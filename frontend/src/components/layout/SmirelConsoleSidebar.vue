<template>
  <aside
    class="sidebar smirel-console-sidebar"
    :class="[
      sidebarCollapsed ? 'w-[72px]' : 'w-64',
      { '-translate-x-full lg:translate-x-0': !mobileOpen },
    ]"
  >
    <div class="sidebar-header" :class="{ 'sidebar-header-collapsed': sidebarCollapsed }">
      <router-link
        :to="homePath"
        class="sidebar-logo flex h-9 w-9 items-center justify-center overflow-hidden"
        @click="closeMobile"
      >
        <img :src="siteLogo || '/smirel-mark.svg'" alt="Smirel" class="h-full w-full object-contain" />
      </router-link>
      <div class="sidebar-brand" :class="{ 'sidebar-brand-collapsed': sidebarCollapsed }">
        <router-link :to="homePath" class="sidebar-brand-title" @click="closeMobile">
          {{ siteName || 'Smirel API' }}
        </router-link>
        <span class="smirel-console-brand-caption">{{ isAdmin ? copy.adminConsole : copy.developerConsole }}</span>
      </div>
    </div>

    <nav class="sidebar-nav scrollbar-hide">
      <section v-for="section in navSections" :key="section.id" class="sidebar-section">
        <div v-if="!sidebarCollapsed" class="sidebar-section-title">
          {{ section.label }}
        </div>

        <router-link
          v-for="item in section.items"
          :key="item.path"
          :to="item.path"
          class="sidebar-link"
          :class="{
            'sidebar-link-active': isActive(item),
            'sidebar-link-collapsed': sidebarCollapsed,
          }"
          :title="sidebarCollapsed ? item.label : undefined"
          @click="closeMobile"
        >
          <svg
            class="h-5 w-5 flex-shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.6"
            aria-hidden="true"
          >
            <path
              v-if="item.kind === 'overview'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4 4h6v6H4V4zm10 0h6v6h-6V4zM4 14h6v6H4v-6zm10 0h6v6h-6v-6z"
            />
            <path
              v-else-if="item.kind === 'key'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z"
            />
            <path
              v-else-if="item.kind === 'chart'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4 19V9m6 10V5m6 14v-7m4 7H2"
            />
            <path
              v-else-if="item.kind === 'network'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 3v4m0 10v4M3 12h4m10 0h4M6.35 6.35l2.8 2.8m5.7 5.7 2.8 2.8m0-11.3-2.8 2.8m-5.7 5.7-2.8 2.8M12 8a4 4 0 100 8 4 4 0 000-8z"
            />
            <path
              v-else-if="item.kind === 'security'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 3l7 3v5c0 4.4-2.8 8.1-7 10-4.2-1.9-7-5.6-7-10V6l7-3zm-2 9 1.4 1.4L15 9.8"
            />
            <path
              v-else-if="item.kind === 'user'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.5 20a7.5 7.5 0 0115 0"
            />
            <path
              v-else-if="item.kind === 'settings'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 8.25A3.75 3.75 0 1112 15.75 3.75 3.75 0 0112 8.25zm0-5.25 1 2.1a7.9 7.9 0 012 .8l2.2-.8 1.7 1.7-.8 2.2c.35.63.62 1.3.8 2L21 12l-2.1 1c-.18.7-.45 1.37-.8 2l.8 2.2-1.7 1.7-2.2-.8a7.9 7.9 0 01-2 .8L12 21l-1-2.1a7.9 7.9 0 01-2-.8l-2.2.8-1.7-1.7.8-2.2a7.9 7.9 0 01-.8-2L3 12l2.1-1a7.9 7.9 0 01.8-2l-.8-2.2 1.7-1.7 2.2.8a7.9 7.9 0 012-.8L12 3z"
            />
            <path
              v-else
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M8 12h8m-4-4 4 4-4 4M5 5h5m-5 14h5"
            />
          </svg>
          <span
            class="sidebar-label"
            :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
          >
            {{ item.label }}
          </span>
        </router-link>
      </section>
    </nav>

    <div class="smirel-console-sidebar-footer mt-auto border-t">
      <button
        type="button"
        class="sidebar-link w-full"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? copy.theme : undefined"
        @click="toggleTheme"
      >
        <svg class="h-5 w-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
          <path v-if="isDark" stroke-linecap="round" stroke-linejoin="round" d="M12 3v2m0 14v2M3 12h2m14 0h2M5.6 5.6 7 7m10 10 1.4 1.4M18.4 5.6 17 7M7 17l-1.4 1.4M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
          <path v-else stroke-linecap="round" stroke-linejoin="round" d="M20.5 15.5A8 8 0 018.5 3.5 8.5 8.5 0 1020.5 15.5z" />
        </svg>
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }">
          {{ isDark ? copy.lightMode : copy.darkMode }}
        </span>
      </button>

      <button
        type="button"
        class="sidebar-link mt-1 hidden w-full lg:flex"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? copy.expand : undefined"
        @click="toggleSidebar"
      >
        <svg class="h-5 w-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            :d="sidebarCollapsed ? 'M9 5l7 7-7 7' : 'M15 5l-7 7 7 7'"
          />
        </svg>
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }">
          {{ copy.collapse }}
        </span>
      </button>
    </div>
  </aside>

  <transition name="fade">
    <button
      v-if="mobileOpen"
      type="button"
      class="fixed inset-0 z-30 bg-black/45 lg:hidden"
      aria-label="Close navigation"
      @click="closeMobile"
    ></button>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAppStore, useAuthStore } from '@/stores'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'

type NavKind = 'overview' | 'key' | 'chart' | 'network' | 'security' | 'user' | 'settings' | 'link'

type NavItem = {
  path: string
  label: string
  kind: NavKind
  exact?: boolean
}

type NavSection = {
  id: string
  label: string
  items: NavItem[]
}

const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)
const isDark = ref(document.documentElement.classList.contains('dark'))
const isZh = computed(() => String(locale.value).toLowerCase().startsWith('zh'))
const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const siteName = computed(() => appStore.siteName || 'Smirel API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

const copy = computed(() => isZh.value ? {
  developerConsole: '开发者控制台', adminConsole: '运营控制台', overview: '概览', api: 'API 与模型',
  billing: '计费与账户', operations: '运营总览', access: '用户与权限', supply: '上游与路由',
  commerce: '商业化', governance: '平台治理', personal: '我的账户', theme: '切换主题',
  lightMode: '浅色模式', darkMode: '深色模式', collapse: '收起导航', expand: '展开导航',
} : {
  developerConsole: 'Developer Console', adminConsole: 'Operations Console', overview: 'Overview', api: 'API & Models',
  billing: 'Billing & Account', operations: 'Operations', access: 'Users & Access', supply: 'Upstream & Routing',
  commerce: 'Commerce', governance: 'Governance', personal: 'My Account', theme: 'Switch theme',
  lightMode: 'Light mode', darkMode: 'Dark mode', collapse: 'Collapse navigation', expand: 'Expand navigation',
})

const userCustomItems = computed<NavItem[]>(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
    .map((item) => ({ path: `/custom/${item.id}`, label: String(item.label), kind: 'link' }))
})

const adminCustomItems = computed<NavItem[]>(() => adminSettingsStore.customMenuItems
  .filter((item) => item.visibility === 'admin')
  .slice()
  .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
  .map((item) => ({ path: `/custom/${item.id}`, label: String(item.label), kind: 'link' })))

const userSections = computed<NavSection[]>(() => {
  const simple = authStore.isSimpleMode
  const apiItems: NavItem[] = [
    { path: '/keys', label: t('nav.apiKeys'), kind: 'key' },
    { path: '/usage', label: t('nav.usage'), kind: 'chart' },
  ]

  if (!simple && isFeatureFlagEnabled(FeatureFlags.availableChannels)) {
    apiItems.push({ path: '/available-channels', label: t('nav.availableChannels'), kind: 'network' })
  }
  if (isFeatureFlagEnabled(FeatureFlags.channelMonitor)) {
    apiItems.push({ path: '/monitor', label: t('nav.channelStatus'), kind: 'network' })
  }
  if (!simple && isFeatureFlagEnabled(FeatureFlags.modelPlaza)) {
    apiItems.push({ path: '/model-plaza?embedded=1', label: t('nav.modelPlaza'), kind: 'overview' })
  }
  if (!simple && canUseBatchImage.value) {
    apiItems.push({ path: '/batch-image', label: t('nav.batchImage'), kind: 'overview' })
  }

  const accountItems: NavItem[] = []
  if (!simple) {
    accountItems.push({ path: '/subscriptions', label: t('nav.mySubscriptions'), kind: 'link' })
    if (isFeatureFlagEnabled(FeatureFlags.payment)) {
      accountItems.push(
        { path: '/purchase', label: t('nav.buySubscription'), kind: 'link' },
        { path: '/orders', label: t('nav.myOrders'), kind: 'link' },
      )
    }
    accountItems.push({ path: '/redeem', label: t('nav.redeem'), kind: 'link' })
    if (isFeatureFlagEnabled(FeatureFlags.affiliate)) {
      accountItems.push({ path: '/affiliate', label: t('nav.affiliate'), kind: 'link' })
    }
  }
  accountItems.push({ path: '/profile', label: t('nav.profile'), kind: 'user' }, ...userCustomItems.value)

  return [
    { id: 'overview', label: copy.value.overview, items: [{ path: '/dashboard', label: t('nav.dashboard'), kind: 'overview', exact: true }] },
    { id: 'api', label: copy.value.api, items: apiItems },
    { id: 'billing', label: copy.value.billing, items: accountItems },
  ]
})

const adminSections = computed<NavSection[]>(() => {
  const simple = authStore.isSimpleMode
  const operations: NavItem[] = [{ path: '/admin/dashboard', label: t('nav.dashboard'), kind: 'overview', exact: true }]
  if (!simple && adminSettingsStore.opsMonitoringEnabled) {
    operations.push({ path: '/admin/ops', label: t('nav.ops'), kind: 'chart' })
  }

  const access: NavItem[] = simple ? [] : [
    { path: '/admin/users', label: t('nav.users'), kind: 'user' },
    { path: '/admin/groups', label: t('nav.groups'), kind: 'security' },
  ]

  const supply: NavItem[] = [{ path: '/admin/accounts', label: t('nav.accounts'), kind: 'network' }]
  if (!simple) {
    supply.push({ path: '/admin/channels/pricing', label: t('nav.channelPricing'), kind: 'network' })
    if (isFeatureFlagEnabled(FeatureFlags.channelMonitor)) {
      supply.push({ path: '/admin/channels/monitor', label: t('nav.channelMonitor'), kind: 'chart' })
    }
    supply.push({ path: '/admin/proxies', label: t('nav.proxies'), kind: 'network' })
  }

  const commerce: NavItem[] = []
  if (!simple) {
    commerce.push({ path: '/admin/subscriptions', label: t('nav.subscriptions'), kind: 'link' })
    if (adminSettingsStore.paymentEnabled) {
      commerce.push(
        { path: '/admin/orders/dashboard', label: t('nav.paymentDashboard'), kind: 'chart' },
        { path: '/admin/orders', label: t('nav.orderManagement'), kind: 'link', exact: true },
        { path: '/admin/orders/plans', label: t('nav.paymentPlans'), kind: 'link' },
      )
    }
    commerce.push(
      { path: '/admin/redeem', label: t('nav.redeemCodes'), kind: 'link' },
      { path: '/admin/promo-codes', label: t('nav.promoCodes'), kind: 'link' },
    )
    if (isFeatureFlagEnabled(FeatureFlags.affiliate)) {
      commerce.push(
        { path: '/admin/affiliates/invites', label: t('nav.affiliateInviteRecords'), kind: 'link' },
        { path: '/admin/affiliates/rebates', label: t('nav.affiliateRebateRecords'), kind: 'link' },
        { path: '/admin/affiliates/transfers', label: t('nav.affiliateTransferRecords'), kind: 'link' },
      )
    }
  }

  const governance: NavItem[] = [{ path: '/admin/usage', label: t('nav.usage'), kind: 'chart' }]
  if (!simple) {
    governance.push(
      { path: '/admin/announcements', label: t('nav.announcements'), kind: 'link' },
      { path: '/admin/audit-logs', label: t('nav.auditLogs'), kind: 'security' },
    )
    if (isFeatureFlagEnabled(FeatureFlags.riskControl)) {
      governance.push(
        { path: '/admin/risk-control', label: t('nav.contentModeration'), kind: 'security' },
        { path: '/admin/prompt-audit', label: t('nav.promptAudit'), kind: 'security' },
      )
    }
  }
  governance.push({ path: '/admin/settings', label: t('nav.settings'), kind: 'settings' }, ...adminCustomItems.value)

  return [
    { id: 'operations', label: copy.value.operations, items: operations },
    { id: 'access', label: copy.value.access, items: access },
    { id: 'supply', label: copy.value.supply, items: supply },
    { id: 'commerce', label: copy.value.commerce, items: commerce },
    { id: 'governance', label: copy.value.governance, items: governance },
    { id: 'personal', label: copy.value.personal, items: [
      { path: '/keys', label: t('nav.apiKeys'), kind: 'key' },
      { path: '/profile', label: t('nav.profile'), kind: 'user' },
    ] },
  ].filter((section) => section.items.length > 0)
})

const navSections = computed(() => isAdmin.value ? adminSections.value : userSections.value)

function isActive(item: NavItem) {
  const target = item.path.split('?')[0]
  return item.exact ? route.path === target : route.path === target || route.path.startsWith(`${target}/`)
}

function closeMobile() {
  if (mobileOpen.value) appStore.setMobileOpen(false)
}

function toggleSidebar() {
  appStore.toggleSidebar()
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

watch(isAdmin, (value) => {
  if (value) void adminSettingsStore.fetch()
}, { immediate: true })

onMounted(() => {
  void refreshBatchImageAccess()
})
</script>

<style scoped>
.smirel-console-brand-caption {
  display: block;
  margin-top: 2px;
  overflow: hidden;
  color: #718784;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.045em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-brand {
  min-width: 0;
  flex: 1 1 auto;
  max-width: 12rem;
  white-space: nowrap;
  transition: max-width 0.2s ease, opacity 0.14s ease, transform 0.14s ease;
}

.sidebar-brand-collapsed,
.sidebar-label-collapsed {
  max-width: 0 !important;
  overflow: hidden;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

.sidebar-header-collapsed {
  gap: 0;
  padding-left: 1.125rem !important;
  padding-right: 1.125rem !important;
}

.sidebar-label {
  display: block;
  min-width: 0;
  max-width: 12rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: max-width 0.2s ease, opacity 0.12s ease, transform 0.12s ease;
}

.sidebar-link-collapsed {
  gap: 0 !important;
  justify-content: center;
  padding-left: 0.7rem !important;
  padding-right: 0.7rem !important;
}

.smirel-console-sidebar-footer {
  border-color: rgba(255, 255, 255, 0.07) !important;
  padding: 10px 12px 14px;
}
</style>
