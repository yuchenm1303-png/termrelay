<template>
  <div class="sw2-page">
    <div class="sw2-environment" aria-hidden="true"></div>

    <div class="sw2-shell">
      <header class="sw2-bar">
        <router-link to="/home" class="sw2-brand" aria-label="返回 Smirel 首页">
          <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
          <span v-else class="sw2-brand-fallback">{{ siteName }}</span>
        </router-link>

        <nav class="sw2-nav" aria-label="工作区主导航">
          <router-link to="/dashboard" class="sw2-nav-item sw2-nav-item--active" aria-current="page">
            Overview
          </router-link>
          <span class="sw2-nav-item" aria-disabled="true">API Keys</span>
          <span class="sw2-nav-item" aria-disabled="true">Models</span>
          <span class="sw2-nav-item" aria-disabled="true">Usage</span>
          <span class="sw2-nav-item" aria-disabled="true">Billing</span>
        </nav>

        <div class="sw2-account-area">
          <span v-if="authStore.isAdmin" class="sw2-mode">ADMIN</span>
          <div class="sw2-account" aria-label="当前账户">
            <span class="sw2-account-dot" aria-hidden="true"></span>
            <span class="sw2-account-copy">{{ accountLabel }}</span>
          </div>
        </div>
      </header>

      <main class="sw2-main">
        <section class="sw2-heading">
          <div class="sw2-heading-copy">
            <p class="sw2-overline">SMIREL WORKSPACE</p>
            <h1>Overview</h1>
            <p>这里将集中呈现账户状态、调用情况和近期活动。当前版本只确认二级工作区的导航、空间和材质，不接入旧 Dashboard 结构。</p>
          </div>
          <div class="sw2-heading-state"><i aria-hidden="true"></i><span>Workspace active</span></div>
        </section>

        <section class="sw2-stage" aria-label="工作区内容区域">
          <header class="sw2-stage-head">
            <div class="sw2-stage-label">
              <span>CONTENT AREA</span>
              <strong>Workspace canvas</strong>
            </div>
            <span class="sw2-stage-mark">V2 FOUNDATION</span>
          </header>

          <div class="sw2-stage-body">
            <div class="sw2-stage-note">
              <span>SECONDARY UI V2</span>
              <strong>先把工作区本身做对，再放业务内容。</strong>
              <p>下一步会在这个母版里重新设计 Overview，而不是把旧 Dashboard、Sidebar 或卡片重新接回来。</p>
            </div>
          </div>

          <footer class="sw2-stage-foot">
            <span>Single root surface · no nested glass</span>
            <span>Business modules intentionally not mounted</span>
          </footer>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-secondary-v2.css'

const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const accountLabel = computed(() => authStore.user?.email || (authStore.isAdmin ? 'Administrator' : 'Account'))
</script>
