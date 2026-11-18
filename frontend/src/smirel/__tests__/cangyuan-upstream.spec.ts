import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import CreateUpstreamAccountModal from '../components/CreateUpstreamAccountModal.vue'
import { api } from '../core/api'

vi.mock('../core/api', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
  previewMode: false,
  getErrorMessage: (error: unknown) => error instanceof Error ? error.message : String(error),
}))

const get = vi.mocked(api.get)
const post = vi.mocked(api.post)

function element<T extends HTMLElement>(selector: string): T {
  const found = document.querySelector<T>(selector)
  if (!found) throw new Error(`Missing element: ${selector}`)
  return found
}

function change<T extends HTMLInputElement | HTMLSelectElement>(input: T, value: string) {
  input.value = value
  input.dispatchEvent(new Event('input', { bubbles: true }))
  input.dispatchEvent(new Event('change', { bubbles: true }))
}

describe('Cangyuan New API setup', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({
      data: [{ id: 99, name: 'openai-default', platform: 'openai', status: 'active' }],
    } as never)
    post.mockImplementation(async (url) => {
      if (url === '/admin/accounts/models/sync-upstream-preview') {
        return { data: { models: ['verified-chat-model'], filtered_count: 1, pricing: { 'verified-chat-model': { billing_mode: 'tiered_expr', tiered: true } } } } as never
      }
      return { data: { id: 14 } } as never
    })
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('requires explicit approved models and never silently attaches openai-default', async () => {
    const wrapper = mount(CreateUpstreamAccountModal, {
      props: { show: true },
      attachTo: document.body,
    })
    await flushPromises()

    const preset = Array.from(document.querySelectorAll<HTMLSelectElement>('select'))
      .find((select) => select.querySelector('option[value="cangyuan"]'))
    expect(preset).toBeDefined()
    change(preset!, 'cangyuan')
    await nextTick()

    expect(element<HTMLInputElement>('.input-with-prefix input').value)
      .toBe('https://ai.cangyuansuanli.cn')
    expect(element<HTMLButtonElement>('.save-button').disabled).toBe(true)

    change(element<HTMLInputElement>('.secret-input input'), 'test-not-a-real-key')
    await nextTick()
    element<HTMLButtonElement>('.model-probe button').click()
    await flushPromises()

    expect(post).toHaveBeenCalledWith(
      '/admin/accounts/models/sync-upstream-preview',
      expect.objectContaining({
        platform: 'openai',
        type: 'apikey',
        vendor: 'cangyuan',
        base_url: 'https://ai.cangyuansuanli.cn',
        api_key: 'test-not-a-real-key',
      }),
    )
    expect(element<HTMLButtonElement>('.save-button').disabled).toBe(true)

    const choices = Array.from(document.querySelectorAll<HTMLLabelElement>('.cangyuan-model'))
    expect(choices).toHaveLength(1)
    expect(choices[0].textContent).toContain('分档计费')
    expect(document.querySelector('.cangyuan-warning')?.textContent).toContain('排除了 1 个')
    const selected = choices.find((label) => label.textContent?.includes('verified-chat-model'))
    expect(selected).toBeDefined()
    selected!.querySelector<HTMLInputElement>('input')!.click()
    await nextTick()

    element<HTMLButtonElement>('.save-button').click()
    await flushPromises()
    const create = post.mock.calls.find(([url]) => url === '/admin/accounts')
    expect(create).toBeDefined()
    expect(create![1]).toEqual(expect.objectContaining({
      platform: 'openai',
      type: 'apikey',
      group_ids: [],
      skip_default_group_bind: true,
      extra: {
        upstream_vendor: 'cangyuan',
        openai_passthrough: false,
        openai_responses_mode: 'force_chat_completions',
      },
      credentials: {
        api_key: 'test-not-a-real-key',
        base_url: 'https://ai.cangyuansuanli.cn',
        model_mapping: { 'verified-chat-model': 'verified-chat-model' },
      },
    }))
    wrapper.unmount()
  })
  it('reports media models separately without enabling a media-only account', async () => {
    post.mockImplementation(async (url) => {
      if (url === '/admin/accounts/models/sync-upstream-preview') {
        return {
          data: {
            models: [],
            filtered_count: 2,
            media_models: {
              image: [{ id: 'photo-model', billing_mode: 'tiered_expr', tiered: true }],
              video: [{ id: 'video-model', billing_mode: 'per_request', tiered: false }],
            },
          },
        } as never
      }
      return { data: { id: 14 } } as never
    })
    const wrapper = mount(CreateUpstreamAccountModal, { props: { show: true }, attachTo: document.body })
    await flushPromises()
    const preset = Array.from(document.querySelectorAll<HTMLSelectElement>('select'))
      .find((select) => select.querySelector('option[value="cangyuan"]'))
    expect(preset).toBeDefined()
    change(preset!, 'cangyuan')
    await nextTick()

    const key = element<HTMLInputElement>('.secret-input input')
    change(key, 'safe-test-only-key')
    element<HTMLButtonElement>('.model-probe button').click()
    await flushPromises()

    const preview = element<HTMLElement>('.cangyuan-media-preview')
    expect(preview.textContent).toContain('photo-model')
    expect(preview.textContent).toContain('video-model')
    expect(preview.textContent).toContain('分档')
    expect(preview.textContent).toContain('按次')
    expect(document.querySelectorAll('.cangyuan-model')).toHaveLength(0)
    expect(element<HTMLButtonElement>('.save-button').disabled).toBe(true)

    change(key, 'a-different-key')
    await nextTick()
    expect(document.querySelector('.cangyuan-media-preview')).toBeNull()
    expect(element<HTMLButtonElement>('.save-button').disabled).toBe(true)
    expect(post.mock.calls.filter(([url]) => url === '/admin/accounts')).toHaveLength(0)
    wrapper.unmount()
  })

})
