package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildChatCompletionsKeepaliveSSEIsSDKVisibleChunk(t *testing.T) {
	frame := buildChatCompletionsKeepaliveSSE("gpt-5.5")
	require.True(t, strings.HasPrefix(frame, "data: "))
	require.True(t, strings.HasSuffix(frame, "\n\n"))

	payloadText := strings.TrimSuffix(strings.TrimPrefix(frame, "data: "), "\n\n")
	var payload struct {
		ID      string            `json:"id"`
		Object  string            `json:"object"`
		Created int64             `json:"created"`
		Model   string            `json:"model"`
		Choices []json.RawMessage `json:"choices"`
	}
	require.NoError(t, json.Unmarshal([]byte(payloadText), &payload))
	require.Equal(t, "chatcmpl-keepalive", payload.ID)
	require.Equal(t, "chat.completion.chunk", payload.Object)
	require.Positive(t, payload.Created)
	require.Equal(t, "gpt-5.5", payload.Model)
	require.NotNil(t, payload.Choices)
	require.Empty(t, payload.Choices)
}
