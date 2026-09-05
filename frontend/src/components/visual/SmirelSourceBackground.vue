<template>
  <div
    v-if="enabled"
    class="smirel-source-background"
    :class="{ 'is-ready': ready }"
    :style="{ '--smirel-source-image': `url(&quot;${SOURCE_IMAGE}&quot;)` }"
    aria-hidden="true"
  ></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const SOURCE_IMAGE = 'https://raw.githubusercontent.com/yuchenm1303-png/Chen-s-Homepage/6ffcfa3f9f4203e8a91197fb0a1a34f4efe0b4ea/download/wallpaper-beach-blue-v1-original.png'

const enabled = ref(false)
const ready = ref(false)
let image: HTMLImageElement | null = null

onMounted(() => {
  enabled.value = document.documentElement.classList.contains('relay-standalone')
  if (!enabled.value) return

  image = new Image()
  image.decoding = 'async'
  image.fetchPriority = 'high'
  image.onload = () => {
    ready.value = true
  }
  image.onerror = () => {
    console.warn('[smirel-source-background] wallpaper failed to load')
  }
  image.src = SOURCE_IMAGE
})

onBeforeUnmount(() => {
  if (image) {
    image.onload = null
    image.onerror = null
    image = null
  }
})
</script>
