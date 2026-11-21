import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CangyuanMediaSmokeTest from '../components/CangyuanMediaSmokeTest.vue'

describe('Cangyuan one-click media smoke test', () => {
  const key = 'sk-qa-secret-key-must-never-appear-in-dom'
  beforeEach(() => {
    vi.useFakeTimers()
    vi.stubGlobal('confirm', vi.fn(() => true))
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('sends one paid POST with idempotency then only GET status checks', async () => {
    const requests: Array<{ url: string; method: string; headers: Record<string, string>; body?: string }> = []
    vi.stubGlobal('fetch', vi.fn(async (url: string, init: RequestInit) => {
      const method = init.method || 'GET'
      requests.push({ url, method, headers: init.headers as Record<string, string>, body: init.body as string | undefined })
      if (method === 'POST') {
        return {
          ok: true, status: 202,
          json: async () => ({ task_id: 'cyjob_qa123', status: 'queued' }),
        }
      }
      return {
        ok: true, status: 200,
        json: async () => ({ task_id: 'cyjob_qa123', status: 'completed', data: [{ url: 'https://example.com/qa-image.png' }] }),
      }
    }))

    const wrapper = mount(CangyuanMediaSmokeTest, { props: { apiKey: key } })
    expect(wrapper.html()).not.toContain(key)
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(requests.filter(request => request.method === 'POST')).toHaveLength(1)
    const post = requests.find(request => request.method === 'POST')!
    expect(post.url).toBe('/v1/images/generations')
    expect(post.headers.Authorization).toBe('Bearer ' + key)
    expect(post.headers['Idempotency-Key']).toMatch(/^media-qa-/)
    expect(JSON.parse(post.body || '{}')).toEqual(expect.objectContaining({
      model: 'gpt-image-2.6', n: 1, async: true, response_format: 'url',
    }))
    expect(wrapper.text()).toContain('已提交一次，不可重复')

    await vi.advanceTimersByTimeAsync(4500)
    await flushPromises()
    expect(requests.filter(request => request.method === 'POST')).toHaveLength(1)
    expect(requests.filter(request => request.method === 'GET')).toHaveLength(1)
    expect(wrapper.text()).toContain('completed')
    expect(wrapper.find('a[href="https://example.com/qa-image.png"]').exists()).toBe(true)
    expect(wrapper.html()).not.toContain(key)
    wrapper.unmount()
  })

  it('does not retry a network-ambiguous POST', async () => {
    const post = vi.fn(async () => { throw Error('network dropped') })
    vi.stubGlobal('fetch', post)
    const wrapper = mount(CangyuanMediaSmokeTest, { props: { apiKey: key } })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(post).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('禁止重试')
    await wrapper.find('button').trigger('click')
    expect(post).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })
})
