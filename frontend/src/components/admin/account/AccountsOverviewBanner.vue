<template>
  <section class="accounts-overview" aria-labelledby="accounts-overview-title">
    <div class="accounts-overview-copy">
      <p class="accounts-overview-kicker">TERMRELAY / UPSTREAM CONTROL</p>
      <h2 id="accounts-overview-title">{{ copy.title }}</h2>
      <p class="accounts-overview-description">{{ copy.description }}</p>

      <div class="accounts-overview-actions">
        <button type="button" class="accounts-overview-primary" @click="openCreateDialog">
          <Icon name="plus" size="sm" />
          <span>{{ copy.connect }}</span>
        </button>
        <router-link to="/keys" class="accounts-overview-secondary">
          <Icon name="key" size="sm" />
          <span>{{ copy.keys }}</span>
        </router-link>
      </div>
    </div>

    <div class="accounts-flow" aria-hidden="true">
      <div class="accounts-flow-node accounts-flow-node-source">
        <span class="accounts-flow-icon">OA</span>
        <div>
          <small>UPSTREAM</small>
          <strong>OpenAI OAuth</strong>
        </div>
      </div>

      <span class="accounts-flow-link accounts-flow-link-a"></span>

      <div class="accounts-flow-node accounts-flow-node-gateway">
        <span class="accounts-flow-core"></span>
        <div>
          <small>GATEWAY</small>
          <strong>TermRelay</strong>
        </div>
      </div>

      <span class="accounts-flow-link accounts-flow-link-b"></span>

      <div class="accounts-flow-node accounts-flow-node-client">
        <span class="accounts-flow-icon">API</span>
        <div>
          <small>DOWNSTREAM</small>
          <strong>Bearer Key</strong>
        </div>
      </div>
    </div>

    <div class="accounts-overview-steps">
      <div>
        <span>01</span>
        <p>{{ copy.stepOne }}</p>
      </div>
      <div>
        <span>02</span>
        <p>{{ copy.stepTwo }}</p>
      </div>
      <div>
        <span>03</span>
        <p>{{ copy.stepThree }}</p>
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
        title: '上游账号接入中心',
        description:
          '连接并维护模型服务账号，统一查看认证状态、调度能力、额度窗口与健康检查。首个节点建议使用 OpenAI OAuth。',
        connect: '连接上游账号',
        keys: '管理下游 Key',
        stepOne: '选择 OpenAI 与 OAuth 认证方式',
        stepTwo: '完成授权并进行连通性测试',
        stepThree: '创建下游 API Key 开始调用'
      }
    : {
        title: 'Upstream account control',
        description:
          'Connect and maintain model provider accounts with centralized authentication health, scheduling capacity, quota windows and connectivity checks. OpenAI OAuth is recommended for the first node.',
        connect: 'Connect upstream account',
        keys: 'Manage downstream keys',
        stepOne: 'Choose OpenAI and OAuth authentication',
        stepTwo: 'Authorize and verify connectivity',
        stepThree: 'Issue a downstream API key'
      }
)

function openCreateDialog() {
  const button = document.querySelector<HTMLButtonElement>('[data-tour="accounts-create-btn"]')
  button?.click()
}
</script>
