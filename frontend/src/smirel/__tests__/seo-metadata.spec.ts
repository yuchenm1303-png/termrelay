import { beforeEach, describe, expect, it } from 'vitest'
import { landingDescription, landingTitle, syncRouteSEO } from '../core/seo'

describe('route-aware SEO metadata', () => {
  beforeEach(() => {
    document.head.innerHTML = '<title>Initial</title><meta name="description" content="old">'
  })

  it('makes the public homepage discoverable with a single canonical', () => {
    syncRouteSEO('/home')
    syncRouteSEO('/')
    expect(document.title).toBe(landingTitle)
    expect(document.querySelector('meta[name="description"]')?.getAttribute('content')).toBe(landingDescription)
    expect(document.querySelector('meta[name="robots"]')?.getAttribute('content')).toBe('index,follow')
    expect(document.querySelectorAll('link[rel="canonical"]')).toHaveLength(1)
    expect(document.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe('https://muxway.dev/home')
  })

  it('removes marketing canonical and noindexes the logged-in workspace', () => {
    syncRouteSEO('/home')
    syncRouteSEO('/admin/dashboard')
    expect(document.querySelector('meta[name="robots"]')?.getAttribute('content')).toBe('noindex,nofollow')
    expect(document.querySelector('link[rel="canonical"]')).toBeNull()
    expect(document.querySelector('meta[property="og:url"]')).toBeNull()
  })

  it('keeps login and unknown routes out of the index', () => {
    for (const path of ['/login', '/register', '/some-missing-route']) {
      syncRouteSEO(path)
      expect(document.querySelector('meta[name="robots"]')?.getAttribute('content')).toBe('noindex,nofollow')
    }
  })
})
