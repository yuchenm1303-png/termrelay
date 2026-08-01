<template>
  <div v-if="hasHomeContent" class="min-h-screen bg-[#05070b]">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="termrelay-shell">
    <header class="site-header page-frame">
      <router-link to="/home" class="brand" aria-label="TermRelay home">
        <span class="brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="" />
          <span v-else>TR</span>
        </span>
        <span class="brand-copy">
          <strong>{{ siteName }}</strong>
          <small>{{ copy.brandSubtitle }}</small>
        </span>
      </router-link>

      <nav class="site-nav" :aria-label="copy.primaryNavigation">
        <a href="#system">{{ copy.navSystem }}</a>
        <a href="#console">{{ copy.navConsole }}</a>
        <a href="#capabilities">{{ copy.navCapabilities }}</a>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
          {{ copy.navDocs }}
        </a>

        <div class="language-switch" :aria-label="copy.languageLabel" role="group">
          <button
            type="button"
            :class="{ active: currentLanguage === 'zh' }"
            :disabled="switchingLocale"
            @click="changeLocale('zh')"
          >
            中文
          </button>
          <button
            type="button"
            :class="{ active: currentLanguage === 'en' }"
            :disabled="switchingLocale"
            @click="changeLocale('en')"
          >
            EN
          </button>
        </div>

        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-cta">
          {{ isAuthenticated ? copy.openConsole : copy.signIn }}
        </router-link>
      </nav>
    </header>

    <main>
      <section id="system" class="hero page-frame">
        <div class="hero-copy">
          <p class="eyebrow"><span></span>{{ copy.eyebrow }}</p>
          <h1>
            <span>{{ copy.heroLineOne }}</span>
            <em>{{ copy.heroLineTwo }}</em>
          </h1>
          <p class="hero-description">{{ localizedSubtitle }}</p>

          <div class="hero-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="primary-action">
              <span>{{ isAuthenticated ? copy.enterConsole : copy.initializeSession }}</span>
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
            </router-link>

            <button type="button" class="secondary-action" @click="copyBaseUrl">
              <span>{{ copied ? copy.copied : copy.copyBaseUrl }}</span>
              <code>{{ baseUrl }}</code>
            </button>
          </div>

          <div class="signal-row" :aria-label="copy.gatewayCapabilities">
            <span><i class="signal online"></i>{{ copy.gatewayReachable }}</span>
            <span><i class="signal"></i>{{ copy.responsesApi }}</span>
            <span><i class="signal"></i>{{ copy.oauthUpstream }}</span>
          </div>
        </div>

        <aside class="node-card" :aria-label="copy.nodeOverview">
          <div class="node-card-head">
            <div>
              <p>{{ copy.nodeOverview }}</p>
              <strong>TR-01</strong>
            </div>
            <span class="node-live"><i></i>{{ copy.live }}</span>
          </div>

          <div class="node-signal-map" aria-hidden="true">
            <span class="map-line map-line-a"></span>
            <span class="map-line map-line-b"></span>
            <span class="map-flare map-flare-a"></span>
            <span class="map-flare map-flare-b"></span>
            <span class="map-dot map-dot-a"></span>
            <span class="map-dot map-dot-b"></span>
            <span class="map-dot map-dot-c"></span>
          </div>

          <div class="node-metrics">
            <div>
              <span>{{ copy.route }}</span>
              <strong>Responses</strong>
            </div>
            <div>
              <span>{{ copy.auth }}</span>
              <strong>OAuth</strong>
            </div>
            <div>
              <span>{{ copy.mode }}</span>
              <strong>{{ copy.privateMode }}</strong>
            </div>
            <div>
              <span>{{ copy.state }}</span>
              <strong>{{ copy.ready }}</strong>
            </div>
          </div>

          <div class="node-endpoint">
            <span>{{ copy.endpoint }}</span>
            <code>{{ baseUrl }}</code>
            <button type="button" @click="copyBaseUrl">{{ copied ? copy.done : copy.copy }}</button>
          </div>
        </aside>
      </section>

      <section id="console" class="console-section page-frame">
        <div class="console-intro">
          <p class="section-kicker">{{ copy.consoleKicker }}</p>
          <h2>{{ copy.consoleTitle }}</h2>
          <p>{{ copy.consoleDescription }}</p>

          <div class="quick-command">
            <span>{{ copy.quickStart }}</span>
            <code>export OPENAI_BASE_URL={{ baseUrl }}</code>
            <button type="button" @click="copyBaseUrl">{{ copied ? copy.done : copy.copy }}</button>
          </div>

          <div class="terminal-meta">
            <div><span>{{ copy.node }}</span><strong>TR-01</strong></div>
            <div><span>{{ copy.mode }}</span><strong>{{ copy.privateMode }}</strong></div>
            <div><span>{{ copy.time }}</span><strong>{{ clock }}</strong></div>
          </div>
        </div>

        <div class="terminal-wrap" :aria-label="copy.terminalPreview">
          <div class="terminal-window">
            <div class="terminal-bar">
              <div class="window-controls" aria-hidden="true">
                <span></span><span></span><span></span>
              </div>
              <div class="terminal-title">jack@termrelay: ~/gateway</div>
              <div class="terminal-state"><i></i>{{ copy.live }}</div>
            </div>

            <div class="terminal-body">
              <div class="terminal-line delay-1">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span>status --verbose</span>
              </div>

              <div class="terminal-output delay-2">
                <div><span>{{ copy.service }}</span><strong>TermRelay Gateway</strong></div>
                <div><span>{{ copy.transport }}</span><strong>OpenAI Responses</strong></div>
                <div><span>{{ copy.auth }}</span><strong>Bearer / OAuth upstream</strong></div>
                <div><span>{{ copy.endpoint }}</span><strong>{{ baseUrl }}</strong></div>
              </div>

              <div class="terminal-line delay-3">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span>curl {{ baseUrl }}/models</span>
              </div>

              <div class="terminal-json delay-4">
                <span>{</span>
                <span class="indent">"object": "list",</span>
                <span class="indent">"route": "ready",</span>
                <span class="indent">"stream": true</span>
                <span>}</span>
              </div>

              <div class="terminal-line delay-5">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span class="cursor"></span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="capabilities" class="capabilities page-frame">
        <div class="section-heading">
          <div>
            <p class="section-kicker">cat capabilities.md</p>
            <span>{{ copy.capabilitiesLabel }}</span>
          </div>
          <h2>{{ copy.capabilitiesHeadline }}</h2>
        </div>

        <div class="capability-grid">
          <article v-for="item in copy.capabilities" :key="item.index">
            <div class="article-topline">
              <span class="article-index">{{ item.index }}</span>
              <span class="article-tag">{{ item.tag }}</span>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </section>

      <section class="manifesto page-frame">
        <div>
          <p class="section-kicker">README / 00</p>
          <span>{{ copy.manifestoLabel }}</span>
        </div>
        <p>{{ copy.manifesto }}</p>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'">
          {{ copy.launchConsole }} <span>↗</span>
        </router-link>
      </section>
    </main>

    <footer class="page-frame">
      <div>
        <span>© {{ currentYear }} {{ siteName }}</span>
        <span>{{ copy.builtOn }}</span>
      </div>
      <a href="https://github.com/yuchenm1303-png/termrelay" target="_blank" rel="noopener noreferrer">
        {{ copy.githubSource }}
      </a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { setLocale } from '@/i18n'
import { sanitizeUrl } from '@/utils/url'

const authStore = useAuthStore()
const appStore = useAppStore()
const { locale } = useI18n()

const COPY = {
  en: {
    brandSubtitle: 'personal ai relay node',
    primaryNavigation: 'Primary navigation',
    navSystem: 'system',
    navConsole: 'console',
    navCapabilities: 'capabilities',
    navDocs: 'docs',
    languageLabel: 'Language',
    openConsole: 'open console',
    signIn: 'sign in',
    eyebrow: 'self-hosted gateway / node 01',
    heroLineOne: 'Your models.',
    heroLineTwo: 'Your endpoint.',
    fallbackSubtitle:
      'A terminal-inspired AI relay gateway for your own models, credentials and clients.',
    enterConsole: 'Enter console',
    initializeSession: 'Initialize session',
    copied: 'Copied',
    copyBaseUrl: 'Copy base URL',
    gatewayCapabilities: 'Gateway capabilities',
    gatewayReachable: 'gateway reachable',
    responsesApi: 'responses api',
    oauthUpstream: 'oauth upstream',
    nodeOverview: 'Node overview',
    live: 'live',
    route: 'route',
    auth: 'auth',
    mode: 'mode',
    privateMode: 'private',
    state: 'state',
    ready: 'ready',
    endpoint: 'endpoint',
    done: 'done',
    copy: 'copy',
    consoleKicker: 'operator console / preview',
    consoleTitle: 'One calm surface for the entire relay.',
    consoleDescription:
      'Inspect the route, credentials and request flow without turning the gateway into another crowded SaaS panel.',
    quickStart: 'quick start',
    node: 'node',
    time: 'time',
    terminalPreview: 'TermRelay terminal preview',
    service: 'service',
    transport: 'transport',
    capabilitiesLabel: 'gateway primitives',
    capabilitiesHeadline: 'Built as infrastructure. Presented as a personal node.',
    capabilities: [
      {
        index: '01',
        title: 'Unified endpoint',
        description:
          'Expose one stable API base URL while the gateway handles upstream routing and protocol compatibility.',
        tag: '/v1'
      },
      {
        index: '02',
        title: 'Credential control',
        description:
          'Keep downstream keys separate from upstream OAuth credentials, with revocation and usage visibility.',
        tag: 'auth'
      },
      {
        index: '03',
        title: 'Streaming native',
        description:
          'Forward long-running Responses API streams through a focused, observable relay path.',
        tag: 'sse'
      },
      {
        index: '04',
        title: 'Observable by default',
        description:
          'Track requests, latency, failures and account health from one operator console.',
        tag: 'logs'
      }
    ],
    manifestoLabel: 'design intent',
    manifesto:
      'TermRelay is a small, self-hosted AI gateway with the clarity of infrastructure and the character of a personal terminal.',
    launchConsole: 'launch console',
    builtOn: 'built on Sub2API',
    githubSource: 'github / source'
  },
  zh: {
    brandSubtitle: '个人 AI 中转节点',
    primaryNavigation: '主导航',
    navSystem: '系统',
    navConsole: '控制台',
    navCapabilities: '能力',
    navDocs: '文档',
    languageLabel: '语言切换',
    openConsole: '打开控制台',
    signIn: '登录',
    eyebrow: '自托管网关 / 节点 01',
    heroLineOne: '你的模型。',
    heroLineTwo: '你的入口。',
    fallbackSubtitle: '一个为你的模型、凭证与客户端而设计的终端风格 AI 中转网关。',
    enterConsole: '进入控制台',
    initializeSession: '初始化会话',
    copied: '已复制',
    copyBaseUrl: '复制 Base URL',
    gatewayCapabilities: '网关能力',
    gatewayReachable: '网关在线',
    responsesApi: 'Responses API',
    oauthUpstream: 'OAuth 上游',
    nodeOverview: '节点概览',
    live: '在线',
    route: '路由',
    auth: '认证',
    mode: '模式',
    privateMode: '私有',
    state: '状态',
    ready: '就绪',
    endpoint: '入口',
    done: '完成',
    copy: '复制',
    consoleKicker: '运营控制台 / 预览',
    consoleTitle: '用一个安静、清晰的界面管理整条中转链路。',
    consoleDescription: '查看路由、凭证和请求状态，同时避免把网关做成拥挤、同质化的 SaaS 后台。',
    quickStart: '快速接入',
    node: '节点',
    time: '时间',
    terminalPreview: 'TermRelay 终端预览',
    service: '服务',
    transport: '协议',
    capabilitiesLabel: '网关基础能力',
    capabilitiesHeadline: '以基础设施为内核，以个人节点的方式呈现。',
    capabilities: [
      {
        index: '01',
        title: '统一 API 入口',
        description: '对外提供稳定的 Base URL，由网关负责上游路由和协议兼容。',
        tag: '/v1'
      },
      {
        index: '02',
        title: '凭证隔离管理',
        description: '将下游 API Key 与上游 OAuth 凭证分离，并支持撤销和用量查看。',
        tag: '认证'
      },
      {
        index: '03',
        title: '原生流式转发',
        description: '稳定转发长时间运行的 Responses API 流，并保持链路清晰可观测。',
        tag: 'SSE'
      },
      {
        index: '04',
        title: '默认可观测',
        description: '在统一控制台查看请求、延迟、错误与上游账号健康状态。',
        tag: '日志'
      }
    ],
    manifestoLabel: '设计意图',
    manifesto:
      'TermRelay 是一套小型、自托管的 AI 网关，既保留基础设施应有的清晰，也保留个人终端的气质。',
    launchConsole: '启动控制台',
    builtOn: '基于 Sub2API 构建',
    githubSource: 'GitHub / 源码'
  }
} as const

const currentLanguage = computed<'en' | 'zh'>(() => (locale.value === 'zh' ? 'zh' : 'en'))
const copy = computed(() => COPY[currentLanguage.value])

const siteName = computed(() => {
  const name = appStore.cachedPublicSettings?.site_name || appStore.siteName
  return !name || name === 'Sub2API' ? 'TermRelay' : name
})
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true
  })
)
const configuredSubtitle = computed(() =>
  (appStore.cachedPublicSettings?.site_subtitle || '').trim()
)
const localizedSubtitle = computed(() => configuredSubtitle.value || copy.value.fallbackSubtitle)
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = new Date().getFullYear()
const baseUrl = computed(() => {
  const configured = appStore.apiBaseUrl?.trim()
  if (configured) return configured.replace(/\/$/, '')
  if (typeof window === 'undefined') return '/v1'

  const basePath = import.meta.env.BASE_URL === '/' ? '' : import.meta.env.BASE_URL.replace(/\/$/, '')
  return `${window.location.origin}${basePath}/v1`
})

const copied = ref(false)
const clock = ref('00:00:00')
const switchingLocale = ref(false)
let clockTimer: number | undefined
let copiedTimer: number | undefined

function updateClock() {
  clock.value = new Intl.DateTimeFormat(currentLanguage.value === 'zh' ? 'zh-CN' : 'en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).format(new Date())
}

async function changeLocale(code: 'en' | 'zh') {
  if (switchingLocale.value || currentLanguage.value === code) return

  switchingLocale.value = true
  try {
    await setLocale(code)
    updateClock()
  } finally {
    switchingLocale.value = false
  }
}

async function copyBaseUrl() {
  try {
    await navigator.clipboard.writeText(baseUrl.value)
    copied.value = true
    if (copiedTimer) window.clearTimeout(copiedTimer)
    copiedTimer = window.setTimeout(() => {
      copied.value = false
    }, 1800)
  } catch {
    copied.value = false
  }
}

onMounted(() => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})

onBeforeUnmount(() => {
  if (clockTimer) window.clearInterval(clockTimer)
  if (copiedTimer) window.clearTimeout(copiedTimer)
})
</script>

<style scoped>
.termrelay-shell {
  --line: rgba(218, 232, 232, 0.1);
  --line-strong: rgba(218, 232, 232, 0.18);
  --text: #f3f1eb;
  --muted: #a5aaae;
  --dim: #6f777c;
  --rose: #d9a5b7;
  --iris: #b7a8cf;
  --foam: #87d4cf;
  --pine: #2f7880;
  position: relative;
  min-height: 100vh;
  overflow: clip;
  color: var(--text);
  background: transparent;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.page-frame {
  position: relative;
  z-index: 2;
  width: min(1160px, calc(100% - 48px));
  margin-inline: auto;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  min-height: 88px;
  border-bottom: 1px solid var(--line);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: inherit;
  text-decoration: none;
}

.brand-mark {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: linear-gradient(145deg, rgba(135, 212, 207, 0.12), rgba(183, 168, 207, 0.08));
  color: var(--foam);
  font: 700 11px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.08em;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.brand-copy {
  display: grid;
  gap: 3px;
}

.brand-copy strong {
  font-size: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.brand-copy small {
  color: var(--dim);
  font: 500 9px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.05em;
}

.site-nav {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 20px;
}

.site-nav > a {
  color: var(--muted);
  font: 500 11px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-decoration: none;
  transition: color 160ms ease;
}

.site-nav > a:hover {
  color: var(--text);
}

.site-nav .nav-cta {
  padding: 11px 15px;
  border: 1px solid var(--line-strong);
  border-radius: 9px;
  background: rgba(5, 10, 12, 0.46);
  color: var(--text);
  white-space: nowrap;
}

.language-switch {
  display: inline-grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px;
  padding: 3px;
  border: 1px solid var(--line);
  border-radius: 9px;
  background: rgba(4, 9, 11, 0.42);
}

.language-switch button {
  min-width: 43px;
  padding: 7px 9px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--dim);
  font: 600 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  cursor: pointer;
}

.language-switch button.active {
  background: rgba(135, 212, 207, 0.1);
  color: var(--text);
  box-shadow: inset 0 0 0 1px rgba(135, 212, 207, 0.08);
}

.language-switch button:disabled {
  cursor: wait;
  opacity: 0.6;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(330px, 0.72fr);
  align-items: center;
  gap: clamp(48px, 8vw, 104px);
  min-height: 690px;
  padding: 82px 0 68px;
}

.hero-copy {
  position: relative;
}

.eyebrow,
.section-kicker {
  color: var(--foam);
  font: 600 10px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 24px;
}

.eyebrow span {
  width: 28px;
  height: 1px;
  background: currentColor;
  box-shadow: 0 0 9px rgba(135, 212, 207, 0.5);
}

.hero h1 {
  max-width: 720px;
  margin: 0;
  font-size: clamp(58px, 7vw, 94px);
  font-weight: 620;
  line-height: 0.94;
  letter-spacing: -0.06em;
}

.hero h1 span,
.hero h1 em {
  display: block;
}

.hero h1 em {
  margin-top: 13px;
  color: transparent;
  font-style: normal;
  background: linear-gradient(90deg, #e2b4c0 0%, #beaecd 47%, #91d4d0 100%);
  -webkit-background-clip: text;
  background-clip: text;
  text-shadow: 0 18px 48px rgba(89, 164, 163, 0.08);
}

.hero-description {
  max-width: 610px;
  margin: 30px 0 0;
  color: rgba(213, 216, 216, 0.72);
  font-size: clamp(16px, 1.6vw, 19px);
  line-height: 1.75;
}

.hero-actions {
  display: flex;
  align-items: stretch;
  gap: 12px;
  margin-top: 34px;
}

.primary-action,
.secondary-action {
  min-height: 56px;
  border-radius: 11px;
  font: 600 11px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.primary-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 28px;
  min-width: 182px;
  padding: 0 20px;
  color: #071012;
  background: linear-gradient(135deg, #f2efea, #d7e9e6);
  text-decoration: none;
  box-shadow: 0 16px 40px rgba(3, 8, 11, 0.28);
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.primary-action:hover {
  transform: translateY(-2px);
  box-shadow: 0 20px 48px rgba(3, 8, 11, 0.36);
}

.primary-action svg {
  width: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.7;
}

.secondary-action {
  display: grid;
  min-width: 245px;
  gap: 6px;
  padding: 10px 16px;
  border: 1px solid var(--line);
  background: rgba(4, 9, 11, 0.42);
  color: var(--text);
  text-align: left;
  cursor: pointer;
}

.secondary-action code {
  max-width: 230px;
  overflow: hidden;
  color: var(--dim);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.signal-row {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  margin-top: 28px;
  color: var(--dim);
  font: 500 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.signal-row span,
.node-live {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.signal,
.node-live i,
.terminal-state i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--dim);
}

.signal.online,
.node-live i,
.terminal-state i {
  background: var(--foam);
  box-shadow: 0 0 9px rgba(135, 212, 207, 0.74);
}

.node-card {
  position: relative;
  overflow: hidden;
  padding: 22px;
  border: 1px solid rgba(135, 212, 207, 0.13);
  border-radius: 18px;
  background: linear-gradient(145deg, rgba(9, 17, 20, 0.82), rgba(4, 8, 11, 0.72));
  box-shadow:
    0 34px 90px rgba(0, 0, 0, 0.32),
    inset 0 1px 0 rgba(255, 255, 255, 0.035);
  backdrop-filter: blur(18px);
}

.node-card::before {
  content: '';
  position: absolute;
  width: 220px;
  height: 220px;
  right: -110px;
  top: -120px;
  border-radius: 50%;
  background: rgba(73, 190, 185, 0.08);
  filter: blur(28px);
}

.node-card-head {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 20px;
}

.node-card-head p,
.node-endpoint span,
.node-metrics span,
.terminal-meta span {
  margin: 0;
  color: var(--dim);
  font: 500 8px/1.3 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.node-card-head strong {
  display: block;
  margin-top: 5px;
  font: 600 20px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.04em;
}

.node-live {
  color: var(--foam);
  font: 600 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.node-signal-map {
  position: relative;
  height: 150px;
  margin: 22px -2px 18px;
  overflow: hidden;
  border-top: 1px solid rgba(135, 212, 207, 0.08);
  border-bottom: 1px solid rgba(135, 212, 207, 0.08);
  background:
    linear-gradient(rgba(135, 212, 207, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(135, 212, 207, 0.025) 1px, transparent 1px);
  background-size: 34px 34px;
}

.map-line,
.map-flare,
.map-dot {
  position: absolute;
}

.map-line {
  height: 1px;
  transform-origin: left center;
  background: linear-gradient(90deg, transparent, rgba(135, 212, 207, 0.26), transparent);
}

.map-line-a {
  width: 62%;
  left: 12%;
  top: 62%;
  transform: rotate(-17deg);
}

.map-line-b {
  width: 42%;
  left: 46%;
  top: 34%;
  transform: rotate(22deg);
}

.map-flare {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: var(--foam);
  box-shadow: 0 0 14px rgba(135, 212, 207, 0.52);
}

.map-flare::before,
.map-flare::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  background: linear-gradient(90deg, transparent, rgba(135, 212, 207, 0.8), transparent);
}

.map-flare::before {
  width: 28px;
  height: 1px;
}

.map-flare::after {
  width: 1px;
  height: 28px;
  background: linear-gradient(180deg, transparent, rgba(135, 212, 207, 0.8), transparent);
}

.map-flare-a {
  right: 17%;
  top: 28%;
}

.map-flare-b {
  left: 34%;
  bottom: 25%;
  opacity: 0.72;
}

.map-dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: rgba(240, 239, 232, 0.72);
  box-shadow: 0 0 7px rgba(240, 239, 232, 0.16);
}

.map-dot-a { left: 14%; top: 28%; }
.map-dot-b { left: 57%; top: 20%; background: rgba(217, 165, 183, 0.72); }
.map-dot-c { right: 29%; bottom: 20%; }

.node-metrics {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.node-metrics div {
  display: grid;
  gap: 5px;
  padding: 12px;
  border: 1px solid rgba(218, 232, 232, 0.075);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.018);
}

.node-metrics strong {
  font: 600 11px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.node-endpoint {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  padding: 12px;
  border: 1px solid rgba(218, 232, 232, 0.075);
  border-radius: 10px;
  background: rgba(2, 6, 8, 0.36);
}

.node-endpoint code {
  overflow: hidden;
  color: var(--muted);
  font: 500 8px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-endpoint button,
.quick-command button {
  border: 0;
  background: transparent;
  color: var(--foam);
  font: 600 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  cursor: pointer;
  text-transform: uppercase;
}

.console-section {
  display: grid;
  grid-template-columns: minmax(220px, 0.42fr) minmax(0, 1.58fr);
  gap: clamp(34px, 6vw, 76px);
  align-items: start;
  padding: 42px 0 92px;
  border-top: 1px solid var(--line);
}

.console-intro {
  position: sticky;
  top: 32px;
  padding-top: 12px;
}

.console-intro h2 {
  margin: 18px 0 0;
  font-size: clamp(27px, 3.1vw, 42px);
  font-weight: 560;
  line-height: 1.1;
  letter-spacing: -0.035em;
}

.console-intro > p:not(.section-kicker) {
  margin: 18px 0 0;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.72;
}

.quick-command {
  display: grid;
  gap: 9px;
  margin-top: 28px;
  padding: 14px;
  border: 1px solid var(--line);
  border-radius: 11px;
  background: rgba(4, 9, 11, 0.38);
}

.quick-command span {
  color: var(--dim);
  font: 500 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.quick-command code {
  overflow-wrap: anywhere;
  color: var(--muted);
  font: 500 9px/1.55 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.quick-command button {
  justify-self: start;
  padding: 0;
}

.terminal-meta {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-top: 12px;
}

.terminal-meta div {
  display: grid;
  gap: 4px;
  padding: 10px;
  border: 1px solid rgba(218, 232, 232, 0.07);
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.014);
}

.terminal-meta strong {
  color: var(--muted);
  font: 600 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.terminal-wrap {
  position: relative;
}

.terminal-window {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(135, 212, 207, 0.14);
  border-radius: 16px;
  background: linear-gradient(150deg, rgba(12, 17, 21, 0.96), rgba(4, 7, 10, 0.96));
  box-shadow:
    0 34px 90px rgba(0, 0, 0, 0.38),
    inset 0 1px 0 rgba(255, 255, 255, 0.035);
}

.terminal-bar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  min-height: 46px;
  padding: 0 15px;
  border-bottom: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.018);
}

.window-controls {
  display: flex;
  gap: 7px;
}

.window-controls span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #4c4a57;
}

.window-controls span:first-child { background: var(--rose); }
.window-controls span:nth-child(2) { background: #e1bd7c; }
.window-controls span:last-child { background: var(--foam); }

.terminal-title,
.terminal-state {
  font: 500 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.terminal-title {
  color: var(--dim);
}

.terminal-state {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  color: var(--foam);
  text-transform: uppercase;
}

.terminal-body {
  min-height: 410px;
  padding: 30px 30px 34px;
  font: 500 12px/1.9 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.terminal-line,
.terminal-output,
.terminal-json {
  opacity: 0;
  animation: terminal-in 0.42s ease forwards;
}

.delay-1 { animation-delay: 0.15s; }
.delay-2 { animation-delay: 0.48s; }
.delay-3 { animation-delay: 0.82s; }
.delay-4 { animation-delay: 1.16s; }
.delay-5 { animation-delay: 1.5s; }

@keyframes terminal-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.terminal-line {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  color: #d4d6d5;
}

.prompt { color: var(--foam); }
.path { color: var(--iris); }

.terminal-output {
  display: grid;
  gap: 3px;
  margin: 18px 0 23px;
  padding: 15px 16px;
  border-left: 1px solid var(--pine);
  background: rgba(49, 120, 128, 0.055);
}

.terminal-output div {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
}

.terminal-output span {
  color: var(--dim);
}

.terminal-output strong {
  overflow: hidden;
  color: #c9cdcc;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-json {
  display: grid;
  margin: 13px 0 22px;
  color: var(--rose);
}

.terminal-json .indent {
  padding-left: 18px;
  color: #bfc3c3;
}

.cursor {
  width: 7px;
  height: 15px;
  margin-top: 4px;
  background: var(--foam);
  box-shadow: 0 0 10px rgba(135, 212, 207, 0.5);
  animation: blink 1s step-end infinite;
}

@keyframes blink { 50% { opacity: 0; } }

.capabilities {
  padding: 92px 0 100px;
  border-top: 1px solid var(--line);
}

.section-heading {
  display: grid;
  grid-template-columns: 0.52fr 1.48fr;
  gap: 46px;
  align-items: start;
  margin-bottom: 42px;
}

.section-heading p,
.manifesto p.section-kicker {
  margin: 0;
}

.section-heading > div > span,
.manifesto > div > span {
  display: block;
  margin-top: 8px;
  color: var(--dim);
  font: 500 9px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.section-heading h2 {
  max-width: 790px;
  margin: 0;
  font-size: clamp(35px, 4.5vw, 58px);
  font-weight: 560;
  line-height: 1.06;
  letter-spacing: -0.045em;
}

.capability-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.capability-grid article {
  min-height: 190px;
  padding: 24px;
  border: 1px solid rgba(218, 232, 232, 0.085);
  border-radius: 14px;
  background: linear-gradient(145deg, rgba(7, 13, 16, 0.55), rgba(3, 7, 9, 0.38));
  transition: transform 180ms ease, border-color 180ms ease, background 180ms ease;
}

.capability-grid article:hover {
  transform: translateY(-3px);
  border-color: rgba(135, 212, 207, 0.17);
  background: linear-gradient(145deg, rgba(8, 17, 20, 0.66), rgba(3, 8, 10, 0.46));
}

.article-topline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.article-index,
.article-tag {
  font: 500 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.article-index {
  color: var(--dim);
}

.article-tag {
  padding: 6px 8px;
  border: 1px solid var(--line);
  border-radius: 6px;
  color: var(--foam);
}

.capability-grid h3 {
  margin: 34px 0 10px;
  font-size: 19px;
  font-weight: 600;
}

.capability-grid p {
  max-width: 480px;
  margin: 0;
  color: var(--muted);
  font-size: 13px;
  line-height: 1.72;
}

.manifesto {
  display: grid;
  grid-template-columns: 0.38fr 1.35fr auto;
  gap: 44px;
  align-items: end;
  padding: 68px 0;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
}

.manifesto > p {
  margin: 0;
  color: #d3d5d2;
  font-size: clamp(23px, 2.7vw, 35px);
  line-height: 1.36;
  letter-spacing: -0.026em;
}

.manifesto a {
  color: var(--rose);
  font: 600 10px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-decoration: none;
  white-space: nowrap;
}

.manifesto a span {
  margin-left: 7px;
}

footer {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 30px 0 42px;
  color: var(--dim);
  font: 500 8px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

footer div {
  display: flex;
  gap: 18px;
}

footer a {
  color: var(--muted);
  text-decoration: none;
}

@media (max-width: 1020px) {
  .site-nav > a:not(.nav-cta) {
    display: none;
  }

  .hero {
    grid-template-columns: minmax(0, 1fr) minmax(300px, 0.78fr);
    gap: 42px;
  }

  .hero h1 {
    font-size: clamp(54px, 7vw, 72px);
  }

  .console-section {
    grid-template-columns: 1fr;
  }

  .console-intro {
    position: static;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 0.7fr);
    gap: 24px 34px;
    align-items: start;
  }

  .console-intro .section-kicker,
  .console-intro h2,
  .console-intro > p {
    grid-column: 1;
  }

  .quick-command,
  .terminal-meta {
    grid-column: 2;
  }

  .quick-command {
    grid-row: 1 / span 3;
    margin-top: 0;
  }
}

@media (max-width: 790px) {
  .page-frame {
    width: min(100% - 34px, 1160px);
  }

  .site-header {
    min-height: 78px;
  }

  .brand-copy small {
    display: none;
  }

  .site-nav {
    gap: 9px;
  }

  .site-nav .nav-cta {
    padding-inline: 11px;
    font-size: 9px;
  }

  .language-switch button {
    min-width: 37px;
    padding-inline: 7px;
  }

  .hero {
    grid-template-columns: 1fr;
    gap: 48px;
    min-height: auto;
    padding: 72px 0 66px;
  }

  .hero-copy {
    max-width: 660px;
  }

  .node-card {
    width: min(100%, 600px);
  }

  .console-intro {
    display: block;
  }

  .quick-command {
    margin-top: 24px;
  }

  .terminal-meta {
    margin-top: 12px;
  }

  .section-heading,
  .manifesto {
    grid-template-columns: 1fr;
  }

  .manifesto {
    align-items: start;
  }
}

@media (max-width: 560px) {
  .brand-copy {
    display: none;
  }

  .site-nav .nav-cta {
    display: none;
  }

  .hero {
    padding-top: 60px;
  }

  .hero h1 {
    font-size: clamp(48px, 15vw, 66px);
  }

  .hero-actions {
    flex-direction: column;
  }

  .primary-action,
  .secondary-action {
    width: 100%;
    box-sizing: border-box;
  }

  .secondary-action code {
    max-width: 100%;
  }

  .node-card {
    padding: 18px;
  }

  .node-signal-map {
    height: 126px;
  }

  .console-section {
    padding-bottom: 72px;
  }

  .terminal-body {
    min-height: 350px;
    padding: 23px 18px 28px;
    font-size: 10px;
  }

  .terminal-output div {
    grid-template-columns: 72px minmax(0, 1fr);
  }

  .terminal-title {
    display: none;
  }

  .terminal-bar {
    grid-template-columns: 1fr 1fr;
  }

  .capabilities {
    padding: 74px 0 78px;
  }

  .capability-grid {
    grid-template-columns: 1fr;
  }

  .capability-grid article {
    min-height: 170px;
  }

  .manifesto {
    padding: 54px 0;
  }

  footer,
  footer div {
    flex-direction: column;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
