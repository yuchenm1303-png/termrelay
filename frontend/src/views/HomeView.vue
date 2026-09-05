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

  <div v-else class="smv3-home">
    <header class="smv3-home-nav">
      <router-link to="/home" class="smv3-home-brand">
        <img v-if="siteLogo" :src="siteLogo" alt="Smirel" />
        <span>{{ siteName }}</span>
      </router-link>

      <nav class="smv3-home-links">
        <router-link to="/model-plaza">{{ copy.models }}</router-link>
        <router-link to="/key-usage">{{ copy.keyUsage }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a>
        <LocaleSwitcher />
        <button type="button" class="smv3-home-secondary" style="height: 34px; padding: 0 10px" @click="toggleTheme">
          {{ isDark ? 'Light' : 'Dark' }}
        </button>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="smv3-home-cta">
          {{ isAuthenticated ? copy.console : copy.login }}
        </router-link>
      </nav>
    </header>

    <main>
      <section class="smv3-home-hero">
        <div>
          <div class="smv3-home-kicker">UNIFIED AI API INFRASTRUCTURE</div>
          <h1>{{ copy.heroTitle }}</h1>
          <p>{{ copy.heroDescription }}</p>

          <div class="smv3-home-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="smv3-home-primary">
              {{ isAuthenticated ? copy.openConsole : copy.start }}
            </router-link>
            <router-link to="/model-plaza" class="smv3-home-secondary">{{ copy.exploreModels }}</router-link>
          </div>

          <div class="smv3-home-proof">
            <span>OpenAI-compatible</span>
            <span>{{ copy.oneKey }}</span>
            <span>{{ copy.serverRouting }}</span>
            <span>{{ copy.usageTracking }}</span>
          </div>
        </div>

        <div class="smv3-home-command">
          <div class="smv3-home-command-head">
            <div class="smv3-home-lights"><i></i><i></i><i></i></div>
            <span>smirel / quickstart</span>
            <span>HTTPS</span>
          </div>
          <div class="smv3-home-command-body">
            <div class="smv3-home-endpoint">
              <span>API BASE</span>
              <code>{{ apiBase }}</code>
            </div>
            <pre class="smv3-home-code">curl {{ apiBase }}/chat/completions \
  -H <b>"Authorization: Bearer sk-..."</b> \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.5",
    "messages": [{"role":"user","content":"Hello"}]
  }'</pre>
            <div class="smv3-home-metrics">
              <div class="smv3-home-metric"><strong>1</strong><span>Base URL</span></div>
              <div class="smv3-home-metric"><strong>1</strong><span>API Key</span></div>
              <div class="smv3-home-metric"><strong>N</strong><span>Models</span></div>
            </div>
          </div>
        </div>
      </section>

      <section class="smv3-home-band">
        <article v-for="provider in providers" :key="provider.name">
          <strong>{{ provider.name }}</strong>
          <span>{{ provider.description }}</span>
        </article>
      </section>

      <section class="smv3-home-section">
        <div class="smv3-home-section-head">
          <h2>{{ copy.sectionTitle }}</h2>
          <p>{{ copy.sectionDescription }}</p>
        </div>
        <div class="smv3-home-feature-grid">
          <article v-for="feature in features" :key="feature.index" class="smv3-home-feature">
            <div class="smv3-home-feature-index">{{ feature.index }}</div>
            <h3>{{ feature.title }}</h3>
            <p>{{ feature.description }}</p>
          </article>
        </div>
      </section>
    </main>

    <footer class="smv3-home-footer">
      <span>© {{ currentYear }} {{ siteName }}</span>
      <span>{{ siteSubtitle || 'Unified AI API Gateway' }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const isDark = ref(document.documentElement.classList.contains('dark'))
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
        console: '进入控制台',
        openConsole: '打开控制台',
        start: '开始使用',
        exploreModels: '浏览模型目录',
        heroTitle: '把复杂的 AI 上游，变成一个简单的 API。',
        heroDescription: 'Smirel 为开发者提供统一的模型入口。你只需要维护一个 Base URL 和自己的 API Key，模型路由、上游账号、故障切换与用量统计都留在服务端。',
        oneKey: '统一 API Key',
        serverRouting: '服务端路由',
        usageTracking: '用量可观测',
        sectionTitle: '从第一次请求，到日常运营，都在同一个工作区。',
        sectionDescription: '用户只看到调用 API 所需要的信息；上游账户、路由策略和运营能力则留在独立的管理工作区。',
      }
    : {
        models: 'Models & Pricing',
        keyUsage: 'Key Usage',
        docs: 'API Docs',
        login: 'Sign in',
        console: 'Open console',
        openConsole: 'Open console',
        start: 'Get started',
        exploreModels: 'Explore models',
        heroTitle: 'Turn complex AI upstreams into one simple API.',
        heroDescription: 'Smirel gives developers one model gateway. Keep one Base URL and your own API key while routing, upstream accounts, failover and usage tracking stay on the server.',
        oneKey: 'One API key',
        serverRouting: 'Server-side routing',
        usageTracking: 'Observable usage',
        sectionTitle: 'From the first request to daily operations, in one workspace.',
        sectionDescription: 'Customers see only what they need to call the API. Upstream accounts, routing policies and operations stay in a separate admin workspace.',
      },
)

const providers = computed(() => [
  { name: 'OpenAI', description: isZh.value ? 'Responses、Chat Completions 与兼容 SDK。' : 'Responses, Chat Completions and compatible SDKs.' },
  { name: 'Anthropic', description: isZh.value ? 'Claude 系列模型统一接入。' : 'Unified access to Claude models.' },
  { name: 'Gemini', description: isZh.value ? 'Gemini 模型与 CLI 工作流。' : 'Gemini models and CLI workflows.' },
  { name: 'More', description: isZh.value ? '继续扩展更多模型与协议。' : 'Extend to more models and protocols.' },
])

const features = computed(() =>
  isZh.value
    ? [
        { index: '01', title: '创建与隔离 API Key', description: '每个用户管理自己的凭证、额度和访问策略，不暴露上游账户信息。' },
        { index: '02', title: '模型目录与价格透明', description: '直接查看可用模型、能力与计费，不需要理解内部路由分组。' },
        { index: '03', title: '调用、Token 与费用可追踪', description: '统一查看使用日志、消费和服务状态，排查问题更直接。' },
      ]
    : [
        { index: '01', title: 'Create isolated API keys', description: 'Each customer controls their own credentials, quota and access without seeing upstream accounts.' },
        { index: '02', title: 'Transparent models and pricing', description: 'Browse available models, capabilities and pricing without internal routing concepts.' },
        { index: '03', title: 'Track requests, tokens and spend', description: 'Usage logs, spend and service health live in one place for faster troubleshooting.' },
      ],
)

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>
