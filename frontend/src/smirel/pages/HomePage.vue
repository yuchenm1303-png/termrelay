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
    <header class="home-topbar">
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
            <div v-if="menuOpen" class="account-menu">
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
          <RouterLink to="/register" class="top-console">注册</RouterLink>
        </template>
      </div>
    </header>

    <main class="home-content">
      <section class="hero-section">
        <div class="hero-message">
          <h1>模型很多，<br><strong>接口一个。</strong></h1>
          <p>OpenAI、Claude、Gemini、Grok 接入 Smirel 后，你只维护一个 Base URL 和一套 Key。</p>
          <div class="hero-actions">
            <RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">
              {{ isAuthenticated ? '进入控制台' : '创建 API Key' }}
            </RouterLink>
            <RouterLink to="/model-plaza" class="text-button">查看模型与价格 →</RouterLink>
          </div>
        </div>

        <div class="route-signature" aria-label="Smirel model routing">
          <div class="route-providers">
            <span>OpenAI</span>
            <span>Claude</span>
            <span>Gemini</span>
            <span>Grok</span>
          </div>
          <div class="route-track" aria-hidden="true"><i></i></div>
          <div class="route-endpoint">
            <small>BASE URL</small>
            <code>{{ apiBase }}</code>
            <button type="button" @click="copyBase">{{ copied ? '已复制' : '复制' }}</button>
          </div>
        </div>

        <div class="route-meta">
          <span><i></i>服务正常</span>
          <span>OpenAI Compatible</span>
          <span>按实际用量计费</span>
        </div>
      </section>

      <section class="change-section">
        <div class="change-copy">
          <span>接入</span>
          <h2>你真正要改的，<br>只有一行。</h2>
          <p>原来的 SDK、请求结构和业务代码都不用重写。</p>
        </div>

        <div class="change-code" aria-label="Base URL change example">
          <div class="code-line code-line--old"><b>−</b><code>base_url="https://api.openai.com/v1"</code></div>
          <div class="code-line code-line--new"><b>+</b><code>base_url="https://api.smirel.com/v1"</code></div>
        </div>
      </section>

      <section class="capability-list">
        <article>
          <span>01</span>
          <h2>换模型，不换接口。</h2>
          <p>应用继续请求同一个地址，模型按需要切换。</p>
          <RouterLink to="/model-plaza">查看模型 →</RouterLink>
        </article>

        <article>
          <span>02</span>
          <h2>Key 按项目分开。</h2>
          <p>每个项目独立创建和停用，权限边界更清楚。</p>
          <RouterLink :to="isAuthenticated ? '/keys' : '/register'">管理 Key →</RouterLink>
        </article>

        <article>
          <span>03</span>
          <h2>每次调用，都有记录。</h2>
          <p>请求、Token 和费用放在一起看，不需要自己对账。</p>
          <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'">查看用量 →</RouterLink>
        </article>
      </section>

      <section class="closing-section">
        <div>
          <span>SMIREL</span>
          <h2>创建一个 Key，开始调用。</h2>
        </div>
        <RouterLink :to="isAuthenticated ? consolePath : '/register'">
          {{ isAuthenticated ? '进入控制台' : '开始使用' }} →
        </RouterLink>
      </section>
    </main>

    <footer class="home-footer">
      <span>© {{ new Date().getFullYear() }} Smirel</span>
      <span>api.smirel.com/v1</span>
    </footer>
  </div>
</template>
