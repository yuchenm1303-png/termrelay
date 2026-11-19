package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const cangyuanMediaSubmitBodyLimit = 1 << 20

// TrySubmit opportunistically handles only a model explicitly listed on a
// schedulable Cangyuan media-only account in the authenticated OpenAI group.
// It restores the request body and returns false for all ordinary OpenAI/Grok
// models so existing /v1/images and /v1/videos routing stays unchanged.
func (h *CangyuanMediaHandler) TrySubmit(c *gin.Context, kind service.CangyuanMediaKind) bool {
	if h == nil || h.public == nil || h.billing == nil {
		return false
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.GroupID == nil || *apiKey.GroupID <= 0 ||
		apiKey.Group == nil || apiKey.Group.ID != *apiKey.GroupID ||
		apiKey.Group.Platform != service.PlatformOpenAI {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type"))), "application/json") {
		return false
	}
	data, readErr := io.ReadAll(io.LimitReader(c.Request.Body, cangyuanMediaSubmitBodyLimit+1))
	c.Request.Body = io.NopCloser(bytes.NewReader(data))
	if readErr != nil || len(data) > cangyuanMediaSubmitBodyLimit {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": gin.H{"type": "invalid_request_error", "message": "Media request body is too large or unreadable"},
		})
		return true
	}
	var modelInput struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(data, &modelInput) != nil || strings.TrimSpace(modelInput.Model) == "" {
		return false
	}
	candidates, err := h.public.Candidates(c.Request.Context(), *apiKey.GroupID, kind, modelInput.Model)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"type": "api_error", "message": "Media account scheduling unavailable"},
		})
		return true
	}
	if len(candidates) == 0 {
		return false
	}
	c.Header("Cache-Control", "no-store")
	if apiKey.User == nil || apiKey.Group.Status != service.StatusActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"type": "permission_error", "message": "Media group is unavailable"},
		})
		return true
	}
	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	// The current Cangyuan implementation freezes prepaid balance; do not
	// silently charge that balance on a subscription-mode API key.
	if subscription != nil && apiKey.Group.IsSubscriptionType() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"type": "permission_error", "message": "Cangyuan media currently requires a balance-billed group"},
		})
		return true
	}
	if err := h.billing.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey,
		apiKey.Group, subscription, service.QuotaPlatform(c.Request.Context(), apiKey)); err != nil {
		status, code, message, retryAfter := billingErrorDetails(err)
		if retryAfter > 0 {
			c.Header("Retry-After", "5")
		}
		c.JSON(status, gin.H{"error": gin.H{"type": code, "message": message}})
		return true
	}
	owner := service.CangyuanMediaJobOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID}
	job, err := h.public.Submit(c.Request.Context(), owner, *apiKey.GroupID,
		kind, modelInput.Model, data, c.GetHeader("Idempotency-Key"))
	if err != nil && (job == nil || job.BillingState == "pending" ||
		job.Status == service.CangyuanMediaJobReserved) {
		status := http.StatusServiceUnavailable
		message := "Cangyuan media request could not be accepted"
		switch {
		case errors.Is(err, service.ErrBatchImageInsufficientBalance):
			status, message = http.StatusPaymentRequired, "Insufficient balance for approved media price"
		case errors.Is(err, service.ErrCangyuanMediaPriceNotFound):
			status, message = http.StatusUnprocessableEntity, "This media model and option combination has no enabled verified price"
		case errors.Is(err, service.ErrCangyuanMediaIdempotencyConflict):
			status, message = http.StatusConflict, "Idempotency-Key conflicts with an existing media request"
		case errors.Is(err, service.ErrCangyuanMediaNoAccount):
			status, message = http.StatusServiceUnavailable, "No available media account for this group"
		case err != nil && strings.Contains(err.Error(), "Idempotency-Key"):
			status, message = http.StatusBadRequest, "A valid Idempotency-Key header is required"
		}
		c.JSON(status, gin.H{"error": gin.H{"type": "media_error", "message": message}})
		return true
	}
	if job == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"type": "media_error", "message": "Media task was not created"},
		})
		return true
	}
	if err != nil && job.Status == service.CangyuanMediaJobFailed {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{"type": "media_error", "message": "Upstream rejected the media request"},
			"task_id": job.ID,
			"status": "failed",
		})
		return true
	}
	status := string(job.Status)
	if err != nil && (job.Status == service.CangyuanMediaJobSubmitting ||
		job.Status == service.CangyuanMediaJobSubmitUnknown) {
		// An upstream timeout may mean the submission was accepted.
		status = "submit_unknown"
	}
	c.Header("Retry-After", "6")
	c.JSON(http.StatusAccepted, gin.H{
		"id": job.ID,
		"task_id": job.ID,
		"object": "media.task",
		"model": job.Model,
		"status": status,
		"created_at": job.CreatedAt,
		"message": "Query the task status with the same API key; never resubmit an ambiguous generation",
	})
	return true
}
