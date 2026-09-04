<template>
  <section class="mb-6 overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
    <div class="grid gap-6 p-5 sm:p-6 xl:grid-cols-[1.15fr_.85fr] xl:items-center">
      <div>
        <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
          API Access
        </div>
        <h1 class="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
          {{ copy.title }}
        </h1>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-600 dark:text-dark-300">
          {{ copy.description }}
        </p>
        <div class="mt-5 flex flex-wrap gap-2">
          <router-link
            to="/model-plaza?embedded=1"
            class="inline-flex h-10 items-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
          >
            {{ copy.models }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex h-10 items-center rounded-xl border border-gray-200 bg-white px-4 text-sm font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100 dark:hover:bg-dark-700"
          >
            {{ copy.docs }}
          </a>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-200 bg-gray-950 p-4 shadow-inner dark:border-dark-700">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div class="min-w-0">
            <div class="text-[10px] font-semibold uppercase tracking-[0.14em] text-gray-500">Base URL</div>
            <code class="mt-1 block truncate text-xs font-medium text-gray-200">{{ apiBase }}</code>
          </div>
          <button
            type="button"
            class="shrink-0 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-[11px] font-semibold text-gray-300 transition hover:bg-white/10 hover:text-white"
            @click="copyBase"
          >
            {{ copied ? copy.copied : copy.copy }}
          </button>
        </div>
        <pre class="overflow-x-auto whitespace-pre-wrap break-words font-mono text-[11px] leading-5 text-gray-300"><code>curl {{ apiBase }}/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_NAME","messages":[{"role":"user","content":"Hello"}]}'</code></pre>
      </div>
    </div>

    <div class="grid border-t border-gray-100 bg-gray-50/70 sm:grid-cols-3 dark:border-dark-800 dark:bg-dark-950/50">
      <div v-for="(step, index) in copy.steps" :key="step" class="flex items-center gap-3 px-5 py-3.5 sm:px-6">
        <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-gray-900 text-[11px] font-semibold text-white dark:bg-white dark:text-gray-900">{{ index + 1 }}</span>
        <span class="text-xs font-medium text-gray-600 dark:text-dark-300">{{ step }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { locale } = useI18n()
const copied = ref(false)
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const apiBase = computed(() => {
  const configured = appStore.cachedPublicSettings?.api_base_url?.trim()
  return (configured || `${window.location.origin.replace(/\/$/, '')}/v1`).replace(/\/$/, '')
})

const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
)

const copy = computed(() =>
  isZh.value
    ? {
        title: '创建 Key 后，就可以直接调用统一 AI API',
        description: '下方仍然是完整的 API Key 管理能力。普通用户只需要创建一个 Key、选择可用模型，然后把 Base URL 和 Key 配到 OpenAI-compatible 客户端即可。',
        models: '查看模型与价格',
        docs: 'API 文档',
        copy: '复制',
        copied: '已复制',
        steps: ['创建 API Key', '选择模型', '复制 Base URL 开始调用']
      }
    : {
        title: 'Create a key and start calling the unified AI API',
        description: 'The full API key management tools remain below. For normal usage, create one key, choose an available model, then configure the Base URL and key in any OpenAI-compatible client.',
        models: 'Models & Pricing',
        docs: 'API Docs',
        copy: 'Copy',
        copied: 'Copied',
        steps: ['Create an API key', 'Choose a model', 'Copy the Base URL and call the API']
      }
)

async function copyBase() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 1600)
  } catch (error) {
    console.warn('Failed to copy API base URL:', error)
  }
}
</script>
