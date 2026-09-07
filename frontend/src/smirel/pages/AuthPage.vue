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
const submitLabels: Record<string, string> = {
  login: '登录',
  register: '创建账户',
  forgot: '发送重置链接',
  reset: '更新密码',
}
const title = computed(() => titles[kind.value] || titles.login)
const subtitle = computed(() => subtitles[kind.value] || '')
const submitLabel = computed(() => submitLabels[kind.value] || submitLabels.login)

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
    <RouterLink to="/home" class="auth-brand brand-link">
      <img :src="logoUrl" alt="Smirel" />
      <span>
        <strong>Smirel</strong>
        <small>API SERVICE</small>
      </span>
    </RouterLink>

    <main class="auth-layout">
      <section class="auth-intro" aria-label="Smirel Console">
        <span class="auth-kicker">SMIREL CONSOLE</span>
        <h2>统一管理你的<br />API 工作区。</h2>
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
          <span>SMIREL ACCOUNT</span>
          <i><b></b>SECURE ACCESS</i>
        </div>

        <header>
          <h1>{{ title }}</h1>
          <p>{{ subtitle }}</p>
        </header>

        <form @submit.prevent="submit">
          <label>
            <span>邮箱</span>
            <input
              v-model="email"
              type="email"
              autocomplete="email"
              required
              placeholder="name@example.com"
            />
          </label>

          <label v-if="kind !== 'forgot'">
            <span>密码</span>
            <input
              v-model="password"
              type="password"
              :autocomplete="kind === 'login' ? 'current-password' : 'new-password'"
              required
              placeholder="••••••••"
            />
          </label>

          <label v-if="kind === 'register'">
            <span>确认密码</span>
            <input
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
              required
              placeholder="••••••••"
            />
          </label>

          <label v-if="kind === 'reset'">
            <span>重置令牌</span>
            <input v-model="token" type="text" required placeholder="Reset token" />
          </label>

          <p v-if="error" class="form-error">{{ error }}</p>
          <p v-if="message" class="form-success">{{ message }}</p>

          <button class="auth-submit" type="submit" :disabled="loading">
            <span>{{ loading ? '处理中…' : (previewMode ? '进入预览控制台' : submitLabel) }}</span>
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
      <span>Smirel · API Service</span>
      <span>Console Access</span>
    </footer>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  padding: 0;
  color: #f4f7fa;
  background:
    radial-gradient(circle at 74% 38%, rgba(47, 113, 164, .10), transparent 30%),
    radial-gradient(circle at 18% 82%, rgba(42, 76, 106, .06), transparent 34%),
    #07090d;
}

.auth-page::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, transparent 0 49.96%, rgba(255,255,255,.018) 50%, transparent 50.04%),
    linear-gradient(180deg, rgba(255,255,255,.014), transparent 18%);
  opacity: .65;
}

.auth-brand {
  position: absolute;
  z-index: 3;
  top: 34px;
  left: 40px;
  gap: 11px;
}

.auth-brand img {
  width: 39px;
  height: 39px;
}

.auth-brand > span {
  gap: 2px;
}

.auth-brand strong {
  color: #f4f6f8;
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -.025em;
}

.auth-brand small {
  color: #6f7a86;
  font-size: .60rem;
  font-weight: 650;
  letter-spacing: .16em;
}

.auth-layout {
  position: relative;
  z-index: 2;
  width: min(1120px, calc(100vw - 96px));
  min-height: 100vh;
  margin: 0 auto;
  padding: 118px 0 92px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 440px;
  gap: 92px;
  align-items: center;
}

.auth-intro {
  max-width: 530px;
  padding: 14px 0 18px;
}

.auth-kicker {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  color: #708090;
  font-size: .69rem;
  font-weight: 700;
  letter-spacing: .16em;
}

.auth-kicker::before {
  content: '';
  width: 22px;
  height: 1px;
  background: #4385b9;
}

.auth-intro h2 {
  margin: 22px 0 0;
  color: #f1f4f7;
  font-size: clamp(2.7rem, 4.2vw, 3.7rem);
  font-weight: 620;
  line-height: 1.05;
  letter-spacing: -.055em;
}

.auth-intro > p {
  max-width: 470px;
  margin: 20px 0 0;
  color: #8a949f;
  font-size: .95rem;
  line-height: 1.7;
}

.auth-capabilities {
  width: min(440px, 100%);
  margin-top: 42px;
  border-top: 1px solid #20262d;
}

.auth-capabilities > div {
  min-height: 74px;
  display: grid;
  grid-template-columns: 42px 1fr;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid #20262d;
}

.auth-capabilities b {
  color: #56616c;
  font: 650 .69rem/1 ui-monospace, SFMono-Regular, Menlo, monospace;
}

.auth-capabilities span {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.auth-capabilities strong {
  color: #cbd1d7;
  font-size: .84rem;
  font-weight: 650;
}

.auth-capabilities small {
  color: #68737e;
  font-size: .73rem;
  line-height: 1.45;
}

.auth-card {
  width: 100%;
  margin: 0;
  padding: 34px 34px 28px;
  border: 1px solid #252c34;
  border-radius: 16px;
  background: #0d1015;
  box-shadow: 0 26px 80px rgba(0,0,0,.30), inset 0 1px rgba(255,255,255,.018);
}

.auth-card-meta {
  min-height: 22px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: #65717d;
  font-size: .66rem;
  font-weight: 700;
  letter-spacing: .13em;
}

.auth-card-meta i {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #657b70;
  font-style: normal;
  font-size: .60rem;
  letter-spacing: .08em;
}

.auth-card-meta i b {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #53bf8e;
  box-shadow: 0 0 10px rgba(83,191,142,.22);
}

.auth-card header {
  margin-top: 24px;
}

.auth-card header h1 {
  margin: 0;
  color: #f5f7f9;
  font-size: 2rem;
  font-weight: 670;
  line-height: 1.12;
  letter-spacing: -.045em;
}

.auth-card header p {
  margin: 10px 0 0;
  color: #8a949f;
  font-size: .86rem;
  line-height: 1.6;
}

.auth-card form {
  display: flex;
  flex-direction: column;
  gap: 17px;
  margin-top: 29px;
}

.auth-card label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: #a6afb8;
  font-size: .77rem;
  font-weight: 620;
}

.auth-card label > span {
  padding-left: 1px;
}

.auth-card input {
  width: 100%;
  height: 50px;
  padding: 0 14px;
  border: 1px solid #29313a;
  border-radius: 10px;
  outline: none;
  background: #090c10;
  color: #edf1f4;
  font-size: .90rem;
  transition: border-color .16s ease, background-color .16s ease, box-shadow .16s ease;
}

.auth-card input::placeholder {
  color: #4d5660;
}

.auth-card input:hover {
  border-color: #35404b;
}

.auth-card input:focus {
  border-color: #3a6f9b;
  background: #0b0f14;
  box-shadow: 0 0 0 3px rgba(60,126,178,.10);
}

.auth-submit {
  width: 100%;
  height: 50px;
  margin-top: 3px;
  padding: 0 16px;
  border: 1px solid #dce2e7;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  background: #eef2f5;
  color: #101419;
  font: inherit;
  font-size: .86rem;
  font-weight: 700;
  cursor: pointer;
  transition: background-color .16s ease, border-color .16s ease, transform .16s ease;
}

.auth-submit:hover:not(:disabled) {
  border-color: #fff;
  background: #fff;
  transform: translateY(-1px);
}

.auth-submit:disabled {
  cursor: wait;
  opacity: .58;
}

.auth-submit b {
  font-size: .95rem;
  font-weight: 500;
  transition: transform .16s ease;
}

.auth-submit:hover:not(:disabled) b {
  transform: translateX(2px);
}

.form-error,
.form-success {
  margin: -2px 0 0;
  padding: 10px 12px;
  border-radius: 8px;
  font-size: .76rem;
  line-height: 1.5;
}

.auth-card footer {
  margin-top: 21px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: #737d87;
  font-size: .74rem;
}

.auth-card footer a {
  color: #aab4be;
  transition: color .15s ease;
}

.auth-card footer a:hover {
  color: #f1f4f6;
}

.auth-card footer span a {
  margin-left: 3px;
  color: #7fb8e1;
}

.auth-page-footer {
  position: absolute;
  z-index: 2;
  left: 40px;
  right: 40px;
  bottom: 27px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #444e58;
  font-size: .64rem;
  letter-spacing: .04em;
}

@media (max-width: 980px) {
  .auth-layout {
    width: min(460px, calc(100vw - 48px));
    grid-template-columns: 1fr;
    gap: 0;
    padding-top: 118px;
  }

  .auth-intro {
    display: none;
  }
}

@media (max-width: 560px) {
  .auth-page {
    overflow: auto;
  }

  .auth-brand {
    top: 22px;
    left: 22px;
  }

  .auth-brand img {
    width: 36px;
    height: 36px;
  }

  .auth-layout {
    width: calc(100vw - 28px);
    min-height: 100vh;
    padding: 96px 0 78px;
    align-items: start;
  }

  .auth-card {
    padding: 27px 22px 24px;
    border-radius: 14px;
  }

  .auth-card-meta {
    font-size: .61rem;
  }

  .auth-card-meta i {
    font-size: .56rem;
  }

  .auth-card header {
    margin-top: 21px;
  }

  .auth-card header h1 {
    font-size: 1.78rem;
  }

  .auth-card form {
    margin-top: 25px;
  }

  .auth-card footer {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
  }

  .auth-page-footer {
    left: 22px;
    right: 22px;
    bottom: 20px;
    font-size: .59rem;
  }
}
</style>
