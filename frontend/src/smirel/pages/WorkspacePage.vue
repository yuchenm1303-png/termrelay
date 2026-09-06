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
    <header class="page-heading"><div><span class="eyebrow">{{ feature.toUpperCase() }}</span><h1>{{ title }}</h1><p v-if="isDashboard">账户概览、请求量、Token 与成本。</p><p v-else-if="isKeys">管理调用 Smirel API 的项目凭证。</p><p v-else-if="isUsage">查看最近请求、模型、Token 和费用。</p><p v-else-if="isProfile">管理当前账户信息。</p><p v-else>Smirel 独立前端功能模块。</p></div><button v-if="isDashboard || isKeys || isUsage" class="ghost-button" type="button" :disabled="loading" @click="load">{{ loading ? '刷新中…' : '刷新' }}</button></header>
    <p v-if="error" class="inline-error">{{ error }}</p>

    <template v-if="isDashboard">
      <div class="metric-grid">
        <article class="glass metric-card"><span>今日请求</span><strong>{{ Number(stats?.today_requests || 0).toLocaleString() }}</strong><small>requests</small></article>
        <article class="glass metric-card"><span>今日 Token</span><strong>{{ Number(stats?.today_tokens || 0).toLocaleString() }}</strong><small>tokens</small></article>
        <article class="glass metric-card"><span>今日费用</span><strong>${{ Number(stats?.today_actual_cost || 0).toFixed(3) }}</strong><small>actual cost</small></article>
        <article class="glass metric-card"><span>有效密钥</span><strong>{{ stats?.active_api_keys || 0 }} / {{ stats?.total_api_keys || 0 }}</strong><small>API keys</small></article>
      </div>
      <div class="glass information-panel"><span class="eyebrow">WORKSPACE</span><h2>统一入口，独立管理。</h2><p>Base URL 固定为 <code>https://api.smirel.com/v1</code>。每个项目使用独立 API Key，方便停用、追踪和分账。</p></div>
    </template>

    <template v-else-if="isKeys">
      <div class="glass action-strip"><input v-model="newKeyName" placeholder="新密钥名称，例如 Production" @keydown.enter="createKey" /><button class="primary-button" type="button" :disabled="loading" @click="createKey">创建 API Key</button></div>
      <div class="glass data-table"><div class="table-head"><span>名称</span><span>密钥</span><span>状态</span><span>创建时间</span><span></span></div><div v-for="item in keys" :key="item.id" class="table-row"><strong>{{ item.name || `Key #${item.id}` }}</strong><code>{{ item.key || '••••••••' }}</code><span><i class="status-dot"></i>{{ item.status || 'active' }}</span><span>{{ item.created_at || '—' }}</span><button type="button" @click="removeKey(item.id)">删除</button></div><p v-if="!keys.length && !loading" class="empty-state">还没有 API Key。</p></div>
    </template>

    <template v-else-if="isUsage">
      <div class="glass data-table usage-table"><div class="table-head"><span>时间</span><span>模型</span><span>Endpoint</span><span>Token</span><span>费用</span></div><div v-for="(item, index) in usage" :key="item.id || index" class="table-row"><span>{{ item.created_at || '—' }}</span><strong>{{ item.model || '—' }}</strong><code>{{ item.endpoint || '—' }}</code><span>{{ Number(item.total_tokens || 0).toLocaleString() }}</span><span>${{ Number(item.actual_cost || 0).toFixed(4) }}</span></div><p v-if="!usage.length && !loading" class="empty-state">暂无用量记录。</p></div>
    </template>

    <template v-else-if="isProfile">
      <div class="glass profile-panel"><div class="profile-avatar">{{ (state.user?.username || state.user?.email || 'S').slice(0,1).toUpperCase() }}</div><div><span>ACCOUNT</span><h2>{{ state.user?.username || 'Smirel Account' }}</h2><p>{{ state.user?.email }}</p></div><dl><div><dt>角色</dt><dd>{{ state.user?.role === 'admin' ? '管理员' : '用户' }}</dd></div><div><dt>状态</dt><dd>{{ state.user?.status || 'active' }}</dd></div><div><dt>余额</dt><dd>${{ Number(state.user?.balance || 0).toFixed(2) }}</dd></div></dl></div>
    </template>

    <template v-else>
      <div class="glass information-panel module-panel"><span class="eyebrow">SMIREL MODULE</span><h2>{{ title }}</h2><p>这个入口已经完全脱离旧前端，当前只连接 Smirel 自己的页面层。后端业务能力保持原样，具体操作面板会继续在这里补齐，而不会重新接回任何 legacy UI。</p><div class="module-contract"><span>UI</span><strong>Smirel Native</strong><span>Backend</span><strong>TermRelay API</strong></div></div>
    </template>
  </section>
</template>
