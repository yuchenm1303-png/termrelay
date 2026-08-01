<template>
  <div v-if="hasHomeContent" class="min-h-screen bg-[#05070b]">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="termrelay-shell">
    <div class="ambient ambient-a"></div>
    <div class="ambient ambient-b"></div>
    <div class="grid-layer"></div>
    <div class="scanlines"></div>

    <header class="site-header">
      <router-link to="/home" class="brand" aria-label="TermRelay home">
        <span class="brand-mark">
          <img v-if="siteLogo" :src="siteLogo" alt="" />
          <span v-else>tr</span>
        </span>
        <span class="brand-copy">
          <strong>{{ siteName }}</strong>
          <small>personal ai relay node</small>
        </span>
      </router-link>

      <nav class="site-nav" aria-label="Primary navigation">
        <a href="#system">system</a>
        <a href="#capabilities">capabilities</a>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">docs</a>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-cta">
          {{ isAuthenticated ? 'open console' : 'sign in' }}
        </router-link>
      </nav>
    </header>

    <main>
      <section class="hero" id="system">
        <div class="hero-copy">
          <p class="eyebrow"><span></span> self-hosted gateway / node 01</p>
          <h1>
            Your models.
            <em>Your endpoint.</em>
          </h1>
          <p class="hero-description">
            {{ siteSubtitle }}
          </p>

          <div class="hero-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="primary-action">
              <span>{{ isAuthenticated ? 'Enter console' : 'Initialize session' }}</span>
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
            </router-link>
            <button type="button" class="secondary-action" @click="copyBaseUrl">
              <span>{{ copied ? 'Copied' : 'Copy base URL' }}</span>
              <code>{{ baseUrl }}</code>
            </button>
          </div>

          <div class="signal-row" aria-label="Gateway capabilities">
            <span><i class="signal online"></i> gateway reachable</span>
            <span><i class="signal"></i> responses api</span>
            <span><i class="signal"></i> oauth upstream</span>
          </div>
        </div>

        <div class="terminal-wrap" aria-label="TermRelay terminal preview">
          <div class="terminal-glow"></div>
          <div class="terminal-window">
            <div class="terminal-bar">
              <div class="window-controls" aria-hidden="true">
                <span></span><span></span><span></span>
              </div>
              <div class="terminal-title">jack@termrelay: ~/gateway</div>
              <div class="terminal-state"><i></i> live</div>
            </div>

            <div class="terminal-body">
              <div class="terminal-line delay-1">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span>status --verbose</span>
              </div>
              <div class="terminal-output delay-2">
                <div><span>service</span><strong>TermRelay Gateway</strong></div>
                <div><span>transport</span><strong>OpenAI Responses</strong></div>
                <div><span>auth</span><strong>Bearer / OAuth upstream</strong></div>
                <div><span>endpoint</span><strong>{{ baseUrl }}</strong></div>
              </div>
              <div class="terminal-line delay-3">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span>curl {{ baseUrl }}/models</span>
              </div>
              <div class="terminal-json delay-4">
                <span>{</span>
                <span class="indent">"object": "list",</span>
                <span class="indent">"route": "ready",</span>
                <span class="indent">"stream": true</span>
                <span>}</span>
              </div>
              <div class="terminal-line delay-5">
                <span class="prompt">jack@termrelay</span><span class="path">:~$</span>
                <span class="cursor"></span>
              </div>
            </div>
          </div>

          <div class="terminal-meta">
            <span>NODE</span><strong>TR-01</strong>
            <span>MODE</span><strong>PRIVATE</strong>
            <span>TIME</span><strong>{{ clock }}</strong>
          </div>
        </div>
      </section>

      <section class="command-strip" aria-label="Quick start command">
        <div class="command-label">quick start</div>
        <code><span>$</span> export OPENAI_BASE_URL={{ baseUrl }}</code>
        <button type="button" @click="copyBaseUrl">{{ copied ? 'done' : 'copy' }}</button>
      </section>

      <section class="capabilities" id="capabilities">
        <div class="section-heading">
          <p>cat capabilities.md</p>
          <h2>Built as a gateway, presented as a personal node.</h2>
        </div>

        <div class="capability-grid">
          <article>
            <div class="article-index">01</div>
            <div>
              <h3>Unified endpoint</h3>
              <p>Expose one stable API base URL while the gateway handles upstream routing and protocol compatibility.</p>
            </div>
            <span class="article-tag">/v1</span>
          </article>

          <article>
            <div class="article-index">02</div>
            <div>
              <h3>Credential control</h3>
              <p>Keep downstream keys separate from upstream OAuth credentials, with revocation and usage visibility.</p>
            </div>
            <span class="article-tag">auth</span>
          </article>

          <article>
            <div class="article-index">03</div>
            <div>
              <h3>Streaming native</h3>
              <p>Forward long-running Responses API streams without turning the interface into a generic SaaS dashboard.</p>
            </div>
            <span class="article-tag">sse</span>
          </article>

          <article>
            <div class="article-index">04</div>
            <div>
              <h3>Observable by default</h3>
              <p>Track requests, latency, failures and account health from a focused operator console.</p>
            </div>
            <span class="article-tag">logs</span>
          </article>
        </div>
      </section>

      <section class="manifesto">
        <div class="manifesto-label">README / 00</div>
        <p>
          TermRelay is not designed to look like another anonymous API marketplace.
          It is a small, self-hosted AI gateway with the character of a personal terminal.
        </p>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'">
          launch console <span>↗</span>
        </router-link>
      </section>
    </main>

    <footer>
      <div>
        <span>© {{ currentYear }} {{ siteName }}</span>
        <span>built on Sub2API</span>
      </div>
      <a href="https://github.com/yuchenm1303-png/termrelay" target="_blank" rel="noopener noreferrer">
        github / source
      </a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(
  () => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TermRelay'
)
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true
  })
)
const siteSubtitle = computed(
  () =>
    appStore.cachedPublicSettings?.site_subtitle ||
    'A terminal-inspired AI relay gateway for your own models, credentials and clients.'
)
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const currentYear = new Date().getFullYear()
const baseUrl = computed(() =>
  typeof window === 'undefined' ? '/v1' : `${window.location.origin}/v1`
)

const copied = ref(false)
const clock = ref('00:00:00')
let clockTimer: number | undefined
let copiedTimer: number | undefined

function updateClock() {
  clock.value = new Intl.DateTimeFormat('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).format(new Date())
}

async function copyBaseUrl() {
  try {
    await navigator.clipboard.writeText(baseUrl.value)
    copied.value = true
    if (copiedTimer) window.clearTimeout(copiedTimer)
    copiedTimer = window.setTimeout(() => {
      copied.value = false
    }, 1800)
  } catch {
    copied.value = false
  }
}

onMounted(() => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

onBeforeUnmount(() => {
  if (clockTimer) window.clearInterval(clockTimer)
  if (copiedTimer) window.clearTimeout(copiedTimer)
})
</script>

<style scoped>
.termrelay-shell {
  --bg: #06070b;
  --panel: rgba(14, 16, 23, 0.78);
  --panel-solid: #0d0f16;
  --line: rgba(226, 232, 255, 0.12);
  --line-strong: rgba(226, 232, 255, 0.22);
  --text: #f3f0ff;
  --muted: #9d9aae;
  --dim: #6f6c7d;
  --rose: #ebbcba;
  --iris: #c4a7e7;
  --foam: #9ccfd8;
  --pine: #31748f;
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% -20%, rgba(196, 167, 231, 0.13), transparent 38%),
    var(--bg);
  color: var(--text);
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.termrelay-shell::before {
  content: "";
  position: fixed;
  inset: 0;
  z-index: 30;
  pointer-events: none;
  opacity: 0.025;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.85'/%3E%3C/svg%3E");
}

.ambient,
.grid-layer,
.scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.ambient {
  border-radius: 999px;
  filter: blur(100px);
  opacity: 0.13;
}

.ambient-a {
  width: 34rem;
  height: 34rem;
  left: -18rem;
  top: 16rem;
  background: var(--pine);
}

.ambient-b {
  width: 30rem;
  height: 30rem;
  right: -14rem;
  top: 4rem;
  background: var(--iris);
}

.grid-layer {
  background-image:
    linear-gradient(rgba(196, 167, 231, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(196, 167, 231, 0.035) 1px, transparent 1px);
  background-size: 54px 54px;
  mask-image: linear-gradient(to bottom, black, transparent 82%);
}

.scanlines {
  z-index: 20;
  opacity: 0.13;
  background: repeating-linear-gradient(
    to bottom,
    transparent 0,
    transparent 3px,
    rgba(255, 255, 255, 0.025) 4px
  );
}

.site-header,
.hero,
.command-strip,
.capabilities,
.manifesto,
footer {
  position: relative;
  z-index: 5;
  width: min(1180px, calc(100% - 40px));
  margin-inline: auto;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 0;
  border-bottom: 1px solid var(--line);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: inherit;
  text-decoration: none;
}

.brand-mark {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  background: linear-gradient(145deg, rgba(196, 167, 231, 0.18), rgba(49, 116, 143, 0.1));
  color: var(--iris);
  font: 700 13px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.brand-copy {
  display: grid;
  gap: 2px;
}

.brand-copy strong {
  font-size: 14px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.brand-copy small {
  color: var(--dim);
  font: 500 10px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.05em;
}

.site-nav {
  display: flex;
  align-items: center;
  gap: 25px;
}

.site-nav a {
  color: var(--muted);
  font: 500 12px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-decoration: none;
  transition: color 160ms ease;
}

.site-nav a:hover {
  color: var(--text);
}

.site-nav .nav-cta {
  padding: 10px 14px;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.035);
  color: var(--text);
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.02fr) minmax(430px, 0.98fr);
  align-items: center;
  gap: clamp(48px, 8vw, 110px);
  min-height: 690px;
  padding: 78px 0 62px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 22px;
  color: var(--foam);
  font: 600 11px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.11em;
  text-transform: uppercase;
}

.eyebrow span {
  width: 26px;
  height: 1px;
  background: currentColor;
  box-shadow: 0 0 10px currentColor;
}

.hero h1 {
  max-width: 720px;
  margin: 0;
  font-size: clamp(52px, 7vw, 90px);
  font-weight: 620;
  line-height: 0.94;
  letter-spacing: -0.065em;
}

.hero h1 em {
  display: block;
  margin-top: 12px;
  color: transparent;
  font-style: normal;
  -webkit-text-stroke: 1px rgba(235, 188, 186, 0.88);
  text-shadow: 0 0 42px rgba(235, 188, 186, 0.12);
}

.hero-description {
  max-width: 610px;
  margin: 30px 0 0;
  color: var(--muted);
  font-size: clamp(16px, 1.7vw, 19px);
  line-height: 1.75;
}

.hero-actions {
  display: flex;
  align-items: stretch;
  gap: 12px;
  margin-top: 34px;
}

.primary-action,
.secondary-action {
  min-height: 54px;
  border-radius: 10px;
  font: 600 12px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.primary-action {
  display: inline-flex;
  align-items: center;
  gap: 24px;
  padding: 0 20px;
  background: var(--text);
  color: #090a0e;
  text-decoration: none;
  transition: transform 180ms ease, box-shadow 180ms ease;
}

.primary-action:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 40px rgba(196, 167, 231, 0.13);
}

.primary-action svg {
  width: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.7;
}

.secondary-action {
  display: grid;
  gap: 5px;
  padding: 9px 16px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.025);
  color: var(--text);
  text-align: left;
  cursor: pointer;
}

.secondary-action code {
  max-width: 220px;
  overflow: hidden;
  color: var(--dim);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.signal-row {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  margin-top: 28px;
  color: var(--dim);
  font: 500 10px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.signal-row span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.signal {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--dim);
}

.signal.online {
  background: var(--foam);
  box-shadow: 0 0 10px var(--foam);
}

.terminal-wrap {
  position: relative;
}

.terminal-glow {
  position: absolute;
  inset: 8% 3% 12%;
  border-radius: 30px;
  background: rgba(196, 167, 231, 0.12);
  filter: blur(55px);
}

.terminal-window {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 14px;
  background: linear-gradient(150deg, rgba(18, 20, 29, 0.96), rgba(7, 8, 12, 0.96));
  box-shadow:
    0 40px 100px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
  transform: perspective(1200px) rotateY(-4deg) rotateX(1.5deg);
}

.terminal-bar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  min-height: 44px;
  padding: 0 14px;
  border-bottom: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.025);
}

.window-controls {
  display: flex;
  gap: 7px;
}

.window-controls span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #4c4a57;
}

.window-controls span:first-child { background: var(--rose); }
.window-controls span:nth-child(2) { background: #f6c177; }
.window-controls span:last-child { background: var(--foam); }

.terminal-title,
.terminal-state {
  font: 500 9px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.terminal-title {
  color: var(--dim);
}

.terminal-state {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  color: var(--foam);
  text-transform: uppercase;
}

.terminal-state i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 8px currentColor;
}

.terminal-body {
  min-height: 390px;
  padding: 26px 28px 32px;
  font: 500 12px/1.9 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.terminal-line,
.terminal-output,
.terminal-json {
  opacity: 0;
  animation: terminal-in 0.45s ease forwards;
}

.delay-1 { animation-delay: 0.2s; }
.delay-2 { animation-delay: 0.65s; }
.delay-3 { animation-delay: 1.1s; }
.delay-4 { animation-delay: 1.55s; }
.delay-5 { animation-delay: 2s; }

@keyframes terminal-in {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}

.terminal-line {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  color: #d7d3e3;
}

.prompt { color: var(--foam); }
.path { color: var(--iris); }

.terminal-output {
  display: grid;
  gap: 2px;
  margin: 17px 0 22px;
  padding: 14px 16px;
  border-left: 1px solid var(--pine);
  background: rgba(49, 116, 143, 0.055);
}

.terminal-output div {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
}

.terminal-output span {
  color: var(--dim);
}

.terminal-output strong {
  overflow: hidden;
  color: #c8c4d5;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-json {
  display: grid;
  margin: 13px 0 22px;
  color: var(--rose);
}

.terminal-json .indent {
  padding-left: 18px;
  color: #b9b5c5;
}

.cursor {
  width: 8px;
  height: 15px;
  margin-top: 4px;
  background: var(--foam);
  box-shadow: 0 0 12px rgba(156, 207, 216, 0.55);
  animation: blink 1s step-end infinite;
}

@keyframes blink { 50% { opacity: 0; } }

.terminal-meta {
  display: grid;
  grid-template-columns: auto 1fr auto 1fr auto 1fr;
  gap: 8px 12px;
  margin: 13px 12px 0;
  color: var(--dim);
  font: 500 8px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.terminal-meta strong {
  color: var(--muted);
  font-weight: 600;
}

.command-strip {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 22px;
  min-height: 68px;
  padding: 0 22px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.025);
}

.command-label,
.command-strip code,
.command-strip button {
  font: 500 10px/1.2 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.command-label {
  color: var(--dim);
  text-transform: uppercase;
}

.command-strip code {
  overflow: hidden;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.command-strip code span {
  margin-right: 10px;
  color: var(--foam);
}

.command-strip button {
  border: 0;
  background: transparent;
  color: var(--iris);
  cursor: pointer;
}

.capabilities {
  padding: 120px 0 90px;
}

.section-heading {
  display: grid;
  grid-template-columns: 0.55fr 1.45fr;
  gap: 40px;
  align-items: start;
  margin-bottom: 46px;
}

.section-heading p,
.manifesto-label {
  margin: 8px 0 0;
  color: var(--foam);
  font: 600 10px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

.section-heading h2 {
  max-width: 760px;
  margin: 0;
  font-size: clamp(32px, 4.5vw, 55px);
  font-weight: 560;
  line-height: 1.08;
  letter-spacing: -0.045em;
}

.capability-grid {
  border-top: 1px solid var(--line);
}

.capability-grid article {
  display: grid;
  grid-template-columns: 70px minmax(0, 1fr) auto;
  gap: 26px;
  align-items: start;
  padding: 30px 0;
  border-bottom: 1px solid var(--line);
  transition: padding 180ms ease, background 180ms ease;
}

.capability-grid article:hover {
  padding-inline: 18px;
  background: rgba(255, 255, 255, 0.018);
}

.article-index,
.article-tag {
  color: var(--dim);
  font: 500 10px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.capability-grid h3 {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
}

.capability-grid p {
  max-width: 700px;
  margin: 0;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.7;
}

.article-tag {
  padding: 6px 8px;
  border: 1px solid var(--line);
  border-radius: 6px;
  color: var(--iris);
}

.manifesto {
  display: grid;
  grid-template-columns: 0.35fr 1.25fr auto;
  gap: 42px;
  align-items: end;
  padding: 65px 0;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
}

.manifesto p {
  margin: 0;
  color: #d4d0df;
  font-size: clamp(22px, 2.7vw, 34px);
  line-height: 1.35;
  letter-spacing: -0.025em;
}

.manifesto a {
  color: var(--rose);
  font: 600 11px/1 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-decoration: none;
  white-space: nowrap;
}

.manifesto a span {
  margin-left: 7px;
}

footer {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 28px 0 40px;
  color: var(--dim);
  font: 500 9px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  text-transform: uppercase;
}

footer div {
  display: flex;
  gap: 18px;
}

footer a {
  color: var(--muted);
  text-decoration: none;
}

@media (max-width: 940px) {
  .hero {
    grid-template-columns: 1fr;
    min-height: auto;
    padding-top: 90px;
  }

  .terminal-wrap {
    width: min(100%, 620px);
  }

  .terminal-window {
    transform: none;
  }

  .section-heading,
  .manifesto {
    grid-template-columns: 1fr;
  }

  .manifesto {
    align-items: start;
  }
}

@media (max-width: 680px) {
  .site-header,
  .hero,
  .command-strip,
  .capabilities,
  .manifesto,
  footer {
    width: min(100% - 28px, 1180px);
  }

  .site-nav > a:not(.nav-cta) {
    display: none;
  }

  .hero {
    gap: 52px;
    padding: 66px 0 48px;
  }

  .hero h1 {
    font-size: clamp(48px, 15vw, 70px);
  }

  .hero-actions {
    flex-direction: column;
  }

  .primary-action {
    justify-content: space-between;
  }

  .secondary-action code {
    max-width: 100%;
  }

  .terminal-body {
    min-height: 350px;
    padding: 22px 18px 28px;
    font-size: 10px;
  }

  .terminal-output div {
    grid-template-columns: 72px minmax(0, 1fr);
  }

  .terminal-meta {
    grid-template-columns: auto 1fr auto 1fr;
  }

  .terminal-meta span:nth-of-type(3),
  .terminal-meta strong:nth-of-type(3) {
    display: none;
  }

  .command-strip {
    grid-template-columns: 1fr auto;
    gap: 10px;
    padding: 15px;
  }

  .command-label {
    display: none;
  }

  .capabilities {
    padding-top: 90px;
  }

  .capability-grid article {
    grid-template-columns: 38px minmax(0, 1fr);
    gap: 16px;
  }

  .article-tag {
    display: none;
  }

  footer,
  footer div {
    flex-direction: column;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
