// Local-only uploaded images/video: blobs never go to the API or repository.
// IndexedDB avoids base64 media in localStorage and keeps browser drafts bounded.
// Cross-device sync/export of binary assets is intentionally a later milestone.
const DB_NAME = 'muxway-creative-assets-v1'
const STORE = 'assets'

interface AssetRow { key: string; blob: Blob; name: string; updatedAt: number }

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    if (typeof indexedDB === 'undefined') {
      reject(Error('当前浏览器不支持本地素材存储'))
      return
    }
    const request = indexedDB.open(DB_NAME, 1)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(STORE)) db.createObjectStore(STORE, { keyPath: 'key' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(Error('无法访问浏览器素材存储'))
  })
}

function assetKey(userId: number, assetId: string): string {
  if (!Number.isSafeInteger(userId) || userId < 0 || !/^cr_[a-zA-Z0-9_-]{4,75}$/.test(assetId)) {
    throw Error('素材归属信息无效')
  }
  return userId + ':' + assetId
}

export async function storeCreativeAsset(userId: number, assetId: string, file: File): Promise<void> {
  if (!['image/png', 'image/jpeg', 'image/webp', 'image/gif', 'video/mp4', 'video/webm'].includes(file.type)) {
    throw Error('仅支持 PNG、JPEG、WebP、GIF、MP4 或 WebM')
  }
  const max = file.type.startsWith('video/') ? 32 * 1024 * 1024 : 12 * 1024 * 1024
  if (file.size === 0 || file.size > max) throw Error('图片不能超过 12MB，视频不能超过 32MB')
  const db = await openDb()
  try {
    await new Promise<void>((resolve, reject) => {
      const tx = db.transaction(STORE, 'readwrite')
      tx.objectStore(STORE).put({ key: assetKey(userId, assetId), blob: file, name: file.name.slice(0, 120), updatedAt: Date.now() } as AssetRow)
      tx.oncomplete = () => resolve()
      tx.onerror = () => reject(Error('素材保存失败，请检查浏览器存储空间'))
      tx.onabort = () => reject(Error('素材保存被取消'))
    })
  } finally { db.close() }
}

export async function readCreativeAsset(userId: number, assetId: string): Promise<Blob | null> {
  const db = await openDb()
  try {
    return await new Promise<Blob | null>((resolve, reject) => {
      const tx = db.transaction(STORE, 'readonly')
      const request = tx.objectStore(STORE).get(assetKey(userId, assetId))
      request.onsuccess = () => {
        const row = request.result as AssetRow | undefined
        resolve(row?.blob instanceof Blob ? row.blob : row?.blob || null)
      }
      request.onerror = () => reject(Error('素材读取失败'))
    })
  } finally { db.close() }
}
