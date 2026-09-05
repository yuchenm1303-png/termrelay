<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { computed, onMounted, onBeforeUnmount, watch } from 'vue'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import AdminComplianceDialog from '@/components/admin/AdminComplianceDialog.vue'
import SmirelPortalShell from '@/components/layout/SmirelPortalShell.vue'
import SmirelBackground from '@/components/visual/SmirelBackground.vue'
import SmirelCursor from '@/components/visual/SmirelCursor.vue'
import { resolveRouteDocumentTitle } from '@/router/title'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import { useAppStore, useAuthStore, useSubscriptionStore, useAnnouncementStore, useAdminComplianceStore, useAdminSettingsStore } from '@/stores'
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

const PORTAL_PATHS = [
  '/key-usage',
  '/model-plaza',
  '/setup',
  '/payment/result',
  '/payment/airwallex',
  '/payment/qrcode',
]

const isEmbeddedModelPlaza = computed(() =>
  route.path === '/model-plaza'
  && route.query.embedded === '1'
  && authStore.isAuthenticated,
)

const usePortalShell = computed(() => {
  if (isEmbeddedModelPlaza.value) return false
  return route.name === 'NotFound'
    || route.path.startsWith('/legal/')
    || PORTAL_PATHS.some((path) => route.path === path || route.path.startsWith(`${path}/`))
})

function updateDocumentTitle() {
  const customMenuItems = [
    ...(appStore.cachedPublicSettings?.custom_menu_items ?? []),
    ...(authStore.isAdmin ? adminSettingsStore.customMenuItems : []),
  ]
  document.title = resolveRouteDocumentTitle(route, appStore.siteName, customMenuItems)
}

watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) updateFavicon(newLogo)
  },
  { immediate: true },
)

watch(
  [
    () => route.fullPath,
    () => route.meta.title,
    () => route.meta.titleKey,
    () => appStore.siteName,
    () => appStore.cachedPublicSettings?.custom_menu_items,
    () => authStore.isAdmin,
    () => adminSettingsStore.customMenuItems,
  ],
  updateDocumentTitle,
  { deep: true },
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
  { immediate: true },
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
    // If setup endpoint fails, assume normal mode and continue.
  }

  await appStore.fetchPublicSettings()
  updateDocumentTitle()
})
</script>

<template>
  <SmirelBackground />
  <NavigationProgress />
  <RouterView v-slot="{ Component }">
    <SmirelPortalShell v-if="usePortalShell">
      <component :is="Component" />
    </SmirelPortalShell>
    <component :is="Component" v-else />
  </RouterView>
  <Toast />
  <AnnouncementPopup />
  <AdminComplianceDialog />
  <SmirelCursor />
</template>
