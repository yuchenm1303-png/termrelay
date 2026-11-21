// Creation canvas project format v1. Documents are intentionally local to the
// signed-in browser/user until an authenticated cloud persistence API exists.
// Never put access tokens or API key material into a canvas document.
export type CanvasNodeKind = 'prompt' | 'image' | 'video' | 'note'
export type CanvasTaskStatus = 'idle' | 'reserved' | 'queued' | 'in_progress' | 'submitting' | 'submit_unknown' | 'completed' | 'failed'

export interface CreativeNode {
  id: string
  kind: CanvasNodeKind
  x: number
  y: number
  title: string
  prompt: string
  model: string
  options: string
  resultUrl: string
  assetId: string
  taskId: string
  taskKeyId: number | null
  status: CanvasTaskStatus
}

export interface CreativeLink { id: string; from: string; to: string }
export interface CreativeProject {
  version: 1
  id: string
  title: string
  createdAt: string
  updatedAt: string
  nodes: CreativeNode[]
  links: CreativeLink[]
}
export interface CreativeShelf { activeId: string; projects: CreativeProject[] }

const prefix = 'muxway.creation.projects.v1'
const maxNodes = 120
const maxLinks = 250
const kinds: CanvasNodeKind[] = ['prompt', 'image', 'video', 'note']
const statuses: CanvasTaskStatus[] = ['idle', 'reserved', 'queued', 'in_progress', 'submitting', 'submit_unknown', 'completed', 'failed']
const safeId = /^[a-zA-Z0-9_-]{1,80}$/

function clipped(value: unknown, size: number): string {
  return typeof value === 'string' ? value.slice(0, size) : ''
}
function coOrd(value: unknown, fallback: number): number {
  const num = Number(value)
  return Number.isFinite(num) ? Math.max(-8000, Math.min(8000, num)) : fallback
}
export function safeMediaUrl(value: unknown): string {
  if (typeof value !== 'string' || value.length > 2048) return ''
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'https:' && parsed.username === '' && parsed.password === '' ? parsed.href : ''
  } catch { return '' }
}
export function creativeId(): string {
  const random = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID().replace(/-/g, '')
    : String(Date.now()) + Math.random().toString(36).slice(2)
  return 'cr_' + random
}

export function makeNode(kind: CanvasNodeKind, x: number, y: number): CreativeNode {
  const titles = { prompt: '提示词', image: '图片生成', video: '视频生成', note: '便签' }
  return {
    id: creativeId(), kind, x: coOrd(x, 100), y: coOrd(y, 120),
    title: titles[kind], prompt: '', model: kind === 'image' ? 'gpt-image-2.6' : kind === 'video' ? 'omni-fast' : '',
    options: kind === 'image' ? '{"n":1}' : '{}',
    resultUrl: '', assetId: '', taskId: '', taskKeyId: null, status: 'idle',
  }
}

export function starterProject(): CreativeProject {
  const now = new Date().toISOString()
  const prompt = makeNode('prompt', 80, 170)
  prompt.prompt = '一只漫步在未来城市的橘猫，电影质感，柔和光影'
  const image = makeNode('image', 430, 170)
  const video = makeNode('video', 780, 170)
  return {
    version: 1, id: creativeId(), title: '我的第一张创作画布',
    createdAt: now, updatedAt: now,
    nodes: [prompt, image, video],
    links: [{ id: creativeId(), from: prompt.id, to: image.id }, { id: creativeId(), from: image.id, to: video.id }],
  }
}
export function blankProject(title = '未命名画布'): CreativeProject {
  const now = new Date().toISOString()
  return { version: 1, id: creativeId(), title: title.slice(0, 90), createdAt: now, updatedAt: now, nodes: [], links: [] }
}

export function normalizeProject(raw: unknown): CreativeProject | null {
  if (!raw || typeof raw !== 'object') return null
  const source = raw as Partial<CreativeProject>
  if (source.version !== 1 || !Array.isArray(source.nodes) || !Array.isArray(source.links) ||
    source.nodes.length > maxNodes || source.links.length > maxLinks ||
    !safeId.test(String(source.id ?? ''))) return null
  const ids = new Set<string>()
  const nodes: CreativeNode[] = []
  for (const item of source.nodes) {
    if (!item || typeof item !== 'object' || !safeId.test(String(item.id ?? '')) ||
      !kinds.includes(item.kind) || ids.has(item.id)) return null
    ids.add(item.id)
    nodes.push({
      id: item.id, kind: item.kind, x: coOrd(item.x, 100), y: coOrd(item.y, 100),
      title: clipped(item.title, 90), prompt: clipped(item.prompt, 6000),
      model: clipped(item.model, 200), options: clipped(item.options, 2500) || '{}',
      resultUrl: safeMediaUrl(item.resultUrl), assetId: safeId.test(String(item.assetId ?? '')) ? item.assetId : '',
      taskId: /^cyjob_[a-zA-Z0-9_-]{1,72}$/.test(String(item.taskId ?? '')) ? item.taskId : '',
      taskKeyId: Number.isSafeInteger(item.taskKeyId) && Number(item.taskKeyId) > 0 ? Number(item.taskKeyId) : null,
      // A browser reload during an in-flight POST is ambiguous. Treat it as
      // unknown rather than offering another paid submission.
      status: item.status === 'submitting' && !item.taskId
        ? 'submit_unknown'
        : statuses.includes(item.status) ? item.status : 'idle',
    })
  }
  const links: CreativeLink[] = []
  const pairs = new Set<string>()
  for (const edge of source.links) {
    if (!edge || typeof edge !== 'object' || !safeId.test(String(edge.id ?? '')) ||
      !ids.has(edge.from) || !ids.has(edge.to) || edge.from === edge.to) return null
    const pair = edge.from + ':' + edge.to
    if (pairs.has(pair)) continue
    pairs.add(pair)
    links.push({ id: edge.id, from: edge.from, to: edge.to })
  }
  return {
    version: 1, id: source.id as string,
    title: clipped(source.title, 90) || '未命名画布',
    createdAt: clipped(source.createdAt, 40) || new Date().toISOString(),
    updatedAt: clipped(source.updatedAt, 40) || new Date().toISOString(),
    nodes, links,
  }
}

export function canConnect(project: CreativeProject, from: string, to: string): boolean {
  if (from === to || project.links.length >= maxLinks || project.links.some(x => x.from === from && x.to === to)) return false
  const fromNode = project.nodes.find(x => x.id === from)
  const toNode = project.nodes.find(x => x.id === to)
  if (!fromNode || !toNode || fromNode.kind === 'video' || toNode.kind === 'prompt') return false
  // Source-to-target must be a DAG so chained generation can later run safely.
  const checked = new Set<string>()
  const pending = [to]
  while (pending.length) {
    const current = pending.pop()!
    if (current === from) return false
    if (checked.has(current)) continue
    checked.add(current)
    for (const edge of project.links) if (edge.from === current) pending.push(edge.to)
  }
  return true
}

function storageKey(userId: number | string): string {
  return prefix + '.user-' + String(userId).replace(/[^0-9]/g, '')
}

export function loadCreativeShelf(userId: number | string, storage: Pick<Storage, 'getItem'> = localStorage): CreativeShelf {
  try {
    const raw = storage.getItem(storageKey(userId))
    if (raw && raw.length <= 750_000) {
      const parsed = JSON.parse(raw) as Partial<CreativeShelf>
      const projects = Array.isArray(parsed.projects) ? parsed.projects.slice(0, 24)
        .map(normalizeProject).filter((item): item is CreativeProject => Boolean(item)) : []
      if (projects.length) {
        return { activeId: projects.some(item => item.id === parsed.activeId) ? parsed.activeId! : projects[0].id, projects }
      }
    }
  } catch { /* corrupted or storage-disabled browsers start with a clean local canvas */ }
  const initial = starterProject()
  return { activeId: initial.id, projects: [initial] }
}

export function saveCreativeShelf(userId: number | string, shelf: CreativeShelf, storage: Pick<Storage, 'setItem'> = localStorage): void {
  const projects = shelf.projects.slice(0, 24)
  if (!projects.length || projects.some(item => !normalizeProject(item))) throw Error('画布数据无效，未写入本地储存')
  const serialized = JSON.stringify({ activeId: shelf.activeId, projects })
  if (serialized.length > 750_000) throw Error('本地画布容量已满，请导出备份或减少项目')
  storage.setItem(storageKey(userId), serialized)
}

export function importCreativeProject(input: string): CreativeProject {
  if (input.length > 500_000) throw Error('文件太大，无法安全导入')
  let parsed: unknown
  try { parsed = JSON.parse(input) } catch { throw Error('不是有效的画布 JSON 文件') }
  const doc = normalizeProject(parsed)
  if (!doc) throw Error('文件格式不受支持或数据不完整')
  const now = new Date().toISOString()
  return { ...doc, id: creativeId(), title: doc.title + ' · 导入', createdAt: now, updatedAt: now,
    // Imported projects cannot take over another user's private media job or
    // local IndexedDB asset ID. Keep prompt/model history and HTTPS results.
    nodes: doc.nodes.map(node => ({ ...node, status: 'idle', taskId: '', taskKeyId: null, assetId: '' })),
  }
}
