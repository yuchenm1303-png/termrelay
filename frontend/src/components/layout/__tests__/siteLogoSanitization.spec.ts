import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const sidebarSource = readFileSync(resolve(dir, '../SmirelGlassSidebar.vue'), 'utf8')
const homeViewSource = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')
const keyUsageViewSource = readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8')

describe('site_logo sanitization', () => {
  it('Glass sidebar sanitizes siteLogo', () => {
    expect(sidebarSource).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sidebarSource).toContain("sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })")
  })

  it('HomeView sanitizes siteLogo', () => {
    expect(homeViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo')
  })

  it('KeyUsageView sanitizes siteLogo', () => {
    expect(keyUsageViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo')
  })

  it('all active surfaces allow relative and data image URLs', () => {
    for (const src of [sidebarSource, homeViewSource, keyUsageViewSource]) {
      expect(src).toContain('allowRelative: true')
      expect(src).toContain('allowDataUrl: true')
    }
  })
})
