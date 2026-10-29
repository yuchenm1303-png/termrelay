<script setup lang="ts">
import { computed, provide, ref, shallowRef, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import HomeAccountMenu from './smirel/components/HomeAccountMenu.vue'
import WorkspaceShell from './smirel/components/WorkspaceShell.vue'
import {
  resolveRouteMotion,
  workspaceRouteSettledKey,
  type RouteMotion,
  type RouteMotionEndpoint,
} from './smirel/core/route-motion'

const route = useRoute()
const useWorkspace = computed(() => route.meta.shell === 'workspace')
const workspaceRouteSettled = ref(!useWorkspace.value)
provide(workspaceRouteSettledKey, workspaceRouteSettled)
const shell = ref<InstanceType<typeof WorkspaceShell> | null>(null)
const routeMotion = shallowRef<RouteMotion>({ kind: 'section', direction: 1 })

// Shell entrance on first visit and route entrance during internal navigation
// report separately. This prevents fetching data from racing CSS choreography.
watch(() => route.path, (next, previous) => {
  if (next !== previous) workspaceRouteSettled.value = !useWorkspace.value
}, { flush: 'sync' })
function markWorkspaceEntered() { workspaceRouteSettled.value = true }

// Resolved synchronously so the leaving and entering pages share one motion.
let previousWorkspaceRoute: RouteMotionEndpoint | null = null
watch(() => route.path, () => {
  const current = useWorkspace.value ? { path: route.path, feature: route.meta.feature } : null
  if (current) routeMotion.value = resolveRouteMotion(current, previousWorkspaceRoute)
  previousWorkspaceRoute = current
}, { flush: 'sync', immediate: true })

function onWorkspaceBeforeEnter() { shell.value?.resetScroll() }
</script>

<template>
  <RouterView v-slot="{ Component }">
    <template v-if="useWorkspace">
      <WorkspaceShell ref="shell" @canvas-entered="markWorkspaceEntered">
        <div class="workspace-route-motion" :data-route-motion="routeMotion.kind"
          :data-route-direction="routeMotion.direction > 0 ? 'forward' : 'back'">
          <Transition name="workspace-route" mode="out-in"
            @before-enter="onWorkspaceBeforeEnter" @after-enter="markWorkspaceEntered">
            <div :key="route.path" class="workspace-route-stage" :class="{ 'workspace-route-stage--keys': route.path === '/keys' }">
              <component :is="Component" />
            </div>
          </Transition>
        </div>
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
