import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initI18n } from './i18n'
import { useAppStore } from '@/stores/app'
import { updateFavicon } from '@/utils/branding'
import { isIOSDevice } from '@/utils/device'
import {
  applyStandaloneSmirelPreferences,
  applyStandaloneSmirelPublicSettings,
  isStandaloneSmirelFrontend,
  SMIREL_API_BASE_URL,
  SMIREL_SITE_LOGO,
  SMIREL_SITE_NAME,
} from '@/utils/smirelStandalone'
import './style.css'
import './styles/smirel-relay.css'
import './styles/smirel-console-shell.css'
import './styles/smirel-public-pages.css'
import './styles/smirel-relay-interactions.css'

function applyStandaloneRelayShell() {
  if (!isStandaloneSmirelFrontend()) return

  // Mark the independently hosted Smirel console without forcing a theme.
  // Light/dark preference remains user-controlled instead of being coupled to
  // the legacy wallpaper/glass skin.
  document.documentElement.classList.add('relay-standalone')
  applyStandaloneSmirelPreferences()
}

async function applyStandaloneRelayBrand(appStore: ReturnType<typeof useAppStore>) {
  if (!isStandaloneSmirelFrontend()) return

  // Injected settings bypass Axios, while API-loaded settings pass through the
  // response interceptor. Normalize both paths with the same central policy so
  // no component can observe a different brand/API endpoint depending on how
  // the settings were loaded.
  const loadedSettings = await appStore.fetchPublicSettings()
  const sourceSettings = loadedSettings || appStore.cachedPublicSettings
  if (sourceSettings) {
    const brandedSettings = applyStandaloneSmirelPublicSettings(sourceSettings)
    appStore.cachedPublicSettings = brandedSettings
    window.__APP_CONFIG__ = { ...brandedSettings }
  }

  appStore.siteName = SMIREL_SITE_NAME
  appStore.siteLogo = SMIREL_SITE_LOGO
  appStore.apiBaseUrl = SMIREL_API_BASE_URL
  appStore.publicSettingsLoaded = true
}

function initIOSViewportZoomFix() {
  // iOS Safari 在输入框字号小于 16px 时聚焦会自动放大页面，且失焦后不会恢复。
  // 限制 maximum-scale 可阻止该行为；iOS 10+ 用户仍可双指手动缩放，不影响可访问性。
  // 仅在 iOS 设备上注入，避免影响 Android Chrome 的手动缩放能力。
  if (!isIOSDevice()) return

  const viewport = document.querySelector('meta[name="viewport"]')
  if (!viewport) return

  const content = viewport.getAttribute('content') || ''
  if (/maximum-scale/i.test(content)) return
  viewport.setAttribute('content', `${content}, maximum-scale=1.0`)
}

function initThemeClass() {
  const savedTheme = localStorage.getItem('theme')
  const shouldUseDark =
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

async function bootstrap() {
  // Apply theme class globally before app mount to keep all routes consistent.
  initThemeClass()
  applyStandaloneRelayShell()
  initIOSViewportZoomFix()

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)

  // Initialize settings from injected config BEFORE mounting (prevents flash)
  // This must happen after pinia is installed but before router and i18n
  const appStore = useAppStore()
  appStore.initFromInjectedConfig()
  await applyStandaloneRelayBrand(appStore)

  // Set document title immediately after config is loaded
  if (appStore.siteName && appStore.siteName !== 'Sub2API') {
    document.title = `${appStore.siteName} - AI API Gateway`
  }
  updateFavicon(appStore.siteLogo)

  await initI18n()

  app.use(router)
  app.use(i18n)

  // 等待路由器完成初始导航后再挂载，避免竞态条件导致的空白渲染
  await router.isReady()
  app.mount('#app')
}

bootstrap()
