<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, getErrorMessage, previewMode } from '../core/api'

interface AdminStats {
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

interface Snapshot {
  generated_at?: string
  stats?: AdminStats
}

const loading = ref(false)
const error = ref('')
const snapshot = ref<Snapshot | null>(null)
const stats = computed<AdminStats>(() => snapshot.value?.stats ?? {})

function compact(value: unknown) {
  const n = Number(value || 0)
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`
  return n.toLocaleString()
}

function money(value: unknown) {
  return `$${Number(value || 0).toFixed(2)}`
}

function duration(value: unknown) {
  const n = Number(value || 0)
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${Math.round(n)}ms`
}

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
      snapshot.value = (await api.get<Snapshot>('/admin/dashboard/snapshot-v2', {
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
  <section class="workspace-page admin-overview">
    <header class="page-heading">
      <div><h1>控制台</h1><p>查看平台最重要的运行数据。</p></div>
      <button class="ghost-button" type="button" :disabled="loading" @click="load">{{ loading ? '刷新中…' : '刷新' }}</button>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <div class="admin-summary-grid">
      <article class="glass metric-card">
        <span>活跃用户</span>
        <strong>{{ compact(stats.active_users) }}</strong>
        <small>共 {{ compact(stats.total_users) }} 位用户</small>
      </article>
      <article class="glass metric-card">
        <span>上游账户</span>
        <strong>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }}</strong>
        <small>当前可用账户</small>
      </article>
      <article class="glass metric-card">
        <span>今日请求</span>
        <strong>{{ compact(stats.today_requests) }}</strong>
        <small>RPM {{ compact(stats.rpm) }}</small>
      </article>
    </div>

    <section class="glass admin-status-panel">
      <header>
        <div><h2>运行状态</h2><p>核心服务当前正常。</p></div>
        <span class="service-ok"><i></i>正常</span>
      </header>
      <div class="admin-status-grid">
        <div><span>今日 Token</span><strong>{{ compact(stats.today_tokens) }}</strong></div>
        <div><span>今日成本</span><strong>{{ money(stats.today_actual_cost) }}</strong></div>
        <div><span>平均响应</span><strong>{{ duration(stats.average_duration_ms) }}</strong></div>
      </div>
      <footer>API Keys {{ compact(stats.total_api_keys) }} · 数据实时更新</footer>
    </section>
  </section>
</template>
