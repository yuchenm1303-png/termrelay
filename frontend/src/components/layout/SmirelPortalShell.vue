<template>
  <div class="smp-portal">
    <div class="smp-shell">
      <header class="smh-topbar smp-topbar">
        <router-link to="/home" class="smh-brand">
          <span class="smh-brand-mark">
            <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
            <span v-else>S</span>
          </span>
          <span class="smh-brand-copy">
            <strong>{{ siteName }}</strong>
            <small>AI API Gateway</small>
          </span>
        </router-link>

        <div class="smh-topbar-right">
          <nav class="smh-portal-nav" aria-label="Smirel portal navigation">
            <router-link to="/home">{{ copy.home }}</router-link>
            <router-link to="/model-plaza">{{ copy.models }}</router-link>
            <router-link to="/key-usage">{{ copy.keyUsage }}</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
            <router-link :to="isAuthenticated ? dashboardPath : '/login'">
              {{ isAuthenticated ? copy.console : copy.login }}
            </router-link>
          </nav>
          <div class="smh-locale"><LocaleSwitcher /></div>
          <span class="smh-service-status"><i></i>{{ copy.online }}</span>
        </div>
      </header>

      <main class="smp-body">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel API')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))

const copy = computed(() => isZh.value
  ? { home: '首页', models: '模型与价格', keyUsage: 'Key 用量', docs: 'API 文档', console: '进入控制台', login: '登录', online: '服务在线' }
  : { home: 'Home', models: 'Models & Pricing', keyUsage: 'Key Usage', docs: 'API Docs', console: 'Open Console', login: 'Sign In', online: 'Online' })
</script>
