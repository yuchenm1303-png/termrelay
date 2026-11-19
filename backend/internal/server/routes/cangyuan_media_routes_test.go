package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// Cangyuan GET and POST run through the authenticated /v1 stack.
func TestCangyuanMediaRoutesRegistered(t *testing.T) {
	router := newGatewayRoutesTestRouter()
	registered := make(map[string]struct{})
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}
	require.Contains(t, registered, "GET /v1/images/generations/:task_id")
	require.Contains(t, registered, "GET /v1/images/edits/:task_id")
	require.Contains(t, registered, "GET /v1/videos/:request_id")
	require.Contains(t, registered, "POST /v1/videos")
	require.Contains(t, registered, "POST /videos")
	// Existing Grok video route must remain available and untouched.
	require.Contains(t, registered, "POST /v1/videos/generations")

	for _, path := range []string{
		"/v1/images/generations/cyjob_0123456789abcdef0123456789abcdef",
		"/v1/images/edits/cyjob_0123456789abcdef0123456789abcdef",
		"/v1/videos/cyjob_0123456789abcdef0123456789abcdef",
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusNotFound, w.Code, path)
		require.NotContains(t, w.Body.String(), "video_url")
	}
}
