<template>
  <AppLayout v-if="isEmbedded">
    <section class="smg-page">
      <header class="smg-page-head"><div><div class="smg-page-kicker">SMIREL MODEL CATALOG</div><h1 class="smg-page-title">{{ copy.title }}</h1><p class="smg-page-description">{{ copy.description }}</p></div></header>
      <div class="smg-catalog-panel"><ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" embedded /></div>
    </section>
  </AppLayout>

  <div v-else class="smp-page">
    <header class="smp-nav">
      <router-link to="/home" class="smp-brand"><img v-if="siteLogo" :src="siteLogo" alt="Smirel" /><span>{{ siteName }}</span></router-link>
      <div class="smp-nav-actions">
        <router-link to="/home">{{ copy.home }}</router-link><router-link to="/key-usage">{{ copy.keyUsage }}</router-link>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docs }}</a><LocaleSwitcher />
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="smp-primary-action">{{ isAuthenticated ? copy.console : copy.login }}</router-link>
      </div>
    </header>
    <main class="smp-main">
      <section class="smp-hero">
        <div><div class="smp-kicker">SMIREL MODEL CATALOG</div><h1>{{ copy.title }}</h1><p>{{ copy.description }}</p></div>
        <aside class="smp-hero-aside"><span>API BASE</span><code>{{ apiBase }}</code><span class="smp-aside-gap">COMPATIBILITY</span><code>OpenAI-compatible</code></aside>
      </section>
      <div class="smp-catalog-panel"><ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" /></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import ModelPlazaContent from '@/components/modelPlaza/ModelPlazaContent.vue'
import { getModelPlaza, type ModelPlazaResponse } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { sanitizeUrl } from '@/utils/url'

const route=useRoute();const { locale }=useI18n();const appStore=useAppStore();const authStore=useAuthStore()
const isZh=computed(()=>locale.value.toLowerCase().startsWith('zh'));const isEmbedded=computed(()=>route.query.embedded==='1'&&authStore.isAuthenticated);const isAuthenticated=computed(()=>authStore.isAuthenticated);const dashboardPath=computed(()=>(authStore.isAdmin?'/admin/dashboard':'/dashboard'))
const siteName=computed(()=>appStore.cachedPublicSettings?.site_name||appStore.siteName||'Smirel API');const siteLogo=computed(()=>sanitizeUrl(appStore.cachedPublicSettings?.site_logo||appStore.siteLogo||'',{allowRelative:true,allowDataUrl:true}));const docUrl=computed(()=>sanitizeUrl(appStore.cachedPublicSettings?.doc_url||appStore.docUrl||''));const apiBase=computed(()=>appStore.cachedPublicSettings?.api_base_url||appStore.apiBaseUrl||'https://api.smirel.com/v1')
const copy=computed(()=>isZh.value?{title:'模型与价格',description:'从产品视角查看可以调用的模型、能力与价格。这里不暴露上游账号、内部路由分组或调度细节。',home:'首页',keyUsage:'Key 用量',docs:'API 文档',console:'进入控制台',login:'登录'}:{title:'Models & pricing',description:'Browse callable models, capabilities and pricing from the product perspective. Upstream accounts and internal routing stay hidden.',home:'Home',keyUsage:'Key usage',docs:'API Docs',console:'Open console',login:'Sign in'})
const data=ref<ModelPlazaResponse|null>(null);const loading=ref(true);const loadFailed=ref(false)
onMounted(async()=>{void appStore.fetchPublicSettings();try{data.value=await getModelPlaza()}catch{loadFailed.value=true}finally{loading.value=false}})
</script>
