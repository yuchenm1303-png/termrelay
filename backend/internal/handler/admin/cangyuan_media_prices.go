package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// CangyuanMediaPriceHandler is mounted under the existing administrator
// authentication/audit middleware. Price writes additionally require 2FA.
// These routes DO NOT enable customer-facing media generation by themselves.
type CangyuanMediaPriceHandler struct {
	registry *service.CangyuanMediaPriceRegistry
}

func NewCangyuanMediaPriceHandler(registry *service.CangyuanMediaPriceRegistry) *CangyuanMediaPriceHandler {
	return &CangyuanMediaPriceHandler{registry: registry}
}

func (h *CangyuanMediaPriceHandler) List(c *gin.Context) {
	if h == nil || h.registry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cangyuan pricing storage unavailable"})
		return
	}
	rules, err := h.registry.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cangyuan pricing storage unavailable"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *CangyuanMediaPriceHandler) Upsert(c *gin.Context) {
	if h == nil || h.registry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cangyuan pricing storage unavailable"})
		return
	}
	// Admin pricing is intentionally small. Reject oversized/unknown data
	// instead of persisting a secret, prompt or arbitrary provider JSON.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var rule service.CangyuanMediaPriceRule
	if err := decoder.Decode(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Cangyuan price rule"})
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cangyuan price rule must contain one JSON object"})
		return
	}
	if rule.ID != c.Param("id") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cangyuan price rule ID mismatch"})
		return
	}
	if err := h.registry.Put(c.Request.Context(), rule); err != nil {
		// Do not expose potentially sensitive raw request contents or
		// internal database error messages.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cangyuan price rule was rejected or could not be saved"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"id": rule.ID, "saved": true})
}

func (h *CangyuanMediaPriceHandler) Delete(c *gin.Context) {
	if h == nil || h.registry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cangyuan pricing storage unavailable"})
		return
	}
	if err := h.registry.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cangyuan price rule could not be removed"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusNoContent)
}
