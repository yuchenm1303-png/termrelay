<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSession } from '../core/session'
import '../styles/home-layout.css'

const { state, isAuthenticated, isAdmin, logout } = useSession()
const copied = ref(false)
const menuOpen = ref(false)
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const apiBase = 'https://api.smirel.com/v1'
const consolePath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const initials = computed(() => (state.user?.username || state.user?.email || 'S').slice(0, 1).toUpperCase())

async function copyBase() {
  await navigator.clipboard.writeText(apiBase)
  copied.value = true
  window.setTimeout(() => { copied.value = false }, 1400)
}

async function signOut() {
  menuOpen.value = false
  await logout()
}
</script>

<template>
  <div class="home-page">
    <div class="site-environment" aria-hidden="true"></div>

    <header class="home-topbar glass">
      <RouterLink to="/home" class="brand-link">
        <img :src="logoUrl" alt="Smirel" />
        <span><strong>Smirel</strong><small>API SERVICE</small></span>
      </RouterLink>

      <nav class="home-nav">
        <RouterLink to="/model-plaza">模型与价格</RouterLink>
        <RouterLink to="/key-usage">用量查询</RouterLink>
        <a href="https://api.smirel.com" target="_blank" rel="noreferrer">接入文档</a>
      </nav>

      <div class="home-actions">
        <span class="service-state"><i></i>服务正常</span>
        <template v-if="isAuthenticated">
          <RouterLink :to="consolePath" class="top-console">{{ isAdmin ? '管理控制台' : '控制台' }}</RouterLink>
          <div class="account-menu-wrap">
            <button class="account-trigger" type="button" @click="menuOpen = !menuOpen">{{ initials }}</button>
            <div v-if="menuOpen" class="account-menu glass">
              <strong>{{ state.user?.username || 'Smirel Account' }}</strong>
              <small>{{ state.user?.email }}</small>
              <RouterLink to="/profile" @click="menuOpen = false">账户设置</RouterLink>
              <RouterLink to="/keys" @click="menuOpen = false">API Keys</RouterLink>
              <RouterLink to="/usage" @click="menuOpen = false">用量与日志</RouterLink>
              <button type="button" @click="signOut">退出登录</button>
            </div>
          </div>
        </template>
        <template v-else>
          <RouterLink to="/login" class="quiet-link">登录</RouterLink>
          <RouterLink to="/register" class="top-console">注册账户</RouterLink>
        </template>
      </div>
    </header>

    <main class="home-content">
      <section class="hero-shell">
        <div class="hero-section">
          <div class="hero-kicker"><span></span>UNIFIED AI GATEWAY</div>
          <h1>统一的大模型接口，<br><em>让接入保持简单。</em></h1>
          <p class="hero-copy">用一个稳定的 API 地址接入 OpenAI、Claude 和 Gemini。统一管理密钥、模型与用量，让应用侧不再重复维护多套接入逻辑。</p>
          <div class="hero-actions">
            <RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">{{ isAuthenticated ? '进入控制台' : '开始使用' }} <span>→</span></RouterLink>
            <RouterLink to="/model-plaza" class="secondary-button">查看模型价格</RouterLink>
          </div>
          <div class="hero-assurance">
            <span><i></i> OpenAI Compatible</span>
            <span>按实际用量计费</span>
            <span>多模型统一入口</span>
          </div>
        </div>

        <section class="gateway-card glass" aria-label="Smirel API endpoint">
          <header class="gateway-card-head">
            <div>
              <span class="gateway-mark">S</span>
              <p><strong>Smirel Gateway</strong><small>Production endpoint</small></p>
            </div>
            <span class="gateway-live"><i></i>Operational</span>
          </header>

          <div class="gateway-endpoint">
            <span>BASE URL</span>
            <div class="endpoint-input">
              <code>{{ apiBase }}</code>
              <button type="button" @click="copyBase">{{ copied ? '已复制' : '复制' }}</button>
            </div>
          </div>

          <div class="gateway-capabilities">
            <div><span>01</span><p><strong>统一模型入口</strong><small>OpenAI · Anthropic · Google</small></p><b>→</b></div>
            <div><span>02</span><p><strong>独立 API Key</strong><small>按项目管理访问凭证</small></p><b>→</b></div>
            <div><span>03</span><p><strong>请求与用量</strong><small>调用、Token 与费用记录</small></p><b>→</b></div>
          </div>

          <footer class="gateway-meta">
            <span>AUTH<strong>Bearer API Key</strong></span>
            <span>PROTOCOL<strong>OpenAI Compatible</strong></span>
            <span>BILLING<strong>Pay as you go</strong></span>
          </footer>
        </section>
      </section>

      <section class="provider-strip" aria-label="Supported model ecosystems">
        <span class="provider-strip-label">MODEL ECOSYSTEM</span>
        <div class="provider-list">
          <span><b>O</b>OpenAI</span>
          <span><b>A</b>Anthropic</span>
          <span><b>G</b>Google Gemini</span>
          <span><b>X</b>Grok</span>
        </div>
        <RouterLink to="/model-plaza">查看全部模型 →</RouterLink>
      </section>

      <section class="feature-section">
        <div class="section-heading">
          <p class="eyebrow">CORE FEATURES</p>
          <h2>把模型接入需要的东西，收进一个入口。</h2>
          <p>不增加复杂概念，只保留真正影响开发和使用体验的基础能力。</p>
        </div>

        <div class="feature-grid">
          <article class="feature-card glass">
            <div class="feature-icon">01</div>
            <h3>一键接入</h3>
            <p>创建 Key、替换 Base URL，即可沿用熟悉的 OpenAI SDK 调用方式。</p>
            <RouterLink :to="isAuthenticated ? '/keys' : '/register'">开始接入 →</RouterLink>
          </article>
          <article class="feature-card glass">
            <div class="feature-icon">02</div>
            <h3>统一模型路由</h3>
            <p>模型可以切换，应用里的接入地址不用跟着变化。</p>
            <RouterLink to="/model-plaza">浏览模型 →</RouterLink>
          </article>
          <article class="feature-card glass">
            <div class="feature-icon">03</div>
            <h3>用量清晰</h3>
            <p>请求、Token 与费用集中记录，知道每一次调用花在了哪里。</p>
            <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'">查看用量 →</RouterLink>
          </article>
          <article class="feature-card glass">
            <div class="feature-icon">04</div>
            <h3>密钥独立管理</h3>
            <p>按项目创建和停用 API Key，让不同应用的访问边界更清楚。</p>
            <RouterLink :to="isAuthenticated ? '/keys' : '/login'">管理密钥 →</RouterLink>
          </article>
        </div>
      </section>

      <section class="integration-section glass">
        <div class="integration-copy">
          <p class="eyebrow">QUICK START</p>
          <h2>改一个 Base URL，<br>就可以开始调用。</h2>
          <p>保持原来的 SDK 和调用习惯，只把客户端指向 Smirel。</p>
          <div class="integration-steps">
            <span><b>01</b> 创建 API Key</span>
            <span><b>02</b> 设置 Base URL</span>
            <span><b>03</b> 选择模型并发送请求</span>
          </div>
        </div>

        <div class="code-panel">
          <header><span><i></i><i></i><i></i></span><b>quickstart.py</b><small>Python</small></header>
          <pre><code><span class="code-muted">from</span> openai <span class="code-muted">import</span> OpenAI

client = OpenAI(
    base_url=<span class="code-accent">"https://api.smirel.com/v1"</span>,
    api_key=<span class="code-accent">"sk-..."</span>
)

response = client.responses.create(
    model=<span class="code-accent">"gpt-5"</span>,
    input=<span class="code-accent">"Hello"</span>
)</code></pre>
        </div>
      </section>

      <section class="quick-grid">
        <RouterLink to="/model-plaza" class="glass quick-card"><span>MODELS</span><strong>模型与价格</strong><small>查看当前模型、价格与服务信息。</small><b>查看 →</b></RouterLink>
        <RouterLink :to="isAuthenticated ? '/keys' : '/login'" class="glass quick-card"><span>API KEY</span><strong>密钥管理</strong><small>创建、停用和管理项目密钥。</small><b>管理 →</b></RouterLink>
        <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'" class="glass quick-card"><span>USAGE</span><strong>用量查询</strong><small>查询请求、Token 与费用记录。</small><b>查询 →</b></RouterLink>
      </section>
    </main>

    <footer class="home-footer">
      <span>© {{ new Date().getFullYear() }} Smirel</span>
      <span>Unified AI API Gateway · api.smirel.com/v1</span>
    </footer>
  </div>
</template>
