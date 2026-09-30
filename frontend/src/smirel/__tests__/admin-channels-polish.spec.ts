import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type AtRule, type Declaration, type Rule } from 'postcss'

const src = resolve(process.cwd(), 'src')
const channelsPage = readFileSync(resolve(src, 'smirel/pages/AdminChannelsPage.vue'), 'utf8')
const main = readFileSync(resolve(src, 'main.ts'), 'utf8')
const css = postcss.parse(readFileSync(resolve(src, 'smirel/styles/admin-channels-polish.css'), 'utf8'))

function rules(selector: string, maxWidth?: string): Rule[] {
  const nodes = maxWidth
    ? css.nodes.filter((node): node is AtRule =>
      node.type === 'atrule' && node.name === 'container' && node.params.includes(maxWidth))
        .flatMap((node) => node.nodes || [])
    : css.nodes
  return nodes.filter((node): node is Rule =>
    node.type === 'rule' && postcss.list.comma(node.selector).some((part) => part.trim() === selector))
}

function declaration(selector: string, property: string, maxWidth?: string): string {
  const all = rules(selector, maxWidth).flatMap((rule) =>
    rule.nodes.filter((node): node is Declaration => node.type === 'decl' && node.prop === property))
  expect(all.length, selector + ' should declare ' + property).toBeGreaterThan(0)
  return all.at(-1)!.value
}

function trackCount(template: string) {
  const minmax = (template.match(/minmax\(/g) || []).length
  // Every fixed track in this table is a whole px token (not a minmax arg).
  const fixed = (template.match(/(?:^|\s)\d+px(?=\s|$)/g) || []).length
  return minmax + fixed
}

describe('channels workspace visual integrity', () => {
  it('loads its theme layer after the shared workspace atmosphere', () => {
    expect(main.indexOf("import './smirel/styles/workspace-ambient.css'")).toBeGreaterThan(-1)
    expect(main.indexOf("import './smirel/styles/admin-channels-polish.css'"))
      .toBeGreaterThan(main.indexOf("import './smirel/styles/workspace-ambient.css'"))
    expect(declaration(".workspace-root .channels-page", 'width')).toContain('1480px')
    expect(declaration("html.smirel-app[data-theme='light'] .workspace-root .channels-page", '--ch-panel')).toBe('#fff')
  })

  it('uses actual channel data, readable group labels, and honest pricing fallbacks', () => {
    expect(channelsPage).toContain("'/admin/channels'")
    expect(channelsPage).toContain('api.put(`/admin/channels/${c.id}`')
    expect(channelsPage).toContain('groupLabels(c).slice(0, 2)')
    expect(channelsPage).toContain('billingSourceLabel(c.billing_model_source)')
    expect(channelsPage).toContain("hasFilters ? '没有符合条件的渠道'")
    expect(channelsPage).toContain('class="pricing-fallback"')
    expect(declaration('.workspace-root .channels-page .group-chip', 'overflow-wrap')).toBe('anywhere')
  })

  it('keeps responsive row and header tracks aligned without squeezed action labels', () => {
    const head = '.workspace-root .channels-page .thead'
    const row = '.workspace-root .channels-page .row'
    const expected = [[undefined, 7], ['1199px', 6], ['979px', 5], ['739px', 3], ['499px', 2]] as const
    for (const [maxWidth, count] of expected) {
      const rowTracks = declaration(row, 'grid-template-columns', maxWidth)
      if (maxWidth !== '499px') {
        expect(declaration(head, 'grid-template-columns', maxWidth)).toBe(rowTracks)
      }
      expect(trackCount(rowTracks)).toBe(count)
    }
    expect(declaration('.workspace-root .channels-page .actions', 'flex-wrap')).toBe('nowrap')
    expect(declaration('.workspace-root .channels-page .actions button', 'white-space')).toBe('nowrap')
    expect(declaration('.workspace-root .channels-page .actions button > span', 'display', '499px')).toBe('none')
    expect(channelsPage).not.toContain('.thead>span:nth-child(5)')
  })

  it('scales toolbar and modal for smaller screens, with accessible controls', () => {
    expect(channelsPage).toContain('aria-label="筛选渠道状态"')
    expect(channelsPage).toContain(':aria-label="(c.status ===')
    expect(channelsPage).toContain('aria-label="关闭编辑窗口"')
    expect(rules('.workspace-root .channels-page .toolbar', '739px').length).toBeGreaterThan(0)
    expect(css.nodes.some((node) => node.type === 'atrule' && node.name === 'media' &&
      node.params.includes('prefers-reduced-motion'))).toBe(true)
    expect(declaration('.workspace-root .channels-page .panel', 'container-name')).toBe('channel-panel')
  })
})
