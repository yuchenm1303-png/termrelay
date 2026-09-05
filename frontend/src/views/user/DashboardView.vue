<template>
  <AppLayout>
    <section class="smg-dashboard">
      <div class="smg-dashboard-hero">
        <GlassSurface class="smg-dashboard-intro">
          <div class="smg-dashboard-kicker">SMIREL API WORKSPACE</div>
          <h1>{{ copy.heroTitle }}</h1>
          <p>{{ copy.heroDescription }}</p>
          <div class="smg-dashboard-actions">
            <router-link to="/keys" class="smg-button smg-button--primary">{{ copy.manageKeys }}</router-link>
            <router-link to="/model-plaza?embedded=1" class="smg-button">{{ copy.models }}</router-link>
            <router-link to="/usage" class="smg-button">{{ copy.usage }}</router-link>
          </div>
        </GlassSurface>

        <GlassSurface class="smg-dashboard-quickstart">
          <div class="smg-dashboard-kicker">{{ copy.quickstart }}</div>
          <div class="smg-dashboard-code">
            <span>BASE_URL</span>
            <code>{{ apiBase }}</code>
            <span>AUTH</span>
            <code>Bearer sk-••••••••</code>
            <span>REQUEST</span>
            <code>curl {{ apiBase }}/chat/completions</code>
          </div>
          <button type="button" class="smg-button smg-button--wide" @click="copyApiBase">
            {{ copied ? copy.copied : copy.copyBase }}
          </button>
        </GlassSurface>
      </div>

      <div class="smg-dashboard-metrics">
        <GlassMetric
          v-for="card in statCards"
          :key="card.label"
          :label="card.label"
          :value="card.value"
          :note="card.note"
        />
      </div>

      <GlassSurface class="smg-dashboard-panel">
        <header class="smg-dashboard-panel-head">
          <div>
            <span>{{ copy.trend }}</span>
            <strong>{{ startDate }} → {{ endDate }}</strong>
          </div>
        </header>
        <div class="smg-dashboard-chart-surface">
          <UserDashboardCharts
            v-model:startDate="startDate"
            v-model:endDate="endDate"
            v-model:granularity="granularity"
            :loading="loadingCharts"
            :trend="trendData"
            :models="modelStats"
            @dateRangeChange="loadCharts"
            @granularityChange="loadCharts"
            @refresh="refreshAll"
          />
        </div>
      </GlassSurface>

      <div class="smg-dashboard-lower">
        <GlassSurface class="smg-dashboard-panel">
          <header class="smg-dashboard-panel-head">
            <div><span>{{ copy.recent }}</span><strong>{{ copy.usage }}</strong></div>
            <router-link to="/usage">{{ copy.viewAll }}</router-link>
          </header>
          <div v-if="loadingUsage" class="smg-dashboard-empty">{{ copy.loading }}</div>
          <div v-else-if="recentUsage.length === 0" class="smg-dashboard-empty">{{ copy.empty }}</div>
          <div v-else class="smg-dashboard-recent">
            <div v-for="log in recentUsage" :key="log.id" class="smg-dashboard-recent-row">
              <div class="smg-dashboard-request">
                <strong>{{ log.model || 'model' }}</strong>
                <span>{{ formatTime(log.created_at) }}</span>
              </div>
              <span>{{ formatCompact((log.input_tokens || 0) + (log.output_tokens || 0)) }} tokens</span>
              <span>${{ formatMoney(log.actual_cost || 0, 4) }}</span>
              <span>{{ formatDuration(log.duration_ms) }}</span>
            </div>
          </div>
        </GlassSurface>

        <GlassSurface class="smg-dashboard-panel">
          <header class="smg-dashboard-panel-head">
            <div><span>{{ copy.next }}</span><strong>{{ copy.shortcuts }}</strong></div>
          </header>
          <nav class="smg-dashboard-shortcuts">
            <router-link v-for="action in actions" :key="action.path" :to="action.path" class="smg-dashboard-shortcut">
              <div>
                <strong>{{ action.title }}</strong>
                <span>{{ action.description }}</span>
              </div>
              <b>→</b>
            </router-link>
          </nav>
        </GlassSurface>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import GlassMetric from '@/components/glass/GlassMetric.vue'
import GlassSurface from '@/components/glass/GlassSurface.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import type { ModelStat, TrendDataPoint, UsageLog } from '@/types'
import { formatDateLocalInput } from '@/utils/format'

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const user = computed(() => authStore.user)
const copied = ref(false)

const stats = ref<UserStatsType | null>(null)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatDateLocalInput(new Date()))
const granularity = ref('day')

const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')
    .trim()
    .replace(/\/$/, ''),
)

const copy = computed(() =>
  isZh.value
    ? {
        heroTitle: '你的 API 入口、模型和用量，现在都在一个工作区里。',
        heroDescription: '这里是客户真正需要的控制面：创建 API Key、选择模型、查看调用与费用。上游账号和调度细节不会暴露到用户工作区。',
        manageKeys: '管理 API Key',
        models: '浏览模型',
        usage: '查看用量',
        quickstart: 'QUICKSTART',
        copied: '已复制 Base URL',
        copyBase: '复制 Base URL',
        trend: '调用趋势与模型分布',
        recent: '最近调用',
        viewAll: '查看全部 →',
        loading: '正在读取用量数据…',
        empty: '还没有调用记录。创建 API Key 后发起第一次请求即可看到数据。',
        next: '常用入口',
        shortcuts: 'SHORTCUTS',
        balance: '可用余额',
        keys: 'API Keys',
        requests: '今日请求',
        spend: '今日消费',
        tokens: '今日 Tokens',
        rpm: '当前 RPM',
        tpm: '当前 TPM',
        latency: '平均响应',
        active: '个启用',
        totalRequests: '累计请求',
        actualSpend: '实际结算',
        inputOutput: '输入 + 输出',
        perMinute: '每分钟请求',
        tokenRate: '每分钟 Token',
        average: '平均耗时',
      }
    : {
        heroTitle: 'Your API entry point, models and usage in one workspace.',
        heroDescription: 'This is the customer control plane: create API keys, choose models, and inspect usage and spend. Upstream accounts and scheduling details stay behind the platform.',
        manageKeys: 'Manage API Keys',
        models: 'Browse models',
        usage: 'View usage',
        quickstart: 'QUICKSTART',
        copied: 'Base URL copied',
        copyBase: 'Copy Base URL',
        trend: 'Request trend & model mix',
        recent: 'Recent requests',
        viewAll: 'View all →',
        loading: 'Loading usage data…',
        empty: 'No requests yet. Create an API key and send your first request.',
        next: 'Common actions',
        shortcuts: 'SHORTCUTS',
        balance: 'Available balance',
        keys: 'API Keys',
        requests: 'Requests today',
        spend: 'Spend today',
        tokens: 'Tokens today',
        rpm: 'Current RPM',
        tpm: 'Current TPM',
        latency: 'Avg response',
        active: 'active',
        totalRequests: 'Total requests',
        actualSpend: 'Actual billed',
        inputOutput: 'Input + output',
        perMinute: 'Requests per minute',
        tokenRate: 'Tokens per minute',
        average: 'Average duration',
      },
)

const statCards = computed(() => [
  { label: copy.value.balance, value: `$${formatMoney(user.value?.balance || 0, 2)}`, note: copy.value.actualSpend },
  { label: copy.value.keys, value: String(stats.value?.total_api_keys || 0), note: `${stats.value?.active_api_keys || 0} ${copy.value.active}` },
  { label: copy.value.requests, value: formatCompact(stats.value?.today_requests || 0), note: `${copy.value.totalRequests}: ${formatCompact(stats.value?.total_requests || 0)}` },
  { label: copy.value.spend, value: `$${formatMoney(stats.value?.today_actual_cost || 0, 4)}`, note: copy.value.actualSpend },
  { label: copy.value.tokens, value: formatCompact(stats.value?.today_tokens || 0), note: copy.value.inputOutput },
  { label: copy.value.rpm, value: formatCompact(stats.value?.rpm || 0), note: copy.value.perMinute },
  { label: copy.value.tpm, value: formatCompact(stats.value?.tpm || 0), note: copy.value.tokenRate },
  { label: copy.value.latency, value: formatDuration(stats.value?.average_duration_ms), note: copy.value.average },
])

const actions = computed(() => [
  { path: '/keys', title: copy.value.manageKeys, description: isZh.value ? '创建、停用与管理凭证' : 'Create, disable and manage credentials' },
  { path: '/model-plaza?embedded=1', title: copy.value.models, description: isZh.value ? '查看模型能力和价格' : 'Browse model capabilities and pricing' },
  { path: '/usage', title: copy.value.usage, description: isZh.value ? '定位请求、Token 和费用' : 'Inspect requests, tokens and spend' },
  { path: '/profile', title: isZh.value ? '账户设置' : 'Account settings', description: isZh.value ? '管理个人资料和安全设置' : 'Manage profile and security' },
])

function formatCompact(value: number): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(value >= 10_000_000 ? 0 : 1)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(value >= 10_000 ? 0 : 1)}K`
  return Number(value || 0).toLocaleString()
}

function formatMoney(value: number, digits: number): string {
  return Number(value || 0).toFixed(digits)
}

function formatDuration(ms: number | null | undefined): string {
  const value = Number(ms || 0)
  if (!value) return '—'
  if (value < 1000) return `${Math.round(value)} ms`
  return `${(value / 1000).toFixed(2)} s`
}

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(isZh.value ? 'zh-CN' : 'en-US', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function copyApiBase() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 1400)
  } catch (error) {
    console.warn('Failed to copy Base URL:', error)
  }
}

async function loadStats() {
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  }
}

async function loadCharts() {
  loadingCharts.value = true
  try {
    const [trend, models] = await Promise.all([
      usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }),
      usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value }),
    ])
    trendData.value = trend.trend || []
    modelStats.value = models.models || []
  } catch (error) {
    console.error('Failed to load dashboard charts:', error)
  } finally {
    loadingCharts.value = false
  }
}

async function loadRecent() {
  loadingUsage.value = true
  try {
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = res.items.slice(0, 6)
  } catch (error) {
    console.error('Failed to load recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}

function refreshAll() {
  void loadStats()
  void loadCharts()
  void loadRecent()
}

onMounted(refreshAll)
</script>
