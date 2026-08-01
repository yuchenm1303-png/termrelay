<template>
  <div class="auth-shell">
    <div class="auth-grid"></div>
    <div class="auth-scanlines"></div>
    <div class="auth-glow auth-glow-a"></div>
    <div class="auth-glow auth-glow-b"></div>

    <header class="auth-header">
      <router-link to="/home" class="auth-brand" aria-label="TermRelay home">
        <span class="auth-brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="" />
          <span v-else>tr</span>
        </span>
        <span class="auth-brand-copy">
          <strong>{{ siteName }}</strong>
          <small>{{ localizedNodeLabel }}</small>
        </span>
      </router-link>

      <div class="auth-header-actions">
        <LocaleSwitcher />
        <router-link to="/home" class="back-home">
          <span>{{ localizedBackLabel }}</span>
          <span aria-hidden="true">↗</span>
        </router-link>
      </div>
    </header>

    <main class="auth-main">
      <section class="auth-intro" aria-label="TermRelay status">
        <p class="auth-eyebrow"><span></span>{{ localizedEyebrow }}</p>
        <h1>{{ localizedHeadline }}</h1>
        <p class="auth-intro-copy">{{ localizedIntro }}</p>

        <div class="auth-terminal" aria-label="Gateway terminal status">
          <div class="auth-terminal-bar">
            <div class="auth-window-controls" aria-hidden="true">
              <span></span><span></span><span></span>
            </div>
            <span>jack@termrelay: ~/auth</span>
            <strong><i></i>{{ localizedLiveLabel }}</strong>
          </div>

          <div class="auth-terminal-body">
            <div class="auth-command">
              <span class="auth-prompt">jack@termrelay</span><span class="auth-path">:~$</span>
              <span>gateway auth --status</span>
            </div>

            <div class="auth-status-list">
              <div><span>{{ localizedServiceLabel }}</span><strong>TermRelay Gateway</strong></div>
              <div><span>{{ localizedModeLabel }}</span><strong>{{ localizedPrivateLabel }}</strong></div>
              <div><span>{{ localizedProtocolLabel }}</span><strong>Bearer / OAuth</strong></div>
              <div><span>{{ localizedStateLabel }}</span><strong class="auth-ready">{{ localizedReadyLabel }}</strong></div>
            </div>

            <div class="auth-command">
              <span class="auth-prompt">jack@termrelay</span><span class="auth-path">:~$</span>
              <span class="auth-cursor"></span>
            </div>
          </div>
        </div>
      </section>

      <section class="auth-panel" aria-label="Authentication panel">
        <div v-if="settingsLoaded" class="auth-panel-head">
          <span class="auth-panel-index">AUTH / 01</span>
          <div>
            <h2>{{ siteName }}</h2>
            <p>{{ siteSubtitle }}</p>
          </div>
        </div>

        <div class="auth-card">
          <slot />
        </div>

        <div class="auth-footer-slot">
          <slot name="footer" />
        </div>

        <div class="auth-copyright">
          <span>© {{ currentYear }} {{ siteName }}</span>
          <span>{{ localizedBuiltLabel }}</span>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const { locale } = useI18n()

const isChinese = computed(() => locale.value === 'zh')
const siteName = computed(() => appStore.siteName || 'TermRelay')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)
const siteSubtitle = computed(
  () =>
    appStore.cachedPublicSettings?.site_subtitle ||
    (isChinese.value ? '个人 AI 中转节点' : 'Personal AI relay node')
)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const currentYear = computed(() => new Date().getFullYear())

const localizedNodeLabel = computed(() =>
  isChinese.value ? '个人 AI 中转节点' : 'personal ai relay node'
)
const localizedBackLabel = computed(() => (isChinese.value ? '返回首页' : 'back home'))
const localizedEyebrow = computed(() =>
  isChinese.value ? '安全访问 / 节点 01' : 'secure access / node 01'
)
const localizedHeadline = computed(() =>
  isChinese.value ? '进入你的 AI 网关。' : 'Enter your AI gateway.'
)
const localizedIntro = computed(() =>
  isChinese.value
    ? '使用管理员凭证登录，管理 API Key、上游账号、调用记录与节点状态。'
    : 'Sign in with your administrator credentials to manage keys, upstream accounts, request logs and node health.'
)
const localizedLiveLabel = computed(() => (isChinese.value ? '在线' : 'live'))
const localizedServiceLabel = computed(() => (isChinese.value ? '服务' : 'service'))
const localizedModeLabel = computed(() => (isChinese.value ? '模式' : 'mode'))
const localizedPrivateLabel = computed(() => (isChinese.value ? '私有节点' : 'private node'))
const localizedProtocolLabel = computed(() => (isChinese.value ? '协议' : 'protocol'))
const localizedStateLabel = computed(() => (isChinese.value ? '状态' : 'state'))
const localizedReadyLabel = computed(() => (isChinese.value ? '等待登录' : 'awaiting sign-in'))
const localizedBuiltLabel = computed(() =>
  isChinese.value ? '基于 Sub2API 构建' : 'built on Sub2API'
)

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-shell {
  --auth-bg: #06070b;
  --auth-line: rgba(226, 232, 255, 0.12);
  --auth-line-strong: rgba(226, 232, 255, 0.22);
  --auth-text: #f3f0ff;
  --auth-muted: #9d9aae;
  --auth-dim: #6f6c7d;
  --auth-rose: #ebbcba;
  --auth-iris: #c4a7e7;
  --auth-foam: #9ccfd8;
  --auth-pine: #31748f;
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at 70% 10%, rgba(196, 167, 231, 0.11), transparent 34%),
    var(--auth-bg);
  color: var(--auth-text);
}

.auth-shell::before {
  content: "";
  position: fixed;
  inset: 0;
  z-index: 40;
  pointer-events: none;
  opacity: 0.024;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.85'/%3E%3C/svg%3E");
}

.auth-grid,
.auth-scanlines,
.auth-glow {
  position: absolute;
  pointer-events: none;
}

.auth-grid,
.auth-scanlines {
  inset: 0;
}

.auth-grid {
  background-image:
    linear-gradient(rgba(196, 167, 231, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(196, 167, 231, 0.035) 1px, transparent 1px);
  background-size: 54px 54px;
  mask-image: linear-gradient(to bottom, black, transparent 86%);
}

.auth-scanlines {
  z-index: 30;
  opacity: 0.12;
  background: repeating-linear-gradient(
    to bottom,
    transparent 0,
    transparent 3px,
    rgba(255, 255, 255, 0.024) 4px
  );
}

.auth-glow {
  border-radius: 999px;
  filter: blur(110px);
  opacity: 0.12;
}

.auth-glow-a {
  width: 34rem;
  height: 34rem;
  left: -18rem;
  top: 18rem;
  background: var(--auth-pine);
}

.auth-glow-b {
  width: 30rem;
  height: 30rem;
  right: -14rem;
  top: 8rem;
  background: var(--auth-iris);
}

.auth-header,
.auth-main {
  position: relative;
  z-index: 5;
  width: min(1180px, calc(100% - 40px));
  margin-inline: auto;
}

.auth-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 24px 0;
  border-bottom: 1px solid var(--auth-line);
}

.auth-brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: inherit;
  text-decoration: none;
}

.auth-brand-mark {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--auth-line-strong);
  border-radius: 10px;
  background: linear-gradient(145deg, rgba(196, 167, 231, 0.18), rgba(49, 116, 143, 0.1));
  color: var(--auth-iris);
  font: 700 13px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.auth-brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.auth-brand-copy {
  display: grid;
  gap: 2px;
}

.auth-brand-copy strong {
  font-size: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.auth-brand-copy small {
  color: var(--auth-dim);
  font: 500 10px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.auth-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.auth-header-actions :deep(button) {
  border: 1px solid var(--auth-line);
  background: rgba(255, 255, 255, 0.025);
  color: var(--auth-muted);
}

.auth-header-actions :deep(.absolute) {
  border-color: var(--auth-line-strong);
  background: #10121a;
}

.back-home {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 13px;
  border: 1px solid var(--auth-line);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.025);
  color: var(--auth-muted);
  font: 600 10px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-decoration: none;
}

.auth-main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(390px, 0.72fr);
  align-items: center;
  gap: clamp(54px, 8vw, 110px);
  min-height: calc(100vh - 90px);
  padding: 62px 0 74px;
}

.auth-intro {
  max-width: 650px;
}

.auth-eyebrow {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 22px;
  color: var(--auth-foam);
  font: 600 11px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.11em;
  text-transform: uppercase;
}

.auth-eyebrow span {
  width: 26px;
  height: 1px;
  background: currentColor;
  box-shadow: 0 0 10px currentColor;
}

.auth-intro h1 {
  max-width: 720px;
  margin: 0;
  font-size: clamp(48px, 6.5vw, 82px);
  font-weight: 620;
  line-height: 0.98;
  letter-spacing: -0.06em;
}

.auth-intro-copy {
  max-width: 590px;
  margin: 27px 0 0;
  color: var(--auth-muted);
  font-size: 17px;
  line-height: 1.75;
}

.auth-terminal {
  margin-top: 38px;
  overflow: hidden;
  border: 1px solid var(--auth-line-strong);
  border-radius: 13px;
  background: linear-gradient(150deg, rgba(18, 20, 29, 0.94), rgba(7, 8, 12, 0.95));
  box-shadow: 0 32px 90px rgba(0, 0, 0, 0.45), inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.auth-terminal-bar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  min-height: 42px;
  padding: 0 13px;
  border-bottom: 1px solid var(--auth-line);
  color: var(--auth-dim);
  font: 500 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.auth-window-controls {
  display: flex;
  gap: 7px;
}

.auth-window-controls span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--auth-dim);
}

.auth-window-controls span:first-child { background: var(--auth-rose); }
.auth-window-controls span:nth-child(2) { background: #f6c177; }
.auth-window-controls span:last-child { background: var(--auth-foam); }

.auth-terminal-bar strong {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  color: var(--auth-foam);
  font-weight: 600;
  text-transform: uppercase;
}

.auth-terminal-bar strong i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 8px currentColor;
}

.auth-terminal-body {
  padding: 24px 26px 27px;
  font: 500 11px/1.9 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.auth-command {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  color: #d7d3e3;
}

.auth-prompt { color: var(--auth-foam); }
.auth-path { color: var(--auth-iris); }

.auth-status-list {
  display: grid;
  gap: 3px;
  margin: 16px 0 21px;
  padding: 13px 15px;
  border-left: 1px solid var(--auth-pine);
  background: rgba(49, 116, 143, 0.055);
}

.auth-status-list div {
  display: grid;
  grid-template-columns: 90px minmax(0, 1fr);
  gap: 12px;
}

.auth-status-list span {
  color: var(--auth-dim);
}

.auth-status-list strong {
  overflow: hidden;
  color: #c8c4d5;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.auth-status-list .auth-ready {
  color: var(--auth-rose);
}

.auth-cursor {
  width: 8px;
  height: 14px;
  margin-top: 4px;
  background: var(--auth-foam);
  box-shadow: 0 0 12px rgba(156, 207, 216, 0.55);
  animation: auth-blink 1s step-end infinite;
}

@keyframes auth-blink { 50% { opacity: 0; } }

.auth-panel {
  width: 100%;
  max-width: 470px;
  justify-self: end;
}

.auth-panel-head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 16px;
  align-items: start;
  margin-bottom: 15px;
}

.auth-panel-index {
  padding-top: 6px;
  color: var(--auth-foam);
  font: 600 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.auth-panel-head h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 620;
  letter-spacing: -0.02em;
}

.auth-panel-head p {
  margin: 5px 0 0;
  color: var(--auth-dim);
  font-size: 12px;
  line-height: 1.55;
}

.auth-card {
  padding: 30px;
  border: 1px solid var(--auth-line-strong);
  border-radius: 14px;
  background: rgba(12, 14, 20, 0.82);
  box-shadow: 0 28px 80px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.045);
  backdrop-filter: blur(16px);
}

.auth-card :deep(h2),
.auth-card :deep(label),
.auth-card :deep(.text-gray-900),
.auth-card :deep(.dark\:text-white) {
  color: var(--auth-text) !important;
}

.auth-card :deep(p),
.auth-card :deep(.text-gray-500),
.auth-card :deep(.dark\:text-dark-400) {
  color: var(--auth-muted) !important;
}

.auth-card :deep(input) {
  border-color: var(--auth-line) !important;
  background: rgba(255, 255, 255, 0.028) !important;
  color: var(--auth-text) !important;
}

.auth-card :deep(input::placeholder) {
  color: var(--auth-dim) !important;
}

.auth-card :deep(.btn-primary) {
  border-color: transparent !important;
  background: var(--auth-text) !important;
  color: #090a0e !important;
  box-shadow: none !important;
}

.auth-card :deep(.btn-secondary) {
  border-color: var(--auth-line) !important;
  background: rgba(255, 255, 255, 0.03) !important;
  color: var(--auth-text) !important;
}

.auth-card :deep(a) {
  color: var(--auth-iris) !important;
}

.auth-footer-slot {
  margin-top: 17px;
  text-align: center;
  font-size: 13px;
}

.auth-footer-slot :deep(p) {
  color: var(--auth-muted) !important;
}

.auth-footer-slot :deep(a) {
  color: var(--auth-iris) !important;
}

.auth-copyright {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
  color: var(--auth-dim);
  font: 500 8px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

@media (max-width: 940px) {
  .auth-main {
    grid-template-columns: 1fr;
    padding-top: 76px;
  }

  .auth-intro {
    max-width: 720px;
  }

  .auth-panel {
    max-width: 620px;
    justify-self: start;
  }
}

@media (max-width: 680px) {
  .auth-header,
  .auth-main {
    width: min(100% - 28px, 1180px);
  }

  .auth-brand-copy small,
  .back-home span:first-child {
    display: none;
  }

  .auth-main {
    gap: 54px;
    padding: 52px 0 64px;
  }

  .auth-intro h1 {
    font-size: clamp(46px, 15vw, 68px);
  }

  .auth-terminal-body {
    padding-inline: 18px;
    font-size: 10px;
  }

  .auth-card {
    padding: 23px;
  }

  .auth-copyright {
    flex-direction: column;
  }
}

@media (max-width: 450px) {
  .auth-brand-copy {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
