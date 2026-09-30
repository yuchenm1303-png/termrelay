import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { syncRouteSEO } from './smirel/core/seo'
import i18n from './smirel/core/i18n'
import { restoreInterfacePreferences } from './smirel/core/preferences'
import { restoreNotifications } from './smirel/core/notifications'
import { restoreSession } from './smirel/core/session'
import './smirel/styles/app.css'
import './smirel/styles/button-interactions.css'
import './smirel/styles/workspace-topbar.css'
import './smirel/styles/workspace-breadcrumb-polish.css'
import './smirel/styles/workspace-utility-popovers.css'
import './smirel/styles/interface-preferences.css'
import './smirel/styles/admin-groups-light.css'
import './smirel/styles/models-commercial.css'
import './smirel/styles/model-catalog-workspace.css'
import './smirel/styles/user-usage-polish.css'
import './smirel/styles/user-usage-readability.css'
import './smirel/styles/provider-logos.css'
import './smirel/styles/provider-logos-live.css'
import './smirel/styles/model-capability-filter.css'
import './smirel/styles/model-card-readability.css'
import './smirel/styles/api-keys-overview-readability.css'
import './smirel/styles/api-key-modal-alignment.css'
import './smirel/styles/workspace-account-alignment.css'
import './smirel/styles/oauth-callback-polish.css'
import './smirel/styles/light-theme-hardening.css'
import './smirel/styles/admin-payment-light.css'
import './smirel/styles/admin-resource-light.css'
import './smirel/styles/workspace-light-tail.css'
import './smirel/styles/user-light-complete.css'
import './smirel/styles/user-usage-trend-polish.css'
import './smirel/styles/model-card-compact.css'
import './smirel/styles/provider-segmented.css'
import './smirel/styles/admin-groups-polish.css'
import './smirel/styles/group-display-icons.css'
import './smirel/styles/account-settings-light.css'
import './smirel/styles/route-transitions.css'
import './smirel/styles/async-data-motion.css'
import './smirel/styles/workspace-responsive.css'
import './smirel/styles/workspace-nav-controls.css'
import './smirel/styles/admin-accounts-polish.css'
import './smirel/styles/workspace-ambient.css'

async function bootstrap() {
  document.documentElement.classList.add('smirel-app')
  // Keep the crawlable document title intact until the router resolves.
  restoreInterfacePreferences()
  restoreNotifications()
  await restoreSession()

  const app = createApp(App)
  app.use(i18n)
  app.use(router)
  await router.isReady()
  syncRouteSEO(router.currentRoute.value.path)
  app.mount('#app')
}

void bootstrap()
