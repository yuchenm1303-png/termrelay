<template>
  <div class="spx-auth-page spg-page">
    <div class="spg-environment" aria-hidden="true"></div>

    <div class="spx-auth-shell">
      <header class="spx-auth-topbar">
        <router-link to="/home" class="spg-brand spx-auth-brand" aria-label="Smirel 首页">
          <img v-if="siteLogo" :src="siteLogo" alt="" class="spg-brand-logo" />
          <span v-else class="spg-brand-fallback">S</span>
          <span class="spg-brand-copy">
            <strong>{{ siteName }}</strong>
            <small>ACCOUNT ACCESS</small>
          </span>
        </router-link>

        <div class="spx-auth-links">
          <router-link to="/home">{{ copy.home }}</router-link>
          <router-link to="/model-plaza">{{ copy.models }}</router-link>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
          <LocaleSwitcher />
        </div>
      </header>

      <main class="spx-auth-main">
        <section class="spx-auth-story">
          <p class="spg-overline">SMIREL WORKSPACE</p>
          <h1>{{ copy.title }}</h1>
          <p class="spx-auth-description">{{ copy.description }}</p>

          <div class="spx-auth-facts" aria-label="账户能力">
            <article v-for="fact in facts" :key="fact.title" class="spx-auth-fact">
              <strong>{{ fact.title }}</strong>
              <span>{{ fact.description }}</span>
            </article>
          </div>

          <div class="spg-surface spx-auth-endpoint">
            <div>
              <span>BASE URL</span>
              <code>{{ apiBase }}</code>
            </div>
            <span class="spg-pill">HTTPS</span>
          </div>
        </section>

        <section class="spg-surface spx-auth-panel">
          <div class="spx-auth-panel-head">
            <div>
              <p class="spg-overline">ACCOUNT</p>
              <strong>{{ copy.panelTitle }}</strong>
            </div>
            <span class="spg-pill">SECURE</span>
          </div>

          <div class="spx-auth-form">
            <slot />
          </div>

          <div class="spx-auth-panel-footer">
            <div class="spx-auth-footer-slot"><slot name="footer" /></div>
            <span>© {{ currentYear }} {{ siteName }}</span>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-shared-glass-v1.css'
import '@/styles/smirel-shared-interactions-v1.css'
import '@/styles/smirel-secondary-auth-v1.css'

const appStore = useAppStore()
const { locale } = useI18n()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const apiBase = computed(() => (
  appStore.cachedPublicSettings?.api_base_url
  || appStore.apiBaseUrl
  || 'https://api.smirel.com/v1'
).trim().replace(/\/$/, ''))

const copy = computed(() => isZh.value ? {
  home: '首页',
  models: '模型与价格',
  docs: '接入文档',
  title: '管理密钥、额度与调用记录。',
  description: '登录后即可创建 API Key、查看模型价格、余额和每笔调用明细。所有账户操作都在你的 Smirel 工作区内完成。',
  panelTitle: '账户访问',
} : {
  home: 'Home',
  models: 'Models & Pricing',
  docs: 'Documentation',
  title: 'Manage keys, balance and usage.',
  description: 'Sign in to create API keys, review model pricing, balance and request-level usage. Your account controls stay inside your Smirel workspace.',
  panelTitle: 'Account access',
})

const facts = computed(() => isZh.value ? [
  { title: '独立密钥', description: '每个项目单独管理访问凭证。' },
  { title: '按量计费', description: '余额与消费明细清晰可查。' },
  { title: '调用记录', description: '请求、Token 与费用统一查看。' },
] : [
  { title: 'Separate keys', description: 'Manage credentials per project.' },
  { title: 'Usage billing', description: 'Review balance and spend clearly.' },
  { title: 'Request history', description: 'Inspect requests, tokens and cost.' },
])

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
