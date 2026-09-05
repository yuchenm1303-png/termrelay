import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const componentSource = readFileSync(resolve(dir, '../TablePageLayout.vue'), 'utf8')

describe('TablePageLayout Glass composition', () => {
  it('uses one canonical layout for admin and customer table pages', () => {
    expect(componentSource).toContain('class="smg-page"')
    expect(componentSource).toContain('<GlassSurface v-if="$slots.filters" class="smg-page-toolbar">')
    expect(componentSource).toContain('<GlassSurface class="smg-data-panel">')
    expect(componentSource).not.toContain('smv3-')
  })
})
