<template>
  <header class="smg-topbar">
    <div class="smg-topbar-title">
      <button class="smg-mobile-menu" type="button" aria-label="Open navigation" @click="appStore.toggleMobileSidebar()">
        <span></span><span></span><span></span>
      </button>
      <div>
        <span>{{ copy.console }}</span>
        <strong>{{ pageTitle }}</strong>
      </div>
    </div>

    <div class="smg-topbar-actions">
      <button type="button" class="smg-endpoint" @click="copyEndpoint">
        <i></i>
        <code>{{ compactEndpoint }}</code>
        <span>{{ copied ? copy.copied : copy.copy }}</span>
      </button>

      <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="smg-topbar-link">
        {{ copy.docs }}
      </a>
      <LocaleSwitcher class="smg-locale" />
      <router-link to="/profile" class="smg-balance">
        <span>{{ copy.balance }}</span>
        <strong>${{ balance }}</strong>
      </router-link>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'

const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const copied = ref(false)

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1').trim().replace(/\/$/, ''),
)
const compactEndpoint = computed(() => apiBase.value.replace(/^https?:\/\//, ''))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const balance = computed(() => Number(authStore.user?.balance || 0).toFixed(2))

const copy = computed(() => isZh.value
  ? { console: 'SMIREL API', copy: '复制', copied: '已复制', docs: 'API 文档', balance: '余额' }
  : { console: 'SMIREL API', copy: 'Copy', copied: 'Copied', docs: 'API Docs', balance: 'Balance' })

const pageTitle = computed(() => {
  const key = route.meta.titleKey
  if (typeof key === 'string' && key) {
    const translated = t(key)
    if (translated !== key) return translated
  }
  return String(route.meta.title || (isZh.value ? '工作区' : 'Workspace'))
})

async function copyEndpoint() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    copied.value = true
    window.setTimeout(() => { copied.value = false }, 1400)
  } catch (error) {
    console.warn('Failed to copy API endpoint:', error)
  }
}
</script>
