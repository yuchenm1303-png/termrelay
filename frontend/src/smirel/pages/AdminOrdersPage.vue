<script setup lang="ts">
import { computed, ref } from 'vue'
import WorkspaceNavIcon from '../components/WorkspaceNavIcon.vue'

type OrderStatus = 'all' | 'paid' | 'pending' | 'failed' | 'refunded'

type OrderRow = {
  id: string
  user: string
  product: string
  amount: number
  channel: string
  status: Exclude<OrderStatus, 'all'>
  createdAt: string
}

const query = ref('')
const activeStatus = ref<OrderStatus>('all')

const orders: OrderRow[] = [
  { id: 'SM202609070184', user: 'cheng@example.com', product: 'API 余额充值', amount: 299, channel: '支付宝', status: 'paid', createdAt: '09-07 10:42' },
  { id: 'SM202609070183', user: 'lin@example.com', product: 'API 余额充值', amount: 99, channel: '微信支付', status: 'paid', createdAt: '09-07 10:38' },
  { id: 'SM202609070182', user: 'alex@example.com', product: 'API 余额充值', amount: 499, channel: 'Stripe', status: 'pending', createdAt: '09-07 10:31' },
  { id: 'SM202609070181', user: 'xu@example.com', product: 'API 余额充值', amount: 59, channel: '支付宝', status: 'failed', createdAt: '09-07 10:27' },
  { id: 'SM202609070180', user: 'mika@example.com', product: 'API 余额充值', amount: 199, channel: 'Stripe', status: 'paid', createdAt: '09-07 10:18' },
  { id: 'SM202609070179', user: 'ren@example.com', product: 'API 余额充值', amount: 129, channel: '微信支付', status: 'refunded', createdAt: '09-07 09:56' },
  { id: 'SM202609070178', user: 'yang@example.com', product: 'API 余额充值', amount: 399, channel: '支付宝', status: 'paid', createdAt: '09-07 09:41' },
]

const statusMeta: Record<Exclude<OrderStatus, 'all'>, { label: string; tone: string }> = {
  paid: { label: '已支付', tone: 'success' },
  pending: { label: '处理中', tone: 'warning' },
  failed: { label: '支付失败', tone: 'danger' },
  refunded: { label: '已退款', tone: 'muted' },
}

const filters = computed(() => [
  { value: 'all' as const, label: '全部', count: orders.length },
  { value: 'paid' as const, label: '已支付', count: orders.filter((item) => item.status === 'paid').length },
  { value: 'pending' as const, label: '处理中', count: orders.filter((item) => item.status === 'pending').length },
  { value: 'failed' as const, label: '失败', count: orders.filter((item) => item.status === 'failed').length },
  { value: 'refunded' as const, label: '退款', count: orders.filter((item) => item.status === 'refunded').length },
])

const filteredOrders = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  return orders.filter((item) => {
    const matchesStatus = activeStatus.value === 'all' || item.status === activeStatus.value
    const matchesQuery = !keyword || [item.id, item.user, item.product, item.channel]
      .some((value) => value.toLowerCase().includes(keyword))
    return matchesStatus && matchesQuery
  })
})

const paidOrders = computed(() => orders.filter((item) => item.status === 'paid'))
const paidAmount = computed(() => paidOrders.value.reduce((sum, item) => sum + item.amount, 0))
const attentionCount = computed(() => orders.filter((item) => item.status === 'pending' || item.status === 'failed').length)

function money(value: number) {
  return `¥${value.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
</script>

<template>
  <section class="workspace-page admin-orders-page">
    <header class="orders-heading">
      <div>
        <div class="orders-eyebrow"><i></i><span>ORDER OPERATIONS</span></div>
        <h1>订单管理</h1>
        <p>集中查看平台订单、支付状态与异常交易。</p>
      </div>
      <div class="orders-heading-actions">
        <span class="orders-demo-badge"><i></i>示例数据</span>
        <RouterLink class="orders-outline-link" to="/admin/orders/dashboard">
          支付概览 <WorkspaceNavIcon name="chart" />
        </RouterLink>
      </div>
    </header>

    <section class="orders-summary" aria-label="订单概览">
      <article>
        <span>当前记录</span>
        <strong>{{ orders.length }}</strong>
        <small>ORDERS</small>
      </article>
      <article>
        <span>已支付</span>
        <strong>{{ paidOrders.length }}</strong>
        <small>{{ Math.round((paidOrders.length / orders.length) * 100) }}% SUCCESS</small>
      </article>
      <article>
        <span>实收金额</span>
        <strong>{{ money(paidAmount) }}</strong>
        <small>PAID AMOUNT</small>
      </article>
      <article :class="{ attention: attentionCount > 0 }">
        <span>需要关注</span>
        <strong>{{ attentionCount }}</strong>
        <small>PENDING / FAILED</small>
      </article>
    </section>

    <section class="orders-panel">
      <header class="orders-panel-head">
        <div class="orders-panel-title">
          <span>ORDERS</span>
          <strong>全部订单</strong>
          <small>{{ filteredOrders.length }} 条结果</small>
        </div>

        <label class="orders-search">
          <WorkspaceNavIcon name="search" />
          <input v-model="query" type="search" placeholder="搜索订单号、用户或渠道" aria-label="搜索订单" />
          <button v-if="query" type="button" aria-label="清空搜索" @click="query = ''">×</button>
        </label>
      </header>

      <nav class="orders-filters" aria-label="订单状态筛选">
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

      <div class="orders-table-wrap">
        <div class="orders-table">
          <div class="orders-row orders-table-head">
            <span>订单号</span>
            <span>用户</span>
            <span>订单内容</span>
            <span>金额</span>
            <span>支付渠道</span>
            <span>状态</span>
            <span>创建时间</span>
          </div>

          <div v-for="item in filteredOrders" :key="item.id" class="orders-row">
            <code>{{ item.id }}</code>
            <span class="order-user">{{ item.user }}</span>
            <span>{{ item.product }}</span>
            <strong class="order-amount">{{ money(item.amount) }}</strong>
            <span class="order-channel">{{ item.channel }}</span>
            <span>
              <i class="order-status" :class="`tone-${statusMeta[item.status].tone}`">
                <b></b>{{ statusMeta[item.status].label }}
              </i>
            </span>
            <time>{{ item.createdAt }}</time>
          </div>

          <div v-if="!filteredOrders.length" class="orders-empty">
            <span><WorkspaceNavIcon name="receipt" /></span>
            <strong>没有匹配的订单</strong>
            <small>调整搜索条件或订单状态后重试。</small>
          </div>
        </div>
      </div>

      <footer class="orders-panel-foot">
        <span>当前展示商业版订单管理 UI 示例。</span>
        <span>真实订单接口接入后，此处直接替换为实时数据，不改变页面结构。</span>
      </footer>
    </section>
  </section>
</template>

<style scoped>
.admin-orders-page {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding-bottom: 36px;
  color: #e9ecf0;
}

.orders-heading {
  min-height: 112px;
  margin-bottom: 26px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
}

.orders-eyebrow {
  margin-bottom: 12px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #69717b;
  font: 700 .65rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .12em;
}

.orders-eyebrow i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4fc99b;
}

.orders-heading h1 {
  margin: 0;
  color: #f7f8fa;
  font-size: 2.15rem;
  line-height: 1.05;
  font-weight: 680;
  letter-spacing: -.042em;
}

.orders-heading p {
  margin: 9px 0 0;
  color: #7d8590;
  font-size: .88rem;
  line-height: 1.55;
}

.orders-heading-actions {
  display: flex;
  align-items: center;
  gap: 9px;
}

.orders-demo-badge,
.orders-outline-link {
  height: 38px;
  padding: 0 13px;
  border: 1px solid #272b32;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: #0e1014;
  color: #9199a3;
  font-size: .73rem;
}

.orders-demo-badge i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #d3a04f;
}

.orders-outline-link {
  color: #c5cad0;
  transition: background-color .15s ease, border-color .15s ease, color .15s ease;
}

.orders-outline-link :deep(.workspace-nav-icon) { width: 14px; height: 14px; color: #707984; }
.orders-outline-link:hover { border-color: #3a4048; background: #14171b; color: #fff; }

.orders-summary {
  margin-bottom: 14px;
  border: 1px solid #22262d;
  border-radius: 12px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  overflow: hidden;
  background: #0f1115;
}

.orders-summary article {
  min-height: 126px;
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  border-right: 1px solid #22262d;
}

.orders-summary article:last-child { border-right: 0; }
.orders-summary article > span { color: #858d97; font-size: .76rem; font-weight: 600; }
.orders-summary article > strong { margin-top: 10px; color: #f1f3f5; font-size: 1.62rem; line-height: 1; font-weight: 670; letter-spacing: -.035em; }
.orders-summary article > small { margin-top: 10px; color: #5f6873; font: 650 .59rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .09em; }
.orders-summary article.attention > strong { color: #e1b15f; }

.orders-panel {
  border: 1px solid #22262d;
  border-radius: 12px;
  background: #0d0f13;
  overflow: hidden;
}

.orders-panel-head {
  min-height: 82px;
  padding: 16px 18px 16px 21px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border-bottom: 1px solid #20242a;
}

.orders-panel-title { display: flex; align-items: baseline; gap: 9px; }
.orders-panel-title > span { color: #626b75; font: 700 .6rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
.orders-panel-title > strong { color: #e9ecf0; font-size: .96rem; font-weight: 650; }
.orders-panel-title > small { color: #646d77; font-size: .68rem; }

.orders-search {
  width: min(330px, 36vw);
  height: 40px;
  padding: 0 10px 0 12px;
  border: 1px solid #2a2f37;
  border-radius: 8px;
  display: grid;
  grid-template-columns: 16px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  background: #0a0c0f;
  transition: border-color .15s ease, box-shadow .15s ease;
}

.orders-search:focus-within { border-color: #385d79; box-shadow: 0 0 0 3px rgba(47,150,232,.07); }
.orders-search :deep(.workspace-nav-icon) { width: 15px; height: 15px; color: #68717b; }
.orders-search input { width: 100%; border: 0; outline: 0; background: transparent; color: #dfe3e7; font: inherit; font-size: .78rem; }
.orders-search input::placeholder { color: #59616b; }
.orders-search button { width: 22px; height: 22px; border: 0; border-radius: 5px; background: transparent; color: #66707a; cursor: pointer; font-size: 1rem; }
.orders-search button:hover { background: #171a20; color: #bdc3ca; }

.orders-filters {
  min-height: 54px;
  padding: 9px 14px;
  display: flex;
  align-items: center;
  gap: 5px;
  border-bottom: 1px solid #20242a;
  background: #0f1115;
}

.orders-filters button {
  min-height: 34px;
  padding: 0 10px;
  border: 1px solid transparent;
  border-radius: 7px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: transparent;
  color: #7d8590;
  cursor: pointer;
  font: inherit;
  font-size: .73rem;
  font-weight: 600;
}

.orders-filters button:hover { background: #15181d; color: #c8cdd3; }
.orders-filters button.active { border-color: #2c3946; background: #111a23; color: #dce9f4; }
.orders-filters button b { min-width: 20px; height: 20px; padding: 0 5px; border-radius: 5px; display: grid; place-items: center; background: #171a1f; color: #68717b; font-size: .63rem; }
.orders-filters button.active b { background: #1a2a38; color: #80bde7; }

.orders-table-wrap { width: 100%; overflow-x: auto; }
.orders-table { min-width: 1040px; }
.orders-row {
  min-height: 58px;
  padding: 0 20px;
  display: grid;
  grid-template-columns: 1.3fr 1.35fr 1.15fr .72fr .8fr .78fr .82fr;
  gap: 16px;
  align-items: center;
  border-bottom: 1px solid #1d2026;
  color: #9ca3ac;
  font-size: .74rem;
}

.orders-row:not(.orders-table-head):hover { background: #111419; }
.orders-row:last-of-type { border-bottom: 0; }
.orders-table-head { min-height: 43px; background: #0b0d11; color: #626a74; font-size: .64rem; font-weight: 700; letter-spacing: .04em; }
.orders-row code { color: #b9c1ca; font: 600 .7rem/1.3 ui-monospace, SFMono-Regular, Menlo, monospace; }
.order-user { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #aeb5bd; }
.order-amount { color: #e5e8eb; font-size: .77rem; font-weight: 650; }
.order-channel { color: #a7aeb7; }
.orders-row time { color: #717a84; font: .68rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; }

.order-status {
  min-height: 25px;
  width: max-content;
  padding: 0 8px;
  border: 1px solid #2b3037;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #12151a;
  color: #929aa4;
  font-size: .65rem;
  font-style: normal;
  font-weight: 620;
}

.order-status > b { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.order-status.tone-success { border-color: rgba(76,195,148,.22); background: rgba(76,195,148,.06); color: #68cda7; }
.order-status.tone-warning { border-color: rgba(211,160,79,.24); background: rgba(211,160,79,.06); color: #d4a75e; }
.order-status.tone-danger { border-color: rgba(210,100,100,.22); background: rgba(210,100,100,.06); color: #d87979; }
.order-status.tone-muted { color: #78818b; }

.orders-empty { min-height: 240px; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #6c7580; }
.orders-empty > span { width: 40px; height: 40px; margin-bottom: 12px; border: 1px solid #282d34; border-radius: 10px; display: grid; place-items: center; background: #111419; }
.orders-empty :deep(.workspace-nav-icon) { width: 18px; height: 18px; }
.orders-empty strong { color: #b9c0c8; font-size: .8rem; }
.orders-empty small { margin-top: 6px; font-size: .69rem; }

.orders-panel-foot {
  min-height: 48px;
  padding: 0 20px;
  border-top: 1px solid #20242a;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  color: #5f6872;
  background: #0b0d11;
  font-size: .65rem;
}

@media (max-width: 900px) {
  .orders-heading { min-height: auto; align-items: flex-start; flex-direction: column; }
  .orders-heading-actions { width: 100%; }
  .orders-summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .orders-summary article:nth-child(2) { border-right: 0; }
  .orders-summary article:nth-child(-n+2) { border-bottom: 1px solid #22262d; }
  .orders-panel-head { align-items: flex-start; flex-direction: column; }
  .orders-search { width: 100%; }
  .orders-panel-foot { padding: 12px 16px; align-items: flex-start; flex-direction: column; gap: 4px; }
}

@media (max-width: 560px) {
  .orders-heading h1 { font-size: 1.8rem; }
  .orders-summary { grid-template-columns: 1fr; }
  .orders-summary article { min-height: 105px; border-right: 0; border-bottom: 1px solid #22262d; }
  .orders-summary article:last-child { border-bottom: 0; }
  .orders-heading-actions { flex-wrap: wrap; }
  .orders-filters { overflow-x: auto; }
}
</style>
