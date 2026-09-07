<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import HomeAccountMenu from './smirel/components/HomeAccountMenu.vue'
import WorkspaceShell from './smirel/components/WorkspaceShell.vue'

const route = useRoute()
const useWorkspace = computed(() => route.meta.shell === 'workspace')
</script>

<template>
  <RouterView v-slot="{ Component }">
    <template v-if="useWorkspace">
      <WorkspaceShell>
        <component :is="Component" />
      </WorkspaceShell>
      <HomeAccountMenu variant="workspace" />
    </template>
    <component :is="Component" v-else />
  </RouterView>
</template>
