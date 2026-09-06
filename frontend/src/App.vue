<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import WorkspaceShell from './smirel/components/WorkspaceShell.vue'

const route = useRoute()
const useWorkspace = computed(() => route.meta.shell === 'workspace')
</script>

<template>
  <RouterView v-slot="{ Component }">
    <WorkspaceShell v-if="useWorkspace">
      <component :is="Component" />
    </WorkspaceShell>
    <component :is="Component" v-else />
  </RouterView>
</template>
