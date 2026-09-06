<template>
  <div class="spg-page sw2-page" :class="{ 'sw2-page--admin': isAdminWorkspace }">
    <div class="spg-environment sw2-environment" aria-hidden="true"></div>

    <div class="sw2-console">
      <aside class="sw2-sidebar spg-surface smirel-card" :class="{ 'sw2-sidebar--open': mobileNavOpen }">
        <div class="sw2-sidebar-brand-row">
          <router-link to="/home" class="sw2-brand" aria-label="返回 Smirel 首页" @click="closeMobileNav">
            <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
            <span v-else class="sw2-brand-fallback">{{ siteName }}</span>
          </router-link>
          <button class="sw2-sidebar-close" type="button" aria-label="关闭导航" @click="closeMobileNav">×</button>
        </div>

        <div class="sw2-sidebar-context smirel-card smirel-card--quiet">
          <div class="sw2-sidebar-context-head">
            <span>{{ isAdminWorkspace ? 'ADMIN CONSOLE' : 'WORKSPACE' }}</span>
            <span class="sw2-context-role">{{ isAdminWorkspace ? 'ADMIN' : 'USER' }}</span>
          </div>
          <strong>{{ isAdminWorkspace ? '平台管理' : 'Smirel API' }}</strong>
          <small>{{ navItemCount }} 个已接入功能</small>
        </div>

        <nav class="sw2-side-nav" :aria-label="isAdminWorkspace ? '管理员导航' : '工作区导航'">
          <section v-for="section in navSections" :key="section.label" class="sw2-nav-section">
            <p>{{ section.label }}</p>
            <router-link
              v-for="item in section.items"
              :key="item.to"
              :to="item.to"
              class="sw2-side-item smirel-nav-item"
              :class="{
                'sw2-side-item--active': isActive(item.to),
                'smirel-nav-item--active': isActive(item.to),
              }"
              :aria-current="isActive(item.to) ? 'page' : undefined"
              @click="closeMobileNav"
            >
              <span class="sw2-side-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" /></svg>
              </span>
              <span>{{ item.label }}</span>
            </router-link>
          </section>
        </nav>
      </aside>

      <button v-if="mobileNavOpen" class="sw2-nav-scrim" type="button" aria-label="关闭导航" @click="closeMobileNav"></button>

      <section class="sw2-workspace">
        <header class="sw2-topbar smirel-card">
          <div class="sw2-topbar-left">
            <button class="sw2-menu-button" type="button" aria-label="打开导航" @click="mobileNavOpen = true">
              <span></span><span></span><span></span>
            </button>
            <div class="sw2-breadcrumb">
              <span>Smirel</span><b>/</b><span>{{ isAdminWorkspace ? 'Admin' : 'Workspace' }}</span><b>/</b><strong>{{ currentNavLabel }}</strong>
            </div>
          </div>

          <div class="sw2-topbar-right">
            <span class="sw2-endpoint">{{ apiBaseCompact }}</span>
            <span v-if="isAdminWorkspace" class="sw2-role-pill">ADMIN</span>
            <router-link to="/profile" class="sw2-topbar-avatar" aria-label="账户设置">
              <img v-if="avatarUrl" :src="avatarUrl" alt="" />
              <span v-else>{{ initials }}</span>
            </router-link>
          </div>
        </header>

        <main class="sw2-canvas" :class="{ 'sw2-canvas--route': !showAdminOverview }">
          <div class="sw2-canvas-inner">
            <template v-if="showAdminOverview">
              <header class="sw2-page-head">
                <div>
                  <p>ADMIN CONSOLE</p>
                  <h1>Overview</h1>
                  <span>平台健康、上游资源、流量与成本，一屏完成日常运营判断。</span>
                </div>
              </header>
              <SmirelAdminOverviewV2 />
            </template>

            <div v-else class="sw2-route-stage">
              <slot />
            </div>
          </div>
        </main>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { isSmirelUiPreview } from '@/utils/smirelUiPreview'
import { buildWorkspaceNavigation } from '@/components/layout/smirelWorkspaceNavigation'
import SmirelAdminOverviewV2 from '@/views/admin/SmirelAdminOverviewV2.vue'
import '@/styles/smirel-secondary-v2.css'
import '@/styles/smirel-shared-glass-v1.css'
import '@/styles/smirel-sidebar-fixed-v2.css'
import '@/styles/smirel-workspace-functional-v2.css'
import '@/styles/smirel-card-system-v1.css'
import '@/styles/smirel-navigation-v1.css'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const mobileNavOpen = ref(false)

const settings = computed(() => appStore.cachedPublicSettings)
const user = computed(() => authStore.user)
const siteName = computed(() => settings.value?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() => sanitizeUrl(
  settings.value?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const avatarUrl = computed(() => sanitizeUrl(user.value?.avatar_url || '', { allowRelative: true, allowDataUrl: true }))
const displayName = computed(() => {
  if (isSmirelUiPreview) return '管理员预览'

  const username = user.value?.username?.trim()
  if (username) return username

  const email = user.value?.email?.trim()
  if (email) return email.split('@')[0]

  return authStore.isAdmin ? 'Administrator' : 'Account'
})
const initials = computed(() => (displayName.value || 'S').trim().slice(0, 1).toUpperCase())
const isAdminWorkspace = computed(() => authStore.isAdmin || route.path.startsWith('/admin'))
const showAdminOverview = computed(() => isAdminWorkspace.value && route.path === '/admin/dashboard')
const apiBaseCompact = computed(() => (
  settings.value?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1'
).replace(/^https?:\/\//, '').replace(/\/$/, ''))

const navSections = computed(() => buildWorkspaceNavigation(router, {
  isAdmin: isAdminWorkspace.value,
  isSimpleMode: authStore.isSimpleMode,
  paymentEnabled: settings.value?.payment_enabled !== false,
  riskControlEnabled: settings.value?.risk_control_enabled === true,
  modelPlazaEnabled: settings.value?.model_plaza_enabled !== false,
  availableChannelsEnabled: settings.value?.available_channels_enabled !== false,
  channelMonitorEnabled: settings.value?.channel_monitor_enabled !== false,
  affiliateEnabled: settings.value?.affiliate_enabled === true,
}))

const navItemCount = computed(() => navSections.value.reduce((total, section) => total + section.items.length, 0))
const currentNavLabel = computed(() => {
  const activeItem = navSections.value
    .flatMap((section) => section.items)
    .find((item) => item.to.split('?')[0] === route.path)

  if (activeItem) return activeItem.label
  return String(route.meta.title || (isAdminWorkspace.value ? 'Admin' : 'Workspace'))
})

function isActive(target: string): boolean {
  return target.split('?')[0] === route.path
}

function closeMobileNav() {
  mobileNavOpen.value = false
}

watch(() => route.fullPath, closeMobileNav)
</script>
