import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss from 'postcss'

const component = readFileSync(resolve(process.cwd(), 'src/smirel/pages/AuthPage.vue'), 'utf8')
const styles = postcss.parse(component.split('<style scoped>')[1]?.split('</style>')[0] || '')

describe('Auth submit direction icon', () => {
  it('uses a decorative vector icon instead of the old font arrow', () => {
    expect(component).not.toContain('<b aria-hidden="true">→</b>')
    expect(component).toContain('<span class="auth-submit-direction" aria-hidden="true">')
    expect(component).toContain('viewBox="0 0 24 24" fill="none" focusable="false"')
    expect(component).toContain('class="auth-submit-label"')
  })
  it('retains the busy indicator, centered label and reduced-motion alternative', () => {
    expect(component).toContain('v-if="loading" class="auth-submit-spinner"')
    expect(component).toContain(':disabled="loading || turnstilePending || Boolean(turnstileLoadError)"')
    let centered = false
    let spinner = false
    let reducedMotion = false
    styles.walkRules((rule) => {
      if (rule.selector === '.auth-submit-label') centered = rule.nodes.some((node) => node.type === 'decl' && node.prop === 'text-align' && node.value === 'center')
      if (rule.selector === '.auth-submit-spinner') spinner = true
    })
    styles.walkAtRules('media', (media) => {
      if (media.params === '(prefers-reduced-motion: reduce)' && media.toString().includes('.auth-submit-spinner')) reducedMotion = true
    })
    expect(centered).toBe(true)
    expect(spinner).toBe(true)
    expect(reducedMotion).toBe(true)
  })
})
