<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { api, getErrorMessage, previewMode } from '../core/api'
import { pushNotification } from '../core/notifications'
import { useSession } from '../core/session'

interface ApiKeyRow { id: number; name?: string; key?: string; status?: string; created_at?: string; [key: string]: unknown }
interface UsageRow { id?: number; model?: string; endpoint?: string; total_tokens?: number; actual_cost?: number; created_at?: string; [key: string]: unknown }
interface DashboardStats { total_api_keys?: number; active_api_keys?: number; total_requests?: number; total_tokens?: number; total_actual_cost?: number; today_requests?: number; today_tokens?: number; today_actual_cost?: number; [key: string]: unknown }

const route = useRoute()
const { t } = useI18n()
const { state } = useSession()
const feature = computed(() => String(route.meta.feature || 'module'))
const loading = ref(false)
const error = ref('')
const stats = ref<DashboardStats | null>(null)
const keys = ref<ApiKeyRow[]>([])
const usage = ref<UsageRow[]>([])
const newKeyName = ref('')

const isDashboard = computed(() => feature.value === 'dashboard')
const isKeys = computed(() => feature.value === 'keys')
const isUsage = computed(() => feature.value === 'usage')
const isProfile = computed(() => feature.value === 'profile')
const accountBalance = computed(() => Number(state.user?.balance || 0))
const visibleUsageTokens = computed(() => usage.value.reduce((sum, item) => sum + Number(item.total_tokens || 0), 0))
const visibleUsageCost = computed(() => usage.value.reduce((sum, item) => sum + Number(item.actual_cost || 0), 0))

const featureTitleKeys: Record<string, string> = {
  dashboard: 'nav.dashboard',
  keys: 'nav.keys',
  usage: 'nav.usage',
  subscriptions: 'nav.subscriptions',
  purchase: 'nav.purchase',
  orders: 'nav.orders',
  profile: 'nav.profile',
  'admin-dashboard': 'nav.adminDashboard',
  'admin-users': 'nav.adminUsers',
  'admin-accounts': 'nav.adminAccounts',
  'admin-groups': 'nav.adminGroups',
  'admin-channels': 'nav.adminChannels',
  'admin-usage': 'nav.adminUsage',
  'admin-ops': 'nav.adminOps',
  'admin-payment-dashboard': 'nav.adminPayment',
  'admin-orders': 'nav.adminOrders',
  'admin-settings': 'nav.adminSettings',
}

const featureDescriptionKeys: Record<string, string> = {
  dashboard: 'workspace.descriptions.dashboard',
  keys: 'workspace.descriptions.keys',
  usage: 'workspace.descriptions.usage',
  profile: 'workspace.descriptions.profile',
  'admin-users': 'workspace.descriptions.adminUsers',
  'admin-accounts': 'workspace.descriptions.adminAccounts',
  'admin-groups': 'workspace.descriptions.adminGroups',
  'admin-channels': 'workspace.descriptions.adminChannels',
  'admin-usage': 'workspace.descriptions.adminUsage',
  'admin-ops': 'workspace.descriptions.adminOps',
  'admin-payment-dashboard': 'workspace.descriptions.adminPayment',
  'admin-orders': 'workspace.descriptions.adminOrders',
  'admin-settings': 'workspace.descriptions.adminSettings',
}

const title = computed(() => {
  const key = featureTitleKeys[feature.value]
  return key ? t(key) : String(route.meta.title || 'Workspace')
})

const pageDescription = computed(() => {
  const key = featureDescriptionKeys[feature.value]
  return key ? t(key) : t('workspace.descriptions.generic')
})

async function load() {
  error.value = ''
  if (previewMode) {
    if (isDashboard.value) stats.value = { total_api_keys: 4, active_api_keys: 3, total_requests: 12480, total_tokens: 8294000, total_actual_cost: 18.72, today_requests: 842, today_tokens: 612340, today_actual_cost: 1.94 }
    if (isKeys.value) keys.value = [{ id: 1, name: 'Production', key: 'sk-••••••••9F2A', status: 'active', created_at: '2026-09-01' }, { id: 2, name: 'Development', key: 'sk-••••••••71CD', status: 'active', created_at: '2026-08-28' }]
    if (isUsage.value) usage.value = [{ id: 1, model: 'gpt-5.6', endpoint: '/v1/responses', total_tokens: 18420, actual_cost: 0.082, created_at: '2026-09-06 14:20' }, { id: 2, model: 'claude-sonnet', endpoint: '/v1/messages', total_tokens: 9820, actual_cost: 0.051, created_at: '2026-09-06 14:12' }]
    return
  }
  if (!isDashboard.value && !isKeys.value && !isUsage.value) return
  loading.value = true
  try {
    if (isDashboard.value) stats.value = (await api.get<DashboardStats>('/usage/dashboard/stats')).data
    if (isKeys.value) {
      const data = (await api.get<{ items?: ApiKeyRow[] } | ApiKeyRow[]>('/keys', { params: { page: 1, page_size: 50 } })).data
      keys.value = Array.isArray(data) ? data : (data.items || [])
    }
    if (isUsage.value) {
      const data = (await api.get<{ items?: UsageRow[] } | UsageRow[]>('/usage', { params: { page: 1, page_size: 30 } })).data
      usage.value = Array.isArray(data) ? data : (data.items || [])
    }
  } catch (caught) { error.value = getErrorMessage(caught) } finally { loading.value = false }
}

async function createKey() {
  const keyName = newKeyName.value.trim()
  if (!keyName) return

  if (previewMode) {
    keys.value.unshift({ id: Date.now(), name: keyName, key: 'sk-preview-new', status: 'active', created_at: new Date().toISOString().slice(0, 10) })
    newKeyName.value = ''
    pushNotification({
      title: t('workspace.keyCreatedTitle'),
      message: t('workspace.keyCreatedMessage', { name: keyName }),
      tone: 'success',
    })
    return
  }

  loading.value = true
  try {
    const created = (await api.post<ApiKeyRow>('/keys', { name: keyName })).data
    keys.value.unshift(created)
    newKeyName.value = ''
    pushNotification({
      title: t('workspace.keyCreatedTitle'),
      message: t('workspace.keyCreatedMessage', { name: keyName }),
      tone: 'success',
    })
  } catch (caught) { error.value = getErrorMessage(caught) } finally { loading.value = false }
}

async function removeKey(id: number) {
  if (!window.confirm(t('workspace.confirmDelete'))) return
  if (!previewMode) await api.delete(`/keys/${id}`)
  keys.value = keys.value.filter((item) => item.id !== id)
  pushNotification({
    title: t('workspace.keyDeletedTitle'),
    message: t('workspace.keyDeletedMessage'),
    tone: 'info',
  })
}

watch(() => route.fullPath, () => void load())
onMounted(() => void load())
</script>

<template>
  <section class="workspace-page">
    <header class="page-heading">
      <div>
        <h1>{{ title }}</h1>
        <p>{{ pageDescription }}</p>
      </div>
      <button v-if="isDashboard || isKeys || isUsage" class="ghost-button" type="button" :disabled="loading" @click="load">{{ loading ? t('workspace.refreshing') : t('workspace.refresh') }}</button>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <template v-if="isDashboard">
      <section class="glass account-summary">
        <div class="summary-balance">
          <span>{{ t('workspace.availableBalance') }}</span>
          <strong>${{ accountBalance.toFixed(2) }}</strong>
          <small>{{ t('workspace.keyAvailability', { active: stats?.active_api_keys || 0, total: stats?.total_api_keys || 0 }) }}</small>
        </div>
        <div class="summary-meta">
          <span>API Endpoint</span>
          <strong>https://api.smirel.com/v1</strong>
          <small>OpenAI compatible</small>
        </div>
        <RouterLink to="/purchase" class="primary-button">{{ t('workspace.purchase') }}</RouterLink>
      </section>

      <div class="metric-grid metric-grid-three">
        <article class="glass metric-card"><span>{{ t('workspace.todayRequests') }}</span><strong>{{ Number(stats?.today_requests || 0).toLocaleString() }}</strong></article>
        <article class="glass metric-card"><span>{{ t('workspace.todayTokens') }}</span><strong>{{ Number(stats?.today_tokens || 0).toLocaleString() }}</strong></article>
        <article class="glass metric-card"><span>{{ t('workspace.todayCost') }}</span><strong>${{ Number(stats?.today_actual_cost || 0).toFixed(3) }}</strong></article>
      </div>
    </template>

    <template v-else-if="isKeys">
      <section class="glass action-strip">
        <input v-model="newKeyName" :aria-label="`${t('workspace.key')} ${t('workspace.name')}`" :placeholder="t('workspace.keyNamePlaceholder')" @keydown.enter="createKey" />
        <button class="primary-button" type="button" :disabled="loading" @click="createKey">{{ t('workspace.createKey') }}</button>
      </section>
      <div class="glass data-table">
        <div class="table-toolbar">
          <div><strong>API Keys</strong><span class="table-count">{{ keys.length }}</span></div>
        </div>
        <div class="table-head"><span>{{ t('workspace.name') }}</span><span>{{ t('workspace.key') }}</span><span>{{ t('workspace.status') }}</span><span>{{ t('workspace.createdAt') }}</span><span></span></div>
        <div v-for="item in keys" :key="item.id" class="table-row">
          <strong>{{ item.name || `Key #${item.id}` }}</strong>
          <code>{{ item.key || '••••••••' }}</code>
          <span><i class="status-dot"></i>{{ item.status || 'active' }}</span>
          <span>{{ item.created_at || '—' }}</span>
          <button type="button" @click="removeKey(item.id)">{{ t('workspace.delete') }}</button>
        </div>
        <p v-if="!keys.length && !loading" class="empty-state">{{ t('workspace.noKeys') }}</p>
      </div>
    </template>

    <template v-else-if="isUsage">
      <section class="glass table-toolbar standalone-toolbar">
        <div><strong>{{ t('workspace.recentRequests') }}</strong><span class="table-count">{{ usage.length }}</span></div>
        <div class="usage-total"><span>{{ visibleUsageTokens.toLocaleString() }} Tokens</span><span>${{ visibleUsageCost.toFixed(4) }}</span></div>
      </section>
      <div class="glass data-table usage-table">
        <div class="table-head"><span>{{ t('workspace.time') }}</span><span>{{ t('workspace.model') }}</span><span>{{ t('workspace.endpoint') }}</span><span>{{ t('workspace.token') }}</span><span>{{ t('workspace.cost') }}</span></div>
        <div v-for="(item, index) in usage" :key="item.id || index" class="table-row">
          <span>{{ item.created_at || '—' }}</span>
          <strong>{{ item.model || '—' }}</strong>
          <code>{{ item.endpoint || '—' }}</code>
          <span>{{ Number(item.total_tokens || 0).toLocaleString() }}</span>
          <span>${{ Number(item.actual_cost || 0).toFixed(4) }}</span>
        </div>
        <p v-if="!usage.length && !loading" class="empty-state">{{ t('workspace.noUsage') }}</p>
      </div>
    </template>

    <template v-else-if="isProfile">
      <section class="glass profile-panel">
        <div class="profile-avatar">{{ (state.user?.username || state.user?.email || 'S').slice(0,1).toUpperCase() }}</div>
        <div class="profile-copy"><h2>{{ state.user?.username || 'Smirel Account' }}</h2><p>{{ state.user?.email }}</p></div>
        <dl>
          <div><dt>{{ t('workspace.role') }}</dt><dd>{{ state.user?.role === 'admin' ? t('shell.roleAdmin') : t('workspace.user') }}</dd></div>
          <div><dt>{{ t('workspace.status') }}</dt><dd>{{ state.user?.status || 'active' }}</dd></div>
          <div><dt>{{ t('workspace.availableBalance') }}</dt><dd>${{ accountBalance.toFixed(2) }}</dd></div>
        </dl>
      </section>
    </template>

    <template v-else>
      <section class="glass module-panel">
        <h2>{{ title }}</h2>
        <p>{{ t('workspace.modulePending') }}</p>
      </section>
    </template>
  </section>
</template>
