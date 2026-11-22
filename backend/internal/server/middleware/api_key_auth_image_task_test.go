package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAsyncMediaTaskRead(t *testing.T) {
	// Legacy task routes and Cangyuan's current image/video polling routes
	// require authentication and ownership but never another billing check.
	for _, path := range []string{
		"/v1/images/tasks/imgtask_123",
		"/images/tasks/imgtask_123",
		"/v1/images/generations/cyjob_123abc",
		"/v1/images/edits/cyjob_a-b_123",
		"/v1/videos/cyjob_d932d140250643209a1b9bbf890f39df",
	} {
		require.True(t, isAsyncMediaTaskRead(http.MethodGet, path), path)
		require.False(t, isAsyncMediaTaskRead(http.MethodPost, path), path)
		require.False(t, isAsyncMediaTaskRead(http.MethodDelete, path), path)
	}

	// Only exact owner-scoped media task identifiers count as read-only paths.
	// New generation, listing, malformed IDs and unrelated routes still pass
	// through the ordinary quota/subscription billing middleware.
	for _, path := range []string{
		"/v1/images/generations",
		"/v1/images/edits",
		"/v1/videos",
		"/v1/videos/",
		"/v1/videos/example",
		"/v1/videos/cyjob_",
		"/v1/videos/cyjob_x/other",
		"/v1/videos/cyjob_x%2Fother",
		"/v1/videos/cyjob_中文",
		"/v1/videos/cyjob_x?mode=submit",
		"/v1/video/cyjob_123",
		"/v1/images/generations/cyjob_x/other",
		"/v1/images/edits/cyjob_",
	} {
		require.False(t, isAsyncMediaTaskRead(http.MethodGet, path), path)
	}
}
