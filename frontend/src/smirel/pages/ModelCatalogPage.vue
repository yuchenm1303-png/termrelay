<script setup lang="ts">
import { computed, ref } from 'vue'
import { interfacePreferences } from '../core/preferences'

type ProviderKey = 'openai' | 'anthropic' | 'google'

type ModelEntry = {
  family: string
  api: string
  input: string
  output: string
  note: string
}

type ProviderCatalog = {
  key: ProviderKey
  index: string
  name: string
  protocol: string
  summary: string
  models: ModelEntry[]
}

const isZh = computed(() => interfacePreferences.locale === 'zh-CN')
const activeProvider = ref<ProviderKey>('openai')

const copy = computed(() => isZh.value
  ? {
      eyebrow: 'MODEL CATALOG',
      title: '模型与价格',
      description: '在控制台内统一查看可用模型、接口方式与计费结构。实际可用模型和结算费率以当前账户为准。',
      access: '统一接入',
      accessValue: 'https://api.smirel.com/v1',
      accessHint: '一个 Base URL 接入多家模型服务',
      billing: '计费方式',
      billingValue: '按实际用量',
      billingHint: '输入与输出分别记录并结算',
      record: '费用记录',
      recordValue: '控制台可查',
      recordHint: '请求、Token 与费用集中查看',
      provider: '模型服务商',
      modelFamily: '模型系列',
      api: '接口',
      inputPrice: '输入价格',
      outputPrice: '输出价格',
      useCase: '适用场景',
      status: '计费状态',
      liveRate: '实时费率',
      noteTitle: '价格说明',
      note: '这里展示计费结构，不写死可能变化的上游单价。最终费用以请求发生时的控制台结算信息为准。',
    }
  : {
      eyebrow: 'MODEL CATALOG',
      title: 'Models & Pricing',
      description: 'Review available models, API formats, and billing structure directly inside the console. Availability and settlement rates follow the current account.',
      access: 'Unified access',
      accessValue: 'https://api.smirel.com/v1',
      accessHint: 'One Base URL for multiple model providers',
      billing: 'Billing',
      billingValue: 'Usage based',
      billingHint: 'Input and output usage are recorded separately',
      record: 'Cost records',
      recordValue: 'Visible in console',
      recordHint: 'Requests, tokens, and costs in one place',
      provider: 'Model providers',
      modelFamily: 'Model family',
      api: 'API',
      inputPrice: 'Input price',
      outputPrice: 'Output price',
      useCase: 'Best for',
      status: 'Billing status',
      liveRate: 'Live rate',
      noteTitle: 'Pricing note',
      note: 'This page shows the billing structure without hard-coding upstream prices that may change. Final charges follow the settlement data at request time.',
    })

const providers = computed<ProviderCatalog[]>(() => isZh.value
  ? [
      {
        key: 'openai',
        index: '01',
        name: 'OpenAI',
        protocol: 'Responses / Chat Completions',
        summary: '覆盖通用对话、推理与多模态调用。',
        models: [
          { family: 'GPT 通用模型', api: 'Responses / Chat Completions', input: '实时费率', output: '实时费率', note: '通用任务' },
          { family: 'GPT 轻量模型', api: 'Responses / Chat Completions', input: '实时费率', output: '实时费率', note: '高频调用' },
          { family: '推理模型', api: 'Responses API', input: '实时费率', output: '实时费率', note: '复杂推理' },
        ],
      },
      {
        key: 'anthropic',
        index: '02',
        name: 'Anthropic',
        protocol: 'Messages API',
        summary: '面向长文本、代码与复杂工作流。',
        models: [
          { family: 'Claude 高性能模型', api: 'Messages API', input: '实时费率', output: '实时费率', note: '复杂任务' },
          { family: 'Claude 通用模型', api: 'Messages API', input: '实时费率', output: '实时费率', note: '日常生产' },
          { family: 'Claude 轻量模型', api: 'Messages API', input: '实时费率', output: '实时费率', note: '低延迟' },
        ],
      },
      {
        key: 'google',
        index: '03',
        name: 'Google',
        protocol: 'Gemini API',
        summary: '覆盖 Gemini 通用与多模态模型。',
        models: [
          { family: 'Gemini Pro 系列', api: 'Gemini API', input: '实时费率', output: '实时费率', note: '高质量生成' },
          { family: 'Gemini Flash 系列', api: 'Gemini API', input: '实时费率', output: '实时费率', note: '快速响应' },
          { family: 'Gemini 多模态系列', api: 'Gemini API', input: '实时费率', output: '实时费率', note: '多模态任务' },
        ],
      },
    ]
  : [
      {
        key: 'openai',
        index: '01',
        name: 'OpenAI',
        protocol: 'Responses / Chat Completions',
        summary: 'General chat, reasoning, and multimodal workloads.',
        models: [
          { family: 'GPT general models', api: 'Responses / Chat Completions', input: 'Live rate', output: 'Live rate', note: 'General workloads' },
          { family: 'GPT lightweight models', api: 'Responses / Chat Completions', input: 'Live rate', output: 'Live rate', note: 'High-frequency calls' },
          { family: 'Reasoning models', api: 'Responses API', input: 'Live rate', output: 'Live rate', note: 'Complex reasoning' },
        ],
      },
      {
        key: 'anthropic',
        index: '02',
        name: 'Anthropic',
        protocol: 'Messages API',
        summary: 'Long-context, coding, and complex workflow workloads.',
        models: [
          { family: 'Claude high-performance', api: 'Messages API', input: 'Live rate', output: 'Live rate', note: 'Complex workloads' },
          { family: 'Claude general', api: 'Messages API', input: 'Live rate', output: 'Live rate', note: 'Daily production' },
          { family: 'Claude lightweight', api: 'Messages API', input: 'Live rate', output: 'Live rate', note: 'Low latency' },
        ],
      },
      {
        key: 'google',
        index: '03',
        name: 'Google',
        protocol: 'Gemini API',
        summary: 'Gemini general-purpose and multimodal model access.',
        models: [
          { family: 'Gemini Pro family', api: 'Gemini API', input: 'Live rate', output: 'Live rate', note: 'High-quality generation' },
          { family: 'Gemini Flash family', api: 'Gemini API', input: 'Live rate', output: 'Live rate', note: 'Fast response' },
          { family: 'Gemini multimodal', api: 'Gemini API', input: 'Live rate', output: 'Live rate', note: 'Multimodal tasks' },
        ],
      },
    ])

const activeProviderData = computed(() => providers.value.find((provider) => provider.key === activeProvider.value) || providers.value[0])
</script>

<template>
  <section class="workspace-page model-catalog-page">
    <header class="page-heading model-catalog-heading">
      <div>
        <div class="model-catalog-kicker"><i aria-hidden="true"></i><span>{{ copy.eyebrow }}</span></div>
        <h1>{{ copy.title }}</h1>
        <p>{{ copy.description }}</p>
      </div>
    </header>

    <section class="model-catalog-overview" aria-label="Model access overview">
      <article>
        <span>{{ copy.access }}</span>
        <strong><code>{{ copy.accessValue }}</code></strong>
        <small>{{ copy.accessHint }}</small>
      </article>
      <article>
        <span>{{ copy.billing }}</span>
        <strong>{{ copy.billingValue }}</strong>
        <small>{{ copy.billingHint }}</small>
      </article>
      <article>
        <span>{{ copy.record }}</span>
        <strong>{{ copy.recordValue }}</strong>
        <small>{{ copy.recordHint }}</small>
      </article>
    </section>

    <section class="model-provider-section">
      <header class="model-section-heading">
        <div>
          <span>{{ copy.provider }}</span>
          <strong>{{ activeProviderData.name }}</strong>
        </div>
      </header>

      <div class="model-provider-tabs" role="tablist" :aria-label="copy.provider">
        <button
          v-for="provider in providers"
          :key="provider.key"
          type="button"
          role="tab"
          :class="{ active: activeProvider === provider.key }"
          :aria-selected="activeProvider === provider.key"
          @click="activeProvider = provider.key"
        >
          <b>{{ provider.index }}</b>
          <span>
            <strong>{{ provider.name }}</strong>
            <small>{{ provider.protocol }}</small>
          </span>
        </button>
      </div>
    </section>

    <section class="model-catalog-panel">
      <header class="model-catalog-panel-head">
        <div>
          <span>{{ activeProviderData.name.toUpperCase() }}</span>
          <h2>{{ activeProviderData.name }}</h2>
          <p>{{ activeProviderData.summary }}</p>
        </div>
        <div class="model-rate-status">
          <i aria-hidden="true"></i>
          <span>{{ copy.status }}</span>
          <strong>{{ copy.liveRate }}</strong>
        </div>
      </header>

      <div class="model-catalog-table-wrap">
        <table class="model-catalog-table">
          <thead>
            <tr>
              <th>{{ copy.modelFamily }}</th>
              <th>{{ copy.api }}</th>
              <th>{{ copy.inputPrice }}</th>
              <th>{{ copy.outputPrice }}</th>
              <th>{{ copy.useCase }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="model in activeProviderData.models" :key="model.family">
              <td><strong>{{ model.family }}</strong></td>
              <td><code>{{ model.api }}</code></td>
              <td><span class="model-rate-chip">{{ model.input }}</span></td>
              <td><span class="model-rate-chip">{{ model.output }}</span></td>
              <td>{{ model.note }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="model-pricing-note">
      <span>{{ copy.noteTitle }}</span>
      <strong>{{ copy.note }}</strong>
    </section>
  </section>
</template>
