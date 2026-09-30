import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const css = postcss.parse(read('src/smirel/styles/workspace-ambient.css'))
const shell = read('src/smirel/components/WorkspaceShell.vue')
const entry = read('src/main.ts')
const accounts = read('src/smirel/styles/admin-accounts-polish.css')
const lightTail = read('src/smirel/styles/workspace-light-tail.css')
const plane = '.workspace-root .workspace-main > .workspace-atmosphere'

function rule(selector: string, media?: string): Rule {
  let found: Rule | undefined
  css.walkRules((node) => {
    if (!postcss.list.comma(node.selector).some(part => part.trim() === selector)) return
    if (media && (node.parent?.type !== 'atrule' || node.parent.params !== media)) return
    if (!media && node.parent?.type === 'atrule') return
    found = node
  })
  expect(found, `Missing ${selector} / ${media || 'base'}`).toBeDefined()
  return found!
}
function decl(selector: string, name: string, media?: string): Declaration {
  const node = rule(selector, media).nodes
    .filter((item): item is Declaration => item.type === 'decl' && item.prop === name)
    .at(-1)
  expect(node, `Missing ${selector} ${name}`).toBeDefined()
  return node!
}

describe('full-shell shared aurora', () => {
  const root = '.workspace-root'
  const main = '.workspace-root .workspace-main'
  const canvas = '.workspace-root .workspace-main > .workspace-canvas'
  const light = "html.smirel-app[data-theme='light']"
  const dark = "html.smirel-app[data-theme='dark']"

  it('uses one whole-viewport atmosphere with no colored seam at the sidebar or header', () => {
    expect(shell.match(/class="workspace-atmosphere"/g)).toHaveLength(1)
    expect(shell.indexOf('class="workspace-atmosphere"'))
      .toBeLessThan(shell.indexOf('class="workspace-canvas"'))
    expect(entry).toContain("import './smirel/styles/workspace-ambient.css'")
    expect(decl(plane, 'inset').value).toBe('0')
    expect(decl(plane, 'position').value).toBe('fixed')
    expect(decl(plane, 'pointer-events').value).toBe('none')
    expect(decl(plane, 'mask-image').value).toContain('transparent 100%')
    expect(decl(plane, '-webkit-mask-image').value).toContain('transparent 100%')
    expect(decl(root, 'isolation').value).toBe('isolate')
    expect(decl(root, 'width').value).toBe('100%')
    expect(decl(root, 'background').value).toBe('var(--workspace-shell-paper)')
    expect(decl(root, 'background').important).toBe(true)
    expect(decl(main, 'background').value).toBe('transparent')
    expect(decl(main, 'background').important).toBe(true)
    expect(decl(canvas, 'background').value).toBe('transparent')
    expect(decl(canvas, 'background').important).toBe(true)
    expect(lightTail).toContain('.workspace-root .workspace-main,')
    expect(accounts).not.toMatch(/\.admin-accounts-page::before\s*\{/)
    // The ambient palette is inherited throughout the entire workspace.
    expect(decl(`${light} ${root}`, '--workspace-shell-paper').value).toBe('#f8faff')
    expect(decl(root, '--workspace-shell-paper').value).toBe('#0d121b')
    expect(decl(`${light}:has(.workspace-root)`, 'background').value).toBe('#f8faff')
  })

  it('lets sidebar and topbar share the field without harsh borders or isolated whites', () => {
    for (const theme of [light, dark]) {
      const side = `${theme} ${root} .workspace-sidebar`
      const top = `${theme} ${root} .workspace-main .workspace-topbar.workspace-topbar--contextual`
      const sidebarPaint = decl(side, 'background')
      const topbarPaint = decl(top, 'background')
      expect(sidebarPaint.value).toMatch(/^rgba\(/)
      expect(topbarPaint.value).toMatch(/^rgba\(/)
      expect(sidebarPaint.important).toBe(true)
      expect(topbarPaint.important).toBe(true)
      expect(decl(side, 'border-right').value).toMatch(/rgba\(/)
      expect(decl(side, 'box-shadow').value).toBe('none')
      expect(decl(top, 'box-shadow').value).toBe('none')
    }
    expect(decl(`${light} ${root} .workspace-sidebar`, 'background').value)
      .toBe('rgba(251, 252, 255, .73)')
    expect(decl(`${light} ${root} .workspace-main .workspace-topbar.workspace-topbar--contextual`, 'background').value)
      .toBe('rgba(251, 252, 255, .82)')
    expect(decl(`${light} body:has(.workspace-root)`, 'background').value).toBe('#f8faff')
    expect(decl(`${dark} body:has(.workspace-root)`, 'background').value).toBe('#0d121b')
  })

  it('places the original logo on the sidebar without a separate white brand tile', () => {
    const row = `${root} .workspace-sidebar .workspace-brand-row`
    const link = `${row} .brand-link`
    expect(decl(row, 'background').value).toBe('transparent')
    expect(decl(row, 'background').important).toBe(true)
    expect(decl(row, 'box-shadow').value).toBe('none')
    for (const state of ['', ':hover', ':active']) {
      expect(decl(link + state, 'background').value).toBe('transparent')
      expect(decl(link + state, 'background').important).toBe(true)
      expect(decl(link + state, 'border-color').value).toBe('transparent')
      expect(decl(link + state, 'box-shadow').value).toBe('none')
    }
    expect(decl(link + ':focus-visible', 'outline').value).toContain('2px solid')
    expect(shell).toContain('class="workspace-brand-mark"')
    expect(shell).toContain('<img :src="logoUrl" alt=""')
  })

  it('keeps the original drift, neutral center, and safe responsive drawer', () => {
    expect(decl(plane, 'background').value).toContain('--workspace-ambient-base-blue')
    expect(decl(plane, 'background').value).toContain('--workspace-ambient-base-pink')
    expect(decl(`${plane}::before`, 'animation').value).toContain('workspace-ambient-breathe')
    expect(decl(`${plane}::after`, 'animation').value).toContain('workspace-ambient-counterflow')
    const names: string[] = []
    css.walkAtRules('keyframes', atRule => names.push(atRule.params))
    expect(names).toContain('workspace-ambient-breathe')
    expect(names).toContain('workspace-ambient-counterflow')
    const mobile = '(max-width: 1279px), (max-width: 1366px) and (pointer: coarse)'
    expect(decl(`${light} ${root} .workspace-sidebar`, 'background', mobile).value)
      .toBe('rgba(250, 252, 255, .97)')
    expect(decl(`${dark} ${root} .workspace-sidebar`, 'background', mobile).value)
      .toBe('rgba(13, 20, 32, .97)')
    expect(decl(`${plane}::after`, 'display', '(max-width: 720px)').value).toBe('none')
    for (const part of ['::before', '::after']) {
      const animation = decl(`${plane}${part}`, 'animation', '(prefers-reduced-motion: reduce)')
      expect(animation.value).toBe('none')
      expect(animation.important).toBe(true)
    }
  })
})
