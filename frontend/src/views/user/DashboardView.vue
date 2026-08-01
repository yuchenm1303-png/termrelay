<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="termrelay-dashboard-hero">
        <div>
          <p class="termrelay-dashboard-kicker">TERMRELAY / CONTROL CENTER</p>
          <h2>{{ copy.welcome }}, {{ operatorName }}</h2>
          <p>{{ copy.description }}</p>
        </div>

        <div class="termrelay-dashboard-status" :aria-label="copy.statusLabel">
          <div>
            <span>{{ copy.gateway }}</span>
            <strong>{{ copy.online }}</strong>
          </div>
          <div>
            <span>{{ copy.route }}</span>
            <strong>Responses API</strong>
          </div>
          <div>
            <span>{{ copy.auth }}</span>
            <strong>Bearer Key</strong>
          </div>
          <div>
            <span>{{ copy.endpoint }}</span>
            <strong>/v1</strong>
          </div>
        </div>
      </section>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <UserDashboardStats
          :stats="stats"
          :balance="user?.balance || 0"
          :is-simple="authStore.isSimpleMode"
          :platform-quotas="platformQuotas"
        />

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

        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2">
            <UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" />
          </div>
          <div class="lg:col-span-1">
            <UserDashboardQuickActions />
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'
import { formatDateLocalInput } from '@/utils/format'

const { locale } = useI18n()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const copy = computed(() =>
  locale.value === 'zh'
    ? {
        welcome: '欢迎回来',
        description:
          '这里是你的 AI 网关控制中心。统一管理 API Key、调用用量、模型请求和上游服务状态。',
        statusLabel: '网关状态',
        gateway: '网关',
        online: '运行正常',
        route: '协议路由',
        auth: '下游认证',
        endpoint: 'API 入口'
      }
    : {
        welcome: 'Welcome back',
        description:
          'This is your AI gateway control center for API keys, usage, model traffic and upstream health.',
        statusLabel: 'Gateway status',
        gateway: 'Gateway',
        online: 'Operational',
        route: 'Protocol route',
        auth: 'Client auth',
        endpoint: 'API endpoint'
      }
)

const operatorName = computed(() => user.value?.email?.split('@')[0] || 'Operator')
const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatDateLocalInput(new Date()))
const granularity = ref('day')

const loadStats = async () => {
  loading.value = true
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const res = await Promise.all([
      usageAPI.getDashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value as any
      }),
      usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })
    ])
    trendData.value = res[0].trend || []
    modelStats.value = res[1].models || []
  } catch (error) {
    console.error('Failed to load charts:', error)
  } finally {
    loadingCharts.value = false
  }
}

const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = res.items.slice(0, 5)
  } catch (error) {
    console.error('Failed to load recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}

const loadPlatformQuotas = async () => {
  try {
    const data = await getMyPlatformQuotas()
    platformQuotas.value = data.platform_quotas ?? []
  } catch (error) {
    console.warn('Failed to load platform quotas:', error)
    platformQuotas.value = []
  }
}

const refreshAll = () => {
  loadStats()
  loadCharts()
  loadRecent()
  loadPlatformQuotas()
}

onMounted(() => {
  refreshAll()
})
</script>
