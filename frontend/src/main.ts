import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initI18n } from './i18n'
import { useAppStore } from '@/stores/app'
import { updateFavicon } from '@/utils/branding'
import { isIOSDevice } from '@/utils/device'
import { installSmirelDownloadInteractions } from '@/utils/smirelDownloadInteractions'
import {
  applyStandaloneSmirelPreferences,
  applyStandaloneSmirelPublicSettings,
  isStandaloneSmirelFrontend,
  SMIREL_API_BASE_URL,
  SMIREL_SITE_LOGO,
  SMIREL_SITE_NAME,
} from '@/utils/smirelStandalone'
import { installSmirelUiPreview, isSmirelUiPreview } from '@/utils/smirelUiPreview'
import './style.css'
import './styles/smirel-public-v5.css'
import './styles/smirel-glass-v5.css'
import './styles/smirel-glass-pages-v5.css'
import './styles/smirel-glass-controls-v5.css'
import './styles/smirel-glass-dashboard-v5.css'
import './styles/smirel-glass-auth-v5.css'
import './styles/smirel-wordmark-v1.css'
import './styles/smirel-shared-material-v1.css'
import './styles/smirel-card-system-v1.css'
import './styles/smirel-inset-system-v1.css'
import './styles/smirel-home-topbar-v1.css'
import './styles/smirel-shared-interactions-v1.css'

function applyStandaloneRelayShell() {
  if (!isStandaloneSmirelFrontend()) return
  document.documentElement.classList.add('relay-standalone')
  applyStandaloneSmirelPreferences()
}

async function applyStandaloneRelayBrand(appStore: ReturnType<typeof useAppStore>) {
  if (!isStandaloneSmirelFrontend()) return

  // GitHub Pages is a static review target. It intentionally does not call the
  // production settings API; production/standalone deployments keep the real
  // settings bootstrap below.
  if (!isSmirelUiPreview) {
    const loadedSettings = await appStore.fetchPublicSettings()
    const sourceSettings = loadedSettings || appStore.cachedPublicSettings
    if (sourceSettings) {
      const brandedSettings = applyStandaloneSmirelPublicSettings(sourceSettings)
      appStore.cachedPublicSettings = brandedSettings
      window.__APP_CONFIG__ = { ...brandedSettings }
    }
  }

  appStore.siteName = SMIREL_SITE_NAME
  appStore.siteLogo = SMIREL_SITE_LOGO
  appStore.apiBaseUrl = SMIREL_API_BASE_URL
  appStore.publicSettingsLoaded = true
}

function initIOSViewportZoomFix() {
  if (!isIOSDevice()) return
  const viewport = document.querySelector('meta[name="viewport"]')
  if (!viewport) return
  const content = viewport.getAttribute('content') || ''
  if (/maximum-scale/i.test(content)) return
  viewport.setAttribute('content', `${content}, maximum-scale=1.0`)
}

function initThemeClass() {
  const savedTheme = localStorage.getItem('theme')
  const shouldUseDark = savedTheme === 'dark'
    || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

async function bootstrap() {
  initThemeClass()
  applyStandaloneRelayShell()
  initIOSViewportZoomFix()

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)

  const appStore = useAppStore()
  appStore.initFromInjectedConfig()

  // Must run after Pinia registration and before Router starts its first guard.
  installSmirelUiPreview()

  await applyStandaloneRelayBrand(appStore)
  if (appStore.siteName && appStore.siteName !== 'Sub2API') {
    document.title = `${appStore.siteName} - AI API Gateway`
  }
  updateFavicon(appStore.siteLogo)

  await initI18n()
  app.use(router)
  app.use(i18n)
  await router.isReady()
  app.mount('#app')
  installSmirelDownloadInteractions()
}

bootstrap()
