import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const root = resolve(process.cwd(), 'src')
const ops = readFileSync(resolve(root, 'smirel/components/AdminOpsPage.vue'), 'utf8')
const workspace = readFileSync(resolve(root, 'smirel/pages/WorkspacePage.vue'), 'utf8')
const main = readFileSync(resolve(root, 'main.ts'), 'utf8')
const css = postcss.parse(readFileSync(resolve(root, 'smirel/styles/admin-ops-polish.css'), 'utf8'))
const base = '.workspace-root .ops-host-page .admin-ops-workspace'

function matchingRules(selector: string) {
  const found: Rule[] = []
  css.walkRules((rule) => {
    if (postcss.list.comma(rule.selector).some((part) => part.trim() === selector)) found.push(rule)
  })
  return found
}

function property(selector: string, name: string) {
  const declarations = matchingRules(selector).flatMap((rule) =>
    rule.nodes.filter((node): node is Declaration => node.type === 'decl' && node.prop === name))
  expect(declarations.length, selector + ' declares ' + name).toBeGreaterThan(0)
  return declarations.at(-1)?.value
}

describe('operations monitor visual contract', () => {
  it('loads finishing styles after workspace theme rules without altering other routes', () => {
    expect(main.indexOf("import './smirel/styles/admin-ops-polish.css'"))
      .toBeGreaterThan(main.indexOf("import './smirel/styles/model-card-finish.css'"))
    expect(workspace).toContain(":class=\"{ 'ops-host-page': isAdminOps }\"")
    expect(workspace).toContain("'ops-page-heading': isAdminOps")
    expect(property(base, '--ops-panel')).toBe('#151c28')
    expect(property("html.smirel-app[data-theme='light'] " + base, '--ops-panel')).toBe('#fff')
    expect(property(base + ' .ops-snapshot-bar', 'border-radius')).toBe('13px')
  })

  it('preserves source truth and does not fabricate chart trends for one-day data', () => {
    expect(ops).toContain("api.get<AdminDashboardStats>('/admin/dashboard/stats')")
    expect(ops).toContain("api.get<UsageTrendResponse>('/admin/dashboard/trend'")
    expect(ops).toContain('trend.value.length')
    expect(ops).toContain('单日记录')
    expect(ops).toContain('<polyline v-if="trend.length > 1"')
    expect(ops).toContain('class="ops-plot-point"')
    expect(ops).toContain(':cy="chart.singlePointY"')
    expect(ops).toContain('仅一天记录 · 无法构成趋势')
    expect(ops).toContain(':aria-label="\`\${chart.label}')
    expect(ops).toContain("stats.stats_stale")
  })

  it('keeps text readable, prevents the empty-chart look and styles period controls', () => {
    expect(property(base + ' .ops-metric-label', 'font-size')).toBe('12px')
    expect(property(base + ' .ops-trend-plot .ops-plot-point', 'fill')).toBe('currentColor')
    expect(property(base + ' .ops-period-switch button', 'font-family')).toBe('inherit')
    expect(property(base + ' .ops-trend-card', 'grid-template-columns')).toContain('minmax(0, 1fr)')
    expect(ops).toContain(':aria-pressed="periodDays === days"')
  })

  it('covers short viewports, sidebar-width changes and reduced-motion preferences', () => {
    const widths = ['1200px', '1080px', '760px', '420px']
    for (const width of widths) {
      expect(css.nodes.some((node) => node.type === 'atrule' && node.name === 'media' &&
        node.params.includes(width)), width).toBe(true)
    }
    expect(css.nodes.some((node) => node.type === 'atrule' && node.name === 'media' &&
      node.params.includes('prefers-reduced-motion'))).toBe(true)
    expect(matchingRules(base + ' .ops-single-axis').length).toBeGreaterThan(0)
  })
})
