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
            <router-link to="/model-plaza">模型价格</router-link>
            <router-link to="/key-usage">用量查询</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">接入文档</a>
          </nav>
          <span class="smh-status"><i></i>服务正常</span>
          <router-link :to="accountPath" class="smh-login-link">
            {{ isAuthenticated ? '进入控制台' : '登录' }}
          </router-link>
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

        <section id="access" class="smh-glass smh-access-bar" aria-label="API 接入信息">
          <div class="smh-access-main">
            <span>BASE URL</span>
            <code>{{ apiBase }}</code>
          </div>
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
          <button type="button" class="smh-copy-button" @click="copyBaseUrl">
            {{ copied ? '已复制' : '复制地址' }}
          </button>
        </section>

        <section class="smh-glass smh-process-card">
          <div class="smh-process-copy">
            <p class="smh-section-label">开始使用</p>
            <h2>三步接入。</h2>
            <p>注册、创建 Key、替换地址。其他调用方式不用改。</p>
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
              <p>给项目生成一把密钥。</p>
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
            <p>看当前可用模型、单价和服务状态。</p>
            <b>查看 →</b>
          </router-link>

          <router-link :to="keyPath" class="smh-glass smh-quick-card">
            <span class="smh-quick-label">API KEY</span>
            <h3>密钥</h3>
            <p>创建和管理项目密钥。</p>
            <b>{{ isAuthenticated ? '管理 →' : '登录 →' }}</b>
          </router-link>

          <router-link to="/key-usage" class="smh-glass smh-quick-card">
            <span class="smh-quick-label">USAGE</span>
            <h3>用量</h3>
            <p>查看请求、Token 和费用。</p>
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
import { computed, onMounted, ref } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-shared-glass-v1.css'

const appStore = useAppStore()
const authStore = useAuthStore()
const copied = ref(false)

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

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
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
  padding: 24px 0 34px;
}

.smh-header {
  min-height: 58px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 28px;
}

.smh-header-right,
.smh-nav,
.smh-status {
  display: flex;
  align-items: center;
}

.smh-header-right {
  gap: 18px;
}

.smh-nav {
  gap: 22px;
}

.smh-nav a,
.smh-status,
.smh-login-link {
  color: rgba(255, 255, 255, .68);
  font-size: .72rem;
}

.smh-nav a,
.smh-login-link {
  transition: color .18s ease, background-color .18s ease;
}

.smh-nav a:hover,
.smh-login-link:hover {
  color: rgba(255, 255, 255, .98);
}

.smh-status {
  gap: 7px;
  white-space: nowrap;
}

.smh-status i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #83efc7;
  box-shadow: 0 0 12px rgba(131, 239, 199, .42);
}

.smh-login-link {
  min-height: 36px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, .10);
  border-radius: 999px;
  background: rgba(6, 16, 25, .16);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
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
.smh-quick-label {
  margin: 0;
  color: rgba(255, 255, 255, .44);
  font-size: .64rem;
  letter-spacing: .15em;
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
  color: rgba(255, 255, 255, .72);
  font-size: .95rem;
  line-height: 1.74;
}

.smh-hero-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 27px;
}

.smh-button {
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 17px;
  border-radius: 6px;
  font-size: .78rem;
  font-weight: 620;
  transition: background-color .18s ease, border-color .18s ease, transform .18s ease;
}

.smh-button:hover {
  transform: translateY(-1px);
}

.smh-button--primary {
  border: 1px solid rgba(255, 255, 255, .14);
  background: rgba(255, 255, 255, .16);
  color: rgba(255, 255, 255, .98);
  -webkit-backdrop-filter: blur(14px);
  backdrop-filter: blur(14px);
}

.smh-button--primary:hover {
  background: rgba(255, 255, 255, .22);
}

.smh-button--quiet {
  border: 1px solid rgba(255, 255, 255, .08);
  background: rgba(4, 10, 16, .14);
  color: rgba(255, 255, 255, .72);
}

.smh-button--quiet:hover {
  color: rgba(255, 255, 255, .96);
  border-color: rgba(255, 255, 255, .14);
}

.smh-glass {
  border: 1px solid rgba(255, 255, 255, .08);
  border-radius: 8px;
  background: rgba(5, 18, 29, .42);
  -webkit-backdrop-filter: blur(22px) saturate(118%);
  backdrop-filter: blur(22px) saturate(118%);
}

.smh-access-bar {
  display: grid;
  grid-template-columns: minmax(330px, 2fr) repeat(3, minmax(125px, .72fr)) auto;
  align-items: stretch;
  overflow: hidden;
}

.smh-access-main,
.smh-access-item {
  min-width: 0;
  min-height: 108px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 19px 22px;
}

.smh-access-item {
  border-left: 1px solid rgba(255, 255, 255, .07);
}

.smh-access-main span,
.smh-access-item span {
  color: rgba(255, 255, 255, .40);
  font-size: .60rem;
  letter-spacing: .10em;
}

.smh-access-main code {
  margin-top: 9px;
  overflow: hidden;
  color: rgba(255, 255, 255, .97);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 1rem;
  font-weight: 580;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smh-access-item strong {
  margin-top: 8px;
  color: rgba(255, 255, 255, .82);
  font-size: .72rem;
  font-weight: 610;
}

.smh-copy-button {
  align-self: center;
  min-width: 92px;
  min-height: 40px;
  margin: 0 18px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, .12);
  border-radius: 6px;
  background: rgba(255, 255, 255, .10);
  color: rgba(255, 255, 255, .86);
  font: inherit;
  font-size: .72rem;
  cursor: pointer;
  transition: background-color .18s ease, border-color .18s ease;
}

.smh-copy-button:hover {
  background: rgba(255, 255, 255, .16);
  border-color: rgba(255, 255, 255, .18);
}

.smh-process-card {
  display: grid;
  grid-template-columns: minmax(280px, .82fr) minmax(0, 1.68fr);
  gap: 0;
  margin-top: 18px;
  overflow: hidden;
}

.smh-process-copy {
  min-height: 260px;
  padding: 34px 34px 30px;
  border-right: 1px solid rgba(255, 255, 255, .07);
}

.smh-process-copy h2,
.smh-routing-copy h2 {
  margin: 10px 0 0;
  font-size: clamp(1.9rem, 3vw, 2.7rem);
  line-height: 1.1;
  font-weight: 560;
  letter-spacing: -.038em;
}

.smh-process-copy > p:last-of-type,
.smh-routing-copy > p:last-of-type {
  max-width: 520px;
  margin: 16px 0 0;
  color: rgba(255, 255, 255, .54);
  font-size: .82rem;
  line-height: 1.72;
}

.smh-text-link {
  display: inline-flex;
  margin-top: 30px;
  color: rgba(255, 255, 255, .78);
  font-size: .73rem;
  font-weight: 600;
}

.smh-text-link:hover {
  color: rgba(255, 255, 255, .98);
}

.smh-steps {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.smh-step {
  min-height: 260px;
  padding: 34px 30px 30px;
}

.smh-step + .smh-step {
  border-left: 1px solid rgba(255, 255, 255, .07);
}

.smh-step > span {
  color: rgba(255, 255, 255, .34);
  font-size: .62rem;
  letter-spacing: .10em;
}

.smh-step h3 {
  margin: 88px 0 0;
  font-size: 1.04rem;
  font-weight: 620;
}

.smh-step p {
  margin: 9px 0 0;
  color: rgba(255, 255, 255, .47);
  font-size: .75rem;
  line-height: 1.6;
}

.smh-routing-card {
  display: grid;
  grid-template-columns: minmax(0, .92fr) minmax(440px, 1.08fr);
  gap: 0;
  margin-top: 18px;
  overflow: hidden;
}

.smh-routing-copy {
  min-height: 360px;
  padding: 38px 38px 34px;
  border-right: 1px solid rgba(255, 255, 255, .07);
}

.smh-provider-line {
  display: flex;
  flex-wrap: wrap;
  gap: 22px;
  margin-top: 28px;
  color: rgba(255, 255, 255, .74);
  font-size: .77rem;
}

.smh-provider-line span + span::before {
  content: '/';
  margin-right: 22px;
  color: rgba(255, 255, 255, .20);
}

.smh-client-example {
  min-width: 0;
  margin: 22px;
  align-self: stretch;
  border: 1px solid rgba(255, 255, 255, .07);
  border-radius: 6px;
  background: rgba(0, 0, 0, .11);
  overflow: hidden;
}

.smh-client-example-head {
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
  border-bottom: 1px solid rgba(255, 255, 255, .07);
  color: rgba(255, 255, 255, .40);
  font-size: .65rem;
}

.smh-client-example pre {
  margin: 0;
  padding: 30px 30px 34px;
  overflow-x: auto;
}

.smh-client-example code {
  color: rgba(255, 255, 255, .84);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: .81rem;
  line-height: 1.8;
}

.smh-quick-grid {
  display: grid;
  grid-template-columns: 1.28fr .86fr .86fr;
  gap: 12px;
  margin-top: 18px;
}

.smh-quick-card {
  position: relative;
  min-height: 178px;
  padding: 26px 26px 24px;
  transition: transform .2s ease, background-color .2s ease;
}

.smh-quick-card:hover {
  transform: translateY(-2px);
  background: rgba(5, 18, 29, .50);
}

.smh-quick-card h3 {
  margin: 32px 0 0;
  font-size: 1.08rem;
  font-weight: 620;
}

.smh-quick-card p {
  max-width: 88%;
  margin: 8px 0 0;
  color: rgba(255, 255, 255, .47);
  font-size: .75rem;
  line-height: 1.58;
}

.smh-quick-card b {
  position: absolute;
  right: 24px;
  bottom: 22px;
  color: rgba(255, 255, 255, .58);
  font-size: .71rem;
  font-weight: 580;
}

.smh-footer {
  min-height: 76px;
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  color: rgba(255, 255, 255, .34);
  font-size: .64rem;
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

  .smh-access-bar {
    grid-template-columns: minmax(280px, 1.7fr) repeat(3, minmax(115px, .7fr));
  }

  .smh-copy-button {
    grid-column: 1 / -1;
    justify-self: end;
    margin: 0 18px 16px;
  }

  .smh-process-card {
    grid-template-columns: minmax(250px, .72fr) minmax(0, 1.78fr);
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

  .smh-access-bar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .smh-access-main {
    grid-column: 1 / -1;
    border-bottom: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-access-item {
    border-left: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-access-item:nth-of-type(even) {
    border-left: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-copy-button {
    grid-column: 1 / -1;
    justify-self: start;
    margin: 16px 20px 18px;
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
    grid-template-columns: 1.25fr 1fr;
  }

  .smh-quick-card:last-child {
    grid-column: 1 / -1;
  }
}

@media (max-width: 720px) {
  .smh-shell {
    width: calc(100vw - 28px);
    padding-top: 16px;
  }

  .smh-status {
    display: none;
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
    font-size: .86rem;
  }

  .smh-hero-actions {
    flex-wrap: wrap;
    margin-top: 22px;
  }

  .smh-access-bar {
    grid-template-columns: 1fr;
  }

  .smh-access-main,
  .smh-access-item {
    min-height: 88px;
    padding: 17px 18px;
  }

  .smh-access-item,
  .smh-access-item:nth-of-type(even) {
    border-left: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-access-main code {
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .smh-process-copy,
  .smh-routing-copy {
    padding: 28px 24px 26px;
  }

  .smh-steps {
    grid-template-columns: 1fr;
  }

  .smh-step,
  .smh-step + .smh-step {
    min-height: 132px;
    padding: 24px;
    border-left: 0;
  }

  .smh-step + .smh-step {
    border-top: 1px solid rgba(255, 255, 255, .07);
  }

  .smh-step h3 {
    margin-top: 28px;
  }

  .smh-client-example {
    margin: 14px;
  }

  .smh-client-example pre {
    padding: 24px 20px 28px;
  }

  .smh-client-example code {
    font-size: .74rem;
  }

  .smh-quick-grid {
    grid-template-columns: 1fr;
  }

  .smh-quick-card:last-child {
    grid-column: auto;
  }

  .smh-quick-card {
    min-height: 150px;
    padding: 22px;
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

@media (prefers-reduced-motion: reduce) {
  .smh-nav a,
  .smh-login-link,
  .smh-button,
  .smh-copy-button,
  .smh-quick-card {
    transition: none !important;
  }
}
</style>
