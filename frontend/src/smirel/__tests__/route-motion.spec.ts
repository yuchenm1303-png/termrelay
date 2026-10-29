import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { resolveRouteMotion } from '../core/route-motion'

const at = (path: string, feature: string) => ({ path, feature })

describe('workspace route motion', () => {
  it('slides laterally between pages of the same sidebar section', () => {
    expect(resolveRouteMotion(at('/orders', 'orders'), at('/subscriptions', 'subscriptions')))
      .toEqual({ kind: 'sibling', direction: 1 })
    expect(resolveRouteMotion(at('/subscriptions', 'subscriptions'), at('/orders', 'orders')))
      .toEqual({ kind: 'sibling', direction: -1 })
  })

  it('follows sidebar order when crossing sections', () => {
    expect(resolveRouteMotion(at('/subscriptions', 'subscriptions'), at('/dashboard', 'dashboard')))
      .toEqual({ kind: 'section', direction: 1 })
    expect(resolveRouteMotion(at('/usage', 'usage'), at('/support', 'support')))
      .toEqual({ kind: 'section', direction: -1 })
  })

  it('treats user/admin switches as a depth change and neutral routes as same-side', () => {
    expect(resolveRouteMotion(at('/admin/dashboard', 'admin-dashboard'), at('/dashboard', 'dashboard')).kind)
      .toBe('workspace')
    expect(resolveRouteMotion(at('/keys', 'keys'), at('/admin/dashboard', 'admin-dashboard')))
      .toEqual({ kind: 'sibling', direction: 1 })
  })

  it('falls back to a forward section move for unknown or first routes', () => {
    expect(resolveRouteMotion(at('/custom/1', 'custom'), at('/dashboard', 'dashboard')))
      .toEqual({ kind: 'section', direction: 1 })
    expect(resolveRouteMotion(at('/dashboard', 'dashboard'), null)).toEqual({ kind: 'section', direction: 1 })
  })

  it('keeps the light theme clarity override off transitioning stages', () => {
    const css = readFileSync(resolve(process.cwd(), 'src/smirel/styles/workspace-light-tail.css'), 'utf8')
    expect(css).toContain('.workspace-route-stage:not(.workspace-route-enter-active):not(.workspace-route-leave-active)')
    const route = readFileSync(resolve(process.cwd(), 'src/smirel/styles/route-transitions.css'), 'utf8')
    expect(route).not.toMatch(/clip-path:\s*inset/)
  })
})
