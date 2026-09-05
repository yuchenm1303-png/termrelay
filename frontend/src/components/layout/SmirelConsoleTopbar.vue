<template>
  <header class="smv3-topbar">
    <div class="smv3-topbar-left">
      <button
        type="button"
        class="smv3-mobile-menu"
        aria-label="Open navigation"
        @click="appStore.toggleMobileSidebar()"
      >
        <span></span><span></span><span></span>
      </button>
      <div class="smv3-breadcrumb">
        <span>{{ isAdmin ? copy.admin : copy.console }}</span>
        <b>/</b>
        <strong>{{ pageTitle }}</strong>
      </div>
    </div>

    <div class="smv3-topbar-actions">
      <button type="button" class="smv3-endpoint-pill" @click="copyEndpoint">
        <span class="smv3-endpoint-dot"></span>
        <code>{{ compactEndpoint }}</code>
        <span>{{ copied ? copy.copied : copy.copy }}</span>
      </button>

      <a
        v-if="docUrl"
        :href="docUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="smv3-topbar-link"
      >
        {{ copy.docs }}
      </a>

      <LocaleSwitcher class="smv3-locale" />

      <router-link v-if="user" to="/profile" class="smv3-topbar-account">
        <span class="smv3-topbar-avatar">{{ initials }}</span>
        <span class="smv3-topbar-account-copy">
          <strong>{{ displayName }}</strong>
          <span v-if="!isAdmin">${{ balance }}</span>
          <span v-else>{{ copy.admin }}</span>
        </span>
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
const isAdmin = computed(() => authStore.isAdmin)
const user = computed(() => authStore.user)
const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')
    .trim()
    .replace(/\/$/, ''),
)
const compactEndpoint = computed(() => apiBase.value.replace(/^https?:\/\//, ''))
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)

const copy = computed(() =>
  isZh.value
    ? { console: '开发者控制台', admin: '运营控制台', copy: '复制', copied: '已复制', docs: 'API 文档' }
    : { console: 'Developer Console', admin: 'Operations Console', copy: 'Copy', copied: 'Copied', docs: 'API Docs' },
)

const pageTitle = computed(() => {
  const key = route.meta.titleKey
  if (typeof key === 'string' && key) {
    const translated = t(key)
    if (translated !== key) return translated
  }
  return String(route.meta.title || (isAdmin.value ? copy.value.admin : copy.value.console))
})

const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || 'User')
const initials = computed(() => displayName.value.slice(0, 2).toUpperCase())
const balance = computed(() => Number(user.value?.balance || 0).toFixed(2))

async function copyEndpoint() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 1400)
  } catch (error) {
    console.warn('Failed to copy API endpoint:', error)
  }
}
</script>
