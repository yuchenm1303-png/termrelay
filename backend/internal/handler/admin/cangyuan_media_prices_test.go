package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cangyuanAdminPriceTestStore struct {
	mu   sync.Mutex
	data map[string]string
}

func (s *cangyuanAdminPriceTestStore) Set(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}
func (s *cangyuanAdminPriceTestStore) GetAll(_ context.Context) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := make(map[string]string, len(s.data))
	for key, value := range s.data {
		values[key] = value
	}
	return values, nil
}
func (s *cangyuanAdminPriceTestStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func TestCangyuanMediaAdminPriceCreateListDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &cangyuanAdminPriceTestStore{data: map[string]string{}}
	h := NewCangyuanMediaPriceHandler(service.NewCangyuanMediaPriceRegistry(store))
	router := gin.New()
	router.GET("/admin/cangyuan-media/prices", h.List)
	router.PUT("/admin/cangyuan-media/prices/:id", h.Upsert)
	router.DELETE("/admin/cangyuan-media/prices/:id", h.Delete)

	// This fixture bypasses the router's real administrator middleware;
	// the production route in routes/admin.go enforces admin + 2FA writes.
	valid := []byte(`{
		"id":"video-tier-720",
		"model":"video-model",
		"kind":"video_generation",
		"options":{"duration":5,"resolution":"720p"},
		"snapshot":{
			"pricing_version":"2026-11-reviewed",
			"currency":"USD",
			"tier":"720p-five-seconds",
			"units":1,
			"unit_sale_price_usd":"2.50",
			"unit_provider_cost_usd":"1.25",
			"verified_by_admin":false
		},
		"enabled":true
	}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/cangyuan-media/prices/video-tier-720", bytes.NewReader(valid))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/admin/cangyuan-media/prices", nil))
	require.Equal(t, http.StatusOK, list.Code)
	require.Equal(t, "no-store", list.Header().Get("Cache-Control"))
	var result struct {
		Data []service.CangyuanMediaPriceRule `json:"data"`
	}
	require.NoError(t, json.Unmarshal(list.Body.Bytes(), &result))
	require.Len(t, result.Data, 1)
	require.True(t, result.Data[0].Snapshot.VerifiedByAdmin)
	require.Equal(t, "720p-five-seconds", result.Data[0].Snapshot.Tier)

	for _, tc := range []struct {
		name string
		url  string
		body []byte
	}{
		{"wrong ID", "/admin/cangyuan-media/prices/another", valid},
		{"unknown JSON key", "/admin/cangyuan-media/prices/video-tier-720", append(append([]byte{}, valid[:len(valid)-1]...), []byte(`,"untrusted":true}`)...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, tc.url, bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/cangyuan-media/prices/video-tier-720", nil))
	require.Equal(t, http.StatusNoContent, w.Code)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/cangyuan-media/prices", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"data":[]`)
}
