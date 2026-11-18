package handler

import (
	"errors"
	"net/http"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// CangyuanMediaHandler exposes only the owner-scoped task read API.
// Submission and settlement routes remain unregistered until admin-approved
// prices, quota reservation, entitlement and usage logging are available.
type CangyuanMediaHandler struct {
	jobs *service.CangyuanMediaJobService
}

func NewCangyuanMediaHandler(jobs *service.CangyuanMediaJobService) *CangyuanMediaHandler {
	return &CangyuanMediaHandler{jobs: jobs}
}

func (h *CangyuanMediaHandler) GetImageGeneration(c *gin.Context) {
	h.get(c, service.CangyuanImageGeneration, "task_id")
}

func (h *CangyuanMediaHandler) GetImageEdit(c *gin.Context) {
	h.get(c, service.CangyuanImageEdit, "task_id")
}

func (h *CangyuanMediaHandler) GetVideo(c *gin.Context) {
	h.get(c, service.CangyuanVideoGeneration, "request_id")
}

func (h *CangyuanMediaHandler) get(c *gin.Context, expectedKind service.CangyuanMediaKind, pathParam string) {
	c.Header("Cache-Control", "no-store")
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.ID <= 0 || apiKey.UserID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"type": "authentication_error", "message": "Invalid API key"},
		})
		return
	}
	if h == nil || h.jobs == nil || h.jobs.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"type": "api_error", "message": "Cangyuan media task storage is unavailable"},
		})
		return
	}

	job, err := h.jobs.Get(c.Request.Context(),
		service.CangyuanMediaJobOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID},
		c.Param(pathParam))
	if errors.Is(err, service.ErrCangyuanMediaJobNotFound) ||
		(err == nil && (job == nil || job.Kind != expectedKind || job.BillingState == "disabled")) {
		// Do not reveal whether another user's job ID exists.
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"type": "not_found_error", "message": "Media task not found"},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{"type": "api_error", "message": "Media task storage is unavailable"},
		})
		return
	}
	// An upstream completion is not itself proof of local payment. Never
	// expose expensive image/video URLs while SQL settlement is pending.
	// Disabled/unpriced jobs are hidden entirely by the owner check above.
	if job.Status == service.CangyuanMediaJobCompleted && job.BillingState != "settled" {
		c.Header("Retry-After", "6")
		c.JSON(http.StatusAccepted, gin.H{
			"id": job.ID,
			"task_id": job.ID,
			"status": "in_progress",
			"message": "Media result is being finalized",
		})
		return
	}
	status := string(job.Status)
	object := "image.generation.task"
	if expectedKind == service.CangyuanVideoGeneration {
		object = "video"
	}
	response := gin.H{
		"id":         job.ID,
		"task_id":    job.ID,
		"object":     object,
		"model":      job.Model,
		"status":     status,
		"created_at": job.CreatedAt,
		"expires_at": job.ExpiresAt,
	}
	switch job.Status {
	case service.CangyuanMediaJobReserved, service.CangyuanMediaJobSubmitting,
		service.CangyuanMediaJobQueued, service.CangyuanMediaJobProcessing:
		c.Header("Retry-After", "6")
	case service.CangyuanMediaJobCompleted:
		if expectedKind == service.CangyuanVideoGeneration {
			if len(job.ResultURLs) > 0 {
				response["video_url"] = job.ResultURLs[0]
			}
		} else {
			data := make([]gin.H, 0, len(job.ResultURLs))
			for _, resultURL := range job.ResultURLs {
				data = append(data, gin.H{"url": resultURL})
			}
			response["data"] = data
		}
	case service.CangyuanMediaJobSubmitUnknown:
		response["message"] = "Upstream submission status is unknown; task requires reconciliation"
	}
	c.JSON(http.StatusOK, response)
}
