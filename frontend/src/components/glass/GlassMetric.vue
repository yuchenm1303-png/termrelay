<template>
  <GlassSurface as="article" class="smg-metric">
    <span class="smg-metric-label">{{ label }}</span>
    <strong class="smg-metric-value">{{ value }}</strong>
    <small class="smg-metric-note">{{ note }}</small>
  </GlassSurface>
</template>

<script setup lang="ts">
import GlassSurface from './GlassSurface.vue'

defineProps<{
  label: string
  value: string
  note: string
}>()
</script>
