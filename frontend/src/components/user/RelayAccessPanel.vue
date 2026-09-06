<template>
  <GlassSurface class="spw-access-panel">
    <div class="spw-access-grid">
      <div class="spw-access-copy">
        <div class="spw-access-kicker">API ACCESS</div>
        <h1>{{ copy.title }}</h1>
        <p>{{ copy.description }}</p>

        <div class="spw-access-actions">
          <button type="button" class="smg-button smg-button--primary" @click="openCreateKey">
            {{ copy.create }}
          </button>
          <router-link to="/model-plaza?embedded=1" class="smg-button">
            {{ copy.models }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="smg-button"
          >
            {{ copy.docs }}
          </a>
        </div>
      </div>

      <div class="spw-access-code">
        <div class="spw-access-code-head">
          <div class="min-w-0">
            <span>BASE URL</span>
            <code>{{ apiBase }}</code>
          </div>
          <button type="button" class="smg-button" @click="copyBase">
            {{ copied ? copy.copied : copy.copy }}
          </button>
        </div>
        <pre><code>curl {{ apiBase }}/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_NAME","messages":[{"role":"user","content":"Hello"}]}'</code></pre>
      </div>
    </div>

    <div class="spw-access-steps">
      <div v-for="(step, index) in copy.steps" :key="step" class="spw-access-step">
        <b>{{ index + 1 }}</b>
        <span>{{ step }}</span>
      </div>
    </div>
  </GlassSurface>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'
import GlassSurface from '@/components/glass/GlassSurface.vue'

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
        title: '创建 API Key 后即可开始调用。',
        description: '创建凭证、选择可用模型，然后把 Base URL 和 Key 配置到 OpenAI-compatible 客户端。密钥状态、额度和调用记录都在当前工作区统一管理。',
        create: '创建 API Key',
        models: '模型与价格',
        docs: 'API 文档',
        copy: '复制',
        copied: '已复制',
        steps: ['创建 API Key', '选择模型', '配置 Base URL 开始调用']
      }
    : {
        title: 'Create an API key and start calling.',
        description: 'Create a credential, choose an available model, then configure the Base URL and key in any OpenAI-compatible client. Key status, quota and request history stay in this workspace.',
        create: 'Create API Key',
        models: 'Models & Pricing',
        docs: 'API Docs',
        copy: 'Copy',
        copied: 'Copied',
        steps: ['Create an API key', 'Choose a model', 'Configure the Base URL and call']
      }
)

function openCreateKey() {
  const button = document.querySelector<HTMLElement>('[data-tour="keys-create-btn"]')
  button?.click()
}

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
