<script setup lang="ts">
defineProps<{ name: string }>()
</script>

<template>
  <svg
    class="workspace-nav-icon"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.8"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'dashboard'">
      <rect width="7" height="9" x="3" y="3" rx="1" />
      <rect width="7" height="5" x="14" y="3" rx="1" />
      <rect width="7" height="9" x="14" y="12" rx="1" />
      <rect width="7" height="5" x="3" y="16" rx="1" />
    </template>

    <template v-else-if="name === 'users'">
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </template>

    <template v-else-if="name === 'server'">
      <rect width="20" height="8" x="2" y="2" rx="2" ry="2" />
      <rect width="20" height="8" x="2" y="14" rx="2" ry="2" />
      <line x1="6" x2="6.01" y1="6" y2="6" />
      <line x1="6" x2="6.01" y1="18" y2="18" />
    </template>

    <template v-else-if="name === 'layers'">
      <path d="m12 2-10 5 10 5 10-5-10-5Z" />
      <path d="m2 12 10 5 10-5" />
      <path d="m2 17 10 5 10-5" />
    </template>

    <template v-else-if="name === 'network'">
      <rect width="6" height="6" x="9" y="2" rx="1" />
      <rect width="6" height="6" x="2" y="16" rx="1" />
      <rect width="6" height="6" x="16" y="16" rx="1" />
      <path d="M12 8v4" />
      <path d="M5 16v-2a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v2" />
    </template>

    <template v-else-if="name === 'chart'">
      <path d="M3 3v18h18" />
      <path d="M7 16v-3" />
      <path d="M12 16V9" />
      <path d="M17 16V5" />
    </template>

    <template v-else-if="name === 'activity'">
      <path d="M3 12h4l3-8 4 16 3-8h4" />
    </template>

    <template v-else-if="name === 'credit-card'">
      <rect width="20" height="14" x="2" y="5" rx="2" />
      <line x1="2" x2="22" y1="10" y2="10" />
      <line x1="6" x2="10" y1="15" y2="15" />
    </template>

    <template v-else-if="name === 'receipt'">
      <path d="M4 2v20l2-2 2 2 2-2 2 2 2-2 2 2 2-2 2 2V2l-2 2-2-2-2 2-2-2-2 2-2-2-2 2-2-2Z" />
      <path d="M16 8h-6" />
      <path d="M16 12h-6" />
      <path d="M13 16h-3" />
    </template>

    <template v-else-if="name === 'settings'">
      <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.09a2 2 0 0 1 1 1.74v.5a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.09a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2Z" />
      <circle cx="12" cy="12" r="3" />
    </template>

    <template v-else-if="name === 'key'">
      <circle cx="7.5" cy="15.5" r="5.5" />
      <path d="m21 2-9.6 9.6" />
      <path d="m15.5 7.5 3 3L22 7" />
    </template>

    <template v-else-if="name === 'wallet'">
      <path d="M20 7V5a2 2 0 0 0-2-2H5a3 3 0 0 0 0 6h15a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2H5a3 3 0 0 1-3-3V6" />
      <path d="M16 14h2" />
    </template>

    <template v-else-if="name === 'user'">
      <circle cx="12" cy="12" r="10" />
      <circle cx="12" cy="10" r="3" />
      <path d="M7 20.66V19a5 5 0 0 1 10 0v1.66" />
    </template>

    <template v-else-if="name === 'arrow-up-right'">
      <path d="M7 17 17 7" />
      <path d="M9 7h8v8" />
    </template>

    <circle v-else cx="12" cy="12" r="4" />
  </svg>
</template>
