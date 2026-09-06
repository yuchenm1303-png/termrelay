import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')

const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')

const mainSource = read('main.ts')
const primitiveSource = read('components/glass/GlassSurface.vue')
const workspaceSource = read('views/SmirelWorkspaceV2.vue')
const cardSystem = read('styles/smirel-card-system-v1.css')
const sharedMaterial = read('styles/smirel-shared-material-v1.css')
const workspaceLayout = read('styles/smirel-secondary-v2.css')
const workspaceFunctional = read('styles/smirel-workspace-functional-v2.css')
const authLayout = read('components/layout/AuthLayout.vue')
const authCss = read('styles/smirel-glass-auth-v5.css')
const dashboardCss = read('styles/smirel-glass-dashboard-v5.css')

describe('Smirel canonical card system', () => {
  it('loads one global card material and exposes it through GlassSurface', () => {
    expect(mainSource).toContain("import './styles/smirel-card-system-v1.css'")
    expect(primitiveSource).toContain('class="smirel-card smg-surface"')
    expect(primitiveSource).toContain('`smirel-card--${tone}`')
    expect(cardSystem).toContain('--smirel-card-radius: 14px')
    expect(cardSystem).toContain('.smirel-card--data')
    expect(cardSystem).toContain('.smirel-card--quiet')
    expect(cardSystem).toContain('.smirel-card--active')
  })

  it('makes workspace shell cards call the canonical primitive instead of drawing material locally', () => {
    expect(workspaceSource).toContain('sw2-sidebar spg-surface smirel-card')
    expect(workspaceSource).toContain('sw2-sidebar-context smirel-card smirel-card--quiet')
    expect(workspaceSource).toContain('sw2-side-item smirel-card smirel-card--quiet smirel-card--interactive')
    expect(workspaceSource).toContain('sw2-account-card smirel-card smirel-card--quiet smirel-card--interactive')
    expect(workspaceSource).toContain('sw2-topbar smirel-card')

    expect(workspaceLayout).not.toContain('--sw2-radius')
    expect(workspaceLayout).not.toContain('--sw2-panel:')
    expect(workspaceLayout).not.toContain('--sw2-sidebar-bg')
    expect(workspaceLayout).not.toContain('--sw2-topbar-bg')
    expect(workspaceLayout).not.toContain('border-radius: 0')
    expect(workspaceFunctional).not.toContain('linear-gradient(145deg, rgba(158, 216, 250')
    expect(workspaceFunctional).not.toContain('linear-gradient(100deg, rgba(161, 219, 253')
  })

  it('keeps shared material focused on the scene rather than defining a second card recipe', () => {
    expect(sharedMaterial).not.toContain('--spg-surface:')
    expect(sharedMaterial).not.toContain('--spg-radius:')
    expect(sharedMaterial).not.toContain('border-radius: 0')
    expect(sharedMaterial).toContain('Card/surface material moved to smirel-card-system-v1.css')
  })

  it('composes authentication cards through GlassSurface and keeps page css material-free', () => {
    expect(authLayout).toContain('<GlassSurface as="header" class="smg-auth-nav">')
    expect(authLayout).toContain('tone="data"')
    expect(authCss).not.toContain('background: var(--smg-data-strong)')
    expect(authCss).not.toContain('background: var(--smg-data);')
  })

  it('routes dashboard nested data cards through the centralized compatibility adapter', () => {
    expect(cardSystem).toContain('.smg-dashboard-code')
    expect(cardSystem).toContain('.smg-dashboard-chart-surface')
    expect(dashboardCss).not.toContain('background: var(--smg-data-strong)')
  })

  it('keeps homepage, admin and legacy route aliases pinned to the same material during migration', () => {
    expect(cardSystem).toContain('.spg-page.smh-home .smh-glass')
    expect(cardSystem).toContain('.sw2-admin-module')
    expect(cardSystem).toContain('.sw2-route-content .smg-surface')
    expect(cardSystem).toContain('.sw2-route-content .card')
  })
})
