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
        <div class="hero-copy-block">
          <span class="hero-kicker">SMIREL / MODEL API</span>
          <h1>统一的大模型接口，<br>让您的 vibecoding 更高效。</h1>
          <p>OpenAI、Claude、Gemini、Grok 共用一个 Base URL。保留熟悉的 SDK 和调用方式，把模型切换、Key 和用量留给 Smirel。</p>

          <div class="hero-actions">
            <RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">
              {{ isAuthenticated ? '进入控制台' : '开始使用' }}
            </RouterLink>
            <RouterLink to="/model-plaza" class="secondary-button">查看模型</RouterLink>
          </div>

          <div class="hero-trust">
            <span><i></i>服务正常</span>
            <span>OpenAI Compatible</span>
            <span>按实际用量计费</span>
          </div>
        </div>

        <div class="route-board" aria-label="Smirel unified model route">
          <div class="route-board-head">
            <span>ROUTING / LIVE</span>
            <small><i></i>在线</small>
          </div>

          <div class="route-base">
            <div>
              <small>BASE URL</small>
              <code>{{ apiBase }}</code>
            </div>
            <button type="button" @click="copyBase">{{ copied ? '已复制' : '复制' }}</button>
          </div>

          <div class="route-lanes">
            <div class="route-lane"><span>OpenAI</span><i></i><b>统一入口</b></div>
            <div class="route-lane"><span>Claude</span><i></i><b>统一入口</b></div>
            <div class="route-lane"><span>Gemini</span><i></i><b>统一入口</b></div>
            <div class="route-lane"><span>Grok</span><i></i><b>统一入口</b></div>
          </div>

          <div class="route-board-foot">
            <span>Bearer API Key</span>
            <span>OpenAI SDK</span>
            <RouterLink to="/model-plaza">更多模型 →</RouterLink>
          </div>
        </div>
      </section>

      <section class="integration-section">
        <div class="section-heading">
          <span>接入</span>
          <h2>接入，不换写法。</h2>
          <p>原来的业务代码继续用。把请求地址换成 Smirel，就可以开始调用。</p>
        </div>

        <div class="integration-panel">
          <div class="integration-code" aria-label="Base URL change example">
            <div class="code-caption">
              <span>base_url.py</span>
              <small>Python</small>
            </div>
            <div class="code-line code-line--old"><b>−</b><code>base_url="https://api.openai.com/v1"</code></div>
            <div class="code-line code-line--new"><b>+</b><code>base_url="https://api.smirel.com/v1"</code></div>
          </div>

          <div class="integration-notes">
            <div>
              <span>01</span>
              <strong>Base URL</strong>
              <p>统一请求入口，不需要维护多套地址。</p>
            </div>
            <div>
              <span>02</span>
              <strong>API Key</strong>
              <p>按项目创建、停用和管理访问凭证。</p>
            </div>
            <div>
              <span>03</span>
              <strong>Model</strong>
              <p>模型按需要切换，接入方式保持一致。</p>
            </div>
          </div>
        </div>
      </section>

      <section class="workspace-section">
        <div class="workspace-copy">
          <span>工作区</span>
          <h2>模型、Key 和用量，<br>放在一个地方。</h2>
          <p>首页只负责让您快速开始。后续的模型选择、密钥和调用记录都回到控制台处理。</p>
        </div>

        <div class="workspace-links">
          <RouterLink to="/model-plaza">
            <span>MODEL</span>
            <strong>查看模型与价格</strong>
            <p>挑选模型，查看当前价格和可用状态。</p>
            <b>→</b>
          </RouterLink>
          <RouterLink :to="isAuthenticated ? '/keys' : '/register'">
            <span>KEYS</span>
            <strong>管理 API Key</strong>
            <p>为不同项目创建独立访问凭证。</p>
            <b>→</b>
          </RouterLink>
          <RouterLink :to="isAuthenticated ? '/usage' : '/key-usage'">
            <span>USAGE</span>
            <strong>查看调用与用量</strong>
            <p>请求、Token 和费用集中查看。</p>
            <b>→</b>
          </RouterLink>
        </div>
      </section>

      <section class="closing-section">
        <div>
          <span>READY</span>
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
