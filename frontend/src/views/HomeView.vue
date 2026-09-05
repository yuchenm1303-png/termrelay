<template>
  <!-- Admin-configured custom home always has highest priority. -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact mode remains available for operators who explicitly enable it. -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-white text-slate-950 dark:bg-dark-950 dark:text-white"
  >
    <header class="border-b border-slate-200/80 px-4 py-4 sm:px-6 dark:border-dark-800">
      <nav class="mx-auto flex max-w-6xl items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-9 w-9 rounded-xl object-contain" />
          <span class="truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex items-center gap-2">
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 items-center justify-center rounded-xl bg-slate-950 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
          <LocaleSwitcher />
          <button
            class="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-50 dark:border-dark-700 dark:text-dark-300 dark:hover:bg-dark-800"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <svg v-if="isDark" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="12" r="4" />
              <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
            </svg>
            <svg v-else class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79Z" />
            </svg>
          </button>
        </div>
      </nav>
    </header>

    <main class="flex flex-1 items-center justify-center px-5 py-20">
      <div class="max-w-3xl text-center">
        <div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-10 w-10 object-contain" />
        </div>
        <h1 class="text-4xl font-semibold tracking-tight md:text-5xl">{{ siteName }}</h1>
        <p class="mx-auto mt-5 max-w-2xl text-base leading-7 text-slate-600 dark:text-dark-300">{{ siteSubtitle }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-11 items-center justify-center rounded-xl bg-primary-600 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-primary-700"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
        </router-link>
      </div>
    </main>

    <footer class="border-t border-slate-200 px-5 py-5 text-center text-sm text-slate-500 dark:border-dark-800 dark:text-dark-400">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Standalone product portal: same visual language as Smirel Downloads / Service Monitor. -->
  <div v-else class="relay-portal min-h-screen">
    <div class="relay-portal-cosmos" aria-hidden="true">
      <div class="relay-portal-glow relay-portal-glow-cyan"></div>
      <div class="relay-portal-glow relay-portal-glow-violet"></div>
      <span v-for="star in stars" :key="star.id" class="relay-star" :class="star.className" :style="star.style"></span>
    </div>

    <main class="relay-public-shell">
      <header class="relay-public-topbar relay-fade">
        <router-link to="/home" class="relay-brand" aria-label="Smirel API 首页">
          <span class="relay-brand-mark">
            <img :src="siteLogo || '/logo.svg'" alt="" />
          </span>
          <span class="relay-brand-copy">
            <strong>{{ siteName }}</strong>
            <small>AI Relay</small>
          </span>
        </router-link>

        <div class="relay-topbar-right">
          <nav class="relay-public-nav" aria-label="页面导航">
            <router-link to="/model-plaza">模型与价格</router-link>
            <router-link to="/key-usage">Key 用量</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">API 文档</a>
          </nav>
          <LocaleSwitcher />
          <div class="relay-service-status"><i></i><span>Gateway</span></div>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="relay-topbar-action">
            {{ isAuthenticated ? '控制台' : '登录' }}
          </router-link>
        </div>
      </header>

      <section class="relay-portal-title relay-fade">
        <div>
          <p class="relay-eyebrow">AI RELAY · UNIFIED ACCESS</p>
          <h1>{{ siteName }}</h1>
          <p>一个 Base URL，一个 API Key，统一访问 GPT、Claude、Gemini 等主流模型。</p>
        </div>
        <div class="relay-title-meta" aria-label="服务状态">
          <span class="relay-title-chip live"><i></i>Online</span>
          <span class="relay-title-chip">OpenAI Compatible</span>
        </div>
      </section>

      <section class="relay-dashboard relay-fade">
        <article class="relay-card relay-gateway-card">
          <div class="relay-product-row">
            <div class="relay-product-icon"><img :src="siteLogo || '/logo.svg'" alt="" /></div>
            <div class="relay-product-copy">
              <strong>Unified API Gateway</strong>
              <span>Multi-model AI access</span>
            </div>
            <span class="relay-channel">STABLE CHANNEL</span>
          </div>

          <div class="relay-card-head">
            <div>
              <p class="relay-kicker">API ENDPOINT</p>
              <h2>统一接口</h2>
            </div>
            <span class="relay-mini-badge">V1 · HTTPS</span>
          </div>

          <div class="relay-endpoint-block">
            <code>{{ apiBase }}</code>
            <span>READY</span>
          </div>

          <p class="relay-summary">
            客户端只需要模型名、Base URL 和自己的 API Key。模型接入、路由和故障切换在服务端完成。
          </p>

          <div class="relay-state-row" aria-label="接口能力">
            <span class="live"><i></i>Gateway online</span>
            <span><i></i>OpenAI SDK compatible</span>
            <span><i></i>Server-side routing</span>
          </div>

          <div class="relay-meta-grid">
            <div><span>ENDPOINT</span><strong>1 Base URL</strong></div>
            <div><span>ACCESS</span><strong>1 API Key</strong></div>
            <div><span>MODELS</span><strong>Multiple</strong></div>
          </div>

          <div class="terminal-container relay-terminal">
            <div class="relay-terminal-head">
              <span>QUICK START</span>
              <small>chat/completions</small>
            </div>
            <pre><code><span class="relay-code-command">curl</span> {{ apiBase }}/chat/completions \
  -H <span class="relay-code-string">"Authorization: Bearer sk-..."</span> \
  -H <span class="relay-code-string">"Content-Type: application/json"</span> \
  -d <span class="relay-code-json">'{"model":"model-name","messages":[...]}'</span></code></pre>
          </div>

          <div class="relay-primary-actions">
            <router-link :to="isAuthenticated ? '/keys' : '/register'" class="relay-action relay-action-primary">
              <span>
                <strong>{{ isAuthenticated ? '管理 API Key' : '免费开始使用' }}</strong>
                <small>{{ isAuthenticated ? '查看和创建访问密钥' : '创建账户后生成访问密钥' }}</small>
              </span>
              <b>→</b>
            </router-link>
            <router-link to="/model-plaza" class="relay-action relay-action-secondary">
              <span><strong>查看可用模型</strong><small>模型与价格</small></span>
              <b>↗</b>
            </router-link>
          </div>
        </article>

        <aside class="relay-card relay-access-card">
          <div class="relay-card-head">
            <div>
              <p class="relay-kicker">ACCOUNT ACCESS</p>
              <h2>{{ isAuthenticated ? '账户已连接' : '账户入口' }}</h2>
            </div>
            <span class="relay-mini-badge">ACCESS</span>
          </div>

          <p class="relay-access-intro">
            {{ isAuthenticated ? '你的账户已经可以进入控制台管理密钥、用量与模型。' : '登录或创建账户后，即可生成自己的 API Key 并开始调用。' }}
          </p>

          <div v-if="isAuthenticated" class="relay-signed-panel">
            <div class="relay-signed-state">
              <span class="relay-account-avatar">✓</span>
              <div><p class="relay-kicker">SIGNED IN</p><strong>Authenticated session</strong></div>
            </div>
            <router-link :to="dashboardPath" class="relay-account-button">进入控制台</router-link>
            <router-link to="/keys" class="relay-account-link">API 密钥</router-link>
          </div>

          <div v-else class="relay-login-actions">
            <router-link to="/login" class="relay-account-button">登录账户</router-link>
            <router-link to="/register" class="relay-account-link">创建账户</router-link>
          </div>

          <div class="relay-access-flow" aria-label="使用流程">
            <span>ACCESS FLOW</span>
            <div>
              <p><b>01</b><strong>账户登录</strong></p>
              <p><b>02</b><strong>生成 API Key</strong></p>
              <p><b>03</b><strong>开始调用</strong></p>
            </div>
          </div>

          <div class="relay-provider-list">
            <span>PROVIDERS</span>
            <div>
              <p v-for="provider in providers" :key="provider.name"><i :class="provider.className"></i>{{ provider.name }}</p>
            </div>
          </div>

          <div class="relay-account-footer"><span>Authorized API access</span><span>relay.smirel.com</span></div>
        </aside>
      </section>

      <section class="relay-utility-grid relay-fade">
        <router-link to="/model-plaza" class="relay-card relay-utility-card">
          <span class="relay-utility-icon">⌘</span><span class="relay-utility-overline">MODELS</span><h3>模型与价格</h3><p>查看当前可调用模型与计费信息。</p>
        </router-link>
        <router-link :to="isAuthenticated ? '/keys' : '/login'" class="relay-card relay-utility-card">
          <span class="relay-utility-icon">◇</span><span class="relay-utility-overline">ACCESS</span><h3>API 密钥</h3><p>创建、停用和管理项目访问密钥。</p>
        </router-link>
        <router-link :to="isAuthenticated ? '/usage' : '/key-usage'" class="relay-card relay-utility-card">
          <span class="relay-utility-icon">⌁</span><span class="relay-utility-overline">USAGE</span><h3>调用记录</h3><p>查看 Token、费用与模型调用情况。</p>
        </router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="relay-card relay-utility-card">
          <span class="relay-utility-icon">?</span><span class="relay-utility-overline">DOCS</span><h3>API 文档</h3><p>查看接口格式、鉴权和接入说明。</p>
        </a>
        <router-link v-else :to="isAuthenticated ? dashboardPath : '/login'" class="relay-card relay-utility-card">
          <span class="relay-utility-icon">▣</span><span class="relay-utility-overline">CONSOLE</span><h3>控制台</h3><p>进入账户控制台管理你的 API 服务。</p>
        </router-link>
      </section>

      <footer class="relay-public-footer relay-fade">
        <span>© {{ currentYear }} {{ siteName }}</span>
        <span class="relay-footer-status"><i></i>Unified AI API Gateway</span>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TermRelay')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
const siteSubtitle = computed(() =>
  appStore.cachedPublicSettings?.site_subtitle || 'Unified AI API Gateway',
)
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(
  () => appStore.cachedPublicSettings?.compact_home_enabled === true,
)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')
    .trim()
    .replace(/\/$/, ''),
)

const providers = [
  { name: 'OpenAI', className: 'is-green' },
  { name: 'Anthropic', className: 'is-amber' },
  { name: 'Gemini', className: 'is-cyan' },
  { name: 'More', className: 'is-violet' },
]

const stars = Array.from({ length: 18 }, (_, index) => {
  const x = (13 + index * 37) % 96
  const y = (17 + index * 29) % 92
  const size = 1 + (index % 4) * 0.45
  const tone = index % 5 === 0 ? 'rose' : index % 3 === 0 ? 'ivory' : 'cyan'
  return {
    id: index,
    className: `relay-star-${tone}`,
    style: {
      left: `${x}%`,
      top: `${y}%`,
      width: `${size}px`,
      height: `${size}px`,
      animationDelay: `${-(index % 7) * 0.7}s`,
    },
  }
})

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  if (document.documentElement.classList.contains('relay-standalone')) {
    isDark.value = true
    document.documentElement.classList.add('dark')
    return
  }

  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.relay-portal {
  position: relative;
  overflow: hidden;
  color: var(--relay-text, #fff);
  background: transparent;
}

.relay-portal-cosmos {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(7,14,33,.20), rgba(5,9,25,.35));
}

.relay-portal-cosmos::after {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 45%, transparent 34%, rgba(14,20,48,.20) 100%);
}

.relay-portal-glow {
  position: absolute;
  width: 58vw;
  height: 58vw;
  border-radius: 50%;
  filter: blur(32px);
  opacity: .42;
}

.relay-portal-glow-cyan {
  right: -15vw;
  top: 2vh;
  background: radial-gradient(circle, rgba(89,214,217,.22), transparent 66%);
}

.relay-portal-glow-violet {
  left: -18vw;
  bottom: -18vw;
  background: radial-gradient(circle, rgba(191,124,215,.19), transparent 65%);
}

.relay-star {
  position: absolute;
  z-index: 2;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 8px currentColor, 0 0 22px currentColor;
  animation: relay-home-twinkle 6.8s ease-in-out infinite;
}

.relay-star-cyan { color: rgb(205,235,242); }
.relay-star-ivory { color: rgb(255,240,208); }
.relay-star-rose { color: rgb(241,211,225); }

@keyframes relay-home-twinkle {
  0%, 100% { opacity: .42; transform: scale(.85); }
  36% { opacity: .95; transform: scale(1.45); }
  68% { opacity: .55; transform: scale(.95); }
}

.relay-public-shell {
  position: relative;
  z-index: 2;
  width: min(1180px, calc(100vw - 44px));
  min-height: 100vh;
  margin: 0 auto;
  padding: 24px 0 28px;
}

.relay-public-topbar {
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.relay-brand,
.relay-topbar-right,
.relay-public-nav,
.relay-service-status,
.relay-title-meta,
.relay-product-row,
.relay-card-head,
.relay-state-row,
.relay-primary-actions,
.relay-signed-state,
.relay-public-footer,
.relay-footer-status {
  display: flex;
  align-items: center;
}

.relay-brand { gap: 11px; }
.relay-brand-mark {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,.20);
  background: rgba(255,255,255,.13);
  backdrop-filter: blur(10px);
}
.relay-brand-mark img { width: 25px; height: 25px; object-fit: contain; }
.relay-brand-copy { display: flex; flex-direction: column; gap: 1px; }
.relay-brand-copy strong { font-size: 1rem; font-weight: 680; }
.relay-brand-copy small { color: rgba(255,255,255,.42); font-size: .62rem; letter-spacing: .14em; text-transform: uppercase; }

.relay-topbar-right { gap: 15px; }
.relay-public-nav { gap: 6px; }
.relay-public-nav a,
.relay-topbar-action {
  padding: 8px 10px;
  border-radius: 6px;
  color: rgba(255,255,255,.62);
  font-size: .76rem;
  font-weight: 560;
  transition: color .18s, background .18s;
}
.relay-public-nav a:hover { color: #fff; background: rgba(255,255,255,.07); }
.relay-topbar-action { color: #fff; border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.10); }
.relay-topbar-action:hover { background: rgba(255,255,255,.17); }

.relay-service-status { gap: 7px; color: rgba(255,255,255,.60); font-size: .7rem; }
.relay-service-status i,
.relay-title-chip.live i,
.relay-state-row .live i,
.relay-footer-status i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #83efc7;
  box-shadow: 0 0 12px rgba(131,239,199,.55);
}

.relay-portal-title {
  margin: 35px 0 22px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 28px;
}
.relay-eyebrow,
.relay-kicker,
.relay-utility-overline {
  color: rgba(255,255,255,.45);
  font-size: .66rem;
  font-weight: 650;
  letter-spacing: .15em;
  text-transform: uppercase;
}
.relay-portal-title h1 {
  margin-top: 6px;
  font-size: clamp(2.3rem, 4.2vw, 3.6rem);
  font-weight: 620;
  letter-spacing: -.045em;
}
.relay-portal-title > div:first-child > p:last-child {
  margin-top: 9px;
  max-width: 680px;
  color: rgba(255,255,255,.68);
  font-size: .88rem;
  line-height: 1.75;
}
.relay-title-meta { gap: 7px; flex-wrap: wrap; justify-content: flex-end; }
.relay-title-chip,
.relay-mini-badge,
.relay-channel {
  padding: 6px 9px;
  border-radius: 999px;
  border: 1px solid rgba(255,255,255,.08);
  color: rgba(255,255,255,.62);
  background: rgba(255,255,255,.07);
  font-size: .61rem;
  letter-spacing: .07em;
  white-space: nowrap;
}
.relay-title-chip.live { display: inline-flex; align-items: center; gap: 6px; color: rgba(216,255,240,.85); }

.relay-dashboard {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(320px, .82fr);
  gap: 18px;
  align-items: stretch;
}
.relay-card {
  border: 1px solid rgba(255,255,255,.12);
  border-radius: 8px;
  background: rgba(0,0,0,.25);
  backdrop-filter: blur(14px) saturate(115%);
  box-shadow: 0 18px 50px rgba(5,10,28,.12);
  transition: transform .25s ease, background .25s ease, border-color .25s ease;
}
.relay-card:hover { border-color: rgba(255,255,255,.16); background: rgba(0,0,0,.30); }
.relay-gateway-card { padding: 28px; }
.relay-access-card { padding: 26px; }

.relay-product-row { gap: 10px; padding-bottom: 20px; border-bottom: 1px solid rgba(255,255,255,.08); }
.relay-product-icon {
  width: 40px; height: 40px; display: grid; place-items: center; flex: 0 0 40px;
  border-radius: 7px; border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.09);
}
.relay-product-icon img { width: 26px; height: 26px; object-fit: contain; }
.relay-product-copy { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.relay-product-copy strong { font-size: .88rem; font-weight: 650; }
.relay-product-copy span { color: rgba(255,255,255,.42); font-size: .66rem; }
.relay-channel { margin-left: auto; }

.relay-card-head { justify-content: space-between; gap: 18px; margin-top: 22px; }
.relay-card-head h2 { margin-top: 5px; font-size: 1.35rem; font-weight: 620; }

.relay-endpoint-block {
  margin-top: 22px;
  min-height: 62px;
  padding: 13px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 15px;
  border: 1px solid rgba(255,255,255,.08);
  border-radius: 7px;
  background: rgba(0,0,0,.16);
}
.relay-endpoint-block code { overflow: hidden; text-overflow: ellipsis; color: rgba(255,255,255,.90); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: .82rem; }
.relay-endpoint-block span { color: #bdf6df; font-size: .61rem; font-weight: 700; letter-spacing: .08em; }
.relay-summary { margin-top: 18px; color: rgba(255,255,255,.67); font-size: .8rem; line-height: 1.75; }

.relay-state-row { margin-top: 17px; gap: 14px; flex-wrap: wrap; color: rgba(255,255,255,.48); font-size: .65rem; }
.relay-state-row span { display: inline-flex; align-items: center; gap: 6px; }
.relay-state-row i { width: 5px; height: 5px; border-radius: 50%; background: rgba(255,255,255,.28); }

.relay-meta-grid { margin-top: 19px; display: grid; grid-template-columns: repeat(3, minmax(0,1fr)); gap: 7px; }
.relay-meta-grid > div { padding: 10px 11px; border-radius: 6px; background: rgba(0,0,0,.14); }
.relay-meta-grid span { display: block; color: rgba(255,255,255,.34); font-size: .56rem; letter-spacing: .10em; }
.relay-meta-grid strong { display: block; margin-top: 5px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: rgba(255,255,255,.78); font-size: .7rem; font-weight: 600; }

.relay-terminal { width: 100%; margin-top: 18px; overflow: hidden; border: 1px solid rgba(255,255,255,.08); border-radius: 7px; background: rgba(4,8,18,.38); }
.relay-terminal-head { padding: 9px 12px; display: flex; justify-content: space-between; gap: 15px; border-bottom: 1px solid rgba(255,255,255,.07); color: rgba(255,255,255,.36); font-size: .58rem; letter-spacing: .12em; }
.relay-terminal-head small { letter-spacing: 0; text-transform: none; }
.relay-terminal pre { padding: 14px; overflow-x: auto; white-space: pre-wrap; word-break: break-word; color: rgba(255,255,255,.67); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: .68rem; line-height: 1.8; }
.relay-code-command { color: #bfeff3; }
.relay-code-string { color: #ffe9a9; }
.relay-code-json { color: #aeeed2; }

.relay-primary-actions { gap: 8px; margin-top: 18px; }
.relay-action { min-height: 54px; padding: 10px 13px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-radius: 7px; transition: transform .2s, background .2s; }
.relay-action:hover { transform: translateY(-1px); }
.relay-action span { display: flex; flex-direction: column; gap: 2px; }
.relay-action strong { font-size: .78rem; font-weight: 630; }
.relay-action small { color: rgba(255,255,255,.42); font-size: .60rem; }
.relay-action b { font-size: 1.05rem; font-weight: 400; }
.relay-action-primary { flex: 1.35; border: 1px solid rgba(255,255,255,.14); background: rgba(255,255,255,.13); }
.relay-action-primary:hover { background: rgba(255,255,255,.19); }
.relay-action-secondary { flex: 1; border: 1px solid rgba(255,255,255,.08); background: rgba(0,0,0,.12); }
.relay-action-secondary:hover { background: rgba(255,255,255,.08); }

.relay-access-intro { margin-top: 17px; color: rgba(255,255,255,.60); font-size: .76rem; line-height: 1.7; }
.relay-login-actions,
.relay-signed-panel { margin-top: 24px; display: grid; gap: 9px; }
.relay-signed-state { gap: 11px; margin-bottom: 4px; }
.relay-account-avatar { width: 39px; height: 39px; display: grid; place-items: center; border-radius: 50%; border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.09); color: #bdf6df; }
.relay-signed-state strong { display: block; margin-top: 4px; color: rgba(255,255,255,.78); font-size: .76rem; }
.relay-account-button,
.relay-account-link { min-height: 43px; display: flex; align-items: center; justify-content: center; border-radius: 7px; font-size: .75rem; font-weight: 620; }
.relay-account-button { border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.12); }
.relay-account-button:hover { background: rgba(255,255,255,.19); }
.relay-account-link { color: rgba(255,255,255,.58); background: rgba(0,0,0,.12); }
.relay-account-link:hover { color: #fff; background: rgba(255,255,255,.07); }

.relay-access-flow,
.relay-provider-list { margin-top: 24px; padding-top: 17px; border-top: 1px solid rgba(255,255,255,.08); }
.relay-access-flow > span,
.relay-provider-list > span { color: rgba(255,255,255,.34); font-size: .57rem; letter-spacing: .13em; }
.relay-access-flow > div { margin-top: 11px; display: grid; gap: 7px; }
.relay-access-flow p { display: flex; align-items: center; gap: 10px; color: rgba(255,255,255,.63); font-size: .68rem; }
.relay-access-flow b { width: 22px; color: rgba(255,255,255,.32); font-size: .58rem; }
.relay-access-flow strong { font-weight: 570; }
.relay-provider-list > div { margin-top: 11px; display: grid; grid-template-columns: repeat(2, minmax(0,1fr)); gap: 7px; }
.relay-provider-list p { display: flex; align-items: center; gap: 7px; color: rgba(255,255,255,.56); font-size: .66rem; }
.relay-provider-list i { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,.4); }
.relay-provider-list .is-green { background: #83efc7; }
.relay-provider-list .is-amber { background: #f2c77e; }
.relay-provider-list .is-cyan { background: #bfeff3; }
.relay-provider-list .is-violet { background: #c5a8e8; }
.relay-account-footer { margin-top: 23px; padding-top: 14px; border-top: 1px solid rgba(255,255,255,.08); display: flex; justify-content: space-between; gap: 12px; color: rgba(255,255,255,.34); font-size: .59rem; }

.relay-utility-grid { display: grid; grid-template-columns: repeat(4, minmax(0,1fr)); gap: 18px; margin-top: 18px; }
.relay-utility-card { min-height: 142px; padding: 20px; }
.relay-utility-card:hover { transform: translateY(-2px) scale(1.008); background: rgba(0,0,0,.36); }
.relay-utility-icon { display: block; color: rgba(255,255,255,.78); font-size: 1.25rem; line-height: 1; }
.relay-utility-overline { display: block; margin-top: 17px; font-size: .55rem; }
.relay-utility-card h3 { margin-top: 5px; font-size: .89rem; font-weight: 620; }
.relay-utility-card p { margin-top: 6px; color: rgba(255,255,255,.48); font-size: .66rem; line-height: 1.55; }

.relay-public-footer { justify-content: space-between; gap: 16px; margin-top: 18px; color: rgba(255,255,255,.32); font-size: .60rem; }
.relay-footer-status { gap: 7px; }
.relay-fade { animation: relay-fade-in .55s both; }
@keyframes relay-fade-in { from { opacity: 0; transform: translateY(7px); filter: blur(4px); } to { opacity: 1; transform: translateY(0); filter: blur(0); } }

@media (max-width: 900px) {
  .relay-public-shell { width: min(100% - 28px, 760px); padding-top: 15px; }
  .relay-public-nav { display: none; }
  .relay-dashboard { grid-template-columns: 1fr; }
  .relay-portal-title { margin-top: 26px; align-items: flex-start; flex-direction: column; }
  .relay-title-meta { justify-content: flex-start; }
  .relay-utility-grid { grid-template-columns: repeat(2, minmax(0,1fr)); }
}

@media (max-width: 560px) {
  .relay-service-status { display: none; }
  .relay-topbar-right { gap: 7px; }
  .relay-gateway-card, .relay-access-card { padding: 20px; }
  .relay-meta-grid { grid-template-columns: 1fr; }
  .relay-primary-actions { align-items: stretch; flex-direction: column; }
  .relay-utility-grid { grid-template-columns: 1fr; gap: 10px; }
  .relay-utility-card { min-height: 116px; }
  .relay-public-footer { align-items: flex-start; flex-direction: column; }
  .relay-product-row { align-items: flex-start; flex-wrap: wrap; }
  .relay-channel { margin-left: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .relay-star, .relay-fade { animation: none; }
}
</style>
