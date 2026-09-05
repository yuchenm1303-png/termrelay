<template>
  <div class="relay-app-shell smirel-console-shell min-h-screen">
    <AppSidebar />

    <div
      class="smirel-console-workspace relative min-h-screen transition-[margin] duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <AppHeader />

      <main class="smirel-page-stage">
        <div class="smirel-page-frame">
          <RelayAccessPanel v-if="showRelayAccessPanel" />
          <BillingAccessPanel v-if="showBillingAccessPanel" />
          <slot />
        </div>
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
