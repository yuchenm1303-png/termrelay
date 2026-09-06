<template>
  <div class="spw-layout spg-page">
    <div class="spg-environment" aria-hidden="true"></div>

    <div class="spw-app-grid">
      <SmirelGlassSidebar />

      <div class="spw-workspace">
        <SmirelGlassTopbar />
        <main class="spw-main">
          <div class="spw-main-inner">
            <RelayAccessPanel v-if="showRelayAccessPanel" />
            <BillingAccessPanel v-if="showBillingAccessPanel" />
            <slot />
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import '@/styles/smirel-shared-glass-v1.css'
import '@/styles/smirel-shared-interactions-v1.css'
import '@/styles/smirel-workspace-v1.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import SmirelGlassSidebar from './SmirelGlassSidebar.vue'
import SmirelGlassTopbar from './SmirelGlassTopbar.vue'
import RelayAccessPanel from '@/components/user/RelayAccessPanel.vue'
import BillingAccessPanel from '@/components/user/BillingAccessPanel.vue'

const route = useRoute()
const authStore = useAuthStore()
const isAdmin = computed(() => authStore.user?.role === 'admin')
const showRelayAccessPanel = computed(() => !isAdmin.value && route.path === '/keys')
const showBillingAccessPanel = computed(
  () => !isAdmin.value && (route.path === '/subscriptions' || route.path === '/purchase'),
)

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true,
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
