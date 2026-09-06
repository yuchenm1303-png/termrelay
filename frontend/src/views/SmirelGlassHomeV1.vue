<template>
  <div class="spg-page smh-home">
    <div class="spg-environment smh-environment" aria-hidden="true"></div>

    <div class="smh-hero-shell">
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
            <h1>OpenAI、Claude、Gemini，用一个 API 地址接入。</h1>
            <p class="smh-hero-description">
              把客户端的 Base URL 改成 Smirel，再填入自己的 API Key。现有 OpenAI SDK 和常用客户端可以直接继续用。
            </p>
            <div class="smh-hero-actions">
              <router-link :to="startPath" class="smh-button smh-button--primary">
                {{ isAuthenticated ? '打开控制台' : '注册账户' }}
              </router-link>
              <router-link to="/model-plaza" class="smh-button smh-button--quiet">
                查看模型价格
              </router-link>
            </div>
          </div>
        </section>

        <section id="access" class="smh-access-bar" aria-label="API 接入信息">
          <div class="smh-access-main">
            <span>BASE URL</span>
            <code>{{ apiBase }}</code>
          </div>
          <div class="smh-access-item">
            <span>认证</span>
            <strong>Bearer API Key</strong>
          </div>
          <div class="smh-access-item">
            <span>接口</span>
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
      </main>
    </div>

    <section class="smh-body">
      <div class="smh-body-inner">
        <section class="smh-steps-section">
          <div class="smh-section-heading">
            <p>开始调用</p>
            <h2>第一次调用，只需要三步。</h2>
            <span>不用换客户端，也不用改原来的请求格式。</span>
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
              <p>为项目生成自己的密钥。</p>
            </div>
            <div class="smh-step">
              <span>03</span>
              <h3>替换 Base URL</h3>
              <p>把客户端地址改成上面的 API 地址。</p>
            </div>
          </div>
        </section>

        <section class="smh-routing-section">
          <div class="smh-routing-copy">
            <p class="smh-section-label">模型接入</p>
            <h2>换模型，不用换接入方式。</h2>
            <p>
              OpenAI、Anthropic 和 Google 的上游配置由服务端处理。客户端仍然使用同一个 Base URL 和 API Key。
            </p>
            <div class="smh-provider-line" aria-label="支持的模型服务">
              <span>OpenAI</span>
              <span>Anthropic</span>
              <span>Google</span>
            </div>
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

        <section class="smh-links-section" aria-label="常用入口">
          <router-link to="/model-plaza" class="smh-link-row">
            <span class="smh-link-index">01</span>
            <div>
              <h3>模型与价格</h3>
              <p>查看可用模型和当前单价。</p>
            </div>
            <b>查看 →</b>
          </router-link>

          <router-link :to="keyPath" class="smh-link-row">
            <span class="smh-link-index">02</span>
            <div>
              <h3>API Key</h3>
              <p>创建、停用和管理自己的密钥。</p>
            </div>
            <b>{{ isAuthenticated ? '管理 →' : '登录 →' }}</b>
          </router-link>

          <router-link to="/key-usage" class="smh-link-row">
            <span class="smh-link-index">03</span>
            <div>
              <h3>用量记录</h3>
              <p>按 Key 查看请求、Token 和费用。</p>
            </div>
            <b>查询 →</b>
          </router-link>
        </section>

        <footer class="smh-footer">
          <span>© {{ currentYear }} {{ siteName }}</span>
          <div>
            <span>{{ apiBase }}</span>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">接入文档</a>
          </div>
        </footer>
      </div>
    </section>
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
  background: #111a21;
}

.smh-home a {
  color: inherit;
  text-decoration: none;
}

.smh-environment::after {
  background:
    linear-gradient(180deg, rgba(4, 10, 16, .15), rgba(4, 10, 16, .34)),
    linear-gradient(90deg, rgba(5, 12, 18, .24) 0%, rgba(5, 12, 18, .08) 52%, rgba(5, 12, 18, .18) 100%);
}

.smh-hero-shell {
  position: relative;
  z-index: 1;
  width: min(1500px, calc(100vw - 96px));
  margin: 0 auto;
  padding: 24px 0 48px;
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
  gap: 20px;
}

.smh-nav {
  gap: 24px;
}

.smh-nav a,
.smh-status,
.smh-login-link {
  color: rgba(255, 255, 255, .68);
  font-size: .72rem;
}

.smh-nav a,
.smh-login-link {
  transition: color .18s ease;
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
  border: 1px solid rgba(255, 255, 255, .12);
  border-radius: 999px;
  background: rgba(7, 15, 22, .16);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.smh-hero {
  min-height: 430px;
  display: flex;
  align-items: center;
  padding: 40px 0 54px;
}

.smh-hero-copy {
  width: min(760px, 60vw);
}

.smh-kicker,
.smh-section-label,
.smh-section-heading > p {
  margin: 0;
  color: rgba(255, 255, 255, .46);
  font-size: .66rem;
  letter-spacing: .15em;
}

.smh-hero h1 {
  max-width: 760px;
  margin: 12px 0 0;
  font-size: clamp(3.1rem, 5.2vw, 4.7rem);
  line-height: 1.02;
  font-weight: 560;
  letter-spacing: -.052em;
}

.smh-hero-description {
  max-width: 670px;
  margin: 20px 0 0;
  color: rgba(255, 255, 255, .72);
  font-size: .96rem;
  line-height: 1.78;
}

.smh-hero-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 28px;
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
  background: rgba(4, 10, 16, .16);
  color: rgba(255, 255, 255, .72);
}

.smh-button--quiet:hover {
  color: rgba(255, 255, 255, .96);
  border-color: rgba(255, 255, 255, .14);
}

.smh-access-bar {
  display: grid;
  grid-template-columns: minmax(330px, 2fr) repeat(3, minmax(130px, .72fr)) auto;
  align-items: stretch;
  border: 1px solid rgba(255, 255, 255, .10);
  border-radius: 8px;
  background: rgba(7, 18, 27, .56);
  -webkit-backdrop-filter: blur(24px) saturate(118%);
  backdrop-filter: blur(24px) saturate(118%);
  overflow: hidden;
}

.smh-access-main,
.smh-access-item {
  min-width: 0;
  min-height: 112px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 20px 22px;
}

.smh-access-item {
  border-left: 1px solid rgba(255, 255, 255, .08);
}

.smh-access-main span,
.smh-access-item span {
  color: rgba(255, 255, 255, .42);
  font-size: .61rem;
  letter-spacing: .10em;
}

.smh-access-main code {
  margin-top: 9px;
  overflow: hidden;
  color: rgba(255, 255, 255, .97);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 1.02rem;
  font-weight: 580;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smh-access-item strong {
  margin-top: 8px;
  color: rgba(255, 255, 255, .82);
  font-size: .73rem;
  font-weight: 610;
}

.smh-copy-button {
  align-self: center;
  min-width: 92px;
  min-height: 40px;
  margin: 0 18px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, .13);
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

.smh-body {
  position: relative;
  z-index: 2;
  background: rgba(14, 23, 30, .985);
  border-top: 1px solid rgba(255, 255, 255, .06);
  color: #fff;
}

.smh-body-inner {
  width: min(1360px, calc(100vw - 96px));
  margin: 0 auto;
  padding: 96px 0 34px;
}

.smh-section-heading {
  max-width: 680px;
}

.smh-section-heading h2,
.smh-routing-copy h2 {
  margin: 12px 0 0;
  font-size: clamp(2rem, 3.3vw, 3rem);
  line-height: 1.12;
  font-weight: 560;
  letter-spacing: -.038em;
}

.smh-section-heading > span {
  display: block;
  margin-top: 14px;
  color: rgba(255, 255, 255, .48);
  font-size: .86rem;
  line-height: 1.7;
}

.smh-steps {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-top: 54px;
  border-top: 1px solid rgba(255, 255, 255, .10);
  border-bottom: 1px solid rgba(255, 255, 255, .10);
}

.smh-step {
  min-height: 180px;
  padding: 32px 34px 34px 0;
}

.smh-step + .smh-step {
  padding-left: 34px;
  border-left: 1px solid rgba(255, 255, 255, .10);
}

.smh-step > span,
.smh-link-index {
  color: rgba(255, 255, 255, .34);
  font-size: .64rem;
  letter-spacing: .10em;
}

.smh-step h3 {
  margin: 34px 0 0;
  font-size: 1.12rem;
  font-weight: 620;
}

.smh-step p {
  margin: 10px 0 0;
  color: rgba(255, 255, 255, .48);
  font-size: .78rem;
  line-height: 1.6;
}

.smh-routing-section {
  display: grid;
  grid-template-columns: minmax(0, .82fr) minmax(460px, 1.18fr);
  gap: 86px;
  align-items: center;
  margin-top: 112px;
  padding-top: 88px;
  border-top: 1px solid rgba(255, 255, 255, .10);
}

.smh-routing-copy > p:last-of-type {
  max-width: 520px;
  margin: 22px 0 0;
  color: rgba(255, 255, 255, .52);
  font-size: .85rem;
  line-height: 1.8;
}

.smh-provider-line {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  margin-top: 34px;
  color: rgba(255, 255, 255, .72);
  font-size: .78rem;
}

.smh-provider-line span + span::before {
  content: '/';
  margin-right: 24px;
  color: rgba(255, 255, 255, .20);
}

.smh-client-example {
  border-top: 1px solid rgba(255, 255, 255, .14);
  border-bottom: 1px solid rgba(255, 255, 255, .10);
  background: rgba(255, 255, 255, .025);
}

.smh-client-example-head {
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
  border-bottom: 1px solid rgba(255, 255, 255, .08);
  color: rgba(255, 255, 255, .40);
  font-size: .66rem;
}

.smh-client-example pre {
  margin: 0;
  padding: 30px 32px 34px;
  overflow-x: auto;
}

.smh-client-example code {
  color: rgba(255, 255, 255, .84);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: .82rem;
  line-height: 1.8;
}

.smh-links-section {
  margin-top: 112px;
  border-top: 1px solid rgba(255, 255, 255, .11);
}

.smh-link-row {
  min-height: 126px;
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr) auto;
  gap: 24px;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, .10);
  transition: background-color .18s ease;
}

.smh-link-row:hover {
  background: rgba(255, 255, 255, .025);
}

.smh-link-row h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 610;
}

.smh-link-row p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, .43);
  font-size: .76rem;
}

.smh-link-row b {
  color: rgba(255, 255, 255, .58);
  font-size: .72rem;
  font-weight: 580;
}

.smh-footer {
  min-height: 88px;
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  color: rgba(255, 255, 255, .32);
  font-size: .65rem;
}

.smh-footer > div {
  display: flex;
  gap: 22px;
}

.smh-footer a:hover {
  color: rgba(255, 255, 255, .70);
}

@media (max-width: 1120px) {
  .smh-hero-shell,
  .smh-body-inner {
    width: calc(100vw - 48px);
  }

  .smh-nav {
    gap: 16px;
  }

  .smh-access-bar {
    grid-template-columns: minmax(280px, 1.6fr) repeat(3, minmax(120px, .7fr));
  }

  .smh-copy-button {
    grid-column: 1 / -1;
    justify-self: end;
    margin: 0 18px 16px;
  }

  .smh-routing-section {
    gap: 52px;
  }
}

@media (max-width: 980px) {
  .smh-header-right {
    gap: 12px;
  }

  .smh-nav {
    display: none;
  }

  .smh-hero {
    min-height: 390px;
  }

  .smh-hero-copy {
    width: min(700px, 90vw);
  }

  .smh-access-bar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .smh-access-main {
    grid-column: 1 / -1;
    border-bottom: 1px solid rgba(255, 255, 255, .08);
  }

  .smh-access-item {
    border-left: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .08);
  }

  .smh-access-item:nth-of-type(even) {
    border-left: 1px solid rgba(255, 255, 255, .08);
  }

  .smh-copy-button {
    grid-column: 1 / -1;
    justify-self: start;
    margin: 16px 20px 18px;
  }

  .smh-routing-section {
    grid-template-columns: 1fr;
    gap: 42px;
  }
}

@media (max-width: 720px) {
  .smh-hero-shell,
  .smh-body-inner {
    width: calc(100vw - 28px);
  }

  .smh-hero-shell {
    padding-top: 16px;
  }

  .smh-status {
    display: none;
  }

  .smh-hero {
    min-height: 350px;
    padding: 36px 0 42px;
  }

  .smh-hero-copy {
    width: 100%;
  }

  .smh-hero h1 {
    font-size: clamp(2.5rem, 13vw, 3.45rem);
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
    min-height: 90px;
    padding: 17px 18px;
  }

  .smh-access-item,
  .smh-access-item:nth-of-type(even) {
    border-left: 0;
    border-bottom: 1px solid rgba(255, 255, 255, .08);
  }

  .smh-access-main code {
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .smh-body-inner {
    padding-top: 72px;
  }

  .smh-steps {
    grid-template-columns: 1fr;
    margin-top: 40px;
  }

  .smh-step,
  .smh-step + .smh-step {
    min-height: 138px;
    padding: 26px 0;
    border-left: 0;
  }

  .smh-step + .smh-step {
    border-top: 1px solid rgba(255, 255, 255, .10);
  }

  .smh-step h3 {
    margin-top: 24px;
  }

  .smh-routing-section,
  .smh-links-section {
    margin-top: 82px;
  }

  .smh-routing-section {
    padding-top: 64px;
  }

  .smh-client-example pre {
    padding: 24px 20px 28px;
  }

  .smh-client-example code {
    font-size: .74rem;
  }

  .smh-link-row {
    min-height: 118px;
    grid-template-columns: 42px minmax(0, 1fr);
    gap: 14px;
    padding: 18px 0;
  }

  .smh-link-row b {
    grid-column: 2;
    margin-top: -4px;
  }

  .smh-footer {
    align-items: flex-start;
    flex-direction: column;
    padding: 34px 0 12px;
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
  .smh-link-row {
    transition: none !important;
  }
}
</style>