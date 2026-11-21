import { beforeEach, describe, expect, it } from 'vitest'
import {
  blankProject, canConnect, importCreativeProject, loadCreativeShelf,
  makeNode, normalizeProject, safeMediaUrl, saveCreativeShelf, starterProject,
} from '../core/creative-canvas'

beforeEach(() => localStorage.clear())

describe('creative canvas project model', () => {
  it('ships a real prompt -> image -> video starter graph without fake output', () => {
    const project = starterProject()
    expect(project.nodes.map(n => n.kind)).toEqual(['prompt', 'image', 'video'])
    expect(project.links).toHaveLength(2)
    expect(project.nodes.every(n => n.status === 'idle' && !n.taskId && !n.resultUrl)).toBe(true)
    expect(canConnect(project, project.nodes[0].id, project.nodes[2].id)).toBe(true)
    expect(canConnect(project, project.nodes[2].id, project.nodes[0].id)).toBe(false)
  })

  it('isolates each signed-in user and never needs API keys in project storage', () => {
    const alpha = loadCreativeShelf(17)
    const beta = loadCreativeShelf(18)
    expect(alpha.activeId).not.toBe(beta.activeId)
    alpha.projects[0].title = 'Private alpha draft'
    saveCreativeShelf(17, alpha)
    expect(loadCreativeShelf(17).projects[0].title).toBe('Private alpha draft')
    expect(loadCreativeShelf(18).projects[0].title).not.toBe('Private alpha draft')
    expect(JSON.stringify(loadCreativeShelf(17))).not.toContain('Authorization')
  })

  it('rejects cycles, duplicate links, self-links and unsupported source/target types', () => {
    const project = blankProject()
    const a = makeNode('prompt', 0, 0)
    const b = makeNode('image', 120, 200)
    const c = makeNode('video', 400, 200)
    project.nodes = [a, b, c]
    expect(canConnect(project, a.id, b.id)).toBe(true)
    project.links.push({ id: 'edge1', from: a.id, to: b.id })
    expect(canConnect(project, a.id, b.id)).toBe(false)
    expect(canConnect(project, b.id, a.id)).toBe(false)
    expect(canConnect(project, b.id, c.id)).toBe(true)
    project.links.push({ id: 'edge2', from: b.id, to: c.id })
    expect(canConnect(project, c.id, b.id)).toBe(false)
    expect(canConnect(project, b.id, b.id)).toBe(false)
  })

  it('rejects malformed imports and strips unsafe URLs, preserving only HTTPS results', () => {
    const project = starterProject()
    const media = project.nodes.find(n => n.kind === 'image')!
    media.resultUrl = 'javascript:alert(1)'
    const reloaded = normalizeProject(project)
    expect(reloaded?.nodes.find(n => n.id === media.id)?.resultUrl).toBe('')
    expect(safeMediaUrl('http://example.org/x.png')).toBe('')
    expect(safeMediaUrl('https://example.com/example.png')).toBe('https://example.com/example.png')
    expect(safeMediaUrl('https://user:password@example.com/x')).toBe('')
    expect(importCreativeProject(JSON.stringify(project)).id).not.toBe(project.id)
    expect(() => importCreativeProject('{')).toThrow()
    expect(() => importCreativeProject(JSON.stringify({ ...project, version: 999 }))).toThrow()
  })

  it('rejects oversized graphs and repairs corrupted local JSON', () => {
    const project = blankProject()
    project.nodes = Array.from({ length: 121 }, (_, index) => makeNode('note', index, 0))
    expect(normalizeProject(project)).toBeNull()
    localStorage.setItem('muxway.creation.projects.v1.user-20', '{invalid')
    expect(loadCreativeShelf(20).projects).toHaveLength(1)
  })

  it('preserves an async task reference across local reload without storing its API credential', () => {
    const project = starterProject()
    const node = project.nodes.find(item => item.kind === 'image')!
    node.status = 'queued'
    node.taskId = 'cyjob_' + 'a'.repeat(32)
    node.taskKeyId = 47
    const shelf = { activeId: project.id, projects: [project] }
    saveCreativeShelf(47, shelf)
    const restored = loadCreativeShelf(47)
    const image = restored.projects[0].nodes.find(x => x.kind === 'image')
    expect(image?.taskId).toBe(node.taskId)
    expect(image?.taskKeyId).toBe(47)
    expect(image?.status).toBe('queued')
    expect(JSON.stringify(restored)).not.toContain('sk-')
  })
})
