<template>
  <div class="termrelay-console-shell min-h-screen" :class="routeClass">
    <div class="console-backdrop" aria-hidden="true">
      <span class="console-aurora console-aurora-cyan"></span>
      <span class="console-aurora console-aurora-violet"></span>
      <span class="console-grid"></span>
    </div>

    <AppSidebar />

    <div
      class="termrelay-console-main relative min-h-screen transition-[margin] duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <AppHeader />

      <main class="termrelay-console-content p-4 md:p-6 lg:p-8">
        <div class="console-content-frame">
          <KeysOverviewBanner v-if="route.path === '/keys'" />
          <AccountsOverviewBanner v-if="route.path === '/admin/accounts'" />
          <UsageOverviewBanner v-if="route.path === '/usage'" />
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import '@/styles/termrelay-console.css'
import '@/styles/termrelay-keys.css'
import '@/styles/termrelay-accounts.css'
import '@/styles/termrelay-usage.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import KeysOverviewBanner from '@/components/keys/KeysOverviewBanner.vue'
import AccountsOverviewBanner from '@/components/admin/account/AccountsOverviewBanner.vue'
import UsageOverviewBanner from '@/components/usage/UsageOverviewBanner.vue'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const routeClass = computed(() => {
  const routeName = String(route.name || 'console')
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/[^a-zA-Z0-9-]/g, '-')
    .toLowerCase()
  return `route-${routeName}`
})

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
