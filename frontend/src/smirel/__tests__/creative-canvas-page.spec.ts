import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import CreativeCanvasPage from '../pages/CreativeCanvasPage.vue'
import { api } from '../core/api'
import { starterProject } from '../core/creative-canvas'

vi.mock('../core/api', () => ({
  previewMode: false,
  api: {
    get: vi.fn(async () => ({ data: [] })),
    post: vi.fn(),
    put: vi.fn(async () => ({ data: { version: 1 } })),
    delete: vi.fn(async () => ({ data: { deleted: true } })),
  },
  getErrorMessage: (error: unknown) => error instanceof Error ? error.message : String(error),
}))
const user = reactive({ id: 99, role: 'admin', email: 'test@example.invalid' })
vi.mock('../core/session', () => ({ useSession: () => ({ state: { user } }) }))

function mountCanvas() {
  return mount(CreativeCanvasPage, {
    attachTo: document.body,
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

describe('creative canvas UI foundation', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    vi.mocked(api.get).mockImplementation(async () => ({ data: [] }) as never)
    vi.mocked(api.put).mockImplementation(async () => ({ data: { version: 1 } }) as never)
    vi.mocked(api.delete).mockImplementation(async () => ({ data: { deleted: true } }) as never)
    vi.stubGlobal('confirm', vi.fn(() => true))
  })
  afterEach(() => {
    user.id = 99
    document.body.innerHTML = ''
    vi.unstubAllGlobals()
  })

  it('renders a draggable graph and a working node inspector', async () => {
    const wrapper = mountCanvas()
    await flushPromises()
    expect(wrapper.findAll('.cc-node')).toHaveLength(3)
    expect(wrapper.findAll('.cc-connectors path').length).toBeGreaterThanOrEqual(3) // marker + 2 links
    expect(wrapper.text()).toContain('AI 创作画布')
    expect(wrapper.text()).toContain('我的第一张创作画布')
    expect(wrapper.find('.cc-inspector').text()).toContain('画布属性')

    await wrapper.findAll('.cc-node')[0].trigger('click')
    expect(wrapper.find('.cc-inspector').text()).toContain('提示词')
    expect(wrapper.find('textarea').element.value).toContain('橘猫')
    wrapper.unmount()
  })

  it('adds, duplicates, connects, undoes and saves nodes locally without an API call', async () => {
    const fetchSpy = vi.fn()
    vi.stubGlobal('fetch', fetchSpy)
    const wrapper = mountCanvas()
    await flushPromises()
    expect(wrapper.findAll('.cc-node')).toHaveLength(3)

    await wrapper.find('.cc-quick-add button').trigger('click')
    expect(wrapper.findAll('.cc-node')).toHaveLength(4)
    await wrapper.find('.cc-node-actions button').trigger('click')
    expect(wrapper.findAll('.cc-node')).toHaveLength(5)

    await wrapper.find('button[title="撤销"]').trigger('click')
    expect(wrapper.findAll('.cc-node')).toHaveLength(4)
    await wrapper.find('button[title="重做"]').trigger('click')
    expect(wrapper.findAll('.cc-node')).toHaveLength(5)

    // Link two valid nodes via their accessible connection ports.
    const source = wrapper.findAll('.cc-node')[0]
    const target = wrapper.findAll('.cc-node')[2] // video accepts input from prompt
    await source.find('.cc-port-out').trigger('click')
    expect(wrapper.text()).toContain('连接中')
    await target.find('.cc-port-in').trigger('click')
    expect(wrapper.text()).not.toContain('连接中：点击目标')
    expect(fetchSpy).not.toHaveBeenCalled()

    // Debounced autosave persists only this user's canvas structure.
    await new Promise(resolve => setTimeout(resolve, 600))
    const stored = localStorage.getItem('muxway.creation.projects.v1.user-99') || ''
    expect(stored).toContain('我的第一张创作画布')
    expect(JSON.parse(stored).projects[0].nodes).toHaveLength(5)
    expect(stored).not.toContain('sk-')
    wrapper.unmount()
  })

  it('keeps unsaved edits with the previous account during a fast user switch', async () => {
    const wrapper = mountCanvas()
    await flushPromises()
    await wrapper.findAll('.cc-node')[0].trigger('click')
    await wrapper.find('.cc-inspector textarea').setValue('只属于用户 99 的私有草稿')
    await nextTick()

    // A rapid logout/login must never write old content under a new user ID.
    user.id = 100
    await nextTick()
    await flushPromises()

    expect(localStorage.getItem('muxway.creation.projects.v1.user-99')).toContain('只属于用户 99 的私有草稿')
    expect(localStorage.getItem('muxway.creation.projects.v1.user-100') || '').not.toContain('只属于用户 99 的私有草稿')
    expect(wrapper.findAll('.cc-node')).toHaveLength(3)
    wrapper.unmount()
  })

  it('restores a cloud document and preserves local-only projects', async () => {
    const remote = starterProject()
    remote.id = 'cr_remote_user99'
    remote.title = 'Cloud restored project'
    vi.mocked(api.get).mockImplementation(async (url) => {
      if (url === '/creative/projects') return { data: { items: [{
        id: remote.id, title: remote.title, version: 4, updated_at: remote.updatedAt,
      }] } } as never
      if (url === '/creative/projects/' + remote.id) return { data: { project: remote, version: 4 } } as never
      return { data: [] } as never
    })
    const wrapper = mountCanvas()
    await flushPromises()
    expect(wrapper.findAll('.cc-project-select option')).toHaveLength(2)
    expect(wrapper.text()).toMatch(/已连接云端|已同步云端/)
    expect(wrapper.find('.cc-project-select').text()).toContain('Cloud restored project')
    wrapper.unmount()
  })

  it('forks a same-user conflict while preserving an already-submitted paid job ID', async () => {
    const local = starterProject()
    local.nodes[1].taskId = 'cyjob_' + 'a'.repeat(32)
    local.nodes[1].taskKeyId = 47
    local.nodes[1].status = 'queued'
    localStorage.setItem('muxway.creation.projects.v1.user-99', JSON.stringify({
      activeId: local.id, projects: [local],
    }))
    const remote = JSON.parse(JSON.stringify(local)) as typeof local
    remote.title = 'Cloud version changed on another device'
    vi.mocked(api.get).mockImplementation(async (url) => {
      if (url === '/creative/projects') return { data: { items: [{
        id: remote.id, title: remote.title, version: 6, updated_at: remote.updatedAt,
      }] } } as never
      if (url === '/creative/projects/' + remote.id) return { data: { project: remote, version: 6 } } as never
      return { data: [] } as never
    })
    const wrapper = mountCanvas()
    await flushPromises()
    const localRecord = localStorage.getItem('muxway.creation.projects.v1.user-99') || ''
    const stored = JSON.parse(localRecord) as { projects: typeof local[] }
    expect(stored.projects).toHaveLength(2)
    const recovery = stored.projects.find(item => item.id !== remote.id)
    expect(recovery?.nodes[1].taskId).toBe(local.nodes[1].taskId)
    expect(recovery?.nodes[1].taskKeyId).toBe(47)
    wrapper.unmount()
  })

  it('requires an approved quote and sends the exact price/version with only one paid POST', async () => {
    vi.mocked(api.get).mockImplementation(async (url) => {
      if (url === '/keys') return { data: [{
        id: 47, name: 'Sandbox image', key: 'sk-qa-test-only', status: 'active',
        group_id: 30, group: { id: 30, name: 'cangyuan-media-sandbox' },
      }] } as never
      return { data: [] } as never
    })
    vi.mocked(api.post).mockImplementation(async () => ({ data: {
      currency: 'USD', kind: 'image', model: 'gpt-image-2.6', api_key_id: 47,
      group_id: 30, units: 1, unit_sale_price_usd: '0.01', total_sale_usd: '0.01000000',
      tier: 'image-1', pricing_version: 'v1', approved: true, disclaimer: '',
    } }) as never)
    const paid = vi.fn(async () => ({
      ok: true, status: 202,
      json: async () => ({ task_id: 'cyjob_' + 'a'.repeat(32), status: 'queued' }),
    }))
    vi.stubGlobal('fetch', paid)
    const wrapper = mountCanvas()
    await flushPromises()
    await wrapper.findAll('.cc-node')[1].trigger('click')
    await wrapper.find('.cc-run-button').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledOnce()
    expect(paid).toHaveBeenCalledTimes(1)
    const [url, init] = paid.mock.calls[0] as unknown as [string, RequestInit]
    expect(url).toBe('/v1/images/generations')
    const headers = init.headers as Record<string, string>
    expect(headers['X-Canvas-Quoted-Sale-USD']).toBe('0.01000000')
    expect(headers['X-Canvas-Quoted-Pricing-Version']).toBe('v1')
    expect(headers['Idempotency-Key']).toMatch(/^canvas-/)
    expect(wrapper.text()).toContain('已排队')
    expect(localStorage.getItem('muxway.creation.projects.v1.user-99')).toContain('cyjob_')
    wrapper.unmount()
  })

  it('does not call any paid API if exact price quote is unavailable', async () => {
    vi.mocked(api.get).mockImplementation(async (url) => {
      if (url === '/keys') return { data: [{ id: 47, key: 'sk-fixture', status: 'active' }] } as never
      return { data: [] } as never
    })
    vi.mocked(api.post).mockRejectedValue(new Error('no reviewed price'))
    const paid = vi.fn()
    vi.stubGlobal('fetch', paid)
    const wrapper = mountCanvas()
    await flushPromises()
    await wrapper.findAll('.cc-node')[1].trigger('click')
    await wrapper.find('.cc-run-button').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledOnce()
    expect(paid).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('没有有效审核报价')
    wrapper.unmount()
  })

  it('creates a separate blank project without overwriting the starter workflow', async () => {
    const wrapper = mountCanvas()
    await flushPromises()
    await wrapper.find('button[title="新建画布"]').trigger('click')
    await nextTick()
    expect(wrapper.findAll('.cc-node')).toHaveLength(0)
    expect(wrapper.find('.cc-canvas-empty').exists()).toBe(true)
    expect(wrapper.findAll('.cc-project-select option')).toHaveLength(2)
    wrapper.unmount()
  })
})
