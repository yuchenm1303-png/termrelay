import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const stylesheet = postcss.parse(read('src/smirel/styles/workspace-ambient.css'))
const legacyLight = read('src/smirel/styles/workspace-light-tail.css')
const accountStyles = read('src/smirel/styles/admin-accounts-polish.css')
const shell = read('src/smirel/components/WorkspaceShell.vue')
const entry = read('src/main.ts')
const plane = '.workspace-root .workspace-main > .workspace-atmosphere'

function ruleFor(selector: string, context?: string): Rule {
  let found: Rule | undefined
  stylesheet.walkRules((rule) => {
    if (!postcss.list.comma(rule.selector).some(part => part.trim() === selector)) return
    if (context && (rule.parent?.type !== 'atrule' || rule.parent.params !== context)) return
    if (!context && rule.parent?.type === 'atrule') return
    found = rule
  })
  expect(found, `Missing ${selector} in ${context || 'base'}`).toBeDefined()
  return found!
}
function declaration(selector: string, name: string, context?: string): Declaration {
  const found = ruleFor(selector, context).nodes.filter(
    (node): node is Declaration => node.type === 'decl' && node.prop === name,
  ).at(-1)
  expect(found, `Missing ${name} for ${selector}`).toBeDefined()
  return found!
}

describe('single-plane workspace aurora', () => {
  it('uses one shared layer, not a stack of rectangular route backgrounds', () => {
    expect(shell.match(/class="workspace-atmosphere"/g)).toHaveLength(1)
    expect(shell.indexOf('class="workspace-atmosphere"'))
      .toBeLessThan(shell.indexOf('class="workspace-canvas"'))
    expect(entry).toContain("import './smirel/styles/workspace-ambient.css'")
    expect(legacyLight).toContain('.workspace-root .workspace-canvas {')
    const main = '.workspace-root .workspace-main'
    const paper = declaration(main, 'background')
    expect(paper.value).toBe('var(--workspace-ambient-paper)')
    expect(paper.important).toBe(true)
    expect(ruleFor(main).nodes.some(node =>
      node.type === 'decl' && ['background-image','background-attachment'].includes(node.prop),
    )).toBe(false)
    expect(declaration('.workspace-root .workspace-main > .workspace-canvas', 'background').value)
      .toBe('transparent')
    expect(declaration('.workspace-root .workspace-main > .workspace-canvas', 'background').important)
      .toBe(true)
    expect(accountStyles).not.toMatch(/\.admin-accounts-page::before\s*\{/)
  })
  it('feathers the glow inside the chrome and keeps table content interactive', () => {
    expect(declaration(plane, 'position').value).toBe('fixed')
    expect(declaration(plane, 'z-index').value).toBe('0')
    expect(declaration(plane, 'pointer-events').value).toBe('none')
    expect(declaration(plane, 'inset').value).toContain('58px')
    expect(declaration(plane, 'background').value).toContain('ellipse 49% 50%')
    expect(declaration(plane, 'background').value).toContain('ellipse 45% 48%')
    expect(declaration(plane, 'mask-image').value).toContain('transparent 100%')
    expect(declaration(plane, '-webkit-mask-image').value).toContain('transparent 100%')
    const light = "html.smirel-app[data-theme='light'] .workspace-root .workspace-main"
    const dark = '.workspace-root .workspace-main'
    expect(declaration(light, '--workspace-ambient-paper').value).toBe('#f8faff')
    expect(declaration(dark, '--workspace-ambient-paper').value).toBe('#0d121b')
    for (const token of ['blue', 'pink', 'violet', 'mint']) {
      const color = declaration(light, `--workspace-ambient-${token}`).value
      const stop = color.match(/rgba\([^)]*,\s*([.\d]+)\)/)
      expect(stop, token).not.toBeNull()
      expect(Number(stop![1]), token).toBeLessThanOrEqual(.12)
    }
  })
  it('runs counter-moving clouds with responsive and reduced-motion behavior', () => {
    const first = declaration(`${plane}::before`, 'background').value
    const second = declaration(`${plane}::after`, 'background').value
    expect(first).toContain('ellipse 35% 50%')
    expect(second).toContain('ellipse 37% 43%')
    expect(declaration(`${plane}::before`, 'animation').value).toContain('workspace-ambient-breathe')
    expect(declaration(`${plane}::after`, 'animation').value).toContain('workspace-ambient-counterflow')
    const names: string[] = []
    stylesheet.walkAtRules('keyframes', rule => names.push(rule.params))
    expect(names).toContain('workspace-ambient-breathe')
    expect(names).toContain('workspace-ambient-counterflow')
    expect(declaration(plane, 'left', '(max-width: 1279px), (max-width: 1366px) and (pointer: coarse)').value).toBe('0')
    expect(declaration(`${plane}::after`, 'display', '(max-width: 720px)').value).toBe('none')
    const reduced = '(prefers-reduced-motion: reduce)'
    for (const pseudo of ['::before', '::after']) {
      const animation = declaration(`${plane}${pseudo}`, 'animation', reduced)
      expect(animation.value).toBe('none')
      expect(animation.important).toBe(true)
    }
  })
})
