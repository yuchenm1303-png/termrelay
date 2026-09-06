import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')
const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')

const mainSource = read('main.ts')
const runtimeSource = read('utils/smirelPageMotion.ts')
const motionCss = read('styles/smirel-shared-interactions-v1.css')

describe('Smirel page motion', () => {
  it('installs page motion without the pointer interaction runtime', () => {
    expect(mainSource).toContain("import { installSmirelPageMotion } from '@/utils/smirelPageMotion'")
    expect(mainSource).toContain("app.mount('#app')\n  installSmirelPageMotion()")
    expect(mainSource).not.toContain('SmirelDownloadInteractions')
  })

  it('keeps dynamically mounted route content on the page-motion runtime', () => {
    expect(runtimeSource).toContain('new MutationObserver')
    expect(runtimeSource).toContain("'.sw2-route-stage > *'")
    expect(runtimeSource).toContain('bindFrom(node)')
  })

  it('keeps page and modal entrance motion', () => {
    expect(motionCss).toContain('.smirel-page-enter')
    expect(motionCss).toContain('transform: scale(1.08)')
    expect(motionCss).toContain('filter: blur(7px)')
    expect(motionCss).toContain('transition: transform .32s ease, opacity .32s ease !important')
    expect(motionCss).toContain('transform: scale(1.05) !important')
  })

  it('does not ship cursor, fireworks or pointer-follow card motion', () => {
    expect(motionCss).not.toContain('.smirel-cursor-dot')
    expect(motionCss).not.toContain('.smirel-cursor-follow')
    expect(motionCss).not.toContain('.smirel-fireworks')
    expect(motionCss).not.toContain('.smirel-download-card-motion')
  })

  it('respects reduced-motion preferences', () => {
    expect(runtimeSource).toContain("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
    expect(motionCss).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
