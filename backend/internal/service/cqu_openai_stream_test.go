package service

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func cquSSEDataPayloads(t *testing.T, raw string) []string {
	t.Helper()
	var payloads []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "data: ") {
			payloads = append(payloads, strings.TrimPrefix(line, "data: "))
		}
	}
	return payloads
}

type cquOpenAIChunkView struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func decodeCQUChunk(t *testing.T, payload string) cquOpenAIChunkView {
	t.Helper()
	var chunk cquOpenAIChunkView
	require.NoError(t, json.Unmarshal([]byte(payload), &chunk))
	require.NotEmpty(t, chunk.Choices)
	return chunk
}

func TestWriteCQUStreamAsOpenAISSEProducesChatCompletionChunks(t *testing.T) {
	stream := strings.Join([]string{
		"event: RUN_STARTED\ndata: {\"threadId\":\"thread-1\",\"runId\":\"run-1\"}\n\n",
		"event: TEXT_MESSAGE_START\ndata: {\"messageId\":\"msg-1\",\"role\":\"assistant\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"你\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"好\"}\n\n",
		"event: TEXT_MESSAGE_END\ndata: {\"messageId\":\"msg-1\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"thread-1\",\"conversationId\":\"conv-1\"}\n\n",
	}, "")

	var out bytes.Buffer
	flushes := 0
	outcome, err := WriteCQUStreamAsOpenAISSE(
		context.Background(),
		strings.NewReader(stream),
		&out,
		func() { flushes++ },
		cquOpenAIStreamOptions{ID: "chatcmpl-test", Model: CQUDefaultModelAlias, Created: 1700000000},
	)
	require.NoError(t, err)
	require.Equal(t, "你好", outcome.Text)
	require.Equal(t, "conv-1", outcome.Identifiers.ConversationKey())
	require.Equal(t, "msg-1", outcome.Identifiers.MessageID)
	require.Positive(t, flushes)

	payloads := cquSSEDataPayloads(t, out.String())
	require.Equal(t, "[DONE]", payloads[len(payloads)-1])

	// role chunk, two content chunks, finish chunk, [DONE]
	require.Len(t, payloads, 5)

	first := decodeCQUChunk(t, payloads[0])
	require.Equal(t, "chatcmpl-test", first.ID)
	require.Equal(t, "chat.completion.chunk", first.Object)
	require.Equal(t, int64(1700000000), first.Created)
	require.Equal(t, CQUDefaultModelAlias, first.Model)
	require.Equal(t, "assistant", first.Choices[0].Delta.Role)
	require.Nil(t, first.Choices[0].FinishReason)

	firstDelta := decodeCQUChunk(t, payloads[1])
	require.Equal(t, "你", firstDelta.Choices[0].Delta.Content)
	require.Equal(t, 0, firstDelta.Choices[0].Index)
	require.Equal(t, "好", decodeCQUChunk(t, payloads[2]).Choices[0].Delta.Content)

	finish := decodeCQUChunk(t, payloads[3])
	require.NotNil(t, finish.Choices[0].FinishReason)
	require.Equal(t, "stop", *finish.Choices[0].FinishReason)
	require.Empty(t, finish.Choices[0].Delta.Content)
}

// TOOL_CALL_* and other non-text events must be invisible to the client and
// must not split or truncate the assistant text around them.
func TestWriteCQUStreamAsOpenAISSEIgnoresToolAndStepEvents(t *testing.T) {
	stream := strings.Join([]string{
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"前\"}\n\n",
		"event: TOOL_CALL_START\ndata: {\"toolCallId\":\"tc-1\",\"toolCallName\":\"search\"}\n\n",
		"event: TOOL_CALL_ARGS\ndata: {\"toolCallId\":\"tc-1\",\"delta\":\"{\\\"q\\\":1}\"}\n\n",
		"event: TOOL_CALL_RESULT\ndata: {\"toolCallId\":\"tc-1\",\"content\":\"结果\"}\n\n",
		"event: TOOL_CALL_END\ndata: {\"toolCallId\":\"tc-1\"}\n\n",
		"event: CUSTOM\ndata: {\"name\":\"progress\",\"value\":1}\n\n",
		"event: STEP_FINISHED\ndata: {\"stepName\":\"s1\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"后\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"t1\"}\n\n",
	}, "")

	var out bytes.Buffer
	outcome, err := WriteCQUStreamAsOpenAISSE(context.Background(), strings.NewReader(stream), &out, nil, cquOpenAIStreamOptions{})
	require.NoError(t, err)
	require.Equal(t, "前后", outcome.Text)

	rendered := out.String()
	require.NotContains(t, rendered, "toolCallId")
	require.NotContains(t, rendered, "TOOL_CALL")
	require.NotContains(t, rendered, "结果")
	require.Contains(t, rendered, `"content":"前"`)
	require.Contains(t, rendered, `"content":"后"`)
}

func TestWriteCQUStreamAsOpenAISSEEmitsUsageOnlyWhenUpstreamReportsIt(t *testing.T) {
	withoutUsage := "event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"a\"}\n\nevent: RUN_FINISHED\ndata: {\"threadId\":\"t\"}\n\n"
	var out bytes.Buffer
	_, err := WriteCQUStreamAsOpenAISSE(context.Background(), strings.NewReader(withoutUsage), &out, nil, cquOpenAIStreamOptions{})
	require.NoError(t, err)
	require.NotContains(t, out.String(), `"usage"`)

	withUsage := "event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"a\"}\n\n" +
		"event: RUN_FINISHED\ndata: {\"threadId\":\"t\",\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":5}}\n\n"
	out.Reset()
	outcome, err := WriteCQUStreamAsOpenAISSE(context.Background(), strings.NewReader(withUsage), &out, nil, cquOpenAIStreamOptions{})
	require.NoError(t, err)
	require.Equal(t, 12, outcome.Usage.InputTokens)
	require.Equal(t, 5, outcome.Usage.OutputTokens)
	require.Contains(t, out.String(), `"prompt_tokens":12`)
	require.Contains(t, out.String(), `"total_tokens":17`)
}

func TestWriteCQUStreamAsOpenAISSEStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	_, err := WriteCQUStreamAsOpenAISSE(ctx, strings.NewReader("event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"x\"}\n\n"), &out, nil, cquOpenAIStreamOptions{})
	require.ErrorIs(t, err, context.Canceled)
	require.NotContains(t, out.String(), "[DONE]")
}

func TestWriteCQUStreamAsOpenAISSEReportsRunError(t *testing.T) {
	stream := "event: RUN_ERROR\ndata: {\"message\":\"agent unavailable\"}\n\n"

	var out bytes.Buffer
	_, err := WriteCQUStreamAsOpenAISSE(context.Background(), strings.NewReader(stream), &out, nil, cquOpenAIStreamOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "agent unavailable")
}

func TestAggregateCQUStreamAsOpenAIJSON(t *testing.T) {
	stream := strings.Join([]string{
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"你好\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"，世界\"}\n\n",
		"event: TEXT_MESSAGE_END\ndata: {\"messageId\":\"m-1\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"t-1\"}\n\n",
	}, "")

	payload, outcome, err := AggregateCQUStreamAsOpenAIJSON(
		context.Background(),
		strings.NewReader(stream),
		cquOpenAIStreamOptions{ID: "chatcmpl-x", Model: CQUDefaultModelAlias, Created: 1700000000},
	)
	require.NoError(t, err)
	require.Equal(t, "你好，世界", outcome.Text)
	require.Equal(t, "t-1", outcome.Identifiers.ConversationKey())
	require.Equal(t, "m-1", outcome.Identifiers.MessageID)

	var completion map[string]any
	require.NoError(t, json.Unmarshal(payload, &completion))
	require.Equal(t, "chat.completion", completion["object"])
	require.Equal(t, "chatcmpl-x", completion["id"])
	require.Equal(t, CQUDefaultModelAlias, completion["model"])
	require.NotContains(t, completion, "usage")

	choices, ok := completion["choices"].([]any)
	require.True(t, ok)
	require.Len(t, choices, 1)
	choice, ok := choices[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "stop", choice["finish_reason"])
	message, ok := choice["message"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "assistant", message["role"])
	require.Equal(t, "你好，世界", message["content"])
}

func TestCQUConversationKeyChainsAcrossTurns(t *testing.T) {
	first := []byte(`{"messages":[{"role":"user","content":"第一问"}]}`)
	require.Empty(t, CQUConversationKey(first), "the opening turn has no prior history")

	seed := cquNextConversationKeySeed(first)
	require.NotEmpty(t, seed)
	stored := cquFinalConversationKey(seed, "第一答")
	require.NotEmpty(t, stored)

	second := []byte(`{"messages":[
		{"role":"user","content":"第一问"},
		{"role":"assistant","content":"第一答"},
		{"role":"user","content":"第二问"}
	]}`)
	require.Equal(t, stored, CQUConversationKey(second))

	// A different history must not collide onto the same conversation.
	other := []byte(`{"messages":[
		{"role":"user","content":"另一个问题"},
		{"role":"assistant","content":"第一答"},
		{"role":"user","content":"第二问"}
	]}`)
	require.NotEqual(t, stored, CQUConversationKey(other))
}

func TestCQUConversationStoreExpiresAndBounds(t *testing.T) {
	store := newCQUConversationStore(0, 2)
	store.Put("a", "conv-a")
	require.Equal(t, "conv-a", store.Get("a"))
	require.Empty(t, store.Get("missing"))

	store.Put("b", "conv-b")
	store.Put("c", "conv-c")
	require.LessOrEqual(t, len(store.entries), 2)

	store.Put("", "conv-none")
	store.Put("d", "")
	require.Empty(t, store.Get(""))
	require.Empty(t, store.Get("d"))
}