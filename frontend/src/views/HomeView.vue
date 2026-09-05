<template>
  <div v-if="hasHomeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="smh-home">
    <main class="smh-shell">
      <header class="smh-topbar">
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
          <nav class="smh-portal-nav" aria-label="Public navigation">
            <router-link to="/model-plaza">{{ copy.models }}</router-link>
            <router-link to="/key-usage">{{ copy.keyUsage }}</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
            <router-link :to="isAuthenticated ? dashboardPath : '/login'">
              {{ isAuthenticated ? copy.console : copy.login }}
            </router-link>
          </nav>
          <div class="smh-locale"><LocaleSwitcher /></div>
          <span class="smh-service-status"><i></i>{{ copy.stable }}</span>
        </div>
      </header>

      <section class="smh-title">
        <div class="smh-title-copy">
          <p class="smh-eyebrow">UNIFIED AI API GATEWAY</p>
          <h1>{{ copy.heroTitle }}</h1>
          <p>{{ copy.heroDescription }}</p>
        </div>
        <div class="smh-title-meta" aria-label="Gateway capabilities">
          <span class="smh-title-chip live">{{ copy.online }}</span>
          <span class="smh-title-chip">OpenAI-compatible</span>
        </div>
      </section>

      <section class="smh-dashboard">
        <article class="smh-primary-card smh-card">
          <div class="smh-product">
            <div class="smh-product-icon">API</div>
            <div class="smh-product-copy">
              <strong>{{ siteName }}</strong>
              <span>{{ copy.productSubtitle }}</span>
            </div>
            <span class="smh-product-channel">{{ copy.stableChannel }}</span>
          </div>

          <div class="smh-card-head">
            <p class="smh-kicker">UNIFIED ENDPOINT</p>
            <span class="smh-pill">HTTPS · /V1</span>
          </div>

          <div class="smh-endpoint-large">
            <code>{{ apiBase }}</code>
            <span>{{ copy.endpointHint }}</span>
          </div>

          <p class="smh-summary">{{ copy.primarySummary }}</p>

          <div class="smh-state-row" aria-label="Gateway states">
            <span class="smh-state-item live"><i></i>{{ copy.oneKey }}</span>
            <span class="smh-state-item"><i></i>{{ copy.serverRouting }}</span>
            <span class="smh-state-item"><i></i>{{ copy.failover }}</span>
          </div>

          <div class="smh-meta">
            <div class="smh-meta-item">
              <span>PROTOCOL</span>
              <strong>OpenAI-compatible</strong>
            </div>
            <div class="smh-meta-item">
              <span>AUTH</span>
              <strong>Bearer API Key</strong>
            </div>
            <div class="smh-meta-item">
              <span>MODELS</span>
              <strong>{{ copy.multiProvider }}</strong>
            </div>
          </div>

          <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="smh-main-action">
            <span class="smh-main-action-copy">
              <strong>{{ isAuthenticated ? copy.openConsole : copy.start }}</strong>
              <small>{{ isAuthenticated ? copy.consoleHint : copy.startHint }}</small>
            </span>
            <span class="smh-main-action-arrow">→</span>
          </router-link>
        </article>

        <aside class="smh-start-card smh-card">
          <div class="smh-start-head">
            <div>
              <p class="smh-kicker">QUICK START</p>
              <h2>{{ copy.quickStart }}</h2>
            </div>
            <span class="smh-pill">API</span>
          </div>

          <p class="smh-start-intro">{{ copy.quickIntro }}</p>

          <div class="smh-connection-panel">
            <div class="smh-connection-line">
              <span>BASE URL</span>
              <code>{{ apiBase }}</code>
            </div>
            <div class="smh-connection-line">
              <span>AUTHORIZATION</span>
              <strong>Bearer sk-...</strong>
            </div>
          </div>

          <div class="smh-access-flow" aria-label="Quick start flow">
            <span class="smh-flow-label">ACCESS FLOW</span>
            <div class="smh-flow-steps">
              <div class="smh-flow-step">
                <span>01</span>
                <strong>{{ copy.stepKey }}</strong>
              </div>
              <div class="smh-flow-step">
                <span>02</span>
                <strong>{{ copy.stepModel }}</strong>
              </div>
              <div class="smh-flow-step">
                <span>03</span>
                <strong>{{ copy.stepRequest }}</strong>
              </div>
            </div>
          </div>

          <router-link to="/model-plaza" class="smh-start-link">{{ copy.exploreModels }}</router-link>
        </aside>
      </section>

      <section class="smh-utility-grid" aria-label="Smirel shortcuts">
        <router-link to="/model-plaza" class="smh-utility-card smh-card">
          <span class="smh-utility-icon">M</span>
          <span class="smh-utility-overline">MODELS</span>
          <h3>{{ copy.models }}</h3>
          <p>{{ copy.modelsDescription }}</p>
        </router-link>

        <router-link :to="isAuthenticated ? '/keys' : '/register'" class="smh-utility-card smh-card">
          <span class="smh-utility-icon">K</span>
          <span class="smh-utility-overline">ACCESS</span>
          <h3>{{ copy.apiKeys }}</h3>
          <p>{{ copy.keysDescription }}</p>
        </router-link>

        <router-link to="/key-usage" class="smh-utility-card smh-card">
          <span class="smh-utility-icon">U</span>
          <span class="smh-utility-overline">USAGE</span>
          <h3>{{ copy.keyUsage }}</h3>
          <p>{{ copy.usageDescription }}</p>
        </router-link>

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="smh-utility-card smh-card"
        >
          <span class="smh-utility-icon">D</span>
          <span class="smh-utility-overline">DOCS</span>
          <h3>{{ copy.docs }}</h3>
          <p>{{ copy.docsDescription }}</p>
        </a>
        <router-link v-else to="/model-plaza" class="smh-utility-card smh-card">
          <span class="smh-utility-icon">D</span>
          <span class="smh-utility-overline">DOCS</span>
          <h3>{{ copy.docs }}</h3>
          <p>{{ copy.docsDescription }}</p>
        </router-link>
      </section>

      <footer class="smh-footer">
        <span>© {{ currentYear }} {{ siteName }}</span>
        <span class="smh-footer-status"><i></i>{{ siteSubtitle || 'Unified AI API Gateway' }}</span>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Unified AI API Gateway')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const isHomeContentUrl = computed(() => /^https?:\/\//.test(homeContent.value.trim()))
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')
    .trim()
    .replace(/\/$/, ''),
)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())

const copy = computed(() =>
  isZh.value
    ? {
        models: '模型与价格',
        keyUsage: 'Key 用量',
        docs: 'API 文档',
        login: '登录',
        console: '控制台',
        stable: 'Stable',
        online: '服务在线',
        stableChannel: 'STABLE GATEWAY',
        heroTitle: '一个 Base URL，连接你需要的 AI 模型。',
        heroDescription: 'Smirel 把模型路由、上游账户、故障切换和用量统计留在服务端。开发者只需要自己的 API Key 和统一入口。',
        productSubtitle: 'Unified AI API infrastructure',
        endpointHint: '统一入口 · OpenAI-compatible',
        primarySummary: '保持现有 SDK 与调用方式，把不同模型和上游统一到一个稳定入口；内部路由变化不再影响你的客户端配置。',
        oneKey: '统一 API Key',
        serverRouting: '服务端路由',
        failover: '故障切换',
        multiProvider: '多模型统一接入',
        openConsole: '打开控制台',
        start: '开始使用 Smirel',
        consoleHint: '查看密钥、用量与模型',
        startHint: '创建账户并获取 API Key',
        quickStart: '三步开始调用',
        quickIntro: '不需要理解内部账户池与路由策略。创建自己的密钥、选择模型，然后按兼容协议发起请求。',
        stepKey: '创建 API Key',
        stepModel: '选择模型',
        stepRequest: '发送请求',
        exploreModels: '浏览模型目录',
        apiKeys: 'API Keys',
        modelsDescription: '查看可用模型、能力与当前价格。',
        keysDescription: '创建并管理自己的调用凭证。',
        usageDescription: '按 Key 查询请求、Token 与消费。',
        docsDescription: '查看接口格式、鉴权和接入示例。',
      }
    : {
        models: 'Models & Pricing',
        keyUsage: 'Key Usage',
        docs: 'API Docs',
        login: 'Sign in',
        console: 'Console',
        stable: 'Stable',
        online: 'Service online',
        stableChannel: 'STABLE GATEWAY',
        heroTitle: 'One Base URL for the AI models you need.',
        heroDescription: 'Smirel keeps model routing, upstream accounts, failover and usage tracking on the server. Developers only keep their own API key and one endpoint.',
        productSubtitle: 'Unified AI API infrastructure',
        endpointHint: 'One endpoint · OpenAI-compatible',
        primarySummary: 'Keep your existing SDK and request flow while multiple models and upstreams sit behind one stable gateway. Internal routing changes no longer affect client configuration.',
        oneKey: 'One API key',
        serverRouting: 'Server-side routing',
        failover: 'Automatic failover',
        multiProvider: 'Multi-model access',
        openConsole: 'Open console',
        start: 'Start with Smirel',
        consoleHint: 'Manage keys, usage and models',
        startHint: 'Create an account and API key',
        quickStart: 'Start in three steps',
        quickIntro: 'You do not need to understand internal account pools or routing policies. Create a key, choose a model and send a compatible request.',
        stepKey: 'Create API key',
        stepModel: 'Choose model',
        stepRequest: 'Send request',
        exploreModels: 'Browse model catalog',
        apiKeys: 'API Keys',
        modelsDescription: 'Browse available models, capabilities and current pricing.',
        keysDescription: 'Create and manage your own API credentials.',
        usageDescription: 'Inspect requests, tokens and spend by API key.',
        docsDescription: 'Read request formats, authentication and examples.',
      },
)

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
