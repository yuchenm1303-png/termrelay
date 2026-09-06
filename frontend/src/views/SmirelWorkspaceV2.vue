<template>
  <div class="sw2-page" :class="{ 'sw2-page--admin': isAdminWorkspace }">
    <div class="sw2-environment" aria-hidden="true"></div>

    <div class="sw2-shell">
      <header class="sw2-bar">
        <router-link to="/home" class="sw2-brand" aria-label="返回 Smirel 首页">
          <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
          <span v-else class="sw2-brand-fallback">{{ siteName }}</span>
        </router-link>

        <nav class="sw2-nav" :aria-label="isAdminWorkspace ? '管理员工作区主导航' : '工作区主导航'">
          <template v-for="item in navItems" :key="item.label">
            <router-link
              v-if="item.to"
              :to="item.to"
              class="sw2-nav-item"
              :class="{ 'sw2-nav-item--active': item.active }"
              :aria-current="item.active ? 'page' : undefined"
            >
              {{ item.label }}
            </router-link>
            <span v-else class="sw2-nav-item" aria-disabled="true">{{ item.label }}</span>
          </template>
        </nav>

        <div class="sw2-account-area">
          <span v-if="isAdminWorkspace" class="sw2-mode">ADMIN</span>
          <div class="sw2-account" aria-label="当前账户">
            <span class="sw2-account-dot" aria-hidden="true"></span>
            <span class="sw2-account-copy">{{ accountLabel }}</span>
          </div>
        </div>
      </header>

      <main class="sw2-main">
        <section class="sw2-heading">
          <div class="sw2-heading-copy">
            <p class="sw2-overline">{{ isAdminWorkspace ? 'SMIREL ADMIN WORKSPACE' : 'SMIREL WORKSPACE' }}</p>
            <h1>Overview</h1>
            <p>{{ headingDescription }}</p>
          </div>
          <div class="sw2-heading-state"><i aria-hidden="true"></i><span>{{ isAdminWorkspace ? 'Admin workspace active' : 'Workspace active' }}</span></div>
        </section>

        <section class="sw2-stage" :aria-label="isAdminWorkspace ? '管理员工作区内容区域' : '工作区内容区域'">
          <header class="sw2-stage-head">
            <div class="sw2-stage-label">
              <span>{{ isAdminWorkspace ? 'ADMIN CONTENT AREA' : 'CONTENT AREA' }}</span>
              <strong>{{ isAdminWorkspace ? 'Operations canvas' : 'Workspace canvas' }}</strong>
            </div>
            <span class="sw2-stage-mark">V2 FOUNDATION</span>
          </header>

          <div class="sw2-stage-body">
            <div class="sw2-stage-note">
              <span>{{ isAdminWorkspace ? 'ADMIN WORKSPACE V2' : 'SECONDARY UI V2' }}</span>
              <strong>{{ stageTitle }}</strong>
              <p>{{ stageDescription }}</p>
            </div>
          </div>

          <footer class="sw2-stage-foot">
            <span>Single root surface · no nested glass</span>
            <span>{{ isAdminWorkspace ? 'Legacy admin modules intentionally not mounted' : 'Business modules intentionally not mounted' }}</span>
          </footer>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-secondary-v2.css'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const accountLabel = computed(() => authStore.user?.email || (authStore.isAdmin ? 'Administrator' : 'Account'))
const isAdminWorkspace = computed(() => authStore.isAdmin)

const navItems = computed(() => {
  if (isAdminWorkspace.value) {
    return [
      { label: 'Overview', to: '/admin/dashboard', active: route.path === '/admin/dashboard' || route.path === '/dashboard' },
      { label: 'Users' },
      { label: 'Accounts' },
      { label: 'Channels' },
      { label: 'Billing' },
      { label: 'Operations' },
      { label: 'Settings' },
    ]
  }

  return [
    { label: 'Overview', to: '/dashboard', active: route.path === '/dashboard' },
    { label: 'API Keys' },
    { label: 'Models' },
    { label: 'Usage' },
    { label: 'Billing' },
  ]
})

const headingDescription = computed(() => isAdminWorkspace.value
  ? '这里会集中呈现平台运行状态、用户与上游资源、调度健康度和近期异常。当前先确认管理员工作区的信息架构与空间，不接回旧 Admin Dashboard。'
  : '这里将集中呈现账户状态、调用情况和近期活动。当前版本只确认二级工作区的导航、空间和材质，不接入旧 Dashboard 结构。')

const stageTitle = computed(() => isAdminWorkspace.value
  ? '先把管理员工作区的骨架做对，再放运营数据。'
  : '先把工作区本身做对，再放业务内容。')

const stageDescription = computed(() => isAdminWorkspace.value
  ? '下一步会在这个母版里重新设计 Admin Overview。Users、Accounts、Channels、Billing、Operations 和 Settings 都会从零组织，不挂载旧后台页面。'
  : '下一步会在这个母版里重新设计 Overview，而不是把旧 Dashboard、Sidebar 或卡片重新接回来。')
</script>
