import { api } from './api'

export interface AdminComplianceAcknowledgement {
  version: string
  document_zh: string
  document_en: string
  admin_user_id: number
  ip_address?: string
  user_agent?: string
  accepted_at: string
}

export interface AdminComplianceStatus {
  required: boolean
  version: string
  document_path_zh: string
  document_path_en: string
  document_url_zh: string
  document_url_en: string
  ack_phrase_zh: string
  ack_phrase_en: string
  acknowledgement?: AdminComplianceAcknowledgement | null
}

let cachedStatus: AdminComplianceStatus | null = null
let inflight: Promise<AdminComplianceStatus> | null = null

export function resetAdminComplianceCache() {
  cachedStatus = null
  inflight = null
}

export async function getAdminComplianceStatus(force = false): Promise<AdminComplianceStatus> {
  if (!force && cachedStatus) return cachedStatus
  if (!force && inflight) return inflight

  const request = api.get<AdminComplianceStatus>('/admin/compliance')
    .then(({ data }) => {
      cachedStatus = data
      return data
    })
    .finally(() => {
      inflight = null
    })

  inflight = request
  return request
}

export async function acceptAdminCompliance(phrase: string, language: 'zh' | 'en'): Promise<AdminComplianceStatus> {
  const { data } = await api.post<AdminComplianceStatus>('/admin/compliance/accept', {
    phrase,
    language,
  })
  cachedStatus = data
  return data
}
