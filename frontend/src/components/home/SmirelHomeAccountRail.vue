<template>
  <div class="smh-account-rail-host">
    <section class="smh-account-rail" aria-label="账户与常用功能">
      <div class="smh-account-rail-identity">
        <span class="smh-account-rail-avatar" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65">
            <circle cx="12" cy="8" r="3.2" />
            <path stroke-linecap="round" d="M5.8 19c.7-3.4 3-5.2 6.2-5.2s5.5 1.8 6.2 5.2" />
          </svg>
        </span>

        <span class="smh-account-rail-copy">
          <small>{{ isAuthenticated ? roleLabel : 'ACCOUNT' }}</small>
          <strong>{{ accountTitle }}</strong>
        </span>

        <span class="smh-account-rail-state" :class="{ 'is-online': isAuthenticated }">
          <i aria-hidden="true"></i>{{ isAuthenticated ? '已登录' : '未登录' }}
        </span>
      </div>

      <nav class="smh-account-rail-actions" aria-label="账户快捷入口">
        <template v-if="!isAuthenticated">
          <router-link to="/login" class="smh-account-rail-action smh-account-rail-action--primary">
            <span class="smh-account-rail-action-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M14 8l4 4-4 4M18 12H8m2-7H6.5A2.5 2.5 0 004 7.5v9A2.5 2.5 0 006.5 19H10" /></svg>
            </span>
            <span><small>ACCOUNT</small><strong>登录</strong></span>
          </router-link>

          <router-link to="/register" class="smh-account-rail-action">
            <span class="smh-account-rail-action-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19a6 6 0 00-12 0m6-8a4 4 0 100-8 4 4 0 000 8zm8-2v6m3-3h-6" /></svg>
            </span>
            <span><small>NEW USER</small><strong>注册账户</strong></span>
          </router-link>
        </template>

        <template v-else>
          <router-link :to="consolePath" class="smh-account-rail-action smh-account-rail-action--primary">
            <span class="smh-account-rail-action-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M4 5h16v14H4V5zm0 4h16M8 15h3" /></svg>
            </span>
            <span><small>WORKSPACE</small><strong>{{ authStore.isAdmin ? '管理控制台' : '进入控制台' }}</strong></span>
          </router-link>

          <router-link to="/profile" class="smh-account-rail-action">
            <span class="smh-account-rail-action-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M12 14a4 4 0 100-8 4 4 0 000 8zm-7 6c.9-3.5 3.3-5.3 7-5.3s6.1 1.8 7 5.3" /></svg>
            </span>
            <span><small>PROFILE</small><strong>账户管理</strong></span>
          </router-link>
        </template>

        <router-link :to="keyPath" class="smh-account-rail-action">
          <span class="smh-account-rail-action-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M14.5 9.5a4.5 4.5 0 11-1.3 3.2L21 5v4h-3v3h-3l-1.8 1.8" /></svg>
          </span>
          <span><small>API KEY</small><strong>密钥管理</strong></span>
        </router-link>

        <router-link :to="usagePath" class="smh-account-rail-action">
          <span class="smh-account-rail-action-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path stroke-linecap="round" stroke-linejoin="round" d="M5 19V9m4 10V5m5 14v-7m5 7V7" /></svg>
          </span>
          <span><small>USAGE</small><strong>{{ isAuthenticated ? '用量与日志' : '用量查询' }}</strong></span>
        </router-link>
      </nav>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores'

const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const consolePath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const keyPath = computed(() => isAuthenticated.value ? '/keys' : '/login')
const usagePath = computed(() => isAuthenticated.value ? '/usage' : '/key-usage')
const roleLabel = computed(() => authStore.isAdmin ? 'ADMIN ACCOUNT' : 'SMIREL ACCOUNT')
const accountTitle = computed(() => {
  if (!isAuthenticated.value) return '登录后管理你的 API 工作区'
  const user = authStore.user
  return user?.username?.trim() || user?.email?.trim() || 'Smirel Account'
})
</script>

<style>
.smh-home-composition {
  position: relative;
  min-height: 100vh;
}

.smh-home-composition .smh-home main {
  padding-top: 96px;
}

.smh-account-rail-host {
  position: absolute;
  z-index: 4;
  top: 94px;
  left: 50%;
  width: min(1480px, calc(100vw - 72px));
  transform: translateX(-50%);
  pointer-events: none;
}

.smh-account-rail {
  min-height: 76px;
  display: grid;
  grid-template-columns: minmax(280px, .86fr) minmax(0, 2.14fr);
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, .10);
  border-radius: 8px;
  background: rgba(6, 18, 29, .48);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, .06), 0 18px 46px rgba(2, 9, 15, .10);
  -webkit-backdrop-filter: blur(24px) saturate(122%);
  backdrop-filter: blur(24px) saturate(122%);
  pointer-events: auto;
}

.smh-account-rail-identity {
  min-width: 0;
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: 13px;
  padding: 13px 18px;
  border-right: 1px solid rgba(255, 255, 255, .08);
}

.smh-account-rail-avatar {
  width: 38px;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, .10);
  border-radius: 7px;
  background: rgba(255, 255, 255, .07);
  color: rgba(255, 255, 255, .78);
}

.smh-account-rail-avatar svg,
.smh-account-rail-action-icon svg {
  width: 19px;
  height: 19px;
}

.smh-account-rail-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.smh-account-rail-copy small,
.smh-account-rail-action small {
  color: rgba(255, 255, 255, .43);
  font-size: .61rem;
  line-height: 1;
  letter-spacing: .11em;
}

.smh-account-rail-copy strong {
  overflow: hidden;
  color: rgba(255, 255, 255, .92);
  font-size: .84rem;
  font-weight: 640;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smh-account-rail-state {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, .52);
  font-size: .68rem;
  white-space: nowrap;
}

.smh-account-rail-state i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: rgba(255, 255, 255, .34);
}

.smh-account-rail-state.is-online i {
  background: #83efc7;
  box-shadow: 0 0 10px rgba(131, 239, 199, .40);
}

.smh-account-rail-actions {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.smh-account-rail-action {
  min-width: 0;
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  padding: 13px 16px;
  color: rgba(255, 255, 255, .76);
  text-decoration: none;
  transition: background-color .18s ease, color .18s ease;
}

.smh-account-rail-action + .smh-account-rail-action {
  border-left: 1px solid rgba(255, 255, 255, .07);
}

.smh-account-rail-action:hover {
  color: rgba(255, 255, 255, .98);
  background: rgba(255, 255, 255, .07);
}

.smh-account-rail-action--primary {
  background: rgba(255, 255, 255, .055);
}

.smh-account-rail-action-icon {
  width: 30px;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, .62);
}

.smh-account-rail-action > span:last-child {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.smh-account-rail-action strong {
  overflow: hidden;
  font-size: .79rem;
  font-weight: 620;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1180px) {
  .smh-account-rail-host {
    width: calc(100vw - 48px);
  }

  .smh-account-rail {
    grid-template-columns: minmax(240px, .74fr) minmax(0, 2.26fr);
  }

  .smh-account-rail-state {
    display: none;
  }
}

@media (max-width: 980px) {
  .smh-home-composition .smh-home main {
    padding-top: 154px;
  }

  .smh-account-rail {
    grid-template-columns: 1fr;
  }

  .smh-account-rail-identity {
    min-height: 58px;
    border-right: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .08);
  }

  .smh-account-rail-actions {
    grid-template-columns: repeat(4, minmax(140px, 1fr));
    overflow-x: auto;
  }
}

@media (max-width: 720px) {
  .smh-home-composition .smh-home main {
    padding-top: 166px;
  }

  .smh-account-rail-host {
    top: 84px;
    width: calc(100vw - 28px);
  }

  .smh-account-rail-identity {
    grid-template-columns: 34px minmax(0, 1fr);
    padding: 12px 14px;
  }

  .smh-account-rail-avatar {
    width: 34px;
    height: 34px;
  }

  .smh-account-rail-state {
    display: none;
  }

  .smh-account-rail-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    overflow: visible;
  }

  .smh-account-rail-action {
    min-height: 58px;
    padding: 10px 12px;
  }

  .smh-account-rail-action:nth-child(3) {
    border-left: 0;
    border-top: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-account-rail-action:nth-child(4) {
    border-top: 1px solid rgba(255, 255, 255, .07);
  }
}

@media (prefers-reduced-motion: reduce) {
  .smh-account-rail-action {
    transition: none !important;
  }
}
</style>
