<template>
  <header class="smirel-console-header sticky top-0 z-30 border-b">
    <div class="smirel-console-header-inner flex items-center justify-between gap-4">
      <div class="flex min-w-0 items-center gap-3">
        <button
          type="button"
          class="smirel-header-icon-button lg:hidden"
          :aria-label="t('common.toggleMenu')"
          @click="toggleMobileSidebar"
        >
          <Icon name="menu" size="md" />
        </button>

        <div class="min-w-0">
          <div class="mb-0.5 hidden items-center gap-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] sm:flex">
            <span class="smirel-header-context">{{ consoleLabel }}</span>
            <span class="smirel-header-separator">/</span>
            <span class="smirel-header-section">{{ sectionLabel }}</span>
          </div>
          <div class="flex min-w-0 items-baseline gap-3">
            <h1 class="truncate text-base font-semibold sm:text-[17px]">{{ pageTitle }}</h1>
            <p v-if="pageDescription" class="hidden max-w-[520px] truncate text-xs xl:block">
              {{ pageDescription }}
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-shrink-0 items-center gap-1.5 sm:gap-2">
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="smirel-header-action hidden md:inline-flex"
        >
          <Icon name="book" size="sm" />
          <span>{{ t('nav.docs') }}</span>
        </a>

        <router-link
          v-if="modelPlazaEnabled"
          to="/model-plaza?embedded=1"
          class="smirel-header-action hidden lg:inline-flex"
        >
          <Icon name="grid" size="sm" />
          <span>{{ t('nav.modelPlaza') }}</span>
        </router-link>

        <AnnouncementBell v-if="user" />
        <LocaleSwitcher />

        <div v-if="user" class="smirel-balance-pill hidden sm:flex">
          <span class="smirel-balance-dot"></span>
          <span class="text-[11px] font-medium">{{ balanceLabel }}</span>
          <span class="font-mono text-xs font-semibold tabular-nums">{{ formatMoney(availableBalance) }}</span>
        </div>

        <div v-if="user" ref="dropdownRef" class="relative">
          <button
            type="button"
            class="smirel-account-trigger"
            :aria-expanded="dropdownOpen"
            :aria-label="t('common.userMenu')"
            @click.stop="dropdownOpen = !dropdownOpen"
          >
            <span class="smirel-account-avatar">
              <img v-if="avatarUrl" :src="avatarUrl" :alt="displayName" class="h-full w-full object-cover" />
              <span v-else>{{ userInitials }}</span>
            </span>
            <span class="hidden min-w-0 text-left md:block">
              <span class="block max-w-28 truncate text-xs font-semibold">{{ displayName }}</span>
              <span class="smirel-account-role block text-[10px]">{{ roleLabel }}</span>
            </span>
            <Icon name="chevronDown" size="xs" class="hidden md:block" />
          </button>

          <transition name="smirel-menu">
            <div v-if="dropdownOpen" class="smirel-account-menu absolute right-0 top-full mt-2 w-64 overflow-hidden">
              <div class="smirel-account-menu-head">
                <p class="truncate text-sm font-semibold">{{ displayName }}</p>
                <p class="mt-0.5 truncate text-xs">{{ user.email }}</p>
              </div>

              <div class="smirel-account-menu-balance sm:hidden">
                <span>{{ balanceLabel }}</span>
                <strong>{{ formatMoney(availableBalance) }}</strong>
              </div>

              <div class="p-1.5">
                <router-link to="/profile" class="smirel-account-menu-item" @click="closeDropdown">
                  <Icon name="user" size="sm" />
                  <span>{{ t('nav.profile') }}</span>
                </router-link>
                <router-link to="/keys" class="smirel-account-menu-item" @click="closeDropdown">
                  <Icon name="key" size="sm" />
                  <span>{{ t('nav.apiKeys') }}</span>
                </router-link>
              </div>

              <div v-if="contactInfo" class="smirel-account-support">
                <span>{{ supportLabel }}</span>
                <strong>{{ contactInfo }}</strong>
              </div>

              <div class="border-t p-1.5">
                <button type="button" class="smirel-account-menu-item smirel-account-menu-danger w-full" @click="handleLogout">
                  <Icon name="arrowLeft" size="sm" />
                  <span>{{ t('nav.logout') }}</span>
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAppStore, useAuthStore } from '@/stores'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { sanitizeUrl } from '@/utils/url'

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()

const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const docUrl = computed(() => sanitizeUrl(appStore.docUrl || ''))
const contactInfo = computed(() => appStore.contactInfo)
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))
const availableBalance = computed(() => Number(user.value?.balance || 0))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')

const consoleLabel = computed(() =>
  isZh.value
    ? isAdmin.value ? 'Smirel 运营控制台' : 'Smirel 开发者控制台'
    : isAdmin.value ? 'Smirel Operations' : 'Smirel Developer',
)

const balanceLabel = computed(() => (isZh.value ? '余额' : 'Balance'))
const supportLabel = computed(() => (isZh.value ? '支持' : 'Support'))
const roleLabel = computed(() =>
  isZh.value
    ? isAdmin.value ? '管理员' : '开发者'
    : isAdmin.value ? 'Administrator' : 'Developer',
)

const userInitials = computed(() => {
  const value = user.value?.username || user.value?.email?.split('@')[0] || 'S'
  return value.slice(0, 2).toUpperCase()
})

const displayName = computed(() =>
  user.value?.username || user.value?.email?.split('@')[0] || 'Smirel User',
)

const pageTitle = computed(() => {
  if (route.name === 'CustomPage') {
    const id = route.params.id as string
    const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
    const item = publicItems.find((entry) => entry.id === id)
      ?? (isAdmin.value ? adminSettingsStore.customMenuItems.find((entry) => entry.id === id) : undefined)
    if (item?.label) return item.label
  }
  const titleKey = route.meta.titleKey as string | undefined
  if (titleKey) return t(titleKey)
  return (route.meta.title as string | undefined) || (isZh.value ? '控制台' : 'Console')
})

const pageDescription = computed(() => {
  const descriptionKey = route.meta.descriptionKey as string | undefined
  if (descriptionKey) return t(descriptionKey)
  return (route.meta.description as string | undefined) || ''
})

const sectionLabel = computed(() => {
  const path = route.path
  if (isAdmin.value) {
    if (/^\/admin\/(accounts|channels|proxies)/.test(path)) return isZh.value ? '上游与路由' : 'Upstream & Routing'
    if (/^\/admin\/(users|groups)/.test(path)) return isZh.value ? '用户与权限' : 'Users & Access'
    if (/^\/admin\/(subscriptions|orders|redeem|promo-codes|affiliates)/.test(path)) return isZh.value ? '商业化' : 'Commerce'
    if (/^\/admin\/(usage|audit-logs|risk-control|prompt-audit|announcements|settings)/.test(path)) return isZh.value ? '平台治理' : 'Governance'
    return isZh.value ? '运营总览' : 'Operations'
  }

  if (/^\/(keys|usage|available-channels|monitor|model-plaza|batch-image)/.test(path)) return isZh.value ? 'API 与模型' : 'API & Models'
  if (/^\/(subscriptions|purchase|orders|redeem|affiliate|profile)/.test(path)) return isZh.value ? '计费与账户' : 'Billing & Account'
  return isZh.value ? '概览' : 'Overview'
})

function formatMoney(value: number) {
  return Number.isFinite(value) ? `$${value.toFixed(2)}` : '$0.00'
}

function toggleMobileSidebar() {
  appStore.toggleMobileSidebar()
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) closeDropdown()
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.smirel-menu-enter-active,
.smirel-menu-leave-active {
  transition: opacity 140ms ease, transform 140ms ease;
}

.smirel-menu-enter-from,
.smirel-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
