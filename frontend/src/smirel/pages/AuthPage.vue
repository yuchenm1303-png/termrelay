<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  api,
  buildOAuthStartUrl,
  getErrorMessage,
  previewMode,
  sanitizeOAuthRedirect,
  type OAuthProvider,
} from '../core/api'
import { useSession } from '../core/session'
import { interfacePreferences } from '../core/preferences'

const route = useRoute()
const router = useRouter()
const { login, register, isAdmin } = useSession()
interface PublicAuthSettings {
  turnstile_enabled?: boolean
  turnstile_site_key?: string
}

interface TurnstileApi {
  render: (container: HTMLElement, options: Record<string, unknown>) => string
  reset: (widgetId?: string) => void
  remove?: (widgetId: string) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileApi
  }
}

const TURNSTILE_SCRIPT_ID = 'smirel-turnstile-script'
const logoUrl = `${import.meta.env.BASE_URL}muxway-mark.svg?v=20260920-ribbon`
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const token = ref(String(route.query.token || ''))
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const message = ref('')
const error = ref('')
const turnstileContainer = ref<HTMLElement | null>(null)
const turnstileEnabled = ref(false)
const turnstileSiteKey = ref('')
const turnstileToken = ref('')
const turnstileLoadError = ref('')
let turnstileWidgetId: string | undefined

const kind = computed(() => String(route.meta.authKind || 'login'))
const showOAuth = computed(() => kind.value === 'login' || kind.value === 'register')
const needsTurnstile = computed(() =>
  !previewMode
  && turnstileEnabled.value
  && Boolean(turnstileSiteKey.value)
  && ['login', 'register', 'forgot'].includes(kind.value),
)
const turnstilePending = computed(() => needsTurnstile.value && !turnstileToken.value && !turnstileLoadError.value)
const turnstileTheme = computed(() => interfacePreferences.resolvedTheme === 'light' ? 'light' : 'dark')
const titles: Record<string, string> = {
  login: '登录 Muxway 模枢',
  register: '创建 Muxway 账户',
  forgot: '找回密码',
  reset: '设置新密码',
}
const subtitles: Record<string, string> = {
  login: '继续进入你的 API 工作区。',
  register: '一个账户管理密钥、用量和服务。',
  forgot: '输入邮箱，我们会发送重置链接。',
  reset: '为你的账户设置新的密码。',
}
const submitLabels: Record<string, string> = {
  login: '登录',
  register: '创建账户',
  forgot: '发送重置链接',
  reset: '更新密码',
}
const title = computed(() => titles[kind.value] || titles.login)
const subtitle = computed(() => subtitles[kind.value] || '')
const submitLabel = computed(() => submitLabels[kind.value] || submitLabels.login)

function loadTurnstileScript(): Promise<void> {
  if (typeof window === 'undefined' || typeof document === 'undefined') return Promise.resolve()
  if (window.turnstile) return Promise.resolve()

  return new Promise((resolve, reject) => {
    const existing = document.getElementById(TURNSTILE_SCRIPT_ID) as HTMLScriptElement | null
    const handleLoad = () => {
      if (window.turnstile) resolve()
      else reject(new Error('Cloudflare Turnstile API unavailable'))
    }
    const handleError = () => reject(new Error('Cloudflare Turnstile script failed to load'))

    if (existing) {
      existing.addEventListener('load', handleLoad, { once: true })
      existing.addEventListener('error', handleError, { once: true })
      window.setTimeout(() => {
        if (window.turnstile) resolve()
      }, 0)
      return
    }

    const script = document.createElement('script')
    script.id = TURNSTILE_SCRIPT_ID
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.addEventListener('load', handleLoad, { once: true })
    script.addEventListener('error', handleError, { once: true })
    document.head.appendChild(script)
  })
}

function disposeTurnstile() {
  turnstileToken.value = ''
  if (turnstileWidgetId && window.turnstile?.remove) {
    try {
      window.turnstile.remove(turnstileWidgetId)
    } catch {
      // The widget may already have been removed by navigation.
    }
  }
  turnstileWidgetId = undefined
  if (turnstileContainer.value) turnstileContainer.value.replaceChildren()
}

function resetTurnstile() {
  turnstileToken.value = ''
  if (!turnstileWidgetId || !window.turnstile) return
  try {
    window.turnstile.reset(turnstileWidgetId)
  } catch {
    turnstileWidgetId = undefined
    void renderTurnstile()
  }
}

async function renderTurnstile() {
  disposeTurnstile()
  turnstileLoadError.value = ''
  if (!needsTurnstile.value) return

  await nextTick()
  const container = turnstileContainer.value
  if (!container) return

  try {
    await loadTurnstileScript()
    if (!window.turnstile) throw new Error('Cloudflare Turnstile API unavailable')

    turnstileWidgetId = window.turnstile.render(container, {
      sitekey: turnstileSiteKey.value,
      theme: turnstileTheme.value,
      appearance: 'always',
      size: 'flexible',
      callback: (value: unknown) => {
        turnstileToken.value = typeof value === 'string' ? value : ''
        turnstileLoadError.value = ''
      },
      'expired-callback': () => {
        turnstileToken.value = ''
      },
      'timeout-callback': () => {
        turnstileToken.value = ''
      },
      'error-callback': () => {
        turnstileToken.value = ''
        turnstileLoadError.value = '安全验证暂时不可用，请稍后重试。'
      },
    })
  } catch {
    turnstileLoadError.value = '安全验证加载失败，请刷新页面后重试。'
  }
}

async function loadPublicAuthSettings() {
  if (previewMode) return
  try {
    const { data } = await api.get<PublicAuthSettings>('/settings/public')
    turnstileEnabled.value = Boolean(data.turnstile_enabled)
    turnstileSiteKey.value = String(data.turnstile_site_key || '').trim()
    if (turnstileEnabled.value && !turnstileSiteKey.value) {
      turnstileLoadError.value = '安全验证尚未完成配置，请联系管理员。'
    }
  } catch {
    // Keep authentication available when public settings cannot be loaded.
    // A server that requires Turnstile will still reject a tokenless request safely.
    turnstileEnabled.value = false
    turnstileSiteKey.value = ''
  }
}

watch(
  () => [kind.value, turnstileEnabled.value, turnstileSiteKey.value, turnstileTheme.value],
  () => void renderTurnstile(),
)

onMounted(() => void loadPublicAuthSettings())
onBeforeUnmount(disposeTurnstile)

function startOAuth(provider: OAuthProvider) {
  error.value = ''
  message.value = ''

  if (previewMode) {
    void router.push('/admin/dashboard')
    return
  }

  const redirect = sanitizeOAuthRedirect(
    typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard',
  )
  sessionStorage.setItem('smirel_oauth_provider', provider)
  sessionStorage.setItem('smirel_oauth_redirect', redirect)
  window.location.assign(buildOAuthStartUrl(provider, redirect))
}

async function submit() {
  error.value = ''
  message.value = ''
  if (kind.value === 'register' && password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (needsTurnstile.value && !turnstileToken.value) {
    error.value = turnstileLoadError.value || '请先完成人机验证。'
    return
  }

  const turnstileTokenForRequest = turnstileToken.value
  loading.value = true
  try {
    if (previewMode) {
      await router.push('/admin/dashboard')
      return
    }

    if (kind.value === 'login') {
      await login(email.value.trim(), password.value, turnstileTokenForRequest)
      const redirect = typeof route.query.redirect === 'string'
        ? route.query.redirect
        : (isAdmin.value ? '/admin/dashboard' : '/dashboard')
      await router.push(redirect)
    } else if (kind.value === 'register') {
      await register(email.value.trim(), password.value, turnstileTokenForRequest)
      const redirect = typeof route.query.redirect === 'string'
        ? route.query.redirect
        : (isAdmin.value ? '/admin/dashboard' : '/dashboard')
      await router.push(redirect)
    } else if (kind.value === 'forgot') {
      await api.post('/auth/forgot-password', {
        email: email.value.trim(),
        turnstile_token: turnstileTokenForRequest,
      })
      message.value = '重置链接已发送，请检查邮箱。'
    } else {
      await api.post('/auth/reset-password', {
        email: email.value.trim(),
        token: token.value.trim(),
        new_password: password.value,
      })
      message.value = '密码已更新，现在可以登录。'
    }
  } catch (caught) {
    error.value = getErrorMessage(caught)
  } finally {
    loading.value = false
    if (turnstileTokenForRequest) resetTurnstile()
  }
}
</script>

<template>
  <div class="auth-page" :class="{ 'is-light': interfacePreferences.resolvedTheme === 'light' }">
    <RouterLink to="/home" class="auth-brand brand-link">
      <img :src="logoUrl" alt="Muxway" />
      <span>
        <strong>Muxway</strong>
        <small>模枢 · API SERVICE</small>
      </span>
    </RouterLink>

    <main class="auth-layout">
      <section class="auth-intro" aria-label="Muxway Console">
        <span class="auth-kicker">MUXWAY CONSOLE</span>
        <h2>统一管理你的<br /><em>API</em> 工作区。</h2>
        <p>密钥、模型、用量与账单，集中在一个清晰、稳定的控制台。</p>

        <div class="auth-capabilities">
          <div>
            <b>01</b>
            <span>
              <strong>统一入口</strong>
              <small>一个账户管理 API Keys 与服务</small>
            </span>
          </div>
          <div>
            <b>02</b>
            <span>
              <strong>清晰用量</strong>
              <small>调用、余额与订单状态集中查看</small>
            </span>
          </div>
          <div>
            <b>03</b>
            <span>
              <strong>账户管理</strong>
              <small>工作区与账户信息统一维护</small>
            </span>
          </div>
        </div>
      </section>

      <section class="auth-card">
        <div class="auth-card-meta">
          <span>MUXWAY ACCOUNT</span>
          <i><b></b>SECURE ACCESS</i>
        </div>

        <header>
          <h1>{{ title }}</h1>
          <p>{{ subtitle }}</p>
        </header>

        <div v-if="showOAuth" class="oauth-login">
          <div class="oauth-actions">
            <button type="button" class="oauth-button" @click="startOAuth('google')">
              <span class="oauth-provider-mark" aria-hidden="true">
                <svg viewBox="0 0 18 18" role="presentation">
                  <path fill="#4285F4" d="M17.64 9.205c0-.639-.057-1.252-.164-1.841H9v3.481h4.844a4.14 4.14 0 0 1-1.797 2.716v2.26h2.909c1.702-1.567 2.684-3.874 2.684-6.616Z" />
                  <path fill="#34A853" d="M9 18c2.43 0 4.467-.806 5.956-2.179l-2.909-2.26c-.806.54-1.835.859-3.047.859-2.344 0-4.328-1.585-5.037-3.714H.957v2.332A9 9 0 0 0 9 18Z" />
                  <path fill="#FBBC05" d="M3.963 10.706A5.41 5.41 0 0 1 3.682 9c0-.592.102-1.168.281-1.706V4.962H.957A9 9 0 0 0 0 9c0 1.452.347 2.827.957 4.038l3.006-2.332Z" />
                  <path fill="#EA4335" d="M9 3.58c1.322 0 2.508.455 3.441 1.346l2.582-2.582C13.463.892 11.426 0 9 0A9 9 0 0 0 .957 4.962l3.006 2.332C4.672 5.165 6.656 3.58 9 3.58Z" />
                </svg>
              </span>
              <span>使用 Google 继续</span>
            </button>
            <button type="button" class="oauth-button" @click="startOAuth('github')">
              <span class="oauth-provider-mark github" aria-hidden="true">
                <svg viewBox="0 0 24 24" role="presentation">
                  <path fill="currentColor" d="M12 .5C5.648.5.5 5.648.5 12c0 5.08 3.292 9.387 7.86 10.907.575.105.785-.25.785-.555 0-.274-.01-1-.016-1.962-3.197.695-3.872-1.54-3.872-1.54-.523-1.33-1.278-1.684-1.278-1.684-1.045-.714.079-.7.079-.7 1.155.081 1.762 1.186 1.762 1.186 1.027 1.76 2.695 1.252 3.352.957.104-.744.402-1.252.732-1.54-2.552-.291-5.236-1.276-5.236-5.68 0-1.255.449-2.281 1.184-3.085-.118-.291-.513-1.462.113-3.048 0 0 .966-.309 3.165 1.178A10.98 10.98 0 0 1 12 6.096c.977.004 1.96.132 2.88.387 2.198-1.487 3.162-1.178 3.162-1.178.627 1.586.232 2.757.114 3.048.737.804 1.183 1.83 1.183 3.085 0 4.415-2.688 5.386-5.248 5.67.413.355.781 1.057.781 2.13 0 1.538-.014 2.779-.014 3.157 0 .308.207.666.79.553C20.21 21.383 23.5 17.078 23.5 12 23.5 5.648 18.352.5 12 .5Z" />
                </svg>
              </span>
              <span>使用 GitHub 继续</span>
            </button>
          </div>
          <div class="oauth-divider"><span>或使用邮箱</span></div>
        </div>

        <form :class="{ 'with-oauth': showOAuth }" @submit.prevent="submit">
          <div class="auth-field">
            <label for="auth-email">邮箱</label>
            <span class="auth-field-control">
              <svg class="auth-field-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><rect x="3.5" y="5.5" width="17" height="13" rx="2.5"/><path d="m4.5 7 7.5 6 7.5-6"/></svg>
              <input
                id="auth-email"
                v-model="email"
                type="email"
                autocomplete="email"
                required
                placeholder="name@example.com"
              />
            </span>
          </div>

          <div v-if="kind !== 'forgot'" class="auth-field">
            <label for="auth-password">密码</label>
            <span class="auth-field-control">
              <svg class="auth-field-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><rect x="5" y="10" width="14" height="11" rx="2.5"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/></svg>
              <input
                id="auth-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                :autocomplete="kind === 'login' ? 'current-password' : 'new-password'"
                required
                placeholder="••••••••"
              />
              <button type="button" class="auth-reveal" :aria-label="showPassword ? '隐藏密码' : '显示密码'" :aria-pressed="showPassword" @click.prevent="showPassword = !showPassword">
                <svg v-if="!showPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><path d="M4 4 20 20M10.7 6.1A11 11 0 0 1 12 6c6 0 9.5 6 9.5 6a15.8 15.8 0 0 1-3.1 3.5M6.8 7.8C4.1 9.4 2.5 12 2.5 12s3.5 6 9.5 6c1.5 0 2.8-.4 4-1"/><path d="M10 10a3 3 0 0 0 4 4"/></svg>
              </button>
            </span>
          </div>

          <div v-if="kind === 'register'" class="auth-field">
            <label for="auth-confirm-password">确认密码</label>
            <span class="auth-field-control">
              <svg class="auth-field-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><rect x="5" y="10" width="14" height="11" rx="2.5"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/></svg>
              <input
                id="auth-confirm-password"
                v-model="confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                autocomplete="new-password"
                required
                placeholder="••••••••"
              />
              <button type="button" class="auth-reveal" :aria-label="showConfirmPassword ? '隐藏确认密码' : '显示确认密码'" :aria-pressed="showConfirmPassword" @click.prevent="showConfirmPassword = !showConfirmPassword">
                <svg v-if="!showConfirmPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><path d="M4 4 20 20M10.7 6.1A11 11 0 0 1 12 6c6 0 9.5 6 9.5 6a15.8 15.8 0 0 1-3.1 3.5M6.8 7.8C4.1 9.4 2.5 12 2.5 12s3.5 6 9.5 6c1.5 0 2.8-.4 4-1"/><path d="M10 10a3 3 0 0 0 4 4"/></svg>
              </button>
            </span>
          </div>

          <div v-if="kind === 'reset'" class="auth-field">
            <label for="auth-token">重置令牌</label>
            <span class="auth-field-control">
              <input id="auth-token" v-model="token" type="text" required placeholder="Reset token" />
            </span>
          </div>

          <div v-if="needsTurnstile" class="turnstile-official">
            <div
              ref="turnstileContainer"
              class="turnstile-widget"
              aria-label="Cloudflare Turnstile 人机验证"
            ></div>
            <p v-if="turnstileLoadError" class="turnstile-error" aria-live="polite">
              {{ turnstileLoadError }}
            </p>
          </div>

          <p v-if="error" class="form-error" role="alert">{{ error }}</p>
          <p v-if="message" class="form-success" role="status">{{ message }}</p>

          <button
            class="auth-submit"
            type="submit"
            :disabled="loading || turnstilePending || Boolean(turnstileLoadError)"
          >
            <span>{{ loading ? '处理中…' : (previewMode ? '进入预览控制台' : (turnstilePending ? '完成安全验证' : submitLabel)) }}</span>
            <b aria-hidden="true">→</b>
          </button>
        </form>

        <footer v-if="kind === 'login'">
          <RouterLink to="/forgot-password">忘记密码？</RouterLink>
          <span>没有账户？ <RouterLink to="/register">注册</RouterLink></span>
        </footer>
        <footer v-else>
          <RouterLink to="/login">← 返回登录</RouterLink>
        </footer>
      </section>
    </main>

    <footer class="auth-page-footer">
      <span>Muxway 模枢 · One API. Every model.</span>
      <span>Console Access</span>
    </footer>
  </div>
</template>

<style scoped>
/* Auth is a dedicated visual surface: theme tokens keep all four auth flows
 * (login, register, forgot and reset) in sync without duplicate overrides. */
.auth-page {
  --auth-ink: #f4f6ff;
  --auth-subtle: #a9b4ce;
  --auth-muted: #8593b3;
  --auth-line: rgba(182, 202, 245, .16);
  --auth-panel: rgba(20, 28, 53, .88);
  --auth-panel-line: rgba(205, 216, 255, .17);
  --auth-field: rgba(10, 18, 38, .65);
  --auth-field-line: rgba(179, 197, 238, .21);
  --auth-focus: #93b4ff;
  --auth-link: #a3baff;
  position: relative;
  display: flex;
  flex-direction: column;
  /* Global app.css still contains the legacy padded auth layout. Reset it
   * here so this dedicated view fits the viewport instead of growing vertically. */
  height: 100vh;
  height: 100dvh;
  min-height: 100svh;
  padding: 0;
  overflow-x: clip;
  /* Only scroll when a short/zoomed viewport, validation message or Turnstile
   * genuinely needs more space; never crop the submit button or challenge. */
  overflow-y: auto;
  isolation: isolate;
  color: var(--auth-ink);
  color-scheme: dark;
  background:
    radial-gradient(ellipse at 80% 25%, rgba(88, 109, 236, .19), transparent 42%),
    radial-gradient(ellipse at 13% 88%, rgba(24, 129, 199, .11), transparent 45%),
    #090e1c;
  font-family: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans SC', 'Microsoft YaHei', sans-serif;
}
.auth-page::before,
.auth-page::after {
  content: '';
  position: absolute;
  z-index: -1;
  pointer-events: none;
}
.auth-page::before {
  inset: 0;
  background:
    radial-gradient(ellipse 40% 38% at 69% 29%, rgba(73, 140, 250, .14), transparent 85%),
    radial-gradient(ellipse 25% 30% at 96% 63%, rgba(136, 93, 248, .17), transparent 85%),
    radial-gradient(ellipse 55% 25% at 43% 106%, rgba(81, 148, 244, .11), transparent 90%);
}
.auth-page::after {
  width: min(63vw, 920px);
  aspect-ratio: 1.2;
  top: 3%;
  right: -12%;
  border: 1px solid rgba(164, 178, 255, .14);
  border-radius: 47% 53% 62% 38% / 55% 42% 58% 45%;
  transform: rotate(-25deg);
  background: linear-gradient(140deg, rgba(87, 144, 255, .13), rgba(162, 111, 246, .07) 45%, transparent 76%);
  box-shadow: inset 0 0 95px rgba(100, 139, 255, .05), 0 0 95px rgba(117, 133, 255, .035);
}
.auth-page.is-light {
  --auth-ink: #182137;
  --auth-subtle: #556580;
  --auth-muted: #71819a;
  --auth-line: #dce5f3;
  --auth-panel: rgba(255, 255, 255, .90);
  --auth-panel-line: rgba(207, 221, 244, .82);
  --auth-field: rgba(255, 255, 255, .93);
  --auth-field-line: #d9e3f3;
  --auth-focus: #578df4;
  --auth-link: #3567db;
  color-scheme: light;
  background:
    radial-gradient(ellipse 39% 52% at 79% 26%, rgba(165, 208, 255, .28), transparent 88%),
    radial-gradient(ellipse 36% 42% at 98% 67%, rgba(204, 189, 255, .23), transparent 88%),
    radial-gradient(ellipse 48% 36% at 9% 94%, rgba(209, 233, 255, .31), transparent 88%),
    #f9fbff;
}
.auth-page.is-light::before {
  background:
    radial-gradient(ellipse 44% 52% at 70% 27%, rgba(99, 182, 255, .17), transparent 81%),
    radial-gradient(ellipse 32% 42% at 94% 63%, rgba(153, 119, 250, .14), transparent 86%),
    radial-gradient(ellipse 54% 30% at 41% 108%, rgba(147, 195, 255, .20), transparent 83%);
}
.auth-page.is-light::after {
  border-color: rgba(145, 175, 250, .18);
  background: linear-gradient(138deg, rgba(173, 208, 255, .15), rgba(173, 155, 255, .09) 51%, transparent 80%);
  box-shadow: inset 0 0 120px rgba(129, 175, 255, .05), 0 0 130px rgba(163, 179, 255, .07);
}
.auth-brand {
  position: absolute;
  z-index: 2;
  top: clamp(25px, 4vh, 39px);
  left: clamp(24px, 3.7vw, 60px);
  display: inline-flex;
  min-height: 52px;
  align-items: center;
  gap: 13px;
  border-radius: 14px;
  text-decoration: none;
}
.auth-brand img {
  display: block;
  width: 56px;
  height: 45px;
  flex: none;
  object-fit: contain;
  filter: drop-shadow(0 4px 6px rgba(89, 106, 230, .14));
}
.auth-brand > span { display: flex; flex-direction: column; gap: 5px; }
.auth-brand strong {
  color: var(--auth-ink);
  font-size: 1.26rem;
  font-weight: 760;
  line-height: 1;
  letter-spacing: -.04em;
}
.auth-brand small {
  color: var(--auth-muted);
  font-size: .65rem;
  font-weight: 700;
  letter-spacing: .18em;
  line-height: 1.25;
}
.auth-brand:focus-visible,
.oauth-button:focus-visible,
.auth-reveal:focus-visible,
.auth-submit:focus-visible,
.auth-card footer a:focus-visible {
  outline: 3px solid var(--auth-focus);
  outline-offset: 3px;
}
.auth-layout {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(430px, 472px);
  align-items: center;
  gap: clamp(50px, 7vw, 128px);
  width: min(1280px, calc(100% - 112px));
  flex: 1 0 auto;
  margin: 0 auto;
  padding: clamp(92px, 11vh, 114px) 0 clamp(20px, 3.5vh, 40px);
}
.auth-intro {
  max-width: 590px;
  padding: 22px 0 20px;
  animation: auth-intro-enter .72s cubic-bezier(.22, 1, .36, 1) both;
}
.auth-kicker {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: var(--auth-muted);
  font-size: .72rem;
  font-weight: 750;
  letter-spacing: .22em;
}
.auth-kicker::before {
  content: '';
  width: 28px;
  height: 2px;
  border-radius: 5px;
  background: linear-gradient(90deg, #3f9ef4, #7959ed);
}
.auth-intro h2 {
  margin: 32px 0 0;
  color: var(--auth-ink);
  font-size: clamp(3.1rem, 4.15vw, 4.7rem);
  font-weight: 790;
  line-height: 1.10;
  letter-spacing: -.055em;
}
.auth-intro h2 em {
  font-style: normal;
  color: #79adff;
  background: linear-gradient(115deg, #57b7f5 0%, #7191fc 47%, #a07bff 95%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.auth-page.is-light .auth-intro h2 em {
  background-image: linear-gradient(115deg, #2279dc 3%, #5668ec 48%, #9345e7 98%);
}
.auth-intro > p {
  max-width: 485px;
  margin: 21px 0 0;
  color: var(--auth-subtle);
  font-size: 1rem;
  font-weight: 450;
  line-height: 1.85;
  letter-spacing: .006em;
}
.auth-capabilities {
  width: min(500px, 100%);
  margin-top: 40px;
  border-top: 1px solid var(--auth-line);
}
.auth-capabilities > div {
  display: grid;
  grid-template-columns: 50px minmax(0, 1fr);
  align-items: center;
  gap: 20px;
  min-height: 96px;
  padding: 13px 7px 13px 0;
  border-bottom: 1px solid var(--auth-line);
  transition: padding-left .28s cubic-bezier(.22, 1, .36, 1);
}
@media (hover: hover) {
  .auth-capabilities > div:hover { padding-left: 5px; }
}
.auth-capabilities b {
  display: grid;
  width: 49px;
  height: 49px;
  place-items: center;
  border: 1px solid rgba(107, 152, 255, .22);
  border-radius: 17px;
  background: rgba(93, 126, 245, .10);
  color: #90aaff;
  font-size: 1.01rem;
  font-weight: 730;
  letter-spacing: -.03em;
}
.auth-capabilities > div:nth-child(2) b {
  border-color: rgba(164, 119, 247, .23);
  background: rgba(156, 111, 242, .10);
  color: #bc9bff;
}
.auth-capabilities > div:nth-child(3) b {
  border-color: rgba(66, 166, 207, .23);
  background: rgba(61, 172, 210, .10);
  color: #71c7e2;
}
.auth-page.is-light .auth-capabilities b { color: #396cdd; background: #eff4ff; }
.auth-page.is-light .auth-capabilities > div:nth-child(2) b { color: #8151d4; background: #f5f0ff; }
.auth-page.is-light .auth-capabilities > div:nth-child(3) b { color: #1683aa; background: #eefbff; }
.auth-capabilities span { display: flex; flex-direction: column; gap: 5px; min-width: 0; }
.auth-capabilities strong { color: var(--auth-ink); font-size: .94rem; font-weight: 710; }
.auth-capabilities small { color: var(--auth-subtle); font-size: .80rem; line-height: 1.55; }
.auth-card {
  position: relative;
  width: 100%;
  min-width: 0;
  /* app.css had margin: 12vh auto 0 (14vh on mobile) on this card.
   * That leftover margin was the main reason the redesigned view scrolled. */
  margin: 0;
  padding: clamp(27px, 3vw, 43px);
  border: 1px solid var(--auth-panel-line);
  border-radius: 25px;
  background: var(--auth-panel);
  box-shadow: 0 28px 88px rgba(0, 0, 0, .22), 0 5px 16px rgba(0, 0, 0, .09), inset 0 1px rgba(255, 255, 255, .08);
  backdrop-filter: blur(22px);
  -webkit-backdrop-filter: blur(22px);
  animation: auth-card-enter .72s .08s cubic-bezier(.22, 1, .36, 1) both;
}
.auth-page.is-light .auth-card {
  box-shadow: 0 28px 85px rgba(60, 93, 149, .12), 0 4px 14px rgba(61, 94, 154, .045), inset 0 1px #fff;
}
.auth-card-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  min-height: 16px;
  color: var(--auth-muted);
  font-size: .67rem;
  font-weight: 750;
  letter-spacing: .18em;
  white-space: nowrap;
}
.auth-card-meta i {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #5fc6a4;
  font-size: .60rem;
  font-weight: 740;
  font-style: normal;
  letter-spacing: .065em;
}
.auth-page.is-light .auth-card-meta i { color: #32856c; }
.auth-card-meta i b {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 50%;
  background: #40b994;
  box-shadow: 0 0 0 4px rgba(64, 185, 148, .10);
}
.auth-card header { margin-top: 31px; }
.auth-card header h1 {
  margin: 0;
  color: var(--auth-ink);
  font-size: clamp(1.65rem, 2.2vw, 2rem);
  font-weight: 770;
  line-height: 1.3;
  letter-spacing: -.045em;
}
.auth-card header p {
  margin: 10px 0 0;
  color: var(--auth-subtle);
  font-size: .89rem;
  line-height: 1.65;
}
.oauth-login { margin-top: 25px; }
.oauth-actions { display: grid; grid-template-columns: minmax(0,1fr) minmax(0,1fr); gap: 11px; }
.oauth-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  min-width: 0;
  height: 49px;
  padding: 0 9px;
  border: 1px solid var(--auth-field-line);
  border-radius: 13px;
  background: var(--auth-field);
  box-shadow: 0 2px 4px rgba(16, 32, 65, .025), inset 0 1px rgba(255, 255, 255, .07);
  color: var(--auth-ink);
  font: inherit;
  font-size: .78rem;
  font-weight: 650;
  white-space: nowrap;
  cursor: pointer;
  transition: border-color .22s ease, box-shadow .22s ease, background-color .22s ease, transform .22s ease;
}
.oauth-button:hover {
  border-color: #7f9fe0;
  box-shadow: 0 6px 16px rgba(76, 121, 211, .12);
  transform: translateY(-2px);
}
.auth-page.is-light .oauth-button:hover { background: #f8faff; }
.oauth-button:active { transform: translateY(0); box-shadow: none; }
.oauth-provider-mark {
  display: grid;
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  place-items: center;
  border-radius: 50%;
  background: rgba(255,255,255,.92);
  color: #212c42;
  box-shadow: 0 1px 4px rgba(29, 51, 104, .08);
}
.oauth-provider-mark svg { display: block; width: 16px; height: 16px; }
.oauth-provider-mark.github svg { width: 17px; height: 17px; }
.oauth-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 20px;
  color: var(--auth-muted);
  font-size: .72rem;
  white-space: nowrap;
}
.oauth-divider::before,
.oauth-divider::after { content: ''; height: 1px; flex: 1; background: var(--auth-line); }
.auth-card form { display: flex; flex-direction: column; gap: 17px; margin-top: 27px; }
.auth-card form.with-oauth { margin-top: 20px; }
.auth-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.auth-field > label {
  align-self: flex-start;
  padding-left: 1px;
  color: var(--auth-ink);
  font-size: .80rem;
  font-weight: 680;
  cursor: pointer;
}
.auth-field-control {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  height: 50px;
  padding: 0 13px;
  border: 1px solid var(--auth-field-line);
  border-radius: 12px;
  background: var(--auth-field);
  box-shadow: inset 0 1px rgba(255,255,255,.045);
  cursor: text;
  transition: border-color .19s ease, box-shadow .19s ease, background-color .19s ease;
}
.auth-field-control:hover { border-color: #829fd5; }
.auth-field-control:focus-within {
  border-color: var(--auth-focus);
  box-shadow: 0 0 0 3px rgba(87, 141, 244, .16), inset 0 1px rgba(255,255,255,.07);
}
.auth-field-icon {
  flex: 0 0 19px;
  width: 19px;
  height: 19px;
  margin-right: 11px;
  color: var(--auth-muted);
}
.auth-card input {
  width: 100%;
  min-width: 0;
  height: 100%;
  padding: 0;
  border: 0;
  outline: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  color: var(--auth-ink);
  font: inherit;
  font-size: .88rem;
  font-weight: 500;
  letter-spacing: 0;
}
.auth-card input::placeholder { color: var(--auth-muted); opacity: .78; }
.auth-card input:focus { outline: none; }
.auth-card input:-webkit-autofill {
  -webkit-text-fill-color: var(--auth-ink);
  -webkit-box-shadow: 0 0 0 100px transparent inset;
  transition: background-color 5000s ease-in-out 0s;
}
.auth-reveal {
  flex: 0 0 34px;
  display: grid;
  width: 34px;
  height: 34px;
  margin: 0 -6px 0 5px;
  padding: 7px;
  place-items: center;
  border: 0;
  border-radius: 9px;
  background: transparent;
  color: var(--auth-muted);
  cursor: pointer;
  transition: color .2s ease, background-color .2s ease;
}
.auth-reveal svg { display: block; width: 19px; height: 19px; }
.auth-reveal:hover { color: var(--auth-link); background: rgba(122, 147, 212, .10); }
.turnstile-official {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  min-height: 65px;
}
.turnstile-widget {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 65px;
}
.turnstile-widget:empty { min-height: 65px; }
.turnstile-widget :deep(iframe) { display: block; max-width: 100%; }
.turnstile-error { margin: 0; color: #d16f7b; font-size: .74rem; line-height: 1.5; }
.auth-submit {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  min-height: 52px;
  margin-top: 2px;
  padding: 0 17px;
  border: 1px solid rgba(160, 181, 255, .27);
  border-radius: 12px;
  background: linear-gradient(105deg, #298ee7 0%, #416feb 50%, #704bf1 100%);
  box-shadow: 0 8px 20px rgba(68, 111, 235, .21), inset 0 1px rgba(255,255,255,.28);
  color: #fff;
  font: inherit;
  font-size: .90rem;
  font-weight: 750;
  cursor: pointer;
  transition: transform .22s ease, box-shadow .22s ease, filter .22s ease;
}
.auth-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 13px 27px rgba(78, 103, 233, .28), inset 0 1px rgba(255,255,255,.3);
  filter: brightness(1.06);
}
.auth-submit:active:not(:disabled) { transform: translateY(0); }
.auth-submit:disabled { cursor: wait; opacity: .58; box-shadow: none; }
.auth-submit b { font-size: 1.05rem; font-weight: 500; transition: transform .22s ease; }
.auth-submit:hover:not(:disabled) b { transform: translateX(4px); }
.form-error,
.form-success {
  margin: 0;
  padding: 11px 13px;
  border: 1px solid rgba(216, 107, 121, .28);
  border-radius: 11px;
  background: rgba(212, 88, 112, .09);
  color: #e99da6;
  font-size: .76rem;
  line-height: 1.5;
}
.form-success {
  border-color: rgba(77, 177, 136, .28);
  background: rgba(66, 169, 126, .09);
  color: #70cf9d;
}
.auth-page.is-light .form-error { color: #ac3454; }
.auth-page.is-light .form-success { color: #247950; }
.auth-card footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 22px;
  color: var(--auth-muted);
  font-size: .77rem;
}
.auth-card footer a { color: var(--auth-link); text-decoration: none; transition: color .2s ease; }
.auth-card footer a:hover { color: var(--auth-ink); }
.auth-card footer span a { margin-left: 4px; font-weight: 700; }
.auth-page-footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 clamp(24px, 3.7vw, 60px) 22px;
  color: var(--auth-muted);
  font-size: .65rem;
  letter-spacing: .045em;
}
@keyframes auth-intro-enter {
  from { opacity: 0; transform: translate3d(0, 15px, 0); }
  to { opacity: 1; transform: translate3d(0, 0, 0); }
}
@keyframes auth-card-enter {
  from { opacity: 0; transform: translate3d(0, 18px, 0) scale(.989); }
  to { opacity: 1; transform: translate3d(0, 0, 0) scale(1); }
}
@media (max-width: 1100px) {
  .auth-layout { width: min(1050px, calc(100% - 80px)); gap: 42px; grid-template-columns: minmax(0, 1fr) minmax(402px, 452px); }
  .auth-intro h2 { font-size: clamp(2.65rem, 4.4vw, 3.5rem); }
}
@media (max-width: 920px) {
  .auth-page::after { width: 90vw; right: -42%; top: 10%; }
  .auth-layout { display: flex; width: min(480px, calc(100% - 48px)); justify-content: center; padding-top: 117px; padding-bottom: 58px; }
  .auth-intro { display: none; }
  .auth-card { padding: 33px; }
}
@media (max-width: 520px) {
  .auth-brand { top: 17px; left: 20px; min-height: 44px; gap: 9px; }
  .auth-brand img { width: 48px; height: 39px; }
  .auth-brand strong { font-size: 1.13rem; }
  .auth-brand small { font-size: .57rem; }
  .auth-layout { width: calc(100% - 28px); padding-top: 92px; padding-bottom: 34px; align-items: flex-start; }
  .auth-card { padding: 26px 21px 25px; border-radius: 21px; }
  .auth-card header { margin-top: 25px; }
  .auth-card header h1 { font-size: 1.7rem; }
  .auth-card-meta { font-size: .59rem; }
  .auth-card-meta i { font-size: .54rem; }
  .oauth-actions { grid-template-columns: 1fr; gap: 9px; }
  .oauth-login { margin-top: 22px; }
  .auth-card form { gap: 15px; }
  .auth-page-footer { flex-wrap: wrap; justify-content: center; padding: 0 20px 18px; font-size: .61rem; text-align: center; }
}
@media (max-height: 760px) and (min-width: 921px) {
  .auth-layout { padding-top: 96px; padding-bottom: 30px; }
  .auth-card { padding-top: 28px; padding-bottom: 28px; }
  .auth-card header { margin-top: 22px; }
  .auth-capabilities { margin-top: 28px; }
  .auth-capabilities > div { min-height: 80px; }
}
/* Laptop/short-window variant: preserve the two-column composition and fit
 * the full register form on a normal desktop viewport (including Turnstile).
 * Emergency overflow stays available for browser zoom, errors and tiny screens. */
@media (max-height: 950px) {
  .auth-brand { top: clamp(18px, 2.7vh, 27px); }
  .auth-layout {
    padding-top: clamp(80px, 10vh, 94px);
    padding-bottom: clamp(14px, 2.4vh, 23px);
  }
  .auth-intro { padding: 8px 0; }
  .auth-intro h2 {
    margin-top: 19px;
    font-size: clamp(2.65rem, 3.7vw, 3.75rem);
  }
  .auth-intro > p { margin-top: 14px; line-height: 1.68; }
  .auth-capabilities { margin-top: 24px; }
  .auth-capabilities > div { min-height: 75px; padding-top: 9px; padding-bottom: 9px; }
  .auth-capabilities b { width: 45px; height: 45px; border-radius: 15px; }
  .auth-card { padding: clamp(23px, 3vh, 30px); }
  .auth-card header { margin-top: 20px; }
  .auth-card header p { margin-top: 6px; line-height: 1.5; }
  .oauth-login { margin-top: 16px; }
  .oauth-button { height: 45px; }
  .oauth-divider { margin-top: 12px; }
  .auth-card form { gap: 11px; margin-top: 18px; }
  .auth-card form.with-oauth { margin-top: 14px; }
  .auth-field { gap: 6px; }
  .auth-field-control { height: 44px; }
  .auth-submit { min-height: 47px; }
  .auth-card footer { margin-top: 15px; }
  .auth-page-footer { padding-bottom: 13px; }
}
@media (max-width: 920px) and (max-height: 950px) {
  .auth-layout { padding-top: clamp(78px, 9.5vh, 90px); padding-bottom: 17px; }
}
@media (max-width: 520px) and (max-height: 950px) {
  .auth-layout { padding-top: 82px; padding-bottom: 12px; }
  .auth-card { padding: 22px 20px; }
}
@media (prefers-reduced-motion: reduce) {
  .auth-intro,
  .auth-card { animation: none; }
  .auth-page *, .auth-page *::before, .auth-page *::after { transition-duration: .01ms !important; }
}
</style>
