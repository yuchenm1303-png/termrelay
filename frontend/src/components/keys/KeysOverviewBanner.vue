<template>
  <section class="keys-overview" aria-labelledby="keys-overview-title">
    <div class="keys-overview-copy">
      <p class="keys-overview-kicker">TERMRELAY / KEY CONTROL</p>
      <h2 id="keys-overview-title">{{ copy.title }}</h2>
      <p class="keys-overview-description">{{ copy.description }}</p>

      <div class="keys-overview-actions">
        <button type="button" class="keys-overview-primary" @click="openCreateDialog">
          <Icon name="plus" size="sm" />
          <span>{{ copy.create }}</span>
        </button>
        <router-link to="/usage" class="keys-overview-secondary">
          <Icon name="chart" size="sm" />
          <span>{{ copy.usage }}</span>
        </router-link>
      </div>
    </div>

    <div class="keys-overview-map" aria-hidden="true">
      <span class="keys-map-orbit keys-map-orbit-a"></span>
      <span class="keys-map-orbit keys-map-orbit-b"></span>
      <span class="keys-map-core"></span>
      <span class="keys-map-node keys-map-node-a"></span>
      <span class="keys-map-node keys-map-node-b"></span>
      <span class="keys-map-node keys-map-node-c"></span>
      <span class="keys-map-line keys-map-line-a"></span>
      <span class="keys-map-line keys-map-line-b"></span>
    </div>

    <div class="keys-overview-specs">
      <div>
        <span>{{ copy.endpoint }}</span>
        <strong>/v1</strong>
      </div>
      <div>
        <span>{{ copy.authentication }}</span>
        <strong>Bearer</strong>
      </div>
      <div>
        <span>{{ copy.policy }}</span>
        <strong>{{ copy.policyValue }}</strong>
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
        title: 'API Key 控制中心',
        description:
          '为不同设备、客户端和使用者分配独立凭证，并分别控制分组、额度、有效期与访问来源。',
        create: '创建 API Key',
        usage: '查看调用记录',
        endpoint: '统一入口',
        authentication: '认证方式',
        policy: '可用策略',
        policyValue: '配额 / IP / 到期'
      }
    : {
        title: 'API key control center',
        description:
          'Issue isolated credentials for clients and operators, then control group routing, quotas, expiration and access origin per key.',
        create: 'Create API key',
        usage: 'View usage records',
        endpoint: 'Unified endpoint',
        authentication: 'Authentication',
        policy: 'Available policy',
        policyValue: 'Quota / IP / expiry'
      }
)

function openCreateDialog() {
  const button = document.querySelector<HTMLButtonElement>('[data-tour="keys-create-btn"]')
  button?.click()
}
</script>
