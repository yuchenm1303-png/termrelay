<template>
  <aside
    class="sidebar smirel-console-sidebar"
    :class="[
      sidebarCollapsed ? 'w-[72px]' : 'w-64',
      { '-translate-x-full lg:translate-x-0': !mobileOpen }
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
        <div
          v-if="!sidebarCollapsed"
          class="sidebar-section-title"
        >
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
          :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
          @click="handleItemClick(item.path)"
        >
          <Icon :name="item.icon" size="md" class="flex-shrink-0" />
          <span
            class="sidebar-label"
            :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
            :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
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
        <Icon :name="isDark ? 'lightbulb' : 'clock'" size="md" class="flex-shrink-0" />
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
        <Icon :name="sidebarCollapsed ? 'arrowRight' : 'arrowLeft'" size="md" class="flex-shrink-0" />
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
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'

type IconName =
  | 'grid'
  | 'key'
  | 'chart'
  | 'globe'
  | 'shield'
  | 'cog'
  | 'user'
  | 'link'
  | 'clock'

type NavItem = {
  path: string
  label: string
  icon: IconName
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
const onboardingStore = useOnboardingStore()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)
const isDark = ref(document.documentElement.classList.contains('dark'))
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const siteName = computed(() => appStore.siteName || 'Smirel API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '/smirel-mark.svg', { allowRelative: true, allowDataUrl: true }),
)

const copy = computed(() =>
  isZh.value
    ? {
        developerConsole: '开发者控制台',
        adminConsole: '运营控制台',
        overview: '概览',
        api: 'API 与模型',
        billing: '计费与账户',
        operations: '运营总览',
        access: '用户与权限',
        supply: '上游与路由',
        commerce: '商业化',
        governance: '平台治理',
        personal: '我的账户',
        theme: '切换主题',
        lightMode: '浅色模式',
        darkMode: '深色模式',
        collapse: '收起导航',
        expand: '展开导航',
      }
    : {
        developerConsole: 'Developer Console',
        adminConsole: 'Operations Console',
        overview: 'Overview',
        api: 'API & Models',
        billing: 'Billing & Account',
        operations: 'Operations',
        access: 'Users & Access',
        supply: 'Upstream & Routing',
        commerce: 'Commerce',
        governance: 'Governance',
        personal: 'My Account',
        theme: 'Switch theme',
        lightMode: 'Light mode',
        darkMode: 'Dark mode',
        collapse: 'Collapse navigation',
        expand: 'Expand navigation',
      },
)

function enabled(flag: (typeof FeatureFlags)[keyof typeof FeatureFlags]) {
  return isFeatureFlagEnabled(flag)
}

const userCustomItems = computed<NavItem[]>(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
    .map((item) => ({ path: `/custom/${item.id}`, label: item.label, icon: 'link' as const }))
})

const adminCustomItems = computed<NavItem[]>(() =>
  adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
    .map((item) => ({ path: `/custom/${item.id}`, label: item.label, icon: 'link' as const })),
)

const userSections = computed<NavSection[]>(() => {
  const simple = authStore.isSimpleMode
  const apiItems: NavItem[] = [
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    { path: '/usage', label: t('nav.usage'), icon: 'chart' },
  ]

  if (!simple && enabled(FeatureFlags.availableChannels)) {
    apiItems.push({ path: '/available-channels', label: t('nav.availableChannels'), icon: 'globe' })
  }
  if (enabled(FeatureFlags.channelMonitor)) {
    apiItems.push({ path: '/monitor', label: t('nav.channelStatus'), icon: 'clock' })
  }
  if (!simple && enabled(FeatureFlags.modelPlaza)) {
    apiItems.push({ path: '/model-plaza?embedded=1', label: t('nav.modelPlaza'), icon: 'grid' })
  }
  if (!simple && canUseBatchImage.value) {
    apiItems.push({ path: '/batch-image', label: t('nav.batchImage'), icon: 'grid' })
  }

  const billingItems: NavItem[] = []
  if (!simple) {
    billingItems.push({ path: '/subscriptions', label: t('nav.mySubscriptions'), icon: 'link' })
    if (enabled(FeatureFlags.payment)) {
      billingItems.push(
        { path: '/purchase', label: t('nav.buySubscription'), icon: 'link' },
        { path: '/orders', label: t('nav.myOrders'), icon: 'link' },
      )
    }
    billingItems.push({ path: '/redeem', label: t('nav.redeem'), icon: 'link' })
    if (enabled(FeatureFlags.affiliate)) {
      billingItems.push({ path: '/affiliate', label: t('nav.affiliate'), icon: 'link' })
    }
  }
  billingItems.push({ path: '/profile', label: t('nav.profile'), icon: 'user' })
  billingItems.push(...userCustomItems.value)

  return [
    {
      id: 'overview',
      label: copy.value.overview,
      items: [{ path: '/dashboard', label: t('nav.dashboard'), icon: 'grid', exact: true }],
    },
    { id: 'api', label: copy.value.api, items: apiItems },
    { id: 'billing', label: copy.value.billing, items: billingItems },
  ].filter((section) => section.items.length > 0)
})

const adminSections = computed<NavSection[]>(() => {
  const simple = authStore.isSimpleMode
  const operations: NavItem[] = [
    { path: '/admin/dashboard', label: t('nav.dashboard'), icon: 'grid', exact: true },
  ]
  if (!simple && adminSettingsStore.opsMonitoringEnabled) {
    operations.push({ path: '/admin/ops', label: t('nav.ops'), icon: 'chart' })
  }

  const access: NavItem[] = simple
    ? []
    : [
        { path: '/admin/users', label: t('nav.users'), icon: 'user' },
        { path: '/admin/groups', label: t('nav.groups'), icon: 'shield' },
      ]

  const supply: NavItem[] = [
    { path: '/admin/accounts', label: t('nav.accounts'), icon: 'globe' },
  ]
  if (!simple) {
    supply.push({ path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: 'link' })
    if (enabled(FeatureFlags.channelMonitor)) {
      supply.push({ path: '/admin/channels/monitor', label: t('nav.channelMonitor'), icon: 'clock' })
    }
    supply.push({ path: '/admin/proxies', label: t('nav.proxies'), icon: 'globe' })
  }

  const commerce: NavItem[] = []
  if (!simple) {
    commerce.push({ path: '/admin/subscriptions', label: t('nav.subscriptions'), icon: 'link' })
    if (adminSettingsStore.paymentEnabled) {
      commerce.push(
        { path: '/admin/orders/dashboard', label: t('nav.paymentDashboard'), icon: 'chart' },
        { path: '/admin/orders', label: t('nav.orderManagement'), icon: 'link', exact: true },
        { path: '/admin/orders/plans', label: t('nav.paymentPlans'), icon: 'link' },
      )
    }
    commerce.push(
      { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: 'link' },
      { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: 'link' },
    )
    if (enabled(FeatureFlags.affiliate)) {
      commerce.push(
        { path: '/admin/affiliates/invites', label: t('nav.affiliateInviteRecords'), icon: 'link' },
        { path: '/admin/affiliates/rebates', label: t('nav.affiliateRebateRecords'), icon: 'link' },
        { path: '/admin/affiliates/transfers', label: t('nav.affiliateTransferRecords'), icon: 'link' },
      )
    }
  }

  const governance: NavItem[] = [
    { path: '/admin/usage', label: t('nav.usage'), icon: 'chart' },
  ]
  if (!simple) {
    governance.push(
      { path: '/admin/announcements', label: t('nav.announcements'), icon: 'link' },
      { path: '/admin/audit-logs', label: t('nav.auditLogs'), icon: 'shield' },
    )
    if (enabled(FeatureFlags.riskControl)) {
      governance.push(
        { path: '/admin/risk-control', label: t('nav.contentModeration'), icon: 'shield' },
        { path: '/admin/prompt-audit', label: t('nav.promptAudit'), icon: 'shield' },
      )
    }
  }
  governance.push({ path: '/admin/settings', label: t('nav.settings'), icon: 'cog' })
  governance.push(...adminCustomItems.value)

  const personal: NavItem[] = [
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    { path: '/profile', label: t('nav.profile'), icon: 'user' },
  ]

  return [
    { id: 'operations', label: copy.value.operations, items: operations },
    { id: 'access', label: copy.value.access, items: access },
    { id: 'supply', label: copy.value.supply, items: supply },
    { id: 'commerce', label: copy.value.commerce, items: commerce },
    { id: 'governance', label: copy.value.governance, items: governance },
    { id: 'personal', label: copy.value.personal, items: personal },
  ].filter((section) => section.items.length > 0)
})

const navSections = computed(() => (isAdmin.value ? adminSections.value : userSections.value))

function isActive(item: NavItem) {
  const current = route.path
  const target = item.path.split('?')[0]
  return item.exact ? current === target : current === target || current.startsWith(`${target}/`)
}

function closeMobile() {
  if (mobileOpen.value) appStore.setMobileOpen(false)
}

function handleItemClick(path: string) {
  closeMobile()
  const selector = path === '/keys' ? '[data-tour="sidebar-my-keys"]' : undefined
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function toggleSidebar() {
  appStore.toggleSidebar()
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

watch(
  isAdmin,
  (value) => {
    if (value) void adminSettingsStore.fetch()
  },
  { immediate: true },
)

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
  white-space: nowrap;
  max-width: 12rem;
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
