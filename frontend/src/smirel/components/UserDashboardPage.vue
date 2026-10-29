<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AsyncDataReveal from './AsyncDataReveal.vue'
import '../styles/user-dashboard.css'

interface DashboardStats {
  total_api_keys?: number
  active_api_keys?: number
  total_requests?: number
  total_tokens?: number
  total_actual_cost?: number
  today_requests?: number
  today_tokens?: number
  today_actual_cost?: number
  [key: string]: unknown
}

const props = defineProps<{
  stats: DashboardStats | null
  balance: number
  loading: boolean
  ready: boolean
}>()

const emit = defineEmits<{
  refresh: []
}>()

const { t, locale } = useI18n()
const copied = ref(false)
const apiBase = 'https://muxway.dev/v1'

const activeKeys = computed(() => Number(props.stats?.active_api_keys || 0))
const totalKeys = computed(() => Number(props.stats?.total_api_keys || 0))
const todayLabel = computed(() => new Intl.DateTimeFormat(locale.value === 'en' ? 'en-US' : 'zh-CN', {
  month: 'short',
  day: 'numeric',
  weekday: 'short',
}).format(new Date()))

const compactNumber = computed(() => new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 2 }))

function formatCount(value: unknown) {
  return Number(value || 0).toLocaleString()
}

// Token totals grow fast; keep them readable at a glance and expose the exact value on hover.
function formatTokens(value: unknown) {
  const amount = Number(value || 0)
  return amount >= 100_000 ? compactNumber.value.format(amount) : amount.toLocaleString()
}

function formatCost(value: unknown) {
  const amount = Number(value || 0)
  if (amount > 0 && amount < 0.01) return '<$0.01'
  return `$${amount.toFixed(2)}`
}

const usageMetrics = computed(() => [
  {
    key: 'requests',
    label: t('workspace.todayRequests'),
    value: formatCount(props.stats?.today_requests),
    exact: formatCount(props.stats?.today_requests),
    total: formatCount(props.stats?.total_requests),
    icon: 'M4 12h4l2.5-6 3 12 2.5-6h4',
  },
  {
    key: 'tokens',
    label: t('workspace.todayTokens'),
    value: formatTokens(props.stats?.today_tokens),
    exact: formatCount(props.stats?.today_tokens),
    total: formatTokens(props.stats?.total_tokens),
    icon: 'M12 3.5 20 8v8l-8 4.5L4 16V8l8-4.5ZM4 8l8 4.5L20 8M12 12.5v8',
  },
  {
    key: 'cost',
    label: t('workspace.todayCost'),
    value: formatCost(props.stats?.today_actual_cost),
    exact: `$${Number(props.stats?.today_actual_cost || 0).toFixed(4)}`,
    total: formatCost(props.stats?.total_actual_cost),
    icon: 'M12 3.5v17M16.2 7.2c-.8-1-2.3-1.7-4.2-1.7-2.4 0-4 1.2-4 3s1.6 2.5 4 3 4 1.3 4 3.2-1.7 3.1-4.2 3.1c-1.9 0-3.5-.7-4.3-1.9',
  },
])

const shortcuts = computed(() => [
  { to: '/keys', title: t('nav.keys'), desc: t('userDashboard.keysDesc'), icon: 'M14.5 4.5a5 5 0 1 1-3.9 8.1L4.5 18.7v1.8h2.8v-1.8h1.8v-1.8h1.8l1.1-1.1A5 5 0 0 1 14.5 4.5ZM16.4 8.2h.01' },
  { to: '/model-plaza', title: t('userDashboard.models'), desc: t('userDashboard.modelsDesc'), icon: 'M12 3.5 20.5 8 12 12.5 3.5 8 12 3.5ZM3.5 12 12 16.5 20.5 12M3.5 16 12 20.5 20.5 16' },
  { to: '/usage', title: t('nav.usage'), desc: t('userDashboard.usageDesc'), icon: 'M4.5 4.5v15h15M8.5 15.5v-4M12.5 15.5v-7M16.5 15.5v-5' },
  { to: '/support', title: t('userDashboard.support'), desc: t('userDashboard.supportDesc'), icon: 'M12 20.5a8.5 8.5 0 1 0 0-17 8.5 8.5 0 0 0 0 17ZM9.6 9.4a2.5 2.5 0 0 1 4.8.9c0 1.7-2.4 2.1-2.4 3.6M12 16.6v.1' },
])

async function copyEndpoint() {
  await navigator.clipboard.writeText(apiBase)
  copied.value = true
  window.setTimeout(() => { copied.value = false }, 1400)
}
</script>

<template>
  <div class="user-console-page">
    <header class="page-heading dash-heading">
      <div>
        <h1>{{ t('nav.dashboard') }}</h1>
        <p>{{ t('workspace.descriptions.dashboard') }}</p>
      </div>
      <button class="ghost-button dash-refresh" type="button" :disabled="loading" @click="emit('refresh')">
        <svg :class="{ spinning: loading }" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M20 11a8 8 0 1 0 2 5.3" />
          <path d="M20 4v7h-7" />
        </svg>
        <span>{{ loading ? t('workspace.refreshing') : t('workspace.refresh') }}</span>
      </button>
    </header>

    <AsyncDataReveal :ready="ready" :content-motion="false" class="user-dashboard-data">
      <template #loading>
        <div class="user-dashboard-loading-layout" role="status" :aria-label="t('workspace.refreshing')">
          <div class="user-dashboard-loading-overview" aria-hidden="true">
            <div v-for="index in 2" :key="index" class="user-dashboard-loading-tile">
              <i class="async-data-skeleton" /><i class="async-data-skeleton" /><i class="async-data-skeleton" />
            </div>
          </div>
          <div class="user-dashboard-loading-metrics" aria-hidden="true">
            <i class="async-data-skeleton" />
            <div><i v-for="index in 3" :key="index" class="async-data-skeleton" /></div>
          </div>
          <div class="user-dashboard-loading-shortcuts" aria-hidden="true">
            <i v-for="index in 4" :key="index" class="async-data-skeleton" />
          </div>
        </div>
      </template>

      <section class="dash-overview">
        <article class="dash-card dash-balance">
          <div class="dash-card-title">
            <span class="dash-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24"><path d="M4.5 7.5c0-1.1.9-2 2-2h11v3M4.5 7.5v10c0 1.1.9 2 2 2h13v-11h-13a2 2 0 0 1-2-2Z" /><path d="M16 14h.01" /></svg>
            </span>
            <span>{{ t('workspace.availableBalance') }}</span>
          </div>
          <strong class="dash-balance-value">${{ balance.toFixed(2) }}</strong>
          <RouterLink to="/keys" class="dash-key-chip">
            <i :class="{ idle: activeKeys === 0 }" aria-hidden="true"></i>
            {{ totalKeys > 0 ? t('userDashboard.activeKeys', { active: activeKeys, total: totalKeys }) : t('userDashboard.noKeys') }}
          </RouterLink>
          <div class="dash-balance-actions">
            <RouterLink to="/purchase" class="dash-button primary">{{ t('workspace.purchase') }}</RouterLink>
            <RouterLink to="/keys" class="dash-button">{{ t('userDashboard.manageKeys') }}</RouterLink>
          </div>
        </article>

        <article class="dash-card dash-endpoint">
          <header class="dash-card-head">
            <div>
              <h2>{{ t('userDashboard.endpointTitle') }}</h2>
              <p>{{ t('userDashboard.endpointHint') }}</p>
            </div>
            <RouterLink to="/model-plaza" class="dash-link">{{ t('userDashboard.models') }} →</RouterLink>
          </header>

          <div class="dash-endpoint-url">
            <span>Base URL</span>
            <code>{{ apiBase }}</code>
            <button type="button" :class="{ copied }" @click="copyEndpoint">
              <svg v-if="!copied" viewBox="0 0 24 24" aria-hidden="true"><rect x="8.5" y="8.5" width="11" height="11" rx="2" /><path d="M15.5 8.5V6.5a2 2 0 0 0-2-2h-7a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h2" /></svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="m5.5 12.5 4 4 9-9" /></svg>
              {{ copied ? t('userDashboard.copied') : t('userDashboard.copy') }}
            </button>
          </div>

          <pre class="dash-snippet"><code><span class="dim">$</span> curl {{ apiBase }}/chat/completions \
  -H <span class="str">"Authorization: Bearer $MUXWAY_API_KEY"</span> \
  -d <span class="str">'{"model": "gpt-5.6", "messages": [...]}'</span></code></pre>

          <div class="dash-endpoint-tags">
            <span><i aria-hidden="true"></i>{{ t('userDashboard.openaiCompatible') }}</span>
            <span>{{ t('userDashboard.bearerAuth') }}</span>
          </div>
        </article>
      </section>

      <section class="dash-card dash-usage">
        <header class="dash-card-head">
          <div class="dash-usage-title">
            <h2>{{ t('userDashboard.usageTitle') }}</h2>
            <span>{{ todayLabel }}</span>
          </div>
          <RouterLink to="/usage" class="dash-link">{{ t('userDashboard.viewUsage') }} →</RouterLink>
        </header>

        <div class="dash-usage-grid">
          <article v-for="metric in usageMetrics" :key="metric.key" :class="`metric-${metric.key}`">
            <div class="dash-metric-label">
              <span class="dash-metric-icon" aria-hidden="true"><svg viewBox="0 0 24 24"><path :d="metric.icon" /></svg></span>
              {{ metric.label }}
            </div>
            <strong :title="metric.exact">{{ metric.value }}</strong>
            <small>{{ t('userDashboard.lifetimeLabel') }} <b>{{ metric.total }}</b></small>
          </article>
        </div>
      </section>

      <section class="dash-shortcuts" :aria-label="t('userDashboard.quickTitle')">
        <RouterLink v-for="item in shortcuts" :key="item.to" :to="item.to" class="dash-card dash-shortcut">
          <span class="dash-icon" aria-hidden="true"><svg viewBox="0 0 24 24"><path :d="item.icon" /></svg></span>
          <span class="dash-shortcut-copy">
            <strong>{{ item.title }}</strong>
            <small>{{ item.desc }}</small>
          </span>
          <b aria-hidden="true">→</b>
        </RouterLink>
      </section>
    </AsyncDataReveal>
  </div>
</template>

<style scoped>
.user-dashboard-loading-layout { display: grid; gap: 16px; }
.user-dashboard-loading-overview { display: grid; grid-template-columns: .8fr 1.2fr; gap: 16px; }
.user-dashboard-loading-tile {
  min-height: 236px;
  padding: 24px;
  border: 1px solid var(--ws-border, rgba(127, 145, 161, .18));
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 20px;
}
.user-dashboard-loading-tile i:nth-child(1) { width: 34%; height: 12px; }
.user-dashboard-loading-tile i:nth-child(2) { width: 56%; height: 34px; }
.user-dashboard-loading-tile i:nth-child(3) { width: 77%; height: 12px; }
.user-dashboard-loading-metrics {
  min-height: 190px;
  padding: 24px;
  border: 1px solid var(--ws-border, rgba(127, 145, 161, .18));
  border-radius: 14px;
  display: grid;
  gap: 24px;
}
.user-dashboard-loading-metrics > i { width: 24%; height: 14px; }
.user-dashboard-loading-metrics > div { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.user-dashboard-loading-metrics > div i { height: 86px; }
.user-dashboard-loading-shortcuts { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
.user-dashboard-loading-shortcuts i { height: 76px; border-radius: 14px; }
@media (max-width: 760px) {
  .user-dashboard-loading-overview,
  .user-dashboard-loading-shortcuts { grid-template-columns: 1fr; }
}
</style>
