<template>
  <component
    :is="as"
    class="smirel-card smg-surface"
    :class="[
      `smirel-card--${tone}`,
      `smg-surface--${tone}`,
      { 'smirel-card--interactive': interactive },
    ]"
  >
    <slot />
  </component>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    as?: string
    tone?: 'glass' | 'data' | 'quiet'
    interactive?: boolean
  }>(),
  {
    as: 'section',
    tone: 'glass',
    interactive: false,
  },
)
</script>
