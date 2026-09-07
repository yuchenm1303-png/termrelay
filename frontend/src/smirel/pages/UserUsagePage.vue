<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import WorkspaceNavIcon from '../components/WorkspaceNavIcon.vue'
import { api, getErrorMessage, previewMode } from '../core/api'

interface UsageRow {
  id?: number
  model?: string
  endpoint?: string
  total_tokens?: number
  actual_cost?: number
  created_at?: string
  [key: string]: unknown
}

const loading = ref(false)
const error = ref('')
const usage = ref<UsageRow[]>([])
const search = ref('')
const modelFilter = ref('all')

const previewRows: UsageRow[] = [
  { id: 1, model: 'gpt-5.6', endpoint: '/v1/responses', total_tokens: 18420, actual_cost: 0.0820, created_at: '2026-09-06 14:20' },
  { id: 2, model: 'claude-sonnet', endpoint: '/v1/messages', total_tokens: 9820, actual_cost: 0.0510, created_at: '2026-09-06 14:12' },
  { id: 3, model: 'gpt-5.6', endpoint: '/v1/responses', total_tokens: 12680, actual_cost: 0.0564, created_at: '2026-09-06 13:54' },
  { id: 4, model: 'gemini-2.5-pro', endpoint: '/v1/chat/completions', total_tokens: 7340, actual_cost: 0.0318, created_at: '2026-09-06 13:31' },
  { id: 5, model: 'claude-sonnet', endpoint: '/v1/messages', total_tokens: 6210, actual_cost: 0.0287, created_at: '2026-09-06 12:48' },
  { id: 6, model: 'gpt-5.6', endpoint: '/v1/responses', total_tokens: 15460, actual_cost: 0.0689, created_at: '2026-09-06 11:26' },
]

const modelOptions = computed(() => Array.from(new Set(usage.value.map((item) => item.model).filter(Boolean) as string[])))

const filteredUsage = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return usage.value.filter((item) => {
    if (modelFilter.value !== 'all' && item.model !== modelFilter.value) return false
    if (!keyword) return true
    return `${item.model || ''} ${item.endpoint || ''}`.toLowerCase().includes(keyword)
  })
})

const totalTokens = computed(() => filteredUsage.value.reduce((sum, item) => sum + Number(item.total_tokens || 0), 0))
const totalCost = computed(() => filteredUsage.value.reduce((sum, item) => sum + Number(item.actual_cost || 0), 0))
const averageCost = computed(() => filteredUsage.value.length ? totalCost.value / filteredUsage.value.length : 0)

const modelBreakdown = computed(() => {
  const totals = new Map<string, number>()
  filteredUsage.value.forEach((item) => {
    const model = item.model || 'Unknown'
    totals.set(model, (totals.get(model) || 0) + Number(item.total_tokens || 0))
  })
  const grandTotal = Math.max(1, Array.from(totals.values()).reduce((sum, value) => sum + value, 0))
  return Array.from(totals.entries())
    .map(([model, tokens]) => ({ model, tokens, percent: Math.round(tokens / grandTotal * 100) }))
    .sort((a, b) => b.tokens - a.tokens)
    .slice(0, 4)
})

function compact(value: number) {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toLocaleString()
}

function money(value: number) {
  return `$${value.toFixed(4)}`
}

function formatTime(value?: string) {
  if (!value) return '—'
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false })
}

function costShare(item: UsageRow) {
  if (!totalCost.value) return 0
  return Math.max(0, Math.min(100, Number(item.actual_cost || 0) / totalCost.value * 100))
}

async function load() {
  error.value = ''
  loading.value = true
  try {
    if (previewMode) {
      usage.value = previewRows
      return
    }
    const data = (await api.get<{ items?: UsageRow[] } | UsageRow[]>('/usage', { params: { page: 1, page_size: 50 } })).data
    usage.value = Array.isArray(data) ? data : (data.items || [])
  } catch (caught) {
    error.value = getErrorMessage(caught)
  } finally {
    loading.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="workspace-page user-usage-page">
    <header class="usage-heading">
      <div>
        <span class="usage-eyebrow"><i></i>USAGE ACTIVITY</span>
        <h1>用量与日志</h1>
        <p>查看最近请求、Token 消耗与费用，快速定位主要模型的使用情况。</p>
      </div>
      <button class="usage-refresh" type="button" :disabled="loading" @click="load">
        <WorkspaceNavIcon name="refresh" />
        <span>{{ loading ? '刷新中' : '刷新数据' }}</span>
      </button>
    </header>

    <p v-if="error" class="inline-error">{{ error }}</p>

    <div class="usage-summary-grid">
      <article class="usage-summary-card">
        <div class="usage-summary-head"><span>当前请求</span><i><WorkspaceNavIcon name="activity" /></i></div>
        <strong>{{ filteredUsage.length }}</strong>
        <footer><span>最近加载记录</span><b>REQUESTS</b></footer>
      </article>
      <article class="usage-summary-card">
        <div class="usage-summary-head"><span>Token 消耗</span><i><WorkspaceNavIcon name="chart" /></i></div>
        <strong>{{ compact(totalTokens) }}</strong>
        <footer><span>{{ totalTokens.toLocaleString() }} Tokens</span><b>TOKENS</b></footer>
      </article>
      <article class="usage-summary-card">
        <div class="usage-summary-head"><span>当前费用</span><i><WorkspaceNavIcon name="wallet" /></i></div>
        <strong>${{ totalCost.toFixed(4) }}</strong>
        <footer><span>当前列表累计</span><b>USD</b></footer>
      </article>
      <article class="usage-summary-card usage-summary-muted">
        <div class="usage-summary-head"><span>平均请求成本</span><i><WorkspaceNavIcon name="receipt" /></i></div>
        <strong>${{ averageCost.toFixed(4) }}</strong>
        <footer><span>按当前记录计算</span><b>AVG / REQ</b></footer>
      </article>
    </div>

    <section class="usage-insight-grid">
      <article class="usage-panel usage-breakdown-panel">
        <header class="usage-panel-head">
          <div><small>MODEL MIX</small><h2>模型用量构成</h2></div>
          <span>按 Token 统计</span>
        </header>
        <div v-if="modelBreakdown.length" class="usage-model-list">
          <div v-for="item in modelBreakdown" :key="item.model" class="usage-model-row">
            <div class="usage-model-meta"><strong>{{ item.model }}</strong><span>{{ compact(item.tokens) }} Tokens</span></div>
            <div class="usage-model-track"><i :style="{ width: `${item.percent}%` }"></i></div>
            <b>{{ item.percent }}%</b>
          </div>
        </div>
        <p v-else class="usage-empty-compact">暂无可统计记录</p>
      </article>

      <article class="usage-panel usage-filter-panel">
        <header class="usage-panel-head">
          <div><small>FILTER</small><h2>筛选日志</h2></div>
          <span>{{ filteredUsage.length }} 条</span>
        </header>
        <div class="usage-filter-body">
          <label class="usage-search">
            <WorkspaceNavIcon name="search" />
            <input v-model="search" type="search" placeholder="搜索模型或 Endpoint" />
          </label>
          <label class="usage-select-wrap">
            <span>模型</span>
            <select v-model="modelFilter">
              <option value="all">全部模型</option>
              <option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option>
            </select>
          </label>
        </div>
      </article>
    </section>

    <section class="usage-log-panel">
      <header class="usage-log-head">
        <div>
          <small>REQUEST LOG</small>
          <h2>最近请求</h2>
        </div>
        <div class="usage-log-total"><span>{{ compact(totalTokens) }} Tokens</span><b>{{ money(totalCost) }}</b></div>
      </header>

      <div class="usage-table-scroll">
        <div class="usage-table">
          <div class="usage-row usage-table-head">
            <span>时间</span><span>模型</span><span>Endpoint</span><span>Token</span><span>费用</span><span>成本占比</span>
          </div>
          <div v-for="(item, index) in filteredUsage" :key="item.id || index" class="usage-row">
            <time>{{ formatTime(item.created_at) }}</time>
            <div class="usage-model-cell"><i></i><strong>{{ item.model || '—' }}</strong></div>
            <code>{{ item.endpoint || '—' }}</code>
            <span class="usage-number">{{ Number(item.total_tokens || 0).toLocaleString() }}</span>
            <strong class="usage-cost">{{ money(Number(item.actual_cost || 0)) }}</strong>
            <div class="usage-share"><span><i :style="{ width: `${costShare(item)}%` }"></i></span><b>{{ costShare(item).toFixed(0) }}%</b></div>
          </div>
          <div v-if="!filteredUsage.length && !loading" class="usage-empty">没有匹配的用量记录</div>
        </div>
      </div>
    </section>

    <footer class="usage-footnote">
      <span>当前页面展示最近加载的请求记录；费用以最终账单结算为准。</span>
      <RouterLink to="/model-plaza">查看模型与价格 →</RouterLink>
    </footer>
  </section>
</template>

<style scoped>
.user-usage-page { width: 100%; max-width: 1280px; margin: 0 auto; padding-bottom: 36px; }
.usage-heading { min-height: 116px; display: flex; align-items: flex-end; justify-content: space-between; gap: 28px; margin-bottom: 27px; }
.usage-heading > div { min-width: 0; }
.usage-eyebrow { margin-bottom: 12px; display: inline-flex; align-items: center; gap: 8px; color: #69727d; font: 700 .64rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
.usage-eyebrow i { width: 6px; height: 6px; border-radius: 50%; background: #4bc494; box-shadow: 0 0 0 4px rgba(75,196,148,.07); }
.usage-heading h1 { margin: 0; color: #f5f6f8; font-size: 2.25rem; line-height: 1.05; letter-spacing: -.045em; font-weight: 720; }
.usage-heading p { margin: 10px 0 0; color: #79828d; font-size: .87rem; line-height: 1.65; }
.usage-refresh { min-height: 40px; padding: 0 14px; border: 1px solid #272c34; border-radius: 9px; display: inline-flex; align-items: center; gap: 8px; color: #aeb4bc; background: #101216; font: inherit; font-size: .75rem; cursor: pointer; transition: border-color .15s ease, color .15s ease, background-color .15s ease; }
.usage-refresh:hover:not(:disabled) { border-color: #373d46; color: #f2f4f6; background: #14171b; }
.usage-refresh:disabled { opacity: .58; cursor: wait; }
.usage-refresh :deep(.workspace-nav-icon) { width: 15px; height: 15px; }

.usage-summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0,1fr)); gap: 12px; }
.usage-summary-card { min-height: 150px; padding: 19px 20px 17px; border: 1px solid #22272e; border-radius: 12px; display: flex; flex-direction: column; background: #0f1115; box-shadow: inset 0 1px rgba(255,255,255,.012); }
.usage-summary-card:hover { border-color: #2d333c; }
.usage-summary-head { display: flex; align-items: center; justify-content: space-between; gap: 14px; }
.usage-summary-head > span { color: #858e99; font-size: .76rem; font-weight: 600; }
.usage-summary-head > i { width: 31px; height: 31px; border: 1px solid #293039; border-radius: 8px; display: grid; place-items: center; color: #78838f; background: #12151a; font-style: normal; }
.usage-summary-head :deep(.workspace-nav-icon) { width: 15px; height: 15px; }
.usage-summary-card > strong { margin-top: 17px; color: #f1f3f5; font-size: 1.72rem; line-height: 1; font-weight: 680; letter-spacing: -.035em; }
.usage-summary-card footer { margin-top: auto; padding-top: 14px; display: flex; justify-content: space-between; align-items: center; gap: 10px; color: #69727c; font-size: .66rem; }
.usage-summary-card footer b { color: #535d68; font: 700 .58rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .09em; }
.usage-summary-muted { background: #0d0f12; }

.usage-insight-grid { margin-top: 12px; display: grid; grid-template-columns: minmax(0,1.55fr) minmax(310px,.72fr); gap: 12px; }
.usage-panel, .usage-log-panel { border: 1px solid #22272e; border-radius: 12px; background: #0d0f13; overflow: hidden; }
.usage-panel-head { min-height: 70px; padding: 16px 20px; border-bottom: 1px solid #1f2329; display: flex; align-items: center; justify-content: space-between; gap: 18px; }
.usage-panel-head small, .usage-log-head small { display: block; margin-bottom: 5px; color: #626c76; font: 700 .58rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .11em; }
.usage-panel-head h2, .usage-log-head h2 { margin: 0; color: #e6e9ec; font-size: .94rem; line-height: 1.2; font-weight: 650; }
.usage-panel-head > span { color: #68717b; font-size: .66rem; }
.usage-model-list { padding: 8px 20px 13px; }
.usage-model-row { min-height: 49px; display: grid; grid-template-columns: 176px 1fr 38px; align-items: center; gap: 16px; }
.usage-model-meta { min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.usage-model-meta strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #d7dbe0; font-size: .73rem; font-weight: 600; }
.usage-model-meta span { color: #646e79; font-size: .61rem; }
.usage-model-track { height: 5px; border-radius: 999px; overflow: hidden; background: #1d2126; }
.usage-model-track i { display: block; height: 100%; border-radius: inherit; background: #58636e; }
.usage-model-row:first-child .usage-model-track i { background: #4b9279; }
.usage-model-row > b { color: #818a94; font-size: .65rem; font-weight: 650; text-align: right; }
.usage-empty-compact { padding: 22px 20px; margin: 0; color: #666f79; font-size: .72rem; }

.usage-filter-body { padding: 16px 20px 19px; display: grid; gap: 14px; }
.usage-search { min-height: 40px; padding: 0 12px; border: 1px solid #282d34; border-radius: 8px; display: flex; align-items: center; gap: 9px; color: #67717b; background: #111318; }
.usage-search:focus-within { border-color: #3a414a; }
.usage-search :deep(.workspace-nav-icon) { width: 15px; height: 15px; flex: 0 0 15px; }
.usage-search input { width: 100%; min-width: 0; border: 0; outline: 0; color: #dce0e4; background: transparent; font: inherit; font-size: .72rem; }
.usage-search input::placeholder { color: #5f6872; }
.usage-select-wrap { display: grid; grid-template-columns: auto 1fr; align-items: center; gap: 14px; }
.usage-select-wrap > span { color: #69727c; font-size: .68rem; }
.usage-select-wrap select { min-height: 40px; padding: 0 11px; border: 1px solid #282d34; border-radius: 8px; color: #cbd0d6; background: #111318; font: inherit; font-size: .72rem; outline: none; }

.usage-log-panel { margin-top: 12px; }
.usage-log-head { min-height: 76px; padding: 17px 20px; border-bottom: 1px solid #20242a; display: flex; align-items: center; justify-content: space-between; gap: 18px; }
.usage-log-total { display: flex; align-items: center; gap: 15px; color: #707984; font-size: .68rem; }
.usage-log-total b { color: #d4d8dd; font-size: .75rem; font-weight: 600; }
.usage-table-scroll { overflow-x: auto; }
.usage-table { min-width: 920px; }
.usage-row { min-height: 58px; padding: 0 20px; display: grid; grid-template-columns: 1.02fr 1.2fr 1.45fr .68fr .67fr 1fr; align-items: center; gap: 16px; border-bottom: 1px solid #1d2025; color: #8b949e; font-size: .71rem; }
.usage-row:last-child { border-bottom: 0; }
.usage-table-head { min-height: 42px; color: #59636e; font-size: .6rem; font-weight: 700; letter-spacing: .04em; }
.usage-row time { color: #747d87; }
.usage-model-cell { min-width: 0; display: flex; align-items: center; gap: 8px; }
.usage-model-cell i { width: 6px; height: 6px; flex: 0 0 6px; border-radius: 50%; background: #596570; }
.usage-model-cell strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #dce0e4; font-size: .73rem; font-weight: 620; }
.usage-row code { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #8b96a2; font: 500 .67rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; }
.usage-number { color: #9ba3ac; font-variant-numeric: tabular-nums; }
.usage-cost { color: #d5d9de; font-size: .72rem; font-weight: 600; font-variant-numeric: tabular-nums; }
.usage-share { display: grid; grid-template-columns: 1fr 30px; align-items: center; gap: 9px; }
.usage-share > span { height: 4px; border-radius: 999px; overflow: hidden; background: #1e2227; }
.usage-share > span i { display: block; height: 100%; border-radius: inherit; background: #596570; }
.usage-share b { color: #6e7781; font-size: .61rem; font-weight: 600; text-align: right; }
.usage-empty { min-height: 130px; display: grid; place-items: center; color: #66707a; font-size: .75rem; }
.usage-footnote { padding: 13px 3px 0; display: flex; align-items: center; justify-content: space-between; gap: 18px; color: #59626c; font-size: .62rem; }
.usage-footnote a { color: #7d8792; }
.usage-footnote a:hover { color: #dce0e4; }

@media (max-width: 1050px) {
  .usage-summary-grid { grid-template-columns: repeat(2,minmax(0,1fr)); }
  .usage-insight-grid { grid-template-columns: 1fr; }
}
@media (max-width: 720px) {
  .usage-heading { min-height: auto; padding-top: 12px; align-items: flex-start; flex-direction: column; }
  .usage-heading h1 { font-size: 1.9rem; }
  .usage-refresh { align-self: flex-end; }
  .usage-summary-grid { grid-template-columns: 1fr; }
  .usage-model-row { grid-template-columns: minmax(120px,1fr) 1fr 36px; gap: 10px; }
  .usage-footnote { align-items: flex-start; flex-direction: column; }
}
</style>
