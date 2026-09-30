import { describe, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import AsyncDataReveal from '../components/AsyncDataReveal.vue'
import { workspaceRouteSettledKey } from '../core/route-motion'

function mountDeferredRegion(initialData: boolean, initialRoute: boolean) {
  const dataReady = ref(initialData)
  const routeSettled = ref(initialRoute)
  const Host = defineComponent({
    setup: () => () => h(AsyncDataReveal, { ready: dataReady.value }, {
      loading: () => h('div', { class: 'placeholder' }, 'waiting for response'),
      default: () => h('div', { class: 'real-cards' }, 'server result'),
    }),
  })
  const wrapper = mount(Host, { global: { provide: { [workspaceRouteSettledKey as symbol]: routeSettled } } })
  return { wrapper, dataReady, routeSettled }
}

async function expectRealCards(wrapper: ReturnType<typeof mountDeferredRegion>['wrapper']) {
  await vi.waitFor(async () => {
    await nextTick()
    expect(wrapper.find('.real-cards').exists()).toBe(true)
  })
}

describe('async data readiness and route choreography', () => {
  it('does not spend card entrance while the network is slower than the route', async () => {
    const { wrapper, dataReady, routeSettled } = mountDeferredRegion(false, false)
    expect(wrapper.find('.placeholder').exists()).toBe(true)
    routeSettled.value = true
    await nextTick()
    expect(wrapper.find('.real-cards').exists()).toBe(false)
    dataReady.value = true
    await expectRealCards(wrapper)
    expect(wrapper.find('.placeholder').exists()).toBe(false)
  })

  it('also waits for the route when data arrives faster than its page animation', async () => {
    const { wrapper, dataReady, routeSettled } = mountDeferredRegion(false, false)
    dataReady.value = true
    await nextTick()
    expect(wrapper.find('.real-cards').exists()).toBe(false)
    routeSettled.value = true
    await expectRealCards(wrapper)
  })

  it('preserves usable content when refresh starts or the previous route begins leaving', async () => {
    const { wrapper, dataReady, routeSettled } = mountDeferredRegion(true, true)
    await expectRealCards(wrapper)
    dataReady.value = false
    routeSettled.value = false
    await nextTick()
    expect(wrapper.find('.real-cards').exists()).toBe(true)
    expect(wrapper.find('.placeholder').exists()).toBe(false)
  })

  it('puts all four async card regions behind explicit initial loading state', () => {
    const read = (file: string) => readFileSync(resolve(process.cwd(), file), 'utf8')
    for (const name of ['ApiKeysPage', 'UserUsagePage', 'ModelCatalogPage', 'AdminOverviewPage']) {
      const page = read(`src/smirel/pages/${name}.vue`)
      expect(page, name).toContain('<AsyncDataReveal :ready="!loading"')
      expect(page, name).toContain('const loading = ref(true)')
    }
    const app = read('src/App.vue')
    expect(app).toContain('workspaceRouteSettled.value = !useWorkspace.value')
    expect(app).toContain('@after-enter="markWorkspaceEntered"')
    const shell = read('src/smirel/components/WorkspaceShell.vue')
    expect(shell).toContain("event.animationName === 'ws-shell-canvas-in'")
  })

  it('uses composited transitions and respects reduced motion', () => {
    const css = readFileSync(resolve(process.cwd(), 'src/smirel/styles/async-data-motion.css'), 'utf8')
    expect(css).toContain('transform 410ms')
    expect(css).not.toMatch(/filter:\s*blur\(/)
    expect(css).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
