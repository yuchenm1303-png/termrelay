<template>
  <component :is="as" class="smg-surface" :class="[`smg-surface--${tone}`]">
    <slot />
  </component>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    as?: string
    tone?: 'glass' | 'data' | 'quiet'
  }>(),
  {
    as: 'section',
    tone: 'glass',
  },
)
</script>
