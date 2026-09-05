<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <div v-if="$slots.filters || $slots.actions" class="smirel-table-toolbar">
      <div v-if="$slots.filters" class="smirel-table-filters">
        <slot name="filters" />
      </div>
      <div v-if="$slots.actions" class="smirel-table-actions">
        <slot name="actions" />
      </div>
    </div>

    <div class="layout-section-scrollable">
      <div class="smirel-data-panel table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <div v-if="$slots.pagination" class="smirel-table-pagination">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.table-page-layout {
  @apply flex min-h-0 flex-col gap-4;
  min-height: calc(100vh - 64px - 4rem);
}

.smirel-table-toolbar {
  @apply flex flex-shrink-0 flex-col gap-4 lg:flex-row lg:items-center lg:justify-between;
}

.smirel-table-filters {
  @apply min-w-0 flex-1;
}

.smirel-table-actions {
  @apply flex flex-shrink-0 items-center lg:justify-end;
}

.layout-section-scrollable {
  @apply flex min-h-0 flex-1 flex-col;
}

.table-scroll-container {
  @apply flex h-full min-h-[360px] flex-col overflow-hidden;
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content;
  display: table;
}

.table-scroll-container :deep(thead) {
  @apply sticky top-0 z-10;
}

.table-scroll-container :deep(th) {
  @apply px-5 py-3.5 text-left text-xs font-semibold uppercase tracking-wide;
}

.table-scroll-container :deep(td) {
  @apply px-5 py-4 text-sm;
}

.smirel-table-pagination {
  @apply flex-shrink-0;
}

.table-page-layout.mobile-mode {
  min-height: auto;
}

.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto min-h-0 overflow-visible;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply min-h-fit flex-none;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
