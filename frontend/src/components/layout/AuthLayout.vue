<template>
  <div class="smv3-auth">
    <header class="smv3-auth-nav">
      <router-link to="/home" class="smv3-auth-brand">
        <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
        <span>{{ siteName }}</span>
      </router-link>
      <div class="smv3-auth-links">
        <router-link to="/model-plaza">{{ copy.models }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
        <LocaleSwitcher />
      </div>
    </header>

    <main class="smv3-auth-main">
      <section class="smv3-auth-story">
        <div class="smv3-auth-story-inner">
          <div class="smv3-auth-story-kicker">{{ copy.kicker }}</div>
          <h1>{{ copy.title }}</h1>
          <p>{{ copy.description }}</p>

          <div class="smv3-auth-facts">
            <article v-for="fact in facts" :key="fact.title" class="smv3-auth-fact">
              <strong>{{ fact.title }}</strong>
              <span>{{ fact.description }}</span>
            </article>
          </div>

          <div class="smv3-code-block" style="margin-top: 22px">
            <span style="color: var(--smv3-faint)">BASE_URL</span><br />
            {{ apiBase }}<br /><br />
            <span style="color: var(--smv3-faint)">AUTH</span><br />
            Authorization: Bearer sk-••••••••
          </div>
        </div>
      </section>

      <section class="smv3-auth-form-wrap">
        <div class="smv3-auth-card">
          <div class="smv3-auth-card-shell">
            <slot />
          </div>
          <div class="smv3-auth-footer">
            <slot name="footer" />
          </div>
          <div class="smv3-auth-footer">© {{ currentYear }} {{ siteName }}</div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
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
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)
const apiBase = computed(() =>
  appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1',
)

const copy = computed(() =>
  isZh.value
    ? {
        kicker: 'SMIREL DEVELOPER PLATFORM',
        title: '把模型接入变成一件简单的事。',
        description: '一个账号、一套 API Key、一个稳定的 Base URL。模型选择、路由与上游资源由 Smirel 在服务端统一处理。',
        models: '模型与价格',
        docs: 'API 文档',
      }
    : {
        kicker: 'SMIREL DEVELOPER PLATFORM',
        title: 'A simpler way to connect AI models.',
        description: 'One account, one API key, one stable base URL. Smirel handles model routing and upstream resources behind the gateway.',
        models: 'Models & Pricing',
        docs: 'API Docs',
      },
)

const facts = computed(() =>
  isZh.value
    ? [
        { title: '统一接口', description: 'OpenAI-compatible 调用方式。' },
        { title: '模型目录', description: '直接查看模型与计费信息。' },
        { title: '可观测', description: '调用、Token 与费用统一追踪。' },
      ]
    : [
        { title: 'One endpoint', description: 'OpenAI-compatible requests.' },
        { title: 'Model catalog', description: 'Browse models and pricing directly.' },
        { title: 'Observable', description: 'Track requests, tokens and spend.' },
      ],
)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
