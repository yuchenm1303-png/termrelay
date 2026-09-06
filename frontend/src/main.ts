import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { restoreSession } from './smirel/core/session'
import './smirel/styles/app.css'
import './smirel/styles/workspace-shell.css'
import './smirel/styles/card-motion.css'
import './smirel/styles/background.css'

async function bootstrap() {
  document.documentElement.classList.add('smirel-app')
  document.documentElement.style.setProperty(
    '--smirel-background-image',
    `url("${import.meta.env.BASE_URL}smirel-cosmic-bg.webp")`,
  )
  document.title = 'Smirel API · Unified AI Gateway'
  await restoreSession()

  const app = createApp(App)
  app.use(router)
  await router.isReady()
  app.mount('#app')
}

void bootstrap()
