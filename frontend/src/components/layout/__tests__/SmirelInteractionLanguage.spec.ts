import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')
const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')

const mainSource = read('main.ts')
const pointerRuntimeSource = read('utils/smirelDownloadInteractions.ts')
const pageRuntimeSource = read('utils/smirelPageMotion.ts')
const motionCss = read('styles/smirel-shared-interactions-v1.css')

describe('Smirel interaction language', () => {
  it('installs card/control and page motion after Vue mounts', () => {
    expect(mainSource).toContain("import { installSmirelDownloadInteractions } from '@/utils/smirelDownloadInteractions'")
    expect(mainSource).toContain("import { installSmirelPageMotion } from '@/utils/smirelPageMotion'")
    expect(mainSource).toContain("app.mount('#app')\n  installSmirelDownloadInteractions()\n  installSmirelPageMotion()")
  })

  it('keeps the card pointer and touch motion amplitudes', () => {
    expect(pointerRuntimeSource).toContain('const hoverScale = utility ? 1.02 : 1.01')
    expect(pointerRuntimeSource).toContain("setCardMotion(card, 0, 0, 0.98, '90ms', 'ease-out')")
    expect(pointerRuntimeSource).toContain('setCardMotion(card, nx * 5, ny * 4, hoverScale')
    expect(pointerRuntimeSource).toContain('setCardMotion(card, nx * 3.2, ny * 2.8, 0.985')
    expect(pointerRuntimeSource).toContain("setCardMotion(card, 0, 0, hoverScale, '95ms', 'ease-out')")
    expect(motionCss).toContain('.smirel-download-card-motion')
    expect(motionCss).toContain('.smirel-download-control-motion')
  })

  it('keeps dynamically mounted content on both motion runtimes', () => {
    expect(pointerRuntimeSource).toContain('new MutationObserver')
    expect(pointerRuntimeSource).toContain('bindCards(node)')
    expect(pointerRuntimeSource).toContain('bindControls(node)')
    expect(pageRuntimeSource).toContain('new MutationObserver')
    expect(pageRuntimeSource).toContain("'.sw2-route-stage > *'")
    expect(pageRuntimeSource).toContain('bindFrom(node)')
  })

  it('keeps page and modal entrance motion', () => {
    expect(motionCss).toContain('.smirel-page-enter')
    expect(motionCss).toContain('transform: scale(1.08)')
    expect(motionCss).toContain('filter: blur(7px)')
    expect(motionCss).toContain('transition: transform .32s ease, opacity .32s ease !important')
    expect(motionCss).toContain('transform: scale(1.05) !important')
  })

  it('does not ship a custom cursor or click-firework runtime', () => {
    expect(pointerRuntimeSource).not.toContain('installCursor')
    expect(pointerRuntimeSource).not.toContain('installFireworks')
    expect(pointerRuntimeSource).not.toContain('smirel-cursor-dot')
    expect(pointerRuntimeSource).not.toContain('smirel-cursor-follow')
    expect(pointerRuntimeSource).not.toContain('smirel-fireworks')
    expect(pointerRuntimeSource).not.toContain("document.documentElement.classList.add('smirel-custom-cursor')")
    expect(motionCss).not.toContain('.smirel-cursor-dot')
    expect(motionCss).not.toContain('.smirel-cursor-follow')
    expect(motionCss).not.toContain('.smirel-fireworks')
    expect(motionCss).not.toContain('cursor: none')
  })

  it('respects reduced-motion preferences', () => {
    expect(pointerRuntimeSource).toContain("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
    expect(pageRuntimeSource).toContain("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
    expect(motionCss).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
