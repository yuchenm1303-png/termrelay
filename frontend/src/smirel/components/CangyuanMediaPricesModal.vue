<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { api, getErrorMessage } from '../core/api'

type MediaKind = 'image_generation' | 'image_edit' | 'video_generation'
type CatalogItem = {
  id: string; kind: 'image_generation' | 'video_generation'
  billing_mode: string; model_price: number; request_unit: string; tiered: boolean
}
type Catalog = { pricing_version: string; items: CatalogItem[] }
type PriceSnapshot = {
  pricing_version: string; currency: string; tier: string; units: number
  unit_sale_price_usd: string; unit_provider_cost_usd: string; verified_by_admin: boolean
}
type PriceRule = {
  id: string; model: string; kind: MediaKind
  options: Record<string, unknown>; snapshot: PriceSnapshot; enabled: boolean
}

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: [] }>()
const loading = ref(false)
const saving = ref(false)
const deleting = ref('')
const error = ref('')
const notice = ref('')
const query = ref('')
const catalog = ref<Catalog | null>(null)
const rules = ref<PriceRule[]>([])
const editorOpen = ref(false)
const editOriginalId = ref('')
const reviewed = ref(false)
const optionsJSON = ref('{"n":1}')
const draft = reactive({
  id: '', model: '', kind: 'image_generation' as MediaKind,
  tier: '', pricingVersion: '', units: 1, salePrice: '', providerCost: '',
  enabled: false,
})
const filteredCatalog = computed(() => {
  const q = query.value.trim().toLowerCase()
  return (catalog.value?.items || []).filter(item =>
    (!q || item.id.toLowerCase().includes(q)) &&
    (draft.kind !== 'video_generation' || item.kind === 'video_generation') &&
    (draft.kind === 'video_generation' || item.kind === 'image_generation'),
  ).slice(0, 90)
})
const activeCount = computed(() => rules.value.filter(rule => rule.enabled).length)
const editing = computed(() => Boolean(editOriginalId.value))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await api.get<{ data: PriceRule[] }>('/admin/cangyuan-media/prices')
    rules.value = Array.isArray(result.data?.data) ? result.data.data : []
  } catch (caught) {
    error.value = '无法读取已审核价格：' + getErrorMessage(caught)
    loading.value = false
    return
  }
  try {
    const response = await api.get<Catalog>('/admin/cangyuan-media/catalog')
    catalog.value = response.data
  } catch (caught) {
    catalog.value = null
    notice.value = '官方模型目录暂时不可用；仍可查看或编辑已有规则。请确认模型、版本及供应商计费后再发布。'
  } finally {
    loading.value = false
  }
}

watch(() => props.show, visible => {
  if (visible) {
    notice.value = ''
    editorOpen.value = false
    void load()
  }
}, { immediate: true })

function resetDraft() {
  editOriginalId.value = ''
  reviewed.value = false
  Object.assign(draft, {
    id: '', model: '', kind: 'image_generation', tier: '',
    pricingVersion: catalog.value?.pricing_version || '', units: 1,
    salePrice: '', providerCost: '', enabled: false,
  })
  optionsJSON.value = '{"n":1}'
  editorOpen.value = true
  error.value = ''
}

function editRule(rule: PriceRule) {
  editOriginalId.value = rule.id
  reviewed.value = false
  Object.assign(draft, {
    id: rule.id, model: rule.model, kind: rule.kind,
    tier: rule.snapshot.tier, pricingVersion: rule.snapshot.pricing_version,
    units: rule.snapshot.units,
    salePrice: rule.snapshot.unit_sale_price_usd,
    providerCost: rule.snapshot.unit_provider_cost_usd,
    enabled: rule.enabled,
  })
  optionsJSON.value = JSON.stringify(rule.options, null, 2)
  error.value = ''
  editorOpen.value = true
}

function chooseModel(item: CatalogItem) {
  draft.model = item.id
  // Image models can be sold under either generation or edit endpoints.
  // Keep an explicitly selected image_edit kind rather than silently
  // changing it back to image_generation when selecting a model.
  if (item.kind === 'video_generation') draft.kind = 'video_generation'
  else if (draft.kind !== 'image_edit') draft.kind = 'image_generation'
  if (!draft.id && !editing.value) {
    const prefix = draft.kind === 'image_edit' ? 'edit-' : draft.kind === 'video_generation' ? 'video-' : ''
    draft.id = (prefix + item.id.toLowerCase().replace(/[^a-z0-9_-]+/g, '-').slice(0, 60) + '-v1').slice(0, 80)
  }
}

function changeKind() {
  if (!editing.value) {
    draft.model = ''
    draft.id = ''
    optionsJSON.value = draft.kind === 'video_generation' ? '{}' : '{"n":1}'
    draft.units = 1
  }
}

async function save() {
  if (saving.value) return
  error.value = ''
  notice.value = ''
  if (!/^[a-z0-9][a-z0-9_-]{0,79}$/.test(draft.id)) {
    error.value = '规则 ID 只能使用小写字母、数字、下划线和连字符（最多 80 位）'
    return
  }
  if (!draft.model.trim() || !draft.tier.trim() || !draft.pricingVersion.trim()) {
    error.value = '请选择模型，并填写具体计费档位和价格版本'
    return
  }
  const money = /^(?:0|[1-9][0-9]*)(?:\.[0-9]{1,8})?$/
  if (!money.test(draft.salePrice) || !money.test(draft.providerCost) ||
    Number(draft.salePrice) <= 0 || Number(draft.providerCost) < 0) {
    error.value = '请手动填写有效的美元售价和美元成本（最多 8 位小数），不能直接复制上游公开标价'
    return
  }
  if (!Number.isInteger(Number(draft.units)) || draft.units < 1 || draft.units > 16 ||
    (draft.kind === 'video_generation' && draft.units !== 1)) {
    error.value = '图片数量为 1–16 张，视频每条任务必须为 1 个输出'
    return
  }
  if (draft.enabled && !reviewed.value) {
    error.value = '启用收费前，请先勾选已核对官方币种、档位与利润的确认框'
    return
  }
  let options: Record<string, unknown>
  try {
    const parsed: unknown = JSON.parse(optionsJSON.value)
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') throw Error('not object')
    options = parsed as Record<string, unknown>
  } catch {
    error.value = '参数档位必须是一个合法 JSON 对象；图片至少需要 {"n":1}'
    return
  }
  if (draft.kind !== 'video_generation' && options.n !== draft.units) {
    error.value = '图片参数中的 n 必须与收费数量一致'
    return
  }
  const body: PriceRule = {
    id: draft.id.trim(), model: draft.model.trim(), kind: draft.kind, options,
    snapshot: {
      pricing_version: draft.pricingVersion.trim(), currency: 'USD',
      tier: draft.tier.trim(), units: Number(draft.units),
      unit_sale_price_usd: draft.salePrice,
      unit_provider_cost_usd: draft.providerCost,
      verified_by_admin: false,
    },
    enabled: draft.enabled,
  }
  saving.value = true
  try {
    await api.put('/admin/cangyuan-media/prices/' + encodeURIComponent(body.id), body)
    if (editing.value && editOriginalId.value !== body.id) {
      notice.value = '已另存为新规则，请手动检查旧规则是否仍然启用'
    } else {
      notice.value = body.enabled ? '已保存并启用价格规则' : '已保存草稿价格（未启用收费）'
    }
    editorOpen.value = false
    await load()
    // load() does not overwrite the success message unless catalog is offline.
    if (catalog.value) notice.value = body.enabled ? '已保存并启用价格规则' : '已保存草稿价格（未启用收费）'
  } catch (caught) {
    error.value = '保存失败：' + getErrorMessage(caught) + '。如启用了 Step-up 2FA，请先完成账户安全验证。'
  } finally {
    saving.value = false
  }
}

async function removeRule(rule: PriceRule) {
  if (deleting.value || !window.confirm('确定删除价格规则“' + rule.id + '”？已提交的任务仍按冻结价格结算。')) return
  deleting.value = rule.id
  error.value = ''
  try {
    await api.delete('/admin/cangyuan-media/prices/' + encodeURIComponent(rule.id))
    await load()
    notice.value = '价格规则已删除'
  } catch (caught) {
    error.value = '删除失败：' + getErrorMessage(caught) + '。如启用了 Step-up 2FA，请先完成账户安全验证。'
  } finally {
    deleting.value = ''
  }
}

function close() {
  if (!saving.value && !deleting.value) emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="cy-pricing-overlay" role="presentation">
      <section class="cy-pricing-panel" role="dialog" aria-modal="true" aria-label="沧元算力媒体价格管理">
        <header class="cy-pricing-head">
          <div>
            <span class="cy-pricing-kicker">CANGYUAN · MEDIA PRICING</span>
            <h2>媒体价格与模型目录</h2>
            <p>先核对官方档位和币种，再由管理员人工审核美元售价。没有审核的规则不会对用户开放。</p>
          </div>
          <button type="button" class="cy-pricing-close" aria-label="关闭价格管理" @click="close">×</button>
        </header>
        <div class="cy-pricing-stats">
          <span>已保存 <strong>{{ rules.length }}</strong> 条</span>
          <span>已启用 <strong>{{ activeCount }}</strong> 条</span>
          <span>官方目录 <strong>{{ catalog?.items.length ?? '—' }}</strong> 个媒体型号</span>
        </div>
        <div class="cy-pricing-alert">
          <strong>计费安全</strong>
          <span>官方 <code>model_price</code> 仅供参考，可能以站点展示币种计价，不是美元成本。分档模型的基础数字也未必是最终费用。禁止自动导入为 USD 收费价。</span>
        </div>
        <p v-if="error" class="cy-pricing-error" role="alert">{{ error }}</p>
        <p v-if="notice" class="cy-pricing-notice" role="status">{{ notice }}</p>
        <div class="cy-pricing-actions">
          <button type="button" class="cy-pricing-primary" :disabled="loading" @click="resetDraft">新增价格规则</button>
          <button type="button" :disabled="loading" @click="load">{{ loading ? '读取中…' : '刷新模型与价格' }}</button>
          <a href="https://ai.cangyuansuanli.cn/console/pricing" target="_blank" rel="noopener noreferrer">核对官方价格</a>
        </div>
        <div v-if="editorOpen" class="cy-pricing-editor">
          <div class="cy-pricing-editor-head">
            <strong>{{ editing ? '编辑审核规则' : '创建审核规则' }}</strong>
            <button type="button" @click="editorOpen = false">收起</button>
          </div>
          <div class="cy-pricing-form">
            <label class="cy-pricing-field">类型
              <select v-model="draft.kind" :disabled="editing" @change="changeKind">
                <option value="image_generation">图片生成</option>
                <option value="image_edit">图片编辑</option>
                <option value="video_generation">视频生成</option>
              </select>
            </label>
            <label class="cy-pricing-field">检索官方模型
              <input v-model="query" type="search" placeholder="搜索模型 ID" />
            </label>
            <div v-if="!editing && catalog?.items.length" class="cy-pricing-models">
              <button v-for="item in filteredCatalog" :key="item.kind + ':' + item.id" type="button" :class="{ picked: draft.model === item.id }" @click="chooseModel(item)">
                <span>{{ item.id }}</span>
                <small>{{ item.tiered ? '分档' : '按次' }} · 参考标价 {{ item.model_price }} {{ item.request_unit }}</small>
              </button>
            </div>
            <label class="cy-pricing-field">精确模型 ID
              <input v-model="draft.model" :readonly="editing" type="text" placeholder="选择或输入上游模型 ID" />
            </label>
            <label class="cy-pricing-field">规则 ID
              <input v-model="draft.id" :readonly="editing" type="text" placeholder="image-1k-v1" />
            </label>
            <label class="cy-pricing-field">版本号
              <input v-model="draft.pricingVersion" type="text" placeholder="请核对官方最新版本" />
            </label>
            <label class="cy-pricing-field">档位说明
              <input v-model="draft.tier" type="text" placeholder="例如 720p-5s / 1024-square" />
            </label>
            <label class="cy-pricing-field">生成数量
              <input v-model.number="draft.units" type="number" min="1" max="16" :readonly="draft.kind === 'video_generation'" />
            </label>
            <label class="cy-pricing-field">单个输出售价（USD）
              <input v-model.trim="draft.salePrice" type="text" inputmode="decimal" placeholder="人工核对后填写" />
            </label>
            <label class="cy-pricing-field">单个输出成本（USD）
              <input v-model.trim="draft.providerCost" type="text" inputmode="decimal" placeholder="先自行确认币种和换算" />
            </label>
            <label class="cy-pricing-field cy-pricing-wide">精确请求参数（JSON）
              <textarea v-model="optionsJSON" rows="3" spellcheck="false" placeholder='{"n":1}'></textarea>
              <small>价格必须对应唯一的真实请求档位。视频请按官方要求填写时长、画质等；不支持任意未知字段。</small>
            </label>
          </div>
          <label class="cy-pricing-check">
            <input v-model="draft.enabled" type="checkbox" />
            <span>保存后立即启用该档位的用户收费</span>
          </label>
          <label v-if="draft.enabled" class="cy-pricing-check cy-pricing-reviewed">
            <input v-model="reviewed" type="checkbox" />
            <span>我已核对模型、选项、官方计费币种、USD 成本和售价，理解此操作会启用真实收费</span>
          </label>
          <footer class="cy-pricing-editor-footer">
            <span>{{ draft.enabled ? '启用规则：满足账号和分组授权后可收费' : '默认保存未启用的安全草稿' }}</span>
            <button type="button" class="cy-pricing-primary" :disabled="saving" @click="save">{{ saving ? '保存中…' : '保存价格规则' }}</button>
          </footer>
        </div>
        <div class="cy-pricing-list-head">
          <strong>已保存规则</strong>
          <span>独立按模型 / 档位计价，不会自动覆盖其他规则</span>
        </div>
        <div v-if="!loading && !rules.length" class="cy-pricing-empty">暂无审核价格。先建立一条停用状态的测试规则，再启用收费。</div>
        <div v-for="rule in rules" :key="rule.id" class="cy-pricing-rule">
          <div>
            <strong>{{ rule.model }}</strong>
            <small>{{ rule.id }} · {{ rule.snapshot.tier }} · {{ rule.kind }}</small>
            <small>USD {{ rule.snapshot.unit_sale_price_usd }} / 输出 · 成本 USD {{ rule.snapshot.unit_provider_cost_usd }} · {{ rule.snapshot.units }} 个输出</small>
          </div>
          <span class="cy-pricing-state" :class="{ enabled: rule.enabled }">{{ rule.enabled ? '已启用' : '未启用' }}</span>
          <button type="button" @click="editRule(rule)">编辑</button>
          <button type="button" class="cy-pricing-remove" :disabled="deleting === rule.id" @click="removeRule(rule)">删除</button>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.cy-pricing-overlay { position:fixed; inset:0; z-index:1100; display:grid; place-items:center; padding:16px; background:rgba(4,8,13,.72); backdrop-filter:blur(16px); }
.cy-pricing-panel { width:min(920px,100%); max-height:min(94vh,980px); overflow:auto; border:1px solid var(--border-color,rgba(155,166,172,.24)); border-radius:20px; background:var(--bg-card,#151b22); color:var(--text-primary,#f0f3f5); padding:24px; box-shadow:0 34px 100px rgba(0,0,0,.32); }
.cy-pricing-head { display:flex; justify-content:space-between; align-items:flex-start; gap:20px; }
.cy-pricing-kicker { font-size:10px; font-weight:700; letter-spacing:.13em; color:var(--text-secondary,#9da8b0); }
.cy-pricing-head h2 { font-size:22px; margin:5px 0 8px; letter-spacing:-.025em; }
.cy-pricing-head p { font-size:12px; max-width:650px; color:var(--text-secondary,#9da8b0); line-height:1.6; margin:0; }
.cy-pricing-close { border:0; background:transparent; font-size:27px; color:inherit; padding:0 8px; cursor:pointer; }
.cy-pricing-stats { display:flex; flex-wrap:wrap; gap:10px 24px; margin:18px 0; font-size:12px; color:var(--text-secondary,#9da8b0); }
.cy-pricing-stats strong { color:var(--text-primary,#f0f3f5); font-size:16px; margin-left:5px; }
.cy-pricing-alert { display:flex; flex-wrap:wrap; gap:7px; padding:13px 15px; border:1px solid rgba(201,151,75,.35); border-radius:12px; background:rgba(181,126,48,.06); font-size:12px; line-height:1.65; }
.cy-pricing-alert strong { color:var(--text-primary,#e6e8eb); }
.cy-pricing-alert span { flex:1 1 380px; color:var(--text-secondary,#b7c2ca); }
.cy-pricing-actions { display:flex; gap:10px; flex-wrap:wrap; margin:16px 0; }
.cy-pricing-actions button,.cy-pricing-actions a,.cy-pricing-editor button,.cy-pricing-rule button { border:1px solid var(--border-color,#3b444c); background:transparent; color:inherit; border-radius:10px; padding:9px 13px; font-size:12px; text-decoration:none; cursor:pointer; }
.cy-pricing-primary { background:var(--text-primary,#e9eef0) !important; border-color:transparent !important; color:var(--bg-primary,#12181c) !important; font-weight:650; }
.cy-pricing-actions button:disabled,.cy-pricing-editor button:disabled,.cy-pricing-rule button:disabled { opacity:.5; cursor:not-allowed; }
.cy-pricing-error,.cy-pricing-notice { font-size:12px; border-radius:9px; padding:10px 12px; margin:8px 0; line-height:1.5; }
.cy-pricing-error { border:1px solid rgba(222,111,111,.4); color:#e88d8d; }
.cy-pricing-notice { border:1px solid rgba(98,175,148,.35); color:var(--text-secondary,#aad5c4); }
.cy-pricing-editor { border:1px solid var(--border-color,#3b444c); border-radius:15px; padding:18px; margin:16px 0 20px; }
.cy-pricing-editor-head,.cy-pricing-editor-footer { display:flex; justify-content:space-between; align-items:center; gap:12px; }
.cy-pricing-editor-head strong { font-size:14px; }
.cy-pricing-editor-head button { border:0; }
.cy-pricing-form { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:13px; margin:16px 0; }
.cy-pricing-field { min-width:0; display:flex; flex-direction:column; gap:6px; font-size:11px; color:var(--text-secondary,#aeb9c2); }
.cy-pricing-field input,.cy-pricing-field select,.cy-pricing-field textarea { box-sizing:border-box; width:100%; border:1px solid var(--border-color,#3b444c); background:var(--bg-primary,#10161b); border-radius:9px; color:var(--text-primary,#f0f3f5); padding:10px 11px; font-size:12px; outline-offset:2px; }
.cy-pricing-field textarea { resize:vertical; font-family:ui-monospace,monospace; }
.cy-pricing-field small { line-height:1.5; }
.cy-pricing-wide,.cy-pricing-models { grid-column:1/-1; }
.cy-pricing-models { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:7px; max-height:146px; overflow:auto; }
.cy-pricing-models button { text-align:left; display:flex; flex-direction:column; align-items:flex-start; gap:3px; overflow:hidden; }
.cy-pricing-models button span { max-width:100%; overflow:hidden; white-space:nowrap; text-overflow:ellipsis; }
.cy-pricing-models button small { font-size:10px; color:var(--text-secondary,#b6bec4); }
.cy-pricing-models button.picked { border-color:var(--text-primary,#c3d0d2); }
.cy-pricing-check { display:flex; gap:9px; align-items:flex-start; font-size:12px; margin:12px 0; line-height:1.6; cursor:pointer; }
.cy-pricing-check input { margin-top:3px; }
.cy-pricing-reviewed { padding:10px; border-radius:9px; border:1px solid rgba(201,151,75,.4); }
.cy-pricing-editor-footer span { font-size:11px; color:var(--text-secondary,#9da8b0); }
.cy-pricing-list-head { display:flex; justify-content:space-between; flex-wrap:wrap; gap:10px; margin:23px 0 12px; font-size:13px; }
.cy-pricing-list-head span,.cy-pricing-empty { font-size:11px; color:var(--text-secondary,#9da8b0); }
.cy-pricing-empty { padding:18px; border:1px dashed var(--border-color,#3b444c); border-radius:11px; }
.cy-pricing-rule { display:flex; align-items:center; gap:12px; padding:12px 0; border-top:1px solid var(--border-color,rgba(155,166,172,.2)); }
.cy-pricing-rule > div { flex:1; min-width:0; display:grid; gap:5px; }
.cy-pricing-rule strong { overflow-wrap:anywhere; font-size:13px; }
.cy-pricing-rule small { font-size:11px; color:var(--text-secondary,#a7b2b9); overflow-wrap:anywhere; }
.cy-pricing-state { font-size:10px; white-space:nowrap; border-radius:10px; padding:4px 7px; background:rgba(155,166,172,.12); }
.cy-pricing-state.enabled { color:#5ebf9b; background:rgba(71,168,132,.12); }
.cy-pricing-rule .cy-pricing-remove { color:#d88686; }
@media (max-width:640px) {
 .cy-pricing-panel { padding:16px; max-height:96vh; border-radius:14px; }
 .cy-pricing-form { grid-template-columns:1fr; }
 .cy-pricing-models { grid-template-columns:1fr; }
 .cy-pricing-rule { flex-wrap:wrap; }
 .cy-pricing-rule > div { flex:1 1 100%; }
}
</style>
