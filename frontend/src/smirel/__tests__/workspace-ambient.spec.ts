import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss from 'postcss'

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const css = postcss.parse(read('src/smirel/styles/workspace-ambient.css'))
const shell = read('src/smirel/components/WorkspaceShell.vue')
const entry = read('src/main.ts')
const legacyLight = postcss.parse(read('src/smirel/styles/workspace-light-tail.css'))

function findRule(selector: string, context?: string) {
  let found: postcss.Rule | undefined
  css.walkRules((rule) => {
    if (!postcss.list.comma(rule.selector).some((part) => part.trim() === selector)) return
    if (context && rule.parent?.type === 'atrule' && rule.parent.params !== context) return
    if (!context && rule.parent?.type === 'atrule') return
    found = rule
  })
  expect(found, `${selector} should exist in ${context ?? 'base'}`).toBeDefined()
  return found!
}

function property(selector: string, prop: string, context?: string) {
  const rule = findRule(selector, context)
  const value = rule.nodes.find((node): node is postcss.Declaration =>
    node.type === 'decl' && node.prop === prop)
  expect(value, `${selector} must define ${prop}`).toBeDefined()
  return value!
}

describe('authenticated workspace ambient layers', () => {
  const atmosphere = '.workspace-root .workspace-main > .workspace-atmosphere'
  it('lives once inside the shared shell, behind all route content', () => {
    expect(shell.match(/class="workspace-atmosphere"/g)).toHaveLength(1)
    expect(shell.indexOf('class="workspace-atmosphere"'))
      .toBeGreaterThan(shell.indexOf('class="workspace-main"'))
    expect(shell.indexOf('class="workspace-atmosphere"'))
      .toBeLessThan(shell.indexOf('class="workspace-canvas"'))
    expect(entry).toContain("import './smirel/styles/workspace-ambient.css'")
    expect(property(atmosphere, 'pointer-events').value).toBe('none')
    expect(property(atmosphere, 'position').value).toBe('fixed')
    expect(property(atmosphere, 'contain').value).toBe('paint')
    expect(property('.workspace-root .workspace-main > .workspace-canvas', 'background').value)
      .toBe('transparent')
    expect(property('.workspace-root .workspace-main > .workspace-canvas', 'background').important)
      .toBe(true)
    // Reproduce the original light-theme conflict: legacy CSS paints the
    // canvas and main opaque. Our direct main gradient wins explicitly,
    // while the actual canvas remains transparent above it.
    const legacyText = legacyLight.toString()
    expect(legacyText).toContain('.workspace-root .workspace-main,')
    expect(legacyText).toContain('.workspace-root .workspace-canvas {')
    const main = '.workspace-root .workspace-main'
    const background = property(main, 'background-image')
    expect(background.important).toBe(true)
    for (const tone of ['blue', 'violet', 'pink']) {
      expect(background.value).toContain(`var(--ambient-surface-${tone})`)
    }
    expect(property(main, 'background-attachment').value).toBe('fixed')
  })
  it('uses quiet blue violet and pink washes with two independent ambient loops', () => {
    const colors = findRule('.workspace-root .workspace-main')
    for (const token of ['--ambient-blue', '--ambient-violet', '--ambient-pink']) {
      expect(colors.nodes.some((node) => node.type === 'decl' && node.prop === token)).toBe(true)
    }
    expect(property(`${atmosphere}::before`, 'animation').value).toContain('workspace-ambient-breathe')
    expect(property(`${atmosphere}::after`, 'animation').value).toContain('workspace-ambient-counterflow')
    const names: string[] = []
    css.walkAtRules('keyframes', (rule) => names.push(rule.params))
    expect(names).toContain('workspace-ambient-breathe')
    expect(names).toContain('workspace-ambient-counterflow')
    expect(property('.workspace-root .workspace-main', '--ambient-blue').value)
      .toBe('rgba(80, 165, 242, .31)')
    const dark = "html.smirel-app[data-theme='dark'] .workspace-root .workspace-main"
    expect(property(dark, '--ambient-blue').value).toBe('rgba(45, 133, 222, .27)')
    expect(property(dark, '--ambient-surface-pink').value)
      .toBe('rgba(194, 91, 170, .13)')
  })
  it('keeps mobile gutters safe and removes animation for reduced-motion users', () => {
    const mobile = '(max-width: 1279px), (max-width: 1366px) and (pointer: coarse)'
    expect(property(atmosphere, 'left', mobile).value).toBe('0')
    expect(property('.workspace-root .workspace-main', 'background-attachment', '(max-width: 720px)').value).toBe('scroll')
    expect(property(`${atmosphere}::after`, 'display', '(max-width: 720px)').value).toBe('none')
    const reduced = '(prefers-reduced-motion: reduce)'
    for (const ending of ['::before', '::after']) {
      const selector = '.workspace-root .workspace-main > .workspace-atmosphere' + ending
      expect(property(selector, 'animation', reduced).value).toBe('none')
      expect(property(selector, 'animation', reduced).important).toBe(true)
    }
  })
})
