<template>
  <section class="smv3-page">
    <header class="smv3-page-head">
      <div>
        <div class="smv3-page-kicker">{{ kicker }}</div>
        <h1 class="smv3-page-title">{{ title }}</h1>
        <p v-if="description" class="smv3-page-description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="smv3-toolbar-actions">
        <slot name="actions" />
      </div>
    </header>

    <div v-if="$slots.filters" class="smv3-toolbar">
      <div class="smv3-toolbar-filters">
        <slot name="filters" />
      </div>
    </div>

    <div class="smv3-panel">
      <div class="smv3-panel-scroll">
        <slot name="table" />
      </div>
      <div v-if="$slots.pagination" class="smv3-pagination">
        <slot name="pagination" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const { t, locale } = useI18n()

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const kicker = computed(() =>
  authStore.isAdmin
    ? (isZh.value ? 'SMIREL OPERATIONS' : 'SMIREL OPERATIONS')
    : (isZh.value ? 'SMIREL DEVELOPER PLATFORM' : 'SMIREL DEVELOPER PLATFORM'),
)

const title = computed(() => {
  const key = route.meta.titleKey
  if (typeof key === 'string' && key) {
    const translated = t(key)
    if (translated !== key) return translated
  }
  return String(route.meta.title || (isZh.value ? '工作区' : 'Workspace'))
})

const description = computed(() => {
  const key = route.meta.descriptionKey
  if (typeof key === 'string' && key) {
    const translated = t(key)
    if (translated !== key) return translated
  }
  return ''
})
</script>
