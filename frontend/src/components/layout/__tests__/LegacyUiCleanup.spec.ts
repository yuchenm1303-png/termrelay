import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')
const mainSource = readFileSync(resolve(srcRoot, 'main.ts'), 'utf8')
const glassCore = readFileSync(resolve(srcRoot, 'styles/smirel-glass-v5.css'), 'utf8')
const layoutIndex = readFileSync(resolve(dir, '../index.ts'), 'utf8')

describe('retired Smirel UI generations', () => {
  it('does not load retired Relay or V3 style layers', () => {
    expect(mainSource).not.toContain('smirel-relay')
    expect(mainSource).not.toContain('smirel-console-shell')
    expect(mainSource).not.toContain('smirel-v3')
  })

  it('keeps the canonical Glass core free of V3 compatibility selectors and tokens', () => {
    expect(glassCore).not.toContain('smv3-')
    expect(glassCore).not.toContain('--smv3-')
  })

  it('exports only active layout components', () => {
    expect(layoutIndex).not.toContain('AppHeader')
    expect(layoutIndex).not.toContain('AppSidebar')
    expect(layoutIndex).toContain('SmirelGlassSidebar')
    expect(layoutIndex).toContain('SmirelGlassTopbar')
  })
})
