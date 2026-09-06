import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')
const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')

const mainSource = read('main.ts')
const insetSystem = read('styles/smirel-inset-system-v1.css')
const cardSystem = read('styles/smirel-card-system-v1.css')

describe('Smirel nested surface inset contract', () => {
  it('loads the geometry contract after the canonical card material', () => {
    const cardImport = "import './styles/smirel-card-system-v1.css'"
    const insetImport = "import './styles/smirel-inset-system-v1.css'"

    expect(mainSource).toContain(cardImport)
    expect(mainSource).toContain(insetImport)
    expect(mainSource.indexOf(cardImport)).toBeLessThan(mainSource.indexOf(insetImport))
    expect(insetSystem).toContain('--smirel-inset-space: 16px')
    expect(insetSystem).toContain('--smirel-inset-gap: 12px')
  })

  it('keeps every Overview child card away from its parent edge', () => {
    expect(insetSystem).toContain('.sw2-admin-health-grid')
    expect(insetSystem).toContain('.sw2-admin-health-foot')
    expect(insetSystem).toContain('.sw2-admin-today-foot')
    expect(insetSystem).toContain('.sw2-admin-hour-shell')
    expect(insetSystem).toContain('.sw2-admin-hour-grid')
    expect(insetSystem).toContain('.sw2-admin-hour-state')
    expect(insetSystem).toContain('.sw2-admin-diagnostics')
    expect(insetSystem).toContain('margin: 0 var(--smirel-inset-space) var(--smirel-inset-space) !important')
    expect(insetSystem).toContain('margin: var(--smirel-inset-space) !important')
    expect(insetSystem).toContain('gap: var(--smirel-inset-gap) !important')

    expect(cardSystem).toContain('.sw2-admin-health-foot')
    expect(cardSystem).toContain('.sw2-admin-today-foot')
    expect(cardSystem).toContain('.sw2-admin-hour-state')
  })

  it('insets routed filter and selection control groups inside their outer cards', () => {
    expect(insetSystem).toContain('.sw2-route-content .smg-page-toolbar')
    expect(insetSystem).toContain('.smg-data-panel .account-bulk-actions')
    expect(insetSystem).toContain('.smg-data-panel .smg-data-pagination')
    expect(cardSystem).toContain('.sw2-route-content .account-bulk-actions')
    expect(cardSystem).toContain('.sw2-route-content .smg-data-pagination')
  })
})
