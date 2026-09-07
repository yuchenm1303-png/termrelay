<script setup lang="ts">
import WorkspaceNavIcon from '../components/WorkspaceNavIcon.vue'

const summary = [
  { label: '今日实收', value: '¥12,486.20', note: '较昨日 +8.4%', icon: 'wallet', tone: 'success' },
  { label: '成功支付', value: '186', note: '共 190 笔支付请求', icon: 'receipt', tone: 'neutral' },
  { label: '支付成功率', value: '97.8%', note: '近 24 小时', icon: 'activity', tone: 'success' },
  { label: '待结算金额', value: '¥3,240.80', note: '预计今日 18:00 结算', icon: 'credit-card', tone: 'warning' },
]

const trend = [38, 52, 47, 66, 61, 79, 92]
const trendDays = ['09/01', '09/02', '09/03', '09/04', '09/05', '09/06', '今日']

const statusRows = [
  { label: '支付成功', value: 186, percent: 97.8, tone: 'success' },
  { label: '处理中', value: 2, percent: 1.1, tone: 'warning' },
  { label: '支付失败', value: 2, percent: 1.1, tone: 'danger' },
]

const channelRows = [
  { label: '支付宝', amount: '¥5,742.60', percent: 46 },
  { label: '微信支付', amount: '¥4,245.30', percent: 34 },
  { label: 'Stripe', amount: '¥2,498.30', percent: 20 },
]

const pendingItems = [
  { label: '支付失败', value: '2 笔', detail: '需要确认支付渠道返回状态', tone: 'danger' },
  { label: '退款处理中', value: '1 笔', detail: '预计 1 个工作日内完成', tone: 'warning' },
  { label: '人工复核', value: '0 笔', detail: '当前无高风险支付订单', tone: 'success' },
]

const transactions = [
  { id: 'SM202609070184', user: 'cheng@example.com', channel: '支付宝', amount: '¥299.00', status: '支付成功', tone: 'success', time: '10:42:18' },
  { id: 'SM202609070183', user: 'lin@example.com', channel: '微信支付', amount: '¥99.00', status: '支付成功', tone: 'success', time: '10:38:51' },
  { id: 'SM202609070182', user: 'alex@example.com', channel: 'Stripe', amount: '¥499.00', status: '处理中', tone: 'warning', time: '10:31:06' },
  { id: 'SM202609070181', user: 'xu@example.com', channel: '支付宝', amount: '¥59.00', status: '支付失败', tone: 'danger', time: '10:27:43' },
  { id: 'SM202609070180', user: 'mika@example.com', channel: 'Stripe', amount: '¥199.00', status: '支付成功', tone: 'success', time: '10:18:29' },
]
</script>

<template>
  <section class="workspace-page payment-overview">
    <header class="payment-heading">
      <div>
        <div class="payment-eyebrow"><i></i><span>PAYMENT OPERATIONS</span></div>
        <h1>支付概览</h1>
        <p>查看平台收款、支付状态、渠道分布与异常交易。</p>
      </div>
      <div class="payment-heading-actions">
        <span class="demo-badge"><i></i> 示例数据</span>
        <RouterLink class="outline-link" to="/admin/orders">订单管理 <b>→</b></RouterLink>
      </div>
    </header>

    <div class="payment-summary-grid">
      <article v-for="item in summary" :key="item.label" class="payment-summary-card" :class="`tone-${item.tone}`">
        <header>
          <span>{{ item.label }}</span>
          <span class="summary-icon"><WorkspaceNavIcon :name="item.icon" /></span>
        </header>
        <strong>{{ item.value }}</strong>
        <footer>
          <span>{{ item.note }}</span>
          <i></i>
        </footer>
      </article>
    </div>

    <div class="payment-main-grid">
      <section class="payment-panel trend-panel">
        <header class="panel-heading">
          <div>
            <span class="panel-kicker">REVENUE</span>
            <h2>近 7 日实收趋势</h2>
          </div>
          <div class="trend-total">
            <span>7 日实收</span>
            <strong>¥68,420.70</strong>
          </div>
        </header>

        <div class="trend-chart" aria-label="近 7 日实收趋势示例图">
          <div class="trend-grid-lines" aria-hidden="true"><i></i><i></i><i></i><i></i></div>
          <div class="trend-bars">
            <div v-for="(height, index) in trend" :key="trendDays[index]" class="trend-column">
              <span class="trend-value" :style="{ height: `${height}%` }"><i></i></span>
              <small>{{ trendDays[index] }}</small>
            </div>
          </div>
        </div>
      </section>

      <section class="payment-panel status-panel">
        <header class="panel-heading compact">
          <div>
            <span class="panel-kicker">STATUS</span>
            <h2>今日支付状态</h2>
          </div>
          <strong class="status-total">190 <small>笔</small></strong>
        </header>

        <div class="payment-status-list">
          <div v-for="item in statusRows" :key="item.label" class="payment-status-row">
            <div class="status-row-head">
              <span><i :class="`tone-${item.tone}`"></i>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </div>
            <div class="status-track"><i :class="`tone-${item.tone}`" :style="{ width: `${item.percent}%` }"></i></div>
            <small>{{ item.percent }}%</small>
          </div>
        </div>
      </section>
    </div>

    <div class="payment-secondary-grid">
      <section class="payment-panel channel-panel">
        <header class="panel-heading compact">
          <div>
            <span class="panel-kicker">CHANNELS</span>
            <h2>支付渠道</h2>
          </div>
          <span class="panel-caption">按今日实收金额</span>
        </header>

        <div class="channel-list">
          <div v-for="item in channelRows" :key="item.label" class="channel-row">
            <div class="channel-copy"><strong>{{ item.label }}</strong><span>{{ item.amount }}</span></div>
            <div class="channel-progress"><i :style="{ width: `${item.percent}%` }"></i></div>
            <b>{{ item.percent }}%</b>
          </div>
        </div>
      </section>

      <section class="payment-panel attention-panel">
        <header class="panel-heading compact">
          <div>
            <span class="panel-kicker">ATTENTION</span>
            <h2>待处理事项</h2>
          </div>
          <span class="attention-count">3 项</span>
        </header>

        <div class="attention-list">
          <div v-for="item in pendingItems" :key="item.label" class="attention-row">
            <i class="attention-dot" :class="`tone-${item.tone}`"></i>
            <div><strong>{{ item.label }}</strong><span>{{ item.detail }}</span></div>
            <b>{{ item.value }}</b>
          </div>
        </div>
      </section>
    </div>

    <section class="payment-panel transaction-panel">
      <header class="transaction-heading">
        <div>
          <span class="panel-kicker">TRANSACTIONS</span>
          <h2>最近交易</h2>
        </div>
        <RouterLink to="/admin/orders">查看全部订单 <b>→</b></RouterLink>
      </header>

      <div class="transaction-table">
        <div class="transaction-row transaction-head">
          <span>订单号</span><span>用户</span><span>支付渠道</span><span>金额</span><span>状态</span><span>时间</span>
        </div>
        <div v-for="item in transactions" :key="item.id" class="transaction-row">
          <code>{{ item.id }}</code>
          <span>{{ item.user }}</span>
          <span>{{ item.channel }}</span>
          <strong>{{ item.amount }}</strong>
          <span><i class="status-pill" :class="`tone-${item.tone}`">{{ item.status }}</i></span>
          <time>{{ item.time }}</time>
        </div>
      </div>
    </section>

    <footer class="payment-demo-note">
      <span>当前页面为管理员支付模块 UI 实例，金额与交易均为示例数据。</span>
      <span>接口接入后将替换为实时支付与结算数据。</span>
    </footer>
  </section>
</template>

<style scoped>
.payment-overview {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding-bottom: 36px;
  color: #e9ecf0;
}

.payment-heading {
  min-height: 112px;
  margin-bottom: 28px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
}

.payment-eyebrow {
  margin-bottom: 12px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #6d7580;
  font: 700 .66rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .12em;
}
.payment-eyebrow i { width: 6px; height: 6px; border-radius: 50%; background: #4bcc99; box-shadow: 0 0 0 4px rgba(75,204,153,.07); }
.payment-heading h1 { margin: 0; color: #f7f8fa; font-size: 2.25rem; line-height: 1.04; letter-spacing: -.045em; }
.payment-heading p { margin: 10px 0 0; color: #7d8590; font-size: .9rem; line-height: 1.6; }

.payment-heading-actions { display: flex; align-items: center; gap: 9px; }
.demo-badge,
.outline-link {
  height: 38px;
  padding: 0 13px;
  border: 1px solid #252a31;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: #0e1014;
  color: #9aa2ac;
  font-size: .73rem;
}
.demo-badge i { width: 6px; height: 6px; border-radius: 50%; background: #d7a14c; }
.outline-link { color: #c2c7cd; transition: border-color .15s ease, background-color .15s ease, color .15s ease; }
.outline-link:hover { border-color: #373d45; background: #14171b; color: #fff; }
.outline-link b { color: #737b85; }

.payment-summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 13px; }
.payment-summary-card {
  --tone: #aab2bc;
  min-height: 154px;
  padding: 20px 21px 17px;
  border: 1px solid #22262d;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  background: #0f1115;
  box-shadow: inset 0 1px rgba(255,255,255,.015);
}
.payment-summary-card.tone-success { --tone: #4fc99b; }
.payment-summary-card.tone-warning { --tone: #d8a454; }
.payment-summary-card.tone-danger { --tone: #d96a6a; }
.payment-summary-card header { display: flex; align-items: center; justify-content: space-between; gap: 12px; color: #8d959f; font-size: .79rem; font-weight: 600; }
.summary-icon { width: 32px; height: 32px; border: 1px solid #2a3038; border-radius: 8px; display: grid; place-items: center; color: var(--tone); background: #13161b; }
.summary-icon :deep(.workspace-nav-icon) { width: 16px; height: 16px; }
.payment-summary-card > strong { margin-top: 18px; color: #f4f5f7; font-size: 1.78rem; line-height: 1; letter-spacing: -.035em; }
.payment-summary-card footer { margin-top: auto; padding-top: 16px; display: flex; align-items: center; justify-content: space-between; gap: 10px; color: #717a84; font-size: .7rem; }
.payment-summary-card footer i { width: 30px; height: 2px; border-radius: 99px; background: var(--tone); opacity: .65; }

.payment-main-grid { margin-top: 14px; display: grid; grid-template-columns: minmax(0, 1.7fr) minmax(300px, .8fr); gap: 14px; }
.payment-secondary-grid { margin-top: 14px; display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.payment-panel { border: 1px solid #22262d; border-radius: 12px; background: #0d0f13; box-shadow: inset 0 1px rgba(255,255,255,.012); overflow: hidden; }
.panel-heading { min-height: 82px; padding: 18px 21px; border-bottom: 1px solid #20242a; display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.panel-heading.compact { min-height: 74px; }
.panel-kicker { display: block; margin-bottom: 6px; color: #626b75; font: 700 .6rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
.panel-heading h2,
.transaction-heading h2 { margin: 0; color: #e9ecf0; font-size: .98rem; font-weight: 650; letter-spacing: -.012em; }
.trend-total { text-align: right; }
.trend-total span { display: block; margin-bottom: 4px; color: #69727c; font-size: .65rem; }
.trend-total strong { color: #e4e7eb; font-size: .96rem; }
.panel-caption { color: #68717b; font-size: .66rem; }
.status-total { color: #e8eaed; font-size: 1.2rem; }
.status-total small { color: #69717b; font-size: .65rem; font-weight: 500; }

.trend-chart { position: relative; height: 236px; padding: 24px 24px 15px; }
.trend-grid-lines { position: absolute; inset: 24px 24px 43px; display: flex; flex-direction: column; justify-content: space-between; pointer-events: none; }
.trend-grid-lines i { width: 100%; height: 1px; background: #1b1e23; }
.trend-bars { position: relative; height: 100%; display: grid; grid-template-columns: repeat(7, 1fr); gap: 16px; align-items: end; }
.trend-column { height: 100%; display: flex; flex-direction: column; justify-content: flex-end; align-items: center; gap: 10px; }
.trend-value { position: relative; width: min(34px, 72%); min-height: 24px; border-radius: 5px 5px 2px 2px; background: #242a31; transition: background-color .15s ease; }
.trend-column:last-child .trend-value { background: #3f8a70; }
.trend-value i { position: absolute; inset: 0 0 auto; height: 1px; background: rgba(255,255,255,.18); }
.trend-column small { color: #656e78; font-size: .62rem; }

.payment-status-list { padding: 9px 20px 15px; }
.payment-status-row { padding: 15px 0; border-bottom: 1px solid #1d2025; }
.payment-status-row:last-child { border-bottom: 0; }
.status-row-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.status-row-head span { display: inline-flex; align-items: center; gap: 8px; color: #a1a8b1; font-size: .74rem; }
.status-row-head span i { width: 6px; height: 6px; border-radius: 50%; }
.status-row-head strong { color: #e4e7ea; font-size: .79rem; }
.status-track { margin-top: 10px; height: 4px; border-radius: 99px; overflow: hidden; background: #1d2127; }
.status-track i { display: block; height: 100%; border-radius: inherit; }
.payment-status-row > small { margin-top: 6px; display: block; color: #616a74; font-size: .6rem; text-align: right; }
.status-row-head i.tone-success,
.status-track i.tone-success,
.attention-dot.tone-success,
.status-pill.tone-success { background-color: #4bc494; }
.status-row-head i.tone-warning,
.status-track i.tone-warning,
.attention-dot.tone-warning,
.status-pill.tone-warning { background-color: #d6a04d; }
.status-row-head i.tone-danger,
.status-track i.tone-danger,
.attention-dot.tone-danger,
.status-pill.tone-danger { background-color: #d56565; }

.channel-list { padding: 7px 21px 14px; }
.channel-row { min-height: 62px; display: grid; grid-template-columns: 118px 1fr 42px; align-items: center; gap: 15px; border-bottom: 1px solid #1d2025; }
.channel-row:last-child { border-bottom: 0; }
.channel-copy { display: flex; flex-direction: column; gap: 4px; }
.channel-copy strong { color: #d9dde1; font-size: .75rem; font-weight: 600; }
.channel-copy span { color: #69727c; font-size: .65rem; }
.channel-progress { height: 5px; border-radius: 99px; overflow: hidden; background: #1d2126; }
.channel-progress i { display: block; height: 100%; border-radius: inherit; background: #555d67; }
.channel-row b { color: #8c949d; font-size: .67rem; font-weight: 600; text-align: right; }

.attention-count { min-width: 38px; height: 25px; padding: 0 8px; border: 1px solid #342d21; border-radius: 99px; display: inline-grid; place-items: center; color: #cba463; background: #17140f; font-size: .63rem; }
.attention-list { padding: 6px 21px 13px; }
.attention-row { min-height: 63px; display: grid; grid-template-columns: 8px 1fr auto; align-items: center; gap: 12px; border-bottom: 1px solid #1d2025; }
.attention-row:last-child { border-bottom: 0; }
.attention-dot { width: 7px; height: 7px; border-radius: 50%; }
.attention-row > div { display: flex; flex-direction: column; gap: 4px; }
.attention-row strong { color: #d9dde1; font-size: .74rem; font-weight: 600; }
.attention-row span { color: #68717b; font-size: .64rem; }
.attention-row > b { color: #aab0b7; font-size: .68rem; font-weight: 600; }

.transaction-panel { margin-top: 14px; }
.transaction-heading { min-height: 76px; padding: 17px 21px; border-bottom: 1px solid #20242a; display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.transaction-heading a { color: #8b939d; font-size: .68rem; }
.transaction-heading a:hover { color: #e8ebee; }
.transaction-heading a b { margin-left: 4px; color: #626b75; }
.transaction-table { min-width: 860px; }
.transaction-row { min-height: 52px; padding: 0 21px; display: grid; grid-template-columns: 1.35fr 1.35fr .8fr .65fr .75fr .55fr; align-items: center; gap: 16px; border-bottom: 1px solid #1d2025; color: #8b939c; font-size: .7rem; }
.transaction-row:last-child { border-bottom: 0; }
.transaction-head { min-height: 40px; color: #59626c; font-size: .61rem; font-weight: 650; letter-spacing: .035em; }
.transaction-row code { color: #a9b0b8; font: 500 .66rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.transaction-row > strong { color: #dfe2e5; font-size: .72rem; }
.transaction-row time { color: #68717b; }
.status-pill { min-width: 58px; height: 24px; padding: 0 8px; border-radius: 6px; display: inline-grid; place-items: center; color: #0c100e; font-size: .61rem; font-style: normal; font-weight: 700; }

.payment-demo-note { margin-top: 14px; padding: 0 4px; display: flex; align-items: center; justify-content: space-between; gap: 16px; color: #555e68; font-size: .62rem; }

@media (max-width: 1100px) {
  .payment-summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .payment-main-grid { grid-template-columns: 1fr; }
}

@media (max-width: 760px) {
  .payment-heading { min-height: auto; padding-top: 12px; align-items: flex-start; flex-direction: column; }
  .payment-heading-actions { width: 100%; }
  .outline-link { margin-left: auto; }
  .payment-heading h1 { font-size: 1.9rem; }
  .payment-summary-grid,
  .payment-secondary-grid { grid-template-columns: 1fr; }
  .transaction-panel { overflow-x: auto; }
  .payment-demo-note { align-items: flex-start; flex-direction: column; }
}
</style>
