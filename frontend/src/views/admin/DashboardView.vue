<template>
  <AppLayout>
    <section class="smw-dashboard">
      <header class="smw-workspace-head">
        <div class="smw-workspace-head__copy">
          <span class="smw-kicker">SMIREL OPERATIONS</span>
          <div class="smw-workspace-title-row">
            <h1>{{ copy.workspaceTitle }}</h1>
            <span class="smw-live-state"><i></i>{{ copy.live }}</span>
          </div>
          <p>{{ copy.workspaceDescription }}</p>
        </div>

        <div class="smw-workspace-head__actions">
          <router-link to="/admin/accounts" class="smw-action smw-action--primary">{{ copy.upstreams }}</router-link>
          <router-link to="/admin/groups" class="smw-action">{{ copy.routing }}</router-link>
          <router-link to="/admin/users" class="smw-action">{{ copy.users }}</router-link>
          <router-link to="/admin/ops" class="smw-action">{{ copy.openOps }}</router-link>
        </div>

        <div class="smw-live-strip" :aria-label="copy.health">
          <div class="smw-live-item">
            <span>{{ copy.activeUpstreams }}</span>
            <strong>{{ stats?.normal_accounts || 0 }}</strong>
          </div>
          <div class="smw-live-item">
            <span>{{ copy.errorUpstreams }}</span>
            <strong :class="{ 'smw-live-value--danger': (stats?.error_accounts || 0) > 0 }">{{ stats?.error_accounts || 0 }}</strong>
          </div>
          <div class="smw-live-item">
            <span>{{ copy.activeUsers }}</span>
            <strong>{{ stats?.active_users || 0 }}</strong>
          </div>
          <div class="smw-live-item">
            <span>RPM</span>
            <strong>{{ formatCompact(stats?.rpm || 0) }}</strong>
          </div>
        </div>
      </header>

      <div v-if="loading && !stats" class="smw-loading">{{ copy.loading }}</div>
      <section v-else class="smw-summary-strip" :aria-label="copy.summary">
        <article v-for="card in statCards" :key="card.label" class="smw-summary-item">
          <span>{{ card.label }}</span>
          <strong>{{ card.value }}</strong>
          <small>{{ card.note }}</small>
        </article>
      </section>

      <div class="smw-toolbar">
        <div class="smw-toolbar__filters">
          <DateRangePicker
            v-model:start-date="startDate"
            v-model:end-date="endDate"
            @change="onDateRangeChange"
          />
          <div class="smw-granularity">
            <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
          </div>
        </div>
        <button class="smw-action" type="button" :disabled="chartsLoading" @click="loadDashboardStats">
          {{ copy.refresh }}
        </button>
      </div>

      <section class="smw-primary-grid">
        <article class="smw-panel smw-panel--wide">
          <header class="smw-panel__head">
            <div>
              <span>{{ copy.analysis }}</span>
              <h2>{{ copy.modelMix }}</h2>
            </div>
            <small>{{ startDate }} → {{ endDate }}</small>
          </header>
          <div class="smw-panel__body">
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
        </article>

        <article class="smw-panel">
          <header class="smw-panel__head">
            <div>
              <span>{{ copy.traffic }}</span>
              <h2>{{ copy.tokenTrend }}</h2>
            </div>
            <small>{{ granularity }}</small>
          </header>
          <div class="smw-panel__body">
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>
        </article>
      </section>

      <section class="smw-secondary-grid">
        <article class="smw-panel smw-panel--trend">
          <header class="smw-panel__head">
            <div>
              <span>{{ copy.users }}</span>
              <h2>{{ copy.userTrend }}</h2>
            </div>
            <small>TOP 12</small>
          </header>
          <div class="smw-panel__body smw-panel__body--chart">
            <div v-if="userTrendLoading" class="smw-loading smw-loading--embedded">{{ copy.loading }}</div>
            <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
            <div v-else class="smw-loading smw-loading--embedded">{{ copy.noData }}</div>
          </div>
        </article>

        <article class="smw-panel smw-panel--actions">
          <header class="smw-panel__head">
            <div>
              <span>OPERATE</span>
              <h2>{{ copy.actions }}</h2>
            </div>
          </header>
          <div class="smw-action-list">
            <router-link v-for="action in quickActions" :key="action.path" :to="action.path" class="smw-action-row">
              <div>
                <strong>{{ action.title }}</strong>
                <span>{{ action.description }}</span>
              </div>
              <b>→</b>
            </router-link>
          </div>
        </article>
      </section>
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
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
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
        workspaceTitle: '运营工作台',
        workspaceDescription: '把平台运行、资源、用量与商业数据收进同一个可操作工作区。',
        live: '实时', health: '实时运行状态', summary: '平台关键指标', activeUpstreams: '正常上游', errorUpstreams: '异常上游', activeUsers: '活跃用户',
        upstreams: '上游账户', routing: '路由策略', users: '用户管理', openOps: '运行状态',
        loading: '正在读取平台数据…', refresh: '刷新数据', analysis: '消费分析', traffic: '调用流量', modelMix: '模型与用户消费分布', tokenTrend: 'Token 趋势', userTrend: '用户调用趋势', noData: '暂无数据', actions: '运营入口',
        apiKeys: 'API Keys', accounts: '上游账户', requests: '今日请求', newUsers: '今日新增', todayTokens: '今日 Tokens', cost: '今日成本', rpm: '当前 RPM', latency: '平均响应',
        activeKeys: '启用密钥', healthyAccounts: '正常 / 异常', totalRequests: '累计请求', totalUsers: '累计用户', billed: '实际结算', accountCost: '上游成本', perMinute: '每分钟请求', average: '平均耗时',
      }
    : {
        workspaceTitle: 'Operations Workspace',
        workspaceDescription: 'Platform runtime, resources, usage and commercial data in one operational workspace.',
        live: 'LIVE', health: 'Live platform health', summary: 'Platform key metrics', activeUpstreams: 'Healthy upstreams', errorUpstreams: 'Errored upstreams', activeUsers: 'Active users',
        upstreams: 'Upstream Accounts', routing: 'Routing Policies', users: 'Users', openOps: 'Operations',
        loading: 'Loading platform data…', refresh: 'Refresh', analysis: 'Spend analysis', traffic: 'Traffic', modelMix: 'Model & user spend mix', tokenTrend: 'Token trend', userTrend: 'User request trend', noData: 'No data', actions: 'Operations',
        apiKeys: 'API Keys', accounts: 'Upstream accounts', requests: 'Requests today', newUsers: 'New users today', todayTokens: 'Tokens today', cost: 'Cost today', rpm: 'Current RPM', latency: 'Avg response',
        activeKeys: 'active keys', healthyAccounts: 'healthy / error', totalRequests: 'total requests', totalUsers: 'total users', billed: 'actual billed', accountCost: 'upstream cost', perMinute: 'requests per minute', average: 'average duration',
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

const statCards = computed(() => [
  { label: copy.value.apiKeys, value: formatCompact(stats.value?.total_api_keys), note: `${formatCompact(stats.value?.active_api_keys)} ${copy.value.activeKeys}` },
  { label: copy.value.accounts, value: formatCompact(stats.value?.total_accounts), note: `${stats.value?.normal_accounts || 0} / ${stats.value?.error_accounts || 0} ${copy.value.healthyAccounts}` },
  { label: copy.value.requests, value: formatCompact(stats.value?.today_requests), note: `${formatCompact(stats.value?.total_requests)} ${copy.value.totalRequests}` },
  { label: copy.value.newUsers, value: `+${formatCompact(stats.value?.today_new_users)}`, note: `${formatCompact(stats.value?.total_users)} ${copy.value.totalUsers}` },
  { label: copy.value.todayTokens, value: formatCompact(stats.value?.today_tokens), note: `${copy.value.billed}: $${formatCost(stats.value?.today_actual_cost)}` },
  { label: copy.value.cost, value: `$${formatCost(stats.value?.today_account_cost)}`, note: copy.value.accountCost },
  { label: copy.value.rpm, value: formatCompact(stats.value?.rpm), note: copy.value.perMinute },
  { label: copy.value.latency, value: formatDuration(stats.value?.average_duration_ms), note: copy.value.average },
])

const quickActions = computed(() => {
  const items = [
    { path: '/admin/accounts', title: copy.value.upstreams, description: isZh.value ? '检查账号池、错误与可调度状态' : 'Inspect pools, errors and schedulability' },
    { path: '/admin/groups', title: copy.value.routing, description: isZh.value ? '配置路由与计费策略' : 'Configure routing and billing policies' },
    { path: '/admin/users', title: copy.value.users, description: isZh.value ? '管理客户、额度与访问权限' : 'Manage customers, quota and access' },
    { path: '/admin/orders/dashboard', title: isZh.value ? '收入概览' : 'Revenue overview', description: isZh.value ? '查看订单和收入数据' : 'Inspect orders and revenue' },
  ]
  if (canUseBatchImage.value) items.push({ path: '/batch-image', title: isZh.value ? '批量生图' : 'Batch image', description: isZh.value ? '打开批量生图工作流' : 'Open the batch image workflow' })
  return items
})

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const chartColors = computed(() => ({ text: isDarkMode.value ? '#dce9e6' : '#465653', grid: isDarkMode.value ? '#253431' : '#e1e9e6' }))
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: { position: 'top' as const, labels: { color: chartColors.value.text, usePointStyle: true, pointStyle: 'circle', padding: 15, font: { size: 10 } } },
    tooltip: { callbacks: { label: (context: any) => `${context.dataset.label}: ${formatCompact(context.raw)}` } },
  },
  scales: {
    x: { grid: { color: chartColors.value.grid }, ticks: { color: chartColors.value.text, font: { size: 9 } } },
    y: { grid: { color: chartColors.value.grid }, ticks: { color: chartColors.value.text, font: { size: 9 }, callback: (value: string | number) => formatCompact(Number(value)) } },
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
    })),
  }
})

function goToUserUsage(item: UserSpendingRankingItem) {
  void router.push({ path: '/admin/usage', query: { user_id: String(item.user_id), start_date: startDate.value, end_date: endDate.value } })
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
    const response = await adminAPI.dashboard.getUserUsageTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value, limit: 12 })
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
    const response = await adminAPI.dashboard.getUserSpendingRanking({ start_date: startDate.value, end_date: endDate.value, limit: rankingLimit })
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