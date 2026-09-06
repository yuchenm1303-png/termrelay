<template>
  <div class="sw2-admin-overview">
    <section v-if="loading && !stats" class="spg-surface sw2-admin-loading">
      <span class="sw2-admin-loading-mark" aria-hidden="true"></span>
      <strong>正在读取平台运行数据</strong>
      <p>同步账户池、请求与结算快照…</p>
    </section>

    <section v-else-if="error && !stats" class="spg-surface sw2-admin-error">
      <strong>平台数据暂时没有加载成功</strong>
      <p>当前 Overview 无法取得最新快照，可以重新请求一次。</p>
      <button type="button" @click="loadSnapshot">重新加载</button>
    </section>

    <template v-else>
      <div class="sw2-admin-toolbar">
        <div class="sw2-admin-snapshot-label">
          <span class="sw2-admin-live-dot" aria-hidden="true"></span>
          <span>LIVE SNAPSHOT</span>
          <strong>运营快照</strong>
        </div>
        <div class="sw2-admin-toolbar-actions">
          <span v-if="lastUpdated">数据更新于 {{ lastUpdated }}</span>
          <button class="sw2-admin-refresh" type="button" :disabled="loading || opsLoading" @click="loadSnapshot">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M20 7v5h-5M4 17v-5h5M18.1 9A7 7 0 006.7 6.4L4 9m2 6a7 7 0 0011.3 2.6L20 15" />
            </svg>
            {{ loading || opsLoading ? '刷新中…' : '刷新数据' }}
          </button>
        </div>
      </div>

      <div class="sw2-admin-top-grid">
        <section class="spg-surface sw2-admin-module sw2-admin-health">
          <header class="sw2-module-head">
            <div>
              <span class="sw2-admin-section-label">PLATFORM HEALTH</span>
              <strong>平台健康</strong>
            </div>
            <span class="sw2-admin-health-state" :class="{ 'sw2-admin-health-state--warning': healthHasIssues }">
              <i aria-hidden="true"></i>{{ healthHasIssues ? '需要关注' : '运行正常' }}
            </span>
          </header>

          <div class="sw2-admin-health-hero">
            <div class="sw2-admin-health-orb" :class="{ 'is-warning': healthHasIssues }" aria-hidden="true">
              <svg v-if="healthHasIssues" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4l9 16H3L12 4zm0 5.5v4.5m0 3h.01" />
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 12.5l4.2 4.2L19 7" />
              </svg>
            </div>
            <div class="sw2-admin-health-copy">
              <span>SYSTEM STATUS</span>
              <strong>{{ healthHasIssues ? '上游资源需要检查' : '核心资源运行正常' }}</strong>
              <p>{{ healthDescription }}</p>
            </div>
          </div>

          <div class="sw2-admin-health-grid">
            <div v-for="item in healthMetrics" :key="item.label" class="sw2-admin-health-metric">
              <span>{{ item.label }}</span>
              <strong :class="{ 'is-warning': item.warning }">{{ item.value }}</strong>
              <small>{{ item.note }}</small>
            </div>
          </div>

          <div class="sw2-admin-health-foot">
            <div v-for="item in secondaryMetrics.slice(0, 2)" :key="item.label">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
        </section>

        <section class="spg-surface sw2-admin-module sw2-admin-today">
          <header class="sw2-module-head">
            <div>
              <span class="sw2-admin-section-label">TODAY</span>
              <strong>今日运行</strong>
            </div>
            <span class="sw2-module-note">{{ todayLabel }}</span>
          </header>

          <div class="sw2-admin-metrics">
            <article v-for="metric in primaryMetrics" :key="metric.label" class="sw2-admin-metric">
              <div class="sw2-admin-metric-head">
                <span class="sw2-admin-metric-icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
                    <path stroke-linecap="round" stroke-linejoin="round" :d="metric.icon" />
                  </svg>
                </span>
                <span>{{ metric.label }}</span>
              </div>
              <strong>{{ metric.value }}</strong>
              <small>{{ metric.note }}</small>
              <span class="sw2-admin-metric-rule" aria-hidden="true"></span>
            </article>
          </div>

          <div class="sw2-admin-today-foot">
            <div v-for="item in secondaryMetrics.slice(2)" :key="item.label">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
              <small>{{ item.note }}</small>
            </div>
          </div>
        </section>
      </div>

      <section class="spg-surface sw2-admin-module sw2-admin-hour">
        <header class="sw2-module-head sw2-admin-hour-head">
          <div>
            <span class="sw2-admin-section-label">LAST 1 HOUR</span>
            <strong>最近一小时</strong>
          </div>
          <div class="sw2-admin-hour-status">
            <span :class="{ 'is-live': opsEnabled && !!opsSnapshot }"><i aria-hidden="true"></i>{{ opsEnabled ? 'OPS MONITORING' : 'MONITORING OFF' }}</span>
            <b v-if="opsEnabled && opsSnapshot">{{ formatCompact(opsOverview?.request_count_total) }} requests</b>
          </div>
        </header>

        <div v-if="!opsEnabled" class="sw2-admin-hour-shell">
          <div class="sw2-admin-monitor-canvas sw2-admin-monitor-canvas--empty">
            <div class="sw2-admin-monitor-axis sw2-admin-monitor-axis--y" aria-hidden="true">
              <span>4</span><span>3</span><span>2</span><span>1</span><span>0</span>
            </div>
            <div class="sw2-admin-monitor-grid" aria-hidden="true"></div>
            <div class="sw2-admin-monitor-axis sw2-admin-monitor-axis--x" aria-hidden="true">
              <span>60m ago</span><span>45m</span><span>30m</span><span>15m</span><span>now</span>
            </div>
            <div class="sw2-admin-monitor-empty-copy">
              <span class="sw2-admin-monitor-empty-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 18V9m5 9V5m5 13v-7m5 7V7" />
                </svg>
              </span>
              <strong>运营监控尚未启用</strong>
              <p>启用 Ops Monitoring 后，这里会呈现真实请求趋势、SLA、错误率与端到端延迟。</p>
            </div>
          </div>

          <aside class="sw2-admin-monitor-insight">
            <span class="sw2-admin-section-label">OPERATIONS</span>
            <strong>实时运行洞察</strong>
            <p>当前基础 Overview 数据保持可用。监控开启后，会补齐请求质量与异常定位视角。</p>
            <ul>
              <li v-for="item in monitorCapabilities" :key="item"><i aria-hidden="true">✓</i>{{ item }}</li>
            </ul>
          </aside>
        </div>

        <div v-else-if="opsLoading && !opsSnapshot" class="sw2-admin-hour-state">
          <span class="sw2-admin-loading-mark" aria-hidden="true"></span>
          <div><strong>正在读取最近一小时运行数据</strong><p>同步吞吐量、SLA 与延迟指标…</p></div>
        </div>

        <div v-else-if="opsError && !opsSnapshot" class="sw2-admin-hour-state sw2-admin-hour-state--error">
          <div><strong>最近一小时数据暂时不可用</strong><p>Platform Health 与今日运行数据不受影响。</p></div>
        </div>

        <div v-else class="sw2-admin-hour-grid">
          <div class="sw2-admin-trend">
            <div class="sw2-admin-trend-copy">
              <span>REQUEST TREND</span>
              <strong>{{ formatCompact(opsOverview?.request_count_total) }}</strong>
              <p>{{ opsTrendDescription }}</p>
            </div>

            <div class="sw2-admin-sparkline" aria-label="最近一小时请求趋势">
              <div class="sw2-admin-sparkline-grid" aria-hidden="true"></div>
              <svg viewBox="0 0 100 36" preserveAspectRatio="none" role="img">
                <line x1="0" y1="35" x2="100" y2="35" class="sw2-admin-sparkline-base" />
                <polyline v-if="sparklinePoints" :points="sparklinePoints" class="sw2-admin-sparkline-line" />
              </svg>
            </div>
          </div>

          <div class="sw2-admin-diagnostics">
            <div v-for="item in opsDiagnostics" :key="item.label" class="sw2-admin-diagnostic">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
              <small>{{ item.note }}</small>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { opsAPI, type OpsDashboardSnapshotV2Response } from '@/api/admin/ops'
import { useAdminSettingsStore } from '@/stores'
import type { DashboardStats } from '@/types'
import '@/styles/smirel-admin-overview-v2.css'

const adminSettingsStore = useAdminSettingsStore()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const error = ref(false)
const lastUpdated = ref('')
const opsSnapshot = ref<OpsDashboardSnapshotV2Response | null>(null)
const opsLoading = ref(false)
const opsError = ref(false)

const opsEnabled = computed(() => adminSettingsStore.opsMonitoringEnabled)
const opsOverview = computed(() => opsSnapshot.value?.overview ?? null)

const metricIcons = {
  requests: 'M4 12l15-7-4.5 14-3.1-5.1L4 12zm7.4 1.9L19 5',
  settlement: 'M5 7.5C5 5.6 8.1 4 12 4s7 1.6 7 3.5S15.9 11 12 11 5 9.4 5 7.5zm0 0V12c0 1.9 3.1 3.5 7 3.5s7-1.6 7-3.5V7.5m-14 4.5v4.5c0 1.9 3.1 3.5 7 3.5s7-1.6 7-3.5V12',
  tokens: 'M6 3.75h9l3 3v13.5H6V3.75zm9 0v3h3M9 11h6m-6 3h6m-6 3h4',
  latency: 'M13 2L5.5 13H11l-1 9 8.5-12H13V2z',
}

const monitorCapabilities = [
  '请求吞吐量与延迟趋势',
  'SLA 与可用性监控',
  '错误率与异常分析',
  '业务限流与峰值定位',
]

function formatLocalDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function toFinite(value: unknown): number {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

function formatCompact(value: number | null | undefined): string {
  const n = toFinite(value)
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return n.toLocaleString()
}

function formatCost(value: number | null | undefined): string {
  const n = toFinite(value)
  return n >= 1 ? `$${n.toFixed(2)}` : `$${n.toFixed(4)}`
}

function formatDuration(value: number | null | undefined): string {
  const n = toFinite(value)
  if (!n) return '—'
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${Math.round(n)}ms`
}

function formatPercentFraction(value: number | null | undefined): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${(value * 100).toFixed(value * 100 >= 10 ? 1 : 2)}%`
}

const todayLabel = computed(() => new Intl.DateTimeFormat('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'short',
}).format(new Date()))

const healthHasIssues = computed(() => toFinite(stats.value?.error_accounts) > 0)
const healthDescription = computed(() => healthHasIssues.value
  ? `当前有 ${formatCompact(stats.value?.error_accounts)} 个上游处于异常状态，优先检查账号池和调度健康度。`
  : '当前没有检测到异常上游，核心调度资源处于正常状态。')

const healthMetrics = computed(() => [
  {
    label: '健康上游',
    value: `${formatCompact(stats.value?.normal_accounts)} / ${formatCompact(stats.value?.total_accounts)}`,
    note: '当前可调度资源',
    warning: false,
  },
  {
    label: '异常上游',
    value: formatCompact(stats.value?.error_accounts),
    note: healthHasIssues.value ? '需要优先处理' : '未发现异常资源',
    warning: healthHasIssues.value,
  },
  {
    label: '活跃用户',
    value: formatCompact(stats.value?.active_users),
    note: '当前活跃账户',
    warning: false,
  },
])

const primaryMetrics = computed(() => [
  { label: '今日请求', value: formatCompact(stats.value?.today_requests), note: `累计 ${formatCompact(stats.value?.total_requests)} 次`, icon: metricIcons.requests },
  { label: '今日结算', value: formatCost(stats.value?.today_actual_cost), note: '用户侧实际结算', icon: metricIcons.settlement },
  { label: '今日 Token', value: formatCompact(stats.value?.today_tokens), note: '输入与输出合计', icon: metricIcons.tokens },
  { label: '平均响应', value: formatDuration(stats.value?.average_duration_ms), note: '当前平均耗时', icon: metricIcons.latency },
])

const secondaryMetrics = computed(() => [
  { label: 'API Keys', value: `${formatCompact(stats.value?.active_api_keys)} / ${formatCompact(stats.value?.total_api_keys)} 启用`, note: '可用密钥' },
  { label: '今日新增用户', value: `+${formatCompact(stats.value?.today_new_users)}`, note: '自然日新增' },
  { label: '当前 RPM', value: formatCompact(stats.value?.rpm), note: '当前请求速率' },
  { label: '今日上游成本', value: formatCost(stats.value?.today_account_cost), note: '上游实际成本' },
])

const sparklinePoints = computed(() => {
  const points = opsSnapshot.value?.throughput_trend?.points ?? []
  if (!points.length) return ''

  const values = points.map((point) => toFinite(point.request_count))
  const max = Math.max(...values)
  const min = Math.min(...values)
  const spread = max - min

  return values.map((value, index) => {
    const x = values.length === 1 ? 50 : (index / (values.length - 1)) * 100
    const y = spread === 0 ? 18 : 32 - ((value - min) / spread) * 27
    return `${x.toFixed(2)},${y.toFixed(2)}`
  }).join(' ')
})

const opsDiagnostics = computed(() => [
  { label: 'SLA', value: formatPercentFraction(opsOverview.value?.sla), note: '近 1 小时成功服务水平' },
  { label: '错误率', value: formatPercentFraction(opsOverview.value?.error_rate), note: `${formatCompact(opsOverview.value?.error_count_sla)} SLA errors` },
  { label: 'P95 延迟', value: formatDuration(opsOverview.value?.duration?.p95_ms), note: '端到端请求时延' },
  { label: '当前 QPS', value: toFinite(opsOverview.value?.qps?.current).toFixed(1), note: `峰值 ${toFinite(opsOverview.value?.qps?.peak).toFixed(1)}` },
])

const opsTrendDescription = computed(() => {
  const overview = opsOverview.value
  if (!overview) return '暂无足够的请求趋势数据。'
  const errors = formatCompact(overview.error_count_sla)
  const limited = formatCompact(overview.business_limited_count)
  return `近一小时 ${formatCompact(overview.request_count_total)} 次请求 · ${errors} 个 SLA 错误 · ${limited} 次业务限流。`
})

async function loadOpsSnapshot() {
  if (!opsEnabled.value) {
    opsSnapshot.value = null
    opsError.value = false
    return
  }

  opsLoading.value = true
  opsError.value = false
  try {
    opsSnapshot.value = await opsAPI.getDashboardSnapshotV2({
      time_range: '1h',
      mode: 'auto',
    })
  } catch (err) {
    opsError.value = true
    console.error('Failed to load admin V2 ops snapshot:', err)
  } finally {
    opsLoading.value = false
  }
}

async function loadSnapshot() {
  loading.value = true
  error.value = false

  const opsPromise = loadOpsSnapshot()
  try {
    const end = new Date()
    const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: formatLocalDate(start),
      end_date: formatLocalDate(end),
      granularity: 'hour',
      include_stats: true,
      include_trend: false,
      include_model_stats: false,
      include_group_stats: false,
      include_users_trend: false,
    })
    stats.value = response.stats || null
    lastUpdated.value = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch (err) {
    error.value = true
    console.error('Failed to load admin V2 overview:', err)
  } finally {
    loading.value = false
    await opsPromise
  }
}

onMounted(() => {
  void loadSnapshot()
})
</script>
