<template>
  <div class="relay-primary-nav-shell border-b border-gray-200/80 bg-white/80 px-4 backdrop-blur-xl md:px-6 lg:px-8 dark:border-dark-800 dark:bg-dark-950/80">
    <nav class="mx-auto flex min-h-12 max-w-[1600px] items-center gap-1 overflow-x-auto py-1 scrollbar-hide" aria-label="Primary user navigation">
      <router-link
        v-for="item in items"
        :key="item.path"
        :to="item.to"
        class="inline-flex h-9 shrink-0 items-center rounded-lg px-3 text-sm font-medium transition-colors"
        :class="isActive(item.path)
          ? 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300'
          : 'text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'"
      >
        {{ item.label }}
      </router-link>

      <a
        v-if="docUrl"
        :href="docUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex h-9 shrink-0 items-center gap-1 rounded-lg px-3 text-sm font-medium text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
      >
        {{ isZh ? 'API 文档' : 'API Docs' }}
        <span aria-hidden="true" class="text-xs">↗</span>
      </a>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'

const route = useRoute()
const { locale } = useI18n()
const appStore = useAppStore()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const paymentEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.payment))
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)

const items = computed(() => {
  const list = [
    { path: '/dashboard', to: '/dashboard', label: isZh.value ? '控制台' : 'Console' },
    { path: '/keys', to: '/keys', label: isZh.value ? 'API 密钥' : 'API Keys' },
  ]

  if (modelPlazaEnabled.value) {
    list.push({
      path: '/model-plaza',
      to: '/model-plaza?embedded=1',
      label: isZh.value ? '模型与价格' : 'Models & Pricing',
    })
  }

  list.push(
    { path: '/usage', to: '/usage', label: isZh.value ? '调用记录' : 'Usage' },
    { path: '/subscriptions', to: '/subscriptions', label: isZh.value ? '额度与套餐' : 'Plans & Quota' },
  )

  if (paymentEnabled.value) {
    list.push({
      path: '/purchase',
      to: '/purchase',
      label: isZh.value ? '充值 / 购买' : 'Top up / Buy',
    })
  }

  return list
})

function isActive(path: string) {
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>
