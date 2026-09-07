<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
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

/*
 * Secondary-page header: keep this component contextual instead of decorative.
 * The page heading below owns the large title; the topbar only communicates
 * location and account identity, so it stays useful without competing with it.
 */
.workspace-topbar.workspace-topbar--contextual {
  height: 64px;
  padding: 0 36px;
  border-bottom: 1px solid #20232a;
  background: rgba(9, 10, 13, .985);
}

.workspace-topbar--contextual .workspace-topbar-left {
  min-width: 0;
  gap: 14px;
}

.workspace-breadcrumb {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 9px;
  white-space: nowrap;
}

.workspace-breadcrumb-root,
.workspace-breadcrumb-group {
  color: #707783;
  font-size: .73rem;
  font-weight: 560;
  letter-spacing: .01em;
}

.workspace-breadcrumb-group {
  color: #8b929c;
}

.workspace-breadcrumb-separator {
  color: #3d424a;
  font-size: .70rem;
  user-select: none;
}

.workspace-topbar--contextual .workspace-topbar-title {
  min-width: 0;
  overflow: hidden;
  color: #e9ecf0;
  font-size: .82rem;
  font-weight: 640;
  letter-spacing: -.005em;
  text-overflow: ellipsis;
}

.workspace-profile-link {
  min-width: 0;
  min-height: 42px;
  padding: 4px 8px 4px 5px;
  border: 1px solid transparent;
  border-radius: 9px;
  display: flex;
  align-items: center;
  gap: 9px;
  color: inherit;
  transition: background-color .15s ease, border-color .15s ease;
}

.workspace-profile-link:hover {
  border-color: #282c34;
  background: #121419;
}

.workspace-profile-link .mini-avatar {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  border-color: #343943;
  background: #181b21;
}

.workspace-profile-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.workspace-profile-copy strong {
  max-width: 148px;
  overflow: hidden;
  color: #d9dde2;
  font-size: .74rem;
  line-height: 1.15;
  font-weight: 620;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workspace-profile-copy small {
  color: #6f7781;
  font-size: .64rem;
  line-height: 1.1;
  font-weight: 540;
}

.workspace-profile-chevron {
  width: 14px;
  height: 14px;
  flex: 0 0 14px;
  color: #555d68;
}

@media (max-width: 1120px) {
  .workspace-topbar.workspace-topbar--contextual {
    padding: 0 24px;
  }
}

@media (max-width: 720px) {
  .workspace-topbar.workspace-topbar--contextual {
    height: 56px;
    padding: 0 14px;
  }

  .workspace-breadcrumb-root,
  .workspace-breadcrumb-group,
  .workspace-breadcrumb-separator,
  .workspace-profile-copy,
  .workspace-profile-chevron {
    display: none;
  }

  .workspace-profile-link {
    min-height: 36px;
    padding: 2px;
  }
}
</style>
