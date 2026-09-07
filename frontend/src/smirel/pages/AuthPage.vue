<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getErrorMessage, previewMode } from '../core/api'
import { useSession } from '../core/session'

interface PublicAuthSettings {
  github_oauth_enabled?: boolean
  google_oauth_enabled?: boolean
}

type OAuthProvider = 'github' | 'google'

const route = useRoute()
const router = useRouter()
const { login, register, isAdmin } = useSession()
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const token = ref(String(route.query.token || ''))
const loading = ref(false)
const oauthLoading = ref<OAuthProvider | ''>('')
const message = ref('')
const error = ref('')
const oauthSettings = ref<PublicAuthSettings>({})
const kind = computed(() => String(route.meta.authKind || 'login'))
const titles: Record<string, string> = {
  login: '登录 Smirel',
  register: '创建 Smirel 账户',
  forgot: '找回密码',
  reset: '设置新密码',
}
const subtitles: Record<string, string> = {
  login: '继续进入你的 API 工作区。',
  register: '一个账户管理密钥、用量和服务。',
  forgot: '输入邮箱，我们会发送重置链接。',
  reset: '为你的账户设置新的密码。',
}
const title = computed(() => titles[kind.value] || titles.login)
const subtitle = computed(() => subtitles[kind.value] || '')
const showOAuth = computed(() => (kind.value === 'login' || kind.value === 'register') && (
  oauthSettings.value.github_oauth_enabled || oauthSettings.value.google_oauth_enabled
))

function requestedRedirect() {
  const value = typeof route.query.redirect === 'string' ? route.query.redirect.trim() : ''
  return value.startsWith('/') && !value.startsWith('//') ? value : '/dashboard'
}

async function loadOAuthSettings() {
  if (previewMode) {
    oauthSettings.value = { github_oauth_enabled: true, google_oauth_enabled: true }
    return
  }
  try {
    const { data } = await api.get<PublicAuthSettings>('/settings/public')
    oauthSettings.value = data || {}
  } catch {
    // Email/password auth must remain usable when public settings are temporarily unavailable.
    oauthSettings.value = {}
  }
}

async function startOAuth(provider: OAuthProvider) {
  error.value = ''
  if (previewMode) {
    await router.push('/admin/dashboard')
    return
  }

  oauthLoading.value = provider
  const base = String(api.defaults.baseURL || '/api/v1').replace(/\/+$/, '')
  const params = new URLSearchParams({ redirect: requestedRedirect() })
  window.location.assign(`${base}/auth/oauth/${provider}/start?${params.toString()}`)
}

async function submit() {
  error.value = ''
  message.value = ''
  if (kind.value === 'register' && password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    if (previewMode) {
      await router.push('/admin/dashboard')
      return
    }

    if (kind.value === 'login') {
      await login(email.value.trim(), password.value)
      const redirect = typeof route.query.redirect === 'string'
        ? route.query.redirect
        : (isAdmin.value ? '/admin/dashboard' : '/dashboard')
      await router.push(redirect)
    } else if (kind.value === 'register') {
      await register(email.value.trim(), password.value)
      await router.push('/dashboard')
    } else if (kind.value === 'forgot') {
      await api.post('/auth/forgot-password', { email: email.value.trim() })
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
  }
}

onMounted(loadOAuthSettings)
</script>

<template>
  <div class="auth-page">
    <div class="site-environment" aria-hidden="true"></div>
    <RouterLink to="/home" class="auth-brand brand-link">
      <img :src="logoUrl" alt="Smirel" />
      <span><strong>Smirel</strong><small>API SERVICE</small></span>
    </RouterLink>
    <main class="auth-card glass">
      <header>
        <span class="eyebrow">SMIREL ACCOUNT</span>
        <h1>{{ title }}</h1>
        <p>{{ subtitle }}</p>
      </header>

      <div v-if="showOAuth" class="oauth-entry">
        <button v-if="oauthSettings.google_oauth_enabled" type="button" class="oauth-button" :disabled="Boolean(oauthLoading)" @click="startOAuth('google')">
          <span class="oauth-mark google-mark">G</span>
          <span>{{ oauthLoading === 'google' ? '正在前往 Google…' : '使用 Google 继续' }}</span>
        </button>
        <button v-if="oauthSettings.github_oauth_enabled" type="button" class="oauth-button" :disabled="Boolean(oauthLoading)" @click="startOAuth('github')">
          <span class="oauth-mark github-mark" aria-hidden="true">
            <svg viewBox="0 0 24 24" role="img"><path fill="currentColor" d="M12 .7a11.5 11.5 0 0 0-3.64 22.41c.58.11.79-.25.79-.56v-2.23c-3.22.7-3.9-1.37-3.9-1.37-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.71.08-.71 1.17.08 1.78 1.2 1.78 1.2 1.04 1.78 2.72 1.27 3.38.97.1-.75.41-1.27.74-1.56-2.57-.29-5.27-1.29-5.27-5.69 0-1.26.45-2.29 1.19-3.1-.12-.29-.52-1.47.11-3.06 0 0 .97-.31 3.16 1.18A11 11 0 0 1 12 6.09c.98 0 1.96.13 2.88.39 2.19-1.49 3.15-1.18 3.15-1.18.63 1.59.23 2.77.11 3.06.74.81 1.19 1.84 1.19 3.1 0 4.41-2.71 5.39-5.29 5.68.42.36.79 1.06.79 2.14v3.27c0 .31.21.68.8.56A11.5 11.5 0 0 0 12 .7Z"/></svg>
          </span>
          <span>{{ oauthLoading === 'github' ? '正在前往 GitHub…' : '使用 GitHub 继续' }}</span>
        </button>
        <div class="oauth-divider"><span>或使用邮箱</span></div>
      </div>

      <form @submit.prevent="submit">
        <label>邮箱<input v-model="email" type="email" autocomplete="email" required placeholder="name@example.com" /></label>
        <label v-if="kind !== 'forgot'">密码<input v-model="password" type="password" :autocomplete="kind === 'login' ? 'current-password' : 'new-password'" required placeholder="••••••••" /></label>
        <label v-if="kind === 'register'">确认密码<input v-model="confirmPassword" type="password" autocomplete="new-password" required placeholder="••••••••" /></label>
        <label v-if="kind === 'reset'">重置令牌<input v-model="token" type="text" required placeholder="Reset token" /></label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <p v-if="message" class="form-success">{{ message }}</p>
        <button class="primary-button auth-submit" type="submit" :disabled="loading || Boolean(oauthLoading)">
          {{ loading ? '处理中…' : (previewMode ? '进入预览控制台' : title) }}
        </button>
      </form>
      <footer v-if="kind === 'login'">
        <RouterLink to="/forgot-password">忘记密码？</RouterLink>
        <span>没有账户？ <RouterLink to="/register">注册</RouterLink></span>
      </footer>
      <footer v-else><RouterLink to="/login">← 返回登录</RouterLink></footer>
    </main>
  </div>
</template>

<style scoped>
.oauth-entry{display:flex;flex-direction:column;gap:8px;margin-top:20px}.oauth-button{width:100%;height:40px;border:1px solid rgba(255,255,255,.10);border-radius:7px;background:rgba(255,255,255,.045);color:rgba(255,255,255,.88);display:flex;align-items:center;justify-content:center;gap:9px;cursor:pointer;transition:.18s ease;font-size:.68rem}.oauth-button:hover{background:rgba(255,255,255,.085);border-color:rgba(255,255,255,.16);transform:translateY(-1px)}.oauth-button:disabled{opacity:.55;cursor:wait;transform:none}.oauth-mark{width:18px;height:18px;display:grid;place-items:center;font-weight:700;font-size:.78rem}.github-mark svg{width:17px;height:17px}.google-mark{font-family:Arial,sans-serif;font-size:.84rem}.oauth-divider{height:20px;display:flex;align-items:center;gap:10px;color:rgba(255,255,255,.34);font-size:.55rem}.oauth-divider:before,.oauth-divider:after{content:"";height:1px;flex:1;background:rgba(255,255,255,.07)}
</style>
