<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import WorkspaceNavIcon from './WorkspaceNavIcon.vue'

const props = defineProps<{ balance: number }>()
const { locale } = useI18n()

const mode = ref<'recharge' | 'subscription'>('recharge')
const selectedAmount = ref<number | null>(50)
const customAmount = ref('')
const presetAmounts = [10, 20, 50, 100, 200, 500]

const copy = computed(() => locale.value === 'zh-CN' ? {
  recharge: '按量充值',
  subscription: '订阅套餐',
  balance: '当前余额',
  balanceHint: '用于 API 调用与服务购买',
  normal: '正常',
  orders: '交易记录',
  amountTitle: '选择充值金额',
  amountHint: '充值后余额将计入当前账户',
  custom: '自定义金额',
  customPlaceholder: '输入充值金额',
  payment: '支付方式',
  paymentHint: '支付通道接入完成后即可在线充值',
  gateway: '在线支付',
  pending: '接入中',
  secure: '支付信息将由支付服务商安全处理，Smirel 不保存完整卡片信息。',
  summary: '充值确认',
  rechargeAmount: '充值金额',
  fee: '支付手续费',
  total: '预计入账',
  unavailable: '支付通道接入中',
  unavailableHint: '当前页面已完成商业版充值流程设计，支付接口尚未开放。',
  currentPlan: '当前方案',
  payg: '按量计费',
  paygPrice: '无月费',
  paygHint: '按实际 API 用量从账户余额中扣除费用。',
  active: '使用中',
  planTitle: '订阅套餐',
  planPrice: '即将开放',
  planHint: '套餐权益、周期计费与自动续费将在计费接口接入后开放。',
  planFeatureOne: '固定周期额度',
  planFeatureTwo: '统一账单管理',
  planFeatureThree: '套餐权益升级',
  preparing: '正在准备',
  helpTitle: '计费方式',
  helpCopy: '按量充值适合灵活使用；订阅套餐适合固定周期与稳定额度需求。',
} : {
  recharge: 'Recharge',
  subscription: 'Subscription',
  balance: 'Current balance',
  balanceHint: 'Available for API usage and service purchases',
  normal: 'Active',
  orders: 'Transactions',
  amountTitle: 'Choose an amount',
  amountHint: 'Funds will be added to the current account',
  custom: 'Custom amount',
  customPlaceholder: 'Enter amount',
  payment: 'Payment method',
  paymentHint: 'Online recharge becomes available when the payment gateway is connected',
  gateway: 'Online payment',
  pending: 'Connecting',
  secure: 'Payment details will be handled securely by the payment provider. Smirel will not store complete card data.',
  summary: 'Recharge summary',
  rechargeAmount: 'Recharge amount',
  fee: 'Processing fee',
  total: 'Expected credit',
  unavailable: 'Payment gateway connecting',
  unavailableHint: 'The commercial recharge flow is ready; the payment API is not enabled yet.',
  currentPlan: 'Current plan',
  payg: 'Pay as you go',
  paygPrice: 'No monthly fee',
  paygHint: 'API charges are deducted from your account balance based on actual usage.',
  active: 'Active',
  planTitle: 'Subscription plans',
  planPrice: 'Coming soon',
  planHint: 'Plan benefits, recurring billing, and renewals will become available after billing integration.',
  planFeatureOne: 'Recurring quota',
  planFeatureTwo: 'Unified billing',
  planFeatureThree: 'Plan benefits',
  preparing: 'Preparing',
  helpTitle: 'Billing options',
  helpCopy: 'Recharge for flexible usage, or choose a subscription when you need a predictable recurring quota.',
})

const effectiveAmount = computed(() => {
  const custom = Number(customAmount.value)
  if (customAmount.value.trim() && Number.isFinite(custom) && custom > 0) return custom
  return selectedAmount.value || 0
})

function chooseAmount(amount: number) {
  selectedAmount.value = amount
  customAmount.value = ''
}

function useCustomAmount() {
  selectedAmount.value = null
}
</script>

<template>
  <div class="billing-page">
    <div class="billing-mode" role="tablist" :aria-label="copy.helpTitle">
      <button type="button" :class="{ active: mode === 'recharge' }" @click="mode = 'recharge'">
        <WorkspaceNavIcon name="wallet" />
        <span>{{ copy.recharge }}</span>
      </button>
      <button type="button" :class="{ active: mode === 'subscription' }" @click="mode = 'subscription'">
        <WorkspaceNavIcon name="credit-card" />
        <span>{{ copy.subscription }}</span>
      </button>
    </div>

    <template v-if="mode === 'recharge'">
      <div class="billing-grid">
        <main class="billing-main">
          <section class="balance-card billing-card">
            <div class="balance-icon"><WorkspaceNavIcon name="wallet" /></div>
            <div class="balance-copy">
              <div class="balance-label">
                <span>{{ copy.balance }}</span>
                <i><b></b>{{ copy.normal }}</i>
              </div>
              <strong>${{ Number(props.balance || 0).toFixed(2) }}</strong>
              <p>{{ copy.balanceHint }}</p>
            </div>
            <RouterLink class="balance-orders" to="/orders">
              {{ copy.orders }} <span>→</span>
            </RouterLink>
          </section>

          <section class="billing-card amount-card">
            <header class="billing-section-head">
              <div>
                <h2>{{ copy.amountTitle }}</h2>
                <p>{{ copy.amountHint }}</p>
              </div>
              <span class="section-index">01</span>
            </header>

            <div class="amount-grid">
              <button
                v-for="amount in presetAmounts"
                :key="amount"
                type="button"
                :class="{ selected: selectedAmount === amount && !customAmount }"
                @click="chooseAmount(amount)"
              >
                <span>$</span>{{ amount }}
              </button>
            </div>

            <label class="custom-amount">
              <span>{{ copy.custom }}</span>
              <div :class="{ focused: selectedAmount === null }">
                <b>$</b>
                <input
                  v-model="customAmount"
                  type="number"
                  inputmode="decimal"
                  min="1"
                  step="1"
                  :placeholder="copy.customPlaceholder"
                  @focus="useCustomAmount"
                />
              </div>
            </label>
          </section>

          <section class="billing-card payment-card">
            <header class="billing-section-head">
              <div>
                <h2>{{ copy.payment }}</h2>
                <p>{{ copy.paymentHint }}</p>
              </div>
              <span class="section-index">02</span>
            </header>

            <div class="payment-option" aria-disabled="true">
              <span class="payment-mark"><WorkspaceNavIcon name="credit-card" /></span>
              <div><strong>{{ copy.gateway }}</strong><small>{{ copy.pending }}</small></div>
              <i>{{ copy.pending }}</i>
            </div>
            <p class="security-note"><WorkspaceNavIcon name="shield" />{{ copy.secure }}</p>
          </section>
        </main>

        <aside class="billing-side">
          <section class="billing-card summary-card">
            <header><span>{{ copy.summary }}</span><WorkspaceNavIcon name="receipt" /></header>
            <dl>
              <div><dt>{{ copy.rechargeAmount }}</dt><dd>${{ effectiveAmount.toFixed(2) }}</dd></div>
              <div><dt>{{ copy.fee }}</dt><dd>—</dd></div>
              <div class="summary-total"><dt>{{ copy.total }}</dt><dd>${{ effectiveAmount.toFixed(2) }}</dd></div>
            </dl>
            <button type="button" disabled>{{ copy.unavailable }}</button>
            <p>{{ copy.unavailableHint }}</p>
          </section>

          <section class="billing-card billing-note">
            <span class="note-icon"><WorkspaceNavIcon name="activity" /></span>
            <div><strong>{{ copy.helpTitle }}</strong><p>{{ copy.helpCopy }}</p></div>
          </section>
        </aside>
      </div>
    </template>

    <template v-else>
      <div class="plan-grid">
        <section class="billing-card plan-card current">
          <header>
            <span class="plan-icon"><WorkspaceNavIcon name="activity" /></span>
            <i>{{ copy.active }}</i>
          </header>
          <div class="plan-kicker">{{ copy.currentPlan }}</div>
          <h2>{{ copy.payg }}</h2>
          <strong>{{ copy.paygPrice }}</strong>
          <p>{{ copy.paygHint }}</p>
          <div class="plan-rule"></div>
          <ul>
            <li><b>✓</b>{{ copy.balanceHint }}</li>
            <li><b>✓</b>{{ copy.orders }}</li>
          </ul>
        </section>

        <section class="billing-card plan-card upcoming">
          <header>
            <span class="plan-icon"><WorkspaceNavIcon name="credit-card" /></span>
            <i>{{ copy.preparing }}</i>
          </header>
          <div class="plan-kicker">SMIREL PLAN</div>
          <h2>{{ copy.planTitle }}</h2>
          <strong>{{ copy.planPrice }}</strong>
          <p>{{ copy.planHint }}</p>
          <div class="plan-rule"></div>
          <ul>
            <li><b>✓</b>{{ copy.planFeatureOne }}</li>
            <li><b>✓</b>{{ copy.planFeatureTwo }}</li>
            <li><b>✓</b>{{ copy.planFeatureThree }}</li>
          </ul>
        </section>
      </div>
    </template>
  </div>
</template>

<style scoped>
.billing-page { width: 100%; max-width: 1240px; margin: 0 auto; padding-bottom: 42px; color: #eef1f4; }
.billing-mode { width: min(680px, 100%); min-height: 44px; margin: 0 0 18px; padding: 3px; display: grid; grid-template-columns: 1fr 1fr; gap: 3px; border: 1px solid #242932; border-radius: 11px; background: #0b0d11; }
.billing-mode button { min-width: 0; height: 36px; border: 0; border-radius: 8px; display: flex; align-items: center; justify-content: center; gap: 8px; background: transparent; color: #77808b; font: inherit; font-size: .78rem; font-weight: 650; cursor: pointer; transition: background-color .15s ease, color .15s ease, box-shadow .15s ease; }
.billing-mode button:hover { color: #c9ced5; }
.billing-mode button.active { background: #151a22; color: #f3f5f7; box-shadow: inset 0 0 0 1px #2c3542; }
.billing-mode :deep(.workspace-nav-icon) { width: 15px; height: 15px; color: currentColor; }
.billing-grid { display: grid; grid-template-columns: minmax(0, 1.65fr) minmax(285px, .72fr); gap: 14px; align-items: start; }
.billing-main, .billing-side { display: grid; gap: 14px; }
.billing-side { position: sticky; top: 86px; }
.billing-card { border: 1px solid #22272f; border-radius: 13px; background: #0e1014; box-shadow: inset 0 1px rgba(255,255,255,.012); }
.balance-card { min-height: 132px; padding: 23px 24px; display: flex; align-items: center; gap: 17px; background: linear-gradient(110deg, #10151d 0%, #0d1015 58%, #0d0f12 100%); border-color: #27313f; }
.balance-icon { width: 46px; height: 46px; flex: 0 0 auto; border: 1px solid #29496a; border-radius: 12px; display: grid; place-items: center; color: #58adf0; background: #11233a; }
.balance-icon :deep(.workspace-nav-icon) { width: 20px; height: 20px; }
.balance-copy { min-width: 0; }
.balance-label { display: flex; align-items: center; gap: 9px; color: #89929d; font-size: .72rem; font-weight: 600; }
.balance-label i { padding: 4px 7px; border-radius: 99px; display: inline-flex; align-items: center; gap: 5px; background: rgba(66,190,141,.08); color: #59c89d; font-style: normal; font-size: .61rem; }
.balance-label i b { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.balance-copy > strong { display: block; margin-top: 7px; color: #f4f7fa; font-size: 2rem; line-height: 1; letter-spacing: -.045em; }
.balance-copy p { margin: 8px 0 0; color: #69727d; font-size: .68rem; }
.balance-orders { margin-left: auto; height: 36px; padding: 0 12px; border: 1px solid #2a3039; border-radius: 9px; display: inline-flex; align-items: center; gap: 9px; color: #aab2bc; background: #111419; font-size: .69rem; transition: color .15s ease, border-color .15s ease, background-color .15s ease; }
.balance-orders:hover { color: #f2f4f6; border-color: #39424d; background: #15191f; }
.balance-orders span { color: #68727e; }
.amount-card, .payment-card { padding: 22px 23px 23px; }
.billing-section-head { margin-bottom: 18px; display: flex; align-items: flex-start; justify-content: space-between; gap: 18px; }
.billing-section-head h2 { margin: 0; color: #e7eaee; font-size: .88rem; font-weight: 650; letter-spacing: -.01em; }
.billing-section-head p { margin: 6px 0 0; color: #68717c; font-size: .66rem; line-height: 1.45; }
.section-index { color: #434b55; font: 650 .62rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .08em; }
.amount-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.amount-grid button { height: 46px; border: 1px solid #2a3038; border-radius: 9px; background: #12151a; color: #ccd1d7; font: inherit; font-size: .82rem; font-weight: 650; cursor: pointer; transition: border-color .15s ease, background-color .15s ease, color .15s ease; }
.amount-grid button:hover { border-color: #3a4552; background: #151920; color: #fff; }
.amount-grid button.selected { border-color: #347ec0; background: #10233a; color: #67b7f1; box-shadow: inset 0 0 0 1px rgba(69,155,224,.16); }
.amount-grid button span { margin-right: 2px; color: #68727d; font-size: .7rem; }
.amount-grid button.selected span { color: #4e98d2; }
.custom-amount { margin-top: 17px; display: block; }
.custom-amount > span { display: block; margin-bottom: 8px; color: #8f97a1; font-size: .68rem; font-weight: 600; }
.custom-amount > div { height: 42px; border: 1px solid #292f37; border-radius: 9px; display: flex; align-items: center; background: #111419; overflow: hidden; transition: border-color .15s ease; }
.custom-amount > div:focus-within, .custom-amount > div.focused { border-color: #3a6286; }
.custom-amount b { padding-left: 13px; color: #65707c; font-size: .75rem; }
.custom-amount input { width: 100%; height: 100%; padding: 0 13px 0 7px; border: 0; outline: 0; background: transparent; color: #e6e9ed; font: inherit; font-size: .76rem; }
.custom-amount input::placeholder { color: #515963; }
.payment-option { min-height: 58px; padding: 10px 12px; border: 1px solid #2b313a; border-radius: 10px; display: flex; align-items: center; gap: 11px; background: #111419; }
.payment-mark { width: 36px; height: 36px; border: 1px solid #303844; border-radius: 9px; display: grid; place-items: center; color: #8995a3; background: #161a20; }
.payment-mark :deep(.workspace-nav-icon) { width: 16px; height: 16px; }
.payment-option div { min-width: 0; }
.payment-option strong { display: block; color: #d6dae0; font-size: .76rem; }
.payment-option small { display: block; margin-top: 3px; color: #626b76; font-size: .61rem; }
.payment-option > i { margin-left: auto; padding: 5px 7px; border-radius: 6px; background: #181d23; color: #68727d; font-style: normal; font-size: .58rem; }
.security-note { margin: 11px 0 0; display: flex; align-items: flex-start; gap: 7px; color: #5f6873; font-size: .62rem; line-height: 1.5; }
.security-note :deep(.workspace-nav-icon) { width: 13px; height: 13px; flex: 0 0 auto; margin-top: 1px; }
.summary-card { padding: 21px 20px 18px; }
.summary-card > header { min-height: 34px; display: flex; align-items: flex-start; justify-content: space-between; color: #dfe3e8; font-size: .82rem; font-weight: 650; }
.summary-card > header :deep(.workspace-nav-icon) { width: 17px; height: 17px; color: #5f6975; }
.summary-card dl { margin: 15px 0 0; }
.summary-card dl > div { min-height: 36px; display: flex; align-items: center; justify-content: space-between; gap: 14px; color: #7f8893; font-size: .68rem; }
.summary-card dd { margin: 0; color: #c9ced4; font-weight: 600; }
.summary-card .summary-total { min-height: 54px; margin-top: 8px; padding-top: 13px; border-top: 1px solid #222831; color: #aeb6bf; }
.summary-card .summary-total dd { color: #61b5f0; font-size: 1.35rem; letter-spacing: -.035em; }
.summary-card > button { width: 100%; height: 42px; margin-top: 14px; border: 1px solid #26313d; border-radius: 9px; background: #141a21; color: #64707c; font: inherit; font-size: .7rem; font-weight: 650; cursor: not-allowed; }
.summary-card > p { margin: 10px 1px 0; color: #59626d; font-size: .6rem; line-height: 1.55; }
.billing-note { padding: 16px; display: flex; align-items: flex-start; gap: 11px; }
.note-icon { width: 34px; height: 34px; flex: 0 0 auto; border-radius: 9px; display: grid; place-items: center; background: #15191f; color: #697582; }
.note-icon :deep(.workspace-nav-icon) { width: 15px; height: 15px; }
.billing-note strong { color: #b9c0c8; font-size: .69rem; }
.billing-note p { margin: 5px 0 0; color: #5e6772; font-size: .61rem; line-height: 1.55; }
.plan-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.plan-card { min-height: 360px; padding: 24px; }
.plan-card.current { border-color: #2b3847; background: linear-gradient(145deg, #111720, #0e1014 54%); }
.plan-card header { display: flex; align-items: center; justify-content: space-between; }
.plan-icon { width: 39px; height: 39px; border: 1px solid #2a313a; border-radius: 10px; display: grid; place-items: center; color: #87919d; background: #15191f; }
.plan-card.current .plan-icon { border-color: #29496a; color: #5eace7; background: #11233a; }
.plan-icon :deep(.workspace-nav-icon) { width: 17px; height: 17px; }
.plan-card header i { padding: 5px 8px; border-radius: 99px; background: #171b20; color: #6f7883; font-style: normal; font-size: .58rem; }
.plan-card.current header i { background: rgba(64,187,140,.08); color: #58c79a; }
.plan-kicker { margin-top: 31px; color: #626b75; font: 700 .59rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
.plan-card h2 { margin: 9px 0 0; color: #edf0f3; font-size: 1.45rem; letter-spacing: -.035em; }
.plan-card > strong { display: block; margin-top: 9px; color: #8f99a4; font-size: .76rem; font-weight: 600; }
.plan-card > p { max-width: 440px; margin: 14px 0 0; color: #68727d; font-size: .69rem; line-height: 1.65; }
.plan-rule { height: 1px; margin: 25px 0 19px; background: #22272e; }
.plan-card ul { margin: 0; padding: 0; display: grid; gap: 12px; list-style: none; }
.plan-card li { display: flex; align-items: center; gap: 9px; color: #8c959f; font-size: .68rem; }
.plan-card li b { width: 18px; height: 18px; border-radius: 50%; display: grid; place-items: center; background: #171c22; color: #63707d; font-size: .58rem; }
.plan-card.current li b { color: #52a9e9; background: #112338; }
@media (max-width: 980px) { .billing-grid { grid-template-columns: 1fr; } .billing-side { position: static; grid-template-columns: 1fr 1fr; } }
@media (max-width: 720px) { .billing-page { padding-bottom: 26px; } .billing-mode { width: 100%; } .balance-card { padding: 19px; align-items: flex-start; flex-wrap: wrap; } .balance-orders { margin-left: 63px; } .amount-card, .payment-card { padding: 19px; } .amount-grid { grid-template-columns: repeat(2, 1fr); } .billing-side, .plan-grid { grid-template-columns: 1fr; } .plan-card { min-height: 0; } }
</style>
