<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { api, getErrorMessage, previewMode } from '../core/api'
import { useSession } from '../core/session'
import {
  blankProject, canConnect, creativeId, importCreativeProject, loadCreativeShelf,
  makeNode, normalizeProject, safeMediaUrl, saveCreativeShelf,
  type CanvasNodeKind, type CanvasTaskStatus, type CreativeNode, type CreativeProject, type CreativeShelf,
} from '../core/creative-canvas'
import { readCreativeAsset, storeCreativeAsset } from '../core/creative-assets'
import {
  listCloudProjects, getCloudProject, putCloudProject, deleteCloudProject, getCreativeMediaQuote,
} from '../core/creative-cloud'
import '../styles/creative-canvas.css'

type KeyOption = {
  id: number
  name?: string
  key?: string
  status?: string
  group_id?: number | null
  group?: { id: number; name: string; platform?: string }
}
type TaskResponse = {
  id?: string
  task_id?: string
  status?: CanvasTaskStatus
  video_url?: string
  data?: Array<{ url?: string }>
  error?: { message?: string }
}
type PointerDrag =
  | { mode: 'pan'; pointerId: number; clientX: number; clientY: number; panX: number; panY: number }
  | { mode: 'node'; pointerId: number; clientX: number; clientY: number; x: number; y: number; nodeId: string; moved: boolean }

const { state } = useSession()
const userId = computed(() => state.user?.id ?? 0)
// Keep the owner bound to its shelf even during logout/fast account changes.
// Otherwise an unmount could write the previous user's draft under user 0.
let shelfOwnerId = userId.value
const shelf = ref<CreativeShelf>(loadCreativeShelf(shelfOwnerId))
const project = computed(() => shelf.value.projects.find(p => p.id === shelf.value.activeId) || shelf.value.projects[0])
const selectedId = ref('')
const selectedNode = computed(() => project.value?.nodes.find(node => node.id === selectedId.value) || null)
const connectionFrom = ref('')
const viewport = ref<HTMLElement | null>(null)
const assetInput = ref<HTMLInputElement | null>(null)
const importInput = ref<HTMLInputElement | null>(null)
const assetTarget = ref('')
const assetUrls = reactive<Record<string, string>>({})
const zoom = ref(0.86)
const panX = ref(65)
const panY = ref(92)
const notice = ref('画布将在本设备自动保存。')
const error = ref('')
const saveState = ref('已保存到本机')
const cloudState = ref('正在连接云端…')
const cloudReady = ref(false)
const cloudConflict = ref(false)
const cloudVersions = new Map<string, number>()
const cloudHashes = new Map<string, string>()
let cloudEpoch = 0
let cloudSaveChain: Promise<void> = Promise.resolve()
const apiKeys = ref<KeyOption[]>([])
const selectedKeyId = ref('')
const keyLoading = ref(false)
const runningNodes = reactive<Record<string, boolean>>({})
const pollingNodes = reactive<Record<string, boolean>>({})
const undoStack = ref<string[]>([])
const redoStack = ref<string[]>([])
let drag: PointerDrag | null = null
const touchPoints = new Map<number, { x: number; y: number }>()
let pinch: { initialDistance: number; initialZoom: number; worldX: number; worldY: number } | null = null
let saveHandle: number | undefined
let destroyed = false
let hydrationVersion = 0
const pollingTimers = new Set<number>()
const maxProjects = 20

const tools: Array<{ kind: CanvasNodeKind; icon: string; label: string; hint: string }> = [
  { kind: 'prompt', icon: 'Aa', label: '提示词', hint: '输入内容或思路' },
  { kind: 'image', icon: '▧', label: '图片生成', hint: '调用图像模型' },
  { kind: 'video', icon: '▷', label: '视频生成', hint: '调用视频模型' },
  { kind: 'note', icon: '✎', label: '创作便签', hint: '记录灵感' },
]
const kindNames: Record<CanvasNodeKind, string> = {
  prompt: '提示词', image: '图片生成', video: '视频生成', note: '便签',
}
const statusNames: Record<CanvasTaskStatus, string> = {
  idle: '未运行', reserved: '已预留', queued: '已排队', in_progress: '生成中', submitting: '提交中',
  submit_unknown: '提交状态未知', completed: '已完成', failed: '失败',
}
const nodeWidth = 294
const nodePortY = 104
const stageStyle = computed(() => ({
  transform: `translate3d(${panX.value}px,${panY.value}px,0) scale(${zoom.value})`,
}))
// Tile the backdrop in screen coordinates rather than clipping it to the
// finite SVG layer; this keeps the dotted canvas visually infinite.
const gridStyle = computed(() => ({
  backgroundSize: `${23 * zoom.value}px ${23 * zoom.value}px`,
  backgroundPosition: `${panX.value}px ${panY.value}px`,
}))
const edges = computed(() => {
  const map = new Map(project.value?.nodes.map(node => [node.id, node]) || [])
  return (project.value?.links || []).flatMap(link => {
    const a = map.get(link.from)
    const b = map.get(link.to)
    if (!a || !b) return []
    const x1 = a.x + nodeWidth
    const y1 = a.y + nodePortY
    const x2 = b.x
    const y2 = b.y + nodePortY
    const dx = Math.max(65, Math.abs(x2 - x1) * 0.48)
    return [{ id: link.id, from: link.from, to: link.to, d: `M ${x1} ${y1} C ${x1 + dx} ${y1}, ${x2 - dx} ${y2}, ${x2} ${y2}` }]
  })
})
const activeKey = computed(() => apiKeys.value.find(k => String(k.id) === selectedKeyId.value) || null)
// Preview 01: keep paid generation out of the general public UI until
// authenticated exact USD price quotes are implemented.
const isRunReady = computed(() => Boolean(state.user?.role === 'admin' &&
  activeKey.value?.key && activeKey.value.status === 'active' && !previewMode))
const canUndo = computed(() => undoStack.value.length > 0)
const canRedo = computed(() => redoStack.value.length > 0)

function feedback(text: string) { notice.value = text; error.value = '' }
function problem(text: string) { error.value = text }
function snapshot() {
  if (!project.value) return
  touchProject()
  undoStack.value.push(JSON.stringify({ nodes: project.value.nodes, links: project.value.links }))
  if (undoStack.value.length > 35) undoStack.value.shift()
  redoStack.value = []
}
function history(action: 'undo' | 'redo') {
  if (!project.value) return
  const source = action === 'undo' ? undoStack.value : redoStack.value
  const target = action === 'undo' ? redoStack.value : undoStack.value
  const previous = source.pop()
  if (!previous) return
  target.push(JSON.stringify({ nodes: project.value.nodes, links: project.value.links }))
  const state = JSON.parse(previous) as Pick<CreativeProject, 'nodes' | 'links'>
  project.value.nodes = state.nodes
  project.value.links = state.links
  if (!project.value.nodes.some(node => node.id === selectedId.value)) selectedId.value = ''
  void hydrateAssets()
}
function touchProject() {
  if (project.value) project.value.updatedAt = new Date().toISOString()
}
function saveNow() {
  if (!project.value) return
  try {
    saveCreativeShelf(shelfOwnerId, shelf.value)
    saveState.value = '已保存到本机'
    if (cloudReady.value) queueCloudSave()
  } catch (err) {
    saveState.value = '保存失败'
    problem(getErrorMessage(err))
  }
}

// Local persistence is always the fallback. Cloud writes are serialized and
// version-gated, so an outdated device cannot silently overwrite revisions.
function queueCloudSave() {
  if (!cloudReady.value || !project.value || previewMode) return
  const owner = shelfOwnerId
  const epoch = cloudEpoch
  const id = project.value.id
  const content = JSON.stringify(project.value)
  if (cloudHashes.get(id) === content) return
  cloudSaveChain = cloudSaveChain.catch(() => {}).then(async () => {
    if (destroyed || !cloudReady.value || cloudEpoch !== epoch || shelfOwnerId !== owner) return
    const current = shelf.value.projects.find(p => p.id === id)
    if (!current) return
    const latest = JSON.stringify(current)
    if (cloudHashes.get(id) === latest) return
    cloudState.value = '正在同步云端…'
    try {
      const version = cloudVersions.get(id) || 0
      const saved = await putCloudProject(JSON.parse(latest) as CreativeProject, version)
      if (destroyed || cloudEpoch !== epoch || shelfOwnerId !== owner) return
      cloudVersions.set(id, saved.version)
      cloudHashes.set(id, latest)
      cloudState.value = '已同步云端'
      if (JSON.stringify(shelf.value.projects.find(p => p.id === id)) !== latest) {
        queueCloudSave() // A second edit arrived while the first write was pending.
      }
    } catch (caught) {
      if (destroyed || cloudEpoch !== epoch || shelfOwnerId !== owner) return
      cloudReady.value = false
      cloudConflict.value = true
      cloudState.value = '云端同步暂停 · 本机草稿安全保留'
      problem('云端保存失败或版本冲突。草稿已保留在本机，请导出备份后点击「重新连接云端」。' +
        getErrorMessage(caught).slice(0, 95))
    }
  })
}

async function connectCloud() {
  const owner = shelfOwnerId
  const epoch = ++cloudEpoch
  cloudReady.value = false
  cloudConflict.value = false
  cloudState.value = previewMode ? '静态预览 · 仅本地存储' : '正在连接云端…'
  cloudVersions.clear()
  cloudHashes.clear()
  if (previewMode || owner <= 0) return
  try {
    const metas = await listCloudProjects()
    if (destroyed || cloudEpoch !== epoch || shelfOwnerId !== owner) return
    const remote = await Promise.all(metas.map(async item => ({
      meta: item, snapshot: await getCloudProject(item.id),
    })))
    if (destroyed || cloudEpoch !== epoch || shelfOwnerId !== owner) return

    const current = shelf.value.projects
    const remoteIds = new Set(remote.map(item => item.meta.id))
    const merged: CreativeProject[] = []
    let recoveryCount = 0
    for (const item of remote) {
      const verified = normalizeProject(item.snapshot.project)
      if (!verified || verified.id !== item.meta.id || item.snapshot.version !== item.meta.version) {
        throw Error('云端项目数据不完整')
      }
      const local = current.find(p => p.id === verified.id)
      if (local && JSON.stringify(local) !== JSON.stringify(verified)) {
        // Never overwrite an unsynced local edit. Fork it as an independent
        // project rather than replaying stale versions over a remote document.
        const recovery = importCreativeProject(JSON.stringify(local))
        recovery.title = (local.title.slice(0, 72) + ' · 本机恢复').slice(0, 90)
        merged.push(recovery)
        recoveryCount += 1
      }
      merged.push(verified)
      cloudVersions.set(verified.id, item.meta.version)
      cloudHashes.set(verified.id, JSON.stringify(verified))
    }
    for (const item of current) if (!remoteIds.has(item.id)) merged.push(item)
    if (merged.length > maxProjects) {
      throw Error('云端与本机项目合计超过 20 个；未自动合并，请先导出本机备份。')
    }
    const active = shelf.value.activeId
    shelf.value = {
      projects: merged.length ? merged : [blankProject()],
      activeId: merged.some(item => item.id === active) ? active : (merged[0]?.id || ''),
    }
    if (!merged.length) shelf.value.activeId = shelf.value.projects[0].id
    cloudReady.value = true
    cloudState.value = '已连接云端'
    saveCreativeShelf(owner, shelf.value)
    if (recoveryCount) {
      feedback('检测到多设备内容差异，已将本机草稿另存为 ' + recoveryCount + ' 个恢复副本；未覆盖云端版本。')
    }
    if (!cloudVersions.has(shelf.value.activeId)) queueCloudSave()
    void hydrateAssets()
  } catch (caught) {
    if (destroyed || cloudEpoch !== epoch || shelfOwnerId !== owner) return
    cloudReady.value = false
    cloudState.value = '离线草稿 · 云端未同步'
    problem('云端连接未完成，本机草稿仍可正常编辑：' + getErrorMessage(caught))
  }
}

function resetView() {
  zoom.value = 0.86
  panX.value = 70
  panY.value = 105
}
function switchProject(id: string) {
  if (!shelf.value.projects.some(item => item.id === id)) return
  shelf.value.activeId = id
  selectedId.value = ''
  connectionFrom.value = ''
  undoStack.value = []
  redoStack.value = []
  resetView()
  void hydrateAssets()
}
function makeProject() {
  if (shelf.value.projects.length >= maxProjects) return problem('最多保留 20 个本地项目，请先导出或删除旧项目。')
  const next = blankProject()
  shelf.value.projects.unshift(next)
  switchProject(next.id)
  feedback('已创建新画布')
}
function copyProject() {
  if (!project.value || shelf.value.projects.length >= maxProjects) return
  const old = project.value
  const ids = new Map<string, string>()
  const nodes = old.nodes.map(node => {
    const id = creativeId()
    ids.set(node.id, id)
    return { ...node, id, taskId: '', taskKeyId: null, status: 'idle' as CanvasTaskStatus }
  })
  const copy = { ...blankProject(old.title + ' · 副本'), nodes,
    links: old.links.map(link => ({ id: creativeId(), from: ids.get(link.from)!, to: ids.get(link.to)! })) }
  shelf.value.projects.unshift(copy)
  switchProject(copy.id)
  feedback('已复制项目和节点，历史生成任务不会重复执行')
}
async function removeProject() {
  if (!project.value || !window.confirm('确定删除当前画布？云端与本机的该项目都会删除，操作无法撤销。')) return
  const id = project.value.id
  if (!cloudReady.value && !previewMode) return problem('云端暂未连接，避免删除后重新同步恢复；请先重新连接。')
  try {
    const version = cloudVersions.get(id)
    if (version && !previewMode) {
      await cloudSaveChain
      const currentVersion = cloudVersions.get(id)
      if (!currentVersion) throw Error('无法取得云端项目版本')
      await deleteCloudProject(id, currentVersion)
      cloudVersions.delete(id)
      cloudHashes.delete(id)
    }
    shelf.value.projects = shelf.value.projects.filter(p => p.id !== id)
    if (shelf.value.projects.length === 0) shelf.value.projects.push(blankProject())
    switchProject(shelf.value.projects[0].id)
    feedback('画布已删除')
  } catch (caught) { problem('无法删除云端项目，本机副本已保留：' + getErrorMessage(caught)) }
}
function insert(kind: CanvasNodeKind) {
  if (!project.value) return
  if (project.value.nodes.length >= 120) return problem('单个画布最多允许 120 个节点。')
  const bounds = viewport.value?.getBoundingClientRect()
  const centerX = bounds ? (bounds.width * 0.48 - panX.value) / zoom.value : 440
  const centerY = bounds ? (bounds.height * 0.48 - panY.value) / zoom.value : 300
  snapshot()
  const count = project.value.nodes.length
  const node = makeNode(kind, centerX - 147 + (count % 4) * 22, centerY - 105 + (count % 3) * 24)
  project.value.nodes.push(node)
  selectedId.value = node.id
  feedback('已添加' + kindNames[kind] + '节点')
}
function choose(id: string) { selectedId.value = id; error.value = '' }
function duplicateNode() {
  if (!project.value || !selectedNode.value || project.value.nodes.length >= 120) return
  snapshot()
  const original = selectedNode.value
  const item: CreativeNode = {
    ...original, id: creativeId(), x: original.x + 42, y: original.y + 55,
    taskId: '', taskKeyId: null, status: 'idle', resultUrl: '',
  }
  project.value.nodes.push(item)
  selectedId.value = item.id
}
function deleteNode() {
  if (!project.value || !selectedNode.value) return
  snapshot()
  const id = selectedNode.value.id
  project.value.nodes = project.value.nodes.filter(node => node.id !== id)
  project.value.links = project.value.links.filter(link => link.from !== id && link.to !== id)
  selectedId.value = ''
  connectionFrom.value = connectionFrom.value === id ? '' : connectionFrom.value
  touchProject()
}
function clearLinks() {
  if (!project.value || !selectedNode.value) return
  snapshot()
  const id = selectedNode.value.id
  project.value.links = project.value.links.filter(link => link.from !== id && link.to !== id)
}
function beginConnect(id: string) {
  connectionFrom.value = connectionFrom.value === id ? '' : id
  if (connectionFrom.value) feedback('连接中：点击目标节点左侧的圆点')
}
function endConnect(id: string) {
  if (!project.value || !connectionFrom.value) return
  if (!canConnect(project.value, connectionFrom.value, id)) return problem('无法连接：不能形成环路、重复连线或不符合节点方向。')
  snapshot()
  project.value.links.push({ id: creativeId(), from: connectionFrom.value, to: id })
  connectionFrom.value = ''
  feedback('节点已连接；运行仍需逐节点手动确认，不会自动扣费。')
}
function clearConnection() { connectionFrom.value = '' }

function touchMidpoint() {
  const [a, b] = [...touchPoints.values()]
  if (!a || !b) return null
  return { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2, distance: Math.hypot(b.x - a.x, b.y - a.y) }
}
function trackPointerDown(event: PointerEvent) {
  if (event.pointerType !== 'touch') return
  touchPoints.set(event.pointerId, { x: event.clientX, y: event.clientY })
  if (touchPoints.size < 2) return
  const midpoint = touchMidpoint()
  const bounds = viewport.value?.getBoundingClientRect()
  if (!midpoint || !bounds || midpoint.distance < 12) return
  pinch = {
    initialDistance: midpoint.distance, initialZoom: zoom.value,
    worldX: (midpoint.x - bounds.left - panX.value) / zoom.value,
    worldY: (midpoint.y - bounds.top - panY.value) / zoom.value,
  }
  drag = null
}
function trackPointerMove(event: PointerEvent) {
  if (!touchPoints.has(event.pointerId)) return
  touchPoints.set(event.pointerId, { x: event.clientX, y: event.clientY })
  if (!pinch || touchPoints.size < 2) return
  const midpoint = touchMidpoint()
  const bounds = viewport.value?.getBoundingClientRect()
  if (!midpoint || !bounds) return
  zoom.value = Math.min(1.7, Math.max(.4, pinch.initialZoom * midpoint.distance / pinch.initialDistance))
  panX.value = midpoint.x - bounds.left - pinch.worldX * zoom.value
  panY.value = midpoint.y - bounds.top - pinch.worldY * zoom.value
}
function trackPointerEnd(event: PointerEvent) {
  touchPoints.delete(event.pointerId)
  if (touchPoints.size < 2) pinch = null
}
function pointerStartPan(event: PointerEvent) {
  if (touchPoints.size > 1) return
  if (event.button !== 0 && event.pointerType === 'mouse') return
  selectedId.value = ''
  drag = { mode: 'pan', pointerId: event.pointerId, clientX: event.clientX, clientY: event.clientY, panX: panX.value, panY: panY.value }
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
}
function pointerStartNode(event: PointerEvent, node: CreativeNode) {
  if (touchPoints.size > 1) return
  if (event.button !== 0 && event.pointerType === 'mouse') return
  if ((event.target as HTMLElement).closest('button, input, textarea, a')) return
  choose(node.id)
  drag = { mode: 'node', pointerId: event.pointerId, clientX: event.clientX, clientY: event.clientY,
    x: node.x, y: node.y, nodeId: node.id, moved: false }
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
}
function pointerMove(event: PointerEvent) {
  if (pinch || !drag || drag.pointerId !== event.pointerId) return
  const dx = event.clientX - drag.clientX
  const dy = event.clientY - drag.clientY
  if (drag.mode === 'pan') {
    panX.value = drag.panX + dx
    panY.value = drag.panY + dy
  } else {
    const movement = drag
    if (!movement.moved && Math.abs(dx) + Math.abs(dy) > 4) {
      snapshot()
      movement.moved = true
    }
    if (!movement.moved) return
    const node = project.value.nodes.find(n => n.id === movement.nodeId)
    if (!node) return
    const oldX = movement.x
    const oldY = movement.y
    node.x = Math.max(-8000, Math.min(8000, Math.round((oldX + dx / zoom.value) / 8) * 8))
    node.y = Math.max(-8000, Math.min(8000, Math.round((oldY + dy / zoom.value) / 8) * 8))
  }
}
function finishPointer(event: PointerEvent) {
  if (!drag || drag.pointerId !== event.pointerId) return
  if (drag.mode === 'node' && drag.moved) touchProject()
  drag = null
}
function zoomTo(next: number, pivotX?: number, pivotY?: number) {
  const old = zoom.value
  const scale = Math.min(1.7, Math.max(0.4, Number(next.toFixed(2))))
  const bounds = viewport.value?.getBoundingClientRect()
  const cx = pivotX ?? (bounds?.width || 880) / 2
  const cy = pivotY ?? (bounds?.height || 550) / 2
  panX.value = cx - (cx - panX.value) * (scale / old)
  panY.value = cy - (cy - panY.value) * (scale / old)
  zoom.value = scale
}
function onWheel(event: WheelEvent) {
  if (event.ctrlKey || event.metaKey) {
    const box = viewport.value?.getBoundingClientRect()
    zoomTo(zoom.value * (event.deltaY > 0 ? 0.92 : 1.08),
      event.clientX - (box?.left || 0), event.clientY - (box?.top || 0))
  } else {
    panX.value -= event.deltaX
    panY.value -= event.deltaY
  }
}
function onKeyboard(event: KeyboardEvent) {
  const target = event.target as HTMLElement | null
  if (target?.closest('input,textarea,select,[contenteditable="true"]')) return
  if (event.key === 'Escape') clearConnection()
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'z') {
    event.preventDefault()
    history(event.shiftKey ? 'redo' : 'undo')
  }
  if (event.key === 'Delete' || event.key === 'Backspace') {
    if (selectedNode.value) { event.preventDefault(); deleteNode() }
  }
}
function downloadProject() {
  if (!project.value) return
  const blob = new Blob([JSON.stringify(project.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'muxway-canvas-' + new Date().toISOString().slice(0, 10) + '.json'
  link.click()
  window.setTimeout(() => URL.revokeObjectURL(url), 2000)
  feedback('项目结构已导出；本地上传的媒体文件不会包含在 JSON 内。')
}
function importPick() { importInput.value?.click() }
async function onImport(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    if (shelf.value.projects.length >= maxProjects) throw Error('本机项目数量已达上限')
    if (file.size > 500_000) throw Error('文件过大')
    const incoming = importCreativeProject(await file.text())
    shelf.value.projects.unshift(incoming)
    switchProject(incoming.id)
    feedback('画布已导入。媒体文件须在此设备重新添加；旧任务不会自动重复提交。')
  } catch (err) { problem(getErrorMessage(err)) }
  finally { input.value = '' }
}
function pickAsset() {
  if (!selectedNode.value) return
  assetTarget.value = selectedNode.value.id
  assetInput.value?.click()
}
async function onAsset(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !assetTarget.value) return
  const node = project.value.nodes.find(n => n.id === assetTarget.value)
  try {
    if (!node) throw Error('节点不存在')
    if (node.kind !== 'image' && node.kind !== 'video') throw Error('只有图片／视频节点支持本地素材')
    if (node.kind === 'image' && !file.type.startsWith('image/')) throw Error('图片节点请上传图片，不能选择视频文件')
    if (node.kind === 'video' && !file.type.startsWith('video/')) throw Error('视频节点请上传视频文件，不能选择图片文件')
    const assetId = creativeId()
    await storeCreativeAsset(shelfOwnerId, assetId, file)
    snapshot()
    node.assetId = assetId
    if (assetUrls[node.id]) URL.revokeObjectURL(assetUrls[node.id])
    assetUrls[node.id] = URL.createObjectURL(file)
    feedback('素材已保存在此浏览器。其他设备不会自动同步。')
  } catch (err) { problem(getErrorMessage(err)) }
  finally { (event.target as HTMLInputElement).value = ''; assetTarget.value = '' }
}
async function hydrateAssets() {
  const version = ++hydrationVersion
  await nextTick()
  if (destroyed || version !== hydrationVersion) return
  const ownerId = shelfOwnerId
  const projectId = project.value.id
  for (const key of Object.keys(assetUrls)) {
    URL.revokeObjectURL(assetUrls[key])
    delete assetUrls[key]
  }
  for (const node of project.value.nodes) {
    if (!node.assetId) continue
    try {
      const data = await readCreativeAsset(ownerId, node.assetId)
      // Never attach another user's local asset after a fast account switch.
      if (!destroyed && version === hydrationVersion && ownerId === shelfOwnerId &&
        project.value.id === projectId && data) {
        assetUrls[node.id] = URL.createObjectURL(data)
      }
    } catch { /* a missing local binary should not prevent editing the project */ }
  }
}
function displayMedia(node: CreativeNode): string {
  return assetUrls[node.id] || safeMediaUrl(node.resultUrl)
}
function typeOfMedia(node: CreativeNode): 'image' | 'video' {
  return node.kind === 'video' ? 'video' : 'image'
}
function inputPrompt(node: CreativeNode): string {
  const prior = project.value.links
    .map(link => link.to === node.id ? project.value.nodes.find(n => n.id === link.from) : null)
    .find((n): n is CreativeNode => Boolean(n?.kind === 'prompt'))
  return (prior?.prompt || node.prompt).trim()
}

async function loadKeys() {
  if (previewMode) return
  keyLoading.value = true
  try {
    const response = await api.get<KeyOption[] | { items?: KeyOption[] }>('/keys', { params: { page: 1, page_size: 100 } })
    const list = Array.isArray(response.data) ? response.data : (response.data.items || [])
    apiKeys.value = list.filter(key => key.status === 'active' && typeof key.key === 'string' && key.key.startsWith('sk-'))
    if (apiKeys.value.length === 1) selectedKeyId.value = String(apiKeys.value[0].id)
  } catch (err) { problem('读取 API Key 失败：' + getErrorMessage(err)) }
  finally { keyLoading.value = false }
}
function safeTaskId(value: unknown): string {
  return typeof value === 'string' && /^cyjob_[a-zA-Z0-9_-]{1,72}$/.test(value) ? value : ''
}
async function fetchTask(node: CreativeNode) {
  const key = apiKeys.value.find(item => item.id === node.taskKeyId)
  if (!key?.key) { problem('请使用创建该任务时的原始 API Key 查询任务，不能用其他密钥代替。'); return }
  if (!safeTaskId(node.taskId) || pollingNodes[node.id]) return
  pollingNodes[node.id] = true
  try {
    const path = node.kind === 'video'
      ? '/v1/videos/' + encodeURIComponent(node.taskId)
      : '/v1/images/generations/' + encodeURIComponent(node.taskId)
    const response = await fetch(path, {
      headers: { Authorization: 'Bearer ' + key.key, Accept: 'application/json' },
      method: 'GET', cache: 'no-store', credentials: 'omit',
    })
    const data = await response.json() as TaskResponse
    if (!response.ok) throw Error('任务查询失败：HTTP ' + response.status)
    if (data.status) node.status = data.status
    if (node.status === 'completed') {
      const result = node.kind === 'video' ? data.video_url : data.data?.[0]?.url
      node.resultUrl = safeMediaUrl(result)
      feedback(node.resultUrl ? '生成完成，素材已显示在画布上。' : '任务完成，但没有返回可用的 HTTPS 素材地址。')
    }
    if (node.status === 'failed') problem('任务失败；请在用量记录中检查是否退款。')
  } catch (err) { problem(getErrorMessage(err)) }
  finally { pollingNodes[node.id] = false }
}
function pollUntilDone(node: CreativeNode, rounds: number) {
  if (destroyed || rounds <= 0 || ['completed','failed','submit_unknown'].includes(node.status)) return
  const timer = window.setTimeout(async () => {
    pollingTimers.delete(timer)
    if (destroyed) return
    await fetchTask(node)
    pollUntilDone(node, rounds - 1)
  }, 6000)
  pollingTimers.add(timer)
}
async function runNode(node: CreativeNode) {
  if (state.user?.role !== 'admin') return problem('当前版本仅管理员可运行真实付费生成。普通用户可以编辑和保存画布。')
  if (!['image','video'].includes(node.kind) || runningNodes[node.id] || node.taskId ||
    node.status === 'submitting' || node.status === 'submit_unknown') return
  const key = activeKey.value
  if (!key?.key || key.status !== 'active') return problem('请先选择一把与媒体分组绑定的有效 API Key。')
  const prompt = inputPrompt(node)
  if (!prompt || prompt.length > 6000) return problem('请输入有效提示词（最多 6000 字）。')
  if (!node.model.trim()) return problem('请输入模型 ID。')
  let options: Record<string, unknown>
  try {
    const decoded: unknown = JSON.parse(node.options || '{}')
    if (!decoded || Array.isArray(decoded) || typeof decoded !== 'object') throw Error('对象')
    options = decoded as Record<string, unknown>
    if (Object.keys(options).some(item => ['model','prompt','idempotency_key'].includes(item))) throw Error('保留参数')
    if (node.kind === 'image' && options.n !== 1) throw Error('图片测试仅允许 n=1')
  } catch { return problem('生成参数必须是合法 JSON 对象；图片生成请填写 {"n":1}。') }
  if (!window.confirm('即将向真实上游提交一次' + kindNames[node.kind] +
    '请求，可能产生费用。费用由后台已审核价格决定，未审核的档位会被拒绝。确认继续？')) return
  runningNodes[node.id] = true
  snapshot()
  node.status = 'submitting'
  node.taskKeyId = key.id
  node.resultUrl = ''
  saveNow() // Persist uncertain state BEFORE the one paid POST.
  const payload: Record<string, unknown> = { ...options, model: node.model.trim(), prompt }
  if (node.kind === 'image') Object.assign(payload, { async: true, response_format: 'url' })
  try {
    const response = await fetch(node.kind === 'image' ? '/v1/images/generations' : '/v1/videos', {
      method: 'POST', cache: 'no-store', credentials: 'omit',
      headers: {
        Authorization: 'Bearer ' + key.key, 'Content-Type': 'application/json',
        Accept: 'application/json', 'Idempotency-Key': 'canvas-' + creativeId(),
      },
      body: JSON.stringify(payload),
    })
    const data = await response.json() as TaskResponse
    if (!response.ok) {
      if ([400,401,402,403,404,413,415,422].includes(response.status)) {
        node.status = 'failed'
        return problem(data.error?.message || '请求被拒绝：HTTP ' + response.status)
      }
      node.status = 'submit_unknown'
      return problem('提交结果不确定（HTTP ' + response.status + '）。为防止重复扣费，请勿重新运行这个节点。')
    }
    const id = safeTaskId(data.task_id || data.id)
    if (!id) {
      node.status = 'submit_unknown'
      return problem('任务已受理但没有可靠任务编号。已禁止重复提交，请在后台核查。')
    }
    node.taskId = id
    node.status = data.status || 'queued'
    feedback('提交成功，任务正在后台处理。此节点不会再次自动提交。')
    pollUntilDone(node, 18)
  } catch {
    node.status = 'submit_unknown'
    problem('网络异常，可能已被上游接受；已锁定重复提交。请联系管理员核查任务。')
  } finally {
    runningNodes[node.id] = false
    saveNow()
  }
}

watch(shelf, () => {
  saveState.value = '正在保存…'
  if (saveHandle) window.clearTimeout(saveHandle)
  saveHandle = window.setTimeout(() => saveNow(), 420)
}, { deep: true })
watch(userId, (nextOwner) => {
  if (nextOwner === shelfOwnerId) return
  cloudReady.value = false
  cloudEpoch += 1
  cloudVersions.clear()
  cloudHashes.clear()
  if (saveHandle) window.clearTimeout(saveHandle)
  try { saveCreativeShelf(shelfOwnerId, shelf.value) } catch { /* best-effort flush of old owner's draft */ }
  shelfOwnerId = nextOwner
  shelf.value = loadCreativeShelf(shelfOwnerId)
  selectedId.value = ''
  selectedKeyId.value = ''
  apiKeys.value = []
  undoStack.value = []
  redoStack.value = []
  for (const timer of pollingTimers) window.clearTimeout(timer)
  pollingTimers.clear()
  void hydrateAssets()
  void loadKeys()
  void connectCloud()
})
onMounted(() => {
  window.addEventListener('keydown', onKeyboard)
  saveNow() // Persist starter project even if the user does not edit it.
  void hydrateAssets()
  void loadKeys()
  void connectCloud()
})
onBeforeUnmount(() => {
  destroyed = true
  cloudReady.value = false
  cloudEpoch += 1
  hydrationVersion += 1
  window.removeEventListener('keydown', onKeyboard)
  if (saveHandle) window.clearTimeout(saveHandle)
  touchPoints.clear()
  pinch = null
  saveNow() // Never lose the final <420ms of edits when navigating away.
  for (const timer of pollingTimers) window.clearTimeout(timer)
  for (const key of Object.keys(assetUrls)) URL.revokeObjectURL(assetUrls[key])
})
</script>

<template>
  <div class="cc-page">
    <header class="cc-header">
      <div class="cc-heading">
        <span class="cc-eyebrow"><span class="cc-orbit"></span> MUXWAY / CREATIVE STUDIO <i>PREVIEW 02</i></span>
        <h1>AI 创作画布</h1>
        <p>一个自由连接的创作空间。由提示词开始，把图片和视频延伸成完整的创作流程。</p>
      </div>
      <div class="cc-head-right">
        <span class="cc-save-indicator"><span></span>{{ saveState }}</span>
        <button type="button" class="cc-outline-button" @click="downloadProject">导出 JSON</button>
        <button type="button" class="cc-outline-button" @click="importPick">导入</button>
        <input ref="importInput" type="file" accept=".json,application/json" hidden @change="onImport" />
        <input ref="assetInput" type="file" accept="image/png,image/jpeg,image/webp,image/gif,video/mp4,video/webm" hidden @change="onAsset" />
      </div>
    </header>
    <div class="cc-toolbar">
      <div class="cc-project-select">
        <span class="cc-label-kicker">CURRENT BOARD</span>
        <select :value="shelf.activeId" aria-label="选择创作画布" @change="switchProject(($event.target as HTMLSelectElement).value)">
          <option v-for="item in shelf.projects" :key="item.id" :value="item.id">{{ item.title }}</option>
        </select>
      </div>
      <div class="cc-toolbar-divider"></div>
      <button type="button" title="新建画布" @click="makeProject">＋ 新建</button>
      <button type="button" title="复制当前画布" @click="copyProject">复制</button>
      <button type="button" title="删除当前画布" class="cc-danger-text" @click="removeProject">删除</button>
      <div class="cc-grow"></div>
      <span class="cc-local-pill">{{ cloudState }}</span>
      <button v-if="!cloudReady && !previewMode" type="button" title="重新连接云端" @click="connectCloud">↻ 重新连接</button>
      <button type="button" title="撤销" :disabled="!canUndo" @click="history('undo')">↶</button>
      <button type="button" title="重做" :disabled="!canRedo" @click="history('redo')">↷</button>
    </div>

    <section class="cc-studio">
      <aside class="cc-toolrail">
        <div class="cc-rail-section">
          <span class="cc-rail-title">CREATION TOOLS</span>
          <h2>创作工具</h2>
          <p>选择一种内容，添加到画布。</p>
        </div>
        <button v-for="item in tools" :key="item.kind" type="button" class="cc-tool" @click="insert(item.kind)">
          <span class="cc-tool-symbol" :class="'cc-symbol-' + item.kind">{{ item.icon }}</span>
          <span><strong>{{ item.label }}</strong><small>{{ item.hint }}</small></span>
          <span class="cc-add-glyph">+</span>
        </button>
        <div class="cc-rail-divider"></div>
        <div class="cc-rail-section cc-rail-guide">
          <span class="cc-rail-title">WORKFLOW</span>
          <strong>从想法到作品</strong>
          <p>拖动节点排列内容，点击右侧圆点，再点另一个节点左侧圆点，即可建立连线。</p>
          <div class="cc-step-mark"><b>01</b> 文本构思</div>
          <div class="cc-step-mark"><b>02</b> 图片创作</div>
          <div class="cc-step-mark"><b>03</b> 视频延伸</div>
        </div>
        <div class="cc-rail-bottom"><span class="cc-pulse"></span> NODE ENGINE <b>v1</b></div>
      </aside>

      <div class="cc-center">
        <div class="cc-canvas-subbar">
          <div class="cc-breadcrumb"><span class="cc-live-dot"></span><strong>{{ project.title }}</strong><span> / {{ project.nodes.length }} 个节点</span></div>
          <div class="cc-quick-add">
            <button type="button" @click="insert('prompt')">＋ 提示词</button>
            <button type="button" @click="insert('image')">＋ 图片</button>
            <button type="button" @click="insert('video')">＋ 视频</button>
          </div>
        </div>
        <div ref="viewport" class="cc-viewport" role="application" aria-label="AI 创作节点画布" :style="gridStyle"
          @pointerdown.capture="trackPointerDown" @pointermove.capture="trackPointerMove"
          @pointerup.capture="trackPointerEnd" @pointercancel.capture="trackPointerEnd"
          @pointerdown.self="pointerStartPan" @pointermove="pointerMove" @pointerup="finishPointer"
          @pointercancel="finishPointer" @wheel.prevent="onWheel">
          <div class="cc-stage" :style="stageStyle" @pointerdown.self="pointerStartPan">
            <svg class="cc-connectors" width="2800" height="1900" viewBox="0 0 2800 1900" aria-hidden="true">
              <defs><marker id="cc-arrow" markerWidth="9" markerHeight="8" refX="7.5" refY="4" orient="auto"><path d="M 0 0 L 8 4 L 0 8 Z" fill="#7c97aa" /></marker></defs>
              <path v-for="edge in edges" :key="edge.id" :d="edge.d" fill="none" stroke="#7897ae" stroke-opacity=".75" stroke-width="2" marker-end="url(#cc-arrow)"/>
            </svg>
            <article v-for="node in project.nodes" :key="node.id"
              class="cc-node" :class="['cc-node-' + node.kind, { 'cc-node-selected': selectedId === node.id }]"
              :style="{ left: node.x + 'px', top: node.y + 'px' }"
              @pointerdown.stop="pointerStartNode($event,node)" @click.stop="choose(node.id)">
              <button v-if="node.kind !== 'prompt'" type="button" class="cc-port cc-port-in"
                :aria-label="'连接到' + node.title" title="连接到这个节点"
                @pointerdown.stop @click.stop="endConnect(node.id)"></button>
              <button v-if="node.kind !== 'video'" type="button" class="cc-port cc-port-out"
                :class="{ 'cc-port-pending': connectionFrom === node.id }"
                :aria-label="'从' + node.title + '发起连线'" title="从这个节点发起连线"
                @pointerdown.stop @click.stop="beginConnect(node.id)"></button>
              <header class="cc-node-header">
                <span class="cc-node-type-icon">{{ tools.find(t => t.kind === node.kind)?.icon }}</span>
                <span class="cc-node-name"><small>{{ kindNames[node.kind] }}</small><strong>{{ node.title }}</strong></span>
                <span class="cc-node-more">•••</span>
              </header>
              <div class="cc-node-body">
                <template v-if="node.kind === 'prompt' || node.kind === 'note'">
                  <p :class="{ 'cc-placeholder': !node.prompt }">{{ node.prompt || (node.kind === 'prompt' ? '在右侧编写提示词…' : '记录你的灵感和说明…') }}</p>
                </template>
                <template v-else>
                  <template v-if="displayMedia(node)">
                    <video v-if="typeOfMedia(node) === 'video'" :src="displayMedia(node)" muted playsinline preload="metadata"></video>
                    <img v-else :src="displayMedia(node)" :alt="node.title" loading="lazy" referrerpolicy="no-referrer" />
                  </template>
                  <div v-else class="cc-media-placeholder">
                    <span>{{ node.kind === 'video' ? '▷' : '▧' }}</span>
                    <small>{{ node.status === 'idle' ? '等待创作内容' : statusNames[node.status] }}</small>
                  </div>
                </template>
              </div>
              <footer class="cc-node-footer">
                <span>{{ node.kind === 'image' || node.kind === 'video' ? node.model || '选择模型' : 'TEXT / INPUT' }}</span>
                <span class="cc-node-status" :class="'cc-status-' + node.status">{{ statusNames[node.status] }}</span>
              </footer>
            </article>
          </div>
          <div v-if="connectionFrom" class="cc-connecting"><span>连接中</span> 点击目标节点的左侧圆点 <button type="button" @click="clearConnection">取消</button></div>
          <div v-if="project.nodes.length === 0" class="cc-canvas-empty">
            <span>＋</span><strong>空白的空间，很多种可能</strong>
            <p>点击左侧工具或顶部快捷按钮，创建你的第一个节点。</p>
          </div>
          <div class="cc-canvas-bottom">
            <div class="cc-canvas-tips"><span>✦</span> 拖拽移动 · 双指捏合缩放 · Ctrl / ⌘ + 滚轮缩放</div>
            <div class="cc-zoom-controls">
              <button type="button" title="缩小" @click="zoomTo(zoom - 0.12)">−</button>
              <span>{{ Math.round(zoom * 100) }}%</span>
              <button type="button" title="放大" @click="zoomTo(zoom + 0.12)">+</button>
              <button type="button" title="复位视图" @click="resetView">适配</button>
            </div>
          </div>
        </div>
        <div class="cc-workspace-foot">
          <span><b>●</b> 创作草稿</span>
          <span>节点只会在你确认后提交到上游 · 不会自动产生费用</span>
        </div>
      </div>

      <aside class="cc-inspector">
        <template v-if="selectedNode">
          <div class="cc-inspector-header"><span>NODE INSPECTOR</span><button type="button" @click="selectedId = ''">×</button></div>
          <h2>{{ kindNames[selectedNode.kind] }}</h2>
          <p class="cc-inspector-intro">编辑选中节点的内容和参数。</p>
          <label class="cc-field">节点名称
            <input v-model="selectedNode.title" type="text" maxlength="90" placeholder="给节点命名" />
          </label>
          <label class="cc-field">提示词 / 内容
            <textarea v-model="selectedNode.prompt" rows="6" maxlength="6000"
              :placeholder="selectedNode.kind === 'prompt' ? '描述你希望生成的画面…' : selectedNode.kind === 'note' ? '写下灵感和备注…' : '可直接填写，或连接提示词节点…'"></textarea>
          </label>
          <template v-if="selectedNode.kind === 'image' || selectedNode.kind === 'video'">
            <div class="cc-inspector-rule"></div>
            <div class="cc-inspector-group-title"><span>GENERATION</span><strong>生成设置</strong></div>
            <label class="cc-field">模型 ID
              <input v-model="selectedNode.model" type="text" maxlength="200" :placeholder="selectedNode.kind === 'image' ? 'gpt-image-2.6' : 'omni-fast'" />
            </label>
            <label class="cc-field">精确参数（JSON）
              <textarea v-model="selectedNode.options" rows="3" spellcheck="false" maxlength="2500" placeholder='{"n":1}'></textarea>
            </label>
            <p class="cc-small-hint">模型和生成参数需与后台审核价一致，否则服务器会拒绝收费。图片默认仅生成一张；视频每次生成一条。</p>
            <label class="cc-field">参考素材
              <button type="button" class="cc-asset-button" @click="pickAsset">＋ 上传本地图片 / 视频</button>
            </label>
            <label class="cc-field">外部素材链接（HTTPS）
              <input :value="selectedNode.resultUrl" type="url" placeholder="https://..." @change="selectedNode!.resultUrl = safeMediaUrl(($event.target as HTMLInputElement).value)" />
            </label>
            <p class="cc-small-hint">当前上传素材仅用于画布展示；图生视频的实际输入映射将在下一阶段接通，避免错误提交不受支持的参数。</p>
            <div class="cc-inspector-rule"></div>
            <div class="cc-inspector-group-title"><span>API / BILLING</span><strong>调用凭证</strong></div>
            <label class="cc-field">我的 API Key（不保存到画布）
              <select v-model="selectedKeyId">
                <option value="">选择已授权分组</option>
                <option v-for="key in apiKeys" :key="key.id" :value="String(key.id)">{{ key.name || 'API Key #' + key.id }} · {{ key.group?.name || '分组 #' + key.group_id }}</option>
              </select>
            </label>
            <button type="button" class="cc-key-refresh" :disabled="keyLoading" @click="loadKeys">{{ keyLoading ? '读取中…' : '↻ 刷新我的 API Key' }}</button>
            <p v-if="!apiKeys.length" class="cc-small-hint">尚无可用 Key。请先在 <RouterLink to="/keys">API 密钥</RouterLink> 创建媒体分组凭证。</p>
            <div v-if="selectedNode.taskId" class="cc-task-info">任务编号 <strong>{{ selectedNode.taskId }}</strong><span>状态：{{ statusNames[selectedNode.status] }}</span></div>
            <div v-if="selectedNode.status === 'submit_unknown'" class="cc-task-warning">提交状态不确定。为避免重复扣费，此节点不能重新发送。请核对使用记录。</div>
            <button v-if="!selectedNode.taskId" type="button" class="cc-run-button"
              :disabled="runningNodes[selectedNode.id] || !isRunReady || selectedNode.status === 'submitting' || selectedNode.status === 'submit_unknown'"
              @click="runNode(selectedNode)">
              <span>✦</span> {{ state.user?.role !== 'admin' ? '付费运行 · 内测未开放' : runningNodes[selectedNode.id] ? '正在提交…' : '运行此节点 · 可能收费' }}
            </button>
            <button v-else type="button" class="cc-run-button cc-run-secondary" :disabled="pollingNodes[selectedNode.id]" @click="fetchTask(selectedNode)">
              {{ pollingNodes[selectedNode.id] ? '查询中…' : '查询生成任务' }}
            </button>
            <p class="cc-small-hint">Preview 01 仅允许管理员手动调用真实生成；普通用户可自由编辑。上线前还需接入精确价格预览。连接线不会自动触发付费执行。</p>
          </template>
          <div class="cc-inspector-rule"></div>
          <div class="cc-node-actions"><button type="button" @click="duplicateNode">复制节点</button><button type="button" @click="clearLinks">清除连线</button><button type="button" class="cc-danger-text" @click="deleteNode">删除节点</button></div>
        </template>
        <template v-else>
          <div class="cc-inspector-header"><span>WORKSPACE</span><span>◇</span></div>
          <h2>画布属性</h2>
          <p class="cc-inspector-intro">选中一个节点，即可编辑提示词、模型和素材。</p>
          <label class="cc-field">画布名称
            <input v-model="project.title" type="text" maxlength="90" placeholder="未命名画布" />
          </label>
          <div class="cc-inspector-rule"></div>
          <div class="cc-stats">
            <div><strong>{{ project.nodes.length }}</strong><small>创作节点</small></div>
            <div><strong>{{ project.links.length }}</strong><small>流程连线</small></div>
          </div>
          <div class="cc-inspector-rule"></div>
          <div class="cc-info-card"><span>✦ CREATIVE WORKFLOW</span><strong>连接你的创作过程</strong><p>用提示词节点控制画面描述，图片节点记录生成结果，视频节点承接后续制作。</p></div>
          <div class="cc-guide-link"><RouterLink to="/keys">管理我的 API 密钥 →</RouterLink></div>
        </template>
      </aside>
    </section>
    <div class="cc-feedback" role="status"><span v-if="error" class="cc-error">{{ error }}</span><span v-else>{{ notice }}</span></div>
  </div>
</template>
