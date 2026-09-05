<template>
  <section class="smg-page">
    <header class="smg-page-head">
      <div>
        <div class="smg-page-kicker">{{ kicker }}</div>
        <h1 class="smg-page-title">{{ title }}</h1>
        <p v-if="description" class="smg-page-description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="smg-page-actions"><slot name="actions" /></div>
    </header>

    <GlassSurface v-if="$slots.filters" class="smg-page-toolbar">
      <div class="smg-page-filters"><slot name="filters" /></div>
    </GlassSurface>

    <GlassSurface class="smg-data-panel">
      <div class="smg-data-scroll"><slot name="table" /></div>
      <div v-if="$slots.pagination" class="smg-data-pagination"><slot name="pagination" /></div>
    </GlassSurface>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import GlassSurface from '@/components/glass/GlassSurface.vue'

const route = useRoute()
const authStore = useAuthStore()
const { t, locale } = useI18n()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const kicker = computed(() => authStore.isAdmin ? 'SMIREL OPERATIONS' : 'SMIREL API WORKSPACE')

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
