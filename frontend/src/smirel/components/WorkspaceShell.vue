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
        <button class="mobile-close" type="button" @click="mobileOpen = false">×</button>
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

    <button v-if="mobileOpen" class="workspace-scrim" type="button" @click="mobileOpen = false"></button>

    <section class="workspace-main">
      <header class="workspace-topbar glass">
        <div class="workspace-topbar-left">
          <button class="mobile-menu" type="button" @click="mobileOpen = true"><span></span><span></span><span></span></button>
          <strong class="workspace-topbar-title">{{ isAdmin ? '管理后台' : 'Smirel Console' }}</strong>
        </div>
        <RouterLink to="/profile" class="mini-avatar">{{ initials }}</RouterLink>
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
</style>
