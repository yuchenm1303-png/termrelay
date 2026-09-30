import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const page = read('src/smirel/pages/ApiKeysPage.vue')
const app = read('src/App.vue')
const cards = postcss.parse(read('src/smirel/styles/api-keys.css'))
const route = postcss.parse(read('src/smirel/styles/route-transitions.css'))

function declarations(root: ReturnType<typeof postcss.parse>, selector: string) {
  const rule = root.nodes.find((node): node is Rule => node.type === 'rule' && node.selector === selector)
  expect(rule, selector).toBeDefined()
  return Object.fromEntries(rule!.nodes.filter((node): node is Declaration => node.type === 'decl')
    .map((node) => [node.prop, node.value]))
}

describe('API credential card entrance', () => {
  it('animates cards after async loading, including the initial group mount', () => {
    expect(page).toContain('const loading = ref(true)')
    expect(page).toContain('<TransitionGroup v-if="keys.length" appear name="key-card"')
    expect(page).toContain(':style="{ \'--key-enter-index\': Math.min(index, 7) }"')
    expect(page).toContain('<AsyncDataReveal :ready="!loading" :content-motion="false"')
    expect(page).toContain('class="api-key-card keys-loading-card"')
    expect(page).toContain('keys-loading-secret')
  })

  it('uses only compositor-friendly properties for large card entrances', () => {
    const enter = declarations(cards, '.workspace-root .api-key-grid > .key-card-enter-active')
    expect(enter.transition).toContain('opacity')
    expect(enter.transition).toContain('transform')
    expect(enter.transition).not.toMatch(/filter|clip-path|box-shadow/)
    expect(enter['transition-delay']).toContain('--key-enter-index')
    const from = declarations(cards, '.workspace-root .api-key-grid > .key-card-enter-from')
    expect(from.transform).toContain('translate3d')
  })

  it('does not stack the expensive route blur/clip/stagger over the credential cards', () => {
    expect(app).toContain("'workspace-route-stage--keys': route.path === '/keys'")
    const enter = declarations(route, '.workspace-route-enter-active.workspace-route-stage--keys')
    expect(enter.transition).not.toMatch(/filter|clip-path/)
    const from = declarations(route, '.workspace-route-enter-from.workspace-route-stage--keys')
    expect(from.filter).toBe('none')
    expect(from['clip-path']).toBe('none')
    const stagger = declarations(route,
      '.workspace-route-enter-active.workspace-route-stage--keys > :first-child > :nth-child(-n + 5)')
    expect(stagger.animation).toBe('none')
  })

  it('honors reduced motion', () => {
    const reduced = cards.nodes.filter((node) => node.type === 'atrule' &&
      node.name === 'media' && node.params === '(prefers-reduced-motion: reduce)')
    expect(reduced.length).toBeGreaterThan(0)
    expect(reduced.at(-1)?.toString()).toContain('.key-card-enter-from')
  })
})
