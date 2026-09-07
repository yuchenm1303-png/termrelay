<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { previewMode } from '../core/api'
import WorkspaceNavIcon from './WorkspaceNavIcon.vue'

type OrderStatus = 'all' | 'paid' | 'pending' | 'failed' | 'refunded'
type VisibleOrderStatus = Exclude<OrderStatus, 'all'>

type OrderRow = {
  id: string
  title: string
  description: string
  amount: number
  channel: string
  status: VisibleOrderStatus
  createdAt: string
}

const { locale } = useI18n()
const query = ref('')
const activeStatus = ref<OrderStatus>('all')

const copy = computed(() => locale.value === 'zh-CN' ? {
  title: '我的订单',
  subtitle: '查看充值与服务购买记录，付款状态、金额和时间一目了然。',
  recharge: '充值余额',
  preview: '预览数据',
  totalOrders: '全部订单',
  completedOrders: '已完成',
  paidAmount: '累计支付',
  records: '订单记录',
  recordsHint: '这里会保留你的充值和购买记录',
  all: '全部',
  paid: '已支付',
  pending: '处理中',
  failed: '失败',
  refunded: '已退款',
  search: '搜索订单号或订单内容',
  order: '订单',
  createdAt: '创建时间',
  amount: '金额',
  channel: '支付方式',
  status: '状态',
  noMatch: '没有找到匹配的订单',
  noMatchHint: '尝试更换状态筛选或搜索关键词。',
  empty: '还没有订单记录',
  emptyHint: '完成充值或购买后，相关订单会统一显示在这里。',
  goRecharge: '前往充值',
  gatewayPending: '当前在线支付通道仍在接入中，正式订单产生后将自动同步到这里。',
} : {
  title: 'My orders',
  subtitle: 'Review recharges and purchases with payment status, amount, and time at a glance.',
  recharge: 'Add funds',
  preview: 'Preview data',
  totalOrders: 'All orders',
  completedOrders: 'Completed',
  paidAmount: 'Total paid',
  records: 'Order history',
  recordsHint: 'Recharge and purchase records are kept here',
  all: 'All',
  paid: 'Paid',
  pending: 'Processing',
  failed: 'Failed',
  refunded: 'Refunded',
  search: 'Search order ID or description',
  order: 'Order',
  createdAt: 'Created',
  amount: 'Amount',
  channel: 'Payment',
  status: 'Status',
  noMatch: 'No matching orders',
  noMatchHint: 'Try another status or search term.',
  empty: 'No orders yet',
  emptyHint: 'Recharge and purchase orders will appear here once they are created.',
  goRecharge: 'Add funds',
  gatewayPending: 'Online payment is still being connected. New orders will sync here automatically once available.',
})

const statusMeta = computed<Record<VisibleOrderStatus, { label: string; tone: string }>>(() => ({
  paid: { label: copy.value.paid, tone: 'success' },
  pending: { label: copy.value.pending, tone: 'warning' },
  failed: { label: copy.value.failed, tone: 'danger' },
  refunded: { label: copy.value.refunded, tone: 'muted' },
}))

const previewOrders: OrderRow[] = previewMode ? [
  {
    id: 'SM202609070184',
    title: 'API 余额充值',
    description: '账户余额充值',
    amount: 100,
    channel: '在线支付',
    status: 'paid',
    createdAt: '2026-09-07 10:42',
  },
  {
    id: 'SM202609050126',
    title: 'API 余额充值',
    description: '账户余额充值',
    amount: 50,
    channel: '在线支付',
    status: 'paid',
    createdAt: '2026-09-05 18:16',
  },
  {
    id: 'SM202609030091',
    title: 'API 余额充值',
    description: '账户余额充值',
    amount: 20,
    channel: '在线支付',
    status: 'refunded',
    createdAt: '2026-09-03 09:24',
  },
] : []

// The billing gateway is not connected yet. Keep production data empty rather than
// presenting synthetic transactions as real orders. This array is the integration
// point for the user order API when it becomes available.
const orders = ref<OrderRow[]>(previewOrders)

const filters = computed(() => [
  { value: 'all' as const, label: copy.value.all, count: orders.value.length },
  { value: 'paid' as const, label: copy.value.paid, count: orders.value.filter((item) => item.status === 'paid').length },
  { value: 'pending' as const, label: copy.value.pending, count: orders.value.filter((item) => item.status === 'pending').length },
  { value: 'refunded' as const, label: copy.value.refunded, count: orders.value.filter((item) => item.status === 'refunded').length },
])

const filteredOrders = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  return orders.value.filter((item) => {
    const matchesStatus = activeStatus.value === 'all' || item.status === activeStatus.value
    const matchesQuery = !keyword || [item.id, item.title, item.description, item.channel]
      .some((value) => value.toLowerCase().includes(keyword))
    return matchesStatus && matchesQuery
  })
})

const completedCount = computed(() => orders.value.filter((item) => item.status === 'paid').length)
const paidAmount = computed(() => orders.value
  .filter((item) => item.status === 'paid')
  .reduce((sum, item) => sum + item.amount, 0))

function money(value: number) {
  return `$${Number(value || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
</script>

<template>
  <div class="user-orders-page">
    <header class="orders-hero">
      <div class="orders-hero-copy">
        <div class="orders-title-line">
          <h1>{{ copy.title }}</h1>
          <span v-if="previewMode" class="preview-badge"><i></i>{{ copy.preview }}</span>
        </div>
        <p>{{ copy.subtitle }}</p>
      </div>

      <RouterLink class="recharge-link" to="/subscriptions">
        <WorkspaceNavIcon name="wallet" />
        <span>{{ copy.recharge }}</span>
        <WorkspaceNavIcon name="arrow-up-right" />
      </RouterLink>
    </header>

    <section class="orders-overview" aria-label="Order overview">
      <article>
        <span>{{ copy.totalOrders }}</span>
        <strong>{{ orders.length }}</strong>
      </article>
      <article>
        <span>{{ copy.completedOrders }}</span>
        <strong>{{ completedCount }}</strong>
      </article>
      <article class="amount-overview">
        <span>{{ copy.paidAmount }}</span>
        <strong>{{ money(paidAmount) }}</strong>
      </article>
    </section>

    <section class="orders-surface">
      <header class="orders-toolbar">
        <div class="orders-toolbar-copy">
          <strong>{{ copy.records }}</strong>
          <span>{{ copy.recordsHint }}</span>
        </div>

        <label v-if="orders.length" class="orders-search">
          <WorkspaceNavIcon name="search" />
          <input v-model="query" type="search" :placeholder="copy.search" :aria-label="copy.search" />
          <button v-if="query" type="button" aria-label="Clear search" @click="query = ''">×</button>
        </label>
      </header>

      <nav v-if="orders.length" class="orders-filters" aria-label="Order status filters">
        <button
          v-for="item in filters"
          :key="item.value"
          type="button"
          :class="{ active: activeStatus === item.value }"
          @click="activeStatus = item.value"
        >
          <span>{{ item.label }}</span>
          <b>{{ item.count }}</b>
        </button>
      </nav>

      <template v-if="orders.length">
        <div class="orders-table" role="table">
          <div class="orders-table-head" role="row">
            <span role="columnheader">{{ copy.order }}</span>
            <span role="columnheader">{{ copy.createdAt }}</span>
            <span role="columnheader">{{ copy.amount }}</span>
            <span role="columnheader">{{ copy.channel }}</span>
            <span role="columnheader">{{ copy.status }}</span>
          </div>

          <div v-for="item in filteredOrders" :key="item.id" class="order-row" role="row">
            <div class="order-identity" role="cell">
              <span class="order-icon"><WorkspaceNavIcon name="receipt" /></span>
              <div>
                <strong>{{ item.title }}</strong>
                <span>{{ item.id }}</span>
              </div>
            </div>
            <time role="cell">{{ item.createdAt }}</time>
            <strong class="order-amount" role="cell">{{ money(item.amount) }}</strong>
            <span class="order-channel" role="cell">{{ item.channel }}</span>
            <span role="cell">
              <i class="order-status" :class="`tone-${statusMeta[item.status].tone}`">
                <b></b>{{ statusMeta[item.status].label }}
              </i>
            </span>
          </div>
        </div>

        <div v-if="!filteredOrders.length" class="orders-empty search-empty">
          <span class="empty-icon"><WorkspaceNavIcon name="search" /></span>
          <strong>{{ copy.noMatch }}</strong>
          <p>{{ copy.noMatchHint }}</p>
        </div>
      </template>

      <div v-else class="orders-empty">
        <span class="empty-icon"><WorkspaceNavIcon name="receipt" /></span>
        <strong>{{ copy.empty }}</strong>
        <p>{{ copy.emptyHint }}</p>
        <RouterLink to="/subscriptions">{{ copy.goRecharge }} <span>→</span></RouterLink>
      </div>

      <footer v-if="!previewMode" class="orders-notice">
        <WorkspaceNavIcon name="shield" />
        <span>{{ copy.gatewayPending }}</span>
      </footer>
    </section>
  </div>
</template>

<style scoped>
.user-orders-page {
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 12px 0 44px;
  color: #edf0f3;
}

.orders-hero {
  min-height: 108px;
  margin-bottom: 22px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 28px;
}

.orders-hero-copy { min-width: 0; }
.orders-title-line { display: flex; align-items: center; gap: 12px; }
.orders-title-line h1 {
  margin: 0;
  color: #f7f8fa;
  font-size: clamp(1.9rem, 2.4vw, 2.35rem);
  line-height: 1.08;
  font-weight: 680;
  letter-spacing: -.043em;
}
.orders-hero-copy > p {
  max-width: 620px;
  margin: 10px 0 0;
  color: #858d97;
  font-size: .88rem;
  line-height: 1.6;
}

.preview-badge {
  height: 25px;
  padding: 0 8px;
  border: 1px solid #343840;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #121419;
  color: #8d949e;
  font-size: .66rem;
  font-weight: 650;
}
.preview-badge i { width: 5px; height: 5px; border-radius: 50%; background: #c99a50; }

.recharge-link {
  height: 42px;
  padding: 0 14px;
  border: 1px solid #343a43;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
  background: #171a1f;
  color: #e4e8ec;
  font-size: .78rem;
  font-weight: 650;
  transition: background-color .15s ease, border-color .15s ease, color .15s ease;
}
.recharge-link:hover { border-color: #484f59; background: #1d2127; color: #fff; }
.recharge-link :deep(.workspace-nav-icon) { width: 15px; height: 15px; color: #9099a4; }
.recharge-link :deep(.workspace-nav-icon:last-child) { width: 13px; height: 13px; }

.orders-overview {
  min-height: 92px;
  margin-bottom: 14px;
  border: 1px solid #252a31;
  border-radius: 11px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  background: #0f1114;
  overflow: hidden;
}
.orders-overview article {
  min-width: 0;
  padding: 19px 22px;
  border-right: 1px solid #23272d;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}
.orders-overview article:last-child { border-right: 0; }
.orders-overview span { color: #747d87; font-size: .71rem; font-weight: 620; }
.orders-overview strong {
  color: #eef1f4;
  font-size: 1.26rem;
  line-height: 1;
  font-weight: 680;
  letter-spacing: -.025em;
  font-variant-numeric: tabular-nums;
}
.orders-overview .amount-overview strong { color: #f5f6f7; }

.orders-surface {
  border: 1px solid #252a31;
  border-radius: 12px;
  background: #0d0f12;
  overflow: hidden;
}

.orders-toolbar {
  min-height: 76px;
  padding: 15px 17px 15px 20px;
  border-bottom: 1px solid #22262c;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}
.orders-toolbar-copy { min-width: 0; display: flex; align-items: baseline; gap: 10px; }
.orders-toolbar-copy strong { color: #e8ebee; font-size: .9rem; font-weight: 650; }
.orders-toolbar-copy span { color: #656e78; font-size: .69rem; }

.orders-search {
  width: min(310px, 36vw);
  height: 38px;
  padding: 0 9px 0 11px;
  border: 1px solid #2b3037;
  border-radius: 8px;
  display: grid;
  grid-template-columns: 15px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  background: #090b0e;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.orders-search:focus-within { border-color: #3b5369; box-shadow: 0 0 0 3px rgba(74,132,183,.07); }
.orders-search :deep(.workspace-nav-icon) { width: 14px; height: 14px; color: #646e78; }
.orders-search input { min-width: 0; border: 0; outline: 0; background: transparent; color: #dfe3e7; font: inherit; font-size: .74rem; }
.orders-search input::placeholder { color: #555d66; }
.orders-search button {
  width: 22px;
  height: 22px;
  padding: 0;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: #69727c;
  cursor: pointer;
  font: inherit;
  font-size: .95rem;
}
.orders-search button:hover { background: #171a1f; color: #c5cbd1; }

.orders-filters {
  min-height: 49px;
  padding: 8px 13px;
  border-bottom: 1px solid #22262c;
  display: flex;
  align-items: center;
  gap: 4px;
  background: #0f1114;
}
.orders-filters button {
  height: 32px;
  padding: 0 9px;
  border: 1px solid transparent;
  border-radius: 7px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: transparent;
  color: #777f89;
  cursor: pointer;
  font: inherit;
  font-size: .71rem;
  font-weight: 620;
}
.orders-filters button:hover { background: #16191d; color: #c1c6cc; }
.orders-filters button.active { border-color: #30363e; background: #191c21; color: #eef1f4; }
.orders-filters button b {
  min-width: 19px;
  height: 19px;
  padding: 0 5px;
  border-radius: 5px;
  display: grid;
  place-items: center;
  background: #101216;
  color: #69717b;
  font-size: .61rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}
.orders-filters button.active b { background: #252a31; color: #aeb5bd; }

.orders-table { width: 100%; }
.orders-table-head,
.order-row {
  display: grid;
  grid-template-columns: minmax(280px, 1.55fr) minmax(145px, .82fr) minmax(100px, .58fr) minmax(120px, .65fr) minmax(105px, .58fr);
  gap: 18px;
  align-items: center;
}
.orders-table-head {
  min-height: 43px;
  padding: 0 20px;
  border-bottom: 1px solid #22262c;
  background: #0a0c0f;
  color: #606974;
  font-size: .64rem;
  font-weight: 650;
}
.order-row {
  min-height: 82px;
  padding: 14px 20px;
  border-bottom: 1px solid #20242a;
  color: #a6aeb7;
  font-size: .75rem;
  transition: background-color .14s ease;
}
.order-row:last-child { border-bottom: 0; }
.order-row:hover { background: #111419; }

.order-identity { min-width: 0; display: flex; align-items: center; gap: 12px; }
.order-icon {
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  border: 1px solid #2b3139;
  border-radius: 8px;
  display: grid;
  place-items: center;
  background: #13161a;
  color: #8b949e;
}
.order-icon :deep(.workspace-nav-icon) { width: 15px; height: 15px; }
.order-identity > div { min-width: 0; }
.order-identity strong { display: block; color: #dce0e4; font-size: .78rem; font-weight: 630; }
.order-identity div > span {
  display: block;
  margin-top: 5px;
  overflow: hidden;
  color: #626b75;
  font: 500 .64rem/1.25 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.order-row time { color: #7a838d; font-variant-numeric: tabular-nums; }
.order-amount { color: #e9ecef; font-size: .79rem; font-weight: 650; font-variant-numeric: tabular-nums; }
.order-channel { color: #8c949e; }

.order-status {
  width: fit-content;
  min-height: 25px;
  padding: 0 8px;
  border: 1px solid transparent;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-style: normal;
  font-size: .67rem;
  font-weight: 650;
  white-space: nowrap;
}
.order-status b { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.order-status.tone-success { border-color: rgba(77,176,133,.18); background: rgba(48,143,102,.08); color: #67c69f; }
.order-status.tone-warning { border-color: rgba(195,148,72,.18); background: rgba(177,126,46,.08); color: #d0a35d; }
.order-status.tone-danger { border-color: rgba(195,91,91,.18); background: rgba(170,68,68,.08); color: #d27c7c; }
.order-status.tone-muted { border-color: #2c3036; background: #14161a; color: #777f88; }

.orders-empty {
  min-height: 330px;
  padding: 52px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}
.empty-icon {
  width: 46px;
  height: 46px;
  margin-bottom: 17px;
  border: 1px solid #2a3037;
  border-radius: 11px;
  display: grid;
  place-items: center;
  background: #13161a;
  color: #747e88;
}
.empty-icon :deep(.workspace-nav-icon) { width: 20px; height: 20px; }
.orders-empty > strong { color: #d9dde1; font-size: .93rem; font-weight: 650; }
.orders-empty > p { max-width: 390px; margin: 9px 0 0; color: #6f7882; font-size: .75rem; line-height: 1.6; }
.orders-empty > a {
  height: 36px;
  margin-top: 21px;
  padding: 0 12px;
  border: 1px solid #30363e;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #171a1f;
  color: #cdd2d7;
  font-size: .72rem;
  font-weight: 630;
}
.orders-empty > a:hover { border-color: #424952; background: #1b1f25; color: #fff; }
.search-empty { min-height: 220px; border-top: 1px solid #20242a; }

.orders-notice {
  min-height: 49px;
  padding: 10px 17px;
  border-top: 1px solid #22262c;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #0b0d10;
  color: #5f6872;
  font-size: .67rem;
  line-height: 1.5;
  text-align: center;
}
.orders-notice :deep(.workspace-nav-icon) { width: 13px; height: 13px; flex: 0 0 auto; color: #67717b; }

@media (max-width: 900px) {
  .orders-table-head,
  .order-row {
    grid-template-columns: minmax(230px, 1.45fr) minmax(125px, .8fr) minmax(90px, .58fr) minmax(95px, .6fr);
  }
  .orders-table-head > :nth-child(4),
  .order-row > :nth-child(4) { display: none; }
}

@media (max-width: 720px) {
  .user-orders-page { padding-top: 0; padding-bottom: 30px; }
  .orders-hero { min-height: 0; margin: 12px 0 18px; align-items: flex-start; }
  .orders-title-line { align-items: flex-start; flex-direction: column; gap: 8px; }
  .orders-title-line h1 { font-size: 1.8rem; }
  .orders-hero-copy > p { margin-top: 8px; font-size: .79rem; }
  .recharge-link { height: 38px; padding: 0 11px; }
  .recharge-link :deep(.workspace-nav-icon:first-child) { display: none; }

  .orders-overview { min-height: 80px; }
  .orders-overview article { padding: 15px 14px; }
  .orders-overview span { font-size: .65rem; }
  .orders-overview strong { font-size: 1.05rem; }

  .orders-toolbar { padding: 14px; align-items: stretch; flex-direction: column; gap: 12px; }
  .orders-toolbar-copy { display: block; }
  .orders-toolbar-copy span { display: block; margin-top: 5px; }
  .orders-search { width: 100%; }
  .orders-filters { overflow-x: auto; scrollbar-width: none; }
  .orders-filters::-webkit-scrollbar { display: none; }

  .orders-table-head { display: none; }
  .order-row {
    min-height: 0;
    padding: 16px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 10px 16px;
  }
  .order-identity { grid-column: 1 / -1; padding-bottom: 4px; }
  .order-row time { grid-column: 1; grid-row: 2; font-size: .69rem; }
  .order-amount { grid-column: 2; grid-row: 2; text-align: right; }
  .order-channel { display: none; }
  .order-row > span:last-child { grid-column: 1 / -1; grid-row: 3; }
  .order-status { margin-top: 2px; }
  .orders-empty { min-height: 280px; }
}

@media (max-width: 470px) {
  .orders-hero { flex-direction: column; gap: 15px; }
  .recharge-link { width: 100%; justify-content: center; }
  .orders-overview { grid-template-columns: 1fr 1fr; }
  .orders-overview article { border-bottom: 1px solid #23272d; }
  .orders-overview article:nth-child(2) { border-right: 0; }
  .orders-overview .amount-overview { grid-column: 1 / -1; border-bottom: 0; border-right: 0; }
}
</style>
