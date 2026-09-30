<script setup lang="ts">
import { computed, provide, ref, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import HomeAccountMenu from './smirel/components/HomeAccountMenu.vue'
import WorkspaceShell from './smirel/components/WorkspaceShell.vue'
import { workspaceRouteSettledKey } from './smirel/core/route-motion'

const route = useRoute()
const useWorkspace = computed(() => route.meta.shell === 'workspace')
const workspaceRouteSettled = ref(!useWorkspace.value)
provide(workspaceRouteSettledKey, workspaceRouteSettled)

// Shell entrance on first visit and route entrance during internal navigation
// report separately. This prevents fetching data from racing CSS choreography.
watch(() => route.path, (next, previous) => {
  if (next !== previous) workspaceRouteSettled.value = !useWorkspace.value
}, { flush: 'sync' })
function markWorkspaceEntered() { workspaceRouteSettled.value = true }
</script>

<template>
  <RouterView v-slot="{ Component }">
    <template v-if="useWorkspace">
      <WorkspaceShell @canvas-entered="markWorkspaceEntered">
        <Transition name="workspace-route" mode="out-in" @after-enter="markWorkspaceEntered">
          <div :key="route.path" class="workspace-route-stage" :class="{ 'workspace-route-stage--keys': route.path === '/keys' }">
            <component :is="Component" />
          </div>
        </Transition>
      </WorkspaceShell>
      <Teleport to=".workspace-topbar-actions">
        <HomeAccountMenu variant="workspace" />
      </Teleport>
    </template>
    <Transition v-else name="app-route" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </RouterView>
</template>
