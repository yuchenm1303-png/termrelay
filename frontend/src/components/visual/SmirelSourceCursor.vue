<template>
  <div ref="dot" class="smirel-source-cursor-dot" aria-hidden="true"></div>
  <div ref="follow" class="smirel-source-cursor-follow is-hidden" aria-hidden="true"></div>
  <canvas ref="fireworks" class="smirel-source-fireworks" aria-hidden="true"></canvas>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const dot = ref<HTMLDivElement | null>(null)
const follow = ref<HTMLDivElement | null>(null)
const fireworks = ref<HTMLCanvasElement | null>(null)

let raf = 0
let curr: { x: number; y: number } | null = null
let prev: { x: number; y: number } | null = null
let clearAnimation: any = null
let fireworkMouseDown: ((event: MouseEvent) => void) | null = null
let resizeFireworks: (() => void) | null = null

function renderFollow() {
  raf = 0
  if (!curr || !follow.value) return

  if (prev) {
    prev.x += (curr.x - prev.x) * 0.35
    prev.y += (curr.y - prev.y) * 0.35
  } else {
    prev = { ...curr }
  }

  follow.value.style.translate = `${prev.x}px ${prev.y}px`
  if (Math.abs(curr.x - prev.x) > 0.01 || Math.abs(curr.y - prev.y) > 0.01) {
    raf = requestAnimationFrame(renderFollow)
  }
}

function queueFollow() {
  if (!raf) raf = requestAnimationFrame(renderFollow)
}

function onMouseMove(event: MouseEvent) {
  if (!dot.value || !follow.value) return
  dot.value.style.translate = `${event.clientX}px ${event.clientY}px`
  dot.value.classList.add('is-visible')

  if (curr === null) follow.value.style.translate = `${event.clientX - 8}px ${event.clientY - 8}px`
  curr = { x: event.clientX - 8, y: event.clientY - 8 }
  follow.value.classList.remove('is-hidden')
  follow.value.classList.add('is-visible')
  queueFollow()
}

function onMouseEnter() {
  dot.value?.classList.add('is-visible')
  follow.value?.classList.remove('is-hidden')
}

function onMouseLeave() {
  dot.value?.classList.remove('is-visible')
  follow.value?.classList.add('is-hidden')
}

function onMouseDown() {
  follow.value?.classList.add('is-active')
}

function onMouseUp() {
  follow.value?.classList.remove('is-active')
}

function loadAnime321(): Promise<any> {
  const globalAnime = (window as any).anime
  if (globalAnime?.version === '3.2.1') return Promise.resolve(globalAnime)

  return new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>('script[data-listing-studio-anime="3.2.1"]')
    if (existing) {
      existing.addEventListener('load', () => resolve((window as any).anime), { once: true })
      existing.addEventListener('error', reject, { once: true })
      return
    }

    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/animejs@3.2.1/lib/anime.min.js'
    script.async = true
    script.dataset.listingStudioAnime = '3.2.1'
    script.addEventListener('load', () => resolve((window as any).anime), { once: true })
    script.addEventListener('error', reject, { once: true })
    document.head.appendChild(script)
  })
}

function installFireworks(anime: any) {
  const canvas = fireworks.value
  if (!canvas || !anime) return
  const context = canvas.getContext('2d')
  if (!context) return

  const colors = ['252, 146, 174', '202, 180, 190', '207, 198, 255']

  resizeFireworks = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    canvas.style.width = `${window.innerWidth}px`
    canvas.style.height = `${window.innerHeight}px`
  }

  function randomEndPosition(origin: { x: number; y: number }) {
    const angle = anime.random(0, 360) * Math.PI / 180
    const diffuseRadius = anime.random(50, 100)
    const signedRadius = [-1, 1][anime.random(0, 1)] * diffuseRadius
    return {
      x: origin.x + signedRadius * Math.cos(angle),
      y: origin.y + signedRadius * Math.sin(angle),
    }
  }

  function createParticle(x: number, y: number) {
    return {
      x,
      y,
      color: `rgba(${colors[anime.random(0, colors.length - 1)]},${anime.random(.2, .8)})`,
      radius: anime.random(10, 20),
      angle: anime.random(0, 360),
      endPos: randomEndPosition({ x, y }),
      draw() {
        context.save()
        context.translate(this.x, this.y)
        context.rotate(this.angle * Math.PI / 180)
        context.beginPath()
        context.moveTo(0, -this.radius)
        context.lineTo(this.radius * Math.sin(Math.PI / 3), this.radius * Math.cos(Math.PI / 3))
        context.lineTo(-this.radius * Math.sin(Math.PI / 3), this.radius * Math.cos(Math.PI / 3))
        context.closePath()
        context.fillStyle = this.color
        context.fill()
        context.restore()
      },
    }
  }

  function createCircle(x: number, y: number) {
    return {
      x,
      y,
      color: 'rgb(233, 179, 237)',
      radius: .1,
      alpha: .5,
      lineWidth: 6,
      draw() {
        context.globalAlpha = this.alpha
        context.beginPath()
        context.arc(this.x, this.y, this.radius, 0, 2 * Math.PI, true)
        context.lineWidth = this.lineWidth
        context.strokeStyle = this.color
        context.stroke()
        context.globalAlpha = 1
      },
    }
  }

  function drawTargets(animation: any) {
    animation.animatables.forEach(({ target }: any) => target.draw?.())
  }

  function createFirework(x: number, y: number) {
    const circle = createCircle(x, y)
    const particles = Array.from({ length: 20 }, () => createParticle(x, y))

    anime.timeline().add({
      targets: particles,
      x: (particle: any) => particle.endPos.x,
      y: (particle: any) => particle.endPos.y,
      radius: 0,
      duration: anime.random(900, 1500),
      easing: 'easeOutExpo',
      update: drawTargets,
    }).add({
      targets: circle,
      radius: anime.random(50, 100),
      lineWidth: 0,
      alpha: { value: 0, easing: 'linear', duration: anime.random(600, 800) },
      duration: anime.random(1200, 1800),
      easing: 'easeOutExpo',
      update: drawTargets,
    }, 0)
  }

  clearAnimation = anime({
    duration: Number.POSITIVE_INFINITY,
    update: () => context.clearRect(0, 0, canvas.width, canvas.height),
  })

  fireworkMouseDown = (event: MouseEvent) => {
    clearAnimation?.play?.()
    createFirework(event.clientX, event.clientY)
  }

  window.addEventListener('resize', resizeFireworks)
  document.addEventListener('mousedown', fireworkMouseDown)
  resizeFireworks()
}

onMounted(() => {
  if (!document.documentElement.classList.contains('relay-standalone')) return

  document.addEventListener('mousemove', onMouseMove, { passive: true })
  document.addEventListener('mouseenter', onMouseEnter)
  document.addEventListener('mouseleave', onMouseLeave)
  document.addEventListener('mousedown', onMouseDown)
  document.addEventListener('mouseup', onMouseUp)

  loadAnime321().then(installFireworks).catch(() => {})
})

onBeforeUnmount(() => {
  if (raf) cancelAnimationFrame(raf)
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseenter', onMouseEnter)
  document.removeEventListener('mouseleave', onMouseLeave)
  document.removeEventListener('mousedown', onMouseDown)
  document.removeEventListener('mouseup', onMouseUp)
  if (fireworkMouseDown) document.removeEventListener('mousedown', fireworkMouseDown)
  if (resizeFireworks) window.removeEventListener('resize', resizeFireworks)
  clearAnimation?.pause?.()
})
</script>
