<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useSession } from '../core/session'

const { state, isAdmin, logout } = useSession()
const open = ref(false)
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)

const initials = computed(() => (state.user?.username || state.user?.email || 'S').slice(0, 1).toUpperCase())
const displayName = computed(() => state.user?.username || state.user?.email?.split('@')[0] || 'Smirel Account')
const secondaryText = computed(() => state.user?.email || (isAdmin.value ? '管理员账户' : 'Smirel 账户'))

function closeMenu(focusTrigger = false) {
  open.value = false
  if (focusTrigger) window.setTimeout(() => trigger.value?.focus(), 0)
}

function toggleMenu() {
  open.value = !open.value
}

function handlePointerDown(event: PointerEvent) {
  if (!open.value) return
  const target = event.target as Node | null
  if (target && !root.value?.contains(target)) closeMenu()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && open.value) closeMenu(true)
}

async function signOut() {
  closeMenu()
  await logout()
}

onMounted(() => {
  document.addEventListener('pointerdown', handlePointerDown)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handlePointerDown)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div ref="root" class="home-account-menu">
    <button
      ref="trigger"
      class="home-account-trigger"
      :class="{ 'is-open': open }"
      type="button"
      aria-haspopup="menu"
      :aria-expanded="open"
      aria-controls="home-account-popover"
      aria-label="账户菜单"
      @click="toggleMenu"
    >
      <span class="home-account-trigger-avatar">{{ initials }}</span>
      <svg class="home-account-trigger-chevron" viewBox="0 0 16 16" aria-hidden="true">
        <path d="m5 6 3 3 3-3" />
      </svg>
    </button>

    <Transition name="home-account-pop">
      <div v-if="open" id="home-account-popover" class="home-account-popover" role="menu">
        <header class="home-account-identity">
          <span class="home-account-avatar">{{ initials }}</span>
          <span class="home-account-copy">
            <strong>{{ displayName }}</strong>
            <small>{{ secondaryText }}</small>
          </span>
          <span class="home-account-role">{{ isAdmin ? 'ADMIN' : 'ACCOUNT' }}</span>
        </header>

        <div class="home-account-divider"></div>

        <nav class="home-account-nav" aria-label="账户快捷入口">
          <RouterLink class="home-account-item" to="/profile" role="menuitem" @click="closeMenu()">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <circle cx="10" cy="7" r="3" />
              <path d="M4.5 16c.8-2.6 2.7-4 5.5-4s4.7 1.4 5.5 4" />
            </svg>
            <span><strong>账户设置</strong><small>资料与账户信息</small></span>
          </RouterLink>

          <RouterLink class="home-account-item" to="/keys" role="menuitem" @click="closeMenu()">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <circle cx="7" cy="10" r="3" />
              <path d="m9.7 8.3 5.8-5.8M12.6 5.4l2 2M14.4 3.6l2 2" />
            </svg>
            <span><strong>API Keys</strong><small>创建与管理密钥</small></span>
          </RouterLink>

          <RouterLink class="home-account-item" to="/usage" role="menuitem" @click="closeMenu()">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="M3 16.5h14M5 14V9.5M10 14V5.5M15 14V8" />
            </svg>
            <span><strong>用量与日志</strong><small>请求、Token 与费用</small></span>
          </RouterLink>
        </nav>

        <div class="home-account-divider"></div>

        <button class="home-account-item home-account-signout" type="button" role="menuitem" @click="signOut">
          <svg viewBox="0 0 20 20" aria-hidden="true">
            <path d="M8.5 4H5.8A1.8 1.8 0 0 0 4 5.8v8.4A1.8 1.8 0 0 0 5.8 16h2.7M12.5 6.5 16 10l-3.5 3.5M8 10h8" />
          </svg>
          <span><strong>退出登录</strong><small>结束当前会话</small></span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.home-account-menu {
  position: relative;
}

.home-account-trigger {
  width: 46px;
  height: 38px;
  padding: 0 7px 0 6px;
  border: 1px solid #2d323b;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  background: #0f1115;
  color: #dce0e5;
  cursor: pointer;
  transition: border-color .16s ease, background-color .16s ease, box-shadow .16s ease;
}

.home-account-trigger:hover,
.home-account-trigger.is-open {
  border-color: #414853;
  background: #14171c;
}

.home-account-trigger.is-open {
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, .018);
}

.home-account-trigger:focus-visible {
  outline: 2px solid rgba(36, 153, 230, .42);
  outline-offset: 2px;
}

.home-account-trigger-avatar {
  width: 25px;
  height: 25px;
  border: 1px solid #353b45;
  border-radius: 7px;
  display: grid;
  place-items: center;
  background: #181b21;
  color: #f1f3f5;
  font-size: .75rem;
  font-weight: 680;
  line-height: 1;
}

.home-account-trigger-chevron {
  width: 12px;
  height: 12px;
  fill: none;
  stroke: #68717d;
  stroke-width: 1.55;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: transform .18s cubic-bezier(.2, .8, .2, 1), stroke .16s ease;
}

.home-account-trigger:hover .home-account-trigger-chevron,
.home-account-trigger.is-open .home-account-trigger-chevron {
  stroke: #a7afb9;
}

.home-account-trigger.is-open .home-account-trigger-chevron {
  transform: rotate(180deg);
}

.home-account-popover {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 80;
  width: min(292px, calc(100vw - 28px));
  padding: 8px;
  border: 1px solid #292f38;
  border-radius: 12px;
  background: #0f1115;
  box-shadow: 0 22px 58px rgba(0, 0, 0, .44), 0 2px 8px rgba(0, 0, 0, .22);
  transform-origin: top right;
}

.home-account-identity {
  min-height: 66px;
  padding: 9px 10px;
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: 11px;
}

.home-account-avatar {
  width: 38px;
  height: 38px;
  border: 1px solid #343a44;
  border-radius: 9px;
  display: grid;
  place-items: center;
  background: #181b21;
  color: #f5f6f8;
  font-size: .88rem;
  font-weight: 700;
}

.home-account-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.home-account-copy strong {
  overflow: hidden;
  color: #eef1f4;
  font-size: .82rem;
  font-weight: 650;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-account-copy small {
  overflow: hidden;
  color: #68717d;
  font-size: .69rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-account-role {
  align-self: start;
  margin-top: 4px;
  color: #58626f;
  font: 700 .58rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .08em;
}

.home-account-divider {
  height: 1px;
  margin: 5px 8px;
  background: #232830;
}

.home-account-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.home-account-item {
  width: 100%;
  min-height: 48px;
  padding: 7px 9px;
  border: 0;
  border-radius: 8px;
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  background: transparent;
  color: #b4bac3;
  cursor: pointer;
  font: inherit;
  text-align: left;
  transition: background-color .14s ease, color .14s ease, transform .14s ease;
}

.home-account-item:hover,
.home-account-item:focus-visible {
  background: #15181e;
  color: #f2f4f6;
  transform: translateX(1px);
}

.home-account-item:focus-visible {
  outline: 1px solid #37404a;
  outline-offset: -1px;
}

.home-account-item > svg {
  width: 17px;
  height: 17px;
  fill: none;
  stroke: #6f7884;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: stroke .14s ease;
}

.home-account-item:hover > svg,
.home-account-item:focus-visible > svg {
  stroke: #aeb7c2;
}

.home-account-item > span {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.home-account-item strong {
  color: inherit;
  font-size: .78rem;
  font-weight: 610;
  line-height: 1.15;
}

.home-account-item small {
  color: #646d78;
  font-size: .66rem;
  line-height: 1.2;
}

.home-account-signout {
  margin-top: 1px;
}

.home-account-signout:hover,
.home-account-signout:focus-visible {
  background: rgba(218, 83, 83, .07);
  color: #f0a0a0;
}

.home-account-signout:hover > svg,
.home-account-signout:focus-visible > svg {
  stroke: #d98282;
}

.home-account-pop-enter-active {
  transition: opacity .15s ease, transform .19s cubic-bezier(.16, 1, .3, 1);
}

.home-account-pop-leave-active {
  transition: opacity .1s ease, transform .12s ease;
}

.home-account-pop-enter-from {
  opacity: 0;
  transform: translateY(-7px) scale(.975);
}

.home-account-pop-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(.988);
}

.home-account-popover .home-account-identity,
.home-account-popover .home-account-item {
  animation: home-account-row-in .22s cubic-bezier(.16, 1, .3, 1) both;
}

.home-account-popover .home-account-item:nth-child(1) { animation-delay: 18ms; }
.home-account-popover .home-account-item:nth-child(2) { animation-delay: 34ms; }
.home-account-popover .home-account-item:nth-child(3) { animation-delay: 50ms; }
.home-account-popover > .home-account-signout { animation-delay: 64ms; }

@keyframes home-account-row-in {
  from {
    opacity: 0;
    transform: translateY(-3px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-account-trigger,
  .home-account-trigger-chevron,
  .home-account-item,
  .home-account-pop-enter-active,
  .home-account-pop-leave-active {
    transition: none !important;
  }

  .home-account-popover .home-account-identity,
  .home-account-popover .home-account-item {
    animation: none !important;
  }
}
</style>
