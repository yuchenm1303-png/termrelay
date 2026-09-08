package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayResponseCommitted_InitiallyFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	require.False(t, gatewayResponseCommitted(c))
}

func TestGatewayResponseCommitted_HeaderOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	c.Status(http.StatusOK)

	require.True(t, gatewayResponseCommitted(c), "WriteHeader must cross the failover boundary even with a zero-byte body")
	require.Equal(t, 0, c.Writer.Size())
}

func TestGatewayResponseCommitted_FlushBeforeFirstSSEChunk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Header("Content-Type", "text/event-stream")

	c.Writer.Flush()

	require.True(t, gatewayResponseCommitted(c), "Flush must cross the failover boundary before the first SSE data chunk")
	require.Equal(t, 0, c.Writer.Size())
}

func TestGatewayResponseCommitted_PartialBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	_, err := c.Writer.Write([]byte("data: {\"id\":\"chunk-1\"}\n\n"))
	require.NoError(t, err)
	require.True(t, gatewayResponseCommitted(c))
}
