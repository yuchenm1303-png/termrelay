package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// gatewayResponseCommitted reports whether any client-visible response state has
// been committed. Once this becomes true, gateway retry/failover must stop: the
// next upstream attempt could duplicate generation or splice two streams into a
// single client response even when the committed response contained no body yet.
func gatewayResponseCommitted(c *gin.Context) bool {
	if c == nil || c.Writer == nil {
		return false
	}
	if service.IsResponseCommitted(c) {
		return true
	}
	return c.Writer.Written()
}
