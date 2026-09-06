<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useSession } from '../core/session'

const route = useRoute()
const { isAuthenticated, isAdmin } = useSession()
const kind = computed(() => String(route.meta.publicKind || 'public'))
const logoUrl = `${import.meta.env.BASE_URL}smirel-logo.png`
const consolePath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const copy = computed(() => {
  if (kind.value === 'models') return { eyebrow: 'MODEL CATALOG', title: '模型与价格', text: 'Smirel 使用统一 Base URL 接入多家模型服务。模型目录与价格将在这个独立页面持续完善。' }
  if (kind.value === 'key-usage') return { eyebrow: 'USAGE LOOKUP', title: '用量查询', text: '登录后可在工作区查看请求、Token 与实际费用。公开 Key 查询能力会在 Smirel 原生页面中提供。' }
  if (kind.value === 'legal') return { eyebrow: 'LEGAL', title: '服务条款与隐私', text: '法律文档属于 Smirel 公共页面体系，不再使用旧站点模板。' }
  if (kind.value === 'payment') return { eyebrow: 'PAYMENT', title: '支付服务', text: '支付流程保留后端能力，前端入口已经切换到 Smirel 原生页面。' }
  if (kind.value === 'callback') return { eyebrow: 'AUTH CALLBACK', title: '正在完成账户验证', text: '身份回调会返回 Smirel 登录流程。若页面长时间没有变化，请返回登录页。' }
  return { eyebrow: 'SMIREL', title: String(route.meta.title || 'Smirel'), text: '该公共功能已经进入 Smirel 独立页面体系。' }
})
</script>

<template>
  <div class="public-page"><div class="site-environment" aria-hidden="true"></div><header class="public-topbar glass"><RouterLink to="/home" class="brand-link"><img :src="logoUrl" alt="Smirel" /><span><strong>Smirel</strong><small>API SERVICE</small></span></RouterLink><div><RouterLink to="/home">首页</RouterLink><RouterLink v-if="isAuthenticated" :to="consolePath">控制台</RouterLink><RouterLink v-else to="/login">登录</RouterLink></div></header><main class="public-surface glass"><span class="eyebrow">{{ copy.eyebrow }}</span><h1>{{ copy.title }}</h1><p>{{ copy.text }}</p><div v-if="kind === 'models'" class="model-list"><article><strong>OpenAI</strong><span>Responses / Chat Completions</span></article><article><strong>Anthropic</strong><span>Messages API</span></article><article><strong>Google</strong><span>Gemini models</span></article></div><div class="public-actions"><RouterLink :to="isAuthenticated ? consolePath : '/register'" class="primary-button">{{ isAuthenticated ? '进入控制台' : '创建账户' }}</RouterLink><RouterLink to="/home" class="secondary-button">返回首页</RouterLink></div></main></div>
</template>
