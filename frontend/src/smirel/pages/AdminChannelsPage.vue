<script setup lang="ts">
import { computed, reactive, ref } from 'vue'

type ChannelState = 'active' | 'standby' | 'paused'

type ChannelRow = {
  id: number
  name: string
  provider: string
  mark: string
  protocol: string
  state: ChannelState
  priority: number
  multiplier: number
  scope: string
  endpoint: string
  note: string
}

type PricingRule = {
  id: number
  name: string
  scope: string
  multiplier: number
  description: string
  state: 'active' | 'draft'
}

const search = ref('')
const stateFilter = ref<'all' | ChannelState>('all')
const showEditor = ref(false)
const editingId = ref<number | null>(null)

const channels = ref<ChannelRow[]>([
  { id: 1, name: 'OpenAI Primary', provider: 'OpenAI', mark: 'O', protocol: 'Responses / Chat', state: 'active', priority: 10, multiplier: 1, scope: 'GPT / o-series', endpoint: 'api.openai.com', note: '主路由 · 默认承载 OpenAI 请求' },
  { id: 2, name: 'Anthropic Messages', provider: 'Anthropic', mark: 'A', protocol: 'Messages API', state: 'active', priority: 20, multiplier: 1.08, scope: 'Claude', endpoint: 'api.anthropic.com', note: 'Claude 系列独立消息协议' },
  { id: 3, name: 'Gemini Fast Lane', provider: 'Google', mark: 'G', protocol: 'Gemini API', state: 'active', priority: 30, multiplier: 0.92, scope: 'Gemini', endpoint: 'generativelanguage.googleapis.com', note: '低延迟通道 · 优先承载 Flash 系列' },
  { id: 4, name: 'Grok Direct', provider: 'xAI', mark: 'X', protocol: 'OpenAI Compatible', state: 'standby', priority: 40, multiplier: 1.05, scope: 'Grok', endpoint: 'api.x.ai', note: '备用通道 · 当前不参与首选路由' },
  { id: 5, name: 'OpenAI Reserve', provider: 'OpenAI', mark: 'O', protocol: 'Responses / Chat', state: 'active', priority: 60, multiplier: 1, scope: 'GPT / o-series', endpoint: 'api.openai.com', note: '故障转移与高峰期备用' },
  { id: 6, name: 'Validation Pool', provider: 'Internal', mark: 'S', protocol: 'OpenAI Compatible', state: 'paused', priority: 90, multiplier: 1, scope: 'Canary', endpoint: 'internal.preview', note: '验证用途 · 暂停参与生产路由' },
])

const pricingRules: PricingRule[] = [
  { id: 1, name: '基础结算', scope: '默认 · 全部模型', multiplier: 1, description: '未命中独立规则时使用的基础倍率。', state: 'active' },
  { id: 2, name: '高性能模型', scope: '旗舰 / 推理模型', multiplier: 1.12, description: '用于高成本、高算力模型的独立结算层。', state: 'active' },
  { id: 3, name: '轻量模型', scope: 'Flash / Mini / Haiku', multiplier: 0.9, description: '面向高频、低延迟场景的轻量倍率。', state: 'active' },
]

const routingSteps = computed(() => [
  { index: '01', title: '首选渠道', value: 'Priority 10–30', detail: `${channels.value.filter((item) => item.state === 'active' && item.priority <= 30).length} 个通道参与首选路由` },
  { index: '02', title: '备用渠道', value: 'Priority 40–70', detail: `${channels.value.filter((item) => item.state !== 'paused' && item.priority > 30 && item.priority <= 70).length} 个通道用于回退` },
  { index: '03', title: '隔离渠道', value: 'Priority 80+', detail: `${channels.value.filter((item) => item.priority > 70).length} 个通道处于验证或隔离层` },
])

const editor = reactive({
  name: '',
  provider: 'OpenAI',
  protocol: 'Responses / Chat',
  priority: 10,
  multiplier: 1,
  scope: 'GPT / o-series',
  endpoint: '',
  note: '',
})

const filteredChannels = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return channels.value.filter((item) => {
    const matchesState = stateFilter.value === 'all' || item.state === stateFilter.value
    const haystack = `${item.name} ${item.provider} ${item.protocol} ${item.scope} ${item.endpoint}`.toLowerCase()
    return matchesState && (!keyword || haystack.includes(keyword))
  })
})

const enabledCount = computed(() => channels.value.filter((item) => item.state === 'active').length)
const providerCount = computed(() => new Set(channels.value.map((item) => item.provider)).size)
const activeRuleCount = computed(() => pricingRules.filter((item) => item.state === 'active').length)
const healthyRatio = computed(() => Math.round((enabledCount.value / channels.value.length) * 100))

function stateLabel(state: ChannelState) {
  if (state === 'active') return '启用'
  if (state === 'standby') return '备用'
  return '暂停'
}

function openCreate() {
  editingId.value = null
  Object.assign(editor, {
    name: '',
    provider: 'OpenAI',
    protocol: 'Responses / Chat',
    priority: 10,
    multiplier: 1,
    scope: 'GPT / o-series',
    endpoint: '',
    note: '',
  })
  showEditor.value = true
}

function openEdit(item: ChannelRow) {
  editingId.value = item.id
  Object.assign(editor, {
    name: item.name,
    provider: item.provider,
    protocol: item.protocol,
    priority: item.priority,
    multiplier: item.multiplier,
    scope: item.scope,
    endpoint: item.endpoint,
    note: item.note,
  })
  showEditor.value = true
}

function saveDemoChannel() {
  const name = editor.name.trim() || `${editor.provider} Channel`
  if (editingId.value) {
    const item = channels.value.find((channel) => channel.id === editingId.value)
    if (item) {
      Object.assign(item, {
        ...editor,
        name,
        mark: editor.provider.slice(0, 1).toUpperCase(),
      })
    }
  } else {
    channels.value.unshift({
      id: Date.now(),
      name,
      provider: editor.provider,
      mark: editor.provider.slice(0, 1).toUpperCase(),
      protocol: editor.protocol,
      state: 'standby',
      priority: Number(editor.priority) || 10,
      multiplier: Number(editor.multiplier) || 1,
      scope: editor.scope,
      endpoint: editor.endpoint || 'preview.endpoint',
      note: editor.note || '演示配置 · 等待正式接口接入',
    })
  }
  showEditor.value = false
}

function toggleState(item: ChannelRow) {
  item.state = item.state === 'active' ? 'paused' : 'active'
}
</script>

<template>
  <section class="workspace-page channels-pricing-page">
    <header class="channels-heading">
      <div class="channels-heading-copy">
        <span class="channels-kicker">ROUTING &amp; PRICING</span>
        <div class="channels-title-line">
          <h1>渠道与价格</h1>
          <span class="demo-state"><i></i>演示配置</span>
        </div>
        <p>统一管理接入渠道、路由优先级与结算倍率。</p>
      </div>
      <button class="channels-primary-button" type="button" @click="openCreate">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M10 4v12M4 10h12" /></svg>
        新增渠道
      </button>
    </header>

    <section class="channels-overview" aria-label="渠道配置概览">
      <article class="overview-primary">
        <div class="overview-label"><span>渠道总数</span><i>CONFIGURED</i></div>
        <strong>{{ channels.length }}<small>个</small></strong>
        <p>{{ providerCount }} 家服务商已纳入路由</p>
      </article>
      <article>
        <div class="overview-label"><span>启用渠道</span><i>{{ healthyRatio }}%</i></div>
        <strong>{{ enabledCount }}<small>/ {{ channels.length }}</small></strong>
        <div class="overview-progress"><i :style="{ width: `${healthyRatio}%` }"></i></div>
      </article>
      <article>
        <div class="overview-label"><span>默认倍率</span><i>BILLING</i></div>
        <strong>1.00<small>×</small></strong>
        <p>基础结算规则</p>
      </article>
      <article>
        <div class="overview-label"><span>计价策略</span><i>ACTIVE</i></div>
        <strong>{{ activeRuleCount }}<small>条</small></strong>
        <p>按模型范围匹配</p>
      </article>
    </section>

    <section class="channels-panel">
      <header class="channels-toolbar">
        <div class="toolbar-copy">
          <div><strong>渠道配置</strong><span>{{ filteredChannels.length }} / {{ channels.length }}</span></div>
          <small>倍率用于演示结算结构，正式价格接入后由服务端返回。</small>
        </div>
        <div class="toolbar-actions">
          <label class="channels-search">
            <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="m13 13 4 4" /></svg>
            <input v-model="search" type="search" placeholder="搜索渠道、服务商或协议" aria-label="搜索渠道" />
          </label>
          <label class="channels-filter">
            <span>状态</span>
            <select v-model="stateFilter" aria-label="筛选渠道状态">
              <option value="all">全部</option>
              <option value="active">启用</option>
              <option value="standby">备用</option>
              <option value="paused">暂停</option>
            </select>
            <svg viewBox="0 0 16 16" aria-hidden="true"><path d="m4 6 4 4 4-4" /></svg>
          </label>
        </div>
      </header>

      <div class="channels-table-wrap">
        <table class="channels-table">
          <thead>
            <tr>
              <th>渠道</th>
              <th>接入协议</th>
              <th>状态</th>
              <th>优先级</th>
              <th>倍率</th>
              <th>模型范围</th>
              <th><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredChannels" :key="item.id">
              <td>
                <div class="channel-identity">
                  <span class="provider-mark">{{ item.mark }}</span>
                  <span><strong>{{ item.name }}</strong><small>{{ item.provider }} · {{ item.endpoint }}</small></span>
                </div>
              </td>
              <td><span class="protocol-cell">{{ item.protocol }}</span></td>
              <td><span class="channel-state" :class="item.state"><i></i>{{ stateLabel(item.state) }}</span></td>
              <td><span class="priority-cell"><b>P{{ item.priority }}</b><small>{{ item.priority <= 30 ? '首选' : item.priority <= 70 ? '回退' : '隔离' }}</small></span></td>
              <td><strong class="multiplier-cell">{{ item.multiplier.toFixed(2) }}×</strong></td>
              <td><span class="scope-chip">{{ item.scope }}</span></td>
              <td>
                <div class="row-actions">
                  <button type="button" :title="item.state === 'active' ? '暂停渠道' : '启用渠道'" @click="toggleState(item)">
                    <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M4 10h12" /><path v-if="item.state !== 'active'" d="M10 4v12" /></svg>
                  </button>
                  <button type="button" title="配置渠道" @click="openEdit(item)">
                    <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M12.8 4.2 15.8 7.2 7.1 15.9 3.8 16.2 4.1 12.9 12.8 4.2Z" /><path d="m11.6 5.4 3 3" /></svg>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filteredChannels.length">
              <td colspan="7" class="channels-empty">没有符合当前条件的渠道。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div class="channels-lower-grid">
      <section class="pricing-panel">
        <header class="subpanel-head">
          <div><span>PRICING RULES</span><strong>计费倍率</strong></div>
          <small>按范围从上到下匹配</small>
        </header>
        <div class="pricing-rule-list">
          <article v-for="rule in pricingRules" :key="rule.id">
            <div class="rule-main">
              <span class="rule-index">0{{ rule.id }}</span>
              <span><strong>{{ rule.name }}</strong><small>{{ rule.scope }}</small></span>
            </div>
            <p>{{ rule.description }}</p>
            <div class="rule-rate"><small>倍率</small><strong>{{ rule.multiplier.toFixed(2) }}×</strong></div>
          </article>
        </div>
      </section>

      <section class="routing-panel">
        <header class="subpanel-head">
          <div><span>ROUTING ORDER</span><strong>路由优先级</strong></div>
          <small>低数值优先</small>
        </header>
        <div class="routing-flow">
          <article v-for="step in routingSteps" :key="step.index">
            <span class="routing-index">{{ step.index }}</span>
            <div><strong>{{ step.title }}</strong><small>{{ step.detail }}</small></div>
            <b>{{ step.value }}</b>
          </article>
        </div>
        <footer class="routing-note">
          <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="7" /><path d="M10 9v4M10 6.5v.2" /></svg>
          <span>后续接入调度接口后，此处直接展示服务端实际优先级与回退关系。</span>
        </footer>
      </section>
    </div>

    <div v-if="showEditor" class="channel-editor-backdrop" @click.self="showEditor = false">
      <section class="channel-editor" role="dialog" aria-modal="true" aria-label="渠道配置">
        <header>
          <div><span>CHANNEL CONFIG</span><strong>{{ editingId ? '编辑渠道' : '新增渠道' }}</strong><small>当前为演示配置，不会写入服务端。</small></div>
          <button type="button" aria-label="关闭" @click="showEditor = false"><svg viewBox="0 0 20 20"><path d="m5 5 10 10M15 5 5 15" /></svg></button>
        </header>
        <div class="editor-grid">
          <label class="editor-wide"><span>渠道名称</span><input v-model="editor.name" placeholder="例如 OpenAI Primary" /></label>
          <label><span>服务商</span><select v-model="editor.provider"><option>OpenAI</option><option>Anthropic</option><option>Google</option><option>xAI</option><option>Internal</option></select></label>
          <label><span>接入协议</span><select v-model="editor.protocol"><option>Responses / Chat</option><option>Messages API</option><option>Gemini API</option><option>OpenAI Compatible</option></select></label>
          <label><span>优先级</span><input v-model.number="editor.priority" type="number" min="1" max="999" /></label>
          <label><span>计费倍率</span><input v-model.number="editor.multiplier" type="number" min="0.01" step="0.01" /></label>
          <label><span>模型范围</span><input v-model="editor.scope" /></label>
          <label><span>上游地址</span><input v-model="editor.endpoint" placeholder="api.example.com" /></label>
          <label class="editor-wide"><span>备注</span><input v-model="editor.note" placeholder="描述这个渠道的用途" /></label>
        </div>
        <footer>
          <button class="editor-cancel" type="button" @click="showEditor = false">取消</button>
          <button class="editor-save" type="button" @click="saveDemoChannel">保存演示配置</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<style scoped>
.channels-pricing-page {
  --cp-surface: #101116;
  --cp-surface-soft: #0c0e12;
  --cp-surface-raised: #14161b;
  --cp-border: rgba(255, 255, 255, 0.075);
  --cp-border-strong: rgba(255, 255, 255, 0.11);
  --cp-text: #f4f5f7;
  --cp-muted: #858b95;
  --cp-subtle: #5d646e;
  --cp-accent: #63b9ea;
  --cp-accent-soft: rgba(99, 185, 234, 0.11);
  --cp-green: #65d8a7;
  --cp-amber: #d8aa61;
  color: var(--cp-text);
  font-size: 15px;
}

.channels-heading {
  min-height: 106px;
  padding: 4px 0 24px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.channels-kicker,
.subpanel-head span,
.channel-editor header span {
  display: block;
  font-size: 10px;
  line-height: 1;
  font-weight: 700;
  letter-spacing: 0.13em;
  color: #65707c;
}

.channels-title-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
}

.channels-title-line h1 {
  margin: 0;
  font-size: 28px;
  line-height: 1.08;
  font-weight: 650;
  letter-spacing: -0.035em;
}

.channels-heading-copy > p {
  margin: 9px 0 0;
  font-size: 13px;
  color: var(--cp-muted);
}

.demo-state {
  height: 24px;
  padding: 0 9px;
  border: 1px solid rgba(99, 185, 234, 0.18);
  border-radius: 999px;
  background: var(--cp-accent-soft);
  color: #8fcef2;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
}

.demo-state i,
.channel-state i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.channels-primary-button {
  min-height: 38px;
  margin-top: 7px;
  padding: 0 14px;
  border: 1px solid rgba(99, 185, 234, 0.28);
  border-radius: 9px;
  background: #182832;
  color: #b7e0f6;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 650;
  cursor: pointer;
  transition: background .16s ease, border-color .16s ease, transform .16s ease;
}

.channels-primary-button:hover {
  background: #1d303d;
  border-color: rgba(99, 185, 234, 0.42);
  transform: translateY(-1px);
}

.channels-primary-button svg,
.row-actions svg,
.channel-editor button svg,
.routing-note svg,
.channels-search svg,
.channels-filter svg {
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.channels-primary-button svg { width: 16px; height: 16px; }

.channels-overview {
  display: grid;
  grid-template-columns: 1.15fr repeat(3, 1fr);
  border: 1px solid var(--cp-border);
  border-radius: 12px;
  overflow: hidden;
  background: var(--cp-surface);
}

.channels-overview article {
  min-height: 112px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.channels-overview article + article { border-left: 1px solid var(--cp-border); }

.overview-primary { background: linear-gradient(115deg, rgba(99, 185, 234, 0.06), transparent 60%); }

.overview-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--cp-muted);
  font-size: 12px;
}

.overview-label i {
  font-style: normal;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: .08em;
  color: #5e6771;
}

.channels-overview strong {
  margin-top: 8px;
  font-size: 25px;
  line-height: 1;
  font-weight: 650;
  letter-spacing: -0.035em;
}

.channels-overview strong small {
  margin-left: 5px;
  font-size: 12px;
  font-weight: 550;
  color: #737a84;
}

.channels-overview p {
  margin: 9px 0 0;
  font-size: 11px;
  color: var(--cp-subtle);
}

.overview-progress {
  width: 100%;
  height: 3px;
  margin-top: 13px;
  border-radius: 999px;
  background: #24272d;
  overflow: hidden;
}

.overview-progress i { display: block; height: 100%; border-radius: inherit; background: var(--cp-green); }

.channels-panel,
.pricing-panel,
.routing-panel {
  border: 1px solid var(--cp-border);
  border-radius: 12px;
  background: var(--cp-surface);
}

.channels-panel { margin-top: 12px; overflow: hidden; }

.channels-toolbar {
  min-height: 76px;
  padding: 14px 16px 14px 20px;
  border-bottom: 1px solid var(--cp-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.toolbar-copy { min-width: 0; }

.toolbar-copy > div { display: flex; align-items: center; gap: 9px; }
.toolbar-copy strong { font-size: 14px; font-weight: 650; }
.toolbar-copy > div span {
  min-width: 28px;
  height: 20px;
  padding: 0 6px;
  border-radius: 6px;
  background: #181a20;
  color: #777e88;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
}
.toolbar-copy small { display: block; margin-top: 5px; font-size: 11px; color: var(--cp-subtle); }

.toolbar-actions { display: flex; align-items: center; gap: 8px; }

.channels-search,
.channels-filter {
  height: 36px;
  border: 1px solid var(--cp-border);
  border-radius: 8px;
  background: #0d0f13;
  color: #737b86;
  display: flex;
  align-items: center;
}

.channels-search { width: 236px; padding: 0 10px; gap: 8px; }
.channels-search svg { width: 15px; height: 15px; flex: 0 0 auto; }
.channels-search input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #d8dbe0;
  font-size: 12px;
}
.channels-search input::placeholder { color: #505760; }

.channels-filter { position: relative; min-width: 126px; padding-left: 10px; }
.channels-filter > span { font-size: 10px; color: #555d66; margin-right: 4px; }
.channels-filter select {
  min-width: 60px;
  height: 100%;
  padding: 0 24px 0 5px;
  appearance: none;
  border: 0;
  outline: 0;
  background: transparent;
  color: #b7bbc1;
  font-size: 12px;
}
.channels-filter svg { position: absolute; right: 8px; width: 13px; height: 13px; pointer-events: none; }

.channels-table-wrap { overflow-x: auto; }
.channels-table { width: 100%; min-width: 1050px; border-collapse: collapse; table-layout: fixed; }
.channels-table th {
  height: 35px;
  padding: 0 14px;
  border-bottom: 1px solid rgba(255,255,255,.055);
  color: #555d67;
  text-align: left;
  font-size: 10px;
  font-weight: 650;
  letter-spacing: .035em;
}
.channels-table th:first-child,
.channels-table td:first-child { padding-left: 20px; width: 24%; }
.channels-table th:nth-child(2) { width: 16%; }
.channels-table th:nth-child(3) { width: 10%; }
.channels-table th:nth-child(4) { width: 10%; }
.channels-table th:nth-child(5) { width: 9%; }
.channels-table th:nth-child(6) { width: 20%; }
.channels-table th:last-child { width: 78px; }
.channels-table td {
  height: 64px;
  padding: 0 14px;
  border-bottom: 1px solid rgba(255,255,255,.05);
  color: #969ca5;
  font-size: 12px;
  vertical-align: middle;
}
.channels-table tbody tr:last-child td { border-bottom: 0; }
.channels-table tbody tr:hover td { background: rgba(255,255,255,.012); }

.channel-identity { display: flex; align-items: center; gap: 11px; min-width: 0; }
.provider-mark {
  width: 31px;
  height: 31px;
  flex: 0 0 auto;
  border: 1px solid #292d34;
  border-radius: 8px;
  background: #17191e;
  color: #c8ccd2;
  display: grid;
  place-items: center;
  font-size: 11px;
  font-weight: 750;
}
.channel-identity > span:last-child { min-width: 0; display: flex; flex-direction: column; }
.channel-identity strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #e4e6e9; font-size: 12.5px; font-weight: 650; }
.channel-identity small { margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #5e6570; font-size: 10.5px; }
.protocol-cell { color: #9da3ac; }
.channel-state { display: inline-flex; align-items: center; gap: 6px; font-size: 11px; font-weight: 600; }
.channel-state.active { color: var(--cp-green); }
.channel-state.standby { color: var(--cp-amber); }
.channel-state.paused { color: #68707b; }
.priority-cell { display: flex; align-items: baseline; gap: 6px; }
.priority-cell b { color: #c7cbd1; font-size: 11px; font-weight: 650; }
.priority-cell small { color: #555d66; font-size: 9.5px; }
.multiplier-cell { color: #d8dbe0; font-size: 12px; font-weight: 650; }
.scope-chip {
  max-width: 100%;
  min-height: 24px;
  padding: 0 8px;
  border: 1px solid #272a31;
  border-radius: 6px;
  background: #15171c;
  color: #898f98;
  display: inline-flex;
  align-items: center;
  font-size: 10.5px;
}
.row-actions { display: flex; align-items: center; justify-content: flex-end; gap: 4px; }
.row-actions button {
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 7px;
  background: transparent;
  color: #626a75;
  display: grid;
  place-items: center;
  cursor: pointer;
}
.row-actions button:hover { border-color: #292d34; background: #17191e; color: #c4c8ce; }
.row-actions svg { width: 14px; height: 14px; }
.channels-empty { height: 96px !important; text-align: center; color: #5d646e !important; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }

.channels-lower-grid { display: grid; grid-template-columns: 1.15fr .85fr; gap: 12px; margin-top: 12px; }
.pricing-panel,
.routing-panel { overflow: hidden; }
.subpanel-head {
  min-height: 68px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--cp-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.subpanel-head strong { display: block; margin-top: 6px; font-size: 14px; font-weight: 650; }
.subpanel-head small { color: #59616b; font-size: 10.5px; }
.pricing-rule-list { padding: 4px 18px; }
.pricing-rule-list article {
  min-height: 62px;
  display: grid;
  grid-template-columns: 1.15fr 1.45fr 76px;
  align-items: center;
  gap: 14px;
  border-bottom: 1px solid rgba(255,255,255,.05);
}
.pricing-rule-list article:last-child { border-bottom: 0; }
.rule-main { display: flex; align-items: center; gap: 10px; min-width: 0; }
.rule-index { color: #424950; font-size: 9px; font-weight: 700; letter-spacing: .08em; }
.rule-main > span:last-child { min-width: 0; display: flex; flex-direction: column; }
.rule-main strong { color: #d8dbe0; font-size: 11.5px; font-weight: 650; }
.rule-main small { margin-top: 4px; color: #5c636d; font-size: 9.5px; }
.pricing-rule-list p { margin: 0; color: #747b85; font-size: 10.5px; line-height: 1.45; }
.rule-rate { justify-self: end; text-align: right; }
.rule-rate small { display: block; color: #525963; font-size: 9px; }
.rule-rate strong { display: block; margin-top: 3px; color: #cdd1d6; font-size: 12px; font-weight: 650; }
.routing-flow { padding: 4px 18px; }
.routing-flow article {
  min-height: 62px;
  display: grid;
  grid-template-columns: 28px 1fr auto;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid rgba(255,255,255,.05);
}
.routing-flow article:last-child { border-bottom: 0; }
.routing-index { width: 24px; height: 24px; border: 1px solid #282c32; border-radius: 7px; display: grid; place-items: center; color: #646c76; font-size: 9px; font-weight: 700; }
.routing-flow div { min-width: 0; display: flex; flex-direction: column; }
.routing-flow div strong { color: #d5d8dd; font-size: 11.5px; font-weight: 650; }
.routing-flow div small { margin-top: 4px; color: #5e6570; font-size: 9.5px; }
.routing-flow article > b { color: #858c96; font-size: 10px; font-weight: 600; }
.routing-note { min-height: 45px; padding: 9px 18px; border-top: 1px solid var(--cp-border); background: #0d0f13; display: flex; align-items: center; gap: 8px; color: #5c646e; font-size: 9.5px; line-height: 1.4; }
.routing-note svg { width: 14px; height: 14px; flex: 0 0 auto; }

.channel-editor-backdrop {
  position: fixed;
  z-index: 120;
  inset: 0;
  padding: 24px;
  background: rgba(0, 0, 0, .62);
  display: grid;
  place-items: center;
}
.channel-editor {
  width: min(620px, 100%);
  border: 1px solid #292d34;
  border-radius: 14px;
  background: #101116;
  box-shadow: 0 28px 90px rgba(0,0,0,.44);
  overflow: hidden;
}
.channel-editor > header {
  min-height: 82px;
  padding: 17px 20px;
  border-bottom: 1px solid var(--cp-border);
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.channel-editor header strong { display: block; margin-top: 7px; font-size: 16px; }
.channel-editor header small { display: block; margin-top: 5px; color: #616873; font-size: 10.5px; }
.channel-editor header button { width: 30px; height: 30px; border: 1px solid #282c32; border-radius: 8px; background: #15171c; color: #777f89; display: grid; place-items: center; cursor: pointer; }
.channel-editor header button svg { width: 14px; height: 14px; }
.editor-grid { padding: 18px 20px; display: grid; grid-template-columns: 1fr 1fr; gap: 13px; }
.editor-grid label { display: flex; flex-direction: column; gap: 6px; }
.editor-grid label > span { color: #777e88; font-size: 10.5px; font-weight: 600; }
.editor-grid input,
.editor-grid select {
  width: 100%;
  height: 38px;
  padding: 0 10px;
  border: 1px solid #272b32;
  border-radius: 8px;
  outline: 0;
  background: #0c0e12;
  color: #d8dbe0;
  font-size: 12px;
}
.editor-grid input:focus,
.editor-grid select:focus { border-color: rgba(99,185,234,.45); box-shadow: 0 0 0 3px rgba(99,185,234,.06); }
.editor-wide { grid-column: 1 / -1; }
.channel-editor > footer { min-height: 62px; padding: 12px 20px; border-top: 1px solid var(--cp-border); display: flex; align-items: center; justify-content: flex-end; gap: 8px; }
.editor-cancel,
.editor-save { min-height: 36px; padding: 0 13px; border-radius: 8px; font-size: 11.5px; font-weight: 650; cursor: pointer; }
.editor-cancel { border: 1px solid #282c32; background: #15171c; color: #8c939d; }
.editor-save { border: 1px solid rgba(99,185,234,.26); background: #1a2a35; color: #b3ddf4; }

@media (max-width: 1100px) {
  .channels-overview { grid-template-columns: 1fr 1fr; }
  .channels-overview article:nth-child(3) { border-left: 0; border-top: 1px solid var(--cp-border); }
  .channels-overview article:nth-child(4) { border-top: 1px solid var(--cp-border); }
  .channels-lower-grid { grid-template-columns: 1fr; }
  .channels-toolbar { align-items: flex-start; flex-direction: column; }
  .toolbar-actions { width: 100%; }
  .channels-search { flex: 1; }
}

@media (max-width: 720px) {
  .channels-heading { min-height: auto; padding-bottom: 20px; flex-direction: column; }
  .channels-primary-button { margin-top: 0; }
  .channels-title-line { align-items: flex-start; flex-direction: column; gap: 8px; }
  .channels-overview { grid-template-columns: 1fr; }
  .channels-overview article + article { border-left: 0; border-top: 1px solid var(--cp-border); }
  .toolbar-actions { align-items: stretch; flex-direction: column; }
  .channels-search,
  .channels-filter { width: 100%; }
  .pricing-rule-list article { grid-template-columns: 1fr auto; padding: 10px 0; }
  .pricing-rule-list p { grid-column: 1 / -1; grid-row: 2; }
  .editor-grid { grid-template-columns: 1fr; }
  .editor-wide { grid-column: auto; }
}
</style>
