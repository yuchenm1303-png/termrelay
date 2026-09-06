<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getErrorMessage, previewMode } from '../core/api'
import { useSession } from '../core/session'

const route = useRoute()
const router = useRouter()
const { login, register, isAdmin } = useSession()
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const token = ref(String(route.query.token || ''))
const loading = ref(false)
const message = ref('')
const error = ref('')
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
      <form @submit.prevent="submit">
        <label>邮箱<input v-model="email" type="email" autocomplete="email" required placeholder="name@example.com" /></label>
        <label v-if="kind !== 'forgot'">密码<input v-model="password" type="password" :autocomplete="kind === 'login' ? 'current-password' : 'new-password'" required placeholder="••••••••" /></label>
        <label v-if="kind === 'register'">确认密码<input v-model="confirmPassword" type="password" autocomplete="new-password" required placeholder="••••••••" /></label>
        <label v-if="kind === 'reset'">重置令牌<input v-model="token" type="text" required placeholder="Reset token" /></label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <p v-if="message" class="form-success">{{ message }}</p>
        <button class="primary-button auth-submit" type="submit" :disabled="loading">
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
