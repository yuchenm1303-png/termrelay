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
      <section class="home-intro-grid">
        <section class="hero-section">
          <p class="eyebrow">SMIREL API</p>
          <h1>一个 API 地址，调用 OpenAI、Claude 和 Gemini。</h1>
          <p class="hero-copy">统一模型接入、密钥管理和用量统计。你的应用只需要维护一个 Base URL 和一套认证方式。</p>
          <div class="hero-actions">
            <RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">{{ isAuthenticated ? '打开控制台' : '注册账户' }}</RouterLink>
            <RouterLink to="/model-plaza" class="secondary-button">查看模型价格</RouterLink>
          </div>
        </section>

        <section class="endpoint-card glass">
          <div class="endpoint-main"><span>BASE URL</span><code>{{ apiBase }}</code></div>
          <button type="button" @click="copyBase">{{ copied ? '已复制' : '复制地址' }}</button>
          <div class="endpoint-meta"><span>认证<strong>Bearer API Key</strong></span><span>协议<strong>OpenAI Compatible</strong></span><span>计费<strong>按实际用量</strong></span></div>
        </section>
      </section>

      <section class="home-feature-grid">
        <article class="glass feature-panel">
          <div class="feature-copy">
            <span class="eyebrow">GET STARTED</span>
            <h2>三步接入。</h2>
            <p>创建账户、生成 Key、替换 Base URL。其余调用方式保持不变。</p>
            <RouterLink class="feature-link" :to="isAuthenticated ? '/keys' : '/register'">开始使用 →</RouterLink>
          </div>
          <div class="compact-steps">
            <div><b>01</b><strong>创建账户</strong><span>进入自己的 Smirel 工作区。</span></div>
            <div><b>02</b><strong>生成 API Key</strong><span>按项目独立管理访问凭证。</span></div>
            <div><b>03</b><strong>替换地址</strong><span>客户端指向 Smirel Base URL。</span></div>
          </div>
        </article>

        <article class="glass feature-panel feature-panel--routing">
          <div class="feature-copy">
            <span class="eyebrow">MODEL ROUTING</span>
            <h2>模型可以换，接入地址不用换。</h2>
            <p>OpenAI、Anthropic、Google 使用统一入口。模型能力和价格在一个地方管理。</p>
            <div class="provider-line"><span>OpenAI</span><span>Anthropic</span><span>Google</span></div>
          </div>
          <div class="compact-code">
            <header><span>OpenAI SDK</span><span>Python</span></header>
            <pre><code>from openai import OpenAI

client = OpenAI(
    base_url="https://api.smirel.com/v1",
    api_key="sk-..."
)</code></pre>
          </div>
        </article>
      </section>

      <section class="quick-grid">
        <RouterLink to="/model-plaza" class="glass quick-card"><span>MODELS</span><strong>模型与价格</strong><small>查看模型、价格与服务状态。</small><b>查看 →</b></RouterLink>
        <RouterLink :to="isAuthenticated ? '/keys' : '/login'" class="glass quick-card"><span>API KEY</span><strong>密钥</strong><small>创建、停用和管理项目密钥。</small><b>管理 →</b></RouterLink>
        <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'" class="glass quick-card"><span>USAGE</span><strong>用量</strong><small>请求、Token 与费用记录。</small><b>查询 →</b></RouterLink>
      </section>
    </main>
    <footer class="home-footer"><span>© {{ new Date().getFullYear() }} Smirel</span><span>api.smirel.com/v1</span></footer>
  </div>
</template>
