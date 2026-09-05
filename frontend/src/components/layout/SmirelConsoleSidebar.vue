<template>
  <aside
    class="smv3-sidebar"
    :class="[
      sidebarCollapsed ? 'smv3-sidebar--collapsed' : '',
      mobileOpen ? 'smv3-sidebar--mobile-open' : '',
    ]"
  >
    <div class="smv3-sidebar-brand">
      <router-link :to="homePath" class="smv3-brand-mark" @click="closeMobile">
        <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
        <span v-else>S</span>
      </router-link>
      <div v-if="!sidebarCollapsed" class="smv3-brand-copy">
        <router-link :to="homePath" @click="closeMobile">{{ siteName }}</router-link>
        <span>{{ isAdmin ? copy.adminConsole : copy.developerConsole }}</span>
      </div>
      <button class="smv3-collapse-button" type="button" @click="appStore.toggleSidebar()">
        <span>{{ sidebarCollapsed ? '›' : '‹' }}</span>
      </button>
    </div>

    <div v-if="!sidebarCollapsed" class="smv3-workspace-card">
      <div class="smv3-workspace-dot"></div>
      <div>
        <strong>{{ isAdmin ? copy.platformWorkspace : copy.apiWorkspace }}</strong>
        <span>{{ isAdmin ? copy.platformWorkspaceHint : copy.apiWorkspaceHint }}</span>
      </div>
    </div>

    <nav class="smv3-sidebar-nav">
      <section v-for="section in navSections" :key="section.label" class="smv3-nav-section">
        <div v-if="!sidebarCollapsed" class="smv3-nav-section-label">{{ section.label }}</div>
        <router-link
          v-for="item in section.items"
          :key="item.path"
          :to="item.path"
          class="smv3-nav-item"
          :class="{ 'smv3-nav-item--active': isActive(item.path) }"
          :title="sidebarCollapsed ? item.label : undefined"
          @click="closeMobile"
        >
          <span class="smv3-nav-glyph">{{ item.glyph }}</span>
          <span v-if="!sidebarCollapsed" class="smv3-nav-label">{{ item.label }}</span>
          <span v-if="!sidebarCollapsed && item.meta" class="smv3-nav-meta">{{ item.meta }}</span>
        </router-link>
      </section>
    </nav>

    <div class="smv3-sidebar-footer">
      <button class="smv3-footer-action" type="button" @click="toggleTheme">
        <span class="smv3-nav-glyph">{{ isDark ? '☀' : '◐' }}</span>
        <span v-if="!sidebarCollapsed">{{ isDark ? copy.lightMode : copy.darkMode }}</span>
      </button>
      <router-link to="/profile" class="smv3-account-chip" @click="closeMobile">
        <span class="smv3-avatar">{{ initials }}</span>
        <span v-if="!sidebarCollapsed" class="smv3-account-copy">
          <strong>{{ displayName }}</strong>
          <span>{{ user?.email || copy.account }}</span>
        </span>
      </router-link>
    </div>
  </aside>

  <button
    v-if="mobileOpen"
    type="button"
    class="smv3-mobile-scrim"
    aria-label="Close navigation"
    @click="closeMobile"
  ></button>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

interface NavItem {
  path: string
  label: string
  glyph: string
  meta?: string
  enabled?: boolean
}

interface NavSection {
  label: string
  items: NavItem[]
}

const route = useRoute()
const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const isDark = ref(document.documentElement.classList.contains('dark'))

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const isAdmin = computed(() => authStore.isAdmin)
const user = computed(() => authStore.user)
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => appStore.siteName || 'Smirel API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }),
)

const copy = computed(() =>
  isZh.value
    ? {
        adminConsole: '运营控制台',
        developerConsole: '开发者控制台',
        platformWorkspace: '平台工作区',
        platformWorkspaceHint: '运营、路由与上游资源',
        apiWorkspace: 'API 工作区',
        apiWorkspaceHint: '密钥、模型与用量',
        overview: '概览',
        build: '构建',
        observe: '观察',
        account: '账户',
        operations: '运营',
        routing: '路由与上游',
        commerce: '商业化',
        governance: '治理',
        dashboard: '工作台',
        apiKeys: 'API 密钥',
        models: '模型与价格',
        usage: '用量与日志',
        access: '可用模型',
        status: '服务状态',
        plans: '订阅与额度',
        purchase: '充值与套餐',
        orders: '我的订单',
        redeem: '兑换码',
        affiliate: '邀请返利',
        batchImage: '批量生图指南',
        profile: '账户设置',
        opsDashboard: '运行状态',
        users: '用户与权限',
        groups: '路由策略',
        channels: '模型路由',
        monitor: '路由健康',
        upstream: '上游账户',
        proxies: '网络出口',
        subscriptions: '套餐管理',
        revenue: '收入概览',
        orderManagement: '订单管理',
        planConfig: '套餐配置',
        announcements: '公告管理',
        redeemAdmin: '兑换码管理',
        promo: '优惠码管理',
        audit: '审计日志',
        usageAdmin: '平台用量',
        risk: '风控规则',
        settings: '平台设置',
        darkMode: '深色模式',
        lightMode: '浅色模式',
      }
    : {
        adminConsole: 'Operations Console',
        developerConsole: 'Developer Console',
        platformWorkspace: 'Platform workspace',
        platformWorkspaceHint: 'Operations, routing and upstreams',
        apiWorkspace: 'API workspace',
        apiWorkspaceHint: 'Keys, models and usage',
        overview: 'Overview',
        build: 'Build',
        observe: 'Observe',
        account: 'Account',
        operations: 'Operations',
        routing: 'Routing & upstream',
        commerce: 'Commerce',
        governance: 'Governance',
        dashboard: 'Workspace',
        apiKeys: 'API Keys',
        models: 'Models & Pricing',
        usage: 'Usage & Logs',
        access: 'Available Models',
        status: 'Service Status',
        plans: 'Plans & Quota',
        purchase: 'Billing & Plans',
        orders: 'My Orders',
        redeem: 'Redeem',
        affiliate: 'Referrals',
        batchImage: 'Batch Image Guide',
        profile: 'Account Settings',
        opsDashboard: 'Operations',
        users: 'Users & Access',
        groups: 'Routing Policies',
        channels: 'Model Routing',
        monitor: 'Route Health',
        upstream: 'Upstream Accounts',
        proxies: 'Egress Network',
        subscriptions: 'Plan Management',
        revenue: 'Revenue Overview',
        orderManagement: 'Orders',
        planConfig: 'Plan Configuration',
        announcements: 'Announcements',
        redeemAdmin: 'Redeem Codes',
        promo: 'Promo Codes',
        audit: 'Audit Logs',
        usageAdmin: 'Platform Usage',
        risk: 'Risk Control',
        settings: 'Platform Settings',
        darkMode: 'Dark mode',
        lightMode: 'Light mode',
      },
)

function enabledItems(items: NavItem[]): NavItem[] {
  return items.filter((item) => item.enabled !== false)
}

const userSections = computed<NavSection[]>(() => [
  {
    label: copy.value.overview,
    items: enabledItems([
      { path: '/dashboard', label: copy.value.dashboard, glyph: '01' },
      { path: '/keys', label: copy.value.apiKeys, glyph: 'K' },
      {
        path: '/model-plaza?embedded=1',
        label: copy.value.models,
        glyph: 'M',
        enabled: settings.value?.model_plaza_enabled !== false,
      },
    ]),
  },
  {
    label: copy.value.observe,
    items: enabledItems([
      { path: '/usage', label: copy.value.usage, glyph: 'U' },
      {
        path: '/available-channels',
        label: copy.value.access,
        glyph: 'A',
        enabled: settings.value?.available_channels_enabled !== false,
      },
      {
        path: '/monitor',
        label: copy.value.status,
        glyph: 'S',
        enabled: settings.value?.channel_monitor_enabled !== false,
      },
    ]),
  },
  {
    label: copy.value.build,
    items: enabledItems([
      { path: '/batch-image', label: copy.value.batchImage, glyph: 'B' },
      { path: '/subscriptions', label: copy.value.plans, glyph: 'P' },
      {
        path: '/purchase',
        label: copy.value.purchase,
        glyph: '$',
        enabled: settings.value?.payment_enabled !== false,
      },
    ]),
  },
  {
    label: copy.value.account,
    items: enabledItems([
      {
        path: '/orders',
        label: copy.value.orders,
        glyph: 'O',
        enabled: settings.value?.payment_enabled !== false,
      },
      { path: '/redeem', label: copy.value.redeem, glyph: 'R' },
      {
        path: '/affiliate',
        label: copy.value.affiliate,
        glyph: 'F',
        enabled: settings.value?.affiliate_enabled === true,
      },
      { path: '/profile', label: copy.value.profile, glyph: 'C' },
    ]),
  },
])

const adminSections = computed<NavSection[]>(() => [
  {
    label: copy.value.operations,
    items: [
      { path: '/admin/dashboard', label: copy.value.dashboard, glyph: '01' },
      { path: '/admin/ops', label: copy.value.opsDashboard, glyph: 'OP' },
      { path: '/admin/users', label: copy.value.users, glyph: 'U' },
      { path: '/admin/usage', label: copy.value.usageAdmin, glyph: 'Σ' },
    ],
  },
  {
    label: copy.value.routing,
    items: [
      { path: '/admin/groups', label: copy.value.groups, glyph: 'R' },
      { path: '/admin/channels/pricing', label: copy.value.channels, glyph: 'M' },
      { path: '/admin/channels/monitor', label: copy.value.monitor, glyph: 'H' },
      { path: '/admin/accounts', label: copy.value.upstream, glyph: 'A' },
      { path: '/admin/proxies', label: copy.value.proxies, glyph: 'N' },
    ],
  },
  {
    label: copy.value.commerce,
    items: enabledItems([
      { path: '/admin/subscriptions', label: copy.value.subscriptions, glyph: 'P' },
      {
        path: '/admin/orders/dashboard',
        label: copy.value.revenue,
        glyph: '$',
        enabled: settings.value?.payment_enabled !== false,
      },
      {
        path: '/admin/orders',
        label: copy.value.orderManagement,
        glyph: 'O',
        enabled: settings.value?.payment_enabled !== false,
      },
      {
        path: '/admin/orders/plans',
        label: copy.value.planConfig,
        glyph: 'PC',
        enabled: settings.value?.payment_enabled !== false,
      },
      { path: '/admin/redeem', label: copy.value.redeemAdmin, glyph: 'R' },
      { path: '/admin/promo-codes', label: copy.value.promo, glyph: '%' },
    ]),
  },
  {
    label: copy.value.governance,
    items: enabledItems([
      { path: '/admin/announcements', label: copy.value.announcements, glyph: '!' },
      { path: '/admin/audit-logs', label: copy.value.audit, glyph: 'L' },
      {
        path: '/admin/risk-control',
        label: copy.value.risk,
        glyph: 'RC',
        enabled: settings.value?.risk_control_enabled === true,
      },
      { path: '/admin/settings', label: copy.value.settings, glyph: 'S' },
    ]),
  },
])

const navSections = computed(() => (isAdmin.value ? adminSections.value : userSections.value))
const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || 'User')
const initials = computed(() => displayName.value.slice(0, 2).toUpperCase())

function isActive(path: string): boolean {
  const cleanPath = path.split('?')[0]
  if (route.path === cleanPath) return true
  if (cleanPath === '/admin/channels/pricing' && route.path.startsWith('/admin/channels')) return true
  if (cleanPath === '/admin/orders/dashboard' && route.path === '/admin/orders/dashboard') return true
  return false
}

function closeMobile() {
  appStore.setMobileOpen(false)
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
