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

interface PendingMouse {
  clientX: number
  clientY: number
}

let installed = false

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

export function installSmirelDownloadInteractions() {
  if (installed || !document.documentElement.classList.contains('relay-standalone')) return
  installed = true

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  document.documentElement.classList.add('smirel-download-interactions')

  const bindCards = installCardInteractions(reducedMotion)
  const bindControls = installControlInteractions()

  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      mutation.addedNodes.forEach((node) => {
        if (!(node instanceof HTMLElement)) return
        bindCards(node)
        bindControls(node)
      })
    })
  })
  observer.observe(document.body, { childList: true, subtree: true })

  window.addEventListener('pagehide', () => {
    observer.disconnect()
  }, { once: true })
}
