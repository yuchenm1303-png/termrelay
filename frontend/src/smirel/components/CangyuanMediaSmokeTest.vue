<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps<{ apiKey: string; apiKeyId: number; canSubmit?: boolean }>()
type MediaTask = {
  id?: string
  task_id?: string
  status?: string
  data?: Array<{ url?: string }>
  message?: string
  error?: { message?: string }
}
const sending = ref(false)
const polling = ref(false)
const taskID = ref('')
const status = ref('')
const error = ref('')
const resultURL = ref('')
let alive = true
const savedTaskSlot = 'cangyuan-media-qa-one-shot:' + props.apiKeyId
function remember(value: string) {
  try { window.sessionStorage.setItem(savedTaskSlot, value) } catch { /* private browsing may disable storage */ }
}
function forget() {
  try { window.sessionStorage.removeItem(savedTaskSlot) } catch { /* no-op */ }
}
onMounted(() => {
  try {
    const previous = window.sessionStorage.getItem(savedTaskSlot) || ''
    if (validTaskID(previous)) {
      taskID.value = previous
      status.value = 'previously_submitted'
    } else if (previous === 'submit_unknown') {
      status.value = 'submit_unknown'
      error.value = '此前提交状态不确定。为防止重复扣费，已禁止再次提交；请管理员核对任务记录。'
    }
  } catch { /* no-op */ }
})
onBeforeUnmount(() => { alive = false })

const target = '/v1/images/generations'
const prompt = 'A minimalist flat icon: one small red circle centered on a pure white background, without words or other elements.'

function validTaskID(value: string): boolean {
  return /^cyjob_[a-zA-Z0-9_-]{1,240}$/.test(value)
}

async function readJson(response: Response): Promise<MediaTask> {
  const body: unknown = await response.json().catch(() => ({}))
  return body && typeof body === 'object' && !Array.isArray(body) ? body as MediaTask : {}
}

async function refreshStatus() {
  if (!validTaskID(taskID.value) || polling.value) return
  if (!props.apiKey.startsWith('sk-')) {
    error.value = '无法在当前页面取得该 Key，请重新打开 API 密钥页面。'
    return
  }
  polling.value = true
  try {
    const response = await fetch(target + '/' + encodeURIComponent(taskID.value), {
      method: 'GET',
      credentials: 'omit',
      cache: 'no-store',
      headers: { Authorization: 'Bearer ' + props.apiKey, Accept: 'application/json' },
    })
    const data = await readJson(response)
    if (!response.ok) {
      error.value = '任务查询失败（HTTP ' + response.status + '），不会重新提交生成请求。'
      return
    }
    status.value = data.status || 'unknown'
    if (status.value === 'completed') {
      const url = data.data?.find(item => typeof item?.url === 'string' && /^https:\/\//.test(item.url))?.url
      if (url) resultURL.value = url
    }
    if (status.value === 'failed') error.value = '上游任务失败。请检查余额冻结是否已自动退回。'
  } catch {
    error.value = '查询暂时失败，可以再次点击「查询任务状态」，不会重复扣费。'
  } finally {
    polling.value = false
  }
}

async function startSingleImage() {
  if (sending.value || taskID.value || props.canSubmit === false) return
  if (!props.apiKey.startsWith('sk-')) {
    error.value = '当前密钥不可用，不能发起测试。'
    return
  }
  if (!window.confirm('这会发起且仅发起一次实际图片生成，按测试分组的已审核价格计费。确认继续？')) return
  sending.value = true
  // Persist the uncertain state BEFORE the POST so reload cannot re-submit it.
  remember('submit_unknown')
  error.value = ''
  resultURL.value = ''
  try {
    const idempotencyKey = 'media-qa-' + crypto.randomUUID()
    const response = await fetch(target, {
      method: 'POST',
      credentials: 'omit',
      cache: 'no-store',
      headers: {
        Authorization: 'Bearer ' + props.apiKey,
        'Idempotency-Key': idempotencyKey,
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        model: 'gpt-image-2.6', prompt, n: 1, async: true, response_format: 'url',
      }),
    })
    const data = await readJson(response)
    if (!response.ok) {
      // A 5xx/timeout can be ambiguous: never create a second paid upstream
      // task even when the browser did receive an HTTP response.
      const definitelyRejected = [400, 401, 402, 403, 404, 413, 415, 422].includes(response.status)
      if (definitelyRejected) {
        forget()
        error.value = '请求被拒绝（HTTP ' + response.status + '）：' +
          (data.error?.message || '检查测试通道是否开启及价格规则是否已审核。').slice(0, 180)
      } else {
        status.value = 'submit_unknown'
        error.value = '请求状态不确定（HTTP ' + response.status + '），为避免重复扣费已禁止重新提交。请管理员检查任务记录。'
      }
      return
    }
    const id = data.task_id || data.id || ''
    if (!validTaskID(id)) {
      error.value = '已提交但任务编号不可识别。为避免重复计费，请勿再次点击生成，联系管理员检查任务记录。'
      status.value = 'submit_unknown'
      return
    }
    taskID.value = id
    remember(id)
    status.value = data.status || 'queued'
  } catch {
    // A dropped response can mean the upstream already accepted the job.
    // Never reissue this POST automatically or allow a second click.
    status.value = 'submit_unknown'
    error.value = '网络中断或结果不确定，为避免重复扣费已锁定重复提交。请管理员按时间检查任务记录。'
  } finally {
    sending.value = false
  }
  // Only GET polls may be repeated, never the paid POST.
  if (validTaskID(taskID.value) && alive) {
    for (let i = 0; i < 20 && alive; i++) {
      await new Promise(resolve => setTimeout(resolve, 4500))
      if (!alive || status.value === 'completed' || status.value === 'failed') break
      await refreshStatus()
      if (status.value === 'completed' || status.value === 'failed') break
    }
  }
}
</script>

<template>
  <section class="cy-test">
    <div class="cy-test-head">
      <div>
        <span class="cy-test-kicker">CANGYUAN · CONTROLLED QA</span>
        <h3>沧元图片生成联调</h3>
        <p>仅测试一张 <code>gpt-image-2.6</code> 图片。密钥只在当前浏览器内存中使用，不另行保存。</p>
      </div>
      <span class="cy-test-mark">1 / 1</span>
    </div>
    <div class="cy-test-note">
      真实请求可能收费。先由管理员确认测试账号、审核价格及余额冻结已就绪，再点击；提交后只查询状态，不会重复生成。
    </div>
    <div class="cy-test-actions">
      <button type="button" class="cy-test-primary" :disabled="canSubmit === false || sending || Boolean(taskID) || Boolean(status === 'submit_unknown')" @click="startSingleImage">
        {{ sending ? '正在提交…' : taskID ? '已提交一次，不可重复' : status === 'submit_unknown' ? '请求不确定，禁止重试' : canSubmit === false ? '额度已用尽，仅可查询' : '生成一张测试图片' }}
      </button>
      <button v-if="taskID" type="button" :disabled="polling" @click="refreshStatus">
        {{ polling ? '查询中…' : '查询任务状态' }}
      </button>
    </div>
    <p v-if="status" class="cy-test-status" role="status">任务状态：<strong>{{ status }}</strong><span v-if="taskID"> · {{ taskID }}</span></p>
    <p v-if="error" class="cy-test-error" role="alert">{{ error }}</p>
    <a v-if="resultURL" :href="resultURL" target="_blank" rel="noopener noreferrer" class="cy-test-result">查看已完成的测试图片</a>
  </section>
</template>

<style scoped>
.cy-test { margin:20px 0; padding:21px; background:var(--bg-card,rgba(140,150,160,.05)); color:var(--text-primary,inherit); border:1px solid var(--border-color,rgba(140,150,160,.22)); border-radius:17px; }
.cy-test-head { display:flex; justify-content:space-between; align-items:flex-start; gap:15px; }
.cy-test-kicker { font-size:10px; letter-spacing:.15em; font-weight:700; color:var(--text-secondary,#8d999f); }
.cy-test h3 { font-size:16px; margin:6px 0; font-weight:700; }
.cy-test p { font-size:12px; line-height:1.7; margin:0; color:var(--text-secondary,#929da5); }
.cy-test code { font-family:ui-monospace,monospace; font-size:11px; }
.cy-test-mark { border:1px solid var(--border-color,rgba(140,150,160,.22)); border-radius:15px; white-space:nowrap; padding:5px 10px; font-size:12px; font-weight:700; }
.cy-test-note { font-size:12px; line-height:1.65; background:rgba(160,135,75,.07); padding:11px 13px; border:1px solid rgba(160,135,75,.18); border-radius:11px; margin:15px 0; }
.cy-test-actions { display:flex; gap:9px; flex-wrap:wrap; }
.cy-test-actions button { cursor:pointer; padding:9px 15px; border:1px solid var(--border-color,rgba(140,150,160,.24)); border-radius:11px; font-size:12px; background:transparent; color:inherit; font-weight:600; }
.cy-test-actions button:disabled { opacity:.53; cursor:not-allowed; }
.cy-test-actions .cy-test-primary { background:var(--text-primary,#e1e8ee); color:var(--bg-primary,#101820); border-color:transparent; }
.cy-test .cy-test-status { margin-top:13px; overflow-wrap:anywhere; }
.cy-test .cy-test-status strong { color:var(--text-primary,inherit); }
.cy-test .cy-test-error { color:#d57373; margin-top:9px; }
.cy-test-result { display:inline-block; margin-top:11px; font-size:12px; color:inherit; text-decoration:underline; }
@media (max-width:600px) { .cy-test { padding:16px; } .cy-test-head { flex-direction:column; gap:8px; } .cy-test-actions { flex-direction:column; } .cy-test-actions button { width:100%; } }
</style>
