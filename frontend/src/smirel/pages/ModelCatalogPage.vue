<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, getErrorMessage } from '../core/api'
import { interfacePreferences } from '../core/preferences'

type Pricing = { billing_mode?: string; input_price?: number | null; output_price?: number | null; cache_write_price?: number | null; cache_read_price?: number | null; per_request_price?: number | null }
type OfficialPricing = { input_price?: number | null; output_price?: number | null; cache_write_price?: number | null; cache_read_price?: number | null }
type PlazaModel = { name: string; platform: string; pricing?: Pricing | null; official_pricing?: OfficialPricing | null }
type PlazaGroup = { id:number; name:string; description?:string; platform:string; subscription_type?:string; rate_multiplier:number; user_rate_multiplier?:number; is_exclusive?:boolean; models:PlazaModel[] }
type PlazaResponse = { description?: string; groups?: PlazaGroup[] }
type Offer = { group: PlazaGroup; model: PlazaModel }
type CatalogModel = { id:string; provider:string; providerKey:string; mark:string; offers:Offer[]; bestRate:number; inputPerM:number|null; outputPerM:number|null; cacheReadPerM:number|null; cacheWritePerM:number|null; priceSource:'channel'|'official'|'none' }
type SortKey = 'recommended'|'name'|'price'|'routes'

const isZh = computed(() => interfacePreferences.locale === 'zh-CN')
const loading = ref(false)
const error = ref('')
const description = ref('')
const groups = ref<PlazaGroup[]>([])
const search = ref('')
const provider = ref('all')
const sortBy = ref<SortKey>('recommended')
const copied = ref('')

function family(model: string) {
  const id = model.toLowerCase()
  if (id.startsWith('claude')) return { key:'anthropic', name:'Anthropic', mark:'A' }
  if (id.startsWith('gemini')) return { key:'google', name:'Google', mark:'G' }
  if (id.startsWith('grok')) return { key:'xai', name:'xAI', mark:'X' }
  if (id.startsWith('deepseek')) return { key:'deepseek', name:'DeepSeek', mark:'D' }
  if (id.startsWith('qwen')) return { key:'qwen', name:'Qwen', mark:'Q' }
  if (id.startsWith('glm')) return { key:'zhipu', name:'GLM', mark:'Z' }
  if (id.startsWith('llama')) return { key:'meta', name:'Meta', mark:'M' }
  if (id.startsWith('gpt') || id.startsWith('o1') || id.startsWith('o3') || id.startsWith('o4')) return { key:'openai', name:'OpenAI', mark:'O' }
  return { key:'other', name:'Other', mark:'AI' }
}
function rate(g: PlazaGroup) { return Number(g.user_rate_multiplier ?? g.rate_multiplier ?? 1) }
function firstPrice(offers: Offer[], field: keyof Pricing): {value:number|null; source:'channel'|'official'|'none'; rate:number} {
  const sorted = [...offers].sort((a,b) => rate(a.group)-rate(b.group))
  for (const o of sorted) {
    const v = o.model.pricing?.[field]
    if (typeof v === 'number') return { value:v, source:'channel', rate:rate(o.group) }
  }
  const officialField = field as keyof OfficialPricing
  for (const o of sorted) {
    const v = o.model.official_pricing?.[officialField]
    if (typeof v === 'number') return { value:v, source:'official', rate:rate(o.group) }
  }
  return { value:null, source:'none', rate:sorted[0] ? rate(sorted[0].group) : 1 }
}

const models = computed<CatalogModel[]>(() => {
  const map = new Map<string, Offer[]>()
  for (const g of groups.value) for (const m of (g.models || [])) {
    const id = String(m.name || '').trim(); if (!id) continue
    const list = map.get(id) || []; list.push({ group:g, model:m }); map.set(id,list)
  }
  return [...map.entries()].map(([id, offers]) => {
    const f = family(id)
    const input = firstPrice(offers,'input_price'), output = firstPrice(offers,'output_price'), cr = firstPrice(offers,'cache_read_price'), cw = firstPrice(offers,'cache_write_price')
    const bestRate = Math.min(...offers.map(o => rate(o.group)))
    const source = input.source !== 'none' ? input.source : output.source
    const perM = (x:{value:number|null;rate:number}) => x.value == null ? null : x.value * 1_000_000 * x.rate
    return { id, provider:f.name, providerKey:f.key, mark:f.mark, offers, bestRate, inputPerM:perM(input), outputPerM:perM(output), cacheReadPerM:perM(cr), cacheWritePerM:perM(cw), priceSource:source }
  })
})
const providers = computed(() => {
  const m = new Map<string,{key:string;name:string;count:number}>()
  for (const x of models.value) { const old=m.get(x.providerKey); if(old) old.count++; else m.set(x.providerKey,{key:x.providerKey,name:x.provider,count:1}) }
  return [...m.values()].sort((a,b)=>b.count-a.count || a.name.localeCompare(b.name))
})
const filtered = computed(() => {
  const q=search.value.trim().toLowerCase()
  const arr=models.value.filter(m => (provider.value==='all'||m.providerKey===provider.value) && (!q || `${m.id} ${m.provider} ${m.offers.map(o=>o.group.name).join(' ')}`.toLowerCase().includes(q)))
  return [...arr].sort((a,b)=>{
    if(sortBy.value==='name') return a.id.localeCompare(b.id)
    if(sortBy.value==='routes') return b.offers.length-a.offers.length || a.id.localeCompare(b.id)
    if(sortBy.value==='price') return (a.inputPerM ?? Number.MAX_SAFE_INTEGER)-(b.inputPerM ?? Number.MAX_SAFE_INTEGER) || a.id.localeCompare(b.id)
    return b.offers.length-a.offers.length || a.bestRate-b.bestRate || a.id.localeCompare(b.id)
  })
})
const routeCount = computed(() => groups.value.reduce((n,g)=>n+(g.models?.length||0),0))

async function loadCatalog() {
  loading.value=true; error.value=''
  try { const r=await api.get<PlazaResponse>('/model-plaza'); description.value=String(r.data?.description||''); groups.value=Array.isArray(r.data?.groups)?r.data.groups:[] }
  catch(e){ groups.value=[]; error.value=getErrorMessage(e) }
  finally{ loading.value=false }
}
function money(v:number|null){ if(v==null) return '—'; if(v<0.01) return `$${v.toFixed(4)}`; if(v<1) return `$${v.toFixed(3)}`; return `$${v.toFixed(2)}` }
function protocol(platform:string){ return ({openai:'OpenAI Compatible',anthropic:'Messages API',gemini:'Gemini API',antigravity:'Antigravity',grok:'OpenAI Compatible',composite:'Composite'} as Record<string,string>)[platform]||platform }
function reset(){ search.value=''; provider.value='all'; sortBy.value='recommended' }
async function copyId(id:string){ if(!navigator.clipboard)return; await navigator.clipboard.writeText(id); copied.value=id; setTimeout(()=>{if(copied.value===id)copied.value=''},1200) }
onMounted(()=>void loadCatalog())
</script>

<template>
  <section class="market-page">
    <header class="heading">
      <div><span class="kicker">MODEL MARKETPLACE</span><div class="title-line"><h1>{{ isZh ? '模型广场' : 'Model Marketplace' }}</h1><span class="live"><i></i>{{ loading ? '同步中' : '实时目录' }}</span></div><p>{{ description || (isZh ? '模型、分组、价格与可用路由全部来自 TermRelay 实时配置。' : 'Models, groups, pricing and routes are loaded from the live TermRelay catalog.') }}</p></div>
      <div class="actions"><label class="search"><svg viewBox="0 0 20 20"><circle cx="8.5" cy="8.5" r="5.5"/><path d="m13 13 4 4"/></svg><input v-model="search" :placeholder="isZh ? '搜索模型、品牌或分组' : 'Search models, providers or groups'" /></label><label class="sort"><span>{{ isZh?'排序':'Sort' }}</span><select v-model="sortBy"><option value="recommended">{{ isZh?'推荐':'Recommended' }}</option><option value="name">{{ isZh?'名称':'Name' }}</option><option value="price">{{ isZh?'输入价格':'Input price' }}</option><option value="routes">{{ isZh?'路由数量':'Routes' }}</option></select></label></div>
    </header>

    <div v-if="error" class="error-box"><strong>{{ isZh?'目录暂不可用':'Catalog unavailable' }}</strong><span>{{ error }}</span><button @click="loadCatalog">{{ isZh?'重新加载':'Retry' }}</button></div>

    <section v-else class="filter-shell">
      <div class="provider-tabs"><button :class="{active:provider==='all'}" @click="provider='all'"><span>{{ isZh?'全部模型':'All models' }}</span><b>{{ models.length }}</b></button><button v-for="p in providers" :key="p.key" :class="{active:provider===p.key}" @click="provider=p.key"><span>{{ p.name }}</span><b>{{ p.count }}</b></button></div>
      <div class="catalog-meta"><span><i></i>{{ isZh?'服务端实时数据':'Live server data' }}</span><small>{{ isZh?'不再展示模拟可用率、延迟或虚构价格。':'No simulated availability, latency, or pricing.' }}</small></div>
    </section>

    <section class="result-head"><div><strong>{{ filtered.length }} {{ isZh?'个模型':'models' }}</strong><span>·</span><span>{{ groups.length }} {{ isZh?'个路由分组':'route groups' }}</span><span>·</span><span>{{ routeCount }} {{ isZh?'条模型路由':'model routes' }}</span></div><div><span>Base URL</span><code>https://api.smirel.com/v1</code><button :disabled="loading" @click="loadCatalog">{{ loading?'…':'↻' }}</button></div></section>

    <div v-if="filtered.length" class="grid">
      <article v-for="(m,index) in filtered" :key="m.id" class="card">
        <header class="card-head"><div class="identity"><span class="mark" :data-provider="m.providerKey">{{ m.mark }}</span><div><h2>{{ m.id }}</h2><div class="tags"><span>{{ m.provider }}</span><span>{{ m.bestRate.toFixed(2) }}×</span><span>{{ m.offers.length }} {{ isZh?'条路由':'routes' }}</span></div></div></div><span class="rank">#{{ String(index+1).padStart(2,'0') }}</span></header>
        <section class="prices"><div><span>{{ isZh?'输入':'Input' }}</span><strong>{{ money(m.inputPerM) }}<small v-if="m.inputPerM!=null"> / M tokens</small></strong></div><div><span>{{ isZh?'输出':'Output' }}</span><strong>{{ money(m.outputPerM) }}<small v-if="m.outputPerM!=null"> / M tokens</small></strong></div><div><span>{{ isZh?'缓存写入':'Cache write' }}</span><strong>{{ money(m.cacheWritePerM) }}<small v-if="m.cacheWritePerM!=null"> / M tokens</small></strong></div><div><span>{{ isZh?'缓存读取':'Cache read' }}</span><strong>{{ money(m.cacheReadPerM) }}<small v-if="m.cacheReadPerM!=null"> / M tokens</small></strong></div></section>
        <section class="access"><div class="id-row"><span>{{ isZh?'模型 ID':'Model ID' }}</span><code>{{ m.id }}</code><button @click="copyId(m.id)">{{ copied===m.id?(isZh?'已复制':'Copied'):(isZh?'复制':'Copy') }}</button></div><div class="source-row"><span>{{ isZh?'价格来源':'Price source' }}</span><strong>{{ m.priceSource==='channel'?(isZh?'渠道配置':'Channel pricing'):m.priceSource==='official'?(isZh?'官方参考价 × 分组倍率':'Official reference × group rate'):(isZh?'未配置':'Not configured') }}</strong></div></section>
        <footer class="routes"><div class="route-title"><span>{{ isZh?'可用路由':'Available routes' }}</span><strong>{{ m.offers.length }}</strong></div><div class="route-list"><div v-for="o in m.offers" :key="`${o.group.id}-${o.model.platform}`"><span><i></i><b>{{ o.group.name }}</b><small>#{{ o.group.id }} · {{ protocol(o.group.platform) }}</small></span><em>{{ rate(o.group).toFixed(2) }}×</em></div></div></footer>
      </article>
    </div>
    <section v-else-if="!loading && !error" class="empty"><strong>{{ models.length ? (isZh?'没有符合筛选条件的模型':'No matching models') : (isZh?'模型目录还没有发布任何模型':'No models have been published yet') }}</strong><span>{{ models.length ? '' : (isZh?'管理员可在「分组与模型」中同步真实上游模型并发布。':'An admin can sync and publish models from Groups & Models.') }}</span><button v-if="models.length" @click="reset">{{ isZh?'清除筛选':'Reset filters' }}</button></section>
    <div v-if="loading && !models.length" class="loading"><i v-for="n in 6" :key="n"></i></div>
  </section>
</template>

<style scoped>
.market-page{--panel:#101217;--border:#252930;--border2:#343a43;--text:#f4f6f8;--soft:#c5cbd2;--muted:#78828d;--green:#43cd98;width:100%;color:var(--text);font-size:14px}.heading{display:flex;align-items:flex-start;justify-content:space-between;gap:26px;padding:2px 0 24px}.kicker{color:#67717d;font-size:.65rem;font-weight:700;letter-spacing:.12em}.title-line{margin-top:8px;display:flex;align-items:center;gap:11px}.heading h1{margin:0;font-size:2.05rem;line-height:1;font-weight:700;letter-spacing:-.045em}.heading p{max-width:720px;margin:11px 0 0;color:#858e99;font-size:.84rem;line-height:1.55}.live{height:26px;padding:0 9px;border:1px solid rgba(67,205,152,.2);border-radius:99px;background:rgba(67,205,152,.055);color:#8bd8ba;display:inline-flex;align-items:center;gap:6px;font-size:.65rem}.live i,.catalog-meta i,.route-list i{width:6px;height:6px;border-radius:50%;background:var(--green)}.actions{display:flex;gap:7px}.search,.sort{height:42px;border:1px solid #2c3139;border-radius:9px;background:#0b0d11;display:flex;align-items:center}.search{width:min(350px,31vw);padding:0 12px;gap:9px}.search svg{width:15px;height:15px;fill:none;stroke:#6d7782;stroke-width:1.45}.search input{width:100%;border:0;outline:0;background:transparent;color:#eef1f4;font-size:.75rem}.sort{padding:0 10px;gap:8px}.sort span{color:#626c77;font-size:.64rem}.sort select{border:0;outline:0;background:transparent;color:#c5cbd2;font-size:.7rem}.error-box{min-height:260px;border:1px solid rgba(225,108,115,.25);border-radius:12px;background:rgba(225,108,115,.05);display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;text-align:center}.error-box strong{font-size:.9rem}.error-box span{color:#d3959a;font-size:.7rem}.error-box button,.empty button{margin-top:6px;height:34px;padding:0 12px;border:1px solid #3a4049;border-radius:7px;background:#15181d;color:#d0d5da}.filter-shell{border:1px solid var(--border);border-radius:12px;background:#101217;overflow:hidden}.provider-tabs{min-height:70px;padding:12px 14px;display:flex;align-items:center;gap:8px;overflow:auto}.provider-tabs button{height:44px;padding:0 13px;border:1px solid transparent;border-radius:8px;background:transparent;color:#939ca6;display:flex;align-items:center;gap:9px;white-space:nowrap;cursor:pointer}.provider-tabs button b{min-width:21px;height:21px;padding:0 6px;border-radius:99px;background:#1c2026;color:#7a848e;display:grid;place-items:center;font-size:.61rem}.provider-tabs button.active{border-color:#39424d;background:#161a20;color:#eff2f5}.provider-tabs button.active b{background:#252c34;color:#cdd3d9}.catalog-meta{min-height:48px;padding:0 16px;border-top:1px solid #23272e;display:flex;align-items:center;gap:8px;color:#8bd8ba;font-size:.65rem}.catalog-meta>span{display:flex;align-items:center;gap:7px}.catalog-meta small{color:#626c77}.result-head{min-height:66px;padding:0 2px;display:flex;align-items:center;justify-content:space-between;color:#747e89;font-size:.69rem}.result-head>div{display:flex;align-items:center;gap:8px}.result-head strong{color:#cfd4da}.result-head code{color:#9da8b3;font-size:.68rem}.result-head button{width:28px;height:28px;border:1px solid #2d323a;border-radius:7px;background:#111419;color:#87919b}.grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}.card{border:1px solid var(--border);border-radius:13px;background:#0f1115;overflow:hidden}.card-head{min-height:92px;padding:17px 20px;display:flex;align-items:center;justify-content:space-between}.identity{min-width:0;display:flex;align-items:center;gap:12px}.mark{width:44px;height:44px;flex:0 0 44px;border:1px solid #323840;border-radius:50%;background:#15191e;color:#d6dbe0;display:grid;place-items:center;font:700 .78rem ui-monospace,SFMono-Regular,monospace}.mark[data-provider=anthropic]{color:#d7b894;background:#1a1714}.mark[data-provider=google]{color:#a8c8ed;background:#13191f}.mark[data-provider=xai]{color:#e0e2e5}.identity>div{min-width:0}.identity h2{margin:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font:680 1rem/1.2 ui-monospace,SFMono-Regular,Menlo,monospace;letter-spacing:-.02em}.tags{margin-top:7px;display:flex;gap:6px;flex-wrap:wrap}.tags span{height:23px;padding:0 7px;border:1px solid #2d323a;border-radius:99px;color:#89939e;display:inline-flex;align-items:center;font-size:.61rem}.rank{color:#5e6873;font:600 .68rem ui-monospace,SFMono-Regular,monospace}.prices{margin:0 15px;display:grid;grid-template-columns:1fr 1fr;border:1px solid #252a31;border-radius:10px;overflow:hidden}.prices>div{min-height:72px;padding:13px 14px;display:flex;flex-direction:column;justify-content:center}.prices>div:nth-child(even){border-left:1px solid #252a31}.prices>div:nth-child(n+3){border-top:1px solid #252a31}.prices span{color:#747e89;font-size:.64rem}.prices strong{margin-top:6px;font-size:.85rem}.prices small{color:#68727d;font-size:.6rem;font-weight:500}.access{margin:14px 15px 0;border:1px solid #252a31;border-radius:10px;overflow:hidden}.id-row,.source-row{min-height:48px;padding:0 12px;display:grid;grid-template-columns:86px minmax(0,1fr) auto;align-items:center;gap:8px}.source-row{grid-template-columns:86px 1fr;border-top:1px solid #252a31}.access span{color:#707a85;font-size:.63rem}.access code{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:#a9c5e2;font-size:.67rem}.access button{height:30px;padding:0 9px;border:1px solid #303640;border-radius:6px;background:#15181d;color:#aeb5bd;font-size:.62rem}.source-row strong{color:#adb5bd;font-size:.66rem;font-weight:560}.routes{margin-top:14px;padding:14px 15px 15px;border-top:1px solid #24282f;background:#0c0e12}.route-title{display:flex;justify-content:space-between;color:#707a85;font-size:.63rem}.route-title strong{color:#8c96a0}.route-list{margin-top:8px;display:grid;gap:5px}.route-list>div{min-height:38px;padding:0 10px;border:1px solid #232830;border-radius:7px;background:#111419;display:flex;align-items:center;justify-content:space-between}.route-list>div>span{min-width:0;display:grid;grid-template-columns:6px auto 1fr;align-items:center;gap:7px}.route-list b{font-size:.65rem;color:#c0c6cc}.route-list small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:#606a75;font-size:.59rem}.route-list em{font-style:normal;color:#82909d;font-size:.62rem}.empty{min-height:300px;border:1px solid var(--border);border-radius:12px;background:#0f1115;display:flex;flex-direction:column;align-items:center;justify-content:center;color:#7a848e}.empty strong{color:#c8ced4}.empty span{margin-top:7px;font-size:.68rem}.loading{display:grid;grid-template-columns:1fr 1fr;gap:14px}.loading i{height:420px;border-radius:13px;background:linear-gradient(90deg,#101217,#171a20,#101217);background-size:200% 100%;animation:sk 1.2s linear infinite}@keyframes sk{to{background-position:-200% 0}}@media(max-width:980px){.heading{flex-direction:column}.actions{width:100%}.search{width:100%;flex:1}.grid{grid-template-columns:1fr}}@media(max-width:620px){.actions{flex-direction:column}.search,.sort{width:100%}.result-head{align-items:flex-start;flex-direction:column;justify-content:center;gap:8px}.prices{grid-template-columns:1fr}.prices>div:nth-child(even){border-left:0}.prices>div+div{border-top:1px solid #252a31}.loading{grid-template-columns:1fr}}
</style>
