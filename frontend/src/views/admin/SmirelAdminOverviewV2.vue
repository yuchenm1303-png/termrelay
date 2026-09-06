<template>
  <div class="sw2-admin-overview">
    <div v-if="loading && !stats" class="sw2-admin-loading">正在读取平台运行数据…</div>

    <div v-else-if="error && !stats" class="sw2-admin-error">
      <span>平台数据暂时没有加载成功。</span>
      <button type="button" @click="loadSnapshot">重新加载</button>
    </div>

    <template v-else>
      <header class="sw2-admin-overview-head">
        <div class="sw2-admin-overview-head-copy">
          <span>PLATFORM PULSE</span>
          <strong>平台运行概览</strong>
        </div>
        <div class="sw2-admin-overview-head-actions">
          <span v-if="lastUpdated" class="sw2-admin-updated">更新于 {{ lastUpdated }}</span>
          <button class="sw2-admin-refresh" type="button" :disabled="loading || opsLoading" @click="loadSnapshot">
            {{ loading || opsLoading ? '刷新中…' : '刷新' }}
          </button>
        </div>
      </header>

      <div class="sw2-admin-grid">
        <section class="sw2-admin-health">
          <span class="sw2-admin-section-label">PLATFORM HEALTH</span>
          <div class="sw2-admin-health-top">
            <div class="sw2-admin-health-copy">
              <strong>{{ healthHasIssues ? '需要关注' : '运行正常' }}</strong>
              <p>{{ healthDescription }}</p>
            </div>
            <span class="sw2-admin-health-state" :class="{ 'sw2-admin-health-state--warning': healthHasIssues }">
              <i aria-hidden="true"></i>{{ healthHasIssues ? 'ATTENTION' : 'OPERATIONAL' }}
            </span>
          </div>

          <div class="sw2-admin-health-list">
            <div class="sw2-admin-health-row">
              <span>健康上游</span>
              <strong>{{ formatCompact(stats?.normal_accounts) }} / {{ formatCompact(stats?.total_accounts) }}</strong>
            </div>
            <div class="sw2-admin-health-row">
              <span>异常上游</span>
              <strong :class="{ 'is-warning': healthHasIssues }">{{ formatCompact(stats?.error_accounts) }}</strong>
            </div>
            <div class="sw2-admin-health-row">
              <span>活跃用户</span>
              <strong>{{ formatCompact(stats?.active_users) }}</strong>
            </div>
          </div>
        </section>

        <section class="sw2-admin-today">
          <span class="sw2-admin-section-label">TODAY</span>
          <p class="sw2-admin-today-intro">今天最需要关注的流量、结算、Token 和响应时间集中在这里，不再拆成一排独立 KPI 卡。</p>

          <div class="sw2-admin-metrics">
            <div v-for="metric in primaryMetrics" :key="metric.label" class="sw2-admin-metric">
              <span>{{ metric.label }}</span>
              <strong>{{ metric.value }}</strong>
              <small>{{ metric.note }}</small>
            </div>
          </div>
        </section>
      </div>

      <div class="sw2-admin-summary">
        <div v-for="item in secondaryMetrics" :key="item.label" class="sw2-admin-summary-row">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>

      <section class="sw2-admin-hour">
        <header class="sw2-admin-hour-head">
          <div>
            <span class="sw2-admin-section-label">LAST 1 HOUR</span>
            <strong>最近一小时</strong>
          </div>
          <span v-if="opsEnabled && opsSnapshot" class="sw2-admin-hour-total">{{ formatCompact(opsOverview?.request_count_total) }} requests</span>
        </header>

        <div v-if="!opsEnabled" class="sw2-admin-hour-empty">
          <strong>运营监控未启用</strong>
          <p>当前 Overview 保持可用；开启 Ops Monitoring 后，这里会显示近一小时的请求趋势和运行诊断。</p>
        </div>

        <div v-else-if="opsLoading && !opsSnapshot" class="sw2-admin-hour-empty">
          <strong>正在读取近一小时运行数据…</strong>
        </div>

        <div v-else-if="opsError && !opsSnapshot" class="sw2-admin-hour-empty">
          <strong>近一小时运行数据暂时不可用</strong>
          <p>Platform Pulse 不受影响，可以稍后再次刷新。</p>
        </div>

        <div v-else class="sw2-admin-hour-grid">
          <div class="sw2-admin-trend">
            <div class="sw2-admin-trend-copy">
              <span>REQUEST TREND</span>
              <strong>{{ formatCompact(opsOverview?.request_count_total) }}</strong>
              <p>{{ opsTrendDescription }}</p>
            </div>

            <div class="sw2-admin-sparkline" aria-label="最近一小时请求趋势">
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

const healthHasIssues = computed(() => toFinite(stats.value?.error_accounts) > 0)
const healthDescription = computed(() => healthHasIssues.value
  ? `当前有 ${formatCompact(stats.value?.error_accounts)} 个上游处于异常状态，优先检查账号池和调度健康度。`
  : '当前没有检测到异常上游。平台核心调度资源处于正常状态。')

const primaryMetrics = computed(() => [
  { label: '今日请求', value: formatCompact(stats.value?.today_requests), note: `累计 ${formatCompact(stats.value?.total_requests)} 次` },
  { label: '今日结算', value: formatCost(stats.value?.today_actual_cost), note: '用户侧实际结算' },
  { label: '今日 Token', value: formatCompact(stats.value?.today_tokens), note: '输入与输出合计' },
  { label: '平均响应', value: formatDuration(stats.value?.average_duration_ms), note: '当前平均耗时' },
])

const secondaryMetrics = computed(() => [
  { label: 'API Keys', value: `${formatCompact(stats.value?.active_api_keys)} / ${formatCompact(stats.value?.total_api_keys)} 启用` },
  { label: '今日新增用户', value: `+${formatCompact(stats.value?.today_new_users)}` },
  { label: '当前 RPM', value: formatCompact(stats.value?.rpm) },
  { label: '今日上游成本', value: formatCost(stats.value?.today_account_cost) },
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
