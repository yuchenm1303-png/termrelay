import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type AtRule, type Declaration, type Rule } from 'postcss'

const root = resolve(process.cwd(), 'src')
const page = readFileSync(resolve(root, 'smirel/pages/ModelCatalogPage.vue'), 'utf8')
const main = readFileSync(resolve(root, 'main.ts'), 'utf8')
const css = postcss.parse(readFileSync(resolve(root, 'smirel/styles/model-card-finish.css'), 'utf8'))
const card = 'html.smirel-app .workspace-root .model-market-page .model-market-card'
const lightCard = "html.smirel-app[data-theme='light'] .workspace-root .model-market-page .model-market-card"

function rule(selector: string, media?: string): Rule {
  const nodes = media
    ? css.nodes.filter((node): node is AtRule =>
      node.type === 'atrule' && node.name === 'media' && node.params.includes(media))
      .flatMap((node) => node.nodes || [])
    : css.nodes.filter((node) => node.type === 'rule')
  const found = nodes.find((node): node is Rule => node.type === 'rule' &&
    postcss.list.comma(node.selector).some((part) => part.trim() === selector))
  expect(found, selector + (media ? ' at ' + media : '')).toBeDefined()
  return found!
}

function value(selector: string, property: string, media?: string): string | undefined {
  const entry = rule(selector, media).nodes
    .filter((node): node is Declaration => node.type === 'decl' && node.prop === property)
  return entry.at(-1)?.value
}

describe('model detail card finishing layer', () => {
  it('loads after existing compact styling and keeps both theme palettes', () => {
    const compact = main.indexOf("import './smirel/styles/model-card-compact.css'")
    const finish = main.indexOf("import './smirel/styles/model-card-finish.css'")
    const ambient = main.indexOf("import './smirel/styles/workspace-ambient.css'")
    expect(compact).toBeGreaterThan(-1)
    expect(finish).toBeGreaterThan(ambient)
    expect(finish).toBeGreaterThan(compact)
    expect(value(card, '--mm-surface')).toBe('#161d29')
    expect(value(lightCard, '--mm-surface')).toBe('#ffffff')
    expect(value(lightCard, 'background')).toBe('var(--mm-surface)')
    expect(value(lightCard + ':hover', 'background')).toBe('var(--mm-surface)')
    expect(value(card, 'border-radius')).toBe('18px')
  })

  it('preserves existing expandable model data, logos and copy interaction', () => {
    expect(page).toContain(':data-provider="model.providerKey"')
    expect(page).toContain('isModelExpanded(model.id)')
    expect(page).toContain('toggleModelDetails(model.id)')
    expect(page).toContain('copyId(model.id)')
    expect(page).toContain(':title="model.id"')
    expect(page).toContain(':title="offer.group.name"')
    expect(page).toContain('model.cacheWritePerM')
    expect(page).toContain('model.cacheReadPerM')
  })

  it('combines cache/spec data instead of stacking miniature bordered tiles', () => {
    expect(value(card + ' .model-secondary-prices', 'gap')).toBe('0')
    expect(value(card + ' .model-secondary-prices', 'overflow')).toBe('hidden')
    expect(value(card + ' .model-secondary-prices > div', 'border')).toBe('0')
    expect(value(card + ' .model-secondary-prices > div + div', 'border-left')).toBe('1px solid var(--mm-line)')
    expect(value(card + ' .model-specs', 'gap')).toBe('0')
    expect(value(card + ' .model-specs > div', 'border')).toBe('0')
    expect(value(card + ' .model-specs > div + div', 'border-left')).toBe('1px solid var(--mm-line)')
    expect(value(card + ' .model-access-block', 'background')).toBe('var(--mm-surface)')
  })

  it('uses one readable route list with long-name protection and stable rates', () => {
    expect(value(card + ' .model-group-list', 'gap')).toBe('0')
    expect(value(card + ' .model-group-item', 'grid-template-columns')).toContain('minmax(0, 1fr)')
    expect(value(card + ' .model-group-item', 'border-top')).toBe('1px solid var(--mm-line)')
    expect(value(card + ' .model-group-copy > b', 'overflow-wrap')).toBe('anywhere')
    expect(value(card + ' .model-group-copy > b', 'white-space')).toBe('normal')
    expect(value(card + ' .model-group-item > em', 'white-space')).toBe('nowrap')
  })

  it('retains responsive two-column pricing and narrow-screen accessible details', () => {
    expect(value(card + ' .model-detail-layout', 'grid-template-columns', '1120px')).toBe('minmax(0, 1fr)')
    expect(value(card + ' .model-primary-prices', 'grid-template-columns', '380px')).toBe('1fr')
    expect(value(card + ' .model-specs', 'grid-template-columns', '380px')).toBe('minmax(0, 1fr)')
    expect(css.nodes.some((node) => node.type === 'atrule' && node.name === 'media' &&
      node.params.includes('prefers-reduced-motion'))).toBe(true)
  })
})
