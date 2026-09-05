<template>
  <section class="mx-auto mb-6 max-w-6xl overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
    <div class="grid gap-5 p-5 sm:p-6 lg:grid-cols-[1fr_auto] lg:items-center">
      <div>
        <div class="mb-2 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
          {{ copy.badge }}
        </div>
        <h1 class="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
          {{ copy.title }}
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
          {{ copy.description }}
        </p>
      </div>

      <div class="flex flex-wrap gap-2 lg:justify-end">
        <router-link
          v-if="route.path !== '/subscriptions'"
          to="/subscriptions"
          class="inline-flex h-10 items-center rounded-xl border border-gray-200 bg-white px-4 text-sm font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100 dark:hover:bg-dark-700"
        >
          {{ copy.myPlans }}
        </router-link>
        <router-link
          v-if="paymentEnabled && route.path !== '/purchase'"
          to="/purchase"
          class="inline-flex h-10 items-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
        >
          {{ copy.buy }}
        </router-link>
        <router-link
          v-if="paymentEnabled"
          to="/orders"
          class="inline-flex h-10 items-center rounded-xl border border-gray-200 bg-white px-4 text-sm font-semibold text-gray-700 transition hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100 dark:hover:bg-dark-700"
        >
          {{ copy.orders }}
        </router-link>
      </div>
    </div>

    <div class="grid border-t border-gray-100 bg-gray-50/80 sm:grid-cols-3 dark:border-dark-800 dark:bg-dark-950/50">
      <div class="border-b border-gray-100 px-5 py-4 sm:border-b-0 sm:border-r dark:border-dark-800">
        <p class="text-[10px] font-semibold uppercase tracking-[0.14em] text-gray-400">{{ copy.balanceLabel }}</p>
        <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">${{ balance.toFixed(2) }}</p>
      </div>
      <div class="border-b border-gray-100 px-5 py-4 sm:border-b-0 sm:border-r dark:border-dark-800">
        <p class="text-[10px] font-semibold uppercase tracking-[0.14em] text-gray-400">{{ copy.nextLabel }}</p>
        <router-link to="/model-plaza?embedded=1" class="mt-1 inline-flex text-sm font-semibold text-gray-800 transition hover:text-primary-600 dark:text-dark-100 dark:hover:text-primary-400">
          {{ copy.models }} →
        </router-link>
      </div>
      <div class="px-5 py-4">
        <p class="text-[10px] font-semibold uppercase tracking-[0.14em] text-gray-400">{{ copy.helpLabel }}</p>
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="mt-1 inline-flex text-sm font-semibold text-gray-800 transition hover:text-primary-600 dark:text-dark-100 dark:hover:text-primary-400"
        >
          {{ copy.docs }} →
        </a>
        <span v-else class="mt-1 block text-sm text-gray-500 dark:text-dark-400">{{ copy.helpFallback }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'

const route = useRoute()
const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const paymentEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.payment))
const balance = computed(() => Number(authStore.user?.balance ?? 0))
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)

const copy = computed(() => {
  const purchase = route.path === '/purchase'
  if (isZh.value) {
    return {
      badge: 'Billing & Quota',
      title: purchase ? '充值或购买套餐' : '管理你的额度与套餐',
      description: purchase
        ? '选择余额充值或订阅套餐，支付完成后额度会自动同步到你的账户。'
        : '查看当前套餐、剩余额度与有效期；需要更多额度时可以直接充值或购买新套餐。',
      myPlans: '我的套餐',
      buy: '充值 / 购买套餐',
      orders: '订单记录',
      balanceLabel: '当前余额',
      nextLabel: '下一步',
      models: '查看模型与价格',
      helpLabel: '接入帮助',
      docs: 'API 文档',
      helpFallback: '完成支付后即可继续使用 API',
    }
  }
  return {
    badge: 'Billing & Quota',
    title: purchase ? 'Top up or purchase a plan' : 'Manage plans and quota',
    description: purchase
      ? 'Choose a balance top-up or subscription plan. Your quota updates automatically after payment.'
      : 'Review active plans, remaining quota, and expiration dates, then add more capacity when needed.',
    myPlans: 'My plans',
    buy: 'Top up / Buy plan',
    orders: 'Orders',
    balanceLabel: 'Balance',
    nextLabel: 'Next step',
    models: 'View models & pricing',
    helpLabel: 'Integration help',
    docs: 'API Docs',
    helpFallback: 'Continue using the API after your payment is completed',
  }
})
</script>
