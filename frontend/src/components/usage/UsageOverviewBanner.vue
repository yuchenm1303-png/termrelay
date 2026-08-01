<template>
  <section class="usage-overview" aria-labelledby="usage-overview-title">
    <div class="usage-overview-copy">
      <p class="usage-overview-kicker">TERMRELAY / REQUEST OBSERVABILITY</p>
      <h2 id="usage-overview-title">{{ copy.title }}</h2>
      <p class="usage-overview-description">{{ copy.description }}</p>

      <div class="usage-overview-actions">
        <router-link to="/keys" class="usage-overview-primary">
          <Icon name="key" size="sm" />
          <span>{{ copy.keys }}</span>
        </router-link>
        <router-link to="/dashboard" class="usage-overview-secondary">
          <Icon name="chart" size="sm" />
          <span>{{ copy.dashboard }}</span>
        </router-link>
      </div>
    </div>

    <div class="usage-signal" aria-hidden="true">
      <div class="usage-signal-head">
        <span>LIVE REQUEST TRACE</span>
        <i></i>
      </div>
      <div class="usage-signal-lanes">
        <span class="usage-lane usage-lane-a"></span>
        <span class="usage-lane usage-lane-b"></span>
        <span class="usage-lane usage-lane-c"></span>
        <span class="usage-lane usage-lane-d"></span>
      </div>
      <div class="usage-signal-foot">
        <span>CLIENT</span>
        <span>GATEWAY</span>
        <span>UPSTREAM</span>
      </div>
    </div>

    <div class="usage-overview-specs">
      <div>
        <span>{{ copy.trace }}</span>
        <strong>Request → Response</strong>
      </div>
      <div>
        <span>{{ copy.metrics }}</span>
        <strong>Tokens / Cost / Latency</strong>
      </div>
      <div>
        <span>{{ copy.errors }}</span>
        <strong>Status / Model / Key</strong>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { locale } = useI18n()

const copy = computed(() =>
  locale.value === 'zh'
    ? {
        title: '调用与请求观测中心',
        description:
          '从下游 API Key 到上游模型账号，统一追踪请求量、Token、成本、模型分布、端点路径和错误响应。',
        keys: '管理 API Key',
        dashboard: '返回控制台',
        trace: '请求链路',
        metrics: '核心指标',
        errors: '错误定位'
      }
    : {
        title: 'Request observability center',
        description:
          'Trace traffic from downstream API keys to upstream model accounts, including requests, tokens, cost, model distribution, endpoint paths and failures.',
        keys: 'Manage API keys',
        dashboard: 'Back to dashboard',
        trace: 'Request path',
        metrics: 'Core metrics',
        errors: 'Failure context'
      }
)
</script>
