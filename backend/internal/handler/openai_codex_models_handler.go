package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (h *OpenAIGatewayHandler) CodexModels(c *gin.Context) {
	lc := service.NewCodexRequestLifecycle()
	reqLog := requestLogger(c, "handler.codex.models")

	if c.Request.Context().Err() != nil {
		reqLog.Warn("codex.models.context_cancelled")
		return
	}

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "invalid_request_error", "API key group is required")
		reqLog.Warn("codex.models.missing_api_key")
		service.CodexMetricsInstance().RecordRequest(http.StatusUnauthorized)
		return
	}
	if apiKey.Group.Platform != service.PlatformOpenAI {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Codex models manifest is only available for OpenAI groups")
		reqLog.Warn("codex.models.wrong_platform", zap.String("platform", apiKey.Group.Platform))
		service.CodexMetricsInstance().RecordRequest(http.StatusNotFound)
		return
	}

	maxAccountSwitches := h.maxAccountSwitches
	if maxAccountSwitches <= 0 {
		maxAccountSwitches = 3
	}
	failedAccountIDs := make(map[int64]struct{})
	switchCount := 0
	var lastUpstreamErr error

	reqLog = reqLog.With(
		zap.Int64("group_id", *apiKey.GroupID),
		zap.String("client_version", c.Query("client_version")),
	)

	for {
		account, err := h.gatewayService.SelectAccountForModelWithExclusions(c.Request.Context(), apiKey.GroupID, "", "", failedAccountIDs)
		if err != nil {
			if c.Request.Context().Err() != nil {
				reqLog.Warn("codex.models.context_cancelled_during_account_select")
				return
			}
			if lastUpstreamErr != nil {
				reqLog.Error("codex.models.no_account_after_retries",
					zap.Error(lastUpstreamErr),
					zap.Int("switch_count", switchCount),
				)
				h.errorResponse(c, infraerrors.Code(lastUpstreamErr), "upstream_error", infraerrors.Message(lastUpstreamErr))
				return
			}
			reqLog.Error("codex.models.no_account_available",
				zap.Error(err),
			)
			h.errorResponse(c, http.StatusServiceUnavailable, "upstream_error", "No available OpenAI accounts")
			return
		}

		setOpsSelectedAccount(c, account.ID, account.Platform)

		reqLog = reqLog.With(
			zap.Int64("account_id", account.ID),
			zap.String("account_type", account.Type),
		)

		manifest, err := h.gatewayService.FetchCodexModelsManifest(c.Request.Context(), account, c.Query("client_version"), c.GetHeader("If-None-Match"))
		if err != nil {
			if c.Request.Context().Err() != nil {
				reqLog.Warn("codex.models.context_cancelled_during_fetch")
				return
			}
			if service.IsRetryableCodexModelsManifestError(err) && switchCount < maxAccountSwitches {
				failedAccountIDs[account.ID] = struct{}{}
				switchCount++
				lastUpstreamErr = err
				service.CodexMetricsInstance().RecordAccountSwitch(false)
				reqLog.Warn("codex.models.retry_with_new_account",
					zap.Int64("failed_account_id", account.ID),
					zap.Int("switch_count", switchCount),
					zap.Error(err),
				)
				continue
			}
			reqLog.Error("codex.models.fetch_failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
			h.errorResponse(c, infraerrors.Code(err), "upstream_error", infraerrors.Message(err))
			return
		}
		if c.Request.Context().Err() != nil {
			reqLog.Warn("codex.models.context_cancelled_after_fetch")
			return
		}

		latencyMs := lc.LatencyMs()
		service.CodexMetricsInstance().RecordLatency(latencyMs)

		if manifest.ETag != "" {
			c.Header("ETag", manifest.ETag)
		}
		if manifest.NotModified {
			c.Status(http.StatusNotModified)
			reqLog.Info("codex.models.not_modified",
				zap.String("etag", manifest.ETag),
				zap.Int64("latency_ms", latencyMs),
			)
			service.CodexMetricsInstance().RecordRequest(http.StatusNotModified)
			return
		}

		reqLog.Info("codex.models.success",
			zap.Int64("account_id", account.ID),
			zap.Int("body_size_bytes", len(manifest.Body)),
			zap.Int64("latency_ms", latencyMs),
			zap.Int("account_switches", switchCount),
		)
		service.CodexMetricsInstance().RecordRequest(http.StatusOK)
		c.Data(http.StatusOK, "application/json", manifest.Body)
		return
	}
}
