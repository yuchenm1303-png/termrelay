package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLandingSnapshotOnlyOnPublicHomepage(t *testing.T) {
	source := []byte(`<!doctype html><html><body><div id="app"></div></body></html>`)
	for _, route := range []string{"/", "/home"} {
		got := string(decorateHomeHTML(source, route))
		if !strings.Contains(got, `<h1>一个接口，通达百模</h1>`) ||
			!strings.Contains(got, `<div id="app"><main`) ||
			!strings.Contains(got, `href="/register"`) {
			t.Errorf("home %s missing crawlable copy or working links", route)
		}
		if strings.Count(got, `<h1>`) != 1 {
			t.Errorf("home %s has duplicate headings", route)
		}
	}
	for _, route := range []string{"/dashboard", "/login", "/admin/dashboard", "/does-not-exist"} {
		if got := string(decorateHomeHTML(source, route)); got != string(source) {
			t.Errorf("private route %s unexpectedly received landing copy", route)
		}
	}
	if got := decorateHomeHTML([]byte("different markup"), "/home"); string(got) != "different markup" {
		t.Fatal("snapshot must not corrupt an unknown index template")
	}
}

func TestHomepageTitleOverridesLegacySiteSettingOnlyForPublicLanding(t *testing.T) {
	oldTitle := `<title>Sub2API - AI API Gateway</title>`
	source := []byte(`<html><head>` + oldTitle + `<link rel="canonical" href="https://muxway.dev/home" /></head><body><div id="app"></div></body></html>`)
	home := string(decorateHomeHTML(source, "/home"))
	if !strings.Contains(home, `<title>Muxway 模枢 | 统一 AI API 接入 · One API. Every model.</title>`) {
		t.Fatal("public homepage must retain its Muxway title after settings injection")
	}
	if strings.Contains(home, oldTitle) {
		t.Fatal("upstream branding must not leak into the indexed homepage title")
	}
	private := string(decorateHomeHTML(source, "/dashboard"))
	if !strings.Contains(private, oldTitle) {
		t.Fatal("private console routes retain the dynamically configured title")
	}
}

func TestSEOHeadersProtectPrivateSPAWithoutTaggingAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SEOHeaders())
	r.Any("/*rest", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	for _, tc := range []struct {
		route string
		want  string
	}{
		{"/", ""}, {"/home", ""},
		{"/dashboard", "noindex, nofollow, noarchive"},
		{"/admin/dashboard", "noindex, nofollow, noarchive"},
		{"/login", "noindex, nofollow, noarchive"},
		{"/register", "noindex, nofollow, noarchive"},
		{"/nonexistent", "noindex, nofollow, noarchive"},
		{"/robots.txt", ""}, {"/sitemap.xml", ""},
		{"/assets/app-abcdef.js", ""}, {"/api/v1/settings/public", ""},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.route, nil))
		if got := w.Header().Get("X-Robots-Tag"); got != tc.want {
			t.Errorf("%s: X-Robots-Tag=%q, want %q", tc.route, got, tc.want)
		}
	}
}
