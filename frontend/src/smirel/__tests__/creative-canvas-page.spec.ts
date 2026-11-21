import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import CreativeCanvasPage from '../pages/CreativeCanvasPage.vue'

vi.mock('../core/api', () => ({
  previewMode: false,
  api: { get: vi.fn(async () => ({ data: [] })) },
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
