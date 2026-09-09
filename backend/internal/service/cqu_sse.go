package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// CQU emits AG-UI-style Server-Sent Events. Keep the event names here instead
// of scattering string literals through the provider adapter.
const (
	CQUSSEEventRunStarted         = "RUN_STARTED"
	CQUSSEEventRunFinished        = "RUN_FINISHED"
	CQUSSEEventRunError           = "RUN_ERROR"
	CQUSSEEventStepStarted        = "STEP_STARTED"
	CQUSSEEventStepFinished       = "STEP_FINISHED"
	CQUSSEEventTextMessageStart   = "TEXT_MESSAGE_START"
	CQUSSEEventTextMessageContent = "TEXT_MESSAGE_CONTENT"
	CQUSSEEventTextMessageEnd     = "TEXT_MESSAGE_END"
	CQUSSEEventToolCallStart      = "TOOL_CALL_START"
	CQUSSEEventToolCallArgs       = "TOOL_CALL_ARGS"
	CQUSSEEventToolCallEnd        = "TOOL_CALL_END"
	CQUSSEEventStateSnapshot      = "STATE_SNAPSHOT"
	CQUSSEEventStateDelta         = "STATE_DELTA"
	CQUSSEEventMessagesSnapshot   = "MESSAGES_SNAPSHOT"
	CQUSSEEventRaw                = "RAW"
	CQUSSEEventCustom             = "CUSTOM"
)

// CQUSSEEvent is the minimally-normalized representation of one upstream SSE
// frame. Data remains available verbatim so future CQU event variants can be
// supported without changing the transport layer.
type CQUSSEEvent struct {
	Type  string
	Data  json.RawMessage
	Delta string
	Raw   []byte
}

// CQUStreamIdentifiers are the server-side conversation handles CQU reports
// while a run is in flight. They are routing identifiers, never credentials,
// so they are safe to log and to reuse on the next turn.
type CQUStreamIdentifiers struct {
	ThreadID       string
	ConversationID string
	RunID          string
	MessageID      string
}

// ConversationKey is the identifier CQU expects back in `conversation_id` on a
// follow-up turn. CQU has used both `conversationId` and `threadId` for it, so
// prefer the explicit one and fall back to the thread handle.
func (ids CQUStreamIdentifiers) ConversationKey() string {
	if value := strings.TrimSpace(ids.ConversationID); value != "" {
		return value
	}
	return strings.TrimSpace(ids.ThreadID)
}

// ParseCQUSSE incrementally parses a CQU event stream. It deliberately does not
// use bufio.Scanner so a large model/tool event is not capped by Scanner's token
// limit.
func ParseCQUSSE(r io.Reader, emit func(CQUSSEEvent) error) error {
	return ParseCQUSSEContext(context.Background(), r, emit)
}

// ParseCQUSSEContext is ParseCQUSSE with cancellation. A client disconnect must
// stop the parse immediately rather than draining the browser stream, so the
// context is checked before every frame is read and before every dispatch.
func ParseCQUSSEContext(ctx context.Context, r io.Reader, emit func(CQUSSEEvent) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if r == nil {
		return errors.New("cqu sse reader is nil")
	}
	if emit == nil {
		return errors.New("cqu sse emitter is nil")
	}

	reader := bufio.NewReader(r)
	var eventName string
	var dataLines [][]byte
	var raw bytes.Buffer

	dispatch := func() error {
		if eventName == "" && len(dataLines) == 0 {
			raw.Reset()
			return nil
		}

		data := bytes.Join(dataLines, []byte("\n"))
		eventType := strings.TrimSpace(eventName)
		if eventType == "" {
			eventType = cquEventTypeFromJSON(data)
		}

		event := CQUSSEEvent{
			Type:  eventType,
			Data:  append(json.RawMessage(nil), data...),
			Raw:   append([]byte(nil), raw.Bytes()...),
			Delta: cquTextDelta(eventType, data),
		}
		if err := emit(event); err != nil {
			return err
		}

		eventName = ""
		dataLines = dataLines[:0]
		raw.Reset()
		return nil
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			_, _ = raw.Write(line)
			trimmed := bytes.TrimRight(line, "\r\n")
			if len(trimmed) == 0 {
				if dispatchErr := dispatch(); dispatchErr != nil {
					return dispatchErr
				}
			} else if trimmed[0] != ':' {
				field, value, found := bytes.Cut(trimmed, []byte(":"))
				if found && len(value) > 0 && value[0] == ' ' {
					value = value[1:]
				}
				switch string(field) {
				case "event":
					eventName = string(value)
				case "data":
					dataLines = append(dataLines, append([]byte(nil), value...))
				}
			}
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("read cqu sse: %w", err)
			}
			if raw.Len() > 0 || eventName != "" || len(dataLines) > 0 {
				if dispatchErr := dispatch(); dispatchErr != nil {
					return dispatchErr
				}
			}
			return nil
		}
	}
}

func cquEventTypeFromJSON(data []byte) string {
	if len(bytes.TrimSpace(data)) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"type", "event", "eventType", "event_type"} {
		if value, ok := payload[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// cquApplyIdentifiers folds one event's routing handles into the running state.
// RUN_STARTED carries the thread/conversation handle for the whole run,
// RUN_FINISHED repeats it, and TEXT_MESSAGE_* carry the assistant message id.
// Values are only ever overwritten by a non-empty replacement so a late event
// without the field cannot erase a handle observed earlier.
func cquApplyIdentifiers(event CQUSSEEvent, ids *CQUStreamIdentifiers) {
	if ids == nil || len(bytes.TrimSpace(event.Data)) == 0 {
		return
	}
	switch event.Type {
	case CQUSSEEventRunStarted, CQUSSEEventRunFinished, CQUSSEEventRunError,
		CQUSSEEventTextMessageStart, CQUSSEEventTextMessageContent, CQUSSEEventTextMessageEnd,
		CQUSSEEventStepStarted, CQUSSEEventStepFinished, CQUSSEEventCustom:
	default:
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return
	}
	assign := func(target *string, keys ...string) {
		if value := cquLookupString(payload, keys...); value != "" {
			*target = value
		}
	}
	assign(&ids.ThreadID, "threadId", "thread_id")
	assign(&ids.ConversationID, "conversationId", "conversation_id", "customConversationId", "custom_conversation_id")
	assign(&ids.RunID, "runId", "run_id")
	assign(&ids.MessageID, "messageId", "message_id")
}

// cquLookupString reads a string field from an event payload, also looking one
// level into a `data` wrapper because CQU nests some events that way.
func cquLookupString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	nested, ok := payload["data"].(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		if value, ok := nested[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func cquTextDelta(eventType string, data []byte) string {
	if eventType != CQUSSEEventTextMessageContent || len(bytes.TrimSpace(data)) == 0 {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}

	// CQU currently follows AG-UI where TEXT_MESSAGE_CONTENT carries `delta`.
	// The fallbacks keep the parser tolerant to observed wrapper variants while
	// still only extracting text from a content event.
	for _, key := range []string{"delta", "content", "text"} {
		if value, ok := payload[key].(string); ok {
			return value
		}
	}
	if nested, ok := payload["data"].(map[string]any); ok {
		for _, key := range []string{"delta", "content", "text"} {
			if value, ok := nested[key].(string); ok {
				return value
			}
		}
	}
	return ""
}
