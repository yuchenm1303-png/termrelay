<template>
  <div class="smb-home">
    <div class="smb-background" aria-hidden="true">
      <div class="smb-grid"></div>
      <div class="smb-glow smb-glow-a"></div>
      <div class="smb-glow smb-glow-b"></div>
      <div class="smb-glow smb-glow-c"></div>
    </div>

    <div class="smb-shell">
      <header class="smb-topbar">
        <router-link to="/home" class="smb-brand" aria-label="Smirel">
          <span class="smb-brand-mark">S</span>
          <span class="smb-brand-copy">
            <strong>Smirel</strong>
            <small>API Access Platform</small>
          </span>
        </router-link>

        <nav class="smb-nav">
          <a href="#models">模型与价格</a>
          <a href="#access">接入方式</a>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">开发文档</a>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="smb-nav-login">
            {{ isAuthenticated ? '进入控制台' : '登录' }}
          </router-link>
        </nav>
      </header>

      <main class="smb-content">
        <section class="smb-hero">
          <div class="smb-hero-copy">
            <div class="smb-eyebrow">SMIREL API GATEWAY</div>
            <h1>一个接口，接入主流模型服务。</h1>
            <p class="smb-lead">
              为团队提供稳定、透明、可管理的 API 接入。统一密钥、统一用量记录、统一账单，现有 OpenAI SDK 可直接迁移。
            </p>

            <div class="smb-actions">
              <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="smb-button smb-button-primary">
                {{ isAuthenticated ? '进入控制台' : '开始使用' }}
              </router-link>
              <a href="#models" class="smb-button smb-button-secondary">查看模型与价格</a>
            </div>

            <div class="smb-proof">
              <span><i></i>统一 API Key</span>
              <span><i></i>按量计费</span>
              <span><i></i>调用记录可查</span>
            </div>
          </div>

          <article class="smb-panel smb-access-panel">
            <div class="smb-panel-head">
              <div>
                <span class="smb-overline">API ACCESS</span>
                <h2>接入信息</h2>
              </div>
              <span class="smb-status-badge"><i></i>服务正常</span>
            </div>

            <div class="smb-access-list">
              <div class="smb-access-row">
                <span>API Base URL</span>
                <code>{{ apiBase }}</code>
              </div>
              <div class="smb-access-row">
                <span>兼容方式</span>
                <strong>OpenAI SDK / HTTP</strong>
              </div>
              <div class="smb-access-row">
                <span>认证方式</span>
                <strong>Bearer API Key</strong>
              </div>
            </div>

            <div class="smb-access-note">
              <span class="smb-overline">MIGRATION</span>
              <p>通常只需要替换 Base URL 和 API Key，即可把现有项目迁移到 Smirel。</p>
            </div>
          </article>
        </section>

        <section class="smb-section" id="models">
          <div class="smb-section-head">
            <div>
              <span class="smb-overline">AVAILABLE SERVICES</span>
              <h2>按业务需要选择模型，不必更换接入方式。</h2>
            </div>
            <p>同一账户下管理不同模型的调用与费用。模型目录、可用状态和价格在控制台统一维护。</p>
          </div>

          <div class="smb-card-grid">
            <article class="smb-panel smb-service-card">
              <span class="smb-card-index">01</span>
              <div>
                <span class="smb-overline">MODEL CATALOG</span>
                <h3>主流模型统一接入</h3>
                <p>通过同一套认证方式调用不同模型，减少多平台账号、密钥和 SDK 的维护成本。</p>
              </div>
            </article>

            <article class="smb-panel smb-service-card">
              <span class="smb-card-index">02</span>
              <div>
                <span class="smb-overline">ACCOUNT CONTROL</span>
                <h3>密钥与额度独立管理</h3>
                <p>为不同项目创建独立 API Key，分别管理额度和使用范围，团队协作更清楚。</p>
              </div>
            </article>

            <article class="smb-panel smb-service-card">
              <span class="smb-card-index">03</span>
              <div>
                <span class="smb-overline">USAGE RECORDS</span>
                <h3>每次调用都有记录</h3>
                <p>请求量、Token、费用和状态集中查看，方便核对成本，也方便定位异常调用。</p>
              </div>
            </article>
          </div>
        </section>

        <section class="smb-section" id="access">
          <div class="smb-section-title-row">
            <div>
              <span class="smb-overline">GET STARTED</span>
              <h2>接入只保留必要步骤。</h2>
            </div>
            <span class="smb-inline-state"><i></i>OpenAI-compatible</span>
          </div>

          <div class="smb-access-grid">
            <article class="smb-panel smb-start-card">
              <div class="smb-start-main">
                <div class="smb-start-number">01</div>
                <div>
                  <span class="smb-overline">CREATE KEY</span>
                  <h3>创建 API Key</h3>
                  <p>登录控制台后创建项目密钥，按需要设置额度与访问范围。</p>
                </div>
              </div>
              <div class="smb-divider"></div>
              <div class="smb-start-main">
                <div class="smb-start-number">02</div>
                <div>
                  <span class="smb-overline">CHANGE ENDPOINT</span>
                  <h3>替换接入地址</h3>
                  <p>保留现有调用代码，将 Base URL 替换为 Smirel 地址。</p>
                </div>
              </div>
              <div class="smb-divider"></div>
              <div class="smb-start-main">
                <div class="smb-start-number">03</div>
                <div>
                  <span class="smb-overline">START REQUESTS</span>
                  <h3>开始调用</h3>
                  <p>请求成功后，用量与费用会自动进入账户记录，无需额外配置统计系统。</p>
                </div>
              </div>
            </article>

            <article class="smb-panel smb-cta-card">
              <span class="smb-overline">ACCOUNT ACCESS</span>
              <h3>准备开始接入？</h3>
              <p>创建账户后即可查看可用模型、价格和接入信息。</p>
              <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="smb-button smb-button-primary smb-button-block">
                {{ isAuthenticated ? '打开控制台' : '创建账户' }}
              </router-link>
              <small>无需改变现有业务架构，按项目逐步迁移即可。</small>
            </article>
          </div>
        </section>
      </main>

      <footer class="smb-footer">
        <span>© {{ currentYear }} Smirel</span>
        <span>API access, usage and billing in one place.</span>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const authStore = useAuthStore()

const apiBase = computed(() =>
  (appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || 'https://api.smirel.com/v1')
    .trim()
    .replace(/\/$/, ''),
)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const currentYear = new Date().getFullYear()

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>

<style src="../styles/smirel-business-system.css"></style>
