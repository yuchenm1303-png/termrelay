import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const stylesheet = postcss.parse(readFileSync(
  resolve(process.cwd(), 'src/smirel/styles/home-storytelling.css'), 'utf8',
))
const rules = stylesheet.nodes.filter((node): node is Rule => node.type === 'rule')
function declaration(selector: string, property: string): Declaration {
  const rule = rules.find((node) => node.selector === selector)
  const entry = rule?.nodes.find((node): node is Declaration =>
    node.type === 'decl' && node.prop === property)
  expect(entry, `${selector} must set ${property}`).toBeDefined()
  return entry!
}

describe('homepage transitions', () => {
  it('uses a shared reading-canvas color beneath every section and footer', () => {
    expect(declaration('.home-page', '--story-paper').value).toBe('#f9fbff')
    expect(declaration('.home-page', 'background').value).toBe('var(--story-paper)')
    expect(declaration("html.smirel-app[data-theme='light'] .home-page", 'background').value)
      .toBe('var(--story-paper)')
    expect(declaration('.home-page .home-section', 'border').value).toBe('0')
    expect(declaration('.home-page .home-section', 'border').important).toBe(true)
    expect(declaration('.home-page .home-section', 'background').value).toBe('transparent')
    expect(declaration('.home-page .home-footer', 'background').value).toBe('transparent')
    expect(declaration('.home-page .home-footer', 'background').important).toBe(true)
    expect(declaration('.home-page .home-section', 'padding-inline').value).toBe('0')
    expect(declaration('.home-page .home-section', 'padding-inline').important).toBe(true)
    expect(declaration("html.smirel-app[data-theme='dark'] .home-page main", 'width').important)
      .toBe(true)
    expect(declaration("html.smirel-app[data-theme='dark'] .home-page .home-footer", 'width').important)
      .toBe(true)
    expect(declaration("html.smirel-app[data-theme='dark'] .home-page .home-closing", 'margin-inline').value)
      .toBe('0')
  })

  it('matches the bottom of both hero themes to their underlying canvas', () => {
    const light = declaration('.home-page .home-hero', 'background')
    const dark = declaration("html.smirel-app[data-theme='dark'] .home-page .home-hero", 'background')
    expect(light.important).toBe(true)
    expect(light.value).toContain('#f9fbff 100%')
    expect(dark.important).toBe(true)
    expect(dark.value).toContain('#0b1018 100%')
    expect(declaration("html.smirel-app[data-theme='dark'] .home-page", '--story-paper').value)
      .toBe('#0b1018')
    expect(declaration('.home-page .home-hero + .home-pain::before', 'display').value)
      .toBe('none')
  })

  it('keeps dark hero geometry full-bleed instead of cropping at the main column', () => {
    // Legacy dark CSS sets width:100% and big viewport padding but the
    // light hero has viewport-sized negative margins. Match those values
    // at the dark selector's higher specificity or the title wraps and
    // a hard vertical seam appears at the right edge of the hero.
    const dark = "html.smirel-app[data-theme='dark'] .home-page .home-hero"
    const lightLayout = '.home-page .home-hero'
    const fullBleedMargin = 'calc(50% - 50vw)'
    expect(declaration(dark, 'width').value).toBe('100vw')
    expect(declaration(dark, 'margin-inline').value).toBe(fullBleedMargin)
    expect(declaration(dark, 'padding-inline').value)
      .toBe('max(24px, calc((100vw - 1320px) / 2))')
    expect(declaration(dark, 'overflow').value).toBe('visible')
    // The basic hero rule still has its responsive spacing and atmospheric
    // fade; dark mode corrects geometry without creating new cloud layers.
    expect(declaration(lightLayout, 'padding-top').value).toContain('clamp(')
    const mobile = stylesheet.nodes
      .filter((node) => node.type === 'atrule' && node.name === 'media' && node.params.includes('820px'))
      .flatMap((node) => node.nodes ?? [])
      .find((node): node is Rule => node.type === 'rule' && node.selector === dark)
    const mobileValues = mobile?.nodes.filter((node): node is Declaration => node.type === 'decl')
    expect(mobileValues?.find((node) => node.prop === 'width')?.value).toBe('auto')
    expect(mobileValues?.find((node) => node.prop === 'margin-inline')?.value).toBe('0')
    expect(mobileValues?.find((node) => node.prop === 'padding-inline')?.value).toBe('0')
  })

  it('lets the capabilities atmosphere extend and fade without moving its content', () => {
    const stage = '.home-page .home-capabilities'
    const wash = '.home-page .home-capabilities::before'
    expect(declaration(stage, 'margin-inline').value).toBe('0')
    expect(declaration(stage, 'margin-inline').important).toBe(true)
    expect(declaration(stage, 'overflow').value).toBe('visible')
    expect(declaration(stage, 'background').value).toBe('transparent')
    expect(declaration(wash, 'width').value).toBe('100vw')
    expect(declaration(wash, 'mask-image').value).toContain('transparent 100%')
    expect(declaration(wash, '-webkit-mask-image').value).toContain('transparent 100%')
    expect(declaration('.home-page .home-capabilities::after', 'display').value).toBe('none')
  })
})
