<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

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
const consolePath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const primaryPath = computed(() => authStore.isAuthenticated ? consolePath.value : '/register')
const primaryLabel = computed(() => authStore.isAuthenticated ? '进入控制台' : '开始使用')
</script>

<template>
  <div class="sr-home smg-shell">
    <div class="smg-environment" aria-hidden="true"></div>
    <div class="sr-home__veil" aria-hidden="true"></div>

    <header class="sr-nav">
      <router-link to="/home" class="sr-brand" aria-label="Smirel Home">
        <span class="sr-brand__mark">
          <img v-if="siteLogo" :src="siteLogo" :alt="siteName" />
          <span v-else>S</span>
        </span>
        <span class="sr-brand__copy">
          <strong>{{ siteName }}</strong>
          <small>AI Gateway</small>
        </span>
      </router-link>

      <nav class="sr-nav__links" aria-label="Primary navigation">
        <a href="#gateway">网关</a>
        <router-link to="/model-plaza">模型与价格</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">文档</a>
        <router-link v-else to="/key-usage">API</router-link>
      </nav>

      <div class="sr-nav__actions">
        <span class="sr-status"><i></i> Gateway online</span>
        <router-link :to="authStore.isAuthenticated ? consolePath : '/login'" class="sr-nav__signin">
          {{ authStore.isAuthenticated ? '控制台' : '登录' }}
        </router-link>
      </div>
    </header>

    <main>
      <section class="sr-hero" aria-labelledby="sr-hero-title">
        <div class="sr-hero__copy">
          <div class="sr-eyebrow"><span>SMIREL</span><i></i><span>UNIFIED AI ACCESS</span></div>
          <h1 id="sr-hero-title">
            一个接口，<br />
            <span>连接所有模型。</span>
          </h1>
          <p class="sr-hero__lead">
            把模型选择、上游账号、故障切换和用量统计留给网关。你的应用只需要一个 Base URL 和一把 API Key。
          </p>

          <div class="sr-hero__actions">
            <router-link :to="primaryPath" class="sr-button sr-button--primary">
              {{ primaryLabel }}
              <span aria-hidden="true">↗</span>
            </router-link>
            <router-link to="/model-plaza" class="sr-button sr-button--ghost">查看模型</router-link>
          </div>

          <div class="sr-endpoint">
            <span>BASE URL</span>
            <code>{{ apiBase }}</code>
            <span class="sr-endpoint__dot"></span>
          </div>
        </div>

        <div class="sr-gateway-stage" id="gateway">
          <div class="sr-gateway-card">
            <div class="sr-gateway-card__head">
              <div>
                <span class="sr-gateway-card__label">LIVE REQUEST PATH</span>
                <strong>Smirel Gateway</strong>
              </div>
              <span class="sr-live"><i></i> LIVE</span>
            </div>

            <div class="sr-route-map" aria-label="Request routing visualization">
              <div class="sr-route-map__origin">
                <span>YOUR APP</span>
                <strong>POST /v1/responses</strong>
                <small>Bearer sk-••••••••</small>
              </div>

              <div class="sr-route-map__spine">
                <i></i><i></i><i></i>
              </div>

              <div class="sr-route-map__core">
                <span class="sr-core-orbit sr-core-orbit--one"></span>
                <span class="sr-core-orbit sr-core-orbit--two"></span>
                <span class="sr-core-dot"></span>
                <strong>SR</strong>
              </div>

              <div class="sr-route-map__branches">
                <div class="sr-provider sr-provider--active"><i></i><span>OpenAI</span><small>12 ms</small></div>
                <div class="sr-provider"><i></i><span>Anthropic</span><small>ready</small></div>
                <div class="sr-provider"><i></i><span>Gemini</span><small>ready</small></div>
              </div>
            </div>

            <div class="sr-gateway-card__foot">
              <div><span>Protocol</span><strong>OpenAI compatible</strong></div>
              <div><span>Routing</span><strong>Automatic</strong></div>
              <div><span>Failover</span><strong>Enabled</strong></div>
            </div>
          </div>

          <span class="sr-stage-note sr-stage-note--top">01 / ROUTE</span>
          <span class="sr-stage-note sr-stage-note--bottom">INFRASTRUCTURE, ABSTRACTED.</span>
        </div>
      </section>

      <section class="sr-proof" aria-label="Gateway capabilities">
        <div class="sr-proof__intro">
          <span>ONE GATEWAY</span>
          <strong>把复杂度停在你的应用之外。</strong>
        </div>
        <div class="sr-proof__item"><strong>01</strong><span>统一接口</span><p>同一个调用方式访问不同模型。</p></div>
        <div class="sr-proof__item"><strong>02</strong><span>自动路由</span><p>上游切换与账号调度由服务端完成。</p></div>
        <div class="sr-proof__item"><strong>03</strong><span>完整观测</span><p>请求、Token、费用和错误集中记录。</p></div>
      </section>

      <section class="sr-workspace">
        <div class="sr-section-heading">
          <span>WORKSPACE</span>
          <h2>只留下真正需要管理的东西。</h2>
          <p>Key、模型、用量和余额属于你的工作区；账号池、路由与故障恢复留在 Smirel 后面。</p>
        </div>

        <div class="sr-workspace-grid">
          <article class="sr-workspace-card sr-workspace-card--wide">
            <div class="sr-card-index">01</div>
            <div class="sr-card-copy"><span>ACCESS</span><h3>API Keys</h3><p>创建独立凭证，控制额度与访问范围。</p></div>
            <div class="sr-key-visual"><span>sk-smirel</span><i></i><i></i><i></i><strong>•••• 7A2F</strong></div>
          </article>

          <article class="sr-workspace-card">
            <div class="sr-card-index">02</div>
            <div class="sr-card-copy"><span>MODELS</span><h3>Model Catalog</h3><p>能力、协议与价格在一个目录里完成选择。</p></div>
            <div class="sr-model-stack"><span>GPT</span><span>Claude</span><span>Gemini</span><span>+</span></div>
          </article>

          <article class="sr-workspace-card">
            <div class="sr-card-index">03</div>
            <div class="sr-card-copy"><span>OBSERVE</span><h3>Usage</h3><p>从单次请求回溯到 Token 与费用。</p></div>
            <div class="sr-bars"><i style="--h:31%"></i><i style="--h:52%"></i><i style="--h:41%"></i><i style="--h:72%"></i><i style="--h:58%"></i><i style="--h:88%"></i><i style="--h:66%"></i></div>
          </article>
        </div>
      </section>

      <section class="sr-quickstart">
        <div class="sr-quickstart__copy">
          <span>QUICK START</span>
          <h2>换一个地址，继续用你熟悉的 SDK。</h2>
          <p>兼容 OpenAI 风格的调用方式。无需为每个模型维护一套新的接入逻辑。</p>
          <router-link :to="primaryPath">创建 API Key <span>→</span></router-link>
        </div>
        <div class="sr-code-panel">
          <div class="sr-code-panel__head"><span>request.py</span><small>Python</small></div>
          <pre><code><span class="sr-code-dim">from</span> openai <span class="sr-code-dim">import</span> OpenAI

client = OpenAI(
    base_url=<b>"{{ apiBase }}"</b>,
    api_key=<b>"sk-..."</b>,
)

response = client.responses.create(
    model=<b>"gpt-5"</b>,
    input=<b>"Hello, Smirel"</b>,
)</code></pre>
        </div>
      </section>
    </main>

    <footer class="sr-footer">
      <div class="sr-footer__brand"><span class="sr-brand__mark sr-brand__mark--small">S</span><strong>{{ siteName }}</strong></div>
      <p>Unified AI gateway for people who would rather build products than maintain upstream complexity.</p>
      <div class="sr-footer__links"><router-link to="/model-plaza">Models</router-link><router-link to="/key-usage">API</router-link><router-link :to="authStore.isAuthenticated ? consolePath : '/login'">Console</router-link></div>
    </footer>
  </div>
</template>

<style scoped>
.sr-home {
  --sr-white: #f7fbff;
  --sr-muted: rgba(237, 246, 252, .64);
  --sr-soft: rgba(232, 244, 252, .42);
  --sr-line: rgba(255, 255, 255, .11);
  --sr-line-strong: rgba(255, 255, 255, .18);
  --sr-panel: rgba(7, 20, 30, .42);
  --sr-panel-deep: rgba(3, 13, 21, .56);
  --sr-accent: #b9e8ff;
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: var(--sr-white);
}

.sr-home__veil {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background: linear-gradient(180deg, rgba(7, 18, 27, .06) 0%, rgba(7, 18, 27, .12) 52%, rgba(5, 14, 22, .42) 100%);
}

.sr-nav {
  position: relative;
  z-index: 10;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  width: min(1440px, calc(100% - 64px));
  min-height: 88px;
  margin: 0 auto;
  border-bottom: 1px solid rgba(255,255,255,.09);
}

.sr-brand, .sr-nav__links, .sr-nav__actions, .sr-footer__brand, .sr-footer__links { display: flex; align-items: center; }
.sr-brand { width: max-content; gap: 12px; color: inherit; text-decoration: none; }
.sr-brand__mark { display: grid; width: 38px; height: 38px; place-items: center; overflow: hidden; border: 1px solid rgba(255,255,255,.17); border-radius: 12px; background: rgba(255,255,255,.08); box-shadow: inset 0 1px rgba(255,255,255,.08); font-size: 13px; font-weight: 760; }
.sr-brand__mark img { width: 100%; height: 100%; object-fit: contain; }
.sr-brand__mark--small { width: 28px; height: 28px; border-radius: 9px; }
.sr-brand__copy { display: flex; flex-direction: column; gap: 1px; }
.sr-brand__copy strong { font-size: 14px; letter-spacing: -.02em; }
.sr-brand__copy small { color: var(--sr-soft); font-size: 8px; letter-spacing: .18em; text-transform: uppercase; }
.sr-nav__links { gap: 34px; }
.sr-nav__links a, .sr-nav__signin { color: rgba(246,251,255,.68); font-size: 12px; text-decoration: none; transition: color .2s ease; }
.sr-nav__links a:hover, .sr-nav__signin:hover { color: #fff; }
.sr-nav__actions { justify-content: flex-end; gap: 20px; }
.sr-status { display: inline-flex; align-items: center; gap: 7px; color: var(--sr-muted); font-size: 10px; letter-spacing: .04em; }
.sr-status i, .sr-live i { width: 6px; height: 6px; border-radius: 50%; background: #91dfb6; box-shadow: 0 0 12px rgba(145,223,182,.72); }
.sr-nav__signin { padding: 9px 14px; border: 1px solid rgba(255,255,255,.12); border-radius: 9px; background: rgba(8,20,30,.20); }

.sr-hero {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, .92fr) minmax(520px, .78fr);
  align-items: center;
  gap: clamp(44px, 7vw, 120px);
  width: min(1440px, calc(100% - 64px));
  min-height: 690px;
  margin: 0 auto;
  padding: 88px 0 104px;
}

.sr-hero::after { content: ''; position: absolute; left: 0; right: 0; bottom: 0; height: 1px; background: linear-gradient(90deg, transparent, rgba(255,255,255,.16) 16%, rgba(255,255,255,.16) 84%, transparent); }
.sr-eyebrow { display: flex; align-items: center; gap: 12px; margin-bottom: 28px; color: rgba(237,247,253,.55); font-size: 9px; font-weight: 650; letter-spacing: .2em; }
.sr-eyebrow i { width: 32px; height: 1px; background: rgba(255,255,255,.2); }
.sr-hero h1 { max-width: 760px; margin: 0; color: #fff; font-size: clamp(52px, 6vw, 92px); font-weight: 540; letter-spacing: -.066em; line-height: .98; text-wrap: balance; }
.sr-hero h1 span { color: rgba(245,250,254,.54); }
.sr-hero__lead { max-width: 610px; margin: 32px 0 0; color: var(--sr-muted); font-size: 16px; line-height: 1.8; }
.sr-hero__actions { display: flex; gap: 10px; margin-top: 34px; }
.sr-button { display: inline-flex; min-height: 46px; align-items: center; justify-content: center; gap: 22px; padding: 0 19px; border-radius: 10px; font-size: 12px; font-weight: 650; text-decoration: none; transition: transform .2s ease, border-color .2s ease, background .2s ease; }
.sr-button:hover { transform: translateY(-1px); }
.sr-button--primary { border: 1px solid rgba(236,249,255,.56); background: rgba(239,249,255,.92); color: #10212c; box-shadow: 0 14px 32px rgba(1,10,17,.14); }
.sr-button--ghost { border: 1px solid rgba(255,255,255,.14); background: rgba(7,18,27,.18); color: #fff; }
.sr-endpoint { display: flex; width: min(100%, 520px); align-items: center; gap: 13px; margin-top: 46px; padding-top: 19px; border-top: 1px solid rgba(255,255,255,.09); }
.sr-endpoint > span:first-child { color: var(--sr-soft); font-size: 8px; font-weight: 700; letter-spacing: .18em; }
.sr-endpoint code { flex: 1; overflow: hidden; color: rgba(246,251,255,.7); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.sr-endpoint__dot { width: 5px; height: 5px; border-radius: 50%; background: var(--sr-accent); }

.sr-gateway-stage { position: relative; min-height: 520px; }
.sr-gateway-stage::before { content: ''; position: absolute; inset: 10% -14% -8% 8%; border-radius: 50%; background: radial-gradient(circle, rgba(197,234,252,.13), transparent 68%); filter: blur(22px); }
.sr-gateway-card { position: relative; z-index: 2; overflow: hidden; margin-top: 18px; border: 1px solid rgba(255,255,255,.14); border-radius: 22px; background: linear-gradient(145deg, rgba(13,31,43,.58), rgba(4,14,22,.50)); box-shadow: 0 32px 90px rgba(2,9,15,.24), inset 0 1px rgba(255,255,255,.05); -webkit-backdrop-filter: blur(24px) saturate(118%); backdrop-filter: blur(24px) saturate(118%); }
.sr-gateway-card::before { content: ''; position: absolute; inset: 0; pointer-events: none; background: linear-gradient(120deg, rgba(255,255,255,.055), transparent 24% 70%, rgba(166,216,243,.035)); }
.sr-gateway-card__head { display: flex; min-height: 86px; align-items: center; justify-content: space-between; padding: 0 24px; border-bottom: 1px solid rgba(255,255,255,.08); }
.sr-gateway-card__head > div { display: flex; flex-direction: column; gap: 4px; }
.sr-gateway-card__head strong { font-size: 15px; font-weight: 590; letter-spacing: -.02em; }
.sr-gateway-card__label { color: var(--sr-soft); font-size: 8px; letter-spacing: .18em; }
.sr-live { display: flex; align-items: center; gap: 7px; padding: 6px 9px; border: 1px solid rgba(143,225,185,.18); border-radius: 99px; color: rgba(186,239,211,.82); background: rgba(93,177,135,.07); font-size: 8px; font-weight: 700; letter-spacing: .12em; }
.sr-route-map { position: relative; display: grid; grid-template-columns: 1fr 46px 86px 1fr; align-items: center; min-height: 292px; padding: 34px 24px; }
.sr-route-map__origin { display: flex; min-width: 0; flex-direction: column; gap: 7px; padding: 17px; border: 1px solid rgba(255,255,255,.09); border-radius: 12px; background: rgba(0,8,13,.19); }
.sr-route-map__origin span, .sr-route-map__origin small { color: var(--sr-soft); font-size: 7px; letter-spacing: .15em; }
.sr-route-map__origin strong { overflow: hidden; color: rgba(250,253,255,.8); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 9px; font-weight: 500; text-overflow: ellipsis; white-space: nowrap; }
.sr-route-map__spine { position: relative; height: 1px; background: rgba(190,231,252,.26); }
.sr-route-map__spine i { position: absolute; top: -2px; width: 5px; height: 5px; border-radius: 50%; background: var(--sr-accent); box-shadow: 0 0 9px rgba(185,232,255,.6); animation: sr-pulse 2.6s infinite ease-in-out; }
.sr-route-map__spine i:nth-child(1) { left: 12%; }
.sr-route-map__spine i:nth-child(2) { left: 46%; animation-delay: .45s; }
.sr-route-map__spine i:nth-child(3) { left: 82%; animation-delay: .9s; }
.sr-route-map__core { position: relative; display: grid; width: 70px; height: 70px; place-items: center; justify-self: center; border: 1px solid rgba(211,240,255,.18); border-radius: 50%; background: rgba(13,31,43,.48); box-shadow: 0 0 34px rgba(159,218,247,.08), inset 0 0 24px rgba(172,224,251,.05); }
.sr-route-map__core strong { position: relative; z-index: 2; font-size: 11px; letter-spacing: .08em; }
.sr-core-dot { position: absolute; width: 5px; height: 5px; border-radius: 50%; background: var(--sr-accent); box-shadow: 0 0 14px rgba(185,232,255,.9); transform: translate(23px,-20px); }
.sr-core-orbit { position: absolute; border: 1px solid rgba(203,237,255,.11); border-radius: 50%; }
.sr-core-orbit--one { inset: -8px; }
.sr-core-orbit--two { inset: -17px; border-color: rgba(203,237,255,.055); }
.sr-route-map__branches { display: flex; flex-direction: column; gap: 8px; margin-left: 6px; }
.sr-provider { position: relative; display: grid; grid-template-columns: 8px 1fr auto; align-items: center; gap: 8px; min-height: 37px; padding: 0 10px; border: 1px solid rgba(255,255,255,.07); border-radius: 9px; background: rgba(2,10,17,.20); }
.sr-provider::before { content: ''; position: absolute; top: 50%; right: 100%; width: 18px; height: 1px; background: rgba(190,231,252,.16); }
.sr-provider i { width: 5px; height: 5px; border-radius: 50%; background: rgba(230,243,251,.28); }
.sr-provider span { color: rgba(248,252,255,.74); font-size: 9px; }
.sr-provider small { color: var(--sr-soft); font-size: 7px; }
.sr-provider--active { border-color: rgba(185,232,255,.16); background: rgba(115,179,213,.07); }
.sr-provider--active i { background: var(--sr-accent); box-shadow: 0 0 9px rgba(185,232,255,.6); }
.sr-gateway-card__foot { display: grid; grid-template-columns: repeat(3,1fr); min-height: 76px; border-top: 1px solid rgba(255,255,255,.08); }
.sr-gateway-card__foot > div { display: flex; flex-direction: column; justify-content: center; gap: 5px; padding: 0 17px; border-right: 1px solid rgba(255,255,255,.07); }
.sr-gateway-card__foot > div:last-child { border-right: 0; }
.sr-gateway-card__foot span { color: var(--sr-soft); font-size: 7px; letter-spacing: .1em; text-transform: uppercase; }
.sr-gateway-card__foot strong { color: rgba(248,252,255,.72); font-size: 9px; font-weight: 550; }
.sr-stage-note { position: absolute; z-index: 3; color: rgba(239,248,253,.30); font-size: 7px; letter-spacing: .2em; }
.sr-stage-note--top { top: 0; right: 2px; }
.sr-stage-note--bottom { bottom: 12px; left: 14px; }

.sr-proof { display: grid; grid-template-columns: 1.25fr repeat(3,1fr); width: min(1440px, calc(100% - 64px)); margin: 0 auto; border-bottom: 1px solid rgba(255,255,255,.09); }
.sr-proof > div { min-height: 150px; padding: 31px 28px; border-right: 1px solid rgba(255,255,255,.08); }
.sr-proof > div:first-child { padding-left: 0; }
.sr-proof > div:last-child { border-right: 0; }
.sr-proof__intro { display: flex; flex-direction: column; justify-content: space-between; }
.sr-proof__intro span, .sr-section-heading > span, .sr-quickstart__copy > span { color: var(--sr-soft); font-size: 8px; font-weight: 650; letter-spacing: .19em; }
.sr-proof__intro strong { max-width: 270px; font-size: 16px; font-weight: 550; line-height: 1.45; }
.sr-proof__item strong { display: block; margin-bottom: 25px; color: rgba(220,241,252,.34); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 9px; }
.sr-proof__item span { display: block; margin-bottom: 7px; font-size: 12px; font-weight: 620; }
.sr-proof__item p { margin: 0; color: var(--sr-muted); font-size: 10px; line-height: 1.6; }

.sr-workspace { width: min(1440px, calc(100% - 64px)); margin: 0 auto; padding: 120px 0 128px; }
.sr-section-heading { max-width: 640px; margin-bottom: 46px; }
.sr-section-heading h2, .sr-quickstart__copy h2 { margin: 16px 0 0; font-size: clamp(34px,4vw,54px); font-weight: 520; letter-spacing: -.05em; line-height: 1.08; }
.sr-section-heading p, .sr-quickstart__copy p { max-width: 560px; margin: 20px 0 0; color: var(--sr-muted); font-size: 13px; line-height: 1.75; }
.sr-workspace-grid { display: grid; grid-template-columns: 1.18fr .82fr; gap: 12px; }
.sr-workspace-card { position: relative; min-height: 280px; overflow: hidden; padding: 27px; border: 1px solid rgba(255,255,255,.10); border-radius: 16px; background: rgba(6,18,28,.28); box-shadow: inset 0 1px rgba(255,255,255,.025); -webkit-backdrop-filter: blur(14px); backdrop-filter: blur(14px); }
.sr-workspace-card--wide { grid-row: span 2; min-height: 572px; }
.sr-card-index { color: rgba(225,243,252,.30); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 8px; }
.sr-card-copy { position: absolute; left: 27px; bottom: 27px; max-width: 360px; }
.sr-card-copy > span { color: var(--sr-soft); font-size: 7px; font-weight: 700; letter-spacing: .18em; }
.sr-card-copy h3 { margin: 8px 0 7px; font-size: 23px; font-weight: 550; letter-spacing: -.035em; }
.sr-card-copy p { margin: 0; color: var(--sr-muted); font-size: 11px; line-height: 1.6; }
.sr-key-visual { position: absolute; top: 90px; left: 50%; display: flex; width: min(74%, 520px); height: 68px; align-items: center; gap: 7px; padding: 0 19px; border: 1px solid rgba(255,255,255,.14); border-radius: 13px; background: rgba(3,12,19,.28); box-shadow: 0 24px 70px rgba(2,9,14,.15); transform: translateX(-50%) rotate(-3deg); }
.sr-key-visual span { color: rgba(243,250,254,.43); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 9px; }
.sr-key-visual i { width: 4px; height: 4px; border-radius: 50%; background: rgba(238,248,253,.25); }
.sr-key-visual strong { margin-left: auto; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; font-weight: 500; letter-spacing: .08em; }
.sr-model-stack { position: absolute; top: 38px; right: 24px; display: flex; gap: 6px; }
.sr-model-stack span { display: grid; min-width: 46px; height: 30px; place-items: center; padding: 0 9px; border: 1px solid rgba(255,255,255,.09); border-radius: 8px; background: rgba(255,255,255,.035); color: rgba(247,251,254,.56); font-size: 8px; }
.sr-bars { position: absolute; top: 46px; right: 28px; display: flex; height: 86px; align-items: flex-end; gap: 5px; }
.sr-bars i { width: 8px; height: var(--h); border-radius: 4px 4px 1px 1px; background: linear-gradient(to top, rgba(152,210,239,.17), rgba(197,234,252,.55)); }

.sr-quickstart { display: grid; grid-template-columns: .82fr 1.18fr; gap: 80px; width: min(1440px, calc(100% - 64px)); margin: 0 auto 118px; padding: 88px; border: 1px solid rgba(255,255,255,.11); border-radius: 22px; background: linear-gradient(145deg, rgba(7,20,30,.31), rgba(3,13,21,.20)); -webkit-backdrop-filter: blur(18px); backdrop-filter: blur(18px); }
.sr-quickstart__copy { align-self: center; }
.sr-quickstart__copy a { display: inline-flex; gap: 18px; margin-top: 27px; color: #fff; font-size: 11px; text-decoration: none; }
.sr-code-panel { overflow: hidden; border: 1px solid rgba(255,255,255,.10); border-radius: 14px; background: rgba(2,9,15,.47); box-shadow: 0 28px 70px rgba(1,8,13,.18); }
.sr-code-panel__head { display: flex; height: 46px; align-items: center; justify-content: space-between; padding: 0 17px; border-bottom: 1px solid rgba(255,255,255,.07); color: rgba(244,250,253,.5); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 8px; }
.sr-code-panel pre { overflow-x: auto; margin: 0; padding: 25px; color: rgba(245,251,254,.72); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; line-height: 1.8; }
.sr-code-panel b { color: #c8ebfc; font-weight: 500; }
.sr-code-dim { color: rgba(193,219,232,.42); }

.sr-footer { display: grid; grid-template-columns: 1fr 1.4fr 1fr; align-items: center; gap: 30px; width: min(1440px, calc(100% - 64px)); min-height: 120px; margin: 0 auto; border-top: 1px solid rgba(255,255,255,.09); color: var(--sr-soft); }
.sr-footer__brand { gap: 9px; color: rgba(249,252,255,.76); font-size: 11px; }
.sr-footer p { margin: 0; font-size: 9px; line-height: 1.6; text-align: center; }
.sr-footer__links { justify-content: flex-end; gap: 22px; }
.sr-footer__links a { color: inherit; font-size: 9px; text-decoration: none; }

@keyframes sr-pulse { 0%,100% { opacity: .22; transform: scale(.78); } 48% { opacity: 1; transform: scale(1.08); } }

@media (max-width: 1100px) {
  .sr-nav { grid-template-columns: 1fr auto; }
  .sr-nav__links { display: none; }
  .sr-hero { grid-template-columns: 1fr; padding-top: 72px; }
  .sr-gateway-stage { width: min(100%, 680px); }
  .sr-proof { grid-template-columns: repeat(3,1fr); }
  .sr-proof__intro { grid-column: 1 / -1; min-height: 100px !important; padding-left: 0 !important; border-right: 0 !important; border-bottom: 1px solid rgba(255,255,255,.08); }
  .sr-quickstart { grid-template-columns: 1fr; gap: 44px; padding: 60px; }
}

@media (max-width: 760px) {
  .sr-nav, .sr-hero, .sr-proof, .sr-workspace, .sr-quickstart, .sr-footer { width: min(100% - 32px, 1440px); }
  .sr-nav { min-height: 74px; }
  .sr-status { display: none; }
  .sr-hero { min-height: auto; gap: 54px; padding: 70px 0 80px; }
  .sr-hero h1 { font-size: clamp(46px,14vw,68px); }
  .sr-hero__lead { font-size: 14px; }
  .sr-route-map { grid-template-columns: 1fr; gap: 16px; }
  .sr-route-map__spine { width: 1px; height: 24px; justify-self: center; }
  .sr-route-map__spine i { display: none; }
  .sr-route-map__branches { margin-left: 0; }
  .sr-provider::before { display: none; }
  .sr-gateway-card__foot { grid-template-columns: 1fr; }
  .sr-gateway-card__foot > div { min-height: 52px; border-right: 0; border-bottom: 1px solid rgba(255,255,255,.06); }
  .sr-proof { grid-template-columns: 1fr; }
  .sr-proof > div { min-height: 118px; padding: 25px 0; border-right: 0; border-bottom: 1px solid rgba(255,255,255,.07); }
  .sr-workspace { padding: 88px 0; }
  .sr-workspace-grid { grid-template-columns: 1fr; }
  .sr-workspace-card--wide { grid-row: auto; min-height: 370px; }
  .sr-workspace-card { min-height: 250px; }
  .sr-key-visual { top: 78px; width: 78%; }
  .sr-quickstart { margin-bottom: 74px; padding: 38px 24px; }
  .sr-footer { grid-template-columns: 1fr; gap: 18px; padding: 32px 0; }
  .sr-footer p { text-align: left; }
  .sr-footer__links { justify-content: flex-start; }
}
</style>
