import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../TablePageLayout.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('TablePageLayout workspace composition', () => {
  it('uses one canonical workspace DOM for admin and user routes', () => {
    expect(componentSource).toContain('class="smg-page"')
    expect(componentSource).toContain('class="smg-data-scroll"')
    expect(componentSource).not.toContain('v-if="authStore.isAdmin"')
    expect(componentSource).not.toContain('smv3-')
  })

  it('keeps the table inside the dedicated scroll container', () => {
    expect(componentSource).toContain('<div class="smg-data-scroll"><slot name="table" /></div>')
  })
})
