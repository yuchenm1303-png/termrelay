<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import WorkspaceNavIcon from './WorkspaceNavIcon.vue'
import { adminNavigation, userNavigation, type NavItem } from '../core/navigation'
import { useSession } from '../core/session'
import '../styles/workspace-layout.css'

interface NavGroup {
  label: string
  items: NavItem[]
}

const route = useRoute()
const mobileOpen = ref(false)
const { state, isAdmin } = useSession()
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const navigation = computed(() => isAdmin.value ? adminNavigation : userNavigation)
const initials = computed(() => (state.user?.username || state.user?.email || 'S').slice(0, 1).toUpperCase())
const accountName = computed(() => state.user?.username || state.user?.email?.split('@')[0] || 'Smirel')
const accountRole = computed(() => isAdmin.value ? '管理员' : '个人账户')

const navIconByFeature: Record<string, string> = {
  dashboard: 'dashboard',
  keys: 'key',
  usage: 'chart',
  subscriptions: 'credit-card',
  purchase: 'wallet',
  orders: 'receipt',
  profile: 'user',
  'admin-dashboard': 'dashboard',
  'admin-users': 'users',
  'admin-accounts': 'server',
  'admin-groups': 'layers',
  'admin-channels': 'network',
  'admin-usage': 'chart',
  'admin-ops': 'activity',
  'admin-payment-dashboard': 'credit-card',
  'admin-orders': 'receipt',
  'admin-settings': 'settings',
}

function take(items: NavItem[], features: string[]) {
  return items.filter((item) => features.includes(item.feature))
}

const navigationGroups = computed<NavGroup[]>(() => {
  const items = navigation.value

  if (!isAdmin.value) {
    return [
      { label: '控制台', items: take(items, ['dashboard', 'keys', 'usage']) },
      { label: '账单', items: take(items, ['subscriptions', 'purchase', 'orders']) },
      { label: '账户', items: take(items, ['profile']) },
    ].filter((group) => group.items.length)
  }

  return [
    { label: '控制台', items: take(items, ['admin-dashboard', 'admin-users']) },
    { label: '资源', items: take(items, ['admin-accounts', 'admin-groups', 'admin-channels']) },
    { label: '运营', items: take(items, ['admin-usage', 'admin-ops']) },
    { label: '交易', items: take(items, ['admin-payment-dashboard', 'admin-orders']) },
    { label: '系统', items: take(items, ['admin-settings']) },
  ].filter((group) => group.items.length)
})

const activeItem = computed(() => navigation.value
  .filter((item) => route.path === item.path || route.path.startsWith(`${item.path}/`))
  .sort((a, b) => b.path.length - a.path.length)[0] || null)

const activeGroupLabel = computed(() => navigationGroups.value.find((group) =>
  group.items.some((item) => item.path === activeItem.value?.path),
)?.label || '')

const breadcrumbGroupLabel = computed(() => (
  activeGroupLabel.value && activeGroupLabel.value !== activeItem.value?.label
    ? activeGroupLabel.value
    : ''
))
</script>

<template>
  <div class="workspace-root">
    <div class="site-environment" aria-hidden="true"></div>
    <aside class="workspace-sidebar glass" :class="{ open: mobileOpen }">
      <div class="workspace-brand-row">
        <RouterLink to="/home" class="brand-link" @click="mobileOpen = false">
          <img :src="logoUrl" alt="Smirel" />
          <span><strong>Smirel</strong><small>API SERVICE</small></span>
        </RouterLink>
        <button class="mobile-close" type="button" aria-label="关闭导航" @click="mobileOpen = false">×</button>
      </div>

      <nav class="workspace-nav">
        <section v-for="group in navigationGroups" :key="group.label" class="workspace-nav-group">
          <div class="workspace-nav-label">{{ group.label }}</div>
          <RouterLink
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            :class="{ active: route.path === item.path }"
            @click="mobileOpen = false"
          >
            <WorkspaceNavIcon :name="navIconByFeature[item.feature] || 'circle'" />
            <span>{{ item.label }}</span>
          </RouterLink>
        </section>
      </nav>
    </aside>

    <button v-if="mobileOpen" class="workspace-scrim" type="button" aria-label="关闭导航" @click="mobileOpen = false"></button>

    <section class="workspace-main">
      <header class="workspace-topbar workspace-topbar--contextual">
        <div class="workspace-topbar-left">
          <button class="mobile-menu" type="button" aria-label="打开导航" @click="mobileOpen = true"><span></span><span></span><span></span></button>
          <nav class="workspace-breadcrumb" aria-label="当前位置">
            <span class="workspace-breadcrumb-root">{{ isAdmin ? '管理后台' : 'Smirel Console' }}</span>
            <span class="workspace-breadcrumb-separator" aria-hidden="true">/</span>
            <template v-if="breadcrumbGroupLabel">
              <span class="workspace-breadcrumb-group">{{ breadcrumbGroupLabel }}</span>
              <span class="workspace-breadcrumb-separator" aria-hidden="true">/</span>
            </template>
            <strong class="workspace-topbar-title">{{ activeItem?.label || (isAdmin ? '控制台' : 'Overview') }}</strong>
          </nav>
        </div>

        <RouterLink to="/profile" class="workspace-profile-link" aria-label="打开账户设置">
          <span class="mini-avatar">{{ initials }}</span>
          <span class="workspace-profile-copy">
            <strong>{{ accountName }}</strong>
            <small>{{ accountRole }}</small>
          </span>
          <svg class="workspace-profile-chevron" viewBox="0 0 16 16" aria-hidden="true">
            <path d="M6 3.5 10.5 8 6 12.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </RouterLink>
      </header>
      <main class="workspace-canvas"><slot /></main>
    </section>
  </div>
</template>

<style scoped>
.workspace-brand-row .brand-link img {
  width: 76px;
  height: auto;
  object-fit: contain;
}

.workspace-nav a {
  gap: 10px;
}

.workspace-nav-icon {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  color: #68717d;
  transition: color .15s ease;
}

.workspace-nav a:hover .workspace-nav-icon {
  color: #aeb6c1;
}

.workspace-nav a.active .workspace-nav-icon {
  color: #73bdf2;
}

.workspace-nav a > span {
  min-width: 0;
}

/*
 * The desktop header is intentionally reduced to an account shortcut. Page
 * headings already provide context, so repeating breadcrumbs here only adds a
 * second visual hierarchy. Mobile keeps the bar because it owns navigation.
 */
.workspace-topbar.workspace-topbar--contextual {
  position: relative;
  top: auto;
  height: 44px;
  padding: 10px 34px 0;
  border-bottom: 0;
  background: transparent;
  align-items: flex-start;
}

.workspace-topbar--contextual .workspace-topbar-left {
  min-width: 0;
}

.workspace-breadcrumb,
.workspace-profile-copy,
.workspace-profile-chevron {
  display: none;
}

.workspace-profile-link {
  min-width: 32px;
  min-height: 32px;
  margin-left: auto;
  padding: 0;
  border: 0;
  border-radius: 8px;
  display: flex;
  align-items: center;
  color: inherit;
  background: transparent;
}

.workspace-profile-link:hover {
  background: #121419;
}

.workspace-profile-link .mini-avatar {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  border-color: #30353e;
  background: #15181e;
}

.workspace-canvas {
  padding-top: 8px;
}

@media (max-width: 1120px) {
  .workspace-topbar.workspace-topbar--contextual {
    padding: 10px 24px 0;
  }
}

@media (max-width: 980px) {
  .workspace-topbar.workspace-topbar--contextual {
    position: sticky;
    top: 0;
    height: 56px;
    padding: 0 14px;
    border-bottom: 1px solid #20232a;
    background: rgba(9, 10, 13, .985);
    align-items: center;
  }

  .workspace-topbar--contextual .workspace-topbar-left {
    display: flex;
    align-items: center;
  }

  .workspace-canvas {
    padding-top: 24px;
  }
}
</style>
