<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import WorkspaceNavIcon from './WorkspaceNavIcon.vue'
import { api, getErrorMessage, previewMode } from '../core/api'
import { interfacePreferences } from '../core/preferences'
import '../styles/admin-ops.css'

interface AdminOpsStats {
  total_users?: number
  active_users?: number
  total_api_keys?: number
  total_accounts?: number
  active_accounts?: number
  today_requests?: number
  today_tokens?: number
  today_actual_cost?: number
  average_duration_ms?: number
  rpm?: number
}

interface AdminOpsSnapshot {
  generated_at?: string
  stats?: AdminOpsStats
}

const { t } = useI18n()
const loading = ref(false)
const error = ref('')
const snapshot = ref<AdminOpsSnapshot | null>(null)
const stats = computed<AdminOpsStats>(() => snapshot.value?.stats ?? {})

function compact(value: unknown) {
  const n = Number(value || 0)
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`
  return n.toLocaleString(interfacePreferences.locale)
}

function money(value: unknown) {
  return `$${Number(value || 0).toFixed(2)}`
}

function duration(value: unknown) {
  const n = Number(value || 0)
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${Math.round(n)}ms`
}

function percent(part: unknown, total: unknown) {
  const totalValue = Number(total || 0)
  if (!totalValue) return 0
  return Math.max(0, Math.min(100, Math.round((Number(part || 0) / totalValue) * 100)))
}

const accountAvailability = computed(() => percent(stats.value.active_accounts, stats.value.total_accounts))
const generatedLabel = computed(() => {
  if (!snapshot.value?.generated_at) return t('admin.waitingFirstSync')
  const date = new Date(snapshot.value.generated_at)
  if (Number.isNaN(date.getTime())) return t('admin.justUpdated')
  return t('admin.updatedAt', {
    time: date.toLocaleTimeString(interfacePreferences.locale, { hour: '2-digit', minute: '2-digit' }),
  })
})

async function load() {
  error.value = ''
  loading.value = true

  try {
    if (previewMode) {
      snapshot.value = {
        generated_at: new Date().toISOString(),
        stats: {
          total_users: 1284,
          active_users: 438,
          total_api_keys: 2168,
          total_accounts: 42,
          active_accounts: 39,
          today_requests: 68420,
          today_tokens: 18400000,
          today_actual_cost: 428.36,
          average_duration_ms: 846,
          rpm: 176,
        },
      }
    } else {
      snapshot.value = (await api.get<AdminOpsSnapshot>('/admin/dashboard/snapshot-v2', {
        params: { include_stats: true },
      })).data
    }
  } catch (caught) {
    error.value = getErrorMessage(caught)
  } finally {
    loading.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <div class="admin-ops-workspace">
    <div class="ops-snapshot-bar">
      <div class="ops-snapshot-state" :class="{ ready: snapshot }">
        <span class="ops-state-icon"><WorkspaceNavIcon name="activity" /></span>
        <span>
          <strong>{{ snapshot ? t('admin.dataSynced') : t('admin.waitingData') }}</strong>
          <small>{{ generatedLabel }}</small>
        </span>
      </div>
      <button class="ops-refresh" type="button" :disabled="loading" @click="load">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M20 6v5h-5M4 18v-5h5" />
          <path d="M6.1 9a7 7 0 0 1 11.5-2.4L20 11M4 13l2.4 4.4A7 7 0 0 0 18 15" />
        </svg>
        <span>{{ loading ? t('workspace.refreshing') : t('admin.refreshData') }}</span>
      </button>
    </div>

    <p v-if="error" class="inline-error ops-error">{{ error }}</p>

    <section class="ops-metric-strip" aria-label="Operations snapshot">
      <article>
        <span class="ops-metric-label">REQUEST RATE</span>
        <div><strong>{{ compact(stats.rpm) }}</strong><small>RPM</small></div>
        <p>{{ t('admin.realtimeLoad') }}</p>
      </article>
      <article>
        <span class="ops-metric-label">LATENCY</span>
        <div><strong>{{ duration(stats.average_duration_ms) }}</strong></div>
        <p>{{ t('admin.averageResponse') }}</p>
      </article>
      <article>
        <span class="ops-metric-label">UPSTREAM</span>
        <div><strong>{{ accountAvailability }}%</strong></div>
        <p>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }} · {{ t('admin.upstreamAccounts') }}</p>
      </article>
      <article>
        <span class="ops-metric-label">REQUESTS</span>
        <div><strong>{{ compact(stats.today_requests) }}</strong></div>
        <p>{{ t('admin.todayRequests') }}</p>
      </article>
    </section>

    <div class="ops-main-grid">
      <section class="ops-resource-panel">
        <header class="ops-panel-head">
          <div>
            <span>RESOURCE STATUS</span>
            <strong>{{ t('admin.operationsOverview') }}</strong>
          </div>
          <RouterLink to="/admin/accounts">{{ t('admin.upstreamAccounts') }} →</RouterLink>
        </header>

        <div class="ops-resource-list">
          <div class="ops-resource-row">
            <span class="ops-resource-icon"><WorkspaceNavIcon name="server" /></span>
            <div class="ops-resource-copy">
              <strong>{{ t('admin.upstreamAccounts') }}</strong>
              <small>{{ t('admin.accountsAvailable', { value: stats.active_accounts || 0 }) }}</small>
            </div>
            <div class="ops-resource-value">
              <strong>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }}</strong>
              <small>{{ t('admin.availability', { value: accountAvailability }) }}</small>
            </div>
            <div class="ops-resource-progress" aria-hidden="true"><i :style="{ width: `${accountAvailability}%` }"></i></div>
          </div>

          <div class="ops-resource-row">
            <span class="ops-resource-icon"><WorkspaceNavIcon name="activity" /></span>
            <div class="ops-resource-copy">
              <strong>{{ t('admin.todayRequests') }}</strong>
              <small>{{ t('admin.realtimeLoad') }}</small>
            </div>
            <div class="ops-resource-value">
              <strong>{{ compact(stats.today_requests) }}</strong>
              <small>RPM {{ compact(stats.rpm) }}</small>
            </div>
          </div>

          <div class="ops-resource-row">
            <span class="ops-resource-icon"><WorkspaceNavIcon name="key" /></span>
            <div class="ops-resource-copy">
              <strong>API Keys</strong>
              <small>ISSUED CREDENTIALS</small>
            </div>
            <div class="ops-resource-value">
              <strong>{{ compact(stats.total_api_keys) }}</strong>
              <small>TOTAL</small>
            </div>
          </div>
        </div>
      </section>

      <section class="ops-volume-panel">
        <header class="ops-panel-head">
          <div>
            <span>TODAY</span>
            <strong>{{ t('admin.resourcePerformance') }}</strong>
          </div>
          <RouterLink to="/admin/usage">{{ t('admin.usageRecords') }} →</RouterLink>
        </header>

        <div class="ops-volume-primary">
          <span>{{ t('admin.todayToken') }}</span>
          <strong>{{ compact(stats.today_tokens) }}</strong>
          <small>TOKENS</small>
        </div>

        <div class="ops-volume-secondary">
          <div>
            <span>{{ t('admin.todayCost') }}</span>
            <strong>{{ money(stats.today_actual_cost) }}</strong>
            <small>USD</small>
          </div>
          <div>
            <span>{{ t('admin.averageResponse') }}</span>
            <strong>{{ duration(stats.average_duration_ms) }}</strong>
            <small>LATENCY</small>
          </div>
        </div>
      </section>
    </div>

    <nav class="ops-shortcuts" aria-label="Operations links">
      <RouterLink to="/admin/accounts">
        <span class="ops-shortcut-icon"><WorkspaceNavIcon name="server" /></span>
        <span><small>UPSTREAM</small><strong>{{ t('admin.upstreamScheduling') }}</strong></span>
        <b>→</b>
      </RouterLink>
      <RouterLink to="/admin/channels/pricing">
        <span class="ops-shortcut-icon"><WorkspaceNavIcon name="network" /></span>
        <span><small>CHANNELS</small><strong>{{ t('nav.adminChannels') }}</strong></span>
        <b>→</b>
      </RouterLink>
      <RouterLink to="/admin/usage">
        <span class="ops-shortcut-icon"><WorkspaceNavIcon name="chart" /></span>
        <span><small>USAGE</small><strong>{{ t('admin.usageRecords') }}</strong></span>
        <b>→</b>
      </RouterLink>
    </nav>
  </div>
</template>
