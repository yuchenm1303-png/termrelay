import { describe, expect, it } from 'vitest'
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { resolve } from 'node:path'

function collect(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const full = resolve(dir, name)
    return statSync(full).isDirectory() ? collect(full) : [full]
  })
}

describe('Smirel frontend purity contract', () => {
  it('contains no legacy product identity in runtime source', () => {
    const src = resolve(process.cwd(), 'src')
    const runtimeFiles = collect(src).filter((file) => !file.includes('__tests__'))
    const offenders = runtimeFiles.filter((file) => /Sub2API|sub2api/.test(readFileSync(file, 'utf8')))
    expect(offenders).toEqual([])
  })

  it('ships only the Smirel application source tree', () => {
    const srcEntries = readdirSync(resolve(process.cwd(), 'src')).sort()
    expect(srcEntries).toEqual(['App.vue', 'main.ts', 'router', 'smirel'])
  })

  it('removes legacy identity from the HTML/package entry points and assets', () => {
    const indexHtml = readFileSync(resolve(process.cwd(), 'index.html'), 'utf8')
    const packageJson = readFileSync(resolve(process.cwd(), 'package.json'), 'utf8')
    expect(indexHtml).not.toMatch(/Sub2API|sub2api/)
    expect(packageJson).not.toMatch(/Sub2API|sub2api/)
    expect(existsSync(resolve(process.cwd(), 'public/logo.svg'))).toBe(false)
  })
})
