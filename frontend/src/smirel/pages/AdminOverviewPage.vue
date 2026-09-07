<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import WorkspaceNavIcon from '../components/WorkspaceNavIcon.vue'
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

function percent(part: unknown, total: unknown) {
  const denominator = Number(total || 0)
  if (!denominator) return 0
  return Math.max(0, Math.min(100, Math.round((Number(part || 0) / denominator) * 100)))
}

const activeUserRate = computed(() => percent(stats.value.active_users, stats.value.total_users))
const accountAvailabilityRate = computed(() => percent(stats.value.active_accounts, stats.value.total_accounts))
const generatedLabel = computed(() => {
  const value = snapshot.value?.generated_at
  if (!value) return '等待首次同步'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '刚刚更新'
  return `更新于 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
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
      <div class="admin-heading-copy">
        <div class="admin-heading-eyebrow"><i></i><span>SMIREL OPERATIONS</span></div>
        <h1>控制台</h1>
        <p>平台核心运行数据、资源状态与实时调用概览。</p>
      </div>

      <div class="admin-heading-actions">
        <div class="admin-snapshot-meta" :class="{ ready: snapshot }">
          <span class="admin-snapshot-icon"><WorkspaceNavIcon name="activity" /></span>
          <span><strong>{{ snapshot ? '数据已同步' : '等待数据' }}</strong><small>{{ generatedLabel }}</small></span>
        </div>
        <button class="ghost-button admin-refresh" type="button" :disabled="loading" @click="load">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M20 6v5h-5M4 18v-5h5" />
            <path d="M6.1 9a7 7 0 0 1 11.5-2.4L20 11M4 13l2.4 4.4A7 7 0 0 0 18 15" />
          </svg>
          <span>{{ loading ? '刷新中…' : '刷新数据' }}</span>
        </button>
      </div>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <div class="admin-summary-grid">
      <article class="glass metric-card admin-metric-card admin-metric-users">
        <div class="admin-metric-head">
          <span>活跃用户</span>
          <span class="admin-metric-icon"><WorkspaceNavIcon name="users" /></span>
        </div>
        <strong>{{ compact(stats.active_users) }}</strong>
        <footer>
          <span>总用户 {{ compact(stats.total_users) }}</span>
          <b>{{ activeUserRate }}% 活跃</b>
        </footer>
      </article>

      <article class="glass metric-card admin-metric-card admin-metric-upstream">
        <div class="admin-metric-head">
          <span>上游账户</span>
          <span class="admin-metric-icon"><WorkspaceNavIcon name="server" /></span>
        </div>
        <strong>{{ stats.active_accounts || 0 }} / {{ stats.total_accounts || 0 }}</strong>
        <footer>
          <span>{{ stats.active_accounts || 0 }} 个账户可用</span>
          <b>{{ accountAvailabilityRate }}% 可用</b>
        </footer>
      </article>

      <article class="glass metric-card admin-metric-card admin-metric-requests">
        <div class="admin-metric-head">
          <span>今日请求</span>
          <span class="admin-metric-icon"><WorkspaceNavIcon name="activity" /></span>
        </div>
        <strong>{{ compact(stats.today_requests) }}</strong>
        <footer>
          <span>实时调用负载</span>
          <b>RPM {{ compact(stats.rpm) }}</b>
        </footer>
      </article>
    </div>

    <section class="glass admin-ops-panel">
      <header class="admin-ops-head">
        <div>
          <span class="admin-ops-kicker">运行概览</span>
          <strong>今日资源消耗与服务性能</strong>
        </div>
        <div class="admin-ops-state" :class="{ ready: snapshot }"><i></i>{{ snapshot ? '快照已同步' : '等待同步' }}</div>
      </header>

      <div class="admin-health-grid">
        <div class="admin-health-stat">
          <span class="admin-health-icon"><WorkspaceNavIcon name="chart" /></span>
          <div>
            <span>今日 Token</span>
            <strong>{{ compact(stats.today_tokens) }}</strong>
            <small>TODAY</small>
          </div>
        </div>

        <div class="admin-health-stat">
          <span class="admin-health-icon"><WorkspaceNavIcon name="wallet" /></span>
          <div>
            <span>今日成本</span>
            <strong>{{ money(stats.today_actual_cost) }}</strong>
            <small>USD</small>
          </div>
        </div>

        <div class="admin-health-stat">
          <span class="admin-health-icon"><WorkspaceNavIcon name="activity" /></span>
          <div>
            <span>平均响应</span>
            <strong>{{ duration(stats.average_duration_ms) }}</strong>
            <small>LATENCY</small>
          </div>
        </div>

        <div class="admin-health-stat">
          <span class="admin-health-icon"><WorkspaceNavIcon name="key" /></span>
          <div>
            <span>API Keys</span>
            <strong>{{ compact(stats.total_api_keys) }}</strong>
            <small>TOTAL</small>
          </div>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.admin-overview {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
}

.admin-page-heading {
  min-height: auto;
  margin-bottom: 30px;
  align-items: flex-end;
}

.admin-heading-copy {
  min-width: 0;
}

.admin-heading-eyebrow {
  margin-bottom: 13px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #66717e;
  font: 700 .66rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .11em;
}

.admin-heading-eyebrow i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #42ce99;
  box-shadow: 0 0 0 4px rgba(66, 206, 153, .07);
}

.admin-page-heading h1 {
  margin-top: 0;
  color: #f6f7f9;
  font-size: 2.3rem;
  font-weight: 700;
  line-height: 1.04;
  letter-spacing: -.045em;
}

.admin-page-heading p {
  margin-top: 10px;
  color: #78818d;
  font-size: .88rem;
}

.admin-heading-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-snapshot-meta {
  min-height: 42px;
  padding: 0 12px 0 8px;
  border: 1px solid #242932;
  border-radius: 9px;
  display: flex;
  align-items: center;
  gap: 9px;
  background: #0d0f13;
}

.admin-snapshot-icon {
  width: 28px;
  height: 28px;
  border: 1px solid #28313a;
  border-radius: 7px;
  display: grid;
  place-items: center;
  color: #74808c;
  background: #12151a;
}

.admin-snapshot-meta.ready .admin-snapshot-icon {
  border-color: #234236;
  color: #57c99c;
  background: #0e1915;
}

.admin-snapshot-icon :deep(.workspace-nav-icon) {
  width: 14px;
  height: 14px;
}

.admin-snapshot-meta > span:last-child {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.admin-snapshot-meta strong {
  color: #b8bec6;
  font-size: .7rem;
  font-weight: 620;
}

.admin-snapshot-meta small {
  color: #5f6873;
  font-size: .62rem;
}

.admin-refresh {
  min-height: 42px;
  padding: 0 14px;
  gap: 8px;
  border-radius: 9px;
  background: #0f1115;
  font-size: .76rem;
}

.admin-refresh svg {
  width: 15px;
  height: 15px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.admin-refresh:disabled svg {
  animation: admin-refresh-spin .9s linear infinite;
}

@keyframes admin-refresh-spin {
  to { transform: rotate(360deg); }
}

.admin-summary-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.admin-metric-card {
  --metric-accent: #4da9eb;
  position: relative;
  isolation: isolate;
  min-height: 156px;
  padding: 22px 23px 20px;
  overflow: hidden;
  border-color: #222730;
  border-radius: 13px;
  background: #0f1115;
}

.admin-metric-card::before {
  content: '';
  position: absolute;
  inset: 0 auto auto 0;
  z-index: -1;
  width: 44%;
  height: 1px;
  background: linear-gradient(90deg, var(--metric-accent), transparent);
  opacity: .7;
}

.admin-metric-card::after {
  content: '';
  position: absolute;
  z-index: -1;
  top: -52px;
  right: -42px;
  width: 150px;
  height: 150px;
  border-radius: 50%;
  background: radial-gradient(circle, color-mix(in srgb, var(--metric-accent) 10%, transparent), transparent 68%);
  pointer-events: none;
}

.admin-metric-upstream {
  --metric-accent: #49c89a;
}

.admin-metric-requests {
  --metric-accent: #789df4;
}

.admin-metric-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.admin-metric-head > span:first-child {
  color: #8b939e;
  font-size: .8rem;
  font-weight: 590;
}

.admin-metric-icon {
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border: 1px solid color-mix(in srgb, var(--metric-accent) 28%, #252a31);
  border-radius: 9px;
  display: grid;
  place-items: center;
  color: var(--metric-accent);
  background: color-mix(in srgb, var(--metric-accent) 7%, #111318);
}

.admin-metric-icon :deep(.workspace-nav-icon) {
  width: 17px;
  height: 17px;
}

.admin-metric-card > strong {
  margin-top: 17px;
  color: #f6f7f9;
  font-size: 2.02rem;
  font-weight: 690;
  line-height: 1;
  letter-spacing: -.04em;
}

.admin-metric-card footer {
  margin-top: auto;
  padding-top: 17px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #66707b;
  font-size: .71rem;
}

.admin-metric-card footer b {
  padding: 4px 7px;
  border: 1px solid #262c34;
  border-radius: 999px;
  color: #88939f;
  background: #11141a;
  font-size: .64rem;
  font-weight: 600;
  white-space: nowrap;
}

.admin-ops-panel {
  position: relative;
  margin-top: 14px;
  padding: 0;
  overflow: hidden;
  border-color: #222730;
  border-radius: 13px;
  background: #0d0f13;
}

.admin-ops-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: linear-gradient(rgba(255,255,255,.014) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.014) 1px, transparent 1px);
  background-size: 32px 32px;
  mask-image: linear-gradient(to bottom, rgba(0,0,0,.55), transparent 72%);
}

.admin-ops-head {
  position: relative;
  z-index: 1;
  min-height: 72px;
  padding: 0 22px;
  border-bottom: 1px solid #22272e;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.admin-ops-head > div:first-child {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.admin-ops-kicker {
  color: #68727e;
  font-size: .67rem;
  font-weight: 650;
}

.admin-ops-head strong {
  color: #dce0e5;
  font-size: .84rem;
  font-weight: 620;
}

.admin-ops-state {
  min-height: 28px;
  padding: 0 9px;
  border: 1px solid #2a2f37;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #7b848f;
  background: #111318;
  font-size: .66rem;
  font-weight: 570;
}

.admin-ops-state i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #69727d;
}

.admin-ops-state.ready {
  border-color: #244337;
  color: #79c8a9;
  background: #0f1815;
}

.admin-ops-state.ready i {
  background: #42ce99;
}

.admin-health-grid {
  position: relative;
  z-index: 1;
  min-height: 142px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.admin-health-stat {
  min-width: 0;
  padding: 24px 22px;
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr);
  align-items: center;
  gap: 14px;
}

.admin-health-stat + .admin-health-stat {
  border-left: 1px solid #232830;
}

.admin-health-icon {
  width: 38px;
  height: 38px;
  border: 1px solid #28303a;
  border-radius: 9px;
  display: grid;
  place-items: center;
  color: #72869a;
  background: #11151a;
}

.admin-health-icon :deep(.workspace-nav-icon) {
  width: 17px;
  height: 17px;
}

.admin-health-stat > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.admin-health-stat > div > span {
  color: #77818c;
  font-size: .71rem;
  font-weight: 560;
}

.admin-health-stat strong {
  margin-top: 6px;
  color: #edf0f3;
  font-size: 1.22rem;
  font-weight: 650;
  line-height: 1.05;
  letter-spacing: -.025em;
}

.admin-health-stat small {
  margin-top: 7px;
  color: #4f5863;
  font: 650 .56rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .1em;
}

@media (max-width: 1080px) {
  .admin-page-heading {
    align-items: flex-start;
  }

  .admin-heading-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .admin-summary-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .admin-health-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .admin-health-stat:nth-child(odd) {
    border-left: 0;
  }

  .admin-health-stat:nth-child(n + 3) {
    border-top: 1px solid #232830;
  }
}

@media (max-width: 820px) {
  .admin-page-heading {
    flex-direction: column;
    gap: 18px;
  }

  .admin-heading-actions {
    width: 100%;
    flex-direction: row;
  }

  .admin-snapshot-meta {
    flex: 1;
  }

  .admin-summary-grid {
    grid-template-columns: 1fr;
  }

  .admin-metric-card {
    min-height: 140px;
  }
}

@media (max-width: 620px) {
  .admin-page-heading h1 {
    font-size: 1.9rem;
  }

  .admin-heading-actions {
    flex-direction: column;
  }

  .admin-refresh {
    width: 100%;
  }

  .admin-ops-head {
    min-height: 86px;
    align-items: flex-start;
    flex-direction: column;
    justify-content: center;
    gap: 10px;
  }

  .admin-health-grid {
    grid-template-columns: 1fr;
  }

  .admin-health-stat,
  .admin-health-stat:nth-child(odd) {
    border-left: 0;
  }

  .admin-health-stat + .admin-health-stat,
  .admin-health-stat:nth-child(n + 3) {
    border-top: 1px solid #232830;
  }
}

@media (prefers-reduced-motion: reduce) {
  .admin-refresh:disabled svg {
    animation: none;
  }
}
</style>
