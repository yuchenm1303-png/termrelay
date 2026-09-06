<template>
  <div class="smg-auth">
    <div class="smg-environment" aria-hidden="true"></div>

    <GlassSurface as="header" class="smg-auth-nav">
      <router-link to="/home" class="smg-auth-brand">
        <span class="smg-brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
          <span v-else>S</span>
        </span>
        <span>{{ siteName }}</span>
      </router-link>
      <div class="smg-auth-links">
        <router-link to="/model-plaza">{{ copy.models }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
        <LocaleSwitcher />
      </div>
    </GlassSurface>

    <main class="smg-auth-main">
      <GlassSurface class="smg-auth-story">
        <div class="smg-auth-story-kicker">{{ copy.kicker }}</div>
        <h1>{{ copy.title }}</h1>
        <p>{{ copy.description }}</p>

        <div class="smg-auth-facts">
          <GlassSurface
            v-for="fact in facts"
            :key="fact.title"
            as="article"
            tone="data"
            class="smg-auth-fact"
          >
            <strong>{{ fact.title }}</strong>
            <span>{{ fact.description }}</span>
          </GlassSurface>
        </div>

        <GlassSurface tone="data" class="smg-auth-code">
          <span>BASE_URL</span>
          <code>{{ apiBase }}</code>
          <span>AUTH</span>
          <code>Authorization: Bearer sk-••••••••</code>
        </GlassSurface>
      </GlassSurface>

      <GlassSurface class="smg-auth-card">
        <div class="smg-auth-card-shell"><slot /></div>
        <div class="smg-auth-footer"><slot name="footer" /></div>
        <div class="smg-auth-footer">© {{ currentYear }} {{ siteName }}</div>
      </GlassSurface>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import GlassSurface from '@/components/glass/GlassSurface.vue'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { locale } = useI18n()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const apiBase = computed(() => appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')

const copy = computed(() => isZh.value ? {
  kicker: 'SMIREL API PLATFORM',
  title: '一个入口，连接你的模型与应用。',
  description: '登录后创建 API Key、选择模型并查看真实调用与费用。路由和上游资源留在平台内部。',
  models: '模型与价格', docs: 'API 文档',
} : {
  kicker: 'SMIREL API PLATFORM',
  title: 'One entry point for your models and apps.',
  description: 'Sign in to create API keys, choose models, and inspect real usage and spend. Routing and upstream resources stay behind the platform.',
  models: 'Models & Pricing', docs: 'API Docs',
})

const facts = computed(() => isZh.value ? [
  { title: '统一接口', description: 'OpenAI-compatible 调用方式。' },
  { title: '模型目录', description: '查看能力与价格。' },
  { title: '用量透明', description: '请求、Token 与费用统一追踪。' },
] : [
  { title: 'One endpoint', description: 'OpenAI-compatible requests.' },
  { title: 'Model catalog', description: 'Browse capabilities and pricing.' },
  { title: 'Transparent usage', description: 'Track requests, tokens and spend.' },
])

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
