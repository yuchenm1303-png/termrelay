<template>
  <div class="sw2-page" :class="{ 'sw2-page--admin': isAdminWorkspace }">
    <div class="sw2-environment" aria-hidden="true"></div>

    <div class="sw2-console">
      <aside class="sw2-sidebar spg-surface" :class="{ 'sw2-sidebar--open': mobileNavOpen }">
        <div class="sw2-sidebar-brand-row">
          <router-link to="/home" class="sw2-brand" aria-label="返回 Smirel 首页" @click="closeMobileNav">
            <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
            <span v-else class="sw2-brand-fallback">{{ siteName }}</span>
          </router-link>
          <button class="sw2-sidebar-close" type="button" aria-label="关闭导航" @click="closeMobileNav">×</button>
        </div>

        <div class="sw2-sidebar-context">
          <span>{{ isAdminWorkspace ? 'ADMIN CONSOLE' : 'WORKSPACE' }}</span>
          <strong>{{ isAdminWorkspace ? '平台管理' : 'Smirel API' }}</strong>
          <small><i></i>{{ isAdminWorkspace ? 'Operations workspace' : 'Developer workspace' }}</small>
        </div>

        <nav class="sw2-side-nav" :aria-label="isAdminWorkspace ? '管理员导航' : '工作区导航'">
          <section v-for="section in navSections" :key="section.label" class="sw2-nav-section">
            <p>{{ section.label }}</p>
            <template v-for="item in section.items" :key="item.label">
              <router-link
                v-if="item.available && item.to"
                :to="item.to"
                class="sw2-side-item"
                :class="{ 'sw2-side-item--active': isActive(item.to) }"
                :aria-current="isActive(item.to) ? 'page' : undefined"
                @click="closeMobileNav"
              >
                <span class="sw2-side-icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" /></svg>
                </span>
                <span>{{ item.label }}</span>
              </router-link>
              <span v-else class="sw2-side-item sw2-side-item--pending" aria-disabled="true">
                <span class="sw2-side-icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" /></svg>
                </span>
                <span>{{ item.label }}</span>
              </span>
            </template>
          </section>
        </nav>

        <div class="sw2-sidebar-foot">
          <div class="sw2-account-card">
            <span class="sw2-avatar">{{ initials }}</span>
            <span class="sw2-account-copy">
              <strong>{{ isAdminWorkspace ? 'Administrator' : 'Account' }}</strong>
              <small>{{ accountLabel }}</small>
            </span>
          </div>
          <router-link to="/home" class="sw2-home-link" @click="closeMobileNav">返回首页 <span>↗</span></router-link>
        </div>
      </aside>

      <button v-if="mobileNavOpen" class="sw2-nav-scrim" type="button" aria-label="关闭导航" @click="closeMobileNav"></button>

      <section class="sw2-workspace">
        <header class="sw2-topbar">
          <div class="sw2-topbar-left">
            <button class="sw2-menu-button" type="button" aria-label="打开导航" @click="mobileNavOpen = true">
              <span></span><span></span><span></span>
            </button>
            <div class="sw2-breadcrumb">
              <span>Smirel</span><b>/</b><strong>{{ isAdminWorkspace ? 'Admin' : 'Workspace' }}</strong>
            </div>
          </div>

          <div class="sw2-topbar-right">
            <span class="sw2-endpoint">{{ apiBaseCompact }}</span>
            <span class="sw2-service-state"><i></i>服务在线</span>
            <span v-if="isAdminWorkspace" class="sw2-role-pill">ADMIN</span>
            <span class="sw2-topbar-avatar">{{ initials }}</span>
          </div>
        </header>

        <main class="sw2-canvas">
          <div class="sw2-canvas-inner">
            <header class="sw2-page-head">
              <div>
                <p>{{ pageMeta.kicker }}</p>
                <h1>{{ pageMeta.title }}</h1>
                <span>{{ pageMeta.description }}</span>
              </div>
              <div class="sw2-page-state"><i></i>{{ isAdminWorkspace ? 'Admin workspace active' : 'Workspace active' }}</div>
            </header>

            <SmirelAdminOverviewV2 v-if="isAdminWorkspace && isOverview" />

            <section v-else class="sw2-panel sw2-placeholder">
              <span>{{ pageMeta.kicker }}</span>
              <strong>{{ pageMeta.title }} 正在重新设计</strong>
              <p>这个页面不会接回旧布局，会在当前 Workspace Canvas 中重新实现。</p>
            </section>
          </div>
        </main>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import SmirelAdminOverviewV2 from '@/views/admin/SmirelAdminOverviewV2.vue'
import '@/styles/smirel-secondary-v2.css'
import '@/styles/smirel-shared-glass-v1.css'
import '@/styles/smirel-sidebar-fixed-v2.css'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const mobileNavOpen = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const accountLabel = computed(() => authStore.user?.email || (authStore.isAdmin ? 'Administrator' : 'Account'))
const initials = computed(() => (accountLabel.value || 'S').trim().slice(0, 1).toUpperCase())
const isAdminWorkspace = computed(() => authStore.isAdmin || route.path.startsWith('/admin'))
const isOverview = computed(() => route.path === '/admin/dashboard' || route.path === '/dashboard')
const apiBaseCompact = computed(() => (
  appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1'
).replace(/^https?:\/\//, '').replace(/\/$/, ''))

const icons = {
  overview: 'M3.75 3.75h6.5v6.5h-6.5v-6.5zm10 0h6.5v6.5h-6.5v-6.5zm-10 10h6.5v6.5h-6.5v-6.5zm10 0h6.5v6.5h-6.5v-6.5z',
  users: 'M16.5 18.75a6 6 0 00-9 0M12 13.5a3.75 3.75 0 100-7.5 3.75 3.75 0 000 7.5zm7.5 5.25a4.5 4.5 0 00-3.2-4.31m.95-7.29a3 3 0 010 5.7',
  usage: 'M4 19V10m5.3 9V5m5.4 14v-7m5.3 7V8',
  accounts: 'M5 7.5h14M5 12h14M5 16.5h9M3.5 4.5h17v15h-17v-15z',
  routing: 'M5 5h5v5H5V5zm9 9h5v5h-5v-5zM10 7.5h3a3 3 0 013 3V14m-8 0v-1.5A2.5 2.5 0 0110.5 10H14',
  channels: 'M4 6.5h16M4 12h16M4 17.5h16M7 4v5m8 1v5m-5 0v5',
  health: 'M3 12h4l2-5 4 10 2.5-6H21',
  billing: 'M4 6.5h16v11H4v-11zm0 3.5h16M7 14h4',
  operations: 'M4 17l4-5 3 2 5-7 4 3M4 20h16',
  audit: 'M6 3.75h9l3 3V20.25H6V3.75zm8.5 0v3.5H18M9 11h6m-6 3h6m-6 3h4',
  settings: 'M12 8.75A3.25 3.25 0 1112 15.25 3.25 3.25 0 0112 8.75zm0-5.25v2m0 12.5v2m8.5-8.5h-2m-13 0h-2m14.51-6.01l-1.42 1.42M7.41 16.59l-1.42 1.42m12.02 0l-1.42-1.42M7.41 7.41L5.99 5.99',
  key: 'M15.5 7.5a4.5 4.5 0 11-8.2 2.6L3.5 14v3h3v3h3l4.2-4.2M15.5 7.5h.01',
  models: 'M12 3l8 4.5-8 4.5-8-4.5L12 3zm-8 9l8 4.5 8-4.5m-16 4.5L12 21l8-4.5',
  profile: 'M12 12a4 4 0 100-8 4 4 0 000 8zm-7 8a7 7 0 0114 0',
}

type NavItem = { label: string; to?: string; icon: string; available?: boolean }
type NavSection = { label: string; items: NavItem[] }

const navSections = computed<NavSection[]>(() => isAdminWorkspace.value ? [
  {
    label: '控制台',
    items: [
      { label: '概览', to: '/admin/dashboard', icon: icons.overview, available: true },
      { label: '用户与权限', icon: icons.users },
      { label: '平台用量', icon: icons.usage },
    ],
  },
  {
    label: '路由与上游',
    items: [
      { label: '上游账户', icon: icons.accounts },
      { label: '路由策略', icon: icons.routing },
      { label: '模型路由', icon: icons.channels },
      { label: '路由健康', icon: icons.health },
    ],
  },
  {
    label: '商业',
    items: [
      { label: '套餐与订单', icon: icons.billing },
    ],
  },
  {
    label: '系统',
    items: [
      { label: '运行状态', icon: icons.operations },
      { label: '审计日志', icon: icons.audit },
      { label: '平台设置', icon: icons.settings },
    ],
  },
] : [
  {
    label: '工作区',
    items: [
      { label: '概览', to: '/dashboard', icon: icons.overview, available: true },
      { label: 'API Keys', icon: icons.key },
      { label: '模型与价格', icon: icons.models },
      { label: '用量与日志', icon: icons.usage },
    ],
  },
  {
    label: '账户',
    items: [
      { label: '账单与额度', icon: icons.billing },
      { label: '账户设置', icon: icons.profile },
    ],
  },
])

const pageMeta = computed(() => isAdminWorkspace.value ? {
  kicker: 'ADMIN CONSOLE',
  title: 'Overview',
  description: '平台健康、上游资源、流量与成本，一屏完成日常运营判断。',
} : {
  kicker: 'WORKSPACE',
  title: 'Overview',
  description: '账户、密钥、用量和模型访问集中在一个工作区。',
})

function isActive(path: string): boolean {
  return path === '/admin/dashboard'
    ? route.path === '/admin/dashboard' || route.path === '/dashboard'
    : route.path === path
}

function closeMobileNav() {
  mobileNavOpen.value = false
}

watch(() => route.fullPath, closeMobileNav)
</script>
