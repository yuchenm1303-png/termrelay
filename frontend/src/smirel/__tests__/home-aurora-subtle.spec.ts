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
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero'))).toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-hero'))).toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-hero")))
      .toBeLessThanOrEqual(.076)
  })
  it('preserves the original animated cloud elements while reducing their saturation', () => {
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::before'))).toBeLessThanOrEqual(.145)
    expect(maxCloudAlpha(gradient(aurora, '.home-page .home-hero::after'))).toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::before")))
      .toBeLessThanOrEqual(.19)
    expect(maxCloudAlpha(gradient(aurora, "html.smirel-app[data-theme='dark'] .home-page .home-hero::after")))
      .toBeLessThanOrEqual(.13)
    expect(homepage).toContain("const opacity = Math.max(.62, .88 - current.scroll * .16)")
    expect(aurora.toString()).toContain("width: min(2160px, 150vw)")
    expect(aurora.toString()).toContain("width: min(1920px, 142vw)")
    expect(aurora.toString()).toContain("width: 174vw")
    expect(storytelling.toString()).toContain("animation-name: home-cloud-tail-idle")
  })
  it('drifts the original clouds while the pointer and scroll remain stationary', () => {
    const ambient = storytelling.toString()
    expect(ambient).toContain('animation: home-cloud-idle 13.5s')
    expect(ambient).toContain('animation-duration: 17.5s')
    const keyframes = new Map<string, postcss.AtRule>()
    storytelling.walkAtRules('keyframes', (rule) => keyframes.set(rule.params, rule))
    for (const name of ['home-cloud-idle', 'home-cloud-tail-idle']) {
      const frames = keyframes.get(name)
      expect(frames, `${name} should animate the original pseudo-elements`).toBeDefined()
      const offsets: string[] = []
      frames?.walkRules((rule) => {
        rule.walkDecls('translate', (decl) => offsets.push(decl.value))
      })
      expect(offsets).toHaveLength(2)
      expect(offsets[0]).not.toBe(offsets[1])
    }
    expect(ambient).toContain('animation: none !important')
    expect(homepage).toContain("hero.style.setProperty('--hero-aurora-x'")
    expect(homepage).toContain("hero.style.setProperty('--hero-aurora-opacity', opacity.toFixed(4))")
  })

  it('keeps secondary section atmosphere restrained rather than a separate color wall', () => {
    expect(maxCloudAlpha(gradient(storytelling, '.home-page .home-capabilities::before')))
      .toBeLessThanOrEqual(.085)
    expect(maxCloudAlpha(gradient(storytelling, "html.smirel-app[data-theme='dark'] .home-page .home-capabilities::before")))
      .toBeLessThanOrEqual(.055)
  })
})
