<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api, getErrorMessage } from '../core/api'

type Account = { id: number; name?: string; notes?: string | null; platform?: string; type?: string; concurrency?: number; priority?: number; rate_multiplier?: number; group_ids?: number[]; credentials?: Record<string, unknown> }
type Group = { id: number; name?: string; platform?: string; status?: string }

const props = defineProps<{ show: boolean; account: Account | null }>()
const emit = defineEmits<{ close: []; updated: [] }>()
const name = ref('')
const notes = ref('')
const concurrency = ref(1)
const priority = ref(0)
const rateMultiplier = ref(1)
const baseUrl = ref('')
const apiKey = ref('')
const groupIds = ref<number[]>([])
const groups = ref<Group[]>([])
const saving = ref(false)
const error = ref('')
const canSave = computed(() => Boolean(props.account && name.value.trim() && !saving.value))
const protocolLabel = computed(() => platformLabel(props.account?.platform))

function platformLabel(value?: string) {
  const labels: Record<string, string> = {
    openai: 'OpenAI Compatible',
    anthropic: 'Anthropic',
    gemini: 'Gemini',
    antigravity: 'Antigravity',
    grok: 'xAI / Grok',
    xai: 'xAI / Grok',
    ollama: 'Ollama',
    composite: 'Composite',
  }
  const key = String(value || '').toLowerCase()
  return labels[key] || value || '未指定'
}

function toggleGroup(id: number) { groupIds.value = groupIds.value.includes(id) ? groupIds.value.filter((value) => value !== id) : [...groupIds.value, id] }
function close() { if (!saving.value) emit('close') }

async function load() {
  const account = props.account
  if (!account) return
  name.value = account.name || ''
  notes.value = account.notes || ''
  concurrency.value = Math.max(1, Number(account.concurrency || 1))
  priority.value = Math.max(0, Number(account.priority || 0))
  rateMultiplier.value = Math.max(0, Number(account.rate_multiplier ?? 1))
  baseUrl.value = String(account.credentials?.base_url || '')
  apiKey.value = ''
  groupIds.value = [...(account.group_ids || [])]
  error.value = ''
  try {
    const response = await api.get<Group[]>('/admin/groups/all', { params: { platform: account.platform || undefined } })
    groups.value = Array.isArray(response.data) ? response.data : []
  } catch (caught) { groups.value = []; error.value = getErrorMessage(caught) }
}

async function submit() {
  const account = props.account
  if (!account || !name.value.trim()) return
  saving.value = true
  error.value = ''
  try {
    const payload: Record<string, unknown> = {
      name: name.value.trim(), notes: notes.value.trim(),
      concurrency: Math.max(1, Math.round(Number(concurrency.value) || 1)),
      priority: Math.max(0, Math.round(Number(priority.value) || 0)),
      rate_multiplier: Math.max(0, Number(rateMultiplier.value) || 0), group_ids: groupIds.value,
    }
    if (account.type === 'apikey') {
      const credentials = { ...(account.credentials || {}) }
      if (baseUrl.value.trim()) credentials.base_url = baseUrl.value.trim()
      if (apiKey.value.trim()) credentials.api_key = apiKey.value.trim()
      payload.credentials = credentials
    }
    await api.put(`/admin/accounts/${account.id}`, payload)
    emit('updated'); emit('close')
  } catch (caught) { error.value = getErrorMessage(caught) } finally { saving.value = false }
}

watch(() => props.show, (visible) => { if (visible) void load() })
</script>

<template>
  <Teleport to="body"><div v-if="show && account" class="backdrop" @click.self="close"><section class="dialog" role="dialog" aria-modal="true" aria-labelledby="edit-account-title">
    <header><div><span>UPSTREAM ACCOUNT</span><h2 id="edit-account-title">编辑上游账户</h2><p>凭据不会在此显示或被覆盖。</p></div><button type="button" aria-label="关闭" @click="close">×</button></header>
    <main>
      <label>账户名称<input v-model="name" maxlength="80" /></label>
      <div class="protocol-card">
        <div><span>上游接口协议</span><strong>{{ protocolLabel }}</strong></div>
        <p>按实际 API 请求格式区分，不等于模型厂商。Claude / Gemini 等模型也可以通过 OpenAI Compatible 上游提供。</p>
      </div>
      <label>备注<textarea v-model="notes" rows="2" maxlength="300" /></label>
      <template v-if="account.type === 'apikey'">
        <label>上游 Base URL<input v-model="baseUrl" placeholder="https://api.example.com" /></label>
        <label>替换 API Key<input v-model="apiKey" type="password" autocomplete="new-password" placeholder="留空则保留原 API Key" /></label>
      </template>
      <div class="numbers"><label>最大并发<input v-model.number="concurrency" type="number" min="1" max="10000" /></label><label>优先级<input v-model.number="priority" type="number" min="0" max="9999" /></label><label>倍率<input v-model.number="rateMultiplier" type="number" min="0" step="0.01" /></label></div>
      <section class="group-section"><div class="section-title"><strong>调度分组</strong><small>仅显示与当前接口协议一致的分组</small></div><div v-if="groups.length" class="groups"><button v-for="group in groups" :key="group.id" type="button" :class="{ selected: groupIds.includes(group.id) }" @click="toggleGroup(group.id)">{{ groupIds.includes(group.id) ? '✓ ' : '' }}{{ group.name || `Group #${group.id}` }}</button></div><p v-else>没有可选择的同协议分组。</p></section>
      <p v-if="error" class="error">{{ error }}</p>
    </main>
    <footer><button type="button" @click="close">取消</button><button type="button" :disabled="!canSave" @click="submit">{{ saving ? '保存中' : '保存修改' }}</button></footer>
  </section></div></Teleport>
</template>

<style scoped>
.backdrop{position:fixed;inset:0;z-index:1400;display:grid;place-items:center;padding:24px;background:#050609cc}.dialog{width:min(620px,100%);border:1px solid #30343d;border-radius:10px;background:#121419;color:#f4f6f8;box-shadow:0 24px 80px #0008}.dialog header,.dialog footer{display:flex;justify-content:space-between;align-items:center;padding:20px 24px;border-bottom:1px solid #282b32}.dialog footer{border-top:1px solid #282b32;border-bottom:0;justify-content:flex-end;gap:10px}.dialog header span{font-size:11px;color:#8d96a2}.dialog h2{margin:5px 0;font-size:20px}.dialog p{margin:0;color:#929aa5;font-size:13px}.dialog main{padding:20px 24px;display:grid;gap:16px}.dialog label{display:grid;gap:7px;color:#cbd1d8;font-size:13px}.dialog input,.dialog textarea{box-sizing:border-box;width:100%;border:1px solid #353a44;border-radius:6px;background:#0d0f13;color:#f4f6f8;padding:9px}.protocol-card{display:flex;align-items:flex-start;justify-content:space-between;gap:18px;padding:12px 14px;border:1px solid #303741;border-radius:8px;background:#0e1116}.protocol-card>div{display:flex;flex-direction:column;gap:5px;flex:0 0 auto}.protocol-card span{color:#7d8793;font-size:11px}.protocol-card strong{color:#dfe6ec;font-size:13px;font-weight:680}.protocol-card p{max-width:330px;color:#727d88;font-size:11px;line-height:1.55;text-align:right}.numbers{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}.section-title{display:flex;align-items:baseline;justify-content:space-between;gap:12px}.section-title small{color:#6f7984;font-size:11px}.groups{display:flex;flex-wrap:wrap;gap:8px;margin-top:8px}.dialog button{border:1px solid #3a404b;border-radius:6px;background:#1a1d23;color:#e4e8ec;padding:8px 12px;cursor:pointer}.groups button.selected,.dialog footer button:last-child{border-color:#52ce9e;background:#1a493a;color:#ddfff1}.error{color:#f08b91!important}@media(max-width:560px){.numbers{grid-template-columns:1fr}.dialog{max-height:calc(100vh - 30px);overflow:auto}.protocol-card{flex-direction:column}.protocol-card p{text-align:left}.section-title{align-items:flex-start;flex-direction:column;gap:4px}}
</style>