<template>
  <AppLayout v-if="isEmbedded">
    <section class="smg-page">
      <header class="smg-page-head">
        <div>
          <div class="smg-page-kicker">SMIREL MODEL CATALOG</div>
          <h1 class="smg-page-title">{{ copy.title }}</h1>
          <p class="smg-page-description">{{ copy.description }}</p>
        </div>
      </header>

      <GlassSurface class="smg-data-panel smp-model-panel">
        <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" embedded />
      </GlassSurface>
    </section>
  </AppLayout>

  <section v-else class="smp-model-page">
    <header class="smp-model-head">
      <div class="smp-model-head__copy">
        <div class="smp-model-kicker">SMIREL MODEL CATALOG</div>
        <h1>{{ copy.title }}</h1>
        <p>{{ copy.description }}</p>
      </div>
      <aside class="smp-model-meta">
        <span>API BASE</span>
        <code>{{ apiBase }}</code>
        <span>COMPATIBILITY</span>
        <code>OpenAI-compatible</code>
      </aside>
    </header>

    <div class="smp-model-panel">
      <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import GlassSurface from '@/components/glass/GlassSurface.vue'
import ModelPlazaContent from '@/components/modelPlaza/ModelPlazaContent.vue'
import { getModelPlaza, type ModelPlazaResponse } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const isEmbedded = computed(() => route.query.embedded === '1' && authStore.isAuthenticated)
const apiBase = computed(() =>
  appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1',
)

const copy = computed(() =>
  isZh.value
    ? {
        title: '模型与价格',
        description: '从产品视角查看可以调用的模型、能力与价格。这里不暴露上游账号、内部路由分组或调度细节。',
      }
    : {
        title: 'Models & pricing',
        description: 'Browse callable models, capabilities and pricing from the product perspective. Upstream accounts and internal routing stay hidden.',
      },
)

const data = ref<ModelPlazaResponse | null>(null)
const loading = ref(true)
const loadFailed = ref(false)

onMounted(async () => {
  void appStore.fetchPublicSettings()
  try {
    data.value = await getModelPlaza()
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
})
</script>
