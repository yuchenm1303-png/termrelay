<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { acceptAdminCompliance, getAdminComplianceStatus, type AdminComplianceStatus } from '../core/adminCompliance'
import { getErrorMessage, sanitizeOAuthRedirect } from '../core/api'
import { interfacePreferences, setLocale } from '../core/preferences'
import { useSession } from '../core/session'

const route = useRoute()
const router = useRouter()
const { state, logout } = useSession()

const status = ref<AdminComplianceStatus | null>(null)
const loading = ref(true)
const submitting = ref(false)
const loadError = ref('')
const submitError = ref('')
const acceptedReading = ref(false)
const typedPhrase = ref('')
const copied = ref(false)

const isZh = computed(() => interfacePreferences.locale === 'zh-CN')
const language = computed<'zh' | 'en'>(() => isZh.value ? 'zh' : 'en')
const expectedPhrase = computed(() => isZh.value ? status.value?.ack_phrase_zh || '' : status.value?.ack_phrase_en || '')
const documentUrl = computed(() => isZh.value ? status.value?.document_url_zh || '' : status.value?.document_url_en || '')
const redirectTarget = computed(() => sanitizeOAuthRedirect(route.query.redirect, '/admin/dashboard'))
const canSubmit = computed(() => Boolean(
  status.value?.required
  && acceptedReading.value
  && typedPhrase.value.trim() === expectedPhrase.value
  && !submitting.value,
))

const copy = computed(() => isZh.value ? {
  eyebrow: 'ADMINISTRATOR ACKNOWLEDGEMENT',
  title: '管理员部署与运营确认',
  subtitle: '在进入管理后台前，请确认当前实例的运营责任与安全要求。每个管理员仅需对当前版本确认一次；条款更新后会再次提示。',
  whyTitle: '为什么需要这一步',
  whyText: '这不是再次验证管理员身份，而是记录实际运营管理员已经阅读当前版本的运营要求。确认记录按管理员账号保存。',
  scopeTitle: '你需要确认的核心事项',
  points: [
    '遵守实际接入的上游服务条款、地区限制、商业使用与转售规则。',
    '对用户、请求内容、日志和账户信息采取合理的数据与隐私保护措施。',
    '收费、退款、对账、税务和第三方分成按照适用规则清晰处理并留痕。',
    '保护管理员账户、服务器、数据库、API Key、OAuth 与支付凭据等敏感信息。',
    '建立限流、风控、审计、备份恢复、滥用处理和安全事件响应机制。',
    '不利用当前实例实施或协助明显违法、欺诈、侵权、盗号或恶意绕过安全控制的行为。',
  ],
  recordTitle: '确认后会记录什么',
  recordText: '系统会记录条款版本、管理员账户、确认时间、IP 地址和 User-Agent，用于证明当前管理员已确认该版本。',
  fullTerms: '阅读完整条款',
  checkbox: '我已阅读并理解完整条款及上述运营要求',
  phraseLabel: '请输入确认短语',
  phraseHint: '为避免误操作，请完整输入下方短语后再确认。',
  copyPhrase: copied.value ? '已复制' : '复制短语',
  submit: submitting.value ? '正在确认…' : '确认并进入管理后台',
  loading: '正在检查确认状态…',
  retry: '重新检查',
  logout: '退出管理员账户',
  account: '当前管理员',
  version: '条款版本',
  failedTitle: '暂时无法检查确认状态',
  accepted: '当前版本已确认，正在进入管理后台…',
  language: 'English',
} : {
  eyebrow: 'ADMINISTRATOR ACKNOWLEDGEMENT',
  title: 'Administrator deployment & operation acknowledgement',
  subtitle: 'Before entering the administration console, confirm the operational responsibilities and security requirements for this instance. Each administrator acknowledges the current version once; an updated version will require acknowledgement again.',
  whyTitle: 'Why this is required',
  whyText: 'This is not another identity check. It records that the actual administrator has reviewed the current operational requirements. The record is stored per administrator account.',
  scopeTitle: 'Core points you are acknowledging',
  points: [
    'Follow the terms, regional restrictions, commercial-use rules, and resale rules of upstream providers actually in use.',
    'Apply reasonable privacy and data safeguards to users, request content, logs, and account information.',
    'Handle charging, refunds, reconciliation, taxes, and third-party revenue sharing under applicable rules with clear records.',
    'Protect administrator accounts, servers, databases, API keys, OAuth credentials, payment secrets, and other sensitive credentials.',
    'Maintain rate limits, risk controls, audit trails, backups, recovery, abuse handling, and security-incident response.',
    'Do not knowingly use the instance to facilitate clearly unlawful, fraudulent, infringing, credential-theft, or malicious security-bypass activity.',
  ],
  recordTitle: 'What is recorded',
  recordText: 'The system records the document version, administrator account, acknowledgement time, IP address, and User-Agent as evidence of acknowledgement for this version.',
  fullTerms: 'Read full terms',
  checkbox: 'I have read and understood the full terms and the operational requirements above',
  phraseLabel: 'Type the confirmation phrase',
  phraseHint: 'To prevent accidental acknowledgement, type the phrase below exactly.',
  copyPhrase: copied.value ? 'Copied' : 'Copy phrase',
  submit: submitting.value ? 'Acknowledging…' : 'Acknowledge and enter admin console',
  loading: 'Checking acknowledgement status…',
  retry: 'Check again',
  logout: 'Sign out administrator',
  account: 'Current administrator',
  version: 'Terms version',
  failedTitle: 'Unable to check acknowledgement status',
  accepted: 'Current version already acknowledged. Opening the admin console…',
  language: '中文',
})

async function loadStatus() {
  loading.value = true
  loadError.value = ''
  try {
    status.value = await getAdminComplianceStatus(true)
    if (!status.value.required) {
      window.setTimeout(() => void router.replace(redirectTarget.value), 250)
    }
  } catch (error) {
    loadError.value = getErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function copyPhrase() {
  if (!expectedPhrase.value) return
  try {
    await navigator.clipboard.writeText(expectedPhrase.value)
    copied.value = true
    window.setTimeout(() => { copied.value = false }, 1800)
  } catch {
    // The phrase remains visible so manual copy is always available.
  }
}

async function submitAcknowledgement() {
  if (!canSubmit.value) return
  submitting.value = true
  submitError.value = ''
  try {
    const next = await acceptAdminCompliance(typedPhrase.value.trim(), language.value)
    status.value = next
    if (!next.required) await router.replace(redirectTarget.value)
  } catch (error) {
    submitError.value = getErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function signOut() {
  await logout()
  await router.replace('/login')
}

function toggleLanguage() {
  setLocale(isZh.value ? 'en-US' : 'zh-CN')
  typedPhrase.value = ''
  acceptedReading.value = false
  submitError.value = ''
}

onMounted(() => { void loadStatus() })
</script>

<template>
  <div class="compliance-page">
    <div class="compliance-environment" aria-hidden="true"></div>

    <header class="compliance-topbar">
      <RouterLink to="/home" class="compliance-brand" aria-label="Muxway home">
        <img src="/smirel-logo.png" alt="" />
        <span><strong>Muxway</strong><small>ADMINISTRATION</small></span>
      </RouterLink>
      <div class="compliance-topbar-actions">
        <button type="button" class="ghost-button" @click="toggleLanguage">{{ copy.language }}</button>
        <button type="button" class="ghost-button" @click="signOut">{{ copy.logout }}</button>
      </div>
    </header>

    <main class="compliance-shell">
      <section v-if="loading" class="compliance-state-card">
        <span class="spinner" aria-hidden="true"></span>
        <strong>{{ copy.loading }}</strong>
      </section>

      <section v-else-if="loadError" class="compliance-state-card error-state">
        <span class="state-mark">!</span>
        <h1>{{ copy.failedTitle }}</h1>
        <p>{{ loadError }}</p>
        <button type="button" class="primary-button" @click="loadStatus">{{ copy.retry }}</button>
      </section>

      <section v-else-if="status && !status.required" class="compliance-state-card">
        <span class="state-mark success">✓</span>
        <strong>{{ copy.accepted }}</strong>
      </section>

      <template v-else-if="status">
        <section class="compliance-hero">
          <div>
            <span class="eyebrow">{{ copy.eyebrow }}</span>
            <h1>{{ copy.title }}</h1>
            <p>{{ copy.subtitle }}</p>
          </div>
          <dl class="compliance-meta">
            <div>
              <dt>{{ copy.account }}</dt>
              <dd>{{ state.user?.email || '—' }}</dd>
            </div>
            <div>
              <dt>{{ copy.version }}</dt>
              <dd>{{ status.version }}</dd>
            </div>
          </dl>
        </section>

        <section class="compliance-grid">
          <article class="compliance-panel context-panel">
            <span class="section-index">01</span>
            <h2>{{ copy.whyTitle }}</h2>
            <p>{{ copy.whyText }}</p>
          </article>

          <article class="compliance-panel requirements-panel">
            <header>
              <div>
                <span class="section-index">02</span>
                <h2>{{ copy.scopeTitle }}</h2>
              </div>
              <a v-if="documentUrl" :href="documentUrl" target="_blank" rel="noopener noreferrer">{{ copy.fullTerms }} ↗</a>
            </header>
            <ol>
              <li v-for="(point, index) in copy.points" :key="point">
                <span>{{ String(index + 1).padStart(2, '0') }}</span>
                <p>{{ point }}</p>
              </li>
            </ol>
          </article>

          <article class="compliance-panel record-panel">
            <span class="section-index">03</span>
            <h2>{{ copy.recordTitle }}</h2>
            <p>{{ copy.recordText }}</p>
          </article>
        </section>

        <section class="compliance-confirm-card">
          <label class="reading-check">
            <input v-model="acceptedReading" type="checkbox" />
            <span aria-hidden="true"></span>
            <strong>{{ copy.checkbox }}</strong>
          </label>

          <div class="phrase-block">
            <div class="phrase-heading">
              <div>
                <span>{{ copy.phraseLabel }}</span>
                <small>{{ copy.phraseHint }}</small>
              </div>
              <button type="button" class="text-button" @click="copyPhrase">{{ copy.copyPhrase }}</button>
            </div>
            <code>{{ expectedPhrase }}</code>
            <input
              v-model="typedPhrase"
              type="text"
              autocomplete="off"
              spellcheck="false"
              :placeholder="expectedPhrase"
              @keyup.enter="submitAcknowledgement"
            />
          </div>

          <p v-if="submitError" class="submit-error">{{ submitError }}</p>

          <div class="confirm-footer">
            <span>{{ status.version }}</span>
            <button type="button" class="primary-button" :disabled="!canSubmit" @click="submitAcknowledgement">
              {{ copy.submit }}
            </button>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<style scoped>
.compliance-page {
  position: relative;
  min-height: 100vh;
  color: #e9edf2;
  background: #07090c;
  overflow-x: hidden;
}

.compliance-environment {
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 18% 8%, rgba(69, 141, 190, 0.11), transparent 31%),
    radial-gradient(circle at 84% 22%, rgba(77, 126, 166, 0.07), transparent 24%),
    linear-gradient(180deg, rgba(255,255,255,0.015), transparent 28%);
}

.compliance-topbar {
  position: relative;
  z-index: 2;
  min-height: 72px;
  padding: 0 36px;
  border-bottom: 1px solid #1b2027;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(8, 10, 14, 0.88);
  backdrop-filter: blur(16px);
}

.compliance-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: inherit;
  text-decoration: none;
}

.compliance-brand img { width: 34px; height: 34px; object-fit: contain; }
.compliance-brand span { display: grid; gap: 2px; }
.compliance-brand strong { font-size: .9rem; letter-spacing: .01em; }
.compliance-brand small { color: #66717e; font-size: .56rem; font-weight: 700; letter-spacing: .13em; }

.compliance-topbar-actions { display: flex; gap: 8px; }
.ghost-button,
.text-button,
.primary-button {
  font: inherit;
  cursor: pointer;
}
.ghost-button {
  min-height: 36px;
  padding: 0 12px;
  border: 1px solid #252b33;
  border-radius: 8px;
  color: #aab3bd;
  background: #0d1014;
}
.ghost-button:hover { border-color: #39424d; color: #eef2f6; }

.compliance-shell {
  position: relative;
  z-index: 1;
  width: min(1120px, calc(100% - 40px));
  margin: 0 auto;
  padding: 70px 0 88px;
}

.compliance-hero {
  padding-bottom: 34px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 330px;
  gap: 48px;
  align-items: end;
}

.eyebrow,
.section-index {
  color: #6faed8;
  font-size: .66rem;
  font-weight: 760;
  letter-spacing: .14em;
}

.compliance-hero h1 {
  max-width: 760px;
  margin: 13px 0 13px;
  font-size: clamp(2rem, 4vw, 3.5rem);
  line-height: 1.03;
  letter-spacing: -.045em;
}
.compliance-hero p {
  max-width: 760px;
  margin: 0;
  color: #89939e;
  font-size: .92rem;
  line-height: 1.75;
}

.compliance-meta {
  margin: 0;
  border: 1px solid #20262d;
  border-radius: 12px;
  overflow: hidden;
  background: #0c0f13;
}
.compliance-meta div { padding: 15px 17px; }
.compliance-meta div + div { border-top: 1px solid #20262d; }
.compliance-meta dt { margin-bottom: 5px; color: #697481; font-size: .66rem; }
.compliance-meta dd { margin: 0; color: #dbe1e7; font-size: .8rem; word-break: break-all; }

.compliance-grid {
  display: grid;
  grid-template-columns: 1fr 1.8fr;
  gap: 14px;
}
.compliance-panel,
.compliance-confirm-card,
.compliance-state-card {
  border: 1px solid #20262e;
  border-radius: 14px;
  background: linear-gradient(180deg, #0e1116, #0b0e12);
  box-shadow: inset 0 1px rgba(255,255,255,.025), 0 18px 50px rgba(0,0,0,.18);
}
.compliance-panel { padding: 26px; }
.compliance-panel h2 { margin: 10px 0 10px; font-size: 1rem; }
.compliance-panel > p,
.record-panel p,
.context-panel p { margin: 0; color: #858f9b; font-size: .8rem; line-height: 1.7; }
.context-panel { min-height: 176px; }
.record-panel { min-height: 176px; }

.requirements-panel {
  grid-row: span 2;
  padding: 26px 28px;
}
.requirements-panel header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-start;
}
.requirements-panel header h2 { margin-bottom: 0; }
.requirements-panel header a {
  flex: 0 0 auto;
  color: #86bfe7;
  font-size: .72rem;
  text-decoration: none;
}
.requirements-panel ol {
  margin: 22px 0 0;
  padding: 0;
  list-style: none;
  border-top: 1px solid #222830;
}
.requirements-panel li {
  min-height: 72px;
  padding: 14px 0;
  border-bottom: 1px solid #222830;
  display: grid;
  grid-template-columns: 34px 1fr;
  gap: 14px;
  align-items: start;
}
.requirements-panel li > span {
  color: #53606e;
  font-size: .65rem;
  font-weight: 750;
  letter-spacing: .08em;
}
.requirements-panel li p { margin: 0; color: #a8b0b9; font-size: .78rem; line-height: 1.65; }

.compliance-confirm-card {
  margin-top: 16px;
  padding: 28px;
}
.reading-check {
  display: flex;
  align-items: center;
  gap: 11px;
  cursor: pointer;
}
.reading-check input { position: absolute; opacity: 0; pointer-events: none; }
.reading-check > span {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  border: 1px solid #3a4652;
  border-radius: 5px;
  background: #090c10;
}
.reading-check input:checked + span {
  border-color: #4c9bd0;
  background: #2f83bc;
  box-shadow: inset 0 0 0 4px #2f83bc;
}
.reading-check input:checked + span::after {
  content: '✓';
  display: block;
  color: white;
  text-align: center;
  font-size: 12px;
  line-height: 16px;
}
.reading-check strong { font-size: .82rem; font-weight: 620; }

.phrase-block { margin-top: 24px; }
.phrase-heading {
  margin-bottom: 10px;
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: end;
}
.phrase-heading > div { display: grid; gap: 4px; }
.phrase-heading span { color: #d6dce2; font-size: .76rem; font-weight: 650; }
.phrase-heading small { color: #6f7a86; font-size: .68rem; }
.text-button { padding: 0; border: 0; color: #80b8df; background: transparent; font-size: .7rem; }
.phrase-block code {
  min-height: 42px;
  padding: 10px 12px;
  border: 1px solid #222a33;
  border-radius: 8px 8px 0 0;
  display: flex;
  align-items: center;
  color: #9eabb8;
  background: #090c10;
  font-size: .72rem;
  white-space: normal;
}
.phrase-block input {
  width: 100%;
  box-sizing: border-box;
  min-height: 48px;
  padding: 0 13px;
  border: 1px solid #303945;
  border-top: 0;
  border-radius: 0 0 8px 8px;
  outline: none;
  color: #edf2f7;
  background: #0c1015;
  font: inherit;
  font-size: .78rem;
}
.phrase-block input:focus { border-color: #3d83b3; box-shadow: 0 0 0 3px rgba(64, 151, 209, .08); }

.submit-error { margin: 13px 0 0; color: #df8585; font-size: .74rem; }
.confirm-footer {
  margin-top: 22px;
  padding-top: 18px;
  border-top: 1px solid #20262d;
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: center;
}
.confirm-footer > span { color: #616d79; font-size: .66rem; }
.primary-button {
  min-height: 44px;
  padding: 0 18px;
  border: 1px solid #3a96d1;
  border-radius: 9px;
  color: #f8fbfd;
  background: #2f8dcc;
  box-shadow: 0 7px 22px rgba(24, 104, 157, .22);
  font-size: .78rem;
  font-weight: 680;
}
.primary-button:hover:not(:disabled) { background: #3898d7; }
.primary-button:disabled { border-color: #29313a; color: #65707b; background: #15191e; box-shadow: none; cursor: not-allowed; }

.compliance-state-card {
  min-height: 300px;
  padding: 44px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 16px;
  text-align: center;
}
.compliance-state-card h1 { margin: 0; font-size: 1.25rem; }
.compliance-state-card p { max-width: 540px; margin: 0; color: #87919c; line-height: 1.65; }
.spinner {
  width: 26px;
  height: 26px;
  border: 2px solid #27313a;
  border-top-color: #6baedb;
  border-radius: 50%;
  animation: compliance-spin .8s linear infinite;
}
.state-mark {
  width: 38px;
  height: 38px;
  border: 1px solid #713939;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #df8f8f;
  background: #3d17171f;
}
.state-mark.success { border-color: #355c49; color: #76c89b; background: #16372524; }
@keyframes compliance-spin { to { transform: rotate(360deg); } }

@media (max-width: 820px) {
  .compliance-topbar { padding: 0 18px; }
  .compliance-brand span { display: none; }
  .compliance-shell { width: min(100% - 24px, 720px); padding-top: 42px; }
  .compliance-hero { grid-template-columns: 1fr; gap: 22px; }
  .compliance-grid { grid-template-columns: 1fr; }
  .requirements-panel { grid-row: auto; }
}

@media (max-width: 560px) {
  .compliance-topbar { min-height: 62px; }
  .compliance-topbar-actions .ghost-button:last-child { display: none; }
  .compliance-shell { padding: 34px 0 54px; }
  .compliance-hero h1 { font-size: 2.15rem; }
  .compliance-panel,
  .compliance-confirm-card { padding: 21px; border-radius: 12px; }
  .requirements-panel header { display: grid; }
  .phrase-heading { align-items: start; }
  .confirm-footer { align-items: stretch; flex-direction: column; }
  .primary-button { width: 100%; }
}

@media (prefers-reduced-motion: reduce) {
  .spinner { animation: none; }
}
</style>
