<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, getErrorMessage, previewMode } from '../core/api'

interface UpstreamAccount {
  id: number
  name?: string
  notes?: string | null
  platform?: string
  type?: string
  status?: string
  schedulable?: boolean
  error_message?: string | null
  current_concurrency?: number
  concurrency?: number
  load_factor?: number
  priority?: number
  rate_multiplier?: number
  group_ids?: number[]
  last_used_at?: string | null
  expires_at?: number | null
  credentials_status?: string
  [key: string]: unknown
}

interface AccountListResponse {
  items?: UpstreamAccount[]
  total?: number
  page?: number
  page_size?: number
}

const loading = ref(false)
const error = ref('')
const accounts = ref<UpstreamAccount[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')
const platform = ref('')
const status = ref('')
const accountType = ref('')
const expandedAccountId = ref<number | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | undefined

const previewAccounts: UpstreamAccount[] = [
  { id: 12, name: 'OpenAI Primary', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, current_concurrency: 7, concurrency: 20, load_factor: 20, priority: 10, rate_multiplier: 1, group_ids: [1, 3], last_used_at: '2026-09-07T01:56:00Z', credentials_status: 'valid', notes: '主调度池 · Responses / Chat Completions' },
  { id: 18, name: 'OpenAI Reserve', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, current_concurrency: 2, concurrency: 12, load_factor: 12, priority: 20, rate_multiplier: 1, group_ids: [1], last_used_at: '2026-09-07T01:48:00Z', credentials_status: 'valid', notes: '高峰期备用池' },
  { id: 21, name: 'Claude Sonnet Pool', platform: 'anthropic', type: 'setup-token', status: 'active', schedulable: true, current_concurrency: 4, concurrency: 10, load_factor: 10, priority: 15, rate_multiplier: 1, group_ids: [2], last_used_at: '2026-09-07T01:51:00Z', credentials_status: 'valid', notes: 'Claude Code / Messages' },
  { id: 25, name: 'Gemini Flash', platform: 'gemini', type: 'oauth', status: 'active', schedulable: true, current_concurrency: 1, concurrency: 8, load_factor: 8, priority: 30, rate_multiplier: 0.9, group_ids: [4], last_used_at: '2026-09-07T01:32:00Z', credentials_status: 'valid' },
  { id: 27, name: 'Grok Direct', platform: 'grok', type: 'apikey', status: 'inactive', schedulable: false, current_concurrency: 0, concurrency: 6, load_factor: 6, priority: 50, rate_multiplier: 1.05, group_ids: [], last_used_at: '2026-09-06T23:16:00Z', credentials_status: 'valid', notes: '手动暂停' },
  { id: 31, name: 'OpenAI Canary', platform: 'openai', type: 'oauth', status: 'error', schedulable: false, current_concurrency: 0, concurrency: 4, load_factor: 4, priority: 90, rate_multiplier: 1, group_ids: [3], last_used_at: '2026-09-06T20:08:00Z', credentials_status: 'expired', error_message: 'OAuth credential refresh required', notes: '仅用于小流量验证' },
]

const platformOptions = computed(() => {
  const values = new Set(accounts.value.map((item) => String(item.platform || '').trim()).filter(Boolean))
  for (const item of previewMode ? previewAccounts : []) values.add(String(item.platform || ''))
  if (platform.value) values.add(platform.value)
  return Array.from(values).sort((a, b) => platformLabel(a).localeCompare(platformLabel(b)))
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
const schedulableCount = computed(() => accounts.value.filter((item) => item.status === 'active' && item.schedulable !== false).length)
const issueCount = computed(() => accounts.value.filter((item) => item.status === 'error' || item.schedulable === false).length)
const currentConcurrency = computed(() => accounts.value.reduce((sum, item) => sum + Number(item.current_concurrency || 0), 0))
const maxConcurrency = computed(() => accounts.value.reduce((sum, item) => sum + Number(item.concurrency || item.load_factor || 0), 0))
const providerCount = computed(() => new Set(accounts.value.map((item) => item.platform).filter(Boolean)).size)
const hasFilters = computed(() => Boolean(search.value.trim() || platform.value || status.value || accountType.value))

function platformLabel(value?: string) {
  const labels: Record<string, string> = {
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    gemini: 'Gemini',
    antigravity: 'Antigravity',
    grok: 'Grok',
    xai: 'xAI',
    ollama: 'Ollama',
  }
  const key = String(value || '').toLowerCase()
  return labels[key] || (value ? value.charAt(0).toUpperCase() + value.slice(1) : 'Unknown')
}

function platformMark(value?: string) {
  const label = platformLabel(value)
  return label.slice(0, 2).toUpperCase()
}

function accountTypeLabel(value?: string) {
  const labels: Record<string, string> = {
    oauth: 'OAuth',
    'setup-token': 'Setup Token',
    apikey: 'API Key',
    upstream: 'Upstream',
    bedrock: 'Bedrock',
    service_account: 'Service Account',
  }
  return labels[String(value || '').toLowerCase()] || value || '—'
}

function healthLabel(item: UpstreamAccount) {
  if (item.status === 'error') return '异常'
  if (item.status === 'inactive') return '已停用'
  if (item.schedulable === false) return '暂停调度'
  return '可调度'
}

function healthClass(item: UpstreamAccount) {
  if (item.status === 'error') return 'danger'
  if (item.status === 'inactive' || item.schedulable === false) return 'muted'
  return 'good'
}

function loadPercent(item: UpstreamAccount) {
  const max = Number(item.concurrency || item.load_factor || 0)
  if (max <= 0) return 0
  return Math.min(100, Math.round((Number(item.current_concurrency || 0) / max) * 100))
}

function formatTime(value?: string | null) {
  if (!value) return '从未使用'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

function formatExpiry(value?: number | null) {
  if (!value) return '未设置'
  const date = new Date(value * 1000)
  if (Number.isNaN(date.getTime())) return '未设置'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(date)
}

function groupLabel(item: UpstreamAccount) {
  const ids = Array.isArray(item.group_ids) ? item.group_ids : []
  if (!ids.length) return '未分组'
  if (ids.length === 1) return `分组 #${ids[0]}`
  return `${ids.length} 个分组`
}

function toggleExpanded(id: number) {
  expandedAccountId.value = expandedAccountId.value === id ? null : id
}

function resetFilters() {
  search.value = ''
  platform.value = ''
  status.value = ''
  accountType.value = ''
  page.value = 1
  void loadAccounts()
}

async function loadAccounts() {
  loading.value = true
  error.value = ''
  try {
    if (previewMode) {
      const keyword = search.value.trim().toLowerCase()
      const filtered = previewAccounts.filter((item) => {
        const matchesSearch = !keyword || `${item.name || ''} ${item.platform || ''} ${item.type || ''} ${item.id}`.toLowerCase().includes(keyword)
        const matchesPlatform = !platform.value || item.platform === platform.value
        const matchesStatus = !status.value || item.status === status.value
        const matchesType = !accountType.value || item.type === accountType.value
        return matchesSearch && matchesPlatform && matchesStatus && matchesType
      })
      total.value = filtered.length
      accounts.value = filtered.slice((page.value - 1) * pageSize, page.value * pageSize)
      return
    }

    const response = await api.get<AccountListResponse>('/admin/accounts', {
      params: {
        page: page.value,
        page_size: pageSize,
        search: search.value.trim() || undefined,
        platform: platform.value || undefined,
        status: status.value || undefined,
        type: accountType.value || undefined,
        sort_by: 'name',
        sort_order: 'asc',
      },
    })
    const data = response.data || {}
    accounts.value = Array.isArray(data.items) ? data.items : []
    total.value = Number(data.total || accounts.value.length)
  } catch (caught) {
    error.value = getErrorMessage(caught)
    accounts.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch([platform, status, accountType], () => {
  page.value = 1
  void loadAccounts()
})

watch(search, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    void loadAccounts()
  }, 260)
})

watch(page, () => void loadAccounts())
onMounted(() => void loadAccounts())
</script>

<template>
  <section class="admin-accounts-page">
    <header class="accounts-heading">
      <div>
        <h1>上游账户</h1>
        <p>统一查看账号健康状态、实时负载与调度配置。</p>
      </div>
      <button class="refresh-button" type="button" :disabled="loading" @click="loadAccounts">
        <span class="refresh-icon" :class="{ spinning: loading }">↻</span>
        {{ loading ? '正在刷新' : '刷新' }}
      </button>
    </header>

    <section class="account-metrics" aria-label="账户概览">
      <article>
        <span>账户总数</span>
        <strong>{{ total.toLocaleString() }}</strong>
        <small>当前筛选范围</small>
      </article>
      <article>
        <span>本页可调度</span>
        <strong>{{ schedulableCount }}</strong>
        <small>{{ issueCount ? `${issueCount} 个需关注` : '运行状态正常' }}</small>
      </article>
      <article>
        <span>实时并发</span>
        <strong>{{ currentConcurrency }}<em>/ {{ maxConcurrency || '—' }}</em></strong>
        <small>当前占用 / 配置上限</small>
      </article>
      <article>
        <span>本页平台</span>
        <strong>{{ providerCount }}</strong>
        <small>独立上游提供方</small>
      </article>
    </section>

    <section class="accounts-panel">
      <header class="accounts-toolbar">
        <div class="toolbar-title">
          <strong>账户池</strong>
          <span>{{ total }} 个账户</span>
        </div>
        <div class="toolbar-controls">
          <label class="search-control">
            <span>⌕</span>
            <input v-model="search" type="search" placeholder="搜索名称、平台或 ID" aria-label="搜索上游账户" />
          </label>
          <select v-model="platform" aria-label="筛选平台">
            <option value="">全部平台</option>
            <option v-for="item in platformOptions" :key="item" :value="item">{{ platformLabel(item) }}</option>
          </select>
          <select v-model="accountType" aria-label="筛选账户类型">
            <option value="">全部类型</option>
            <option value="oauth">OAuth</option>
            <option value="setup-token">Setup Token</option>
            <option value="apikey">API Key</option>
            <option value="upstream">Upstream</option>
            <option value="bedrock">Bedrock</option>
            <option value="service_account">Service Account</option>
          </select>
          <select v-model="status" aria-label="筛选状态">
            <option value="">全部状态</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="error">Error</option>
          </select>
          <button v-if="hasFilters" class="clear-button" type="button" @click="resetFilters">清除</button>
        </div>
      </header>

      <p v-if="error" class="accounts-error">{{ error }}</p>

      <div class="account-table" :class="{ loading }">
        <div class="account-table-head">
          <span>账户</span>
          <span>运行状态</span>
          <span>实时负载</span>
          <span>调度</span>
          <span>分组</span>
          <span>最近使用</span>
          <span></span>
        </div>

        <template v-for="item in accounts" :key="item.id">
          <button class="account-row" type="button" @click="toggleExpanded(item.id)">
            <span class="account-identity">
              <i class="provider-mark" :data-platform="String(item.platform || '').toLowerCase()">{{ platformMark(item.platform) }}</i>
              <span>
                <strong>{{ item.name || `Account #${item.id}` }}</strong>
                <small>{{ platformLabel(item.platform) }} · {{ accountTypeLabel(item.type) }} · #{{ item.id }}</small>
              </span>
            </span>
            <span class="health-cell">
              <i class="health-dot" :class="healthClass(item)"></i>
              <span>
                <strong>{{ healthLabel(item) }}</strong>
                <small>{{ item.status || 'unknown' }}</small>
              </span>
            </span>
            <span class="load-cell">
              <span><b>{{ Number(item.current_concurrency || 0) }}</b> / {{ Number(item.concurrency || item.load_factor || 0) || '—' }}</span>
              <i><b :style="{ width: `${loadPercent(item)}%` }"></b></i>
            </span>
            <span class="routing-cell">
              <strong>P{{ Number(item.priority || 0) }}</strong>
              <small>×{{ Number(item.rate_multiplier ?? 1).toFixed(2) }}</small>
            </span>
            <span class="group-cell">{{ groupLabel(item) }}</span>
            <span class="last-used">{{ formatTime(item.last_used_at) }}</span>
            <span class="row-chevron" :class="{ open: expandedAccountId === item.id }">›</span>
          </button>

          <div v-if="expandedAccountId === item.id" class="account-detail">
            <div>
              <span>账户说明</span>
              <strong>{{ item.notes || '暂无备注' }}</strong>
            </div>
            <div>
              <span>凭据状态</span>
              <strong>{{ item.credentials_status || '—' }}</strong>
            </div>
            <div>
              <span>到期时间</span>
              <strong>{{ formatExpiry(item.expires_at) }}</strong>
            </div>
            <div>
              <span>调度分组</span>
              <strong>{{ item.group_ids?.length ? item.group_ids.map((id) => `#${id}`).join(' · ') : '未分组' }}</strong>
            </div>
            <p v-if="item.error_message" class="account-warning">
              <span>异常原因</span>
              <strong>{{ item.error_message }}</strong>
            </p>
          </div>
        </template>

        <div v-if="!accounts.length && !loading" class="accounts-empty">
          <strong>{{ hasFilters ? '没有符合条件的账户' : '暂无上游账户' }}</strong>
          <span>{{ hasFilters ? '调整筛选条件后再试。' : '账户接入后会在这里集中显示运行状态和调度信息。' }}</span>
          <button v-if="hasFilters" type="button" @click="resetFilters">清除筛选</button>
        </div>

        <div v-if="loading && !accounts.length" class="accounts-loading">
          <i v-for="n in 5" :key="n"></i>
        </div>
      </div>

      <footer class="accounts-footer">
        <span>第 {{ page }} / {{ totalPages }} 页</span>
        <div>
          <button type="button" :disabled="page <= 1 || loading" @click="page -= 1">上一页</button>
          <button type="button" :disabled="page >= totalPages || loading" @click="page += 1">下一页</button>
        </div>
      </footer>
    </section>
  </section>
</template>

<style scoped>
.admin-accounts-page {
  --aa-surface: #101116;
  --aa-surface-soft: #0d0f13;
  --aa-surface-hover: #14161b;
  --aa-border: #23262d;
  --aa-border-strong: #30343d;
  --aa-text: #f3f5f7;
  --aa-text-soft: #c8cdd4;
  --aa-muted: #858d98;
  --aa-subtle: #626b76;
  --aa-blue: #5bbcf5;
  --aa-green: #43cd98;
  --aa-amber: #d7a95b;
  --aa-red: #e16c73;
  width: 100%;
  color: var(--aa-text);
  font-size: 15px;
}

.accounts-heading {
  min-height: 98px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 8px 0 20px;
}

.accounts-heading h1 {
  margin: 0;
  font-size: 2rem;
  line-height: 1.05;
  font-weight: 680;
  letter-spacing: -.045em;
}

.accounts-heading p {
  margin: 10px 0 0;
  color: var(--aa-muted);
  font-size: .86rem;
  line-height: 1.5;
}

.refresh-button,
.clear-button,
.accounts-footer button,
.accounts-empty button {
  border: 1px solid var(--aa-border-strong);
  background: #13151a;
  color: var(--aa-text-soft);
  cursor: pointer;
  transition: border-color .15s ease, background .15s ease, color .15s ease;
}

.refresh-button {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: .78rem;
  font-weight: 620;
}

.refresh-button:hover:not(:disabled),
.clear-button:hover,
.accounts-footer button:hover:not(:disabled),
.accounts-empty button:hover {
  border-color: #454a54;
  background: #181a20;
  color: #fff;
}

.refresh-button:disabled,
.accounts-footer button:disabled {
  opacity: .42;
  cursor: default;
}

.refresh-icon {
  display: inline-block;
  font-size: 1rem;
}

.refresh-icon.spinning {
  animation: account-spin .8s linear infinite;
}

.account-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border: 1px solid var(--aa-border);
  border-radius: 10px;
  background: var(--aa-surface);
  overflow: hidden;
}

.account-metrics article {
  min-height: 112px;
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.account-metrics article + article {
  border-left: 1px solid var(--aa-border);
}

.account-metrics span {
  color: var(--aa-muted);
  font-size: .72rem;
  font-weight: 600;
}

.account-metrics strong {
  margin-top: 7px;
  font-size: 1.55rem;
  line-height: 1;
  font-weight: 680;
  letter-spacing: -.03em;
}

.account-metrics em {
  margin-left: 4px;
  color: var(--aa-subtle);
  font-size: .85rem;
  font-style: normal;
  font-weight: 580;
}

.account-metrics small {
  margin-top: 8px;
  color: var(--aa-subtle);
  font-size: .68rem;
}

.accounts-panel {
  margin-top: 12px;
  border: 1px solid var(--aa-border);
  border-radius: 10px;
  background: var(--aa-surface);
  overflow: hidden;
}

.accounts-toolbar {
  min-height: 66px;
  padding: 11px 14px 11px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border-bottom: 1px solid var(--aa-border);
}

.toolbar-title {
  min-width: max-content;
  display: flex;
  align-items: baseline;
  gap: 9px;
}

.toolbar-title strong {
  font-size: .92rem;
  font-weight: 660;
}

.toolbar-title span {
  color: var(--aa-subtle);
  font-size: .68rem;
}

.toolbar-controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 7px;
  min-width: 0;
}

.search-control {
  width: min(300px, 25vw);
  height: 38px;
  padding: 0 11px;
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--aa-border);
  border-radius: 7px;
  background: #0b0d11;
  color: var(--aa-subtle);
}

.search-control:focus-within {
  border-color: #3a414c;
}

.search-control input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--aa-text);
  font-size: .76rem;
}

.search-control input::placeholder {
  color: #5d6570;
}

.toolbar-controls select {
  height: 38px;
  max-width: 150px;
  padding: 0 30px 0 11px;
  border: 1px solid var(--aa-border);
  border-radius: 7px;
  outline: 0;
  background: #0b0d11;
  color: var(--aa-text-soft);
  font-size: .73rem;
}

.clear-button {
  height: 38px;
  padding: 0 11px;
  border-radius: 7px;
  font-size: .72rem;
}

.accounts-error {
  margin: 12px 14px 0;
  padding: 10px 12px;
  border: 1px solid rgba(225, 108, 115, .28);
  border-radius: 7px;
  background: rgba(225, 108, 115, .08);
  color: #eaa0a5;
  font-size: .74rem;
}

.account-table {
  min-height: 304px;
}

.account-table.loading {
  opacity: .76;
}

.account-table-head,
.account-row {
  display: grid;
  grid-template-columns: minmax(240px, 1.55fr) minmax(118px, .76fr) minmax(120px, .72fr) minmax(90px, .55fr) minmax(100px, .66fr) minmax(110px, .68fr) 24px;
  gap: 18px;
  align-items: center;
}

.account-table-head {
  min-height: 40px;
  padding: 0 18px;
  border-bottom: 1px solid var(--aa-border);
  background: var(--aa-surface-soft);
  color: var(--aa-subtle);
  font-size: .65rem;
  font-weight: 650;
  letter-spacing: .02em;
}

.account-row {
  width: 100%;
  min-height: 72px;
  padding: 0 18px;
  border: 0;
  border-bottom: 1px solid #1d2026;
  background: transparent;
  color: var(--aa-text-soft);
  text-align: left;
  cursor: pointer;
  transition: background .14s ease;
}

.account-row:hover {
  background: var(--aa-surface-hover);
}

.account-identity,
.health-cell {
  min-width: 0;
  display: flex;
  align-items: center;
}

.account-identity {
  gap: 12px;
}

.provider-mark {
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  display: grid;
  place-items: center;
  border: 1px solid #2c3037;
  border-radius: 8px;
  background: #17191e;
  color: #aeb5be;
  font: 650 .62rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  font-style: normal;
  letter-spacing: .02em;
}

.provider-mark[data-platform="openai"] { color: #d6dbdf; }
.provider-mark[data-platform="anthropic"] { color: #d7b997; }
.provider-mark[data-platform="gemini"] { color: #92bff4; }
.provider-mark[data-platform="grok"],
.provider-mark[data-platform="xai"] { color: #d8d9dc; }

.account-identity > span,
.health-cell > span {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.account-identity strong {
  overflow: hidden;
  color: var(--aa-text);
  font-size: .80rem;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-identity small,
.health-cell small,
.routing-cell small {
  margin-top: 4px;
  color: var(--aa-subtle);
  font-size: .64rem;
}

.health-cell {
  gap: 8px;
}

.health-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 7px;
  border-radius: 50%;
  background: var(--aa-subtle);
  box-shadow: 0 0 0 3px rgba(98, 107, 118, .10);
}

.health-dot.good {
  background: var(--aa-green);
  box-shadow: 0 0 0 3px rgba(67, 205, 152, .10);
}

.health-dot.danger {
  background: var(--aa-red);
  box-shadow: 0 0 0 3px rgba(225, 108, 115, .10);
}

.health-dot.muted {
  background: var(--aa-amber);
  box-shadow: 0 0 0 3px rgba(215, 169, 91, .09);
}

.health-cell strong {
  color: var(--aa-text-soft);
  font-size: .73rem;
  font-weight: 620;
}

.load-cell {
  display: flex;
  flex-direction: column;
  gap: 7px;
  color: var(--aa-muted);
  font-size: .69rem;
}

.load-cell > span b {
  color: var(--aa-text-soft);
  font-size: .75rem;
  font-weight: 650;
}

.load-cell > i {
  width: 74px;
  height: 3px;
  overflow: hidden;
  border-radius: 99px;
  background: #272a31;
}

.load-cell > i b {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: #79828e;
}

.routing-cell {
  display: flex;
  flex-direction: column;
}

.routing-cell strong {
  color: var(--aa-text-soft);
  font: 650 .72rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
}

.group-cell,
.last-used {
  color: var(--aa-muted);
  font-size: .70rem;
}

.row-chevron {
  justify-self: end;
  color: var(--aa-subtle);
  font-size: 1.1rem;
  transform: rotate(0deg);
  transition: transform .15s ease, color .15s ease;
}

.row-chevron.open {
  color: var(--aa-text-soft);
  transform: rotate(90deg);
}

.account-detail {
  margin: -1px 0 0;
  padding: 15px 18px 16px 64px;
  display: grid;
  grid-template-columns: 1.4fr .7fr .8fr 1fr;
  gap: 12px 22px;
  border-bottom: 1px solid var(--aa-border);
  background: #0d0f13;
}

.account-detail > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.account-detail span {
  color: var(--aa-subtle);
  font-size: .62rem;
}

.account-detail strong {
  overflow: hidden;
  color: var(--aa-text-soft);
  font-size: .71rem;
  font-weight: 580;
  line-height: 1.45;
  text-overflow: ellipsis;
}

.account-warning {
  grid-column: 1 / -1;
  margin: 1px 0 0;
  padding: 10px 12px;
  display: flex;
  gap: 10px;
  border: 1px solid rgba(225, 108, 115, .22);
  border-radius: 7px;
  background: rgba(225, 108, 115, .06);
}

.account-warning strong {
  color: #e7a1a6;
}

.accounts-empty {
  min-height: 278px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--aa-muted);
  text-align: center;
}

.accounts-empty strong {
  color: var(--aa-text-soft);
  font-size: .88rem;
}

.accounts-empty span {
  margin-top: 7px;
  font-size: .72rem;
}

.accounts-empty button {
  margin-top: 14px;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 7px;
  font-size: .7rem;
}

.accounts-loading {
  padding: 12px 18px;
  display: grid;
  gap: 9px;
}

.accounts-loading i {
  display: block;
  height: 48px;
  border-radius: 7px;
  background: linear-gradient(90deg, #111318 20%, #17191f 50%, #111318 80%);
  background-size: 200% 100%;
  animation: skeleton-shift 1.15s linear infinite;
}

.accounts-footer {
  min-height: 54px;
  padding: 0 14px 0 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--aa-border);
  background: var(--aa-surface-soft);
}

.accounts-footer > span {
  color: var(--aa-subtle);
  font-size: .67rem;
}

.accounts-footer > div {
  display: flex;
  gap: 6px;
}

.accounts-footer button {
  min-height: 32px;
  padding: 0 11px;
  border-radius: 6px;
  font-size: .68rem;
}

@keyframes account-spin { to { transform: rotate(360deg); } }
@keyframes skeleton-shift { to { background-position: -200% 0; } }

@media (max-width: 1180px) {
  .account-table-head,
  .account-row {
    grid-template-columns: minmax(220px, 1.45fr) minmax(118px, .8fr) minmax(112px, .7fr) minmax(86px, .55fr) minmax(98px, .66fr) 24px;
  }

  .account-table-head > span:nth-child(5),
  .account-row > span:nth-child(5) {
    display: none;
  }

  .toolbar-controls select:nth-of-type(2) {
    display: none;
  }
}

@media (max-width: 900px) {
  .accounts-heading {
    min-height: auto;
  }

  .account-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .account-metrics article:nth-child(3) {
    border-left: 0;
    border-top: 1px solid var(--aa-border);
  }

  .account-metrics article:nth-child(4) {
    border-top: 1px solid var(--aa-border);
  }

  .accounts-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-controls {
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .search-control {
    width: 100%;
  }

  .account-table-head,
  .account-row {
    grid-template-columns: minmax(200px, 1.4fr) minmax(110px, .8fr) minmax(105px, .7fr) 24px;
  }

  .account-table-head > span:nth-child(4),
  .account-table-head > span:nth-child(5),
  .account-table-head > span:nth-child(6),
  .account-row > span:nth-child(4),
  .account-row > span:nth-child(5),
  .account-row > span:nth-child(6) {
    display: none;
  }

  .account-detail {
    padding-left: 18px;
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 620px) {
  .accounts-heading h1 { font-size: 1.7rem; }
  .account-metrics { grid-template-columns: 1fr; }
  .account-metrics article + article { border-left: 0; border-top: 1px solid var(--aa-border); }
  .toolbar-controls select { flex: 1 1 120px; max-width: none; }
  .account-table-head { display: none; }
  .account-row { grid-template-columns: minmax(0, 1fr) 24px; min-height: 78px; }
  .account-row > span:not(.account-identity):not(.row-chevron) { display: none; }
  .account-detail { grid-template-columns: 1fr; }
}
</style>
