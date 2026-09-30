<script setup lang="ts">
import { inject, ref, watch } from 'vue'
import { workspaceRouteSettledKey } from '../core/route-motion'

const props = withDefaults(defineProps<{ ready: boolean; contentMotion?: boolean }>(), { contentMotion: true })
const routeSettled = inject(workspaceRouteSettledKey, ref(true))
const resolved = ref(false)

// Latch after first success. Later refreshes preserve the visible cards.
watch([() => props.ready, routeSettled], ([dataReady, routeReady]) => {
  if (dataReady && routeReady) resolved.value = true
}, { immediate: true })
</script>

<template>
  <div class="async-data-boundary" :data-resolved="resolved">
    <Transition name="async-data" mode="out-in" appear>
      <div v-if="resolved" key="ready" class="async-data-content"
        :class="{ 'async-data-content--instant': !contentMotion }"><slot /></div>
      <div v-else key="pending" class="async-data-placeholder" aria-busy="true"><slot name="loading" /></div>
    </Transition>
  </div>
</template>
