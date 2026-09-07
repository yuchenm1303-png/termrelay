<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, getErrorMessage, previewMode } from '../core/api'
import { useSession } from '../core/session'

interface ApiKeyRow { id: number; name?: string; key?: string; status?: string; created_at?: string; [key: string]: unknown }
interface UsageRow { id?: number; model?: string; endpoint?: string; total_tokens?: number; actual_cost?: number; created_at?: string; [key: string]: unknown }
interface DashboardStats { total_api_keys?: number; active_api_keys?: number; total_requests?: number; total_tokens?: number; total_actual_cost?: number; today_requests?: number; today_tokens?: number; today_actual_cost?: number; [key: string]: unknown }

const route = useRoute()
const { state } = useSession()
const feature = computed(() => String(route.meta.feature || 'module'))
const title = computed(() => String(route.meta.title || 'Workspace'))
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

const pageDescription = computed(() => {
  if (isDashboard.value) return '查看账户余额、API Key 和今日调用情况。'
  if (isKeys.value) return '创建和管理用于调用 Smirel API 的访问密钥。'
  if (isUsage.value) return '查看最近请求、Token 与费用。'
  if (isProfile.value) return '查看当前账户信息。'
  return '管理当前 Smirel 服务。'
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
  if (!newKeyName.value.trim()) return
  if (previewMode) {
    keys.value.unshift({ id: Date.now(), name: newKeyName.value.trim(), key: 'sk-preview-new', status: 'active', created_at: new Date().toISOString().slice(0, 10) })
    newKeyName.value = ''
    return
  }
  loading.value = true
  try {
    const created = (await api.post<ApiKeyRow>('/keys', { name: newKeyName.value.trim() })).data
    keys.value.unshift(created)
    newKeyName.value = ''
  } catch (caught) { error.value = getErrorMessage(caught) } finally { loading.value = false }
}

async function removeKey(id: number) {
  if (!window.confirm('确认删除这把 API Key？')) return
  if (!previewMode) await api.delete(`/keys/${id}`)
  keys.value = keys.value.filter((item) => item.id !== id)
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
      <button v-if="isDashboard || isKeys || isUsage" class="ghost-button" type="button" :disabled="loading" @click="load">{{ loading ? '刷新中…' : '刷新' }}</button>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <template v-if="isDashboard">
      <section class="glass account-summary">
        <div class="summary-balance">
          <span>可用余额</span>
          <strong>${{ accountBalance.toFixed(2) }}</strong>
          <small>{{ stats?.active_api_keys || 0 }} / {{ stats?.total_api_keys || 0 }} 个 API Key 可用</small>
        </div>
        <div class="summary-meta">
          <span>API Endpoint</span>
          <strong>https://api.smirel.com/v1</strong>
          <small>OpenAI compatible</small>
        </div>
        <RouterLink to="/purchase" class="primary-button">购买服务</RouterLink>
      </section>

      <div class="metric-grid metric-grid-three">
        <article class="glass metric-card"><span>今日请求</span><strong>{{ Number(stats?.today_requests || 0).toLocaleString() }}</strong></article>
        <article class="glass metric-card"><span>今日 Token</span><strong>{{ Number(stats?.today_tokens || 0).toLocaleString() }}</strong></article>
        <article class="glass metric-card"><span>今日费用</span><strong>${{ Number(stats?.today_actual_cost || 0).toFixed(3) }}</strong></article>
      </div>
    </template>

    <template v-else-if="isKeys">
      <section class="glass action-strip">
        <input v-model="newKeyName" aria-label="API Key 名称" placeholder="密钥名称，例如 Production" @keydown.enter="createKey" />
        <button class="primary-button" type="button" :disabled="loading" @click="createKey">创建 API Key</button>
      </section>
      <div class="glass data-table">
        <div class="table-toolbar">
          <div><strong>API Keys</strong><span class="table-count">{{ keys.length }}</span></div>
        </div>
        <div class="table-head"><span>名称</span><span>密钥</span><span>状态</span><span>创建时间</span><span></span></div>
        <div v-for="item in keys" :key="item.id" class="table-row">
          <strong>{{ item.name || `Key #${item.id}` }}</strong>
          <code>{{ item.key || '••••••••' }}</code>
          <span><i class="status-dot"></i>{{ item.status || 'active' }}</span>
          <span>{{ item.created_at || '—' }}</span>
          <button type="button" @click="removeKey(item.id)">删除</button>
        </div>
        <p v-if="!keys.length && !loading" class="empty-state">还没有 API Key。</p>
      </div>
    </template>

    <template v-else-if="isUsage">
      <section class="glass table-toolbar standalone-toolbar">
        <div><strong>最近请求</strong><span class="table-count">{{ usage.length }}</span></div>
        <div class="usage-total"><span>{{ visibleUsageTokens.toLocaleString() }} Tokens</span><span>${{ visibleUsageCost.toFixed(4) }}</span></div>
      </section>
      <div class="glass data-table usage-table">
        <div class="table-head"><span>时间</span><span>模型</span><span>Endpoint</span><span>Token</span><span>费用</span></div>
        <div v-for="(item, index) in usage" :key="item.id || index" class="table-row">
          <span>{{ item.created_at || '—' }}</span>
          <strong>{{ item.model || '—' }}</strong>
          <code>{{ item.endpoint || '—' }}</code>
          <span>{{ Number(item.total_tokens || 0).toLocaleString() }}</span>
          <span>${{ Number(item.actual_cost || 0).toFixed(4) }}</span>
        </div>
        <p v-if="!usage.length && !loading" class="empty-state">暂无用量记录。</p>
      </div>
    </template>

    <template v-else-if="isProfile">
      <section class="glass profile-panel">
        <div class="profile-avatar">{{ (state.user?.username || state.user?.email || 'S').slice(0,1).toUpperCase() }}</div>
        <div class="profile-copy"><h2>{{ state.user?.username || 'Smirel Account' }}</h2><p>{{ state.user?.email }}</p></div>
        <dl>
          <div><dt>角色</dt><dd>{{ state.user?.role === 'admin' ? '管理员' : '用户' }}</dd></div>
          <div><dt>状态</dt><dd>{{ state.user?.status || 'active' }}</dd></div>
          <div><dt>可用余额</dt><dd>${{ accountBalance.toFixed(2) }}</dd></div>
        </dl>
      </section>
    </template>

    <template v-else>
      <section class="glass module-panel">
        <h2>{{ title }}</h2>
        <p>该功能正在接入商业版控制台。</p>
      </section>
    </template>
  </section>
</template>
