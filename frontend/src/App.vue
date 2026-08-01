<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import CosmicBackground from '@/components/common/CosmicBackground.vue'
import AdminComplianceDialog from '@/components/admin/AdminComplianceDialog.vue'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import { resolveRouteDocumentTitle } from '@/router/title'
import {
  useAppStore,
  useAuthStore,
  useSubscriptionStore,
  useAnnouncementStore,
  useAdminComplianceStore,
  useAdminSettingsStore
} from '@/stores'
import { getSetupStatus } from '@/api/setup'
import { updateFavicon } from '@/utils/branding'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const announcementStore = useAnnouncementStore()
const adminComplianceStore = useAdminComplianceStore()
const adminSettingsStore = useAdminSettingsStore()

const cosmicRoutePrefixes = [
  '/',
  '/home',
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/verify-email',
  '/setup'
]

const showCosmos = computed(() => {
  if (route.path === '/') return true
  return cosmicRoutePrefixes.slice(1).some((prefix) => route.path.startsWith(prefix))
})

function normalizeBrandName() {
  if (!appStore.siteName || appStore.siteName === 'Sub2API') {
    appStore.siteName = 'TermRelay'
  }

  const publicSettings = appStore.cachedPublicSettings
  if (publicSettings && (!publicSettings.site_name || publicSettings.site_name === 'Sub2API')) {
    publicSettings.site_name = 'TermRelay'
  }
}

watch(
  [() => appStore.siteName, () => appStore.cachedPublicSettings?.site_name],
  normalizeBrandName,
  { immediate: true }
)

function updateDocumentTitle() {
  const customMenuItems = [
    ...(appStore.cachedPublicSettings?.custom_menu_items ?? []),
    ...(authStore.isAdmin ? adminSettingsStore.customMenuItems : [])
  ]
  document.title = resolveRouteDocumentTitle(route, appStore.siteName, customMenuItems)
}

watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) updateFavicon(newLogo)
  },
  { immediate: true }
)

watch(
  [
    () => route.fullPath,
    () => route.meta.title,
    () => route.meta.titleKey,
    () => appStore.siteName,
    () => appStore.cachedPublicSettings?.custom_menu_items,
    () => authStore.isAdmin,
    () => adminSettingsStore.customMenuItems
  ],
  updateDocumentTitle,
  { deep: true }
)

function onVisibilityChange() {
  if (document.visibilityState === 'visible' && authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
}

function onAdminComplianceRequired(event: Event) {
  const detail = (event as CustomEvent<Record<string, string>>).detail || {}
  adminComplianceStore.requireAcknowledgement(detail)
}

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated, oldValue) => {
    if (isAuthenticated) {
      if (authStore.isAdmin) {
        adminComplianceStore.fetchStatus().catch((error) => {
          console.error('Failed to fetch admin compliance status:', error)
        })
      }

      subscriptionStore.fetchActiveSubscriptions().catch((error) => {
        console.error('Failed to preload subscriptions:', error)
      })
      subscriptionStore.startPolling()

      if (oldValue === false) {
        setTimeout(() => announcementStore.fetchAnnouncements(true), 3000)
      } else {
        announcementStore.fetchAnnouncements()
      }

      document.addEventListener('visibilitychange', onVisibilityChange)
    } else {
      subscriptionStore.clear()
      announcementStore.reset()
      adminComplianceStore.reset()
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  },
  { immediate: true }
)

router.afterEach(() => {
  if (authStore.isAuthenticated) announcementStore.fetchAnnouncements()
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  window.removeEventListener('admin-compliance-required', onAdminComplianceRequired)
})

onMounted(async () => {
  window.addEventListener('admin-compliance-required', onAdminComplianceRequired)

  try {
    const status = await getSetupStatus()
    if (status.needs_setup && route.path !== '/setup') {
      router.replace('/setup')
      return
    }
  } catch {
    // Static previews and temporarily unavailable backends should still render the UI.
  }

  await appStore.fetchPublicSettings()
  normalizeBrandName()
  updateDocumentTitle()
})
</script>

<template>
  <CosmicBackground v-if="showCosmos" />
  <div class="termrelay-app-layer">
    <NavigationProgress />
    <RouterView />
    <Toast />
    <AnnouncementPopup />
    <AdminComplianceDialog />
  </div>
</template>

<style>
html,
body,
#app {
  min-height: 100%;
  background: #010305;
}

body {
  margin: 0;
  overflow-x: hidden;
}

#app {
  position: relative;
  isolation: isolate;
}

.termrelay-app-layer {
  position: relative;
  z-index: 1;
  min-height: 100vh;
}

.auth-shell {
  background: transparent !important;
}

.auth-shell .auth-grid,
.auth-shell .auth-glow {
  display: none !important;
}

.auth-shell .auth-scanlines {
  opacity: 0.01 !important;
}
</style>
