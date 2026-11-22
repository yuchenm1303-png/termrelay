import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../core/api'
import { starterProject } from '../core/creative-canvas'
import {
  deleteCloudProject, getCloudProject, getCreativeMediaQuote, listCloudProjects,
  putCloudProject,
} from '../core/creative-cloud'

vi.mock('../core/api', () => ({
  api: {
    get: vi.fn(), put: vi.fn(), post: vi.fn(), delete: vi.fn(),
  },
}))

describe('creative cloud API and verified media quote contract', () => {
  beforeEach(() => vi.clearAllMocks())
  afterEach(() => vi.unstubAllGlobals())

  it('lists JWT-scoped projects, then GETs one by encoded ID', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({ data: { items: [{ id: 'cr_a', title: 'Cloud A', version: 2 }] } } as never)
    const list = await listCloudProjects()
    expect(list).toHaveLength(1)
    expect(list[0].version).toBe(2)
    const doc = starterProject()
    vi.mocked(api.get).mockResolvedValueOnce({ data: { project: doc, version: 2 } } as never)
    expect((await getCloudProject(doc.id)).version).toBe(2)
    expect(vi.mocked(api.get).mock.calls[1][0]).toContain(doc.id)
  })

  it('uses CAS expected version for cloud writes, and version for safe delete', async () => {
    const doc = starterProject()
    vi.mocked(api.put).mockResolvedValue({ data: { id: doc.id, version: 5 } } as never)
    const updated = await putCloudProject(doc, 4)
    expect(updated.version).toBe(5)
    expect(api.put).toHaveBeenCalledWith('/creative/projects/' + encodeURIComponent(doc.id), {
      project: doc, expected_version: 4,
    })
    vi.mocked(api.delete).mockResolvedValue({ data: { deleted: true } } as never)
    await deleteCloudProject(doc.id, 5)
    expect(api.delete).toHaveBeenCalledWith('/creative/projects/' + encodeURIComponent(doc.id), { params: { version: 5 } })
  })

  it('accepts only matching authenticated USD quotes and refuses wrong-key/malformed prices', async () => {
    const input = {
      api_key_id: 47, kind: 'image' as const, model: 'gpt-image-2.6',
      payload: { model: 'gpt-image-2.6', prompt: 'a drawing', n: 1, async: true },
    }
    const quote = {
      currency: 'USD', kind: 'image', model: input.model, api_key_id: input.api_key_id,
      group_id: 30, units: 1, unit_sale_price_usd: '0.01', total_sale_usd: '0.01000000',
      tier: 'default', pricing_version: 'v1', approved: true, disclaimer: '',
    }
    vi.mocked(api.post).mockResolvedValueOnce({ data: quote } as never)
    expect((await getCreativeMediaQuote(input)).total_sale_usd).toBe('0.01000000')
    expect(api.post).toHaveBeenCalledWith('/creative/media/quote', input)

    vi.mocked(api.post).mockResolvedValueOnce({ data: { ...quote, api_key_id: 99 } } as never)
    await expect(getCreativeMediaQuote(input)).rejects.toThrow('报价未通过核验')

    vi.mocked(api.post).mockResolvedValueOnce({ data: { ...quote, total_sale_usd: 'NaN' } } as never)
    await expect(getCreativeMediaQuote(input)).rejects.toThrow('报价未通过核验')
  })
})
