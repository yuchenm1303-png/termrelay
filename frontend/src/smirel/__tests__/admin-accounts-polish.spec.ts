import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Rule } from 'postcss'

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

function lastRule(selector: string, media?: string, property?: string): Rule {
  const nodes = media
    ? stylesheet.nodes.filter((node) => node.type === 'atrule' && node.name === 'media' && node.params.includes(media))
        .flatMap((node) => node.type === 'atrule' ? (node.nodes || []) : [])
    : stylesheet.nodes.filter((node) => node.type === 'rule')
  const rules = nodes.filter((node): node is Rule =>
    node.type === 'rule' && postcss.list.comma(node.selector).some((part) => part.trim() === selector)
    && (!property || node.nodes.some((child) => child.type === 'decl' && child.prop === property)))
  expect(rules.length).toBeGreaterThan(0)
  return rules[rules.length - 1]
}

function value(rule: Rule, property: string) {
  return rule.nodes.filter((node) => node.type === 'decl' && node.prop === property)
    .map((node) => node.type === 'decl' ? node.value : '').at(-1)
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

  it('shows complete routing group names instead of clipping the capsule', () => {
    const group = lastRule('.workspace-root .admin-accounts-page .upstream-groups')
    const chip = lastRule('.workspace-root .admin-accounts-page .upstream-groups > b')
    const row = lastRule('.workspace-root .admin-accounts-page .upstream-row')
    expect(value(group, 'flex-wrap')).toBe('wrap')
    expect(value(group, 'overflow')).toBe('visible')
    expect(value(chip, 'white-space')).toBe('normal')
    expect(value(chip, 'overflow-wrap')).toBe('anywhere')
    expect(value(chip, 'height')).toBe('auto')
    expect(value(row, 'padding-block')).toBe('12px')
    expect(value(chip, 'text-overflow')).toBeUndefined()
  })

  it('matches visible row and header columns at each breakpoint', () => {
    const head = '.workspace-root .admin-accounts-page .upstream-table-head'
    const row = '.workspace-root .admin-accounts-page .upstream-row'
    for (const media of [undefined, '1380px', '1180px', '1120px', '860px']) {
      const headColumns = value(lastRule(head, media, 'grid-template-columns'), 'grid-template-columns')
      const rowColumns = value(lastRule(row, media, 'grid-template-columns'), 'grid-template-columns')
      expect(headColumns).toBe(rowColumns)
      expect(headColumns).toBeTruthy()
      // Expected visible columns: check, identity, state, models, load,
      // groups, recency, actions. The legacy breakpoint only hides recency
      // at 1380px, then load at 1120px, then models/groups at 860px.
      const expected = media === '1380px' ? 7 : media === '1180px' || media === '1120px' ? 6 : media === '860px' ? 4 : 8
      const flexibleTracks = (headColumns?.match(/minmax\(/g) || []).length
      const fixedTracks = (headColumns?.match(/(?:^|\\s)\\d+px(?=\\s|$)/g) || []).length
      expect(flexibleTracks + fixedTracks).toBe(expected)
    }
    expect(value(lastRule(row, '640px'), 'grid-template-columns')).toContain('minmax(0, 1fr)')
  })

  it('reserves an unshrinkable action track and never stacks button labels', () => {
    const head = '.workspace-root .admin-accounts-page .upstream-table-head'
    const row = '.workspace-root .admin-accounts-page .upstream-row'
    const actions = '.workspace-root .admin-accounts-page .quick-actions'
    const button = '.workspace-root .admin-accounts-page .quick-button'
    const label = '.workspace-root .admin-accounts-page .quick-button > span'
    const wideTrack = value(lastRule(row, undefined, 'grid-template-columns'), 'grid-template-columns')
    const mediumTrack = value(lastRule(row, '1380px', 'grid-template-columns'), 'grid-template-columns')
    expect(wideTrack?.trim().endsWith('260px')).toBe(true)
    expect(mediumTrack?.trim().endsWith('260px')).toBe(true)
    expect(value(lastRule(head, '1380px', 'grid-template-columns'), 'grid-template-columns')).toBe(mediumTrack)
    expect(value(lastRule(actions, undefined, 'flex-wrap'), 'flex-wrap')).toBe('nowrap')
    expect(value(lastRule(button, undefined, 'min-width'), 'min-width')).toBe('68px')
    expect(value(lastRule(button, undefined, 'flex'), 'flex')).toBe('0 0 auto')
    expect(value(lastRule(label, undefined, 'white-space'), 'white-space')).toBe('nowrap')
    expect(value(lastRule(label, '1180px', 'display'), 'display')).toBe('none')
    expect(value(lastRule(button, '1180px', 'min-width'), 'min-width')).toBe('33px')
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
