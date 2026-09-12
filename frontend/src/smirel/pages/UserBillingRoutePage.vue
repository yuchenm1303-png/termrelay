<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import UserBillingPage from '../components/UserBillingPage.vue'
import { useSession } from '../core/session'

const { locale } = useI18n()
const { state } = useSession()

const title = computed(() => locale.value === 'zh-CN' ? '充值/订阅' : 'Recharge / Subscription')
const description = computed(() => locale.value === 'zh-CN'
  ? '为账户充值余额，或查看订阅方案。'
  : 'Add account balance or review subscription options.')
const accountBalance = computed(() => Number(state.user?.balance || 0))
</script>

<template>
  <section class="workspace-page billing-route-page">
    <header class="page-heading">
      <div>
        <h1>{{ title }}</h1>
        <p>{{ description }}</p>
      </div>
    </header>
    <UserBillingPage :balance="accountBalance" />
  </section>
</template>
