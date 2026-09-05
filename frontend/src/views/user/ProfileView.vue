<template>
  <AppLayout>
    <section class="smg-page" data-testid="profile-shell">
      <header class="smg-page-head">
        <div>
          <div class="smg-page-kicker">SMIREL ACCOUNT</div>
          <h1 class="smg-page-title">{{ copy.title }}</h1>
          <p class="smg-page-description">{{ copy.description }}</p>
        </div>
      </header>

      <div class="smg-profile-grid">
        <div class="smg-profile-column">
          <div class="smg-profile-section-label">{{ copy.identity }}</div>
          <ProfileInfoCard
            :user="user"
            :linuxdo-enabled="linuxdoOAuthEnabled"
            :dingtalk-enabled="dingtalkOAuthEnabled"
            :oidc-enabled="oidcOAuthEnabled"
            :oidc-provider-name="oidcOAuthProviderName"
            :wechat-enabled="wechatOAuthEnabled"
            :wechat-open-enabled="wechatOAuthOpenEnabled"
            :wechat-mp-enabled="wechatOAuthMPEnabled"
          />

          <div class="smg-profile-section-label">{{ copy.password }}</div>
          <ProfilePasswordForm />

          <ProfileBalanceNotifyCard
            v-if="user && balanceLowNotifyEnabled"
            :enabled="user.balance_notify_enabled ?? true"
            :threshold="user.balance_notify_threshold"
            :extra-emails="user.balance_notify_extra_emails ?? []"
            :system-default-threshold="systemDefaultThreshold"
            :user-email="user.email"
          />
        </div>

        <aside class="smg-profile-column">
          <div v-if="contactInfo" class="smg-profile-support">
            <div class="smg-profile-support-label">{{ copy.support }}</div>
            <strong>{{ contactInfo }}</strong>
            <p>{{ copy.supportHint }}</p>
          </div>

          <div class="smg-profile-section-label">{{ copy.security }}</div>
          <ProfileTotpCard />
          <ProfilePasskeyCard :enabled="passkeyEnabled" />
        </aside>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import ProfilePasskeyCard from '@/components/user/profile/ProfilePasskeyCard.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))

const copy = computed(() =>
  isZh.value
    ? {
        title: '账户与安全',
        description: '管理个人资料、登录方式、余额提醒和账户安全。API 产品设置与个人身份信息分开维护。',
        identity: 'IDENTITY & CONNECTIONS',
        password: 'PASSWORD & NOTIFICATIONS',
        security: 'SECURITY',
        support: 'SUPPORT',
        supportHint: '遇到账号、计费或调用问题时，可通过这里的官方联系方式获取支持。',
      }
    : {
        title: 'Account & security',
        description: 'Manage profile, sign-in methods, balance alerts and account security. API product settings stay separate from your identity.',
        identity: 'IDENTITY & CONNECTIONS',
        password: 'PASSWORD & NOTIFICATIONS',
        security: 'SECURITY',
        support: 'SUPPORT',
        supportHint: 'Use this official contact for account, billing or API support.',
      },
)

const contactInfo = ref('')
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const dingtalkOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')
const passkeyEnabled = ref(false)

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) return
      contactInfo.value = settings.contact_info || ''
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
      passkeyEnabled.value = settings.passkey_enabled === true
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>
