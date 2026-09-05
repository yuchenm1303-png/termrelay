<template>
  <div class="relay-auth-shell relative min-h-screen overflow-hidden bg-[#f6f7f9] text-slate-950 dark:bg-[#080b10] dark:text-white">
    <div class="auth-grid pointer-events-none absolute inset-0 opacity-70 dark:opacity-30"></div>
    <div class="pointer-events-none absolute -left-24 top-24 h-72 w-72 rounded-full bg-primary-400/10 blur-3xl"></div>
    <div class="pointer-events-none absolute -right-24 bottom-8 h-80 w-80 rounded-full bg-primary-500/10 blur-3xl"></div>

    <header class="relay-auth-topbar relative z-20 border-b border-slate-200/70 bg-white/75 backdrop-blur-xl dark:border-white/10 dark:bg-[#080b10]/75">
      <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 sm:px-6 lg:px-8">
        <router-link to="/home" class="flex min-w-0 items-center gap-3">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-9 w-9 rounded-full object-contain" />
          <span class="flex min-w-0 flex-col leading-tight">
            <strong class="truncate text-[15px] font-semibold tracking-tight">{{ siteName }}</strong>
            <small class="text-[10px] uppercase tracking-[0.16em] text-white/40">Account Access</small>
          </span>
        </router-link>

        <div class="flex items-center gap-2 sm:gap-3">
          <router-link
            to="/model-plaza"
            class="hidden text-sm font-medium text-slate-600 transition hover:text-slate-950 sm:inline dark:text-dark-300 dark:hover:text-white"
          >
            {{ copy.models }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden text-sm font-medium text-slate-600 transition hover:text-slate-950 sm:inline dark:text-dark-300 dark:hover:text-white"
          >
            {{ copy.docs }}
          </a>
          <LocaleSwitcher />
        </div>
      </div>
    </header>

    <main class="relative z-10 mx-auto grid min-h-[calc(100vh-64px)] max-w-[1180px] items-center gap-10 px-5 py-10 sm:px-6 lg:grid-cols-[1fr_420px] lg:px-8 lg:py-14">
      <section class="hidden max-w-2xl lg:block">
        <div class="mb-5 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-white/46">
          <span class="h-1.5 w-1.5 rounded-full bg-[#83efc7] shadow-[0_0_12px_rgba(131,239,199,.55)]"></span>
          {{ copy.badge }}
        </div>

        <h1 class="max-w-xl text-4xl font-semibold leading-[1.12] tracking-[-0.035em] text-slate-950 dark:text-white xl:text-5xl">
          {{ copy.title }}
        </h1>
        <p class="mt-5 max-w-xl text-base leading-7 text-slate-600 dark:text-dark-300">
          {{ siteSubtitle || copy.description }}
        </p>

        <div class="relay-auth-benefits mt-8 max-w-xl overflow-hidden rounded-2xl border border-slate-200 bg-white/85 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/[0.04]">
          <div class="border-b border-white/10 px-5 py-4">
            <div class="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/40">API Endpoint</div>
            <code class="mt-1 block truncate text-sm font-medium text-slate-800 dark:text-dark-100">{{ apiBase }}</code>
          </div>
          <div class="grid gap-0 sm:grid-cols-3">
            <div v-for="item in benefits" :key="item.title" class="border-b border-white/10 px-5 py-4 last:border-0 sm:border-b-0 sm:border-r sm:last:border-r-0">
              <div class="text-[10px] font-semibold uppercase tracking-[0.12em] text-white/40">{{ item.label }}</div>
              <div class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">{{ item.title }}</div>
              <div class="mt-1 text-xs leading-5 text-slate-500 dark:text-dark-400">{{ item.description }}</div>
            </div>
          </div>
        </div>
      </section>

      <section class="mx-auto w-full max-w-md lg:mx-0">
        <div class="mb-6 text-center lg:hidden">
          <div class="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full border border-white/15 bg-black/20 backdrop-blur">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-9 w-9 object-contain" />
          </div>
          <h1 class="text-xl font-semibold tracking-tight">{{ siteName }}</h1>
          <p class="mt-1 text-sm text-slate-500 dark:text-dark-400">{{ copy.mobileSubtitle }}</p>
        </div>

        <div class="relay-auth-card rounded-3xl border border-slate-200 bg-white p-6 shadow-[0_24px_80px_-30px_rgba(15,23,42,0.24)] sm:p-8 dark:border-white/10 dark:bg-[#10141b] dark:shadow-none">
          <div class="mb-6">
            <p class="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/40">Account Access</p>
          </div>
          <slot />
        </div>

        <div class="mt-5 text-center text-sm">
          <slot name="footer" />
        </div>

        <div class="mt-6 flex items-center justify-center gap-2 text-[10px] uppercase tracking-[0.14em] text-white/34">
          <span class="h-1.5 w-1.5 rounded-full bg-[#83efc7]"></span>
          <span>{{ currentYear }} · {{ siteName }}</span>
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

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TermRelay')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true,
  }),
)
const siteSubtitle = computed(() =>
  appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform',
)
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || `${window.location.origin.replace(/\/$/, '')}/v1`).replace(/\/$/, ''),
)

const copy = computed(() =>
  isZh.value
    ? {
        badge: 'AI RELAY · AUTHORIZED ACCESS',
        title: '一个 API Key，接入你的 AI 工作流',
        description: '登录后创建自己的 API Key，选择模型即可开始调用。',
        models: '模型与价格',
        docs: 'API 文档',
        mobileSubtitle: '登录或注册后即可创建 API Key',
      }
    : {
        badge: 'AI RELAY · AUTHORIZED ACCESS',
        title: 'One API key for your AI workflow',
        description: 'Sign in, create your API key, choose a model, and start calling the API.',
        models: 'Models & Pricing',
        docs: 'API Docs',
        mobileSubtitle: 'Sign in or create an account to get an API key',
      },
)

const benefits = computed(() =>
  isZh.value
    ? [
        { label: 'ENDPOINT', title: '统一入口', description: 'Base URL 与 API Key 保持不变。' },
        { label: 'MODELS', title: '多模型', description: '集中查看当前可用模型与价格。' },
        { label: 'ROUTING', title: '服务端调度', description: '接入、路由与故障切换由平台处理。' },
      ]
    : [
        { label: 'ENDPOINT', title: 'One endpoint', description: 'Keep the same Base URL and API key.' },
        { label: 'MODELS', title: 'Multiple models', description: 'Browse available models and pricing in one place.' },
        { label: 'ROUTING', title: 'Server-side routing', description: 'Access, routing, and failover stay behind the API.' },
      ],
)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.auth-grid {
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: linear-gradient(to bottom, black, transparent 88%);
}

:global(.dark) .auth-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.045) 1px, transparent 1px);
}
</style>
