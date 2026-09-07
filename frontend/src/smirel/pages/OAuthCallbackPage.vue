<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getErrorMessage } from '../core/api'
import { useSession, type OAuthTokenResult } from '../core/session'

interface PendingOAuthPayload {
  access_token?: string
  refresh_token?: string
  expires_in?: number
  step?: string
  error?: string
  provider?: string
  email?: string
  resolved_email?: string
  redirect?: string
  invitation_required?: boolean
  create_account_allowed?: boolean
  existing_account_bindable?: boolean
}

const route = useRoute()
const router = useRouter()
const { loginWithOAuthTokens } = useSession()
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`

const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const pending = ref<PendingOAuthPayload | null>(null)
const password = ref('')
const confirmPassword = ref('')
const invitationCode = ref('')

const providerName = computed(() => {
  const provider = String(pending.value?.provider || '').toLowerCase()
  if (provider === 'github') return 'GitHub'
  if (provider === 'google') return 'Google'
  return '第三方账户'
})

const pendingEmail = computed(() => pending.value?.resolved_email || pending.value?.email || '')
const canCompleteRegistration = computed(() => {
  const provider = String(pending.value?.provider || '').toLowerCase()
  return pending.value?.create_account_allowed !== false && (provider === 'github' || provider === 'google')
})

function safeRedirect(value?: string | null) {
  const redirect = String(value || '').trim()
  return redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/dashboard'
}

function parseTokenPayload(params: URLSearchParams): OAuthTokenResult | null {
  const accessToken = params.get('access_token') || ''
  if (!accessToken) return null
  const expires = Number(params.get('expires_in') || 0)
  return {
    access_token: accessToken,
    refresh_token: params.get('refresh_token') || undefined,
    expires_in: Number.isFinite(expires) && expires > 0 ? expires : undefined,
  }
}

function clearSensitiveFragment() {
  if (!window.location.hash) return
  window.history.replaceState({}, document.title, `${window.location.pathname}${window.location.search}`)
}

async function finishLogin(tokens: OAuthTokenResult, redirect?: string | null) {
  await loginWithOAuthTokens(tokens)
  await router.replace(safeRedirect(redirect))
}

async function resolveCallback() {
  loading.value = true
  error.value = ''
  try {
    const fragment = new URLSearchParams(window.location.hash.replace(/^#/, ''))
    const providerError = fragment.get('error') || (typeof route.query.error === 'string' ? route.query.error : '')
    if (providerError) {
      const detail = fragment.get('error_description') || (typeof route.query.error_description === 'string' ? route.query.error_description : '')
      throw new Error(detail || providerError)
    }

    const fragmentTokens = parseTokenPayload(fragment)
    const fragmentRedirect = fragment.get('redirect')
    if (fragmentTokens) {
      clearSensitiveFragment()
      await finishLogin(fragmentTokens, fragmentRedirect)
      return
    }

    // New-account and identity-adoption flows are browser-bound by HttpOnly cookies.
    // Exchange the pending session instead of exposing provider credentials to the UI.
    const { data } = await api.post<PendingOAuthPayload>('/auth/oauth/pending/exchange', {})
    if (data.access_token) {
      await finishLogin({
        access_token: data.access_token,
        refresh_token: data.refresh_token,
        expires_in: data.expires_in,
      }, data.redirect)
      return
    }
    pending.value = data
  } catch (caught) {
    error.value = getErrorMessage(caught)
  } finally {
    loading.value = false
  }
}

async function completeRegistration() {
  error.value = ''
  if (password.value.length < 6) {
    error.value = '密码至少需要 6 位'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  const provider = String(pending.value?.provider || '').toLowerCase()
  if (provider !== 'github' && provider !== 'google') {
    error.value = '当前第三方登录需要返回登录页继续处理'
    return
  }

  submitting.value = true
  try {
    const { data } = await api.post<OAuthTokenResult>(`/auth/oauth/${provider}/complete-registration`, {
      password: password.value,
      invitation_code: invitationCode.value.trim() || undefined,
    })
    await finishLogin(data, pending.value?.redirect)
  } catch (caught) {
    error.value = getErrorMessage(caught)
  } finally {
    submitting.value = false
  }
}

onMounted(resolveCallback)
</script>

<template>
  <div class="auth-page oauth-callback-page">
    <div class="site-environment" aria-hidden="true"></div>
    <RouterLink to="/home" class="auth-brand brand-link">
      <img :src="logoUrl" alt="Smirel" />
      <span><strong>Smirel</strong><small>API SERVICE</small></span>
    </RouterLink>

    <main class="auth-card glass oauth-callback-card">
      <header>
        <span class="eyebrow">SECURE SIGN IN</span>
        <h1>{{ pending ? `继续使用 ${providerName}` : '正在完成登录' }}</h1>
        <p v-if="loading">正在验证第三方身份并建立 Smirel 会话…</p>
        <p v-else-if="pending && canCompleteRegistration">
          {{ pendingEmail ? `${pendingEmail} 已通过 ${providerName} 验证。` : `${providerName} 身份已验证。` }} 请设置 Smirel 登录密码完成账户创建。
        </p>
        <p v-else-if="!error">登录信息已经返回，正在处理账户状态。</p>
      </header>

      <div v-if="loading" class="oauth-progress" aria-live="polite">
        <span></span><span></span><span></span>
      </div>

      <form v-else-if="pending && canCompleteRegistration" @submit.prevent="completeRegistration">
        <label v-if="pendingEmail">已验证邮箱<input :value="pendingEmail" type="email" disabled /></label>
        <label>设置密码<input v-model="password" type="password" autocomplete="new-password" required minlength="6" placeholder="至少 6 位" /></label>
        <label>确认密码<input v-model="confirmPassword" type="password" autocomplete="new-password" required minlength="6" placeholder="再次输入密码" /></label>
        <label v-if="pending.invitation_required">邀请码<input v-model="invitationCode" type="text" required placeholder="请输入邀请码" /></label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <button class="primary-button auth-submit" type="submit" :disabled="submitting">
          {{ submitting ? '正在创建账户…' : '完成注册并登录' }}
        </button>
      </form>

      <div v-else-if="error || pending" class="oauth-result">
        <p class="form-error">{{ error || '该第三方账户还需要额外的账户处理步骤。' }}</p>
        <RouterLink to="/login" class="primary-button">返回登录</RouterLink>
      </div>
    </main>
  </div>
</template>

<style scoped>
.oauth-callback-card{min-height:270px}.oauth-progress{height:86px;display:flex;align-items:center;justify-content:center;gap:8px}.oauth-progress span{width:7px;height:7px;border-radius:50%;background:rgba(255,255,255,.72);animation:oauth-pulse 1s ease-in-out infinite}.oauth-progress span:nth-child(2){animation-delay:.14s}.oauth-progress span:nth-child(3){animation-delay:.28s}.oauth-result{display:flex;flex-direction:column;gap:12px;margin-top:20px}.oauth-result .primary-button{width:100%}.auth-card input:disabled{opacity:.68;cursor:not-allowed}@keyframes oauth-pulse{0%,100%{opacity:.25;transform:translateY(0)}50%{opacity:1;transform:translateY(-3px)}}
</style>
