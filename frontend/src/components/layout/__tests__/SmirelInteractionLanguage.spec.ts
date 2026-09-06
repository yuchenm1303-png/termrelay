import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(dir, '../../..')
const read = (path: string) => readFileSync(resolve(srcRoot, path), 'utf8')

const mainSource = read('main.ts')
const runtimeSource = read('utils/smirelDownloadInteractions.ts')
const interactionCss = read('styles/smirel-shared-interactions-v1.css')

describe('Smirel download-page interaction language', () => {
  it('installs one standalone runtime after Vue mounts and loads its CSS last', () => {
    expect(mainSource).toContain("import { installSmirelDownloadInteractions } from '@/utils/smirelDownloadInteractions'")
    expect(mainSource).toContain("app.mount('#app')\n  installSmirelDownloadInteractions()")

    const interactionImport = "import './styles/smirel-shared-interactions-v1.css'"
    expect(mainSource).toContain(interactionImport)
    expect(mainSource.lastIndexOf(interactionImport)).toBeGreaterThan(mainSource.lastIndexOf("import './styles/smirel-card-system-v1.css'"))
    expect(mainSource.lastIndexOf(interactionImport)).toBeGreaterThan(mainSource.lastIndexOf("import './styles/smirel-home-topbar-v1.css'"))
  })

  it('ports the exact card pointer and touch amplitudes from the download page', () => {
    expect(runtimeSource).toContain('const hoverScale = utility ? 1.02 : 1.01')
    expect(runtimeSource).toContain("setCardMotion(card, 0, 0, 0.98, '90ms', 'ease-out')")
    expect(runtimeSource).toContain('setCardMotion(card, nx * 5, ny * 4, hoverScale')
    expect(runtimeSource).toContain('setCardMotion(card, nx * 3.2, ny * 2.8, 0.985')
    expect(runtimeSource).toContain("setCardMotion(card, 0, 0, hoverScale, '95ms', 'ease-out')")
  })

  it('arbitrates nested cards and keeps dynamically mounted route content on the same runtime', () => {
    expect(runtimeSource).toContain('findClosestMotionCard(event.target) !== card')
    expect(runtimeSource).toContain("target.closest<HTMLElement>(CARD_SELECTOR)")
    expect(runtimeSource).toContain('new MutationObserver')
    expect(runtimeSource).toContain('bindCards(node)')
    expect(runtimeSource).toContain('bindControls(node)')
    expect(runtimeSource).toContain('bindPageEntry(node)')
  })

  it('ports the download cursor lag and click-firework signature', () => {
    expect(runtimeSource).toContain('prev.x += (curr.x - prev.x) * 0.35')
    expect(runtimeSource).toContain('prev.y += (curr.y - prev.y) * 0.35')
    expect(runtimeSource).toContain("const colors = ['252, 146, 174', '202, 180, 190', '207, 198, 255']")
    expect(runtimeSource).toContain('Array.from({ length: 20 }')
    expect(runtimeSource).toContain('randomInt(50, 100)')
    expect(runtimeSource).toContain('randomInt(900, 1500)')
    expect(interactionCss).toContain('z-index: 10087')
    expect(interactionCss).toContain('z-index: 10086')
    expect(interactionCss).toContain('z-index: 999')
  })

  it('uses the download page shell and modal entrance directions instead of the old shrink motion', () => {
    expect(interactionCss).toContain('animation: smirel-download-shell-in .65s ease both')
    expect(interactionCss).toContain('transform: scale(1.08)')
    expect(interactionCss).toContain('filter: blur(7px)')
    expect(interactionCss).toContain('transition: transform .32s ease, opacity .32s ease !important')
    expect(interactionCss).toContain('transform: scale(1.05) !important')
    expect(interactionCss).toContain('.smirel-download-control-motion:active')
    expect(interactionCss).toContain('transform: scale(.98) !important')
  })

  it('keeps accessibility guards while applying the richer desktop motion', () => {
    expect(runtimeSource).toContain("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
    expect(runtimeSource).toContain("window.matchMedia('(hover: hover) and (pointer: fine)').matches")
    expect(interactionCss).toContain('@media (max-width: 720px), (pointer: coarse)')
    expect(interactionCss).toContain('@media (prefers-reduced-motion: reduce)')
    expect(interactionCss).toContain('pointer-events: none')
  })
})
