package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// This file is the CQU → OpenAI wire translation. It exists so the rest of the
// gateway never sees a CQU-shaped frame: the browser bridge hands back a raw
// CQU (AG-UI style) event stream, and everything below turns it into either an
// OpenAI Chat Completions SSE stream or a single chat.completion object.
//
// Raw CQU SSE is never passed through to the client.

// cquOpenAIStreamOptions carries the client-facing identity of the response.
// Model is the logical model the client asked for (e.g. "cqu-default"), not the
// numeric CQU modelId, which is an internal routing detail.
type cquOpenAIStreamOptions struct {
	ID      string
	Model   string
	Created int64
}

func (o cquOpenAIStreamOptions) withDefaults() cquOpenAIStreamOptions {
	if strings.TrimSpace(o.ID) == "" {
		o.ID = newCQUChatCompletionID()
	}
	if strings.TrimSpace(o.Model) == "" {
		o.Model = CQUDefaultModelAlias
	}
	if o.Created <= 0 {
		o.Created = time.Now().Unix()
	}
	return o
}

func newCQUChatCompletionID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("chatcmpl-cqu%d", time.Now().UnixNano())
	}
	return "chatcmpl-" + hex.EncodeToString(buf)
}

// cquStreamOutcome is what a fully consumed CQU stream produced, independent of
// the wire format it was rendered into.
type cquStreamOutcome struct {
	Identifiers CQUStreamIdentifiers
	Usage       ProviderUsage
	Text        string
}

// consumeCQUStream is the single place that interprets CQU event semantics.
//
// Only TEXT_MESSAGE_CONTENT produces client-visible text. TOOL_CALL_*,
// STEP_*, CUSTOM and STATE_* events are observed for identifiers and usage but
// deliberately emit nothing, so a tool call in the middle of an answer cannot
// truncate or interleave the assistant text.
func consumeCQUStream(ctx context.Context, r io.Reader, onDelta func(string) error) (cquStreamOutcome, error) {
	outcome := cquStreamOutcome{Usage: ProviderUsage{Extra: map[string]int64{}}}
	var text strings.Builder
	var runErr error

	err := ParseCQUSSEContext(ctx, r, func(event CQUSSEEvent) error {
		cquApplyIdentifiers(event, &outcome.Identifiers)
		if len(event.Data) > 0 {
			outcome.Usage = mergeProviderUsage(outcome.Usage, cquBrowserExtractUsage(event.Data))
		}

		switch event.Type {
		case CQUSSEEventTextMessageContent:
			if event.Delta == "" {
				return nil
			}
			_, _ = text.WriteString(event.Delta)
			if onDelta != nil {
				return onDelta(event.Delta)
			}
		case CQUSSEEventRunError:
			// Keep the upstream wording but never the payload: a CQU error body
			// can echo request context we must not re-emit.
			runErr = fmt.Errorf("cqu run failed: %s", cquRunErrorMessage(event.Data))
		}
		return nil
	})

	outcome.Text = text.String()
	if err != nil {
		return outcome, err
	}
	return outcome, runErr
}

// cquRunErrorMessage lifts the upstream failure reason out of a RUN_ERROR
// frame. The value is redacted and truncated because it ends up in an error
// that the gateway logs, and a CQU error string can echo request context.
func cquRunErrorMessage(data []byte) string {
	var payload map[string]any
	if json.Unmarshal(data, &payload) != nil {
		return "unknown error"
	}
	message := cquLookupString(payload, "message", "error", "msg", "reason")
	if message == "" {
		return "unknown error"
	}
	return truncateString(sanitizeUpstreamErrorMessage(message), cquRunErrorMessageMaxBytes)
}

// cquRunErrorMessageMaxBytes keeps a logged upstream reason short enough to be
// a diagnostic rather than a transcript.
const cquRunErrorMessageMaxBytes = 256

// WriteCQUStreamAsOpenAISSE converts a CQU event stream into OpenAI Chat
// Completions SSE and writes it to w, flushing after every frame so a client
// sees tokens as the browser page receives them.
//
// The emitted sequence is: one role chunk, one chunk per text delta, an
// optional usage chunk when CQU reported real counters, a finish chunk with
// finish_reason=stop, and finally `data: [DONE]`.
func WriteCQUStreamAsOpenAISSE(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	flush func(),
	opts cquOpenAIStreamOptions,
) (cquStreamOutcome, error) {
	opts = opts.withDefaults()
	if w == nil {
		return cquStreamOutcome{}, errors.New("cqu sse writer is required")
	}

	writeFrame := func(payload any) error {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode cqu openai chunk: %w", err)
		}
		if _, err := io.WriteString(w, "data: "+string(encoded)+"\n\n"); err != nil {
			return err
		}
		if flush != nil {
			flush()
		}
		return nil
	}

	roleSent := false
	outcome, err := consumeCQUStream(ctx, r, func(delta string) error {
		if !roleSent {
			roleSent = true
			if err := writeFrame(cquOpenAIChunk(opts, map[string]any{"role": "assistant", "content": ""}, nil, nil)); err != nil {
				return err
			}
		}
		return writeFrame(cquOpenAIChunk(opts, map[string]any{"content": delta}, nil, nil))
	})
	if err != nil {
		return outcome, err
	}

	if !roleSent {
		if err := writeFrame(cquOpenAIChunk(opts, map[string]any{"role": "assistant", "content": ""}, nil, nil)); err != nil {
			return outcome, err
		}
	}

	finish := "stop"
	if usage := cquOpenAIUsagePayload(outcome.Usage); usage != nil {
		if err := writeFrame(cquOpenAIChunk(opts, nil, nil, usage)); err != nil {
			return outcome, err
		}
	}
	if err := writeFrame(cquOpenAIChunk(opts, map[string]any{}, &finish, nil)); err != nil {
		return outcome, err
	}
	if _, err := io.WriteString(w, "data: [DONE]\n\n"); err != nil {
		return outcome, err
	}
	if flush != nil {
		flush()
	}
	return outcome, nil
}

// cquOpenAIChunk builds one chat.completion.chunk. A nil delta means the chunk
// carries no choice at all, which is the OpenAI shape for a usage-only chunk.
func cquOpenAIChunk(opts cquOpenAIStreamOptions, delta map[string]any, finishReason *string, usage map[string]any) map[string]any {
	chunk := map[string]any{
		"id":      opts.ID,
		"object":  "chat.completion.chunk",
		"created": opts.Created,
		"model":   opts.Model,
		"choices": []any{},
	}
	if delta != nil {
		choice := map[string]any{
			"index":         0,
			"delta":         delta,
			"finish_reason": nil,
		}
		if finishReason != nil {
			choice["finish_reason"] = *finishReason
		}
		chunk["choices"] = []any{choice}
	}
	if usage != nil {
		chunk["usage"] = usage
	}
	return chunk
}

// cquOpenAIUsagePayload reports CQU's own token counters when it supplied any.
// It never estimates: a fabricated count would flow straight into billing.
func cquOpenAIUsagePayload(usage ProviderUsage) map[string]any {
	if usage.InputTokens <= 0 && usage.OutputTokens <= 0 {
		return nil
	}
	payload := map[string]any{
		"prompt_tokens":     usage.InputTokens,
		"completion_tokens": usage.OutputTokens,
		"total_tokens":      usage.InputTokens + usage.OutputTokens,
	}
	if usage.CachedTokens > 0 {
		payload["prompt_tokens_details"] = map[string]any{"cached_tokens": usage.CachedTokens}
	}
	return payload
}

// AggregateCQUStreamAsOpenAIJSON consumes a CQU stream to completion and
// renders one OpenAI chat.completion object. CQU has no non-streaming endpoint,
// so non-streaming clients are served by buffering the same browser stream.
func AggregateCQUStreamAsOpenAIJSON(ctx context.Context, r io.Reader, opts cquOpenAIStreamOptions) ([]byte, cquStreamOutcome, error) {
	opts = opts.withDefaults()

	outcome, err := consumeCQUStream(ctx, r, nil)
	if err != nil {
		return nil, outcome, err
	}

	completion := map[string]any{
		"id":      opts.ID,
		"object":  "chat.completion",
		"created": opts.Created,
		"model":   opts.Model,
		"choices": []any{
			map[string]any{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": outcome.Text,
				},
				"finish_reason": "stop",
			},
		},
	}
	if usage := cquOpenAIUsagePayload(outcome.Usage); usage != nil {
		completion["usage"] = usage
	}

	encoded, err := json.Marshal(completion)
	if err != nil {
		return nil, outcome, fmt.Errorf("encode cqu chat completion: %w", err)
	}
	return encoded, outcome, nil
}
