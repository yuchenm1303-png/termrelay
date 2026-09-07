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
    <header class="page-heading admin-page-heading">
      <div>
        <h1>控制台</h1>
        <p>平台核心数据与服务状态。</p>
      </div>
      <button class="ghost-button admin-refresh" type="button" :disabled="loading" @click="load">
        {{ loading ? '刷新中…' : '刷新数据' }}
      </button>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <div class="admin-summary-grid">
      <article class="glass metric-card admin-metric-card">
        <span>活跃用户</span>
        <strong>{{ compact(stats.active_users) }}</strong>
        <small>总用户 {{ compact(stats.total_users) }}</small>
      </article>
      <article class="glass metric-card admin-metric-card">
        <span>上游账户</span>
        <strong>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }}</strong>
        <small>{{ stats.active_accounts || 0 }} 个账户可用</small>
      </article>
      <article class="glass metric-card admin-metric-card">
        <span>今日请求</span>
        <strong>{{ compact(stats.today_requests) }}</strong>
        <small>当前 RPM {{ compact(stats.rpm) }}</small>
      </article>
    </div>

    <section class="glass admin-health-strip">
      <div class="admin-health-summary">
        <span>服务状态</span>
        <strong><i></i>运行正常</strong>
        <small>核心服务可用</small>
      </div>
      <div class="admin-health-stat">
        <span>今日 Token</span>
        <strong>{{ compact(stats.today_tokens) }}</strong>
      </div>
      <div class="admin-health-stat">
        <span>今日成本</span>
        <strong>{{ money(stats.today_actual_cost) }}</strong>
      </div>
      <div class="admin-health-stat">
        <span>平均响应</span>
        <strong>{{ duration(stats.average_duration_ms) }}</strong>
      </div>
      <div class="admin-health-stat">
        <span>API Keys</span>
        <strong>{{ compact(stats.total_api_keys) }}</strong>
      </div>
    </section>
  </section>
</template>

<style scoped>
.admin-overview {
  max-width: 1240px;
  margin: 0 auto;
}

.admin-page-heading {
  min-height: auto;
  margin-bottom: 24px;
  align-items: center;
}

.admin-page-heading h1 {
  font-size: 2.12rem;
  font-weight: 690;
  letter-spacing: -.04em;
}

.admin-page-heading p {
  margin-top: 7px;
  color: #7d858f;
  font-size: .86rem;
}

.admin-refresh {
  min-height: 36px;
  padding: 0 13px;
  border-radius: 8px;
  background: #0f1115;
  font-size: .78rem;
}

.admin-summary-grid {
  gap: 12px;
}

.admin-metric-card {
  min-height: 118px;
  padding: 20px 22px;
  border-color: #22262d;
  border-radius: 11px;
  background: #0f1115;
}

.admin-metric-card span {
  color: #868e98;
  font-size: .80rem;
  font-weight: 590;
}

.admin-metric-card strong {
  margin-top: 9px;
  color: #f5f7f9;
  font-size: 1.86rem;
  font-weight: 680;
  letter-spacing: -.035em;
}

.admin-metric-card small {
  margin-top: 7px;
  color: #666f7a;
  font-size: .72rem;
}

.admin-health-strip {
  min-height: 112px;
  margin-top: 12px;
  padding: 0;
  border-color: #22262d;
  border-radius: 11px;
  background: #0f1115;
  display: grid;
  grid-template-columns: minmax(190px, 1.25fr) repeat(4, minmax(140px, 1fr));
  overflow: hidden;
}

.admin-health-summary,
.admin-health-stat {
  min-width: 0;
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.admin-health-stat {
  border-left: 1px solid #252930;
}

.admin-health-summary > span,
.admin-health-stat span {
  color: #747d88;
  font-size: .75rem;
  font-weight: 560;
}

.admin-health-summary strong {
  margin-top: 9px;
  color: #dff7ed;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1rem;
  font-weight: 640;
}

.admin-health-summary strong i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #42ce99;
  box-shadow: 0 0 0 4px rgba(66, 206, 153, .08);
}

.admin-health-summary small {
  margin-top: 7px;
  color: #626b76;
  font-size: .70rem;
}

.admin-health-stat strong {
  margin-top: 9px;
  color: #e9edf1;
  font-size: 1.14rem;
  font-weight: 650;
  letter-spacing: -.02em;
}

@media (max-width: 1080px) {
  .admin-health-strip {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .admin-health-summary {
    grid-column: span 3;
    border-bottom: 1px solid #252930;
  }

  .admin-health-stat {
    border-left: 1px solid #252930;
  }

  .admin-health-stat:nth-of-type(2) {
    border-left: 0;
  }
}

@media (max-width: 720px) {
  .admin-page-heading {
    align-items: flex-start;
  }

  .admin-page-heading h1 {
    font-size: 1.82rem;
  }

  .admin-health-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .admin-health-summary {
    grid-column: span 2;
  }

  .admin-health-stat:nth-of-type(2),
  .admin-health-stat:nth-of-type(4) {
    border-left: 0;
  }

  .admin-health-stat:nth-of-type(n + 4) {
    border-top: 1px solid #252930;
  }
}

@media (max-width: 540px) {
  .admin-refresh {
    width: 100%;
  }

  .admin-health-strip {
    grid-template-columns: 1fr;
  }

  .admin-health-summary {
    grid-column: auto;
  }

  .admin-health-stat {
    border-left: 0;
    border-top: 1px solid #252930;
  }
}
</style>
