import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss from 'postcss'

const root = resolve(process.cwd(), 'src')
const stylesheetPath = resolve(root, 'smirel/styles/admin-accounts-polish.css')
const css = readFileSync(stylesheetPath, 'utf8')
const stylesheet = postcss.parse(css)
const main = readFileSync(resolve(root, 'main.ts'), 'utf8')
const accountsPage = readFileSync(resolve(root, 'smirel/pages/AdminAccountsPage.vue'), 'utf8')

function hasRule(part: string) {
  return stylesheet.nodes.some((node) => node.type === 'rule' && node.selector.includes(part))
    || stylesheet.nodes.some((node) =>
      node.type === 'atrule' && node.nodes?.some((child) => child.type === 'rule' && child.selector.includes(part)),
    )
}

describe('upstream accounts visual contract', () => {
  it('imports the page-specific finishing layer after shared workspace overrides', () => {
    const before = main.indexOf("import './smirel/styles/workspace-nav-controls.css'")
    const after = main.indexOf("import './smirel/styles/admin-accounts-polish.css'")
    expect(before).toBeGreaterThan(-1)
    expect(after).toBeGreaterThan(before)
  })

  it('uses bundled provider logos with explicit provider keys', () => {
    expect(accountsPage).toContain(':data-provider="providerLogoKey(item.platform)"')
    expect(accountsPage).toContain("gemini: 'google'")
    expect(accountsPage).toContain("grok: 'xai'")
    expect(accountsPage).toContain(':title="groupName(id)"')
  })

  it('maintains both themes and responsive overflow handling', () => {
    expect(hasRule(".workspace-root .admin-accounts-page .load-track")).toBe(true)
    expect(hasRule(".workspace-root .admin-accounts-page .upstream-groups > b")).toBe(true)
    expect(hasRule(".workspace-root .admin-accounts-page .quick-button")).toBe(true)
    expect(hasRule("html.smirel-app[data-theme='light'] .workspace-root .admin-accounts-page")).toBe(true)
    expect(stylesheet.nodes.some((node) => node.type === 'atrule' && node.name === 'media' && node.params.includes('640px'))).toBe(true)
    expect(stylesheet.nodes.some((node) => node.type === 'atrule' && node.name === 'media' && node.params.includes('prefers-reduced-motion'))).toBe(true)
  })
})
