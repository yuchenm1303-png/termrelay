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
.billing-page {
  width: 100%;
  max-width: 1260px;
  margin: 0 auto;
  padding-bottom: 48px;
  color: #eef1f4;
}

.billing-mode {
  width: min(720px, 100%);
  min-height: 50px;
  margin: 0 0 22px;
  padding: 4px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  border: 1px solid #262c35;
  border-radius: 12px;
  background: #0b0d11;
  box-shadow: inset 0 1px rgba(255,255,255,.018);
}

.billing-mode button {
  min-width: 0;
  height: 40px;
  border: 0;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  background: transparent;
  color: #949da8;
  font: inherit;
  font-size: .88rem;
  font-weight: 650;
  cursor: pointer;
  transition: background-color .15s ease, color .15s ease, box-shadow .15s ease;
}

.billing-mode button:hover { color: #d8dde3; }
.billing-mode button.active {
  background: #161b22;
  color: #f5f7f9;
  box-shadow: inset 0 0 0 1px #303844, 0 1px 4px rgba(0,0,0,.22);
}
.billing-mode :deep(.workspace-nav-icon) { width: 17px; height: 17px; color: currentColor; }

.billing-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.68fr) minmax(310px, .74fr);
  gap: 18px;
  align-items: start;
}
.billing-main, .billing-side { display: grid; gap: 18px; }
.billing-side { position: sticky; top: 86px; }

.billing-card {
  border: 1px solid #252b33;
  border-radius: 14px;
  background: #0f1115;
  box-shadow: inset 0 1px rgba(255,255,255,.018), 0 8px 28px rgba(0,0,0,.08);
}

.balance-card {
  min-height: 150px;
  padding: 27px 28px;
  display: flex;
  align-items: center;
  gap: 20px;
  background: linear-gradient(112deg, #111821 0%, #0f1217 58%, #0e1013 100%);
  border-color: #2c3744;
}
.balance-icon {
  width: 50px;
  height: 50px;
  flex: 0 0 auto;
  border: 1px solid #315579;
  border-radius: 13px;
  display: grid;
  place-items: center;
  color: #67b8f2;
  background: #12253c;
  box-shadow: inset 0 1px rgba(255,255,255,.025);
}
.balance-icon :deep(.workspace-nav-icon) { width: 22px; height: 22px; }
.balance-copy { min-width: 0; }
.balance-label {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #a5adb7;
  font-size: .82rem;
  font-weight: 620;
}
.balance-label i {
  padding: 5px 8px;
  border-radius: 99px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(66,190,141,.10);
  color: #68d0a8;
  font-style: normal;
  font-size: .71rem;
  font-weight: 650;
}
.balance-label i b { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
.balance-copy > strong {
  display: block;
  margin-top: 9px;
  color: #f7f9fb;
  font-size: 2.25rem;
  line-height: 1;
  font-weight: 700;
  letter-spacing: -.045em;
}
.balance-copy p { margin: 10px 0 0; color: #8a949f; font-size: .79rem; line-height: 1.45; }
.balance-orders {
  margin-left: auto;
  height: 40px;
  padding: 0 14px;
  border: 1px solid #303741;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #c0c7cf;
  background: #12161b;
  font-size: .79rem;
  font-weight: 620;
  transition: color .15s ease, border-color .15s ease, background-color .15s ease;
}
.balance-orders:hover { color: #f4f6f8; border-color: #414b57; background: #171c22; }
.balance-orders span { color: #7b8590; font-size: .88rem; }

.amount-card, .payment-card { padding: 26px 27px 27px; }
.billing-section-head {
  margin-bottom: 21px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}
.billing-section-head h2 {
  margin: 0;
  color: #edf0f3;
  font-size: 1.02rem;
  font-weight: 670;
  letter-spacing: -.015em;
}
.billing-section-head p {
  margin: 7px 0 0;
  color: #87919c;
  font-size: .78rem;
  line-height: 1.55;
}
.section-index {
  color: #66717c;
  font: 650 .70rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .10em;
}

.amount-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.amount-grid button {
  height: 52px;
  border: 1px solid #303741;
  border-radius: 10px;
  background: #13171c;
  color: #d7dce1;
  font: inherit;
  font-size: .94rem;
  font-weight: 670;
  cursor: pointer;
  transition: border-color .15s ease, background-color .15s ease, color .15s ease, transform .15s ease;
}
.amount-grid button:hover { border-color: #46515d; background: #171c22; color: #fff; transform: translateY(-1px); }
.amount-grid button.selected {
  border-color: #3b8ed3;
  background: #112943;
  color: #78c6f8;
  box-shadow: inset 0 0 0 1px rgba(75,167,234,.15);
}
.amount-grid button span { margin-right: 3px; color: #87919b; font-size: .80rem; }
.amount-grid button.selected span { color: #5ba8df; }

.custom-amount { margin-top: 20px; display: block; }
.custom-amount > span {
  display: block;
  margin-bottom: 9px;
  color: #a1aab4;
  font-size: .79rem;
  font-weight: 620;
}
.custom-amount > div {
  height: 48px;
  border: 1px solid #303741;
  border-radius: 10px;
  display: flex;
  align-items: center;
  background: #12161b;
  overflow: hidden;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.custom-amount > div:focus-within, .custom-amount > div.focused {
  border-color: #4779a4;
  box-shadow: 0 0 0 3px rgba(63,126,181,.08);
}
.custom-amount b { padding-left: 15px; color: #88939f; font-size: .86rem; }
.custom-amount input {
  width: 100%;
  height: 100%;
  padding: 0 15px 0 8px;
  border: 0;
  outline: 0;
  background: transparent;
  color: #edf0f3;
  font: inherit;
  font-size: .88rem;
}
.custom-amount input::placeholder { color: #6d7680; }

.payment-option {
  min-height: 66px;
  padding: 12px 14px;
  border: 1px solid #303741;
  border-radius: 11px;
  display: flex;
  align-items: center;
  gap: 13px;
  background: #12161b;
}
.payment-mark {
  width: 40px;
  height: 40px;
  border: 1px solid #37404b;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #a2acb7;
  background: #181d23;
}
.payment-mark :deep(.workspace-nav-icon) { width: 18px; height: 18px; }
.payment-option div { min-width: 0; }
.payment-option strong { display: block; color: #e0e4e8; font-size: .87rem; font-weight: 650; }
.payment-option small { display: block; margin-top: 4px; color: #87919c; font-size: .72rem; }
.payment-option > i {
  margin-left: auto;
  padding: 6px 8px;
  border-radius: 7px;
  background: #1b2027;
  color: #8e98a3;
  font-style: normal;
  font-size: .68rem;
  font-weight: 620;
}
.security-note {
  margin: 13px 0 0;
  display: flex;
  align-items: flex-start;
  gap: 8px;
  color: #7f8994;
  font-size: .73rem;
  line-height: 1.58;
}
.security-note :deep(.workspace-nav-icon) { width: 14px; height: 14px; flex: 0 0 auto; margin-top: 2px; }

.summary-card { padding: 24px 23px 21px; }
.summary-card > header {
  min-height: 38px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  color: #edf0f3;
  font-size: .95rem;
  font-weight: 670;
}
.summary-card > header :deep(.workspace-nav-icon) { width: 19px; height: 19px; color: #7f8994; }
.summary-card dl { margin: 17px 0 0; }
.summary-card dl > div {
  min-height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  color: #9aa3ad;
  font-size: .80rem;
}
.summary-card dd { margin: 0; color: #d8dde2; font-weight: 650; }
.summary-card .summary-total {
  min-height: 62px;
  margin-top: 9px;
  padding-top: 15px;
  border-top: 1px solid #282f37;
  color: #c1c8cf;
}
.summary-card .summary-total dd { color: #6ec0f5; font-size: 1.55rem; font-weight: 700; letter-spacing: -.035em; }
.summary-card > button {
  width: 100%;
  height: 46px;
  margin-top: 16px;
  border: 1px solid #2d3945;
  border-radius: 10px;
  background: #151b22;
  color: #8190a0;
  font: inherit;
  font-size: .80rem;
  font-weight: 650;
  cursor: not-allowed;
}
.summary-card > p { margin: 12px 1px 0; color: #77818c; font-size: .72rem; line-height: 1.62; }

.billing-note { padding: 19px; display: flex; align-items: flex-start; gap: 13px; }
.note-icon {
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: #171c22;
  color: #8995a1;
}
.note-icon :deep(.workspace-nav-icon) { width: 17px; height: 17px; }
.billing-note strong { color: #cbd1d7; font-size: .81rem; font-weight: 650; }
.billing-note p { margin: 6px 0 0; color: #808a95; font-size: .72rem; line-height: 1.62; }

.plan-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 18px; }
.plan-card { min-height: 390px; padding: 28px; }
.plan-card.current {
  border-color: #314154;
  background: linear-gradient(145deg, #121a24, #0f1115 55%);
}
.plan-card header { display: flex; align-items: center; justify-content: space-between; }
.plan-icon {
  width: 43px;
  height: 43px;
  border: 1px solid #313943;
  border-radius: 11px;
  display: grid;
  place-items: center;
  color: #98a3ae;
  background: #171c22;
}
.plan-card.current .plan-icon { border-color: #315579; color: #6ab8ed; background: #12253c; }
.plan-icon :deep(.workspace-nav-icon) { width: 19px; height: 19px; }
.plan-card header i {
  padding: 6px 9px;
  border-radius: 99px;
  background: #1a1f25;
  color: #8d97a2;
  font-style: normal;
  font-size: .68rem;
  font-weight: 650;
}
.plan-card.current header i { background: rgba(64,187,140,.10); color: #68d0a8; }
.plan-kicker {
  margin-top: 34px;
  color: #818b96;
  font: 700 .67rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .13em;
}
.plan-card h2 { margin: 11px 0 0; color: #f1f3f5; font-size: 1.62rem; line-height: 1.15; letter-spacing: -.035em; }
.plan-card > strong { display: block; margin-top: 11px; color: #abb4be; font-size: .88rem; font-weight: 650; }
.plan-card > p { max-width: 460px; margin: 16px 0 0; color: #89939e; font-size: .79rem; line-height: 1.7; }
.plan-rule { height: 1px; margin: 28px 0 21px; background: #292f36; }
.plan-card ul { margin: 0; padding: 0; display: grid; gap: 14px; list-style: none; }
.plan-card li { display: flex; align-items: center; gap: 10px; color: #a1aab4; font-size: .79rem; }
.plan-card li b {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: #1a2027;
  color: #7f8d9b;
  font-size: .65rem;
}
.plan-card.current li b { color: #62b4ed; background: #12263d; }

@media (max-width: 980px) {
  .billing-grid { grid-template-columns: 1fr; }
  .billing-side { position: static; grid-template-columns: 1fr 1fr; }
}

@media (max-width: 720px) {
  .billing-page { padding-bottom: 30px; }
  .billing-mode { width: 100%; }
  .billing-mode button { font-size: .84rem; }
  .balance-card { padding: 22px; align-items: flex-start; flex-wrap: wrap; }
  .balance-copy > strong { font-size: 2rem; }
  .balance-orders { margin-left: 70px; }
  .amount-card, .payment-card { padding: 22px; }
  .amount-grid { grid-template-columns: repeat(2, 1fr); }
  .billing-side, .plan-grid { grid-template-columns: 1fr; }
  .plan-card { min-height: 0; padding: 24px; }
}
</style>
