<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminNavigation, userNavigation } from '../core/navigation'
import { useSession } from '../core/session'

const route = useRoute()
const router = useRouter()
const mobileOpen = ref(false)
const loggingOut = ref(false)
const { state, isAdmin, logout } = useSession()
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const navigation = computed(() => isAdmin.value ? adminNavigation : userNavigation)
const initials = computed(() => (state.user?.username || state.user?.email || 'S').slice(0, 1).toUpperCase())
const accountName = computed(() => state.user?.username || state.user?.email?.split('@')[0] || 'Smirel Account')

async function signOut() {
  loggingOut.value = true
  try {
    await logout()
    if (!import.meta.env.VITE_UI_PREVIEW) await router.replace('/login')
  } finally {
    loggingOut.value = false
  }
}
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

      <div class="workspace-context">
        <span>{{ isAdmin ? 'ADMIN CONSOLE' : 'WORKSPACE' }}</span>
        <strong>{{ isAdmin ? '平台管理' : 'Smirel API' }}</strong>
        <small>{{ navigation.length }} 个功能模块</small>
      </div>

      <nav class="workspace-nav">
        <RouterLink
          v-for="item in navigation"
          :key="item.path"
          :to="item.path"
          :class="{ active: route.path === item.path }"
          @click="mobileOpen = false"
        >
          <i>{{ item.short }}</i><span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="workspace-account">
        <RouterLink to="/profile" class="account-row" @click="mobileOpen = false">
          <b>{{ initials }}</b>
          <span><strong>{{ accountName }}</strong><small>{{ isAdmin ? '管理员账户' : state.user?.email }}</small></span>
        </RouterLink>
        <div class="account-actions">
          <RouterLink to="/home">返回首页</RouterLink>
          <button type="button" :disabled="loggingOut" @click="signOut">{{ loggingOut ? '退出中…' : '退出登录' }}</button>
        </div>
      </div>
    </aside>

    <button v-if="mobileOpen" class="workspace-scrim" type="button" @click="mobileOpen = false"></button>

    <section class="workspace-main">
      <header class="workspace-topbar glass">
        <div class="workspace-topbar-left">
          <button class="mobile-menu" type="button" aria-label="打开导航" @click="mobileOpen = true"><span></span><span></span><span></span></button>
          <div class="workspace-topbar-copy">
            <span>{{ isAdmin ? 'ADMIN CONSOLE' : 'WORKSPACE' }}</span>
            <strong>{{ route.meta.title }}</strong>
          </div>
        </div>
        <div class="workspace-topbar-right">
          <span class="endpoint-pill"><i>API</i><b>api.smirel.com/v1</b></span>
          <span v-if="isAdmin" class="admin-pill">ADMIN</span>
          <RouterLink to="/profile" class="mini-avatar" aria-label="账户设置">{{ initials }}</RouterLink>
        </div>
      </header>
      <main class="workspace-canvas"><slot /></main>
    </section>
  </div>
</template>
