import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')

const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')
const block = (source: string, selector: string) => {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return source.match(new RegExp(`${escaped}\\s*\\{[^}]*\\}`, 's'))?.[0] || ''
}

const mainSource = read('main.ts')
const primitiveSource = read('components/glass/GlassSurface.vue')
const workspaceSource = read('views/SmirelWorkspaceV2.vue')
const cardSystem = read('styles/smirel-card-system-v1.css')
const navigationSystem = read('styles/smirel-navigation-v1.css')
const sharedMaterial = read('styles/smirel-shared-material-v1.css')
const workspaceLayout = read('styles/smirel-secondary-v2.css')
const workspaceFunctional = read('styles/smirel-workspace-functional-v2.css')
const authLayout = read('components/layout/AuthLayout.vue')
const authCss = read('styles/smirel-glass-auth-v5.css')
const dashboardCss = read('styles/smirel-glass-dashboard-v5.css')

const cardMaterialDeclaration = /(background\s*:|border-radius\s*:|backdrop-filter\s*:|box-shadow\s*:)/

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

  it('bridges remaining SPG and SMG card tokens into the canonical material', () => {
    expect(cardSystem).toContain('--spg-radius: var(--smirel-card-radius)')
    expect(cardSystem).toContain('--spg-surface: var(--smirel-card-bg)')
    expect(cardSystem).toContain('--spg-inset: var(--smirel-card-bg-data)')
    expect(cardSystem).toContain('--smg-radius-xl: var(--smirel-card-radius)')
    expect(cardSystem).toContain('--smg-radius-lg: var(--smirel-card-radius)')
    expect(cardSystem).toContain('--smg-glass: var(--smirel-card-bg)')
    expect(cardSystem).toContain('--smg-glass-quiet: var(--smirel-card-bg-quiet)')
    expect(cardSystem).toContain('--smg-data: var(--smirel-card-bg-data)')
  })

  it('normalizes first-level legacy surfaces that can render outside the V2 workspace wrapper', () => {
    expect(cardSystem).toContain('.spg-page .spg-surface')
    expect(cardSystem).toContain('.smg-shell .smg-sidebar')
    expect(cardSystem).toContain('.smg-shell .smg-topbar')
    expect(cardSystem).toContain('.smg-shell .smg-sidebar-context')
    expect(cardSystem).toContain('.smg-shell .smg-profile-support')
    expect(cardSystem).toContain('.smg-shell .smg-catalog-panel')
    expect(cardSystem).toContain('.smg-shell .smg-page-toolbar')
    expect(cardSystem).toContain('.smg-shell .smg-data-panel')
    expect(cardSystem).toContain('.spg-page .spg-surface--interactive')
  })

  it('keeps workspace cards canonical while navigation and account rows stay controls', () => {
    expect(workspaceSource).toContain('sw2-sidebar spg-surface smirel-card')
    expect(workspaceSource).toContain('sw2-sidebar-context smirel-card smirel-card--quiet')
    expect(workspaceSource).toContain('sw2-side-item smirel-nav-item')
    expect(workspaceSource).toContain("'smirel-nav-item--active': isActive(item.to)")
    expect(workspaceSource).not.toContain('sw2-side-item smirel-card')
    expect(workspaceSource).toContain('class="sw2-account-entry"')
    expect(workspaceSource).not.toContain('sw2-account-card smirel-card')
    expect(workspaceSource).toContain('sw2-topbar smirel-card')
    expect(workspaceSource).toContain("import '@/styles/smirel-navigation-v1.css'")

    expect(cardSystem).not.toContain('.sw2-side-item')
    expect(cardSystem).not.toContain('.sw2-account-card')
    expect(cardSystem).not.toContain('.sw2-account-entry')
    expect(navigationSystem).toContain('--smirel-nav-radius: 10px')
    expect(navigationSystem).toContain('.smirel-nav-item')
    expect(navigationSystem).toContain('.smirel-nav-item--active')
    expect(navigationSystem).not.toContain('backdrop-filter')
    expect(navigationSystem).not.toContain('box-shadow')

    expect(workspaceLayout).not.toContain('--sw2-radius')
    expect(workspaceLayout).not.toContain('--sw2-panel:')
    expect(workspaceLayout).not.toContain('--sw2-sidebar-bg')
    expect(workspaceLayout).not.toContain('--sw2-topbar-bg')
    expect(workspaceLayout).not.toContain('border-radius: 0')
    expect(block(workspaceLayout, '.sw2-sidebar-context')).not.toMatch(cardMaterialDeclaration)
    expect(block(workspaceLayout, '.sw2-side-item')).not.toMatch(cardMaterialDeclaration)
    expect(block(workspaceLayout, '.sw2-topbar')).not.toMatch(cardMaterialDeclaration)
    expect(workspaceFunctional).not.toContain('linear-gradient(145deg, rgba(158, 216, 250')
    expect(workspaceFunctional).not.toContain('linear-gradient(100deg, rgba(161, 219, 253')
  })

  it('keeps the workspace footer compact and gives both utility actions one control contract', () => {
    const buttonBase = block(workspaceLayout, '.sw2-page button')
    const accountEntry = block(workspaceFunctional, '.sw2-account-entry')
    const footerActions = block(workspaceFunctional, '.sw2-footer-actions')
    const footerAction = block(workspaceFunctional, '.sw2-footer-action')

    expect(buttonBase).toContain('font-family: inherit')
    expect(buttonBase).not.toContain('font: inherit')
    expect(accountEntry).toContain('min-height: 44px')
    expect(accountEntry).toContain('background: transparent')
    expect(footerActions).toContain('grid-template-columns: repeat(2, minmax(0, 1fr))')
    expect(footerAction).toContain('min-height: 30px')
    expect(footerAction).toContain('font-size: .54rem')
    expect(footerAction).toContain('justify-content: center')
    expect(workspaceSource).toContain('class="sw2-footer-action sw2-home-link"')
    expect(workspaceSource).toContain('class="sw2-footer-action sw2-logout-button"')
    expect(workspaceFunctional).not.toContain('.sw2-page .sw2-home-link,')
    expect(workspaceSource).toContain("if (isSmirelUiPreview) return '预览管理员'")
    expect(workspaceSource).toContain("if (isSmirelUiPreview) return '界面预览 · 非真实账号'")
  })

  it('keeps shared material focused on the scene rather than defining a second card recipe', () => {
    expect(sharedMaterial).not.toContain('--spg-surface:')
    expect(sharedMaterial).not.toContain('--spg-radius:')
    expect(sharedMaterial).not.toContain('border-radius: 0')
    expect(sharedMaterial).toContain('Card/surface material moved to smirel-card-system-v1.css')
  })

  it('composes authentication cards through GlassSurface and keeps their page css material-free', () => {
    expect(authLayout).toContain('<GlassSurface as="header" class="smg-auth-nav">')
    expect(authLayout).toContain('tone="data"')
    expect(block(authCss, '.smg-auth-fact')).not.toMatch(cardMaterialDeclaration)
    expect(block(authCss, '.smg-auth-code')).not.toMatch(cardMaterialDeclaration)
  })

  it('routes dashboard nested data cards through the centralized compatibility adapter', () => {
    expect(cardSystem).toContain('.smg-dashboard-code')
    expect(cardSystem).toContain('.smg-dashboard-chart-surface')
    expect(block(dashboardCss, '.smg-dashboard-code')).not.toMatch(cardMaterialDeclaration)
    expect(block(dashboardCss, '.smg-dashboard-chart-surface')).not.toMatch(cardMaterialDeclaration)
  })

  it('keeps homepage, admin and legacy route aliases pinned to the same material during migration', () => {
    expect(cardSystem).toContain('.spg-page.smh-home .smh-glass')
    expect(cardSystem).toContain('.sw2-admin-module')
    expect(cardSystem).toContain('.sw2-route-content .smg-surface')
    expect(cardSystem).toContain('.sw2-route-content .card')
  })
})
