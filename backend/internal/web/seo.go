package web

import (
  "bytes"
  "net/http"
  "path"
  "strings"

  "github.com/gin-gonic/gin"
)

// The backend serves one SPA document for many routes. Only the public landing
// page is indexable; private SPA URLs receive an HTTP-level noindex even when
// JavaScript is disabled or interrupted.
func SEOHeaders() gin.HandlerFunc {
  return func(c *gin.Context) {
    if (c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead) &&
      isFrontendDocumentPath(c.Request.URL.Path) &&
      !isSearchableHome(c.Request.URL.Path) {
      c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
    }
    c.Next()
  }
}

func isSearchableHome(requestPath string) bool {
  return requestPath == "/" || requestPath == "/home"
}

func isFrontendDocumentPath(requestPath string) bool {
  if requestPath == "/robots.txt" || requestPath == "/sitemap.xml" {
    return false
  }
  for _, prefix := range []string{
    "/api/", "/v1/", "/v1beta/", "/backend-api/", "/antigravity/",
    "/assets/", "/images/", "/videos/", "/.well-known/",
  } {
    if strings.HasPrefix(requestPath, prefix) {
      return false
    }
  }
  if path.Ext(requestPath) != "" {
    return false // static files, including scripts, stylesheets and icons
  }
  return true
}

// The Vue app mounts normally and replaces this crawlable HTML. Inject the
// snapshot only on / and /home: authenticated pages never expose marketing
// markup in their initial document. Keep its claims aligned with HomePage.vue.
const homeSnapshot = `<main class="seo-snapshot" aria-label="Muxway 模枢">
  <header class="seo-snapshot__hero">
    <p class="seo-snapshot__eyebrow">MUXWAY · 模枢</p>
    <h1>一个接口，通达百模</h1>
    <p class="seo-snapshot__lead">One API. Every model.</p>
    <p>沿用熟悉的 SDK 和调用方式，在一个控制台中管理模型、API Key、用量与套餐。</p>
    <nav aria-label="快速入口"><a href="/register">开始使用</a><a href="/home#capabilities">了解平台能力</a></nav>
  </header>
  <section aria-labelledby="seo-capabilities"><h2 id="seo-capabilities">多模型统一 AI API 接入</h2>
    <p>使用统一的 API 入口，在现有开发工具和 SDK 中接入账户可用的模型。支持集中管理 API Key，并在控制台查看请求、Token 与费用记录。</p>
  </section>
  <section aria-labelledby="seo-workflow"><h2 id="seo-workflow">适配已有开发工作流</h2>
    <p>创建账户及 API Key，查看可用模型、套餐和接入文档，再根据具体 SDK 更换 Base URL 与密钥。可用模型、协议与价格以当前控制台为准。</p>
  </section>
  <section aria-labelledby="seo-faq"><h2 id="seo-faq">常见问题</h2>
    <h3>支持哪些模型和协议？</h3>
    <p>以模型广场和当前账户可见模型为准。Muxway 提供统一 API 入口，并保留常见 SDK 的接入方式。</p>
    <h3>现有项目需要重写吗？</h3>
    <p>通常只需替换 Base URL 和 API Key，具体参数和模型名称请以接入文档为准。</p>
    <h3>如何管理团队或项目用量？</h3>
    <p>可以为不同项目创建独立 Key，在控制台查看请求、Token 和费用记录。</p>
  </section>
  <footer><a href="/home#faq">查看完整常见问题</a><a href="/register">注册 Muxway</a></footer>
</main>`

func decorateHomeHTML(body []byte, requestPath string) []byte {
  if !isSearchableHome(requestPath) {
    return body
  }
  // Vite retains the app mount node. An explicit, non-JS fallback avoids
  // shipping a whole second renderer and does not modify Vue's animation code.
  const mount = `<div id="app"></div>`
  content := bytes.Replace(body, []byte(mount), []byte(`<div id="app">`+homeSnapshot+`</div>`), 1)
  return content
}
