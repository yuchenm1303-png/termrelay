package handler

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/service/chatgptweb"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeChatGPTWebGateway struct {
	result chatgptweb.SendResult
}

func (f fakeChatGPTWebGateway) Send(context.Context, chatgptweb.SendRequest) (chatgptweb.SendResult, error) {
	return f.result, nil
}

func (f fakeChatGPTWebGateway) SendStream(_ context.Context, _ chatgptweb.SendRequest, sink chatgptweb.StreamSink) (chatgptweb.SendResult, error) {
	if err := sink(chatgptweb.StreamEvent{Type: chatgptweb.StreamEventUpdate, FinalDelta: "hello"}); err != nil {
		return chatgptweb.SendResult{}, err
	}
	return f.result, nil
}

func TestForwardChatGPTWebNonStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	groupID := int64(3)
	h := &OpenAIGatewayHandler{chatGPTWebGateway: fakeChatGPTWebGateway{result: chatgptweb.SendResult{
		Snapshot: chatgptweb.Snapshot{FinalText: "hello"},
		Session:  chatgptweb.SessionState{ClientSessionID: "session-1", ConversationID: "conversation-1", ParentMessageID: "message-1"},
	}}}
	body := []byte(`{"model":"auto","messages":[{"role":"user","content":"hi"}]}`)
	started := false
	h.forwardChatGPTWeb(c, body, &service.APIKey{GroupID: &groupID}, "hash", false, &started)

	require.Equal(t, 200, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"content":"hello"`)
	require.Equal(t, "conversation-1", recorder.Header().Get("X-TermRelay-Conversation-ID"))
}

func TestForwardChatGPTWebStreamingWritesSSEAndDone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	h := &OpenAIGatewayHandler{chatGPTWebGateway: fakeChatGPTWebGateway{result: chatgptweb.SendResult{Session: chatgptweb.SessionState{ParentMessageID: "message-1"}}}}
	body := []byte(`{"model":"auto","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	started := false
	h.forwardChatGPTWeb(c, body, &service.APIKey{}, "hash", true, &started)

	require.True(t, started)
	require.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), `"content":"hello"`)
	require.True(t, strings.HasSuffix(recorder.Body.String(), "data: [DONE]\n\n"))
}
