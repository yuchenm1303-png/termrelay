import { api } from './api'
import type { CreativeProject } from './creative-canvas'

export interface CloudProjectMetadata {
  id: string
  title: string
  version: number
  created_at: string
  updated_at: string
}
export interface CloudProjectSnapshot {
  project: CreativeProject
  version: number
}
export interface CloudProjectVersion {
  id: string
  version: number
}
export interface CreativeMediaQuote {
  currency: 'USD'
  kind: 'image' | 'video'
  model: string
  api_key_id: number
  group_id: number
  units: number
  unit_sale_price_usd: string
  total_sale_usd: string
  tier: string
  pricing_version: string
  approved: boolean
  disclaimer: string
}
export interface CreativeQuoteRequest {
  api_key_id: number
  kind: 'image' | 'video'
  model: string
  payload: Record<string, unknown>
}
const projectsPath = '/creative/projects'

export async function listCloudProjects(): Promise<CloudProjectMetadata[]> {
  const response = await api.get<{ items?: CloudProjectMetadata[] }>(projectsPath)
  return Array.isArray(response.data?.items) ? response.data.items : []
}
export async function getCloudProject(id: string): Promise<CloudProjectSnapshot> {
  const result = await api.get<CloudProjectSnapshot>(projectsPath + '/' + encodeURIComponent(id))
  return result.data
}
export async function putCloudProject(project: CreativeProject, expectedVersion: number): Promise<CloudProjectVersion> {
  const result = await api.put<CloudProjectVersion>(projectsPath + '/' + encodeURIComponent(project.id),
    { project, expected_version: expectedVersion })
  return result.data
}
export async function deleteCloudProject(id: string, version: number): Promise<void> {
  await api.delete(projectsPath + '/' + encodeURIComponent(id), { params: { version } })
}
export async function getCreativeMediaQuote(input: CreativeQuoteRequest): Promise<CreativeMediaQuote> {
  const result = await api.post<CreativeMediaQuote>('/creative/media/quote', input)
  if (result.data?.currency !== 'USD' || result.data?.approved !== true ||
    result.data?.model !== input.model || result.data?.api_key_id !== input.api_key_id ||
    !/^\d+(?:\.\d{1,8})?$/.test(result.data.total_sale_usd) || !result.data.pricing_version) {
    throw Error('报价未通过核验，请勿执行生成任务')
  }
  return result.data
}
