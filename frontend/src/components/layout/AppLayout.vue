<template>
  <div class="relay-app-shell min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Background Decoration -->
    <div class="relay-console-backdrop pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Primary user workflow navigation. Admin keeps the existing admin IA untouched. -->
      <UserRelayNav v-if="!isAdmin" />

      <!-- Main Content -->
      <main class="p-4 md:p-6 lg:p-8">
        <RelayAccessPanel v-if="showRelayAccessPanel" />
        <BillingAccessPanel v-if="showBillingAccessPanel" />
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import UserRelayNav from './UserRelayNav.vue'
import RelayAccessPanel from '@/components/user/RelayAccessPanel.vue'
import BillingAccessPanel from '@/components/user/BillingAccessPanel.vue'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const showRelayAccessPanel = computed(() => !isAdmin.value && route.path === '/keys')
const showBillingAccessPanel = computed(
  () => !isAdmin.value && (route.path === '/subscriptions' || route.path === '/purchase'),
)

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
