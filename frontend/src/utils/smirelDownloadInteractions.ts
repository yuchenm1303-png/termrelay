const CARD_SELECTOR = [
  '[data-smirel-motion="card"]',
  '[data-smirel-motion="utility"]',
  '.smirel-card--interactive',
  '.spg-surface--interactive',
  '.smh-glass',
  '.smh-quick-card',
  '.sw2-account-card',
  '.sw2-sidebar-context',
  '.sw2-panel',
  '.sw2-admin-module',
  '.sw2-admin-metric',
  '.sw2-admin-diagnostic',
  '.sw2-admin-health-metric',
  '.sw2-admin-health-foot > div',
  '.sw2-admin-today-foot > div',
  '.sw2-route-stage .smg-surface',
  '.sw2-route-stage .smg-data-panel',
  '.sw2-route-stage .card',
  '.sw2-route-stage .glass-card',
  '.smg-shell .smg-surface',
  '.smg-shell .card',
  '.smg-shell .glass-card',
  '.modal-content',
].join(',')

const CARD_EXCLUDE_SELECTOR = [
  '[data-smirel-motion="none"]',
  '.smh-header',
  '.sw2-sidebar',
  '.sw2-topbar',
].join(',')

const UTILITY_CARD_SELECTOR = [
  '[data-smirel-motion="utility"]',
  '.smh-quick-card',
  '.spg-utility-card',
  '.sw2-admin-health-metric',
  '.sw2-admin-health-foot > div',
  '.sw2-admin-today-foot > div',
].join(',')

const CONTROL_SELECTOR = 'a[href], button, [role="button"]'
const CONTROL_EXCLUDE_SELECTOR = [
  '[data-smirel-motion="none"]',
  '.sw2-nav-scrim',
].join(',')

const PREFERRED_PAGE_SHELL_SELECTOR = [
  '.smh-shell',
  '.sw2-console',
  '.smg-shell',
  '[data-smirel-page-shell]',
].join(',')

const FALLBACK_PAGE_SHELL_SELECTOR = '.spg-page:not(.sw2-page)'
const ROUTE_STAGE_CHILD_SELECTOR = '.sw2-route-stage > *'

interface Point {
  x: number
  y: number
}

interface PendingMouse {
  clientX: number
  clientY: number
}

interface FireworkParticle {
  x: number
  y: number
  startX: number
  startY: number
  endX: number
  endY: number
  radius: number
  startRadius: number
  angle: number
  color: string
  duration: number
}

interface FireworkCircle {
  x: number
  y: number
  targetRadius: number
  alphaDuration: number
  duration: number
}

interface FireworkBurst {
  startedAt: number
  particles: FireworkParticle[]
  circle: FireworkCircle
}

let installed = false

function isHTMLElement(value: EventTarget | Node | null): value is HTMLElement {
  return value instanceof HTMLElement
}

function findClosestMotionCard(target: EventTarget | null): HTMLElement | null {
  if (!(target instanceof Element)) return null

  let candidate = target.closest<HTMLElement>(CARD_SELECTOR)
  while (candidate?.matches(CARD_EXCLUDE_SELECTOR)) {
    candidate = candidate.parentElement?.closest<HTMLElement>(CARD_SELECTOR) ?? null
  }
  return candidate
}

function setCardMotion(
  card: HTMLElement,
  x: number,
  y: number,
  scale: number,
  duration: string,
  easing: string,
) {
  card.style.setProperty('--smirel-motion-x', `${x}px`)
  card.style.setProperty('--smirel-motion-y', `${y}px`)
  card.style.setProperty('--smirel-motion-scale', String(scale))
  card.style.setProperty('--smirel-motion-duration', duration)
  card.style.setProperty('--smirel-motion-ease', easing)
}

function installCardInteractions(reducedMotion: boolean) {
  const boundCards = new WeakSet<HTMLElement>()

  const bindCard = (card: HTMLElement) => {
    if (boundCards.has(card) || card.matches(CARD_EXCLUDE_SELECTOR)) return
    boundCards.add(card)

    const utility = card.matches(UTILITY_CARD_SELECTOR)
    const hoverScale = utility ? 1.02 : 1.01
    card.classList.add('smirel-download-card-motion')
    if (utility) card.classList.add('smirel-download-card-motion--utility')

    if (reducedMotion) return

    let mouseInside = false
    let mouseRect: DOMRect | null = null
    let pendingMouse: PendingMouse | null = null
    let mouseRaf = 0
    let touchId: number | null = null
    let touchRect: DOMRect | null = null
    let bounceTimer = 0

    const setHot = (active: boolean) => {
      if (utility) card.classList.toggle('smirel-download-card-motion--hot', active)
    }

    const clearCard = () => {
      if (mouseRaf) {
        cancelAnimationFrame(mouseRaf)
        mouseRaf = 0
      }
      if (bounceTimer) {
        window.clearTimeout(bounceTimer)
        bounceTimer = 0
      }
      pendingMouse = null
      mouseRect = null
      setHot(false)
      setCardMotion(card, 0, 0, 1, '220ms', 'ease')
    }

    const renderMouseMove = () => {
      mouseRaf = 0
      if (!mouseInside || !mouseRect || !pendingMouse) return

      if (findClosestMotionCard(card.ownerDocument.elementFromPoint(
        pendingMouse.clientX,
        pendingMouse.clientY,
      )) !== card) {
        setCardMotion(card, 0, 0, 1, '120ms', 'ease-out')
        setHot(false)
        pendingMouse = null
        return
      }

      const { clientX, clientY } = pendingMouse
      pendingMouse = null
      const nx = (clientX - mouseRect.left) / mouseRect.width - 0.5
      const ny = (clientY - mouseRect.top) / mouseRect.height - 0.5
      setCardMotion(card, nx * 5, ny * 4, hoverScale, '120ms', 'ease-out')
      setHot(true)
    }

    const bounceHome = () => {
      setCardMotion(card, 0, 0, hoverScale, '95ms', 'ease-out')
      bounceTimer = window.setTimeout(() => {
        bounceTimer = 0
        if (touchId === null && !mouseInside) clearCard()
      }, 105)
    }

    card.addEventListener('mouseenter', () => {
      mouseInside = true
      mouseRect = card.getBoundingClientRect()
      card.style.setProperty('--smirel-motion-duration', '120ms')
      card.style.setProperty('--smirel-motion-ease', 'ease-out')
    })

    card.addEventListener('mousemove', (event) => {
      if (findClosestMotionCard(event.target) !== card) {
        setCardMotion(card, 0, 0, 1, '120ms', 'ease-out')
        setHot(false)
        return
      }
      pendingMouse = { clientX: event.clientX, clientY: event.clientY }
      if (!mouseRaf) mouseRaf = requestAnimationFrame(renderMouseMove)
    }, { passive: true })

    card.addEventListener('mousedown', (event) => {
      if (findClosestMotionCard(event.target) !== card) return
      setCardMotion(card, 0, 0, 0.98, '90ms', 'ease-out')
      setHot(utility)
    })

    card.addEventListener('mouseup', (event) => {
      if (!mouseInside || findClosestMotionCard(event.target) !== card) return
      setCardMotion(card, 0, 0, hoverScale, '140ms', 'ease-out')
      setHot(utility)
    })

    card.addEventListener('mouseleave', () => {
      mouseInside = false
      if (touchId === null) clearCard()
    })

    card.addEventListener('touchstart', (event) => {
      if (findClosestMotionCard(event.target) !== card || event.changedTouches.length !== 1) return
      const touch = event.changedTouches[0]
      touchId = touch.identifier
      touchRect = card.getBoundingClientRect()
      setCardMotion(card, 0, 0, 0.98, '85ms', 'ease-out')
      setHot(utility)
    }, { passive: true })

    card.addEventListener('touchmove', (event) => {
      if (touchId === null || !touchRect) return
      const touch = Array.from(event.changedTouches).find((item) => item.identifier === touchId)
      if (!touch) return
      const nx = Math.max(-0.5, Math.min(0.5, (touch.clientX - touchRect.left) / touchRect.width - 0.5))
      const ny = Math.max(-0.5, Math.min(0.5, (touch.clientY - touchRect.top) / touchRect.height - 0.5))
      setCardMotion(card, nx * 3.2, ny * 2.8, 0.985, '70ms', 'linear')
    }, { passive: true })

    const finishTouch = (event: TouchEvent) => {
      if (touchId === null) return
      if (event.changedTouches.length) {
        const matched = Array.from(event.changedTouches).some((item) => item.identifier === touchId)
        if (!matched) return
      }
      touchId = null
      touchRect = null
      bounceHome()
    }

    card.addEventListener('touchend', finishTouch, { passive: true })
    card.addEventListener('touchcancel', finishTouch, { passive: true })
  }

  const bindFrom = (root: ParentNode) => {
    if (root instanceof HTMLElement && root.matches(CARD_SELECTOR)) bindCard(root)
    root.querySelectorAll<HTMLElement>(CARD_SELECTOR).forEach(bindCard)
  }

  bindFrom(document)
  return bindFrom
}

function installControlInteractions() {
  const boundControls = new WeakSet<HTMLElement>()

  const bindControl = (control: HTMLElement) => {
    if (boundControls.has(control) || control.matches(CONTROL_EXCLUDE_SELECTOR)) return
    boundControls.add(control)
    if (control.classList.contains('smirel-download-card-motion')) return
    control.classList.add('smirel-download-control-motion')
  }

  const bindFrom = (root: ParentNode) => {
    if (root instanceof HTMLElement && root.matches(CONTROL_SELECTOR)) bindControl(root)
    root.querySelectorAll<HTMLElement>(CONTROL_SELECTOR).forEach(bindControl)
  }

  bindFrom(document)
  return bindFrom
}

function installPageEntryMotion(reducedMotion: boolean) {
  const animated = new WeakSet<HTMLElement>()

  const animate = (element: HTMLElement) => {
    if (reducedMotion || animated.has(element)) return
    animated.add(element)
    element.classList.add('smirel-download-page-enter')
    element.addEventListener('animationend', () => {
      element.classList.remove('smirel-download-page-enter')
    }, { once: true })
  }

  const preferredShell = document.querySelector<HTMLElement>(PREFERRED_PAGE_SHELL_SELECTOR)
  const fallbackShell = document.querySelector<HTMLElement>(FALLBACK_PAGE_SHELL_SELECTOR)
  animate(preferredShell ?? fallbackShell ?? document.body)

  const bindFrom = (root: ParentNode) => {
    if (root instanceof HTMLElement) {
      if (root.matches(ROUTE_STAGE_CHILD_SELECTOR)) {
        animate(root)
        return
      }
      if (root.matches(PREFERRED_PAGE_SHELL_SELECTOR) || root.matches(FALLBACK_PAGE_SHELL_SELECTOR)) {
        animate(root)
      }
    }

    const routeStageChild = root.querySelector<HTMLElement>(ROUTE_STAGE_CHILD_SELECTOR)
    if (routeStageChild) {
      animate(routeStageChild)
      return
    }

    const preferredShellChild = root.querySelector<HTMLElement>(PREFERRED_PAGE_SHELL_SELECTOR)
    if (preferredShellChild) {
      animate(preferredShellChild)
      return
    }

    const fallbackShellChild = root.querySelector<HTMLElement>(FALLBACK_PAGE_SHELL_SELECTOR)
    if (fallbackShellChild) animate(fallbackShellChild)
  }

  return bindFrom
}

function installCursor(reducedMotion: boolean) {
  const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)').matches
  if (reducedMotion || !finePointer) return () => {}

  const dot = document.createElement('div')
  const follow = document.createElement('div')
  dot.className = 'smirel-cursor-dot'
  follow.className = 'smirel-cursor-follow hidden'
  dot.setAttribute('aria-hidden', 'true')
  follow.setAttribute('aria-hidden', 'true')
  document.body.append(dot, follow)
  document.documentElement.classList.add('smirel-custom-cursor')

  let curr: Point | null = null
  let prev: Point | null = null
  let raf = 0

  const moveDot = (x: number, y: number) => {
    dot.style.translate = `${x}px ${y}px`
  }

  const moveFollow = (x: number, y: number) => {
    follow.style.translate = `${x}px ${y}px`
  }

  const render = () => {
    raf = 0
    if (!curr) return

    if (prev) {
      prev.x += (curr.x - prev.x) * 0.35
      prev.y += (curr.y - prev.y) * 0.35
      moveFollow(prev.x, prev.y)
    } else {
      prev = { ...curr }
      moveFollow(prev.x, prev.y)
    }

    if (Math.abs(curr.x - prev.x) > 0.01 || Math.abs(curr.y - prev.y) > 0.01) {
      raf = requestAnimationFrame(render)
    }
  }

  const queueRender = () => {
    if (!raf) raf = requestAnimationFrame(render)
  }

  const onMouseMove = (event: MouseEvent) => {
    moveDot(event.clientX, event.clientY)
    dot.classList.add('cursor-visible')
    if (curr === null) moveFollow(event.clientX - 8, event.clientY - 8)
    curr = { x: event.clientX - 8, y: event.clientY - 8 }
    follow.classList.remove('hidden')
    queueRender()
  }

  const onMouseEnter = () => {
    dot.classList.add('cursor-visible')
    follow.classList.remove('hidden')
  }

  const onMouseLeave = () => {
    dot.classList.remove('cursor-visible')
    follow.classList.add('hidden')
  }

  const onMouseDown = () => follow.classList.add('active')
  const onMouseUp = () => follow.classList.remove('active')

  document.addEventListener('mousemove', onMouseMove, { passive: true })
  document.addEventListener('mouseenter', onMouseEnter)
  document.addEventListener('mouseleave', onMouseLeave)
  document.addEventListener('mousedown', onMouseDown)
  document.addEventListener('mouseup', onMouseUp)

  return () => {
    if (raf) cancelAnimationFrame(raf)
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseenter', onMouseEnter)
    document.removeEventListener('mouseleave', onMouseLeave)
    document.removeEventListener('mousedown', onMouseDown)
    document.removeEventListener('mouseup', onMouseUp)
    document.documentElement.classList.remove('smirel-custom-cursor')
    dot.remove()
    follow.remove()
  }
}

function randomInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function easeOutExpo(progress: number) {
  if (progress >= 1) return 1
  return 1 - Math.pow(2, -10 * progress)
}

function installFireworks(reducedMotion: boolean) {
  if (reducedMotion) return () => {}

  const colors = ['252, 146, 174', '202, 180, 190', '207, 198, 255']
  const canvas = document.createElement('canvas')
  canvas.className = 'smirel-fireworks'
  canvas.setAttribute('aria-hidden', 'true')
  document.body.appendChild(canvas)

  const context = canvas.getContext('2d')
  if (!context) {
    canvas.remove()
    return () => {}
  }

  let bursts: FireworkBurst[] = []
  let raf = 0

  const resize = () => {
    const ratio = Math.min(window.devicePixelRatio || 1, 2)
    canvas.width = Math.round(window.innerWidth * ratio)
    canvas.height = Math.round(window.innerHeight * ratio)
    canvas.style.width = `${window.innerWidth}px`
    canvas.style.height = `${window.innerHeight}px`
    context.setTransform(ratio, 0, 0, ratio, 0, 0)
  }

  const createBurst = (x: number, y: number) => {
    const particles = Array.from({ length: 20 }, () => {
      const angle = randomInt(0, 360) * Math.PI / 180
      const diffuseRadius = randomInt(50, 100) * (Math.random() < 0.5 ? -1 : 1)
      const radius = randomInt(10, 20)
      const color = colors[randomInt(0, colors.length - 1)]
      return {
        x,
        y,
        startX: x,
        startY: y,
        endX: x + diffuseRadius * Math.cos(angle),
        endY: y + diffuseRadius * Math.sin(angle),
        radius,
        startRadius: radius,
        angle: randomInt(0, 360),
        color: `rgba(${color},${(randomInt(20, 80) / 100).toFixed(2)})`,
        duration: randomInt(900, 1500),
      }
    })

    bursts.push({
      startedAt: performance.now(),
      particles,
      circle: {
        x,
        y,
        targetRadius: randomInt(50, 100),
        alphaDuration: randomInt(600, 800),
        duration: randomInt(1200, 1800),
      },
    })

    if (!raf) raf = requestAnimationFrame(draw)
  }

  const drawTriangle = (particle: FireworkParticle) => {
    context.save()
    context.translate(particle.x, particle.y)
    context.rotate(particle.angle * Math.PI / 180)
    context.beginPath()
    context.moveTo(0, -particle.radius)
    context.lineTo(particle.radius * Math.sin(Math.PI / 3), particle.radius * Math.cos(Math.PI / 3))
    context.lineTo(-particle.radius * Math.sin(Math.PI / 3), particle.radius * Math.cos(Math.PI / 3))
    context.closePath()
    context.fillStyle = particle.color
    context.fill()
    context.restore()
  }

  function draw(now: number) {
    raf = 0
    context.clearRect(0, 0, window.innerWidth, window.innerHeight)

    bursts = bursts.filter((burst) => {
      const elapsed = now - burst.startedAt
      let alive = false

      burst.particles.forEach((particle) => {
        const progress = Math.min(1, elapsed / particle.duration)
        const eased = easeOutExpo(progress)
        particle.x = particle.startX + (particle.endX - particle.startX) * eased
        particle.y = particle.startY + (particle.endY - particle.startY) * eased
        particle.radius = particle.startRadius * (1 - eased)
        if (progress < 1) {
          alive = true
          drawTriangle(particle)
        }
      })

      const circleProgress = Math.min(1, elapsed / burst.circle.duration)
      if (circleProgress < 1) {
        alive = true
        const eased = easeOutExpo(circleProgress)
        const alphaProgress = Math.min(1, elapsed / burst.circle.alphaDuration)
        context.save()
        context.globalAlpha = 0.5 * (1 - alphaProgress)
        context.beginPath()
        context.arc(
          burst.circle.x,
          burst.circle.y,
          burst.circle.targetRadius * eased,
          0,
          Math.PI * 2,
        )
        context.lineWidth = 6 * (1 - eased)
        context.strokeStyle = 'rgb(233, 179, 237)'
        context.stroke()
        context.restore()
      }

      return alive
    })

    if (bursts.length) raf = requestAnimationFrame(draw)
  }

  const onMouseDown = (event: MouseEvent) => createBurst(event.clientX, event.clientY)
  const onResize = () => resize()

  document.addEventListener('mousedown', onMouseDown)
  window.addEventListener('resize', onResize)
  resize()

  return () => {
    document.removeEventListener('mousedown', onMouseDown)
    window.removeEventListener('resize', onResize)
    if (raf) cancelAnimationFrame(raf)
    bursts = []
    canvas.remove()
  }
}

export function installSmirelDownloadInteractions() {
  if (installed || !document.documentElement.classList.contains('relay-standalone')) return
  installed = true

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  document.documentElement.classList.add('smirel-download-interactions')

  const bindCards = installCardInteractions(reducedMotion)
  const bindControls = installControlInteractions()
  const bindPageEntry = installPageEntryMotion(reducedMotion)
  const removeCursor = installCursor(reducedMotion)
  const removeFireworks = installFireworks(reducedMotion)

  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      mutation.addedNodes.forEach((node) => {
        if (!isHTMLElement(node)) return
        bindCards(node)
        bindControls(node)
        bindPageEntry(node)
      })
    })
  })
  observer.observe(document.body, { childList: true, subtree: true })

  window.addEventListener('pagehide', () => {
    observer.disconnect()
    removeCursor()
    removeFireworks()
  }, { once: true })
}
