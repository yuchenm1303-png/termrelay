<template>
  <div class="space-y-5">
    <section
      v-if="!embedded"
      class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900"
      :class="embedded ? '' : 'mt-1'"
    >
      <div class="grid gap-5 p-5 sm:p-6 lg:grid-cols-[1fr_auto] lg:items-center">
        <div>
          <div class="mb-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">
            Model Catalog
          </div>
          <h1 class="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
            {{ productCopy.title }}
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
            {{ productCopy.description }}
          </p>
        </div>
        <div class="flex flex-wrap gap-2 lg:justify-end">
          <router-link
            :to="isAuthenticated ? '/keys' : '/register'"
            class="inline-flex h-10 items-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
          >
            {{ isAuthenticated ? productCopy.manageKeys : productCopy.getStarted }}
          </router-link>
          <button
            type="button"
            class="inline-flex h-10 items-center rounded-xl border border-gray-200 bg-white px-4 text-sm font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100 dark:hover:bg-dark-700"
            @click="copyBase"
          >
            {{ copied ? productCopy.copied : productCopy.copyBase }}
          </button>
        </div>
      </div>
      <div class="border-t border-gray-100 bg-gray-50/80 px-5 py-3.5 sm:px-6 dark:border-dark-800 dark:bg-dark-950/50">
        <div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
          <span class="text-[10px] font-semibold uppercase tracking-[0.14em] text-gray-400">Base URL</span>
          <code class="truncate text-xs font-medium text-gray-700 dark:text-dark-200">{{ apiBase }}</code>
          <span class="hidden text-gray-300 sm:inline dark:text-dark-600">•</span>
          <span class="text-xs text-gray-500 dark:text-dark-400">{{ productCopy.sameKey }}</span>
        </div>
      </div>
    </section>

    <div
      v-if="descriptionHtml"
      class="plaza-description rounded-2xl border border-gray-100 bg-white px-5 py-4 text-sm shadow-card dark:border-dark-700/50 dark:bg-dark-800/50"
      v-html="descriptionHtml"
    ></div>

    <p
      v-if="!isAuthenticated"
      class="flex items-center gap-1.5 text-xs text-gray-400 dark:text-dark-500"
    >
      <Icon name="infoCircle" size="xs" class="h-3.5 w-3.5" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <div v-if="loading" class="flex min-h-[240px] items-center justify-center">
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600/25 border-t-primary-600 dark:border-primary-400/25 dark:border-t-primary-400"></div>
    </div>
    <div
      v-else-if="error"
      class="rounded-2xl border border-red-200 bg-red-50 px-5 py-8 text-center text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
    >
      {{ t('modelPlaza.loadFailed') }}
    </div>
    <template v-else>
      <PlazaFilterBar
        :platforms="platforms"
        :groups="groupOptions"
        :rates="rates"
        :platform="selectedPlatform"
        :group-id="selectedGroupId"
        :rate="selectedRate"
        :search="searchQuery"
        @update:platform="selectedPlatform = $event"
        @update:group-id="selectedGroupId = $event"
        @update:rate="selectedRate = $event"
        @update:search="searchQuery = $event"
      />

      <div v-if="filteredGroups.length > 0" class="space-y-5">
        <PlazaGroupSection v-for="g in filteredGroups" :key="g.id" :group="g" />
      </div>
      <div
        v-else
        class="rounded-2xl border border-dashed border-gray-300 px-5 py-12 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-dark-400"
      >
        {{ searchActive ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import PlazaFilterBar from './PlazaFilterBar.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
  embedded?: boolean
}>()

const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const copied = ref(false)

const apiBase = computed(() => {
  const configured = appStore.cachedPublicSettings?.api_base_url?.trim()
  return (configured || `${window.location.origin.replace(/\/$/, '')}/v1`).replace(/\/$/, '')
})

const productCopy = computed(() =>
  isZh.value
    ? {
        title: '选择模型，然后用同一个 API Key 直接调用',
        description: '不同上游、模型和账号池由 TermRelay 在服务端统一调度。你只需要记住模型名、Base URL 和自己的 API Key。',
        manageKeys: '管理 API Key',
        getStarted: '注册并开始使用',
        copyBase: '复制 Base URL',
        copied: '已复制',
        sameKey: '所有可用模型使用统一入口'
      }
    : {
        title: 'Choose a model, then call it with the same API key',
        description: 'TermRelay handles upstream providers, model routing, and account pools on the server. You only need the model name, Base URL, and your API key.',
        manageKeys: 'Manage API Keys',
        getStarted: 'Create an account',
        copyBase: 'Copy Base URL',
        copied: 'Copied',
        sameKey: 'One unified endpoint for all available models'
      }
)

const selectedPlatform = ref<string>('all')
const selectedGroupId = ref<number | 'all'>('all')
const selectedRate = ref<number | 'all'>('all')
const searchQuery = ref('')

const searchActive = computed(() => searchQuery.value.trim() !== '')

const descriptionHtml = computed(() => {
  const md = props.response?.description?.trim()
  if (!md) return ''
  return DOMPurify.sanitize(marked.parse(md) as string)
})

function effectiveRate(g: ModelPlazaGroup): number {
  return g.user_rate_multiplier ?? g.rate_multiplier
}

const platforms = computed(() =>
  [...new Set((props.response?.groups ?? []).map((g) => g.platform).filter(Boolean))].sort()
)

const groupOptions = computed(() =>
  (props.response?.groups ?? []).map((g) => ({
    id: g.id,
    name: g.name,
    platform: g.platform,
    rate: effectiveRate(g)
  }))
)

const rates = computed(() =>
  [...new Set((props.response?.groups ?? []).map(effectiveRate))].sort((a, b) => a - b)
)

watch(rates, (list) => {
  if (selectedRate.value !== 'all' && !list.includes(selectedRate.value)) {
    selectedRate.value = 'all'
  }
})

const filteredGroups = computed(() => {
  let groups = props.response?.groups ?? []
  if (selectedPlatform.value !== 'all') {
    groups = groups.filter((g) => g.platform === selectedPlatform.value)
  }
  if (selectedGroupId.value !== 'all') {
    groups = groups.filter((g) => g.id === selectedGroupId.value)
  }
  if (selectedRate.value !== 'all') {
    groups = groups.filter((g) => effectiveRate(g) === selectedRate.value)
  }

  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    groups = groups
      .map((g) => ({ ...g, models: g.models.filter((m) => m.name.toLowerCase().includes(q)) }))
      .filter((g) => g.models.length > 0)
  }

  return [...groups].sort(
    (a, b) => effectiveRate(a) - effectiveRate(b) || a.name.localeCompare(b.name)
  )
})

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

<style scoped>
.plaza-description {
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.plaza-description :deep(h1),
.plaza-description :deep(h2),
.plaza-description :deep(h3) {
  @apply mb-2 mt-3 font-semibold text-gray-900 first:mt-0 dark:text-white;
}

.plaza-description :deep(p) {
  @apply mb-2 text-gray-700 last:mb-0 dark:text-dark-200;
}

.plaza-description :deep(a) {
  @apply text-primary-600 underline underline-offset-4 hover:text-primary-700 dark:text-primary-300;
}

.plaza-description :deep(ul) {
  @apply mb-2 list-disc pl-5;
}

.plaza-description :deep(ol) {
  @apply mb-2 list-decimal pl-5;
}

.plaza-description :deep(li) {
  @apply mb-0.5 text-gray-700 dark:text-dark-200;
}

.plaza-description :deep(code) {
  @apply rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs dark:bg-dark-800;
}

.plaza-description :deep(blockquote) {
  @apply my-2 border-l-4 border-gray-300 pl-3 text-gray-600 dark:border-dark-600 dark:text-dark-300;
}
</style>
