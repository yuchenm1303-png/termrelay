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
      <div><span class="eyebrow">ADMIN CONSOLE</span><h1>Overview</h1><p>平台健康、用户、上游资源、流量与成本。</p></div>
      <button class="ghost-button" type="button" :disabled="loading" @click="load">{{ loading ? '刷新中…' : '刷新数据' }}</button>
    </header>
    <p v-if="error" class="inline-error">{{ error }}</p>
    <div class="admin-grid">
      <section class="glass health-panel">
        <header><span>PLATFORM HEALTH</span><b><i></i>运行正常</b></header>
        <div class="health-hero"><strong>核心资源在线</strong><p>平台请求链路、用户服务和上游账户池保持可用。</p></div>
        <div class="health-row">
          <span>上游账户<strong>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }}</strong></span>
          <span>活跃用户<strong>{{ compact(stats.active_users) }}</strong></span>
          <span>API Keys<strong>{{ compact(stats.total_api_keys) }}</strong></span>
        </div>
      </section>
      <section class="glass today-panel">
        <header><span>TODAY</span><b>今日运行</b></header>
        <div class="metric-grid admin-metrics">
          <article><span>请求</span><strong>{{ compact(stats.today_requests) }}</strong></article>
          <article><span>Token</span><strong>{{ compact(stats.today_tokens) }}</strong></article>
          <article><span>实际成本</span><strong>{{ money(stats.today_actual_cost) }}</strong></article>
          <article><span>平均响应</span><strong>{{ duration(stats.average_duration_ms) }}</strong></article>
        </div>
      </section>
    </div>
    <section class="glass operations-panel">
      <header><div><span>OPERATIONS</span><strong>运行摘要</strong></div><small>RPM {{ compact(stats.rpm) }}</small></header>
      <div class="operations-line"><span></span></div>
      <div class="operations-summary">
        <div><span>用户总数</span><strong>{{ compact(stats.total_users) }}</strong></div>
        <div><span>当前 RPM</span><strong>{{ compact(stats.rpm) }}</strong></div>
        <div><span>上游可用率</span><strong>{{ Number(stats.total_accounts || 0) ? `${((Number(stats.active_accounts || 0) / Number(stats.total_accounts || 1)) * 100).toFixed(1)}%` : '—' }}</strong></div>
        <div><span>快照</span><strong>{{ snapshot?.generated_at ? 'LIVE' : '—' }}</strong></div>
      </div>
    </section>
  </section>
</template>
