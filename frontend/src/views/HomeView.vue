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

  <!-- Product landing page -->
  <div v-else class="min-h-screen bg-[#f7f8fa] text-slate-950 dark:bg-[#080b10] dark:text-white">
    <div class="relay-grid pointer-events-none fixed inset-x-0 top-0 h-[680px] opacity-70 dark:opacity-30"></div>

    <header class="relative z-20 border-b border-slate-200/70 bg-white/75 backdrop-blur-xl dark:border-white/10 dark:bg-[#080b10]/75">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 sm:px-6 lg:px-8">
        <router-link to="/home" class="flex min-w-0 items-center gap-3">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-9 w-9 rounded-xl object-contain" />
          <span class="truncate text-[15px] font-semibold tracking-tight">{{ siteName }}</span>
        </router-link>

        <div class="hidden items-center gap-7 md:flex">
          <router-link to="/model-plaza" class="text-sm font-medium text-slate-600 transition hover:text-slate-950 dark:text-dark-300 dark:hover:text-white">
            模型与价格
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm font-medium text-slate-600 transition hover:text-slate-950 dark:text-dark-300 dark:hover:text-white"
          >
            API 文档
          </a>
          <router-link to="/key-usage" class="text-sm font-medium text-slate-600 transition hover:text-slate-950 dark:text-dark-300 dark:hover:text-white">
            Key 用量
          </router-link>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher class="hidden sm:block" />
          <button
            class="flex h-9 w-9 items-center justify-center rounded-xl text-slate-500 transition hover:bg-slate-100 hover:text-slate-800 dark:text-dark-300 dark:hover:bg-white/10 dark:hover:text-white"
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
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex h-9 items-center justify-center rounded-xl bg-slate-950 px-4 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            {{ isAuthenticated ? '进入控制台' : '登录' }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10">
      <section class="mx-auto grid max-w-7xl gap-12 px-5 pb-20 pt-20 sm:px-6 lg:grid-cols-[1.08fr_0.92fr] lg:items-center lg:px-8 lg:pb-28 lg:pt-28">
        <div>
          <div class="mb-6 inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white/80 px-3 py-1.5 text-xs font-semibold text-slate-600 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5 dark:text-dark-200">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
            OpenAI-compatible unified API
          </div>
          <h1 class="max-w-3xl text-5xl font-semibold tracking-[-0.045em] text-slate-950 dark:text-white sm:text-6xl lg:text-7xl">
            一个 API，<br class="hidden sm:block" />连接主流 AI 模型
          </h1>
          <p class="mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-dark-300">
            用统一的接口访问 GPT、Claude、Gemini 等模型。只管理一个 Base URL 和一个 API Key，路由、账号调度与用量统计由 {{ siteName }} 在服务端完成。
          </p>

          <div class="mt-8 flex flex-wrap gap-3">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/register'"
              class="inline-flex h-12 items-center justify-center gap-2 rounded-xl bg-primary-600 px-5 text-sm font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:-translate-y-0.5 hover:bg-primary-700"
            >
              {{ isAuthenticated ? '进入控制台' : '免费开始使用' }}
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
            </router-link>
            <router-link
              to="/model-plaza"
              class="inline-flex h-12 items-center justify-center rounded-xl border border-slate-200 bg-white/80 px-5 text-sm font-semibold text-slate-700 shadow-sm backdrop-blur transition hover:bg-white dark:border-white/10 dark:bg-white/5 dark:text-dark-100 dark:hover:bg-white/10"
            >
              查看可用模型
            </router-link>
          </div>

          <div class="mt-9 flex flex-wrap items-center gap-x-6 gap-y-3 text-sm text-slate-500 dark:text-dark-400">
            <span class="flex items-center gap-2"><span class="h-1.5 w-1.5 rounded-full bg-primary-500"></span>统一 API Key</span>
            <span class="flex items-center gap-2"><span class="h-1.5 w-1.5 rounded-full bg-primary-500"></span>OpenAI SDK 兼容</span>
            <span class="flex items-center gap-2"><span class="h-1.5 w-1.5 rounded-full bg-primary-500"></span>服务端智能路由</span>
          </div>
        </div>

        <div class="terminal-container lg:justify-self-end">
          <div class="overflow-hidden rounded-[22px] border border-slate-200/80 bg-[#0d1117] shadow-2xl shadow-slate-900/15 dark:border-white/10">
            <div class="flex items-center justify-between border-b border-white/10 px-5 py-3.5">
              <div class="flex gap-1.5">
                <span class="h-2.5 w-2.5 rounded-full bg-[#ff5f57]"></span>
                <span class="h-2.5 w-2.5 rounded-full bg-[#febc2e]"></span>
                <span class="h-2.5 w-2.5 rounded-full bg-[#28c840]"></span>
              </div>
              <span class="font-mono text-[11px] text-slate-500">quickstart.sh</span>
              <span class="w-10"></span>
            </div>
            <div class="p-5 sm:p-6">
              <div class="mb-5 flex items-center justify-between rounded-xl border border-white/10 bg-white/[0.035] px-3.5 py-3">
                <div>
                  <p class="text-[10px] uppercase tracking-[0.18em] text-slate-500">API Base</p>
                  <p class="mt-1 break-all font-mono text-xs text-slate-200">{{ apiBase }}</p>
                </div>
                <span class="ml-3 rounded-lg bg-emerald-400/10 px-2 py-1 text-[10px] font-semibold text-emerald-300">READY</span>
              </div>
              <pre class="overflow-x-auto whitespace-pre-wrap break-words font-mono text-[12px] leading-6 text-slate-300"><code><span class="text-sky-300">curl</span> {{ apiBase }}/chat/completions \
  -H <span class="text-amber-200">"Authorization: Bearer sk-..."</span> \
  -H <span class="text-amber-200">"Content-Type: application/json"</span> \
  -d <span class="text-emerald-300">'{
    "model": "model-name",
    "messages": [{"role":"user","content":"Hello"}]
  }'</span></code></pre>
              <div class="mt-6 grid grid-cols-3 gap-2 border-t border-white/10 pt-5 text-center">
                <div>
                  <p class="text-lg font-semibold text-white">1</p>
                  <p class="text-[10px] uppercase tracking-wider text-slate-500">Base URL</p>
                </div>
                <div>
                  <p class="text-lg font-semibold text-white">1</p>
                  <p class="text-[10px] uppercase tracking-wider text-slate-500">API Key</p>
                </div>
                <div>
                  <p class="text-lg font-semibold text-white">N</p>
                  <p class="text-[10px] uppercase tracking-wider text-slate-500">Models</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="border-y border-slate-200/80 bg-white/70 dark:border-white/10 dark:bg-white/[0.025]">
        <div class="mx-auto max-w-7xl px-5 py-9 sm:px-6 lg:px-8">
          <p class="mb-6 text-center text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">One endpoint · multiple providers</p>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <div v-for="provider in providers" :key="provider.name" class="flex items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-3.5 text-sm font-semibold text-slate-700 shadow-sm dark:border-white/10 dark:bg-white/5 dark:text-dark-100">
              <span class="flex h-7 w-7 items-center justify-center rounded-lg text-xs font-bold" :class="provider.badgeClass">{{ provider.mark }}</span>
              {{ provider.name }}
            </div>
          </div>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8 lg:py-24">
        <div class="max-w-2xl">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary-600 dark:text-primary-400">Designed for API users</p>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">从拿到 Key 到第一次调用，路径应该足够短。</h2>
          <p class="mt-4 text-base leading-7 text-slate-600 dark:text-dark-300">用户只需要处理自己的密钥、模型和用量；上游账号、调度策略和路由细节留在服务端。</p>
        </div>
        <div class="mt-10 grid gap-4 md:grid-cols-3">
          <article v-for="feature in features" :key="feature.title" class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-white/10 dark:bg-white/[0.035]">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-slate-50 text-sm font-semibold text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-white">{{ feature.index }}</div>
            <h3 class="mt-5 text-lg font-semibold">{{ feature.title }}</h3>
            <p class="mt-2 text-sm leading-6 text-slate-600 dark:text-dark-300">{{ feature.description }}</p>
          </article>
        </div>
      </section>

      <section class="border-t border-slate-200/80 px-5 py-14 dark:border-white/10">
        <div class="mx-auto flex max-w-7xl flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-xl font-semibold">准备好开始调用了吗？</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-dark-400">创建账户、生成 API Key，然后选择模型。</p>
          </div>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/register'"
            class="inline-flex h-11 shrink-0 items-center justify-center rounded-xl bg-slate-950 px-5 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            {{ isAuthenticated ? '进入控制台' : '创建账户' }}
          </router-link>
        </div>
      </section>
    </main>

    <footer class="border-t border-slate-200/80 px-5 py-6 text-sm text-slate-500 dark:border-white/10 dark:text-dark-400">
      <div class="mx-auto flex max-w-7xl flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <span>&copy; {{ currentYear }} {{ siteName }}</span>
        <span>OpenAI-compatible API gateway</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || `${window.location.origin.replace(/\/$/, '')}/v1`)
    .trim()
    .replace(/\/$/, ''),
)

const isDark = ref(false)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())

const providers = [
  { name: 'OpenAI', mark: 'O', badgeClass: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-300' },
  { name: 'Anthropic', mark: 'A', badgeClass: 'bg-amber-500/10 text-amber-600 dark:text-amber-300' },
  { name: 'Gemini', mark: 'G', badgeClass: 'bg-blue-500/10 text-blue-600 dark:text-blue-300' },
  { name: 'More', mark: '+', badgeClass: 'bg-violet-500/10 text-violet-600 dark:text-violet-300' },
]

const features = [
  { index: '01', title: '创建自己的 API Key', description: '每个用户只管理自己的访问密钥，可按需要创建和停用。' },
  { index: '02', title: '查看模型与价格', description: '在模型目录里找到可调用模型，不需要理解上游账号结构。' },
  { index: '03', title: '追踪调用与费用', description: '在控制台查看 Token、费用和调用记录，定位问题更直接。' },
]

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
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
.relay-grid {
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: linear-gradient(to bottom, black, transparent 95%);
}

:global(.dark) .relay-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.045) 1px, transparent 1px);
}

.terminal-container {
  width: min(100%, 560px);
}
</style>
