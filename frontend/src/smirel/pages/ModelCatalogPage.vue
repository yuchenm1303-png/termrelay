<script setup lang="ts">
import { computed, ref } from 'vue'
import { interfacePreferences } from '../core/preferences'

type ProviderKey = 'openai' | 'anthropic' | 'google' | 'xai'
type CapabilityKey = 'all' | 'text' | 'image' | 'reasoning' | 'code'
type SortKey = 'popular' | 'price' | 'context' | 'latency'

type ModelEntry = {
  rank: number
  provider: ProviderKey
  providerName: string
  mark: string
  name: string
  modelId: string
  family: string
  contextK: number
  contextLabel: string
  inputPrice: number
  outputPrice: number
  cacheReadPrice?: number
  cacheWritePrice?: number
  capabilities: Exclude<CapabilityKey, 'all'>[]
  protocols: string[]
  inputTypes: string
  outputTypes: string
  latencyMs: number
  availability: number
  pricingGroup: string
  note: string
}

const isZh = computed(() => interfacePreferences.locale === 'zh-CN')
const search = ref('')
const provider = ref<'all' | ProviderKey>('all')
const capability = ref<CapabilityKey>('all')
const sortBy = ref<SortKey>('popular')
const copiedId = ref('')

const copy = computed(() => isZh.value
  ? {
      kicker: 'MODEL MARKETPLACE',
      title: '模型广场',
      description: '浏览当前可用模型，直接比较价格、上下文、能力与接口兼容方式。',
      searchPlaceholder: '搜索品牌、模型名称或模型 ID',
      allProviders: '全部模型',
      allCapabilities: '全部能力',
      text: '文本',
      image: '图像',
      reasoning: '推理',
      code: '代码',
      popular: '推荐排序',
      price: '输入价格',
      context: '上下文',
      latency: '响应速度',
      models: '个模型',
      providers: '家服务商',
      baseUrl: '统一 Base URL',
      preview: '演示目录',
      previewHint: '当前为 UI 演示数据，正式接入后由模型目录接口提供。',
      input: '输入',
      output: '输出',
      cacheRead: '缓存读取',
      cacheWrite: '缓存写入',
      perMillion: '/ M tokens',
      inputType: '输入类型',
      outputType: '输出类型',
      contextWindow: '上下文',
      modelId: '模型 ID',
      protocol: '兼容接口',
      live: '最近可用状态',
      recent: '近 30 次',
      copy: '复制',
      copied: '已复制',
      empty: '没有符合当前筛选条件的模型。',
      reset: '清除筛选',
    }
  : {
      kicker: 'MODEL MARKETPLACE',
      title: 'Model Marketplace',
      description: 'Browse available models and compare pricing, context, capabilities, and compatible API formats.',
      searchPlaceholder: 'Search provider, model name, or model ID',
      allProviders: 'All models',
      allCapabilities: 'All capabilities',
      text: 'Text',
      image: 'Image',
      reasoning: 'Reasoning',
      code: 'Code',
      popular: 'Recommended',
      price: 'Input price',
      context: 'Context',
      latency: 'Latency',
      models: 'models',
      providers: 'providers',
      baseUrl: 'Unified Base URL',
      preview: 'Preview catalog',
      previewHint: 'UI preview data. The production catalog will be supplied by the model catalog API.',
      input: 'Input',
      output: 'Output',
      cacheRead: 'Cache read',
      cacheWrite: 'Cache write',
      perMillion: '/ M tokens',
      inputType: 'Input types',
      outputType: 'Output types',
      contextWindow: 'Context',
      modelId: 'Model ID',
      protocol: 'Compatible APIs',
      live: 'Recent availability',
      recent: 'last 30',
      copy: 'Copy',
      copied: 'Copied',
      empty: 'No models match the current filters.',
      reset: 'Reset filters',
    })

const models: ModelEntry[] = [
  {
    rank: 1,
    provider: 'openai', providerName: 'OpenAI', mark: 'O',
    name: 'GPT-5.6 Sol', modelId: 'gpt-5.6-sol', family: 'GPT',
    contextK: 272, contextLabel: '272K', inputPrice: 5, outputPrice: 25,
    cacheReadPrice: 0.5, cacheWritePrice: 6.25,
    capabilities: ['text', 'image', 'reasoning', 'code'], protocols: ['Responses', 'Chat Completions'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 1236, availability: 98.2,
    pricingGroup: '1.0×', note: '高质量通用推理与代码任务',
  },
  {
    rank: 2,
    provider: 'anthropic', providerName: 'Anthropic', mark: 'A',
    name: 'Claude Sonnet 4.6', modelId: 'claude-sonnet-4-6', family: 'Claude',
    contextK: 1000, contextLabel: '1M', inputPrice: 3, outputPrice: 15,
    cacheReadPrice: 0.3, cacheWritePrice: 3.75,
    capabilities: ['text', 'image', 'reasoning', 'code'], protocols: ['Messages', 'Chat Completions'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 1696, availability: 97.8,
    pricingGroup: '1.0×', note: '长上下文、代码与复杂工作流',
  },
  {
    rank: 3,
    provider: 'google', providerName: 'Google', mark: 'G',
    name: 'Gemini 3.8 Flash', modelId: 'gemini-3.8-flash', family: 'Gemini',
    contextK: 1000, contextLabel: '1M', inputPrice: 0.5, outputPrice: 3,
    cacheReadPrice: 0.05,
    capabilities: ['text', 'image', 'reasoning', 'code'], protocols: ['Gemini API', 'OpenAI Compatible'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 742, availability: 99.1,
    pricingGroup: '0.5×', note: '高频、低延迟与多模态调用',
  },
  {
    rank: 4,
    provider: 'openai', providerName: 'OpenAI', mark: 'O',
    name: 'GPT-5.6 Terra', modelId: 'gpt-5.6-terra', family: 'GPT',
    contextK: 272, contextLabel: '272K', inputPrice: 2, outputPrice: 10,
    cacheReadPrice: 0.2, cacheWritePrice: 2.5,
    capabilities: ['text', 'image', 'code'], protocols: ['Responses', 'Chat Completions'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 864, availability: 98.8,
    pricingGroup: '0.8×', note: '兼顾质量、速度与成本的通用模型',
  },
  {
    rank: 5,
    provider: 'anthropic', providerName: 'Anthropic', mark: 'A',
    name: 'Claude Haiku 4.5', modelId: 'claude-haiku-4-5', family: 'Claude',
    contextK: 200, contextLabel: '200K', inputPrice: 1, outputPrice: 5,
    cacheReadPrice: 0.1, cacheWritePrice: 1.25,
    capabilities: ['text', 'image', 'code'], protocols: ['Messages', 'Chat Completions'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 618, availability: 99.4,
    pricingGroup: '0.6×', note: '低延迟、高并发生产任务',
  },
  {
    rank: 6,
    provider: 'xai', providerName: 'xAI', mark: 'X',
    name: 'Grok 4 Fast', modelId: 'grok-4-fast', family: 'Grok',
    contextK: 256, contextLabel: '256K', inputPrice: 2.5, outputPrice: 12,
    capabilities: ['text', 'image', 'reasoning'], protocols: ['OpenAI Compatible'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 1188, availability: 96.7,
    pricingGroup: '1.0×', note: '实时信息与通用推理场景',
  },
  {
    rank: 7,
    provider: 'openai', providerName: 'OpenAI', mark: 'O',
    name: 'GPT-5.5', modelId: 'gpt-5.5', family: 'GPT',
    contextK: 272, contextLabel: '272K', inputPrice: 2.5, outputPrice: 15,
    cacheReadPrice: 0.25, cacheWritePrice: 3.125,
    capabilities: ['text', 'image', 'reasoning', 'code'], protocols: ['Responses', 'Chat Completions'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 1030, availability: 97.4,
    pricingGroup: '1.0×', note: '稳定通用模型与成熟兼容性',
  },
  {
    rank: 8,
    provider: 'google', providerName: 'Google', mark: 'G',
    name: 'Gemini 3.5 Pro', modelId: 'gemini-3.5-pro', family: 'Gemini',
    contextK: 1000, contextLabel: '1M', inputPrice: 2, outputPrice: 12,
    cacheReadPrice: 0.2,
    capabilities: ['text', 'image', 'reasoning', 'code'], protocols: ['Gemini API', 'OpenAI Compatible'],
    inputTypes: '文本 · 图像', outputTypes: '文本', latencyMs: 1472, availability: 97.1,
    pricingGroup: '0.8×', note: '长上下文与高质量多模态生成',
  },
]

const providerOptions = computed(() => {
  const seen = new Map<ProviderKey, { key: ProviderKey; name: string; count: number }>()
  for (const item of models) {
    const current = seen.get(item.provider)
    if (current) current.count += 1
    else seen.set(item.provider, { key: item.provider, name: item.providerName, count: 1 })
  }
  return Array.from(seen.values())
})

const filteredModels = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  const items = models.filter((item) => {
    const matchesProvider = provider.value === 'all' || item.provider === provider.value
    const matchesCapability = capability.value === 'all' || item.capabilities.includes(capability.value as Exclude<CapabilityKey, 'all'>)
    const haystack = `${item.providerName} ${item.name} ${item.modelId} ${item.family} ${item.note}`.toLowerCase()
    return matchesProvider && matchesCapability && (!keyword || haystack.includes(keyword))
  })

  return [...items].sort((a, b) => {
    if (sortBy.value === 'price') return a.inputPrice - b.inputPrice
    if (sortBy.value === 'context') return b.contextK - a.contextK
    if (sortBy.value === 'latency') return a.latencyMs - b.latencyMs
    return a.rank - b.rank
  })
})

function capabilityLabel(value: CapabilityKey) {
  if (value === 'text') return copy.value.text
  if (value === 'image') return copy.value.image
  if (value === 'reasoning') return copy.value.reasoning
  if (value === 'code') return copy.value.code
  return copy.value.allCapabilities
}

function money(value?: number) {
  if (value === undefined) return '—'
  return `$${value < 1 ? value.toFixed(2) : value.toFixed(value % 1 === 0 ? 2 : 2)}`
}

function resetFilters() {
  search.value = ''
  provider.value = 'all'
  capability.value = 'all'
  sortBy.value = 'popular'
}

async function copyModelId(modelId: string) {
  if (typeof navigator === 'undefined' || !navigator.clipboard) return
  await navigator.clipboard.writeText(modelId)
  copiedId.value = modelId
  window.setTimeout(() => {
    if (copiedId.value === modelId) copiedId.value = ''
  }, 1400)
}

function healthTone(item: ModelEntry, index: number) {
  const healthySegments = Math.max(1, Math.round((item.availability / 100) * 30))
  return index <= healthySegments ? 'good' : 'bad'
}
</script>

<template>
  <section class="workspace-page model-market-page">
    <header class="model-market-heading">
      <div class="market-heading-copy">
        <span class="market-kicker">{{ copy.kicker }}</span>
        <div class="market-title-line">
          <h1>{{ copy.title }}</h1>
          <span class="market-count">{{ models.length }} {{ copy.models }}</span>
        </div>
        <p>{{ copy.description }}</p>
      </div>

      <div class="market-heading-actions">
        <label class="market-search">
          <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="m13 13 4 4" /></svg>
          <input v-model="search" type="search" :placeholder="copy.searchPlaceholder" />
        </label>
        <label class="market-sort">
          <span>{{ isZh ? '排序' : 'Sort' }}</span>
          <select v-model="sortBy">
            <option value="popular">{{ copy.popular }}</option>
            <option value="price">{{ copy.price }}</option>
            <option value="context">{{ copy.context }}</option>
            <option value="latency">{{ copy.latency }}</option>
          </select>
        </label>
      </div>
    </header>

    <section class="market-filter-shell">
      <div class="provider-filter" role="tablist" :aria-label="isZh ? '模型服务商' : 'Providers'">
        <button type="button" :class="{ active: provider === 'all' }" @click="provider = 'all'">
          <span>{{ copy.allProviders }}</span>
          <b>{{ models.length }}</b>
        </button>
        <button
          v-for="item in providerOptions"
          :key="item.key"
          type="button"
          :class="{ active: provider === item.key }"
          @click="provider = item.key"
        >
          <span>{{ item.name }}</span>
          <b>{{ item.count }}</b>
        </button>
      </div>

      <div class="market-filter-meta">
        <div class="capability-filter">
          <button
            v-for="item in (['all', 'text', 'image', 'reasoning', 'code'] as CapabilityKey[])"
            :key="item"
            type="button"
            :class="{ active: capability === item }"
            @click="capability = item"
          >{{ capabilityLabel(item) }}</button>
        </div>
        <div class="catalog-state">
          <span><i></i>{{ copy.preview }}</span>
          <small>{{ copy.previewHint }}</small>
        </div>
      </div>
    </section>

    <section class="market-result-head">
      <div>
        <strong>{{ filteredModels.length }} {{ copy.models }}</strong>
        <span>·</span>
        <span>{{ providerOptions.length }} {{ copy.providers }}</span>
      </div>
      <div>
        <span>{{ copy.baseUrl }}</span>
        <code>https://api.smirel.com/v1</code>
      </div>
    </section>

    <div v-if="filteredModels.length" class="model-market-grid">
      <article v-for="item in filteredModels" :key="item.modelId" class="model-market-card">
        <header class="model-card-head">
          <div class="model-identity">
            <span class="provider-mark" :data-provider="item.provider">{{ item.mark }}</span>
            <div>
              <h2>{{ item.name }}</h2>
              <div class="model-tags">
                <span>{{ item.providerName }}</span>
                <span>{{ item.pricingGroup }}</span>
                <span>{{ item.contextLabel }}</span>
              </div>
            </div>
          </div>
          <span class="model-rank">#{{ String(item.rank).padStart(2, '0') }}</span>
        </header>

        <section class="model-price-grid">
          <div>
            <span>{{ copy.input }}</span>
            <strong>{{ money(item.inputPrice) }}<small>{{ copy.perMillion }}</small></strong>
          </div>
          <div>
            <span>{{ copy.output }}</span>
            <strong>{{ money(item.outputPrice) }}<small>{{ copy.perMillion }}</small></strong>
          </div>
          <div>
            <span>{{ copy.cacheWrite }}</span>
            <strong>{{ money(item.cacheWritePrice) }}<small v-if="item.cacheWritePrice !== undefined">{{ copy.perMillion }}</small></strong>
          </div>
          <div>
            <span>{{ copy.cacheRead }}</span>
            <strong>{{ money(item.cacheReadPrice) }}<small v-if="item.cacheReadPrice !== undefined">{{ copy.perMillion }}</small></strong>
          </div>
        </section>

        <section class="model-specs">
          <div><span>{{ copy.inputType }}</span><strong>{{ item.inputTypes }}</strong></div>
          <div><span>{{ copy.outputType }}</span><strong>{{ item.outputTypes }}</strong></div>
          <div><span>{{ copy.contextWindow }}</span><strong>{{ item.contextLabel }}</strong></div>
        </section>

        <section class="model-access-block">
          <div class="model-id-row">
            <span>{{ copy.modelId }}</span>
            <code>{{ item.modelId }}</code>
            <button type="button" @click="copyModelId(item.modelId)">
              {{ copiedId === item.modelId ? copy.copied : copy.copy }}
            </button>
          </div>
          <div class="protocol-list">
            <span>{{ copy.protocol }}</span>
            <div>
              <b v-for="protocol in item.protocols" :key="protocol">{{ protocol }}</b>
            </div>
          </div>
        </section>

        <footer class="model-health">
          <div class="health-label">
            <span>{{ copy.live }} · {{ copy.recent }}</span>
            <strong>{{ item.availability.toFixed(1) }}%</strong>
          </div>
          <div class="health-track" aria-hidden="true">
            <i v-for="index in 30" :key="index" :class="healthTone(item, index)"></i>
          </div>
          <div class="health-latency">
            <span>{{ item.note }}</span>
            <strong>{{ item.latencyMs }}ms</strong>
          </div>
        </footer>
      </article>
    </div>

    <section v-else class="market-empty">
      <strong>{{ copy.empty }}</strong>
      <button type="button" @click="resetFilters">{{ copy.reset }}</button>
    </section>
  </section>
</template>
