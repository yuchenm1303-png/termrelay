<template>
  <aside class="smg-sidebar" :class="{ 'smg-sidebar--open': mobileOpen }">
    <div class="smg-sidebar-head">
      <router-link :to="homePath" class="smg-brand" @click="closeMobile">
        <span class="smg-brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
          <span v-else>S</span>
        </span>
        <span class="smg-brand-copy">
          <strong>{{ siteName }}</strong>
          <small>{{ isAdmin ? copy.operationsConsole : copy.console }}</small>
        </span>
      </router-link>
      <span class="smg-live-pill"><i></i>{{ copy.live }}</span>
    </div>

    <nav class="smg-nav" :aria-label="isAdmin ? copy.operationsConsole : copy.console">
      <section v-for="section in navSections" :key="section.label" class="smg-nav-section">
        <p>{{ section.label }}</p>
        <router-link
          v-for="item in section.items"
          :key="item.path"
          :to="item.path"
          class="smg-nav-item"
          :class="{ 'smg-nav-item--active': isActive(item.path) }"
          @click="closeMobile"
        >
          <span class="smg-nav-copy">
            <strong>{{ item.label }}</strong>
            <small>{{ item.hint }}</small>
          </span>
        </router-link>
      </section>
    </nav>

    <div class="smg-sidebar-foot">
      <router-link to="/profile" class="smg-account" @click="closeMobile">
        <span class="smg-avatar">{{ initials }}</span>
        <span>
          <strong>{{ displayName }}</strong>
          <small>{{ isAdmin ? copy.administrator : (user?.email || copy.account) }}</small>
        </span>
      </router-link>
      <button type="button" class="smg-theme" @click="toggleTheme">
        {{ isDark ? copy.lightMode : copy.darkMode }}
      </button>
    </div>
  </aside>

  <button
    v-if="mobileOpen"
    class="smg-mobile-scrim"
    type="button"
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

const route = useRoute()
const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const isDark = ref(document.documentElement.classList.contains('dark'))

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const isAdmin = computed(() => authStore.isAdmin)
const user = computed(() => authStore.user)
const settings = computed(() => appStore.cachedPublicSettings)
const mobileOpen = computed(() => appStore.mobileOpen)
const siteName = computed(() => appStore.siteName || 'Smirel API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const homePath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

const copy = computed(() => isZh.value ? {
  console: 'API 控制台', operationsConsole: '运营控制台', live: 'ONLINE',
  core: '核心', observe: '用量', resources: '资源', billing: '结算', account: '账户',
  operations: '运营', routing: '路由与上游', commerce: '商业化', governance: '治理',
  dashboard: '工作台', dashboardHint: '余额与运行概览', keys: 'API Keys', keysHint: '创建与管理凭证',
  models: '模型与价格', modelsHint: '能力、价格与选择', usage: '用量与日志', usageHint: '请求、Token 与费用',
  available: '可用模型', availableHint: '当前可调用能力', monitor: '服务状态', monitorHint: '模型与通道可用性',
  batchImage: '批量生图指南', batchImageHint: '批量图像调用说明', plans: '套餐与额度', plansHint: '订阅和可用额度',
  purchase: '充值与账单', purchaseHint: '购买套餐与充值', orders: '我的订单', ordersHint: '查看购买与支付记录',
  redeem: '兑换码', redeemHint: '兑换余额或套餐', affiliate: '邀请返利', affiliateHint: '邀请与返利记录',
  profile: '账户设置', profileHint: '资料与安全',
  adminDashboard: '平台概览', adminDashboardHint: '平台运行与关键指标', opsDashboard: '运行状态', opsDashboardHint: '系统健康与运行状态',
  users: '用户与权限', usersHint: '客户账户与权限管理', usageAdmin: '平台用量', usageAdminHint: '全平台请求与费用',
  groups: '路由策略', groupsHint: '分组与调度规则', channels: '模型路由', channelsHint: '模型价格与路由配置',
  routeHealth: '路由健康', routeHealthHint: '通道健康与可调度状态', upstream: '上游账户', upstreamHint: '上游凭证与资源池',
  proxies: '网络出口', proxiesHint: '代理与出口网络', subscriptionsAdmin: '套餐管理', subscriptionsAdminHint: '用户订阅与额度',
  revenue: '收入概览', revenueHint: '订单与收入趋势', orderManagement: '订单管理', orderManagementHint: '支付与订单处理',
  planConfig: '套餐配置', planConfigHint: '套餐、价格与规则', redeemAdmin: '兑换码管理', redeemAdminHint: '生成与管理兑换码',
  promo: '优惠码管理', promoHint: '促销与优惠规则', announcements: '公告管理', announcementsHint: '客户公告内容',
  audit: '审计日志', auditHint: '关键管理操作记录', risk: '风控规则', riskHint: '请求与账户风险策略',
  settingsAdmin: '平台设置', settingsAdminHint: '全局平台配置', administrator: 'Administrator',
  darkMode: '深色', lightMode: '浅色',
} : {
  console: 'API Console', operationsConsole: 'Operations Console', live: 'ONLINE',
  core: 'Core', observe: 'Usage', resources: 'Resources', billing: 'Billing', account: 'Account',
  operations: 'Operations', routing: 'Routing & Upstream', commerce: 'Commerce', governance: 'Governance',
  dashboard: 'Workspace', dashboardHint: 'Balance and runtime overview', keys: 'API Keys', keysHint: 'Create and manage credentials',
  models: 'Models & Pricing', modelsHint: 'Capabilities, pricing and selection', usage: 'Usage & Logs', usageHint: 'Requests, tokens and spend',
  available: 'Available Models', availableHint: 'Models you can call now', monitor: 'Service Status', monitorHint: 'Model and route availability',
  batchImage: 'Batch Image Guide', batchImageHint: 'Batch image API guidance', plans: 'Plans & Quota', plansHint: 'Subscriptions and quota',
  purchase: 'Billing', purchaseHint: 'Top up and purchase plans', orders: 'My Orders', ordersHint: 'Purchases and payment history',
  redeem: 'Redeem', redeemHint: 'Redeem balance or plans', affiliate: 'Referrals', affiliateHint: 'Invites and referral rewards',
  profile: 'Account Settings', profileHint: 'Profile and security',
  adminDashboard: 'Platform Overview', adminDashboardHint: 'Platform health and key metrics', opsDashboard: 'Operations', opsDashboardHint: 'System health and runtime state',
  users: 'Users & Access', usersHint: 'Customer accounts and permissions', usageAdmin: 'Platform Usage', usageAdminHint: 'Platform-wide requests and spend',
  groups: 'Routing Policies', groupsHint: 'Groups and scheduling rules', channels: 'Model Routing', channelsHint: 'Model pricing and route configuration',
  routeHealth: 'Route Health', routeHealthHint: 'Channel health and schedulability', upstream: 'Upstream Accounts', upstreamHint: 'Upstream credentials and pools',
  proxies: 'Egress Network', proxiesHint: 'Proxy and egress configuration', subscriptionsAdmin: 'Plan Management', subscriptionsAdminHint: 'User subscriptions and quota',
  revenue: 'Revenue Overview', revenueHint: 'Orders and revenue trends', orderManagement: 'Orders', orderManagementHint: 'Payments and order processing',
  planConfig: 'Plan Configuration', planConfigHint: 'Plans, prices and rules', redeemAdmin: 'Redeem Codes', redeemAdminHint: 'Create and manage redeem codes',
  promo: 'Promo Codes', promoHint: 'Promotions and discount rules', announcements: 'Announcements', announcementsHint: 'Customer-facing notices',
  audit: 'Audit Logs', auditHint: 'Administrative activity records', risk: 'Risk Control', riskHint: 'Request and account risk policies',
  settingsAdmin: 'Platform Settings', settingsAdminHint: 'Global platform configuration', administrator: 'Administrator',
  darkMode: 'Dark', lightMode: 'Light',
})

function enabledItems<T extends { enabled?: boolean }>(items: T[]): T[] {
  return items.filter((item) => item.enabled !== false)
}

const userSections = computed(() => [
  {
    label: copy.value.core,
    items: enabledItems([
      { path: '/dashboard', label: copy.value.dashboard, hint: copy.value.dashboardHint },
      { path: '/keys', label: copy.value.keys, hint: copy.value.keysHint },
      { path: '/model-plaza?embedded=1', label: copy.value.models, hint: copy.value.modelsHint, enabled: settings.value?.model_plaza_enabled !== false },
    ]),
  },
  {
    label: copy.value.observe,
    items: enabledItems([
      { path: '/usage', label: copy.value.usage, hint: copy.value.usageHint },
      { path: '/available-channels', label: copy.value.available, hint: copy.value.availableHint, enabled: settings.value?.available_channels_enabled !== false },
      { path: '/monitor', label: copy.value.monitor, hint: copy.value.monitorHint, enabled: settings.value?.channel_monitor_enabled !== false },
    ]),
  },
  {
    label: copy.value.resources,
    items: [
      { path: '/batch-image', label: copy.value.batchImage, hint: copy.value.batchImageHint },
    ],
  },
  {
    label: copy.value.billing,
    items: enabledItems([
      { path: '/subscriptions', label: copy.value.plans, hint: copy.value.plansHint },
      { path: '/purchase', label: copy.value.purchase, hint: copy.value.purchaseHint, enabled: settings.value?.payment_enabled !== false },
    ]),
  },
  {
    label: copy.value.account,
    items: enabledItems([
      { path: '/orders', label: copy.value.orders, hint: copy.value.ordersHint, enabled: settings.value?.payment_enabled !== false },
      { path: '/redeem', label: copy.value.redeem, hint: copy.value.redeemHint },
      { path: '/affiliate', label: copy.value.affiliate, hint: copy.value.affiliateHint, enabled: settings.value?.affiliate_enabled === true },
      { path: '/profile', label: copy.value.profile, hint: copy.value.profileHint },
    ]),
  },
])

const adminSections = computed(() => [
  {
    label: copy.value.operations,
    items: [
      { path: '/admin/dashboard', label: copy.value.adminDashboard, hint: copy.value.adminDashboardHint },
      { path: '/admin/ops', label: copy.value.opsDashboard, hint: copy.value.opsDashboardHint },
      { path: '/admin/users', label: copy.value.users, hint: copy.value.usersHint },
      { path: '/admin/usage', label: copy.value.usageAdmin, hint: copy.value.usageAdminHint },
    ],
  },
  {
    label: copy.value.routing,
    items: [
      { path: '/admin/groups', label: copy.value.groups, hint: copy.value.groupsHint },
      { path: '/admin/channels/pricing', label: copy.value.channels, hint: copy.value.channelsHint },
      { path: '/admin/channels/monitor', label: copy.value.routeHealth, hint: copy.value.routeHealthHint },
      { path: '/admin/accounts', label: copy.value.upstream, hint: copy.value.upstreamHint },
      { path: '/admin/proxies', label: copy.value.proxies, hint: copy.value.proxiesHint },
    ],
  },
  {
    label: copy.value.commerce,
    items: enabledItems([
      { path: '/admin/subscriptions', label: copy.value.subscriptionsAdmin, hint: copy.value.subscriptionsAdminHint },
      { path: '/admin/orders/dashboard', label: copy.value.revenue, hint: copy.value.revenueHint, enabled: settings.value?.payment_enabled !== false },
      { path: '/admin/orders', label: copy.value.orderManagement, hint: copy.value.orderManagementHint, enabled: settings.value?.payment_enabled !== false },
      { path: '/admin/orders/plans', label: copy.value.planConfig, hint: copy.value.planConfigHint, enabled: settings.value?.payment_enabled !== false },
      { path: '/admin/redeem', label: copy.value.redeemAdmin, hint: copy.value.redeemAdminHint },
      { path: '/admin/promo-codes', label: copy.value.promo, hint: copy.value.promoHint },
    ]),
  },
  {
    label: copy.value.governance,
    items: enabledItems([
      { path: '/admin/announcements', label: copy.value.announcements, hint: copy.value.announcementsHint },
      { path: '/admin/audit-logs', label: copy.value.audit, hint: copy.value.auditHint },
      { path: '/admin/risk-control', label: copy.value.risk, hint: copy.value.riskHint, enabled: settings.value?.risk_control_enabled === true },
      { path: '/admin/settings', label: copy.value.settingsAdmin, hint: copy.value.settingsAdminHint },
    ]),
  },
])

const navSections = computed(() => isAdmin.value
  ? [...userSections.value, ...adminSections.value]
  : userSections.value)

const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || 'User')
const initials = computed(() => displayName.value.slice(0, 2).toUpperCase())

function isActive(path: string): boolean {
  const cleanPath = path.split('?')[0]
  if (route.path === cleanPath) return true

  const nestedMatches = navSections.value
    .flatMap((section) => section.items)
    .map((item) => item.path.split('?')[0])
    .filter((candidate) => route.path.startsWith(`${candidate}/`))
    .sort((a, b) => b.length - a.length)

  return nestedMatches[0] === cleanPath
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