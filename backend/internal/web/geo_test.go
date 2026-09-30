package web

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// GEO resources must remain public and self-consistent when the frontend is
// rebuilt. No claim is made that llms.txt influences search ranking.
func TestGeoPublicResources(t *testing.T) {
	public := "../../../frontend/public/"
	guideBytes, err := os.ReadFile(public + "ai-guide.html")
	if err != nil {
		t.Fatalf("public AI guide missing: %v", err)
	}
	guide := string(guideBytes)
	for _, expected := range []string{
		"rel=\"canonical\" href=\"https://muxway.dev/ai-guide.html\"",
		"<h1>Muxway 模枢：统一 AI API 接入说明</h1>",
		"model=\"&lt;your-available-model&gt;\"",
		"https://muxway.dev/v1",
		"以登录后的平台实际信息为准",
	} {
		if !strings.Contains(guide, expected) {
			t.Errorf("AI guide missing %q", expected)
		}
	}
	start := strings.Index(guide, "<script type=\"application/ld+json\">")
	if start == -1 {
		t.Fatal("guide is missing JSON-LD")
	}
	rest := guide[start+len("<script type=\"application/ld+json\">"):]
	end := strings.Index(rest, "</script>")
	if end == -1 {
		t.Fatal("guide JSON-LD is not closed")
	}
	var structured map[string]any
	if err := json.Unmarshal([]byte(rest[:end]), &structured); err != nil {
		t.Fatalf("guide JSON-LD is not valid JSON: %v", err)
	}
	if structured["@id"] != "https://muxway.dev/ai-guide.html#webpage" {
		t.Errorf("guide JSON-LD identity mismatch: %v", structured["@id"])
	}

	llms, err := os.ReadFile(public + "llms.txt")
	if err != nil {
		t.Fatalf("llms.txt missing: %v", err)
	}
	if !strings.Contains(string(llms), "https://muxway.dev/ai-guide.html") {
		t.Fatal("llms.txt does not point at the public guide")
	}

	sitemap, err := os.ReadFile(public + "sitemap.xml")
	if err != nil {
		t.Fatalf("sitemap missing: %v", err)
	}
	if !strings.Contains(string(sitemap), "<loc>https://muxway.dev/ai-guide.html</loc>") {
		t.Fatal("sitemap does not list the AI guide")
	}

	robots, err := os.ReadFile(public + "robots.txt")
	if err != nil {
		t.Fatalf("robots.txt missing: %v", err)
	}
	if !strings.Contains(string(robots), "User-agent: *") {
		t.Fatal("public crawler fallback group missing")
	}
	if strings.Contains(string(robots), "Disallow: /ai-guide") || strings.Contains(string(robots), "Disallow: /llms") {
		t.Fatal("AI resources must not be excluded from public crawling")
	}
	if isFrontendDocumentPath("/ai-guide.html") || isFrontendDocumentPath("/llms.txt") {
		t.Fatal("public GEO files must not receive private-page noindex headers")
	}
	if !strings.Contains(homeSnapshot, `href="/ai-guide.html"`) {
		t.Fatal("crawlable homepage snapshot must link to AI guide")
	}
}
