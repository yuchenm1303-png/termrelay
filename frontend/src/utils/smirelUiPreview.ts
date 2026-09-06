import { adminAPI } from '@/api/admin'
import { useAdminComplianceStore, useAdminSettingsStore, useAuthStore } from '@/stores'
import type { User } from '@/types'

/**
 * GitHub Pages is a static UI review surface, not a production auth surface.
 * This flag is injected only by the Smirel UI preview workflow.
 */
export const isSmirelUiPreview = import.meta.env.VITE_UI_PREVIEW === 'true'

const PREVIEW_ADMIN_USER: User = {
  id: 0,
  username: 'Preview Admin',
  email: 'preview@smirel.local',
  role: 'admin',
  balance: 128.4,
  concurrency: 20,
  status: 'active',
  allowed_groups: null,
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  subscriptions: [],
  last_active_at: new Date().toISOString(),
  created_at: '2026-01-01T00:00:00.000Z',
  updated_at: new Date().toISOString(),
}

const PREVIEW_DASHBOARD_STATS = {
  total_users: 128,
  today_new_users: 4,
  active_users: 37,
  hourly_active_users: 12,
  total_api_keys: 154,
  active_api_keys: 146,
  total_accounts: 12,
  normal_accounts: 12,
  error_accounts: 0,
  total_requests: 124680,
  today_requests: 4287,
  today_tokens: 18_640_000,
  today_actual_cost: 46.71,
  today_account_cost: 31.28,
  average_duration_ms: 842,
  rpm: 18,
  uptime: 99.99,
}

function today(): string {
  return new Date().toISOString().slice(0, 10)
}

function installDashboardPreviewAdapter(): void {
  const dashboard = adminAPI.dashboard as unknown as Record<
    string,
    (...args: any[]) => Promise<any>
  >

  dashboard.getStats = async () => ({ ...PREVIEW_DASHBOARD_STATS })
  dashboard.getRealtimeMetrics = async () => ({
    active_requests: 3,
    requests_per_minute: 18,
    average_response_time: 842,
    error_rate: 0.003,
  })
  dashboard.getUsageTrend = async (params: any = {}) => ({
    trend: [],
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
    granularity: params.granularity || 'hour',
  })
  dashboard.getModelStats = async (params: any = {}) => ({
    models: [],
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
  })
  dashboard.getGroupStats = async (params: any = {}) => ({
    groups: [],
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
  })
  dashboard.getSnapshotV2 = async (params: any = {}) => ({
    generated_at: new Date().toISOString(),
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
    granularity: params.granularity || 'hour',
    stats: params.include_stats === false ? undefined : { ...PREVIEW_DASHBOARD_STATS },
    trend: [],
    models: [],
    groups: [],
    users_trend: [],
  })
  dashboard.getApiKeyUsageTrend = async (params: any = {}) => ({
    trend: [],
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
    granularity: params.granularity || 'hour',
  })
  dashboard.getUserUsageTrend = async (params: any = {}) => ({
    trend: [],
    start_date: params.start_date || today(),
    end_date: params.end_date || today(),
    granularity: params.granularity || 'hour',
  })
  dashboard.getUserSpendingRanking = async () => ({
    ranking: [],
    total_actual_cost: 0,
    total_requests: 0,
    total_tokens: 0,
  })
  dashboard.getBatchUsersUsage = async () => ({ stats: {} })
  dashboard.getBatchApiKeysUsage = async () => ({ stats: {} })
}

/**
 * Install the preview-only runtime adapter after Pinia has been registered.
 * It deliberately does not persist a real auth token and never authenticates
 * against production. The actual workspace/components stay unchanged.
 */
export function installSmirelUiPreview(): void {
  if (!isSmirelUiPreview) return

  document.documentElement.classList.add('smirel-ui-preview')

  const authStore = useAuthStore()
  authStore.$patch({
    token: 'smirel-ui-preview-session',
    user: PREVIEW_ADMIN_USER,
  })

  // Route guards should not make a production compliance request in preview.
  const complianceStore = useAdminComplianceStore()
  complianceStore.$patch({ initialized: true })

  // Keep the Overview on its deterministic non-Ops state in static preview.
  const adminSettingsStore = useAdminSettingsStore()
  adminSettingsStore.setOpsMonitoringEnabledLocal(false)

  installDashboardPreviewAdapter()
}
