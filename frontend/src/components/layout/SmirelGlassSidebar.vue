<template>
  <aside class="smg-sidebar" :class="{ 'smg-sidebar--open': mobileOpen }">
    <div class="smg-sidebar-head">
      <router-link to="/dashboard" class="smg-brand" @click="closeMobile">
        <span class="smg-brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
          <span v-else>S</span>
        </span>
        <span class="smg-brand-copy">
          <strong>{{ siteName }}</strong>
          <small>{{ copy.console }}</small>
        </span>
      </router-link>
      <span class="smg-live-pill"><i></i>{{ copy.live }}</span>
    </div>

    <div class="smg-sidebar-context">
      <span>{{ copy.workspace }}</span>
      <strong>{{ copy.workspaceName }}</strong>
      <small>{{ compactEndpoint }}</small>
    </div>

    <nav class="smg-nav" aria-label="Developer console">
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
          <span class="smg-nav-index">{{ item.index }}</span>
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
          <small>{{ user?.email || copy.account }}</small>
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
const user = computed(() => authStore.user)
const settings = computed(() => appStore.cachedPublicSettings)
const mobileOpen = computed(() => appStore.mobileOpen)
const siteName = computed(() => appStore.siteName || 'Smirel API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const apiBase = computed(() =>
  (settings.value?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1').trim().replace(/\/$/, ''),
)
const compactEndpoint = computed(() => apiBase.value.replace(/^https?:\/\//, ''))

const copy = computed(() => isZh.value ? {
  console: 'API 控制台', live: 'ONLINE', workspace: 'WORKSPACE', workspaceName: 'Smirel API',
  core: '核心', observe: '用量', billing: '结算', account: '账户', dashboard: '工作台', dashboardHint: '余额与运行概览',
  keys: 'API Keys', keysHint: '创建与管理凭证', models: '模型与价格', modelsHint: '能力、价格与选择',
  usage: '用量与日志', usageHint: '请求、Token 与费用', available: '可用模型', availableHint: '当前可调用能力',
  plans: '套餐与额度', plansHint: '订阅和可用额度', purchase: '充值与账单', purchaseHint: '购买套餐与充值',
  profile: '账户设置', profileHint: '资料与安全', darkMode: '深色', lightMode: '浅色',
} : {
  console: 'API Console', live: 'ONLINE', workspace: 'WORKSPACE', workspaceName: 'Smirel API',
  core: 'Core', observe: 'Usage', billing: 'Billing', account: 'Account', dashboard: 'Workspace', dashboardHint: 'Balance and runtime overview',
  keys: 'API Keys', keysHint: 'Create and manage credentials', models: 'Models & Pricing', modelsHint: 'Capabilities, pricing and selection',
  usage: 'Usage & Logs', usageHint: 'Requests, tokens and spend', available: 'Available Models', availableHint: 'Models you can call now',
  plans: 'Plans & Quota', plansHint: 'Subscriptions and quota', purchase: 'Billing', purchaseHint: 'Top up and purchase plans',
  profile: 'Account Settings', profileHint: 'Profile and security', darkMode: 'Dark', lightMode: 'Light',
})

const navSections = computed(() => [
  {
    label: copy.value.core,
    items: [
      { path: '/dashboard', label: copy.value.dashboard, hint: copy.value.dashboardHint, index: '01' },
      { path: '/keys', label: copy.value.keys, hint: copy.value.keysHint, index: '02' },
      { path: '/model-plaza?embedded=1', label: copy.value.models, hint: copy.value.modelsHint, index: '03', enabled: settings.value?.model_plaza_enabled !== false },
    ].filter((item) => item.enabled !== false),
  },
  {
    label: copy.value.observe,
    items: [
      { path: '/usage', label: copy.value.usage, hint: copy.value.usageHint, index: '04' },
      { path: '/available-channels', label: copy.value.available, hint: copy.value.availableHint, index: '05', enabled: settings.value?.available_channels_enabled !== false },
    ].filter((item) => item.enabled !== false),
  },
  {
    label: copy.value.billing,
    items: [
      { path: '/subscriptions', label: copy.value.plans, hint: copy.value.plansHint, index: '06' },
      { path: '/purchase', label: copy.value.purchase, hint: copy.value.purchaseHint, index: '07', enabled: settings.value?.payment_enabled !== false },
    ].filter((item) => item.enabled !== false),
  },
  {
    label: copy.value.account,
    items: [
      { path: '/profile', label: copy.value.profile, hint: copy.value.profileHint, index: '08' },
    ],
  },
])

const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || 'User')
const initials = computed(() => displayName.value.slice(0, 2).toUpperCase())

function isActive(path: string): boolean {
  return route.path === path.split('?')[0]
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
