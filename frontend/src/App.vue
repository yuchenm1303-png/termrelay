<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, onBeforeUnmount, watch } from 'vue'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import AdminComplianceDialog from '@/components/admin/AdminComplianceDialog.vue'
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
    if (newLogo) {
      updateFavicon(newLogo)
    }
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
    () => adminSettingsStore.customMenuItems,
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
  if (authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
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
    // If setup endpoint fails, assume normal mode and continue
  }

  await appStore.fetchPublicSettings()
  updateDocumentTitle()
})
</script>

<template>
  <div class="termrelay-cosmos" aria-hidden="true">
    <div class="cosmic-nebula cosmic-nebula-cyan"></div>
    <div class="cosmic-nebula cosmic-nebula-violet"></div>
    <div class="cosmic-nebula cosmic-nebula-rose"></div>
    <div class="cosmic-stars cosmic-stars-small"></div>
    <div class="cosmic-stars cosmic-stars-large"></div>
    <div class="cosmic-flare cosmic-flare-a"></div>
    <div class="cosmic-flare cosmic-flare-b"></div>
    <div class="cosmic-flare cosmic-flare-c"></div>
    <div class="cosmic-planet"></div>
  </div>

  <NavigationProgress />
  <RouterView />
  <Toast />
  <AnnouncementPopup />
  <AdminComplianceDialog />
</template>

<style>
html,
body,
#app {
  min-height: 100%;
  background: #02040a;
}

.termrelay-cosmos {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  background:
    radial-gradient(circle at 50% 112%, rgba(160, 92, 210, 0.5), transparent 39%),
    radial-gradient(circle at 14% 72%, rgba(43, 205, 211, 0.2), transparent 31%),
    radial-gradient(circle at 88% 70%, rgba(239, 107, 181, 0.18), transparent 30%),
    linear-gradient(180deg, #010207 0%, #02050b 27%, #06212a 53%, #24314f 76%, #3d2454 100%);
}

.termrelay-cosmos::before,
.termrelay-cosmos::after {
  content: '';
  position: absolute;
  inset: 0;
}

.termrelay-cosmos::before {
  opacity: 0.9;
  background-image:
    radial-gradient(circle, rgba(255, 255, 255, 0.95) 0 1px, transparent 1.8px),
    radial-gradient(circle, rgba(116, 255, 244, 0.9) 0 1.2px, transparent 2px),
    radial-gradient(circle, rgba(255, 174, 225, 0.75) 0 0.9px, transparent 1.7px),
    radial-gradient(circle, rgba(255, 255, 255, 0.45) 0 0.7px, transparent 1.5px);
  background-position: 17px 31px, 91px 143px, 151px 73px, 211px 197px;
  background-size: 220px 220px, 310px 310px, 380px 380px, 165px 165px;
  animation: cosmic-star-drift 75s linear infinite;
}

.termrelay-cosmos::after {
  background:
    radial-gradient(ellipse at 50% 100%, rgba(136, 238, 227, 0.24), transparent 36%),
    radial-gradient(ellipse at 18% 92%, rgba(245, 125, 194, 0.2), transparent 30%),
    radial-gradient(ellipse at 82% 91%, rgba(130, 118, 238, 0.25), transparent 32%);
  filter: blur(22px);
}

.cosmic-nebula {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px) saturate(130%);
  mix-blend-mode: screen;
  opacity: 0.35;
}

.cosmic-nebula-cyan {
  width: 58vw;
  height: 34vw;
  left: -20vw;
  top: 45%;
  background: radial-gradient(ellipse, rgba(36, 220, 222, 0.85), rgba(29, 92, 129, 0.25) 55%, transparent 76%);
  animation: cosmic-nebula-left 18s ease-in-out infinite alternate;
}

.cosmic-nebula-violet {
  width: 55vw;
  height: 36vw;
  right: -18vw;
  top: 48%;
  background: radial-gradient(ellipse, rgba(153, 99, 235, 0.7), rgba(62, 50, 134, 0.25) 58%, transparent 78%);
  animation: cosmic-nebula-right 21s ease-in-out infinite alternate;
}

.cosmic-nebula-rose {
  width: 42vw;
  height: 24vw;
  left: 30%;
  bottom: -9vw;
  background: radial-gradient(ellipse, rgba(245, 103, 185, 0.58), rgba(118, 62, 144, 0.18) 58%, transparent 78%);
}

.cosmic-stars {
  position: absolute;
  inset: 0;
}

.cosmic-stars-small {
  opacity: 0.55;
  background-image:
    radial-gradient(circle at 7% 24%, #fff 0 1px, transparent 2px),
    radial-gradient(circle at 18% 47%, #8dfff5 0 1px, transparent 2px),
    radial-gradient(circle at 32% 18%, #fff 0 1px, transparent 2px),
    radial-gradient(circle at 44% 58%, #ffb7dc 0 1px, transparent 2px),
    radial-gradient(circle at 57% 29%, #a7fff8 0 1px, transparent 2px),
    radial-gradient(circle at 69% 64%, #fff 0 1px, transparent 2px),
    radial-gradient(circle at 81% 16%, #8ffcf8 0 1px, transparent 2px),
    radial-gradient(circle at 93% 43%, #fff 0 1px, transparent 2px),
    radial-gradient(circle at 74% 84%, #d5b4ff 0 1px, transparent 2px),
    radial-gradient(circle at 24% 82%, #fff 0 1px, transparent 2px);
  animation: cosmic-twinkle 4.8s ease-in-out infinite alternate;
}

.cosmic-stars-large {
  opacity: 0.85;
  background:
    radial-gradient(circle at 11% 38%, rgba(141, 255, 248, 0.98) 0 2px, rgba(63, 247, 255, 0.25) 4px, transparent 13px),
    radial-gradient(circle at 51% 31%, rgba(255, 255, 255, 0.98) 0 2px, rgba(255, 255, 255, 0.2) 5px, transparent 14px),
    radial-gradient(circle at 87% 25%, rgba(108, 255, 247, 0.98) 0 2px, rgba(63, 247, 255, 0.2) 5px, transparent 15px),
    radial-gradient(circle at 20% 66%, rgba(255, 177, 221, 0.98) 0 2px, rgba(255, 113, 193, 0.18) 5px, transparent 14px),
    radial-gradient(circle at 72% 74%, rgba(166, 137, 255, 0.95) 0 2px, rgba(142, 109, 255, 0.18) 5px, transparent 15px);
}

.cosmic-flare {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: white;
  box-shadow: 0 0 12px 3px rgba(142, 255, 249, 0.65);
  animation: cosmic-flare-pulse 3.6s ease-in-out infinite;
}

.cosmic-flare::before,
.cosmic-flare::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  background: linear-gradient(90deg, transparent, rgba(184, 255, 251, 0.9), transparent);
}

.cosmic-flare::before {
  width: 42px;
  height: 1px;
}

.cosmic-flare::after {
  width: 1px;
  height: 42px;
  background: linear-gradient(180deg, transparent, rgba(184, 255, 251, 0.9), transparent);
}

.cosmic-flare-a { left: 9%; top: 41%; }
.cosmic-flare-b { left: 61%; top: 28%; animation-delay: 1s; }
.cosmic-flare-c { right: 12%; top: 52%; animation-delay: 1.8s; }

.cosmic-planet {
  position: absolute;
  right: clamp(38px, 8vw, 160px);
  top: 64%;
  width: clamp(88px, 10vw, 150px);
  aspect-ratio: 1;
  border-radius: 50%;
  opacity: 0.68;
  background:
    radial-gradient(circle at 34% 27%, rgba(255, 210, 242, 0.25), transparent 19%),
    radial-gradient(circle at 66% 68%, rgba(24, 14, 64, 0.65), transparent 34%),
    linear-gradient(145deg, #83539c, #31335f 55%, #11172d);
  box-shadow:
    inset -18px -20px 30px rgba(3, 5, 18, 0.55),
    0 0 42px rgba(178, 116, 255, 0.2);
}

/* Remove the old technical grid and expose the new app-level sky. */
.termrelay-shell,
.auth-shell {
  background: transparent !important;
}

.termrelay-shell .grid-layer,
.auth-shell .auth-grid {
  background-image: none !important;
}

.termrelay-shell,
.auth-shell,
#app > main,
#app > section {
  position: relative;
  z-index: 1;
}

@keyframes cosmic-star-drift {
  from { transform: translate3d(0, 0, 0); }
  to { transform: translate3d(-80px, 45px, 0); }
}

@keyframes cosmic-twinkle {
  0% { opacity: 0.35; }
  100% { opacity: 0.82; }
}

@keyframes cosmic-flare-pulse {
  0%, 100% { opacity: 0.45; transform: scale(0.85); }
  50% { opacity: 1; transform: scale(1.15); }
}

@keyframes cosmic-nebula-left {
  from { transform: translate3d(-2%, -2%, 0) scale(0.96); }
  to { transform: translate3d(8%, 4%, 0) scale(1.08); }
}

@keyframes cosmic-nebula-right {
  from { transform: translate3d(3%, 2%, 0) scale(0.98); }
  to { transform: translate3d(-7%, -4%, 0) scale(1.07); }
}

@media (max-width: 720px) {
  .cosmic-planet {
    right: -26px;
    top: 70%;
    opacity: 0.45;
  }

  .cosmic-nebula-cyan,
  .cosmic-nebula-violet {
    width: 110vw;
    height: 70vw;
  }
}

@media (prefers-reduced-motion: reduce) {
  .termrelay-cosmos *,
  .termrelay-cosmos::before,
  .termrelay-cosmos::after {
    animation: none !important;
  }
}
</style>
