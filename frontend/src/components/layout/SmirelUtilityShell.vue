<template>
  <div class="smu-shell">
    <header class="smp-nav">
      <router-link to="/home" class="smp-brand"><img v-if="siteLogo" :src="siteLogo" alt="Smirel" /><span>{{ siteName }}</span></router-link>
      <div class="smp-nav-actions"><router-link to="/model-plaza">{{ copy.models }}</router-link><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a><LocaleSwitcher /><router-link :to="isAuthenticated ? dashboardPath : '/login'" class="smp-primary-action">{{ isAuthenticated ? copy.console : copy.login }}</router-link></div>
    </header>
    <div class="smu-context"><div><div class="smp-kicker">SMIREL API</div><strong>{{ contextTitle }}</strong></div><span>{{ copy.context }}</span></div>
    <div class="smu-body"><slot /></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const route=useRoute();const { locale,t }=useI18n();const appStore=useAppStore();const authStore=useAuthStore();const isZh=computed(()=>locale.value.toLowerCase().startsWith('zh'));const isAuthenticated=computed(()=>authStore.isAuthenticated);const dashboardPath=computed(()=>(authStore.isAdmin?'/admin/dashboard':'/dashboard'));const siteName=computed(()=>appStore.cachedPublicSettings?.site_name||appStore.siteName||'Smirel API');const siteLogo=computed(()=>sanitizeUrl(appStore.cachedPublicSettings?.site_logo||appStore.siteLogo||'',{allowRelative:true,allowDataUrl:true}));const docUrl=computed(()=>sanitizeUrl(appStore.cachedPublicSettings?.doc_url||appStore.docUrl||''))
const copy=computed(()=>isZh.value?{models:'模型与价格',docs:'API 文档',console:'进入控制台',login:'登录',context:'独立的 Smirel 服务页面'}:{models:'Models & Pricing',docs:'API Docs',console:'Open console',login:'Sign in',context:'A first-party Smirel service surface'})
const contextTitle=computed(()=>{const key=route.meta.titleKey;if(typeof key==='string'&&key){const translated=t(key);if(translated!==key)return translated}return String(route.meta.title||(isZh.value?'服务页面':'Service'))})
</script>
