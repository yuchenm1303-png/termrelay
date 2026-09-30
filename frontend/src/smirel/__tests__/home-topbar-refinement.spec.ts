import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss from 'postcss'

const src = readFileSync(resolve(process.cwd(), 'src/smirel/pages/HomePage.vue'), 'utf8')
const css = postcss.parse(
  readFileSync(resolve(process.cwd(), 'src/smirel/styles/home-topbar-refinement.css'), 'utf8'),
)

function hasSelector(selector: string) {
  let found = false
  css.walkRules((rule) => { if (rule.selector.includes(selector)) found = true })
  return found
}

describe('Homepage topbar refinement', () => {
  it('keeps the existing logo asset and accessible, section-aware navigation', () => {
    expect(src).toContain("import '../styles/home-topbar-refinement.css'")
    expect(src).toContain('<img :src="logoUrl" alt="" />')
    for (const section of ['capabilities', 'tools', 'pricing', 'faq']) {
      expect(src).toContain(`'is-current': activeSection === '${section}'`)
      expect(src).toContain(`href="#${section}"`)
    }
    expect(src).toContain('aria-controls="home-primary-nav"')
    expect(src).toContain(':aria-expanded="mobileMenuOpen"')
    expect(src).toContain("'is-scrolled': headerScrolled")
  })
  it('centralizes header colors, responsive navigation and grouped utilities', () => {
    expect(hasSelector('.home-topbar.is-scrolled')).toBe(true)
    expect(hasSelector('.home-nav a.is-current::before')).toBe(true)
    expect(hasSelector('.home-workspace-controls .home-utility-button')).toBe(true)
    expect(hasSelector('.home-account-menu--toolbar .home-account-trigger')).toBe(true)
    expect(hasSelector("html.smirel-app[data-theme='dark']")).toBe(true)
    const media: string[] = []
    css.walkAtRules('media', (rule) => media.push(rule.params))
    expect(media).toContain('(max-width: 1040px)')
    expect(media).toContain('(max-width: 620px)')
    expect(media).toContain('(prefers-reduced-motion: reduce)')
    expect(src).toContain('<HomeTopbarControls />')
    expect(src).toContain('<HomeAccountMenu variant="toolbar" />')
  })
})
