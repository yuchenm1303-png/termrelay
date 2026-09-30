// Route-specific metadata must match the HTTP-level SEO policy. This also
// protects standalone builds, where the Go HTTP middleware is not present.
export const landingTitle = 'Muxway 模枢 | 统一 AI API 接入 · One API. Every model.'
export const landingDescription = 'Muxway 模枢提供统一 AI API 入口，沿用熟悉的 SDK 和工具，集中管理可用模型、API Key、用量与套餐。'

function ensureMeta(attribute: string, value: string): HTMLMetaElement {
  let element = document.querySelector<HTMLMetaElement>(`meta[${attribute}="${value}"]`)
  if (!element) {
    element = document.createElement('meta')
    element.setAttribute(attribute, value)
    document.head.appendChild(element)
  }
  return element
}

export function syncRouteSEO(path: string): void {
  const landing = path === '/' || path === '/home'
  const robots = ensureMeta('name', 'robots')
  const canonical = document.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  const ogURL = document.querySelector<HTMLMetaElement>('meta[property="og:url"]')
  if (!landing) {
    robots.content = 'noindex,nofollow'
    canonical?.remove()
    ogURL?.remove()
    return
  }

  document.title = landingTitle
  robots.content = 'index,follow'
  ensureMeta('name', 'description').content = landingDescription
  ensureMeta('property', 'og:title').content = landingTitle
  ensureMeta('property', 'og:description').content = landingDescription
  const link = canonical || document.createElement('link')
  link.rel = 'canonical'
  link.href = 'https://muxway.dev/home'
  if (!link.isConnected) document.head.appendChild(link)
  ensureMeta('property', 'og:url').content = 'https://muxway.dev/home'
}
