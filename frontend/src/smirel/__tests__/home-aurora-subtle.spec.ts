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
  const alphas = [...background.matchAll(/rgba\([^()]+?,\s*(0?\.\d+|1(?:\.0+)?)\)/g)]
    .map((match) => Number(match[1]))
  expect(alphas.length).toBeGreaterThan(0)
  return Math.max(...alphas)
}

describe('subtle existing hero mist', () => {
  it('uses low-alpha static color, including the actual important storytelling layers', () => {
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero'))).toBeLessThanOrEqual(.12)
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-hero'))).toBeLessThanOrEqual(.12)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-hero")))
      .toBeLessThanOrEqual(.08)
  })
  it('preserves the original animated cloud elements while reducing their saturation', () => {
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::before'))).toBeLessThanOrEqual(.20)
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::after'))).toBeLessThanOrEqual(.115)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::before")))
      .toBeLessThanOrEqual(.22)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::after")))
      .toBeLessThanOrEqual(.145)
    expect(homepage).toContain("const opacity = Math.max(.62, .86 - current.scroll * .15)")
  })
  it('keeps secondary section atmosphere restrained rather than a separate color wall', () => {
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-capabilities::before')))
      .toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-capabilities::before")))
      .toBeLessThanOrEqual(.055)
  })
})
