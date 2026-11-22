<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

// Separate, once-only video QA from the already completed image QA. It never
// posts from a poll, reload, mount hook or ambiguous network response.
const props = defineProps<{ apiKey: string; apiKeyId: number; priceApproved?: boolean }>()
type VideoTask = {
  id?: string
  task_id?: string
  status?: string
  video_url?: string
  message?: string
  error?: { message?: string }
}
const endpoint = '/v1/videos'
const model = 'omni-fast'
const prompt = 'A single red circle gently floats against a pure white background, flat minimal animation, no words or logos.'
const slot = 'cangyuan-video-qa-one-shot:' + props.apiKeyId
const sending = ref(false)
const polling = ref(false)
const taskID = ref('')
const status = ref('')
const error = ref('')
const videoURL = ref('')
let alive = true

function validTaskID(value: string): boolean {
  return /^cyjob_[a-zA-Z0-9_-]{1,240}$/.test(value)
}
function remember(value: string) {
  try { window.sessionStorage.setItem(slot, value) } catch { /* private browsing */ }
}
function forget() {
  try { window.sessionStorage.removeItem(slot) } catch { /* private browsing */ }
}
async function readJson(response: Response): Promise<VideoTask> {
  const data: unknown = await response.json().catch(() => ({}))
  return data && typeof data === 'object' && !Array.isArray(data) ? data as VideoTask : {}
}
onMounted(() => {
  try {
    const previous = window.sessionStorage.getItem(slot) || ''
    if (validTaskID(previous)) {
      taskID.value = previous
      status.value = 'previously_submitted'
    } else if (previous === 'submit_unknown') {
      status.value = 'submit_unknown'
      error.value = '此前提交结果不确定。禁止重复提交，请管理员先核查上游任务。'
    }
  } catch { /* private browsing */ }
})
onBeforeUnmount(() => { alive = false })

async function refreshStatus() {
  if (!validTaskID(taskID.value) || polling.value) return
  if (!props.apiKey.startsWith('sk-')) {
    error.value = '本页面无法使用该测试 Key，请重新打开 API 密钥管理页面。'
    return
  }
  polling.value = true
  try {
    const response = await fetch(endpoint + '/' + encodeURIComponent(taskID.value), {
      method: 'GET', credentials: 'omit', cache: 'no-store',
      headers: { Authorization: 'Bearer ' + props.apiKey, Accept: 'application/json' },
    })
    const data = await readJson(response)
    if (!response.ok) {
      error.value = '视频任务查询失败（HTTP ' + response.status + '），不会重新生成。'
      return
    }
    status.value = data.status || 'unknown'
    if (status.value === 'completed' && typeof data.video_url === 'string' && /^https:\/\//.test(data.video_url)) {
      videoURL.value = data.video_url
    }
    if (status.value === 'failed') error.value = '任务失败，请管理员核实冻结余额是否释放。'
    if (status.value === 'submit_unknown') error.value = '上游提交状态不确定，禁止重新提交。'
  } catch {
    error.value = '状态查询暂时失败，可以再次查询，不会重复计费。'
  } finally {
    polling.value = false
  }
}

async function startOneVideo() {
  if (sending.value || taskID.value || status.value === 'submit_unknown' || !props.priceApproved) return
  if (!props.apiKey.startsWith('sk-')) {
    error.value = '测试密钥无效，无法提交视频任务。'
    return
  }
  if (!window.confirm('这将尝试一次真实视频生成，并按已审核的视频价格收费。确认视频模型、售价、成本和测试额度后继续？')) return
  sending.value = true
  // Lock before sending any POST. Even an opaque network failure cannot
  // accidentally trigger a second payable upstream generation.
  remember('submit_unknown')
  error.value = ''
  videoURL.value = ''
  try {
    const response = await fetch(endpoint, {
      method: 'POST', credentials: 'omit', cache: 'no-store',
      headers: {
        Authorization: 'Bearer ' + props.apiKey,
        'Idempotency-Key': 'media-video-qa-' + crypto.randomUUID(),
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({ model, prompt, async: true }),
    })
    const data = await readJson(response)
    if (!response.ok) {
      // Definitive validation/auth rejections do not accept an upstream job.
      // Ambiguous errors must remain locked, exactly as in image QA.
      const definitelyRejected = [400, 401, 402, 403, 404, 413, 415, 422].includes(response.status)
      if (definitelyRejected) {
        forget()
        error.value = '请求已被明确拒绝（HTTP ' + response.status + '），未创建任务。请检查价格和测试权限。' +
          (data.error?.message ? ' ' + data.error.message.slice(0, 120) : '')
      } else {
        status.value = 'submit_unknown'
        error.value = '提交结果不确定（HTTP ' + response.status + '），禁止重新提交，请管理员核对任务与扣费。'
      }
      return
    }
    const id = data.task_id || data.id || ''
    if (!validTaskID(id)) {
      status.value = 'submit_unknown'
      error.value = '返回的任务编号不明确，禁止重新提交，请管理员核实。'
      return
    }
    taskID.value = id
    remember(id)
    status.value = data.status || 'queued'
  } catch {
    status.value = 'submit_unknown'
    error.value = '网络中断后请求结果不明，已禁止重新提交；先排查是否已扣费。'
  } finally {
    sending.value = false
  }
  // A video may take longer than an image. This loop only issues GET requests.
  if (validTaskID(taskID.value) && alive) {
    for (let i = 0; i < 60 && alive; i++) {
      await new Promise(resolve => setTimeout(resolve, 5000))
      if (!alive || status.value === 'completed' || status.value === 'failed' || status.value === 'submit_unknown') break
      await refreshStatus()
    }
  }
}
</script>

<template>
  <section class="cy-video-qa">
    <div class="cy-video-head">
      <div>
        <span class="cy-video-kicker">CANGYUAN · CONTROLLED VIDEO QA</span>
        <h3>沧元视频生成联调</h3>
        <p>仅测试一个 <code>omni-fast</code> 视频任务。密钥不另存，生成操作最多提交一次。</p>
      </div>
      <span class="cy-video-mark">1 / 1</span>
    </div>
    <div class="cy-video-note">
      视频测试尚未启动。必须先确认上游的实际币种、审核美元售价与成本，启用本模型的独立定价规则，并确保测试 Key 额度足够；否则不要点击。
    </div>
    <p v-if="!priceApproved" class="cy-video-pending">当前没有与此请求精确匹配的已启用视频报价，已锁定生成按钮。审核定价后刷新本页。</p>
    <div class="cy-video-actions">
      <button type="button" class="cy-video-primary" :disabled="!priceApproved || sending || Boolean(taskID) || status === 'submit_unknown'" @click="startOneVideo">
        {{ sending ? '正在提交…' : taskID ? '已提交一次，不可重复' : status === 'submit_unknown' ? '结果不明，禁止重试' : !priceApproved ? '待管理员审核视频价格' : '生成一个测试视频' }}
      </button>
      <button v-if="taskID" type="button" :disabled="polling" @click="refreshStatus">
        {{ polling ? '查询中…' : '查询任务状态' }}
      </button>
    </div>
    <p v-if="status" class="cy-video-status" role="status">任务状态：<strong>{{ status }}</strong><span v-if="taskID"> · {{ taskID }}</span></p>
    <p v-if="error" class="cy-video-error" role="alert">{{ error }}</p>
    <a v-if="videoURL" :href="videoURL" target="_blank" rel="noopener noreferrer" class="cy-video-result">查看已完成的测试视频</a>
  </section>
</template>

<style scoped>
.cy-video-qa { margin:20px 0; padding:21px; background:var(--bg-card,rgba(140,150,160,.05)); color:var(--text-primary,inherit); border:1px solid var(--border-color,rgba(140,150,160,.22)); border-radius:17px; }
.cy-video-head { display:flex; justify-content:space-between; align-items:flex-start; gap:15px; }
.cy-video-kicker { font-size:10px; letter-spacing:.15em; font-weight:700; color:var(--text-secondary,#8d999f); }
.cy-video-qa h3 { font-size:16px; margin:6px 0; font-weight:700; }
.cy-video-qa p { font-size:12px; line-height:1.7; margin:0; color:var(--text-secondary,#929da5); }
.cy-video-qa code { font-family:ui-monospace,monospace; font-size:11px; }
.cy-video-mark { border:1px solid var(--border-color,rgba(140,150,160,.22)); border-radius:15px; white-space:nowrap; padding:5px 10px; font-size:12px; font-weight:700; }
.cy-video-note { font-size:12px; line-height:1.65; background:rgba(160,135,75,.07); padding:11px 13px; border:1px solid rgba(160,135,75,.18); border-radius:11px; margin:15px 0; }
.cy-video-actions { display:flex; gap:9px; flex-wrap:wrap; }
.cy-video-actions button { cursor:pointer; padding:9px 15px; border:1px solid var(--border-color,rgba(140,150,160,.24)); border-radius:11px; font-size:12px; background:transparent; color:inherit; font-weight:600; }
.cy-video-actions button:disabled { opacity:.53; cursor:not-allowed; }
.cy-video-actions .cy-video-primary { background:var(--text-primary,#e1e8ee); color:var(--bg-primary,#101820); border-color:transparent; }
.cy-video-qa .cy-video-status { margin-top:13px; overflow-wrap:anywhere; }
.cy-video-qa .cy-video-status strong { color:var(--text-primary,inherit); }
.cy-video-qa .cy-video-error { color:#d57373; margin-top:9px; }
.cy-video-result { display:inline-block; margin-top:11px; font-size:12px; color:inherit; text-decoration:underline; }
@media (max-width:600px) { .cy-video-qa { padding:16px; } .cy-video-head { flex-direction:column; gap:8px; } .cy-video-actions { flex-direction:column; } .cy-video-actions button { width:100%; } }
</style>
