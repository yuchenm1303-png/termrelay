<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'

interface ApiKeyItem {
  id: number
  name?: string
  key?: string
  status?: string
  created_at?: string
  [key: string]: unknown
}

const props = defineProps<{ item: ApiKeyItem }>()
const emit = defineEmits<{ remove: [id: number] }>()
const { locale, t } = useI18n()

const revealed = ref(false)
const copied = ref(false)
let copiedTimer: number | undefined

const labels = computed(() => locale.value === 'zh-CN'
  ? { show: '显示', hide: '隐藏', copy: '复制', copied: '已复制', clickCopy: '点击复制完整密钥' }
  : { show: 'Show', hide: 'Hide', copy: 'Copy', copied: 'Copied', clickCopy: 'Click to copy full key' })

const rawKey = computed(() => String(props.item.key || ''))
const serverMasked = computed(() => /[•*]{3,}/.test(rawKey.value))
const canReveal = computed(() => Boolean(rawKey.value) && !serverMasked.value)

const maskedKey = computed(() => {
  const value = rawKey.value
  if (!value) return '••••••••'
  if (serverMasked.value) return value
  if (value.length <= 12) return '••••••••'
  const middleLength = Math.min(24, Math.max(10, value.length - 11))
  return `${value.slice(0, 7)}${'•'.repeat(middleLength)}${value.slice(-4)}`
})

const displayKey = computed(() => revealed.value && canReveal.value ? rawKey.value : maskedKey.value)

async function writeClipboard(value: string) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      return
    } catch {
      // Fall through to the textarea path for restricted clipboard contexts.
    }
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

async function copyKey() {
  if (!rawKey.value) return
  await writeClipboard(rawKey.value)
  copied.value = true
  if (copiedTimer) window.clearTimeout(copiedTimer)
  copiedTimer = window.setTimeout(() => { copied.value = false }, 1600)
}

function toggleReveal() {
  if (!canReveal.value) return
  revealed.value = !revealed.value
}

onBeforeUnmount(() => {
  if (copiedTimer) window.clearTimeout(copiedTimer)
})
</script>

<template>
  <article class="api-key-card">
    <header class="api-key-card-head">
      <div class="api-key-identity">
        <span class="api-key-mark" aria-hidden="true">
          <svg viewBox="0 0 24 24">
            <circle cx="8" cy="15" r="3.5" />
            <path d="m10.7 12.3 7.6-7.6M15.7 7.3l2 2M13.6 9.4l2 2" />
          </svg>
        </span>
        <div>
          <strong>{{ item.name || `Key #${item.id}` }}</strong>
          <small>API KEY</small>
        </div>
      </div>
      <span class="api-key-state"><i></i>{{ item.status || 'active' }}</span>
    </header>

    <div class="api-key-secret">
      <div class="api-key-secret-head">
        <span>{{ t('workspace.key') }}</span>
        <div class="api-key-secret-actions">
          <button
            v-if="canReveal"
            type="button"
            :aria-label="revealed ? labels.hide : labels.show"
            :title="revealed ? labels.hide : labels.show"
            @click="toggleReveal"
          >
            <svg v-if="!revealed" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6S2 12 2 12Z" />
              <circle cx="12" cy="12" r="2.8" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="m3 3 18 18" />
              <path d="M10.6 6.2A11 11 0 0 1 12 6c6.5 0 10 6 10 6a17 17 0 0 1-2.1 2.8" />
              <path d="M6.6 6.6C3.6 8.3 2 12 2 12s3.5 6 10 6a10.8 10.8 0 0 0 4.3-.9" />
              <path d="M9.9 9.9A3 3 0 0 0 14.1 14.1" />
            </svg>
            <span>{{ revealed ? labels.hide : labels.show }}</span>
          </button>
          <button
            class="api-key-copy-button"
            :class="{ copied }"
            type="button"
            :aria-label="copied ? labels.copied : labels.copy"
            :title="copied ? labels.copied : labels.copy"
            @click="copyKey"
          >
            <svg v-if="!copied" viewBox="0 0 24 24" aria-hidden="true">
              <rect x="9" y="9" width="11" height="11" rx="2" />
              <path d="M15 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h3" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="m5 12 4 4L19 6" />
            </svg>
            <span>{{ copied ? labels.copied : labels.copy }}</span>
          </button>
        </div>
      </div>

      <button class="api-key-value" type="button" :title="labels.clickCopy" @click="copyKey">
        <code>{{ displayKey }}</code>
      </button>
    </div>

    <footer class="api-key-card-foot">
      <span class="api-key-created">
        <b>{{ t('workspace.createdAt') }}</b>
        {{ item.created_at || '—' }}
      </span>
      <button class="api-key-delete" type="button" @click="emit('remove', item.id)">{{ t('workspace.delete') }}</button>
    </footer>
  </article>
</template>
