import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import CangyuanMediaPricesModal from '../components/CangyuanMediaPricesModal.vue'
import { api } from '../core/api'

vi.mock('../core/api', () => ({
  api: { get: vi.fn(), put: vi.fn(), delete: vi.fn() },
  getErrorMessage: (error: unknown) => error instanceof Error ? error.message : String(error),
}))
const get = vi.mocked(api.get)
const put = vi.mocked(api.put)

function field(label: string): HTMLInputElement {
  const wrapper = Array.from(document.querySelectorAll<HTMLLabelElement>('.cy-pricing-field'))
    .find(item => item.textContent?.includes(label))
  const input = wrapper?.querySelector<HTMLInputElement>('input')
  if (!input) throw Error('missing field: ' + label)
  return input
}

async function input(label: string, value: string) {
  const target = field(label)
  target.value = value
  target.dispatchEvent(new Event('input', { bubbles: true }))
  await nextTick()
}

async function newRule() {
  const button = document.querySelector<HTMLButtonElement>('.cy-pricing-actions .cy-pricing-primary')
  expect(button).toBeTruthy()
  button!.click()
  await nextTick()
  const model = document.querySelector<HTMLButtonElement>('.cy-pricing-models button')
  expect(model).toBeTruthy()
  model!.click()
  await nextTick()
  await input('档位说明', 'image-1k')
  await input('单个输出售价', '0.13')
  await input('单个输出成本', '0.08')
}

function save() {
  const button = document.querySelector<HTMLButtonElement>('.cy-pricing-editor-footer .cy-pricing-primary')
  expect(button).toBeTruthy()
  button!.click()
}

describe('Cangyuan media price console', () => {
  beforeEach(() => {
    get.mockReset()
    put.mockReset()
    get.mockImplementation(async url => {
      if (url === '/admin/cangyuan-media/prices') return { data: { data: [] } } as never
      if (url === '/admin/cangyuan-media/catalog') return { data: {
        pricing_version: 'provider-v2026',
        items: [
          { id: 'nano-banana2-1k', kind: 'image_generation', model_price: 0.08, billing_mode: 'per_request', request_unit: 'image', tiered: false },
          { id: 'sd14-seedance-2.0', kind: 'video_generation', model_price: 3, billing_mode: 'per_request', request_unit: 'generation', tiered: false },
        ],
      } } as never
      throw Error('unexpected URL ' + url)
    })
    put.mockResolvedValue({ data: { saved: true } } as never)
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.restoreAllMocks()
  })

  it('requires manually entered USD prices and only saves a disabled draft by default', async () => {
    const wrapper = mount(CangyuanMediaPricesModal, { props: { show: true }, attachTo: document.body })
    await flushPromises()
    expect(document.body.textContent).toContain('计费安全')
    expect(get).toHaveBeenCalledWith('/admin/cangyuan-media/catalog')

    const button = document.querySelector<HTMLButtonElement>('.cy-pricing-actions .cy-pricing-primary')
    button!.click()
    await nextTick()
    document.querySelector<HTMLButtonElement>('.cy-pricing-models button')!.click()
    await nextTick()
    await input('档位说明', 'image-1k')
    save()
    await flushPromises()
    expect(put).not.toHaveBeenCalled()
    expect(document.body.textContent).toContain('手动填写有效的美元售价')

    await input('单个输出售价', '0.13')
    await input('单个输出成本', '0.08')
    save()
    await flushPromises()
    expect(put).toHaveBeenCalledWith(
      '/admin/cangyuan-media/prices/nano-banana2-1k-v1',
      expect.objectContaining({
        model: 'nano-banana2-1k',
        enabled: false,
        options: { n: 1 },
        snapshot: expect.objectContaining({
          unit_provider_cost_usd: '0.08',
          unit_sale_price_usd: '0.13',
          pricing_version: 'provider-v2026',
          verified_by_admin: false,
        }),
      }),
    )
    wrapper.unmount()
  })

  it('keeps image edit billing distinct from image generation when selecting a model', async () => {
    const wrapper = mount(CangyuanMediaPricesModal, { props: { show: true }, attachTo: document.body })
    await flushPromises()
    document.querySelector<HTMLButtonElement>('.cy-pricing-actions .cy-pricing-primary')!.click()
    await nextTick()
    const selector = document.querySelector<HTMLSelectElement>('.cy-pricing-field select')!
    selector.value = 'image_edit'
    selector.dispatchEvent(new Event('change', { bubbles: true }))
    await nextTick()
    document.querySelector<HTMLButtonElement>('.cy-pricing-models button')!.click()
    await nextTick()
    expect(selector.value).toBe('image_edit')
    expect(field('规则 ID').value).toBe('edit-nano-banana2-1k-v1')
    wrapper.unmount()
  })

  it('blocks enabling user billing without explicit review confirmation', async () => {
    const wrapper = mount(CangyuanMediaPricesModal, { props: { show: true }, attachTo: document.body })
    await flushPromises()
    await newRule()
    const enable = document.querySelector<HTMLInputElement>('.cy-pricing-check input')
    expect(enable).toBeTruthy()
    enable!.click()
    await nextTick()
    save()
    await flushPromises()
    expect(put).not.toHaveBeenCalled()
    expect(document.body.textContent).toContain('请先勾选已核对官方币种')

    const reviewed = document.querySelector<HTMLInputElement>('.cy-pricing-reviewed input')
    expect(reviewed).toBeTruthy()
    reviewed!.click()
    await nextTick()
    save()
    await flushPromises()
    expect(put).toHaveBeenCalledWith(
      expect.stringContaining('/admin/cangyuan-media/prices/'),
      expect.objectContaining({ enabled: true }),
    )
    wrapper.unmount()
  })
})
