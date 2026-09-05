import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const topbarSource = readFileSync(resolve(dir, '../SmirelGlassTopbar.vue'), 'utf8')
const homeViewSource = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')
const keyUsageViewSource = readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8')

describe('doc_url sanitization', () => {
  it('Glass topbar sanitizes the configured docs URL', () => {
    expect(topbarSource).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(topbarSource).toContain("sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')")
  })

  it('HomeView sanitizes the configured docs URL', () => {
    expect(homeViewSource).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(homeViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })

  it('KeyUsageView sanitizes the configured docs URL', () => {
    expect(keyUsageViewSource).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(keyUsageViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })
})
