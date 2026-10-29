import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss, { type Declaration, type Rule } from 'postcss'

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const route = postcss.parse(read('src/smirel/styles/route-transitions.css'))

function declarations(selector: string) {
  const rule = route.nodes.find((node): node is Rule => node.type === 'rule' && node.selector === selector)
  expect(rule, selector).toBeDefined()
  return Object.fromEntries(rule!.nodes.filter((node): node is Declaration => node.type === 'decl')
    .map((node) => [node.prop, node.value]))
}

describe('workspace route motion', () => {
  it('enters every page from below with no lateral travel', () => {
    expect(declarations('.workspace-route-enter-from').transform).toMatch(/^translate3d\(0, calc\(18px/)
    expect(declarations('.workspace-route-leave-to').transform).toMatch(/^translate3d\(0, calc\(-8px/)
    expect(read('src/App.vue')).not.toContain('data-route-direction')
    expect(route.toString()).not.toMatch(/data-route-motion|--route-from-x|clip-path:\s*inset/)
  })

  it('keeps the light theme clarity override off transitioning stages', () => {
    expect(read('src/smirel/styles/workspace-light-tail.css'))
      .toContain('.workspace-route-stage:not(.workspace-route-enter-active):not(.workspace-route-leave-active)')
  })
})
