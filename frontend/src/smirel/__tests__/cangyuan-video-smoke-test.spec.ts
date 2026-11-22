import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CangyuanVideoSmokeTest from '../components/CangyuanVideoSmokeTest.vue'

describe('Cangyuan once-only video QA card', () => {
  const key = 'sk-video-private-test-key'
  beforeEach(() => {
    window.sessionStorage.clear()
    vi.useFakeTimers()
    vi.stubGlobal('confirm', vi.fn(() => true))
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('allows GET of a settled video after quota exhaustion but never POST', async () => {
    const requests: string[] = []
    window.sessionStorage.setItem('cangyuan-video-qa-one-shot:47', 'cyjob_paidvideo123')
    vi.stubGlobal('fetch', vi.fn(async (url: string, init: RequestInit) => {
      requests.push((init.method || 'GET') + ' ' + url)
      return { ok: true, status: 200, json: async () => ({
        status: 'completed', video_url: 'https://example.com/settled-video.mp4',
      }) }
    }))
    const wrapper = mount(CangyuanVideoSmokeTest, {
      props: { apiKey: key, apiKeyId: 47, priceApproved: true, canSubmit: false },
    })
    await flushPromises()
    const buttons = wrapper.findAll('button')
    expect((buttons[0].element as HTMLButtonElement).disabled).toBe(true)
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await flushPromises()
    expect(requests).toEqual(['GET /v1/videos/cyjob_paidvideo123'])
    expect(wrapper.find('a[href="https://example.com/settled-video.mp4"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('prevents video submission before an exact admin price is approved', async () => {
    const post = vi.fn()
    vi.stubGlobal('fetch', post)
    const wrapper = mount(CangyuanVideoSmokeTest, {
      props: { apiKey: key, apiKeyId: 47, priceApproved: false },
    })
    await flushPromises()
    const button = wrapper.find('button')
    expect((button.element as HTMLButtonElement).disabled).toBe(true)
    expect(wrapper.text()).toContain('待管理员审核视频价格')
    await button.trigger('click')
    expect(post).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('creates one video task and only polls GET until completed', async () => {
    const requests: Array<{ url: string; method: string; headers: Record<string, string>; body?: string }> = []
    vi.stubGlobal('fetch', vi.fn(async (url: string, init: RequestInit) => {
      const method = init.method || 'GET'
      requests.push({ url, method, headers: init.headers as Record<string, string>, body: init.body as string | undefined })
      if (method === 'POST') {
        return { ok: true, status: 202, json: async () => ({ task_id: 'cyjob_videoqa1', status: 'queued' }) }
      }
      return { ok: true, status: 200, json: async () => ({
        task_id: 'cyjob_videoqa1', status: 'completed', video_url: 'https://example.com/video.mp4',
      }) }
    }))
    const wrapper = mount(CangyuanVideoSmokeTest, { props: { apiKey: key, apiKeyId: 47, priceApproved: true } })
    expect(wrapper.html()).not.toContain(key)
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(requests.filter(request => request.method === 'POST')).toHaveLength(1)
    const post = requests.find(request => request.method === 'POST')!
    expect(post.url).toBe('/v1/videos')
    expect(post.headers.Authorization).toBe('Bearer ' + key)
    expect(post.headers['Idempotency-Key']).toMatch(/^media-video-qa-/)
    expect(JSON.parse(post.body || '{}')).toEqual(expect.objectContaining({
      model: 'omni-fast', async: true,
    }))
    expect(wrapper.text()).toContain('已提交一次，不可重复')
    await vi.advanceTimersByTimeAsync(5000)
    await flushPromises()
    expect(requests.filter(request => request.method === 'POST')).toHaveLength(1)
    expect(requests.filter(request => request.method === 'GET')).toHaveLength(1)
    expect(wrapper.find('a[href="https://example.com/video.mp4"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('allows correction after a definite no-charge 422 without reopening a second accepted job', async () => {
    let posts = 0
    vi.stubGlobal('fetch', vi.fn(async (_url: string, init: RequestInit) => {
      if (init.method === 'POST') {
        posts++
        if (posts === 1) return {
          ok: false, status: 422,
          json: async () => ({ error: { message: 'Missing approved media price' } }),
        }
        return { ok: true, status: 202, json: async () => ({ task_id: 'cyjob_afterpricing', status: 'queued' }) }
      }
      return { ok: true, status: 200, json: async () => ({ status: 'queued' }) }
    }))
    const wrapper = mount(CangyuanVideoSmokeTest, { props: { apiKey: key, apiKeyId: 47, priceApproved: true } })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(posts).toBe(1)
    expect(wrapper.text()).toContain('请求已被明确拒绝')
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(posts).toBe(2)
    expect(wrapper.text()).toContain('已提交一次，不可重复')
    await wrapper.find('button').trigger('click')
    expect(posts).toBe(2)
    wrapper.unmount()
  })

  it('persists uncertain state and prevents repeat paid POST', async () => {
    const post = vi.fn(async () => { throw Error('dropped response') })
    vi.stubGlobal('fetch', post)
    const wrapper = mount(CangyuanVideoSmokeTest, { props: { apiKey: key, apiKeyId: 47, priceApproved: true } })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(post).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('禁止重试')
    wrapper.unmount()
    const restored = mount(CangyuanVideoSmokeTest, { props: { apiKey: key, apiKeyId: 47, priceApproved: true } })
    await flushPromises()
    await restored.find('button').trigger('click')
    expect(post).toHaveBeenCalledTimes(1)
    restored.unmount()
  })
})
