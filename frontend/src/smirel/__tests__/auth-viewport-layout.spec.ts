import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const authSource = readFileSync(resolve(process.cwd(), 'src/smirel/pages/AuthPage.vue'), 'utf8')
const scopedStyles = authSource.split('<style scoped>')[1]?.split('</style>')[0] ?? ''

describe('auth viewport layout contract', () => {
  it('resets the legacy page padding and card margin so auth is not a scrolling landing page', () => {
    const page = scopedStyles.match(/\.auth-page\s*\{([^}]+)\}/)?.[1] ?? ''
    const card = scopedStyles.match(/\.auth-card\s*\{([^}]+)\}/)?.[1] ?? ''
    expect(page).toContain('height: 100dvh;')
    expect(page).toContain('padding: 0;')
    expect(card).toMatch(/margin:\s*0;/)
  })

  it('contains decoration within the page instead of generating a phantom scrollbar', () => {
    const ambientLayer = [...scopedStyles.matchAll(/^\.auth-page::after\s*\{([^}]+)\}/gm)].at(-1)?.[1] ?? ''
    expect(ambientLayer).toMatch(/inset:\s*0;/)
    expect(ambientLayer).not.toMatch(/(?:rotate\(|right:\s*-|top:\s*\d+%|width:\s*min\()/)
    expect(scopedStyles).not.toMatch(/\.auth-page::after\s*\{\s*width:\s*90vw/)
    // Keep normal overflow available for browser zoom, dynamic errors and Turnstile.
    expect(scopedStyles).toContain('overflow-y: auto;')
  })

  it('keeps compact laptop spacing without cropping errors or security checks', () => {
    const compact = scopedStyles.split('@media (max-height: 950px) {')[1]?.split('@media (max-width: 920px)')[0] ?? ''
    expect(compact).toContain('.auth-card form.with-oauth')
    expect(compact).toContain('.auth-layout')
    expect(scopedStyles).toContain('overflow-y: auto;')
    expect(authSource).toContain('v-if="needsTurnstile"')
    expect(authSource).toContain('class="auth-submit"')
  })
})
