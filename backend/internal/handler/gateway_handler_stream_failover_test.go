package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// partialMessageStartSSE 模拟 handleStreamingResponse 已写入的首批 SSE 事件。
const partialMessageStartSSE = "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_01\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"claude-sonnet-4-5\",\"stop_reason\":null,\"stop_sequence\":null,\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}}\n\n" +
	"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n"

// TestStreamWrittenGuard_MessagesPath_AbortFailoverOnSSEContentWritten 验证：
// Forward 一旦向客户端提交任何 SSE 内容，就必须阻止下一次 Forward，避免流拼接。
func TestStreamWrittenGuard_MessagesPath_AbortFailoverOnSSEContentWritten(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	require.False(t, gatewayResponseCommitted(c))

	_, err := c.Writer.Write([]byte(partialMessageStartSSE))
	require.NoError(t, err)
	require.True(t, gatewayResponseCommitted(c), "写入 SSE 内容后必须视为响应已提交")

	failoverErr := &service.UpstreamFailoverError{
		StatusCode:   http.StatusForbidden,
		ResponseBody: []byte(`{"error":{"type":"permission_error","message":"forbidden"}}`),
	}

	h := &GatewayHandler{}
	h.handleFailoverExhausted(c, failoverErr, service.PlatformAnthropic, true)

	body := w.Body.String()
	require.Contains(t, body, "event: message_start")
	require.True(t, strings.HasSuffix(strings.TrimRight(body, "\n"), "}"),
		"响应体应以 SSE error event 的 JSON 对象结尾")
	require.Contains(t, body, `"type":"error"`)

	firstIdx := strings.Index(body, "event: message_start")
	lastIdx := strings.LastIndex(body, "event: message_start")
	assert.Equal(t, firstIdx, lastIdx, "不得因 failover 拼接出第二个 message_start")
}

// TestStreamWrittenGuard_GeminiPath_AbortFailoverOnSSEContentWritten 验证 Gemini 路径行为一致。
func TestStreamWrittenGuard_GeminiPath_AbortFailoverOnSSEContentWritten(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-2.0-flash:streamGenerateContent", nil)

	_, err := c.Writer.Write([]byte(partialMessageStartSSE))
	require.NoError(t, err)
	require.True(t, gatewayResponseCommitted(c))

	failoverErr := &service.UpstreamFailoverError{StatusCode: http.StatusForbidden}
	h := &GatewayHandler{}
	h.handleFailoverExhausted(c, failoverErr, service.PlatformGemini, true)

	body := w.Body.String()
	require.Contains(t, body, "event: message_start")
	require.Contains(t, body, `"type":"error"`)

	firstIdx := strings.Index(body, "event: message_start")
	lastIdx := strings.LastIndex(body, "event: message_start")
	assert.Equal(t, firstIdx, lastIdx, "Gemini 路径不得出现双 message_start")
}

// TestGatewayResponseCommitted_HeaderOnlyCommitBlocksFailover 覆盖旧的 writer-size 守卫漏掉的关键边界：
// 只提交 200/headers、尚未写 body 时 Size 仍可能不变，但此时客户端已经观察到响应，绝不能切上游。
func TestGatewayResponseCommitted_HeaderOnlyCommitBlocksFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	sizeBefore := c.Writer.Size()
	require.False(t, gatewayResponseCommitted(c))

	c.Writer.WriteHeader(http.StatusOK)

	require.Equal(t, sizeBefore, c.Writer.Size(), "header-only commit 不应依赖 body size 变化")
	require.True(t, c.Writer.Written())
	require.True(t, gatewayResponseCommitted(c), "header-only commit 必须立即禁止 retry/failover")
}

func TestGatewayResponseCommitted_NoResponseAllowsFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	require.False(t, c.Writer.Written())
	require.False(t, gatewayResponseCommitted(c), "尚未提交任何响应时仍可按策略 failover")
}
