import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './smirel/core/i18n'
import { restoreInterfacePreferences } from './smirel/core/preferences'
import { restoreNotifications } from './smirel/core/notifications'
import { restoreSession } from './smirel/core/session'
import './smirel/styles/app.css'
import './smirel/styles/button-interactions.css'
import './smirel/styles/workspace-topbar.css'
import './smirel/styles/workspace-utility-popovers.css'
import './smirel/styles/workspace-brand.css'
import './smirel/styles/interface-preferences.css'
import './smirel/styles/select-controls.css'
import './smirel/styles/models-commercial.css'
import './smirel/styles/model-catalog-workspace.css'

async function bootstrap() {
  document.documentElement.classList.add('smirel-app')
  document.title = 'Smirel API · Unified AI Gateway'
  restoreInterfacePreferences()
  restoreNotifications()
  await restoreSession()

  const app = createApp(App)
  app.use(i18n)
  app.use(router)
  await router.isReady()
  app.mount('#app')
}

void bootstrap()
