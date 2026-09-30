import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration } from 'postcss'

function css(path: string) {
  return postcss.parse(readFileSync(resolve(process.cwd(), path), 'utf8'))
}
const aurora = css('src/smirel/styles/home-hero-aurora-motion.css')
const storytelling = css('src/smirel/styles/home-storytelling.css')
const homepage = readFileSync(resolve(process.cwd(), 'src/smirel/pages/HomePage.vue'), 'utf8')

function gradient(root: postcss.Root, selector: string) {
  let result: Declaration | undefined
  root.walkRules(selector, (rule) => {
    rule.walkDecls('background', (entry) => { result = entry })
  })
  expect(result, `${selector} must define the existing aurora gradient`).toBeDefined()
  return result!.value
}

function maxCloudAlpha(background: string) {
  // Exclude neutral hero-to-page fade, evaluate tinted stops only.
  const alphas = [...background.matchAll(/rgba\((\d+),\s*(\d+),\s*(\d+),\s*(0?\.\d+|1(?:\.0+)?)\)/g)]
    .filter((match) => {
      const rgb = [Number(match[1]), Number(match[2]), Number(match[3])]
      return Math.max(...rgb) - Math.min(...rgb) > 20
    })
    .map((match) => Number(match[4]))
  expect(alphas.length).toBeGreaterThan(0)
  return Math.max(...alphas)
}

describe('subtle existing hero mist', () => {
  it('uses low-alpha static color, including the actual important storytelling layers', () => {
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero'))).toBeLessThanOrEqual(.055)
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-hero'))).toBeLessThanOrEqual(.055)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-hero")))
      .toBeLessThanOrEqual(.040)
  })
  it('preserves the original animated cloud elements while reducing their saturation', () => {
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::before'))).toBeLessThanOrEqual(.092)
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::after'))).toBeLessThanOrEqual(.057)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::before")))
      .toBeLessThanOrEqual(.11)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::after")))
      .toBeLessThanOrEqual(.074)
    expect(homepage).toContain("const opacity = Math.max(.54, .76 - current.scroll * .13)")
    expect(aurora.toString()).toContain("width: min(2160px, 150vw)")
    expect(aurora.toString()).toContain("width: min(1920px, 142vw)")
    expect(aurora.toString()).toContain("width: 174vw")
    expect(storytelling.toString()).toContain("animation-name: home-cloud-tail-idle")
  })
  it('keeps secondary section atmosphere restrained rather than a separate color wall', () => {
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-capabilities::before')))
      .toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-capabilities::before")))
      .toBeLessThanOrEqual(.055)
  })
})
