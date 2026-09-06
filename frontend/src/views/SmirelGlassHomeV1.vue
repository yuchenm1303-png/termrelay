<template>
  <div class="spg-page">
    <div class="spg-environment" aria-hidden="true"></div>

    <div class="spg-shell">
      <header class="spg-topbar">
        <router-link to="/home" class="spg-brand" aria-label="Smirel 首页">
          <img v-if="siteLogo" :src="siteLogo" alt="" class="spg-brand-logo" />
          <span v-else class="spg-brand-fallback">S</span>
          <span class="spg-brand-copy">
            <strong>{{ siteName }}</strong>
            <small>API SERVICE</small>
          </span>
        </router-link>

        <div class="spg-topbar-right">
          <nav class="spg-nav" aria-label="首页导航">
            <router-link to="/model-plaza">模型与价格</router-link>
            <router-link to="/key-usage">用量查询</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">接入文档</a>
          </nav>
          <div class="spg-status"><i></i><span>服务正常</span></div>
          <router-link :to="accountPath" class="spg-account-link">
            {{ isAuthenticated ? '进入控制台' : '登录' }}
          </router-link>
        </div>
      </header>

      <main>
        <section class="spg-hero">
          <div class="spg-hero-copy">
            <p class="spg-overline">MODEL API SERVICE</p>
            <h1>一个接口，接入主流模型。</h1>
            <p class="spg-hero-description">
              统一 Base URL、统一密钥、统一账单。兼容 OpenAI SDK 与常用客户端，按实际用量计费，现有项目只需替换接入地址即可使用。
            </p>
          </div>
          <div class="spg-hero-meta" aria-label="服务特点">
            <span class="spg-chip spg-chip--live"><i></i>OpenAI-compatible</span>
            <span class="spg-chip">按量计费</span>
            <span class="spg-chip">用量可查</span>
          </div>
        </section>

        <section id="access" class="spg-primary-grid">
          <article class="spg-surface spg-access-card">
            <div class="spg-card-head">
              <div>
                <p class="spg-overline">API ACCESS</p>
                <h2>接入信息</h2>
              </div>
              <span class="spg-pill">HTTPS</span>
            </div>

            <p class="spg-card-intro">
              把现有客户端的 Base URL 指向 Smirel，并使用你自己的 API Key。SDK、请求格式和业务代码无需重写。
            </p>

            <div class="spg-endpoint-block">
              <span>BASE URL</span>
              <strong>{{ apiBase }}</strong>
            </div>

            <div class="spg-state-row">
              <span><i></i>Bearer API Key</span>
              <span><i></i>OpenAI Compatible</span>
              <span><i></i>按实际用量计费</span>
            </div>

            <div class="spg-meta-grid">
              <div class="spg-inset">
                <span>认证</span>
                <strong>独立 API Key</strong>
              </div>
              <div class="spg-inset">
                <span>模型</span>
                <strong>按目录选择</strong>
              </div>
              <div class="spg-inset">
                <span>记录</span>
                <strong>请求与费用可查</strong>
              </div>
            </div>

            <div class="spg-card-actions">
              <router-link :to="startPath" class="spg-action spg-action--primary">
                <span>{{ isAuthenticated ? '打开控制台' : '创建账户' }}</span><b>→</b>
              </router-link>
              <router-link to="/model-plaza" class="spg-action spg-action--secondary">
                <span>查看模型与价格</span><b>↗</b>
              </router-link>
            </div>
          </article>

          <aside class="spg-surface spg-account-card">
            <div class="spg-card-head">
              <div>
                <p class="spg-overline">ACCOUNT</p>
                <h2>开始使用</h2>
              </div>
              <span class="spg-pill">ACCESS</span>
            </div>

            <p class="spg-card-intro">
              注册后即可创建 API Key、管理额度，并查看每一次调用产生的 Token 与费用。
            </p>

            <div class="spg-account-steps" aria-label="开始使用流程">
              <div class="spg-step"><span>01</span><strong>创建账户</strong><small>完成基础账户注册</small></div>
              <div class="spg-step"><span>02</span><strong>创建密钥</strong><small>为项目生成独立 API Key</small></div>
              <div class="spg-step"><span>03</span><strong>开始调用</strong><small>替换 Base URL 后直接使用</small></div>
            </div>

            <div class="spg-account-note">
              <span>现有 OpenAI SDK</span>
              <strong>无需更换客户端</strong>
            </div>

            <router-link :to="accountPath" class="spg-action spg-action--primary spg-account-action">
              <span>{{ isAuthenticated ? '进入我的工作台' : '登录已有账户' }}</span><b>→</b>
            </router-link>

            <div class="spg-card-footer">
              <span>Secure access</span>
              <span>api.smirel.com</span>
            </div>
          </aside>
        </section>

        <section class="spg-utility-grid" aria-label="常用入口">
          <router-link to="/model-plaza" class="spg-surface spg-surface--interactive spg-utility-card">
            <span class="spg-utility-icon">01</span>
            <span class="spg-overline">MODELS</span>
            <h3>模型与价格</h3>
            <p>查看当前可用模型、计费标准与服务状态。</p>
            <b>→</b>
          </router-link>

          <router-link :to="keyPath" class="spg-surface spg-surface--interactive spg-utility-card">
            <span class="spg-utility-icon">02</span>
            <span class="spg-overline">API KEYS</span>
            <h3>密钥与额度</h3>
            <p>为不同项目创建独立密钥，并管理访问额度。</p>
            <b>→</b>
          </router-link>

          <router-link to="/key-usage" class="spg-surface spg-surface--interactive spg-utility-card">
            <span class="spg-utility-icon">03</span>
            <span class="spg-overline">USAGE</span>
            <h3>用量查询</h3>
            <p>按 Key 查看请求、Token、费用和调用结果。</p>
            <b>→</b>
          </router-link>

          <a :href="docUrl || '#access'" :target="docUrl ? '_blank' : undefined" :rel="docUrl ? 'noopener noreferrer' : undefined" class="spg-surface spg-surface--interactive spg-utility-card">
            <span class="spg-utility-icon">04</span>
            <span class="spg-overline">DOCUMENTATION</span>
            <h3>接入文档</h3>
            <p>查看 Base URL、认证方式和常用客户端配置。</p>
            <b>→</b>
          </a>
        </section>

        <section class="spg-surface spg-provider-card">
          <div class="spg-provider-copy">
            <p class="spg-overline">MODEL ACCESS</p>
            <h2>一套凭证，按需选择模型。</h2>
            <p>模型供应商、协议差异和上游账户由服务端统一处理。你只需要维护自己的 API Key。</p>
          </div>
          <div class="spg-provider-list" aria-label="支持的模型服务">
            <div class="spg-inset"><span>OpenAI</span><strong>Responses / Chat</strong></div>
            <div class="spg-inset"><span>Anthropic</span><strong>Claude</strong></div>
            <div class="spg-inset"><span>Google</span><strong>Gemini</strong></div>
            <div class="spg-inset"><span>更多</span><strong>持续接入</strong></div>
          </div>
        </section>
      </main>

      <footer class="spg-footer">
        <span>© {{ currentYear }} {{ siteName }}</span>
        <span class="spg-footer-status"><i></i>Official API service</span>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/smirel-shared-glass-v1.css'
import '@/styles/smirel-shared-interactions-v1.css'

const appStore = useAppStore()
const authStore = useAuthStore()

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

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
})
</script>

<style scoped>
/*
 * Desktop composition: keep the shared glass language, but let the homepage
 * use the viewport as a canvas instead of sitting inside the shared 1360px
 * content column. This is intentionally scoped to the landing page.
 */
@media (min-width: 981px) {
  .spg-shell {
    width: min(1840px, calc(100vw - 32px));
    padding: 20px 0 28px;
  }

  .spg-topbar {
    min-height: 54px;
    margin-bottom: 28px;
    padding-inline: 8px;
  }

  .spg-hero {
    grid-template-columns: minmax(0, 1fr) minmax(300px, auto);
    gap: 36px;
    margin-bottom: 22px;
    padding-inline: 8px;
  }

  .spg-hero-copy {
    max-width: 980px;
  }

  .spg-hero h1 {
    max-width: 980px;
    margin-top: 8px;
    font-size: clamp(2.8rem, 4.7vw, 4.6rem);
    line-height: 1.01;
  }

  .spg-hero-description {
    max-width: 820px;
    margin-top: 14px;
    line-height: 1.62;
  }

  .spg-hero-meta {
    max-width: 390px;
    padding-bottom: 4px;
  }

  .spg-primary-grid {
    grid-template-columns: minmax(0, 1.62fr) minmax(390px, .88fr);
    gap: 14px;
  }

  .spg-access-card,
  .spg-account-card {
    min-height: 420px;
  }

  .spg-access-card {
    padding: 30px 34px 28px;
  }

  .spg-account-card {
    padding: 30px 28px 26px;
  }

  .spg-card-intro {
    margin-top: 14px;
    line-height: 1.6;
  }

  .spg-endpoint-block {
    margin-top: 22px;
    padding: 15px 18px;
  }

  .spg-state-row {
    margin-top: 12px;
    gap: 7px;
  }

  .spg-meta-grid {
    margin-top: 16px;
    gap: 9px;
  }

  .spg-inset {
    padding: 12px 14px;
  }

  .spg-card-actions {
    margin-top: 16px;
    gap: 9px;
  }

  .spg-action {
    min-height: 48px;
  }

  .spg-account-steps {
    margin-top: 18px;
    gap: 7px;
  }

  .spg-step {
    min-height: 57px;
    padding: 9px 11px;
  }

  .spg-account-note {
    margin-top: 8px;
    padding: 9px 11px;
  }

  .spg-card-footer {
    margin-top: 10px;
    padding-top: 10px;
  }

  .spg-utility-grid {
    margin-top: 14px;
    gap: 10px;
  }

  .spg-utility-card {
    min-height: 146px;
    padding: 20px 22px 18px;
  }

  .spg-utility-card > b {
    top: 18px;
    right: 20px;
  }

  .spg-utility-card .spg-overline {
    margin-top: 13px;
  }

  .spg-utility-card h3 {
    margin-top: 5px;
  }

  .spg-utility-card p {
    max-width: 92%;
    margin-top: 6px;
    line-height: 1.5;
  }

  .spg-provider-card {
    grid-template-columns: minmax(320px, .72fr) minmax(0, 1.28fr);
    margin-top: 14px;
    padding: 26px 30px;
    gap: 28px;
  }

  .spg-provider-list {
    gap: 10px;
  }

  .spg-provider-list .spg-inset {
    min-height: 74px;
  }

  .spg-footer {
    min-height: 46px;
    margin-top: 14px;
    padding-inline: 8px;
  }
}

@media (min-width: 981px) and (max-width: 1220px) {
  .spg-shell {
    width: calc(100vw - 28px);
  }

  .spg-primary-grid {
    grid-template-columns: minmax(0, 1.32fr) minmax(340px, .92fr);
  }
}
</style>