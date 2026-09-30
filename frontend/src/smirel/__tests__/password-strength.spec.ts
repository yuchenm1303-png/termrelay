import { describe, expect, it } from 'vitest'
import { estimatePasswordStrength } from '../core/password-strength'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const component = readFileSync(resolve(process.cwd(), 'src/smirel/pages/AuthPage.vue'), 'utf8')

describe('registration password strength guidance', () => {
  it('starts neutral and reacts to increasing length and diversity', () => {
    expect(estimatePasswordStrength('').level).toBe(0)
    expect(estimatePasswordStrength('bird').level).toBe(1)
    expect(estimatePasswordStrength('midnight').level).toBe(2)
    expect(estimatePasswordStrength('midnightSun42').level).toBe(3)
    expect(estimatePasswordStrength('midnightSun42!').level).toBe(4)
    expect(estimatePasswordStrength('little violet forest morning').level).toBe(4)
  })

  it('does not call common, repeated or email-derived passwords strong', () => {
    expect(estimatePasswordStrength('Password123!').level).toBe(1)
    expect(estimatePasswordStrength('aaaaaaaaaaaaaaa').level).toBe(1)
    expect(estimatePasswordStrength('qwerty-My12345').level).toBe(1)
    expect(estimatePasswordStrength('StarPilot!2048', 'starpilot@example.com').level).toBe(1)
  })

  it('only appears on registration, is a non-interactive accessible meter, and retains the existing submit contract', () => {
    expect(component).toContain('v-if="kind === \'register\'" class="auth-strength"')
    expect(component).toContain('role="meter"')
    expect(component).toContain(':aria-valuenow="passwordStrength.level"')
    expect(component).toContain('class="auth-strength-hint"')
    expect(component).toContain(':disabled="loading || turnstilePending || Boolean(turnstileLoadError)"')
    expect(component).toContain('password.value !== confirmPassword.value')
    expect(component).not.toContain('type="range"')
  })
})
