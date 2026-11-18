package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const testCangyuanJobID = "cyjob_0123456789abcdef0123456789abcdef"

type cangyuanMediaReadStore struct {
	job *service.CangyuanMediaJob
	err error
}

func (s *cangyuanMediaReadStore) Create(_ context.Context, _ *service.CangyuanMediaJob, _ time.Duration) error {
	return nil
}

func (s *cangyuanMediaReadStore) Get(_ context.Context, _ string) (*service.CangyuanMediaJob, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.job == nil {
		return nil, service.ErrCangyuanMediaJobNotFound
	}
	copy := *s.job
	return &copy, nil
}

func (s *cangyuanMediaReadStore) CompareAndSwap(_ context.Context, _ *service.CangyuanMediaJob, _ int64) error {
	return nil
}

func newCangyuanMediaHandlerTestRouter(h *CangyuanMediaHandler, owner service.CangyuanMediaJobOwner) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if owner.UserID > 0 && owner.APIKeyID > 0 {
			c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
				ID: owner.APIKeyID, UserID: owner.UserID,
			})
		}
		c.Next()
	})
	router.GET("/v1/images/generations/:task_id", h.GetImageGeneration)
	router.GET("/v1/images/edits/:task_id", h.GetImageEdit)
	router.GET("/v1/videos/:request_id", h.GetVideo)
	return router
}

func TestCangyuanMediaReadRoutesOnlyReturnOwnedResults(t *testing.T) {
	job := &service.CangyuanMediaJob{
		ID: testCangyuanJobID, UserID: 8, APIKeyID: 13, AccountID: 29,
		Model: "image-fixture", Kind: service.CangyuanImageGeneration,
		Status: service.CangyuanMediaJobCompleted,
		ResultURLs: []string{"https://cdn.example.org/output.png"},
		BillingState: "pending", CreatedAt: 1700000000, ExpiresAt: 1700200000,
	}
	handler := NewCangyuanMediaHandler(&service.CangyuanMediaJobService{Store: &cangyuanMediaReadStore{job: job}})
	owner := service.CangyuanMediaJobOwner{UserID: 8, APIKeyID: 13}
	r := newCangyuanMediaHandlerTestRouter(handler, owner)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/images/generations/"+testCangyuanJobID, nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.Empty(t, w.Header().Get("Retry-After"))
	var output map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &output))
	require.Equal(t, string(service.CangyuanMediaJobCompleted), output["status"])
	require.Equal(t, "https://cdn.example.org/output.png", output["data"].([]any)[0].(map[string]any)["url"])
	for _, private := range []string{"user_id", "api_key_id", "account_id", "billing_state", "price_snapshot", "upstream_task_id", "unit_sale_price_usd", "api_key"} {
		require.NotContains(t, output, private)
	}

	for _, attempt := range []struct {
		name string
		user int64
		key  int64
		path string
		want int
	}{
		{"wrong API key", 8, 99, "/v1/images/generations/" + testCangyuanJobID, 404},
		{"wrong user", 7, 13, "/v1/images/generations/" + testCangyuanJobID, 404},
		{"wrong capability", 8, 13, "/v1/images/edits/" + testCangyuanJobID, 404},
		{"invalid task ID", 8, 13, "/v1/images/generations/cyjob_invalid", 404},
		{"missing authentication", 0, 0, "/v1/images/generations/" + testCangyuanJobID, 401},
	} {
		t.Run(attempt.name, func(t *testing.T) {
			router := newCangyuanMediaHandlerTestRouter(handler, service.CangyuanMediaJobOwner{UserID: attempt.user, APIKeyID: attempt.key})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, attempt.path, nil))
			require.Equal(t, attempt.want, response.Code)
			require.NotContains(t, response.Body.String(), "cdn.example.org")
		})
	}
}

func TestCangyuanMediaReadVideoAndPendingStatus(t *testing.T) {
	job := &service.CangyuanMediaJob{
		ID: testCangyuanJobID, UserID: 8, APIKeyID: 13, AccountID: 29,
		Model: "video-fixture", Kind: service.CangyuanVideoGeneration,
		Status: service.CangyuanMediaJobQueued, BillingState: "disabled",
	}
	handler := NewCangyuanMediaHandler(&service.CangyuanMediaJobService{Store: &cangyuanMediaReadStore{job: job}})
	r := newCangyuanMediaHandlerTestRouter(handler, service.CangyuanMediaJobOwner{UserID: 8, APIKeyID: 13})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/videos/"+testCangyuanJobID, nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "6", w.Header().Get("Retry-After"))
	require.Contains(t, w.Body.String(), "queued")
	require.NotContains(t, w.Body.String(), "video_url")

	job.Status = service.CangyuanMediaJobCompleted
	job.ResultURLs = []string{"https://cdn.example.org/output.mp4"}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/videos/"+testCangyuanJobID, nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Header().Get("Retry-After"))
	require.Contains(t, w.Body.String(), "https://cdn.example.org/output.mp4")
	require.NotContains(t, w.Body.String(), "billing_state")
}

func TestCangyuanMediaReadStorageFailuresRemainPrivate(t *testing.T) {
	handler := NewCangyuanMediaHandler(&service.CangyuanMediaJobService{Store: &cangyuanMediaReadStore{err: context.DeadlineExceeded}})
	r := newCangyuanMediaHandlerTestRouter(handler, service.CangyuanMediaJobOwner{UserID: 8, APIKeyID: 13})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/videos/"+testCangyuanJobID, nil))
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	require.NotContains(t, w.Body.String(), "deadline")
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
}
