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
          <span class="smh-service-status"><i></i>{{ copy.online }}</span>
        </div>
      </header>

      <section class="smh-console-intro">
        <div class="smh-console-intro-copy">
          <p class="smh-eyebrow">SMIREL GATEWAY</p>
          <h1>{{ copy.heroTitle }}</h1>
          <p>{{ copy.heroDescription }}</p>
        </div>
        <div class="smh-intro-status" aria-label="Gateway status">
          <span class="smh-title-chip live">{{ copy.online }}</span>
          <span class="smh-title-chip">OpenAI-compatible</span>
          <span class="smh-title-chip">HTTPS / v1</span>
        </div>
      </section>

      <section class="smh-command-grid">
        <article class="smh-gateway-card smh-card">
          <div class="smh-gateway-head">
            <div class="smh-gateway-identity">
              <span class="smh-product-icon">API</span>
              <div>
                <p class="smh-kicker">GATEWAY OVERVIEW</p>
                <strong>{{ siteName }}</strong>
                <span>{{ copy.productSubtitle }}</span>
              </div>
            </div>
            <span class="smh-product-channel"><i></i>{{ copy.operational }}</span>
          </div>

          <div class="smh-endpoint-block">
            <span>{{ copy.endpointLabel }}</span>
            <code>{{ apiBase }}</code>
          </div>

          <p class="smh-summary">{{ copy.primarySummary }}</p>

          <div class="smh-state-row" aria-label="Gateway capabilities">
            <span class="smh-state-item live"><i></i>{{ copy.oneKey }}</span>
            <span class="smh-state-item"><i></i>{{ copy.serverRouting }}</span>
            <span class="smh-state-item"><i></i>{{ copy.failover }}</span>
            <span class="smh-state-item"><i></i>{{ copy.usageTracking }}</span>
          </div>

          <div class="smh-gateway-controls">
            <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="smh-primary-action">
              <span>
                <strong>{{ isAuthenticated ? copy.openConsole : copy.start }}</strong>
                <small>{{ isAuthenticated ? copy.consoleHint : copy.startHint }}</small>
              </span>
              <span class="smh-action-arrow">→</span>
            </router-link>
            <router-link to="/model-plaza" class="smh-secondary-action">
              <span>{{ copy.exploreModels }}</span>
              <span class="smh-action-arrow">→</span>
            </router-link>
          </div>

          <div class="smh-gateway-meta">
            <div>
              <span>PROTOCOL</span>
              <strong>OpenAI-compatible</strong>
            </div>
            <div>
              <span>AUTH</span>
              <strong>Bearer API Key</strong>
            </div>
            <div>
              <span>ROUTING</span>
              <strong>{{ copy.managedRouting }}</strong>
            </div>
            <div>
              <span>MODELS</span>
              <strong>{{ copy.multiProvider }}</strong>
            </div>
          </div>
        </article>

        <aside class="smh-side-stack">
          <article class="smh-connection-card smh-card">
            <div class="smh-side-card-head">
              <div>
                <p class="smh-kicker">CONNECTION PROFILE</p>
                <h2>{{ copy.connectionProfile }}</h2>
              </div>
              <span class="smh-pill">HTTPS</span>
            </div>

            <div class="smh-connection-panel">
              <div class="smh-connection-line">
                <span>BASE URL</span>
                <code>{{ apiBase }}</code>
              </div>
              <div class="smh-connection-line">
                <span>AUTHORIZATION</span>
                <strong>Bearer sk-...</strong>
              </div>
              <div class="smh-connection-line">
                <span>CONTENT TYPE</span>
                <strong>application/json</strong>
              </div>
            </div>

            <router-link :to="isAuthenticated ? '/keys' : '/register'" class="smh-inline-link">
              <span>{{ copy.manageKeys }}</span>
              <span class="smh-action-arrow">→</span>
            </router-link>
          </article>

          <article class="smh-access-card smh-card">
            <div class="smh-side-card-head">
              <div>
                <p class="smh-kicker">ACCESS FLOW</p>
                <h2>{{ copy.requestFlow }}</h2>
              </div>
              <span class="smh-pill">API</span>
            </div>

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

            <p class="smh-access-note">{{ copy.accessNote }}</p>
          </article>
        </aside>
      </section>

      <section class="smh-provider-rail smh-card" aria-label="Supported providers">
        <div class="smh-provider-rail-copy">
          <p class="smh-kicker">MODEL NETWORK</p>
          <strong>{{ copy.providerRailTitle }}</strong>
        </div>
        <div class="smh-provider-list">
          <div v-for="provider in providers" :key="provider.name" class="smh-provider-item">
            <span class="smh-provider-dot"></span>
            <div>
              <strong>{{ provider.name }}</strong>
              <small>{{ provider.description }}</small>
            </div>
          </div>
        </div>
      </section>

      <section class="smh-control-dock smh-card" aria-label="Gateway controls">
        <router-link :to="isAuthenticated ? '/keys' : '/register'" class="smh-dock-item smh-dock-primary">
          <span class="smh-dock-index">01</span>
          <div>
            <span class="smh-utility-overline">ACCESS</span>
            <h3>{{ copy.apiKeys }}</h3>
            <p>{{ copy.keysDescription }}</p>
          </div>
          <span class="smh-action-arrow">→</span>
        </router-link>

        <router-link to="/model-plaza" class="smh-dock-item">
          <span class="smh-dock-index">02</span>
          <div>
            <span class="smh-utility-overline">MODELS</span>
            <h3>{{ copy.models }}</h3>
            <p>{{ copy.modelsDescription }}</p>
          </div>
          <span class="smh-action-arrow">→</span>
        </router-link>

        <router-link to="/key-usage" class="smh-dock-item">
          <span class="smh-dock-index">03</span>
          <div>
            <span class="smh-utility-overline">USAGE</span>
            <h3>{{ copy.keyUsage }}</h3>
            <p>{{ copy.usageDescription }}</p>
          </div>
          <span class="smh-action-arrow">→</span>
        </router-link>

        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="smh-dock-item">
          <span class="smh-dock-index">04</span>
          <div>
            <span class="smh-utility-overline">DOCS</span>
            <h3>{{ copy.docs }}</h3>
            <p>{{ copy.docsDescription }}</p>
          </div>
          <span class="smh-action-arrow">↗</span>
        </a>
        <router-link v-else to="/model-plaza" class="smh-dock-item">
          <span class="smh-dock-index">04</span>
          <div>
            <span class="smh-utility-overline">DOCS</span>
            <h3>{{ copy.docs }}</h3>
            <p>{{ copy.docsDescription }}</p>
          </div>
          <span class="smh-action-arrow">→</span>
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

const providers = computed(() => [
  { name: 'OpenAI', description: isZh.value ? 'Responses / Chat Completions' : 'Responses / Chat Completions' },
  { name: 'Anthropic', description: isZh.value ? 'Claude 模型' : 'Claude models' },
  { name: 'Gemini', description: isZh.value ? 'Gemini 模型' : 'Gemini models' },
  { name: 'More', description: isZh.value ? '持续扩展协议与模型' : 'More protocols and models' },
])

const copy = computed(() =>
  isZh.value
    ? {
        models: '模型与价格',
        keyUsage: 'Key 用量',
        docs: 'API 文档',
        login: '登录',
        console: '进入控制台',
        online: '服务在线',
        operational: 'Operational',
        heroTitle: '统一模型入口，保持调用简单。',
        heroDescription: 'Smirel 把模型路由、上游账户、故障切换和用量统计留在服务端；客户端只需要一个稳定入口和自己的 API Key。',
        productSubtitle: 'Unified AI API infrastructure',
        endpointLabel: 'PRIMARY ENDPOINT',
        primarySummary: '保持现有 SDK 与调用方式，把不同模型和上游统一到一个稳定入口。内部路由变化不会影响客户端配置。',
        oneKey: '统一 API Key',
        serverRouting: '服务端路由',
        failover: '故障切换',
        usageTracking: '用量可观测',
        managedRouting: '服务端托管',
        multiProvider: '多模型统一接入',
        openConsole: '打开控制台',
        start: '开始使用 Smirel',
        consoleHint: '查看密钥、用量与模型',
        startHint: '创建账户并获取 API Key',
        exploreModels: '浏览模型目录',
        connectionProfile: '连接配置',
        manageKeys: '管理 API Keys',
        requestFlow: '请求路径',
        stepKey: '创建 API Key',
        stepModel: '选择模型',
        stepRequest: '发送请求',
        accessNote: '调用协议保持稳定，路由、上游切换和运行策略由服务端负责。',
        providerRailTitle: '一个入口，连接不同模型协议。',
        apiKeys: 'API Keys',
        modelsDescription: '查看可用模型、能力与当前价格。',
        keysDescription: '创建并管理自己的调用凭证。',
        usageDescription: '按 Key 查询请求、Token 与消费。',
        docsDescription: '查看接口格式、鉴权与接入示例。',
      }
    : {
        models: 'Models & Pricing',
        keyUsage: 'Key Usage',
        docs: 'API Docs',
        login: 'Sign in',
        console: 'Open console',
        online: 'Service online',
        operational: 'Operational',
        heroTitle: 'One model gateway. Keep every request simple.',
        heroDescription: 'Smirel keeps routing, upstream accounts, failover and usage tracking on the server. Clients keep one stable endpoint and their own API key.',
        productSubtitle: 'Unified AI API infrastructure',
        endpointLabel: 'PRIMARY ENDPOINT',
        primarySummary: 'Keep your existing SDK and request flow while multiple models and upstreams sit behind one stable gateway. Internal routing changes stay invisible to clients.',
        oneKey: 'One API key',
        serverRouting: 'Server-side routing',
        failover: 'Automatic failover',
        usageTracking: 'Observable usage',
        managedRouting: 'Server managed',
        multiProvider: 'Multi-model access',
        openConsole: 'Open console',
        start: 'Start with Smirel',
        consoleHint: 'Manage keys, usage and models',
        startHint: 'Create an account and API key',
        exploreModels: 'Browse model catalog',
        connectionProfile: 'Connection profile',
        manageKeys: 'Manage API keys',
        requestFlow: 'Request path',
        stepKey: 'Create API key',
        stepModel: 'Choose model',
        stepRequest: 'Send request',
        accessNote: 'The request contract stays stable while routing, upstream switching and runtime policy stay server-side.',
        providerRailTitle: 'One endpoint across multiple model protocols.',
        apiKeys: 'API Keys',
        modelsDescription: 'Browse available models, capabilities and current pricing.',
        keysDescription: 'Create and manage your own API credentials.',
        usageDescription: 'Inspect requests, tokens and spend by API key.',
        docsDescription: 'Read request formats, authentication and integration examples.',
      },
)

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
