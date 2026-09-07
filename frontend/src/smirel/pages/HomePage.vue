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
        <h1>一个 API，接入主流模型。</h1>
        <p class="hero-copy">OpenAI、Claude、Gemini 等模型共用一套地址和密钥，调用方式保持一致。</p>

        <div class="hero-actions">
          <RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">
            {{ isAuthenticated ? '进入控制台' : '开始使用' }}
          </RouterLink>
          <RouterLink to="/model-plaza" class="secondary-button">查看模型</RouterLink>
        </div>

        <div class="endpoint-bar" aria-label="Smirel API endpoint">
          <span>Base URL</span>
          <code>{{ apiBase }}</code>
          <button type="button" @click="copyBase">{{ copied ? '已复制' : '复制' }}</button>
        </div>

        <div class="hero-meta">
          <span><i></i>服务正常</span>
          <span>OpenAI Compatible</span>
          <span>按实际用量计费</span>
        </div>
      </section>

      <section class="provider-line" aria-label="Supported model providers">
        <span>OpenAI</span>
        <b>·</b>
        <span>Claude</span>
        <b>·</b>
        <span>Gemini</span>
        <b>·</b>
        <span>Grok</span>
        <b>·</b>
        <RouterLink to="/model-plaza">更多模型</RouterLink>
      </section>

      <section class="benefit-grid">
        <RouterLink :to="isAuthenticated ? '/keys' : '/register'" class="benefit-card">
          <span>接入</span>
          <h2>改地址就能用</h2>
          <p>沿用 OpenAI SDK，把 Base URL 换成 Smirel。</p>
          <b>开始接入 →</b>
        </RouterLink>

        <RouterLink to="/model-plaza" class="benefit-card">
          <span>模型</span>
          <h2>模型随时换</h2>
          <p>接入地址不变，按需要选择不同模型。</p>
          <b>查看模型 →</b>
        </RouterLink>

        <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'" class="benefit-card">
          <span>用量</span>
          <h2>账单看得清</h2>
          <p>请求、Token 和费用集中查看。</p>
          <b>查看用量 →</b>
        </RouterLink>
      </section>
    </main>

    <footer class="home-footer">
      <span>© {{ new Date().getFullYear() }} Smirel</span>
      <span>api.smirel.com/v1</span>
    </footer>
  </div>
</template>
