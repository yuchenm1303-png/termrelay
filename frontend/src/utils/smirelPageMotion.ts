const PREFERRED_PAGE_SHELL_SELECTOR = [
  '.smh-shell',
  '.sw2-console',
  '.smg-shell',
  '[data-smirel-page-shell]',
].join(',')

const FALLBACK_PAGE_SHELL_SELECTOR = '.spg-page:not(.sw2-page)'
const ROUTE_STAGE_CHILD_SELECTOR = '.sw2-route-stage > *'

let installed = false

export function installSmirelPageMotion() {
  if (installed || !document.documentElement.classList.contains('relay-standalone')) return
  installed = true
  document.documentElement.classList.add('smirel-page-motion')

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reducedMotion) return

  const animated = new WeakSet<HTMLElement>()

  const animate = (element: HTMLElement | null) => {
    if (!element || animated.has(element)) return
    animated.add(element)
    element.classList.add('smirel-page-enter')
    element.addEventListener('animationend', () => {
      element.classList.remove('smirel-page-enter')
    }, { once: true })
  }

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

    const preferredShell = root.querySelector<HTMLElement>(PREFERRED_PAGE_SHELL_SELECTOR)
    if (preferredShell) {
      animate(preferredShell)
      return
    }

    animate(root.querySelector<HTMLElement>(FALLBACK_PAGE_SHELL_SELECTOR))
  }

  const preferredShell = document.querySelector<HTMLElement>(PREFERRED_PAGE_SHELL_SELECTOR)
  const fallbackShell = document.querySelector<HTMLElement>(FALLBACK_PAGE_SHELL_SELECTOR)
  animate(preferredShell ?? fallbackShell)

  const observer = new MutationObserver((records) => {
    for (const record of records) {
      for (const node of record.addedNodes) {
        if (node instanceof HTMLElement) bindFrom(node)
      }
    }
  })

  observer.observe(document.body, { childList: true, subtree: true })
}
