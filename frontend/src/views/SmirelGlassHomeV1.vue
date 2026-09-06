<template>
  <div class="spg-page smh-home">
    <div class="spg-environment smh-environment" aria-hidden="true"></div>

    <div class="smh-shell">
      <header class="smh-header">
        <router-link to="/home" class="spg-brand" aria-label="Smirel 首页">
          <img v-if="siteLogo" :src="siteLogo" alt="" class="spg-brand-logo" />
          <span v-else class="spg-brand-fallback">S</span>
          <span class="spg-brand-copy">
            <strong>{{ siteName }}</strong>
            <small>API SERVICE</small>
          </span>
        </router-link>

        <div class="smh-header-right">
          <nav class="smh-nav" aria-label="首页导航">
            <router-link to="/model-plaza">模型与价格</router-link>
            <router-link to="/key-usage">用量查询</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">接入文档</a>
          </nav>

          <span class="smh-status"><i></i>服务正常</span>

          <div v-if="isAuthenticated" ref="accountMenuRoot" class="smh-account-area">
            <router-link :to="accountPath" class="smh-console-link">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true">
                <rect x="4" y="4" width="6" height="6" rx="1.2" />
                <rect x="14" y="4" width="6" height="6" rx="1.2" />
                <rect x="4" y="14" width="6" height="6" rx="1.2" />
                <rect x="14" y="14" width="6" height="6" rx="1.2" />
              </svg>
              <span>{{ authStore.isAdmin ? '管理控制台' : '控制台' }}</span>
            </router-link>

            <button
              type="button"
              class="smh-account-trigger"
              aria-label="打开账户菜单"
              aria-haspopup="menu"
              :aria-expanded="accountMenuOpen"
              @click="accountMenuOpen = !accountMenuOpen"
            >
              <img v-if="avatarUrl" :src="avatarUrl" alt="" />
              <span v-else>{{ accountInitial }}</span>
              <svg class="smh-account-chevron" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6.5 8l3.5 3.5L13.5 8" />
              </svg>
            </button>

            <div v-if="accountMenuOpen" class="smh-account-menu" role="menu">
              <div class="smh-account-summary">
                <div class="smh-account-avatar smh-account-avatar--large">
                  <img v-if="avatarUrl" :src="avatarUrl" alt="" />
                  <span v-else>{{ accountInitial }}</span>
                </div>
                <div class="smh-account-summary-copy">
                  <strong>{{ accountLabel }}</strong>
                  <span v-if="accountEmail">{{ accountEmail }}</span>
                  <small>{{ authStore.isAdmin ? '管理员账户' : 'Smirel 账户' }}</small>
                </div>
              </div>

              <div class="smh-account-menu-group">
                <router-link to="/profile" role="menuitem" @click="accountMenuOpen = false">
                  <span>账户设置</span><b>→</b>
                </router-link>
                <router-link to="/keys" role="menuitem" @click="accountMenuOpen = false">
                  <span>API Key</span><b>→</b>
                </router-link>
                <router-link to="/usage" role="menuitem" @click="accountMenuOpen = false">
                  <span>用量与日志</span><b>→</b>
                </router-link>
              </div>

              <button type="button" class="smh-account-logout" role="menuitem" @click="handleLogout">
                退出登录
              </button>
            </div>
          </div>

          <div v-else class="smh-auth-actions">
            <router-link to="/login" class="smh-login-link">登录</router-link>
            <router-link to="/register" class="smh-register-link">注册账户</router-link>
          </div>
        </div>
      </header>

      <main>
        <section class="smh-hero">
          <div class="smh-hero-copy">
            <p class="smh-kicker">SMIREL API</p>
            <h1>一个 API 地址，调用 OpenAI、Claude 和 Gemini。</h1>
            <p class="smh-hero-description">
              把 Base URL 改成 Smirel，再填入自己的 API Key。OpenAI SDK 和常用客户端可以按原来的方式继续用。
            </p>
            <div class="smh-hero-actions">
              <router-link :to="startPath" class="smh-button smh-button--primary">
                {{ isAuthenticated ? '打开控制台' : '注册账户' }}
              </router-link>
              <router-link to="/model-plaza" class="smh-button smh-button--quiet">查看模型价格</router-link>
            </div>
          </div>
        </section>

        <section id="access" class="smh-glass smh-access-card" aria-label="API 接入信息">
          <div class="smh-access-top">
            <div class="smh-access-main">
              <span>BASE URL</span>
              <code>{{ apiBase }}</code>
            </div>
            <button type="button" class="smh-copy-button" @click="copyBaseUrl">
              {{ copied ? '已复制' : '复制地址' }}
            </button>
          </div>

          <div class="smh-access-meta">
            <div class="smh-access-item">
              <span>认证</span>
              <strong>Bearer API Key</strong>
            </div>
            <div class="smh-access-item">
              <span>协议</span>
              <strong>OpenAI Compatible</strong>
            </div>
            <div class="smh-access-item">
              <span>计费</span>
              <strong>按实际用量</strong>
            </div>
          </div>
        </section>

        <section class="smh-glass smh-process-card">
          <div class="smh-process-copy">
            <p class="smh-section-label">开始使用</p>
            <h2>三步接入。</h2>
            <p>注册、创建 Key、替换地址。然后就可以按原来的方式调用。</p>
            <router-link :to="startPath" class="smh-text-link">
              {{ isAuthenticated ? '进入控制台' : '创建账户' }} →
            </router-link>
          </div>

          <div class="smh-steps" aria-label="开始使用流程">
            <div class="smh-step">
              <span>01</span>
              <h3>注册账户</h3>
              <p>登录后进入控制台。</p>
            </div>
            <div class="smh-step">
              <span>02</span>
              <h3>创建 API Key</h3>
              <p>给项目生成一把独立密钥。</p>
            </div>
            <div class="smh-step">
              <span>03</span>
              <h3>替换 Base URL</h3>
              <p>把客户端地址改成 Smirel。</p>
            </div>
          </div>
        </section>

        <section class="smh-glass smh-routing-card">
          <div class="smh-routing-copy">
            <p class="smh-section-label">模型接入</p>
            <h2>模型可以换，接入地址不用换。</h2>
            <p>
              使用支持的模型时，Base URL 和 API Key 保持不变。可用模型和价格直接在模型页查看。
            </p>
            <div class="smh-provider-line" aria-label="支持的模型服务">
              <span>OpenAI</span>
              <span>Anthropic</span>
              <span>Google</span>
            </div>
            <router-link to="/model-plaza" class="smh-text-link">查看模型价格 →</router-link>
          </div>

          <div class="smh-client-example" aria-label="OpenAI SDK 配置示例">
            <div class="smh-client-example-head">
              <span>OpenAI SDK</span>
              <span>Python</span>
            </div>
            <pre><code>from openai import OpenAI

client = OpenAI(
    base_url="{{ apiBase }}",
    api_key="sk-..."
)</code></pre>
          </div>
        </section>

        <section class="smh-quick-grid" aria-label="常用入口">
          <router-link to="/model-plaza" class="smh-glass smh-quick-card smh-quick-card--wide">
            <span class="smh-quick-label">MODELS</span>
            <h3>模型与价格</h3>
            <p>查看当前可用模型、单价和服务状态。</p>
            <b>查看 →</b>
          </router-link>

          <router-link :to="keyPath" class="smh-glass smh-quick-card">
            <span class="smh-quick-label">API KEY</span>
            <h3>密钥</h3>
            <p>创建、停用和管理项目密钥。</p>
            <b>{{ isAuthenticated ? '管理 →' : '登录 →' }}</b>
          </router-link>

          <router-link to="/key-usage" class="smh-glass smh-quick-card">
            <span class="smh-quick-label">USAGE</span>
            <h3>用量</h3>
            <p>查看请求、Token 和费用记录。</p>
            <b>查询 →</b>
          </router-link>
        </section>
      </main>

      <footer class="smh-footer">
        <span>© {{ currentYear }} {{ siteName }}</span>
        <div>
          <span>{{ apiBase }}</span>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">接入文档</a>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-shared-glass-v1.css'

const appStore = useAppStore()
const authStore = useAuthStore()
const copied = ref(false)
const accountMenuOpen = ref(false)
const accountMenuRoot = ref<HTMLElement | null>(null)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Smirel')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const apiBase = computed(() => (
  appStore.cachedPublicSettings?.api_base_url
  || appStore.apiBaseUrl
  || 'https://api.smirel.com/v1'
).trim().replace(/\/$/, ''))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const accountPath = computed(() => isAuthenticated.value
  ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
  : '/login')
const startPath = computed(() => isAuthenticated.value ? accountPath.value : '/register')
const keyPath = computed(() => isAuthenticated.value ? '/keys' : '/login')
const currentYear = computed(() => new Date().getFullYear())
const avatarUrl = computed(() => sanitizeUrl(authStore.user?.avatar_url || '', { allowRelative: true, allowDataUrl: true }))
const accountLabel = computed(() => authStore.user?.username?.trim() || authStore.user?.email?.trim() || 'Smirel Account')
const accountEmail = computed(() => authStore.user?.email?.trim() || '')
const accountInitial = computed(() => accountLabel.value.charAt(0).toUpperCase() || 'S')

async function copyBaseUrl() {
  try {
    await navigator.clipboard.writeText(apiBase.value)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 1600)
  } catch {
    copied.value = false
  }
}

async function handleLogout() {
  accountMenuOpen.value = false
  await authStore.logout()
}

function handleDocumentPointer(event: MouseEvent) {
  if (!accountMenuOpen.value) return
  const target = event.target
  if (target instanceof Node && accountMenuRoot.value?.contains(target)) return
  accountMenuOpen.value = false
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') accountMenuOpen.value = false
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
  document.addEventListener('click', handleDocumentPointer)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentPointer)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>

<style scoped>
.smh-home {
  min-height: 100vh;
  background: transparent;
}

.smh-home a {
  color: inherit;
  text-decoration: none;
}

.smh-environment::after {
  background:
    linear-gradient(180deg, rgba(4, 10, 16, .13), rgba(4, 10, 16, .26)),
    radial-gradient(ellipse at 55% 40%, rgba(5, 12, 18, 0) 25%, rgba(5, 12, 18, .18) 100%);
}

.smh-shell {
  position: relative;
  z-index: 1;
  width: min(1480px, calc(100vw - 72px));
  min-height: 100vh;
  margin: 0 auto;
  padding: 18px 0 34px;
}

.smh-header {
  position: sticky;
  z-index: 30;
  top: 16px;
  min-height: 66px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 26px;
  padding: 7px 9px 7px 16px;
  border: 1px solid rgba(255, 255, 255, .10);
  border-radius: 18px;
  background: rgba(5, 15, 24, .40);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .06),
    0 14px 38px rgba(2, 8, 14, .10);
  -webkit-backdrop-filter: blur(24px) saturate(125%);
  backdrop-filter: blur(24px) saturate(125%);
}

.smh-header-right,
.smh-nav,
.smh-status,
.smh-auth-actions,
.smh-account-area {
  display: flex;
  align-items: center;
}

.smh-header-right {
  min-width: 0;
  flex: 1;
  justify-content: flex-end;
  gap: 12px;
}

.smh-nav {
  gap: 24px;
  margin-right: 6px;
}

.smh-nav a,
.smh-status,
.smh-login-link,
.smh-register-link,
.smh-console-link,
.smh-account-trigger {
  font-size: .78rem;
}

.smh-nav a {
  color: rgba(255, 255, 255, .68);
  transition: color .18s ease;
}

.smh-nav a:hover {
  color: rgba(255, 255, 255, .98);
}

.smh-status {
  gap: 7px;
  padding: 0 4px;
  color: rgba(255, 255, 255, .58);
  white-space: nowrap;
}

.smh-status i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #83efc7;
  box-shadow: 0 0 12px rgba(131, 239, 199, .42);
}

.smh-auth-actions,
.smh-account-area {
  position: relative;
  gap: 8px;
}

.smh-login-link,
.smh-register-link,
.smh-console-link,
.smh-account-trigger {
  min-height: 42px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, .10);
  transition:
    color .18s ease,
    background-color .18s ease,
    border-color .18s ease,
    transform .18s ease;
}

.smh-login-link {
  padding: 0 15px;
  border-color: transparent;
  color: rgba(255, 255, 255, .72);
  background: transparent;
}

.smh-login-link:hover {
  color: rgba(255, 255, 255, .98);
  background: rgba(255, 255, 255, .05);
}

.smh-register-link,
.smh-console-link {
  gap: 8px;
  padding: 0 16px;
  border-radius: 12px;
  color: rgba(255, 255, 255, .94);
  background: rgba(255, 255, 255, .12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, .06);
}

.smh-register-link:hover,
.smh-console-link:hover {
  border-color: rgba(255, 255, 255, .18);
  background: rgba(255, 255, 255, .18);
  transform: translateY(-1px);
}

.smh-console-link svg {
  width: 16px;
  height: 16px;
  opacity: .82;
}

.smh-account-trigger {
  min-width: 54px;
  gap: 4px;
  padding: 4px 8px 4px 5px;
  border-radius: 13px;
  color: rgba(255, 255, 255, .86);
  background: rgba(255, 255, 255, .08);
  cursor: pointer;
}

.smh-account-trigger:hover,
.smh-account-trigger[aria-expanded='true'] {
  border-color: rgba(255, 255, 255, .17);
  background: rgba(255, 255, 255, .14);
}

.smh-account-trigger > img,
.smh-account-trigger > span,
.smh-account-avatar {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 10px;
  background: rgba(117, 178, 255, .20);
  color: rgba(255, 255, 255, .98);
  font-size: .78rem;
  font-weight: 700;
}

.smh-account-trigger > img,
.smh-account-avatar img {
  object-fit: cover;
}

.smh-account-chevron {
  width: 14px;
  height: 14px;
  opacity: .58;
}

.smh-account-menu {
  position: absolute;
  top: calc(100% + 11px);
  right: 0;
  width: 286px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, .11);
  border-radius: 15px;
  background: rgba(5, 15, 24, .78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .06),
    0 22px 54px rgba(2, 8, 14, .26);
  -webkit-backdrop-filter: blur(28px) saturate(128%);
  backdrop-filter: blur(28px) saturate(128%);
}

.smh-account-summary {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  padding: 16px;
}

.smh-account-avatar--large {
  width: 42px;
  height: 42px;
  flex-basis: 42px;
  border-radius: 12px;
  font-size: .92rem;
}

.smh-account-summary-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.smh-account-summary-copy strong,
.smh-account-summary-copy span,
.smh-account-summary-copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smh-account-summary-copy strong {
  color: rgba(255, 255, 255, .94);
  font-size: .90rem;
  font-weight: 650;
}

.smh-account-summary-copy span {
  color: rgba(255, 255, 255, .52);
  font-size: .72rem;
}

.smh-account-summary-copy small {
  color: rgba(131, 239, 199, .72);
  font-size: .65rem;
  letter-spacing: .06em;
}

.smh-account-menu-group {
  padding: 6px;
  border-top: 1px solid rgba(255, 255, 255, .07);
  border-bottom: 1px solid rgba(255, 255, 255, .07);
}

.smh-account-menu-group a {
  min-height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 11px;
  border-radius: 9px;
  color: rgba(255, 255, 255, .76);
  font-size: .80rem;
  transition: background-color .16s ease, color .16s ease;
}

.smh-account-menu-group a:hover {
  color: rgba(255, 255, 255, .98);
  background: rgba(255, 255, 255, .07);
}

.smh-account-menu-group b {
  color: rgba(255, 255, 255, .34);
  font-weight: 500;
}

.smh-account-logout {
  width: calc(100% - 12px);
  min-height: 40px;
  margin: 6px;
  padding: 0 11px;
  border: 0;
  border-radius: 9px;
  background: transparent;
  color: rgba(255, 255, 255, .56);
  font: inherit;
  font-size: .78rem;
  text-align: left;
  cursor: pointer;
  transition: background-color .16s ease, color .16s ease;
}

.smh-account-logout:hover {
  color: rgba(255, 255, 255, .92);
  background: rgba(255, 255, 255, .06);
}

.smh-account-trigger:focus-visible,
.smh-account-logout:focus-visible,
.smh-register-link:focus-visible,
.smh-console-link:focus-visible,
.smh-login-link:focus-visible {
  outline: 2px solid rgba(171, 215, 255, .72);
  outline-offset: 2px;
}

.smh-hero {
  min-height: 390px;
  display: flex;
  align-items: center;
  padding: 34px 0 46px;
}

.smh-hero-copy {
  width: min(790px, 62vw);
}

.smh-kicker,
.smh-section-label,
.smh-quick-label,
.smh-access-main > span,
.smh-access-item > span {
  margin: 0;
  color: rgba(255, 255, 255, .50);
  font-size: .74rem;
  letter-spacing: .12em;
}

.smh-hero h1 {
  max-width: 790px;
  margin: 12px 0 0;
  font-size: clamp(3rem, 5vw, 4.55rem);
  line-height: 1.02;
  font-weight: 560;
  letter-spacing: -.052em;
}

.smh-hero-description {
  max-width: 650px;
  margin: 20px 0 0;
  color: rgba(255, 255, 255, .76);
  font-size: 1rem;
  line-height: 1.74;
}

.smh-hero-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 34px;
}

.smh-button {
  min-height: 58px;
  min-width: 172px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 24px;
  border-radius: 8px;
  font-size: .98rem;
  font-weight: 640;
  letter-spacing: -.01em;
  -webkit-backdrop-filter: blur(18px) saturate(118%);
  backdrop-filter: blur(18px) saturate(118%);
  transition: background-color .18s ease, border-color .18s ease, transform .18s ease, box-shadow .18s ease;
}

.smh-button:hover {
  transform: translateY(-2px);
}

.smh-button--primary {
  min-width: 188px;
  border: 1px solid rgba(255, 255, 255, .22);
  background: rgba(255, 255, 255, .20);
  color: rgba(255, 255, 255, .99);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .12),
    0 12px 30px rgba(3, 10, 16, .12);
}

.smh-button--primary:hover {
  border-color: rgba(255, 255, 255, .30);
  background: rgba(255, 255, 255, .27);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .16),
    0 16px 34px rgba(3, 10, 16, .16);
}

.smh-button--quiet {
  border: 1px solid rgba(255, 255, 255, .14);
  background: rgba(6, 16, 25, .28);
  color: rgba(255, 255, 255, .90);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, .05);
}

.smh-button--quiet:hover {
  color: rgba(255, 255, 255, .99);
  background: rgba(10, 22, 32, .38);
  border-color: rgba(255, 255, 255, .22);
}

.smh-glass {
  border: 1px solid rgba(255, 255, 255, .08);
  border-radius: 8px;
  background: rgba(5, 18, 29, .42);
  -webkit-backdrop-filter: blur(22px) saturate(118%);
  backdrop-filter: blur(22px) saturate(118%);
}

.smh-access-card {
  overflow: hidden;
}

.smh-access-top {
  min-height: 124px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 28px;
  padding: 27px 30px;
}

.smh-access-main {
  min-width: 0;
}

.smh-access-main code {
  display: block;
  margin-top: 10px;
  overflow: hidden;
  color: rgba(255, 255, 255, .98);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: clamp(1.12rem, 1.6vw, 1.34rem);
  font-weight: 620;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smh-copy-button {
  min-width: 106px;
  min-height: 44px;
  padding: 0 17px;
  border: 1px solid rgba(255, 255, 255, .12);
  border-radius: 6px;
  background: rgba(255, 255, 255, .10);
  color: rgba(255, 255, 255, .92);
  font: inherit;
  font-size: .80rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color .18s ease, border-color .18s ease;
}

.smh-copy-button:hover {
  background: rgba(255, 255, 255, .16);
  border-color: rgba(255, 255, 255, .18);
}

.smh-access-meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid rgba(255, 255, 255, .07);
}

.smh-access-item {
  min-height: 94px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 20px 30px;
}

.smh-access-item + .smh-access-item {
  border-left: 1px solid rgba(255, 255, 255, .07);
}

.smh-access-item strong {
  margin-top: 9px;
  color: rgba(255, 255, 255, .92);
  font-size: .96rem;
  font-weight: 630;
}

.smh-process-card,
.smh-routing-card {
  display: grid;
  gap: 0;
  margin-top: 18px;
  overflow: hidden;
}

.smh-process-card {
  grid-template-columns: minmax(340px, .82fr) minmax(0, 1.18fr);
}

.smh-process-copy,
.smh-routing-copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 46px 48px 44px;
}

.smh-process-copy {
  min-height: 360px;
  border-right: 1px solid rgba(255, 255, 255, .07);
}

.smh-process-copy h2,
.smh-routing-copy h2 {
  margin: 12px 0 0;
  font-size: clamp(2.45rem, 3.6vw, 3.3rem);
  line-height: 1.07;
  font-weight: 560;
  letter-spacing: -.042em;
}

.smh-process-copy > p:last-of-type,
.smh-routing-copy > p:last-of-type {
  max-width: 570px;
  margin: 19px 0 0;
  color: rgba(255, 255, 255, .64);
  font-size: 1rem;
  line-height: 1.72;
}

.smh-text-link {
  display: inline-flex;
  align-self: flex-start;
  margin-top: 32px;
  color: rgba(255, 255, 255, .86);
  font-size: .86rem;
  font-weight: 620;
}

.smh-text-link:hover {
  color: rgba(255, 255, 255, .98);
}

.smh-steps {
  display: grid;
  grid-template-columns: 1fr;
  align-content: stretch;
  padding: 18px 32px;
}

.smh-step {
  min-height: 108px;
  display: grid;
  grid-template-columns: 56px minmax(170px, .72fr) minmax(0, 1fr);
  gap: 24px;
  align-items: center;
  padding: 22px 10px;
}

.smh-step + .smh-step {
  border-top: 1px solid rgba(255, 255, 255, .08);
}

.smh-step > span {
  color: rgba(255, 255, 255, .48);
  font-size: .76rem;
  letter-spacing: .10em;
}

.smh-step h3 {
  margin: 0;
  color: rgba(255, 255, 255, .96);
  font-size: 1.28rem;
  line-height: 1.25;
  font-weight: 640;
}

.smh-step p {
  margin: 0;
  color: rgba(255, 255, 255, .62);
  font-size: .95rem;
  line-height: 1.62;
}

.smh-routing-card {
  grid-template-columns: minmax(0, .92fr) minmax(440px, 1.08fr);
}

.smh-routing-copy {
  min-height: 380px;
  border-right: 1px solid rgba(255, 255, 255, .07);
}

.smh-provider-line {
  display: flex;
  flex-wrap: wrap;
  gap: 22px;
  margin-top: 30px;
  color: rgba(255, 255, 255, .86);
  font-size: .95rem;
  font-weight: 600;
}

.smh-provider-line span + span::before {
  content: '/';
  margin-right: 22px;
  color: rgba(255, 255, 255, .24);
}

.smh-client-example {
  min-width: 0;
  margin: 24px;
  align-self: stretch;
  border: 1px solid rgba(255, 255, 255, .07);
  border-radius: 6px;
  background: rgba(0, 0, 0, .11);
  overflow: hidden;
}

.smh-client-example-head {
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 21px;
  border-bottom: 1px solid rgba(255, 255, 255, .07);
  color: rgba(255, 255, 255, .52);
  font-size: .76rem;
}

.smh-client-example pre {
  margin: 0;
  padding: 34px 34px 38px;
  overflow-x: auto;
}

.smh-client-example code {
  color: rgba(255, 255, 255, .90);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: .96rem;
  line-height: 1.82;
}

.smh-quick-grid {
  display: grid;
  grid-template-columns: 1.28fr .86fr .86fr;
  gap: 12px;
  margin-top: 18px;
}

.smh-quick-card {
  position: relative;
  min-height: 188px;
  padding: 30px 30px 28px;
  transition: transform .2s ease, background-color .2s ease;
}

.smh-quick-card:hover {
  transform: translateY(-2px);
  background: rgba(5, 18, 29, .50);
}

.smh-quick-card h3 {
  margin: 24px 0 0;
  font-size: 1.44rem;
  line-height: 1.2;
  font-weight: 640;
}

.smh-quick-card p {
  max-width: 88%;
  margin: 11px 0 0;
  color: rgba(255, 255, 255, .62);
  font-size: .95rem;
  line-height: 1.58;
}

.smh-quick-card b {
  position: absolute;
  right: 28px;
  bottom: 26px;
  color: rgba(255, 255, 255, .74);
  font-size: .84rem;
  font-weight: 610;
}

.smh-footer {
  min-height: 76px;
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  color: rgba(255, 255, 255, .38);
  font-size: .70rem;
}

.smh-footer > div {
  display: flex;
  gap: 22px;
}

.smh-footer a:hover {
  color: rgba(255, 255, 255, .72);
}

@media (max-width: 1180px) {
  .smh-shell {
    width: calc(100vw - 48px);
  }

  .smh-nav {
    gap: 17px;
  }

  .smh-process-card {
    grid-template-columns: minmax(310px, .78fr) minmax(0, 1.22fr);
  }

  .smh-step {
    grid-template-columns: 50px minmax(150px, .7fr) minmax(0, 1fr);
    gap: 18px;
  }

  .smh-routing-card {
    grid-template-columns: minmax(0, .88fr) minmax(400px, 1.12fr);
  }
}

@media (max-width: 980px) {
  .smh-nav {
    display: none;
  }

  .smh-hero {
    min-height: 360px;
  }

  .smh-hero-copy {
    width: min(700px, 90vw);
  }

  .smh-process-card,
  .smh-routing-card {
    grid-template-columns: 1fr;
  }

  .smh-process-copy,
  .smh-routing-copy {
    min-height: auto;
    border-right: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-quick-grid {
    grid-template-columns: 1.2fr 1fr;
  }

  .smh-quick-card:last-child {
    grid-column: 1 / -1;
  }
}

@media (max-width: 720px) {
  .smh-shell {
    width: calc(100vw - 28px);
    padding-top: 12px;
  }

  .smh-header {
    top: 10px;
    min-height: 60px;
    padding: 6px 7px 6px 12px;
    border-radius: 16px;
    gap: 12px;
  }

  .smh-status {
    display: none;
  }

  .smh-console-link {
    width: 42px;
    padding: 0;
  }

  .smh-console-link span {
    display: none;
  }

  .smh-account-trigger {
    min-width: 46px;
    padding-right: 5px;
  }

  .smh-account-chevron {
    display: none;
  }

  .smh-register-link {
    padding: 0 12px;
  }

  .smh-login-link {
    padding: 0 10px;
  }

  .smh-account-menu {
    right: -1px;
    width: min(286px, calc(100vw - 34px));
  }

  .smh-hero {
    min-height: 330px;
    padding: 32px 0 40px;
  }

  .smh-hero-copy {
    width: 100%;
  }

  .smh-hero h1 {
    font-size: clamp(2.45rem, 13vw, 3.4rem);
  }

  .smh-hero-description {
    margin-top: 16px;
    font-size: .92rem;
  }

  .smh-hero-actions {
    width: 100%;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 26px;
  }

  .smh-button {
    min-width: 0;
    min-height: 52px;
    flex: 1 1 158px;
    padding: 0 18px;
    font-size: .90rem;
  }

  .smh-button--primary {
    min-width: 0;
  }

  .smh-access-top {
    grid-template-columns: 1fr;
    gap: 18px;
    padding: 23px 20px 20px;
  }

  .smh-copy-button {
    justify-self: start;
  }

  .smh-access-main code {
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .smh-access-meta {
    grid-template-columns: 1fr;
  }

  .smh-access-item {
    min-height: 82px;
    padding: 17px 20px;
  }

  .smh-access-item + .smh-access-item {
    border-left: 0;
    border-top: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-process-copy,
  .smh-routing-copy {
    padding: 32px 26px 30px;
  }

  .smh-process-copy h2,
  .smh-routing-copy h2 {
    font-size: clamp(2.1rem, 10vw, 2.75rem);
  }

  .smh-process-copy > p:last-of-type,
  .smh-routing-copy > p:last-of-type {
    font-size: .92rem;
  }

  .smh-steps {
    padding: 8px 22px 14px;
  }

  .smh-step {
    min-height: 118px;
    grid-template-columns: 42px minmax(0, 1fr);
    grid-template-rows: auto auto;
    gap: 8px 14px;
    padding: 22px 0;
  }

  .smh-step > span {
    grid-row: 1 / -1;
    align-self: start;
    padding-top: 4px;
  }

  .smh-step h3 {
    font-size: 1.12rem;
  }

  .smh-step p {
    grid-column: 2;
    font-size: .86rem;
  }

  .smh-client-example {
    margin: 14px;
  }

  .smh-client-example pre {
    padding: 24px 20px 28px;
  }

  .smh-client-example code {
    font-size: .80rem;
  }

  .smh-quick-grid {
    grid-template-columns: 1fr;
  }

  .smh-quick-card:last-child {
    grid-column: auto;
  }

  .smh-quick-card {
    min-height: 158px;
    padding: 24px;
  }

  .smh-quick-card h3 {
    margin-top: 18px;
    font-size: 1.28rem;
  }

  .smh-quick-card p {
    font-size: .88rem;
  }

  .smh-footer {
    align-items: flex-start;
    flex-direction: column;
    padding: 30px 0 10px;
  }

  .smh-footer > div {
    flex-direction: column;
    gap: 7px;
  }
}

@media (max-width: 520px) {
  .spg-brand-copy {
    display: none;
  }

  .smh-header-right {
    gap: 5px;
  }

  .smh-login-link {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .smh-nav a,
  .smh-login-link,
  .smh-register-link,
  .smh-console-link,
  .smh-account-trigger,
  .smh-account-menu-group a,
  .smh-account-logout,
  .smh-button,
  .smh-copy-button,
  .smh-quick-card {
    transition: none !important;
  }
}
</style>