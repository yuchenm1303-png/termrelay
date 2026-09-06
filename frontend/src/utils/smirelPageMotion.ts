let installed = false

export function installSmirelPageMotion() {
  if (installed || !document.documentElement.classList.contains('relay-standalone')) return
  installed = true

  // Keep the motion scope class because modal enter/leave feedback still uses it.
  // Route/page transitions are intentionally disabled: navigation should replace
  // the business surface immediately without opacity, blur or scale animation.
  document.documentElement.classList.add('smirel-page-motion')
}
