import { afterEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { createI18n } from 'vue-i18n'
import { flushPromises, mount } from '@vue/test-utils'
import ApiKeysPage from '../pages/ApiKeysPage.vue'
import { api } from '../core/api'
import { workspaceRouteSettledKey } from '../core/route-motion'

// preferences.ts evaluates matchMedia during module import. jsdom does
// not ship it, so install a browser-compatible stub before SFC imports run.
vi.hoisted(() => {
  if (typeof window !== 'undefined') {
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: (media: string) => ({
        matches: false,
        media,
        onchange: null,
        addListener: () => undefined,
        removeListener: () => undefined,
        addEventListener: () => undefined,
        removeEventListener: () => undefined,
        dispatchEvent: () => false,
      }),
    })
  }
})

vi.mock('../core/api', () => ({
  previewMode: false,
  api: { get: vi.fn() },
  getErrorMessage: (error: unknown) => String(error),
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => { resolve = done })
  return { promise, resolve }
}

afterEach(() => vi.clearAllMocks())

describe('API key slow-response card entrance', () => {
  it('holds placeholders for route + key data but never waits for the secondary rates endpoint', async () => {
    const keys = deferred<{ data: unknown }>()
    const groups = deferred<{ data: unknown }>()
    const rates = deferred<{ data: unknown }>()
    vi.mocked(api.get).mockImplementation((url) => {
      if (url === '/keys') return keys.promise as ReturnType<typeof api.get>
      if (url === '/groups/available') return groups.promise as ReturnType<typeof api.get>
      if (url === '/groups/rates') return rates.promise as ReturnType<typeof api.get>
      throw new Error(`Unexpected endpoint: ${url}`)
    })

    const routeSettled = ref(false)
    const i18n = createI18n({
      legacy: false,
      locale: 'zh-CN',
      messages: { 'zh-CN': { workspace: { createdAt: '创建时间', delete: '删除' } } },
    })
    const wrapper = mount(ApiKeysPage, {
      global: {
        plugins: [i18n],
        provide: { [workspaceRouteSettledKey as symbol]: routeSettled },
      },
    })

    expect(wrapper.findAll('.keys-loading-card')).toHaveLength(2)
    expect(wrapper.findAll('.api-key-grid > .api-key-card')).toHaveLength(0)

    keys.resolve({ data: [
      { id: 11, name: 'A', status: 'active', group_id: 1, group: { id: 1, name: 'Team' } },
      { id: 12, name: 'B', status: 'active', group_id: 1, group: { id: 1, name: 'Team' } },
    ] })
    groups.resolve({ data: [{ id: 1, name: 'Team', rate_multiplier: 1 }] })
    await flushPromises()
    await nextTick()

    // Both list endpoints are done. The rates endpoint is still unresolved;
    // the page has the keys but must not spend the entrance before route end.
    expect(vi.mocked(api.get).mock.calls.some(([url]) => url === '/groups/rates')).toBe(true)
    expect(wrapper.findAll('.keys-loading-card')).toHaveLength(2)
    expect(wrapper.findAll('.api-key-grid > .api-key-card')).toHaveLength(0)

    routeSettled.value = true
    await vi.waitFor(async () => {
      await nextTick()
      expect(wrapper.findAll('.api-key-grid > .api-key-card')).toHaveLength(2)
    })
    expect(wrapper.findAll('.keys-loading-card')).toHaveLength(0)
    expect(wrapper.text()).toContain('A')
    expect(wrapper.text()).toContain('B')
    wrapper.unmount()
  })
})
