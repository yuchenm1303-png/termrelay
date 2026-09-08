package handler

import "github.com/gin-gonic/gin"

// gatewayResponseCommitted is the hard failover boundary for gateway requests.
// Gin marks Writer as written as soon as headers are committed (WriteHeaderNow),
// which also happens on Flush before the first response body byte is written.
// Once this returns true, switching to another upstream can no longer safely
// change the HTTP response and may splice two independent generations together.
func gatewayResponseCommitted(c *gin.Context) bool {
	return c != nil && c.Writer != nil && c.Writer.Written()
}
