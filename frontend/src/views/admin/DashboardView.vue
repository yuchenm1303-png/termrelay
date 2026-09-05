<template>
  <AppLayout>
    <section class="smg-admin-dashboard">
      <GlassSurface class="smg-admin-overview">
        <div class="smg-admin-overview-copy">
          <div class="smg-admin-kicker">SMIREL OPERATIONS</div>
          <h1>{{ copy.pageTitle }}</h1>
          <p>{{ copy.pageDescription }}</p>

          <div class="smg-admin-overview-actions">
            <router-link to="/admin/accounts" class="smg-button smg-button--primary">{{ copy.upstreams }}</router-link>
            <router-link to="/admin/groups" class="smg-button">{{ copy.routing }}</router-link>
            <router-link to="/admin/users" class="smg-button">{{ copy.users }}</router-link>
          </div>
        </div>

        <aside class="smg-admin-health">
          <div>
            <div class="smg-admin-health-head">
              <span>{{ copy.health }}</span>
              <span
                class="smg-admin-health-badge"
                :class="{ 'smg-admin-health-badge--warning': healthHasIssues }"
              >
                <i></i>{{ healthLabel }}
              </span>
            </div>

            <dl class="smg-admin-health-list">
              <div class="smg-admin-health-row">
                <dt>{{ copy.healthyUpstreams }}</dt>
                <dd>{{ formatCompact(stats?.normal_accounts) }} / {{ formatCompact(stats?.total_accounts) }}</dd>
              </div>
              <div class="smg-admin-health-row">
                <dt>{{ copy.errorUpstreams }}</dt>
                <dd :class="{ 'is-warning': healthHasIssues }">{{ formatCompact(stats?.error_accounts) }}</dd>
              </div>
              <div class="smg-admin-health-row">
                <dt>{{ copy.activeUsers }}</dt>
                <dd>{{ formatCompact(stats?.active_users) }}</dd>
              </div>
            </dl>
          </div>

          <router-link to="/admin/ops" class="smg-button smg-button--wide">{{ copy.openOps }}</router-link>
        </aside>
      </GlassSurface>

      <GlassSurface v-if="loading && !stats" tone="quiet" class="smg-admin-loading">
        {{ copy.loading }}
      </GlassSurface>

      <template v-else>
        <div class="smg-admin-primary-grid">
          <GlassSurface v-for="card in primaryStatCards" :key="card.label" class="smg-admin-primary-stat">
            <span class="smg-admin-stat-label">{{ card.label }}</span>
            <strong class="smg-admin-stat-value">{{ card.value }}</strong>
            <small class="smg-admin-stat-note">{{ card.note }}</small>
          </GlassSurface>
        </div>

        <GlassSurface tone="quiet" class="smg-admin-secondary-strip">
          <div v-for="card in secondaryStatCards" :key="card.label" class="smg-admin-secondary-stat">
            <span>{{ card.label }}</span>
            <strong>{{ card.value }}</strong>
            <small>{{ card.note }}</small>
          </div>
        </GlassSurface>
      </template>

      <GlassSurface tone="quiet" class="smg-admin-analytics-head">
        <div class="smg-admin-analytics-copy">
          <div class="smg-admin-section-kicker">ANALYTICS</div>
          <h2>{{ copy.analytics }}</h2>
          <p>{{ copy.analyticsDescription }}</p>
        </div>

        <div class="smg-admin-controls">
          <DateRangePicker
            v-model:start-date="startDate"
            v-model:end-date="endDate"
            @change="onDateRangeChange"
          />
          <div class="smg-admin-granularity">
            <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
          </div>
          <button class="smg-button" type="button" :disabled="chartsLoading" @click="loadDashboardStats">
            {{ copy.refresh }}
          </button>
        </div>
      </GlassSurface>

      <div class="smg-admin-analytics-grid">
        <GlassSurface class="smg-admin-panel">
          <div class="smg-admin-panel-head">
            <div>
              <div class="smg-admin-section-kicker">MODEL / SPEND</div>
              <h2>{{ copy.modelMix }}</h2>
            </div>
            <span>{{ startDate }} → {{ endDate }}</span>
          </div>
          <div class="smg-admin-chart-panel">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
          </div>
        </GlassSurface>

        <GlassSurface class="smg-admin-panel smg-admin-token-panel">
          <div class="smg-admin-panel-head">
            <div>
              <div class="smg-admin-section-kicker">TOKENS</div>
              <h2>{{ copy.tokenTrend }}</h2>
            </div>
            <span>{{ granularity }}</span>
          </div>
          <div class="smg-admin-chart-panel">
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>
        </GlassSurface>
      </div>

      <div class="smg-admin-bottom-grid">
        <GlassSurface class="smg-admin-panel">
          <div class="smg-admin-panel-head">
            <div>
              <div class="smg-admin-section-kicker">USERS</div>
              <h2>{{ copy.userTrend }}</h2>
            </div>
            <span>TOP 12</span>
          </div>
          <div class="smg-admin-user-chart">
            <div v-if="userTrendLoading" class="smg-admin-empty">{{ copy.loading }}</div>
            <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
            <div v-else class="smg-admin-empty">{{ copy.noData }}</div>
          </div>
        </GlassSurface>

        <GlassSurface class="smg-admin-panel">
          <div class="smg-admin-panel-head">
            <div>
              <div class="smg-admin-section-kicker">OPERATE</div>
              <h2>{{ copy.actions }}</h2>
            </div>
          </div>
          <div class="smg-admin-actions-list">
            <router-link
              v-for="(action, index) in quickActions"
              :key="action.path"
              :to="action.path"
              class="smg-admin-action"
            >
              <span class="smg-admin-action-index">{{ String(index + 1).padStart(2, '0') }}</span>
              <span class="smg-admin-action-copy">
                <strong>{{ action.title }}</strong>
                <span>{{ action.description }}</span>
              </span>
              <span class="smg-admin-action-arrow">→</span>
            </router-link>
          </div>
        </GlassSurface>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem,
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import GlassSurface from '@/components/glass/GlassSurface.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import '@/styles/smirel-glass-admin-dashboard-v5.css'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t, locale } = useI18n()
const appStore = useAppStore()
const router = useRouter()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12

const formatLocalDate = (date: Date): string =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

const end = new Date()
const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
const granularity = ref<'day' | 'hour'>('hour')
const startDate = ref(formatLocalDate(start))
const endDate = ref(formatLocalDate(end))
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') },
])

const copy = computed(() =>
  isZh.value
    ? {
        pageTitle: '运营工作台',
        pageDescription: '先看平台是否健康，再处理流量、成本与用户。核心状态、消费和趋势集中在同一个工作区。',
        upstreams: '上游账户',
        routing: '路由策略',
        users: '用户管理',
        health: 'LIVE HEALTH',
        healthStable: '运行正常',
        healthWarning: '需要关注',
        healthyUpstreams: '健康上游',
        errorUpstreams: '异常上游',
        activeUsers: '活跃用户',
        openOps: '打开运行状态',
        loading: '正在读取平台数据…',
        refresh: '刷新数据',
        analytics: '流量与消费',
        analyticsDescription: '按时间范围查看模型、Token 与用户活动，把运营判断集中在同一组数据里。',
        modelMix: '模型与消费',
        tokenTrend: 'Token 使用趋势',
        userTrend: '用户调用趋势',
        noData: '暂无数据',
        actions: '运营入口',
        apiKeys: 'API Keys',
        requests: '今日请求',
        newUsers: '今日新增',
        todayTokens: '今日 Tokens',
        settlement: '今日结算',
        upstreamCost: '上游成本',
        rpm: '当前 RPM',
        latency: '平均响应',
        activeKeys: '启用密钥',
        totalRequests: '累计请求',
        totalUsers: '累计用户',
        settlementNote: '用户侧实际结算',
        tokenVolume: '今日 Token 用量',
        accountCost: '今日上游成本',
        perMinute: '每分钟请求',
        average: '平均耗时',
      }
    : {
        pageTitle: 'Operations workspace',
        pageDescription: 'Check platform health first, then manage traffic, cost and users. Core status, spend and trends stay in one workspace.',
        upstreams: 'Upstream Accounts',
        routing: 'Routing Policies',
        users: 'Users',
        health: 'LIVE HEALTH',
        healthStable: 'Operational',
        healthWarning: 'Needs attention',
        healthyUpstreams: 'Healthy upstreams',
        errorUpstreams: 'Upstream errors',
        activeUsers: 'Active users',
        openOps: 'Open operations',
        loading: 'Loading platform data…',
        refresh: 'Refresh',
        analytics: 'Traffic & spend',
        analyticsDescription: 'Review model, token and user activity across the selected time range in one operational view.',
        modelMix: 'Model & spend mix',
        tokenTrend: 'Token usage trend',
        userTrend: 'User request trend',
        noData: 'No data',
        actions: 'Operations',
        apiKeys: 'API Keys',
        requests: 'Requests today',
        newUsers: 'New users today',
        todayTokens: 'Tokens today',
        settlement: 'Settlement today',
        upstreamCost: 'Upstream cost',
        rpm: 'Current RPM',
        latency: 'Avg response',
        activeKeys: 'active keys',
        totalRequests: 'total requests',
        totalUsers: 'total users',
        settlementNote: 'actual customer settlement',
        tokenVolume: 'token volume today',
        accountCost: 'upstream cost today',
        perMinute: 'requests per minute',
        average: 'average duration',
      },
)

function toFinite(value: unknown): number {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
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
  if (n >= 1) return n.toFixed(2)
  return n.toFixed(4)
}

function formatDuration(ms: number | null | undefined): string {
  const n = toFinite(ms)
  if (!n) return '—'
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${Math.round(n)}ms`
}

const healthHasIssues = computed(() => toFinite(stats.value?.error_accounts) > 0)
const healthLabel = computed(() => healthHasIssues.value ? copy.value.healthWarning : copy.value.healthStable)

const primaryStatCards = computed(() => [
  {
    label: copy.value.requests,
    value: formatCompact(stats.value?.today_requests),
    note: `${formatCompact(stats.value?.total_requests)} ${copy.value.totalRequests}`,
  },
  {
    label: copy.value.settlement,
    value: `$${formatCost(stats.value?.today_actual_cost)}`,
    note: copy.value.settlementNote,
  },
  {
    label: copy.value.todayTokens,
    value: formatCompact(stats.value?.today_tokens),
    note: copy.value.tokenVolume,
  },
  {
    label: copy.value.latency,
    value: formatDuration(stats.value?.average_duration_ms),
    note: copy.value.average,
  },
])

const secondaryStatCards = computed(() => [
  {
    label: copy.value.apiKeys,
    value: formatCompact(stats.value?.total_api_keys),
    note: `${formatCompact(stats.value?.active_api_keys)} ${copy.value.activeKeys}`,
  },
  {
    label: copy.value.newUsers,
    value: `+${formatCompact(stats.value?.today_new_users)}`,
    note: `${formatCompact(stats.value?.total_users)} ${copy.value.totalUsers}`,
  },
  {
    label: copy.value.rpm,
    value: formatCompact(stats.value?.rpm),
    note: copy.value.perMinute,
  },
  {
    label: copy.value.upstreamCost,
    value: `$${formatCost(stats.value?.today_account_cost)}`,
    note: copy.value.accountCost,
  },
])

const quickActions = computed(() => {
  const items = [
    { path: '/admin/accounts', title: copy.value.upstreams, description: isZh.value ? '检查账号池、错误与可调度状态' : 'Inspect pools, errors and schedulability' },
    { path: '/admin/groups', title: copy.value.routing, description: isZh.value ? '配置路由与计费策略' : 'Configure routing and billing policies' },
    { path: '/admin/users', title: copy.value.users, description: isZh.value ? '管理客户、额度与访问权限' : 'Manage customers, quota and access' },
    { path: '/admin/orders/dashboard', title: isZh.value ? '收入概览' : 'Revenue overview', description: isZh.value ? '查看订单和收入数据' : 'Inspect orders and revenue' },
  ]
  if (canUseBatchImage.value) {
    items.push({
      path: '/batch-image',
      title: isZh.value ? '批量生图' : 'Batch image',
      description: isZh.value ? '打开批量生图工作流' : 'Open the batch image workflow',
    })
  }
  return items
})

const chartColors = computed(() => ({
  text: 'rgba(235, 244, 251, .68)',
  grid: 'rgba(255, 255, 255, .07)',
}))

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 16,
        font: { size: 10.5 },
      },
    },
    tooltip: {
      callbacks: { label: (context: any) => `${context.dataset.label}: ${formatCompact(context.raw)}` },
    },
  },
  scales: {
    x: {
      grid: { color: chartColors.value.grid },
      ticks: { color: chartColors.value.text, font: { size: 10 } },
    },
    y: {
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        callback: (value: string | number) => formatCompact(Number(value)),
      },
    },
  },
}))

const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null
  const groups = new Map<number, { name: string; data: Map<string, number> }>()
  const dates = new Set<string>()
  userTrend.value.forEach((point) => {
    dates.add(point.date)
    const name = point.username?.trim() || point.email?.trim() || `User ${point.user_id}`
    if (!groups.has(point.user_id)) groups.set(point.user_id, { name, data: new Map() })
    groups.get(point.user_id)!.data.set(point.date, point.tokens)
  })
  const sortedDates = Array.from(dates).sort()
  const colors = ['#38bfae','#5b8def','#d99c3c','#d76473','#8e77d6','#4cb4c7','#88b35d','#cf7fb2','#9d8b62','#6aa78a','#a0a7b8','#c87959']
  return {
    labels: sortedDates,
    datasets: Array.from(groups.values()).map((group, index) => ({
      label: group.name,
      data: sortedDates.map((date) => group.data.get(date) || 0),
      borderColor: colors[index % colors.length],
      backgroundColor: `${colors[index % colors.length]}20`,
      fill: false,
      tension: 0.3,
      borderWidth: 1.5,
      pointRadius: 1.75,
      pointHoverRadius: 4,
    })),
  }
})

function goToUserUsage(item: UserSpendingRankingItem) {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value,
    },
  })
}

function onDateRangeChange(range: { startDate: string; endDate: string; preset: string | null }) {
  const rangeStart = new Date(range.startDate)
  const rangeEnd = new Date(range.endDate)
  const daysDiff = Math.ceil((rangeEnd.getTime() - rangeStart.getTime()) / 86400000)
  granularity.value = daysDiff <= 1 ? 'hour' : 'day'
  void loadChartData()
}

async function loadDashboardSnapshot(includeStats: boolean) {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) loading.value = true
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false,
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) stats.value = response.stats
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

async function loadUsersTrend() {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12,
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) userTrendLoading.value = false
  }
}

async function loadUserSpendingRanking() {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit,
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) rankingLoading.value = false
  }
}

async function loadDashboardStats() {
  await Promise.all([loadDashboardSnapshot(true), loadUsersTrend(), loadUserSpendingRanking()])
}

async function loadChartData() {
  await Promise.all([loadDashboardSnapshot(false), loadUsersTrend(), loadUserSpendingRanking()])
}

onMounted(() => {
  void refreshBatchImageAccess()
  void loadDashboardStats()
})
</script>
