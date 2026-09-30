<script setup lang="ts">
import UiSelect from '../components/ui/UiSelect.vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { api, getErrorMessage } from '../core/api'

type GroupRow={id:number;name:string;platform:string;status:string}
type Channel={id:number;name:string;description?:string;status:string;billing_model_source?:string;restrict_models?:boolean;group_ids?:number[];model_mapping?:Record<string,Record<string,string>>;model_pricing?:Array<{platform:string;models:string[];billing_mode:string;input_price?:number|null;output_price?:number|null}>;updated_at?:string}
type ListResponse<T>={items?:T[];total?:number}
const loading=ref(false), error=ref(''), notice=ref(''), search=ref(''), status=ref('')
const channels=ref<Channel[]>([]), groups=ref<GroupRow[]>([])
const showEditor=ref(false), editingId=ref<number|null>(null), busy=ref(false), togglingId=ref<number|null>(null)
const form=reactive({name:'',description:'',status:'active',billing_model_source:'requested',restrict_models:true,group_ids:[] as number[],platform:'openai',models:''})

const filtered=computed(()=>{const q=search.value.trim().toLowerCase();return channels.value.filter(c=>(!status.value||c.status===status.value)&&(!q||`${c.name} ${c.description||''} ${c.id} ${groupNames(c)}`.toLowerCase().includes(q)))})
const activeCount=computed(()=>channels.value.filter(c=>c.status==='active').length)
const explicitPricingCount=computed(()=>channels.value.reduce((n,c)=>n+pricingCount(c),0))
const hasFilters=computed(()=>Boolean(search.value.trim()||status.value))
const modelCount=computed(()=>channels.value.reduce((n,c)=>n+mappingModels(c).length,0))
const groupCount=computed(()=>new Set(channels.value.flatMap(c=>c.group_ids||[])).size)
function groupName(id:number){return groups.value.find(g=>g.id===id)?.name||'分组 #'+id}
function groupLabels(c:Channel){return (c.group_ids||[]).map(groupName)}
function groupNames(c:Channel){return groupLabels(c).join(' · ')||'未绑定'}
function billingSourceLabel(source?:string){return ({requested:'请求模型',upstream:'上游模型',channel_mapped:'映射模型'} as Record<string,string>)[source||'']||source||'请求模型'}
function resetFilters(){search.value='';status.value=''}
function mappingModels(c:Channel){const s=new Set<string>();Object.values(c.model_mapping||{}).forEach(m=>Object.keys(m||{}).forEach(x=>s.add(x)));return [...s].sort()}
function pricingCount(c:Channel){return (c.model_pricing||[]).reduce((n,p)=>n+(p.models?.length||0),0)}
function platformLabel(p:string){return ({openai:'OpenAI Compatible',anthropic:'Anthropic',gemini:'Gemini',antigravity:'Antigravity',grok:'xAI / Grok'} as Record<string,string>)[p]||p}
async function load(){loading.value=true;error.value='';try{const [cr,gr]=await Promise.all([api.get<ListResponse<Channel>>('/admin/channels',{params:{page:1,page_size:200,sort_by:'created_at',sort_order:'desc'}}),api.get<GroupRow[]>('/admin/groups/all',{params:{include_inactive:true}})]);channels.value=Array.isArray(cr.data?.items)?cr.data.items:[];groups.value=Array.isArray(gr.data)?gr.data:[]}catch(e){error.value=getErrorMessage(e)}finally{loading.value=false}}
function openCreate(){editingId.value=null;Object.assign(form,{name:'',description:'',status:'active',billing_model_source:'requested',restrict_models:true,group_ids:[],platform:'openai',models:''});showEditor.value=true}
function openEdit(c:Channel){editingId.value=c.id;const platforms=Object.keys(c.model_mapping||{});Object.assign(form,{name:c.name,description:c.description||'',status:c.status||'active',billing_model_source:c.billing_model_source||'requested',restrict_models:c.restrict_models!==false,group_ids:[...(c.group_ids||[])],platform:platforms[0]||groups.value.find(g=>c.group_ids?.includes(g.id))?.platform||'openai',models:mappingModels(c).join('\n')});showEditor.value=true}
function parseModels(){return [...new Set(form.models.split(/[\n,]+/).map(x=>x.trim()).filter(Boolean))].sort()}
async function save(){if(!form.name.trim())return;busy.value=true;error.value='';try{const models=parseModels();const base={name:form.name.trim(),description:form.description.trim(),group_ids:form.group_ids,billing_model_source:form.billing_model_source,restrict_models:form.restrict_models,features:'',features_config:{},apply_pricing_to_account_stats:false};if(editingId.value){const existing=channels.value.find(c=>c.id===editingId.value);const payload:any={...base,status:form.status};if(models.length){payload.model_mapping={...(existing?.model_mapping||{}),[form.platform]:Object.fromEntries(models.map(m=>[m,m]))}}await api.put(`/admin/channels/${editingId.value}`,payload)}else{if(!models.length)throw new Error('新建渠道至少填写一个模型');await api.post('/admin/channels',{...base,model_mapping:{[form.platform]:Object.fromEntries(models.map(m=>[m,m]))},model_pricing:[]})}showEditor.value=false;notice.value=editingId.value?'渠道已更新':'渠道已创建';await load()}catch(e){error.value=getErrorMessage(e)}finally{busy.value=false}}
async function toggle(c:Channel){if(togglingId.value!==null)return;togglingId.value=c.id;error.value='';try{await api.put(`/admin/channels/${c.id}`,{status:c.status==='active'?'disabled':'active'});await load()}catch(e){error.value=getErrorMessage(e)}finally{togglingId.value=null}}
onMounted(()=>void load())
</script>

<template>
  <section class="channels-page">
    <header class="head">
      <div class="head-copy">
        <span class="section-eyebrow">ROUTING &amp; CATALOG</span>
        <h1>渠道与价格</h1>
        <p>统一管理模型映射、渠道分组与定价策略。</p>
      </div>
      <div class="head-actions">
        <button class="ghost refresh-action" type="button" :disabled="loading" @click="load">
          <svg :class="{ spinning: loading }" viewBox="0 0 20 20" aria-hidden="true">
            <path d="M16.3 6.4A7 7 0 1 0 17 12M16.3 2.8v3.8h-3.8" />
          </svg>
          {{ loading ? '刷新中' : '刷新数据' }}
        </button>
        <button class="primary create-action" type="button" @click="openCreate">
          <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M10 4v12M4 10h12" /></svg>
          新增渠道
        </button>
      </div>
    </header>

    <section class="stats" aria-label="渠道概览">
      <article class="stat-card stat-channels">
        <div class="stat-heading"><span>渠道总数</span><i class="stat-glyph" aria-hidden="true">01</i></div>
        <strong>{{ channels.length }}</strong>
        <small><span class="stat-status-dot"></span>{{ activeCount }} 个正在启用</small>
      </article>
      <article class="stat-card stat-models">
        <div class="stat-heading"><span>模型映射</span><i class="stat-glyph" aria-hidden="true">02</i></div>
        <strong>{{ modelCount }}</strong>
        <small>当前渠道的模型映射条目</small>
      </article>
      <article class="stat-card stat-groups">
        <div class="stat-heading"><span>关联分组</span><i class="stat-glyph" aria-hidden="true">03</i></div>
        <strong>{{ groupCount }}</strong>
        <small>参与模型路由与展示</small>
      </article>
      <article class="stat-card stat-pricing">
        <div class="stat-heading"><span>显式定价</span><i class="stat-glyph" aria-hidden="true">04</i></div>
        <strong>{{ explicitPricingCount }}</strong>
        <small>其他模型使用参考价格</small>
      </article>
    </section>

    <p v-if="error" class="alert bad" role="alert">{{ error }}</p>
    <p v-if="notice" class="alert good" role="status">{{ notice }}</p>

    <section class="panel">
      <header class="toolbar">
        <div class="toolbar-title">
          <div><strong>渠道配置</strong><span class="toolbar-count">{{ filtered.length }} / {{ channels.length }}</span></div>
          <small>模型路由与计费规则</small>
        </div>
        <div class="toolbar-controls">
          <label class="channel-search">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <circle cx="8.5" cy="8.5" r="5.5" /><path d="m13 13 4 4" />
            </svg>
            <input v-model="search" type="search" aria-label="搜索渠道" placeholder="搜索渠道、分组或 ID" />
          </label>
          <UiSelect
            v-model="status"
            :options="[{ label: '全部状态', value: '' }, { label: '启用', value: 'active' }, { label: '停用', value: 'disabled' }]"
            aria-label="筛选渠道状态"
            min-width="142px"
          />
          <button v-if="hasFilters" class="clear-filters" type="button" @click="resetFilters">清除筛选</button>
        </div>
      </header>

      <div class="table" role="table" aria-label="渠道配置">
        <div class="thead" role="row">
          <span role="columnheader">渠道</span>
          <span role="columnheader">状态</span>
          <span role="columnheader">关联分组</span>
          <span role="columnheader">模型</span>
          <span role="columnheader">定价</span>
          <span role="columnheader">计费依据</span>
          <span class="actions-heading" role="columnheader">快捷操作</span>
        </div>

        <div v-for="c in filtered" :key="c.id" class="row" role="row">
          <span class="identity" role="cell">
            <i class="channel-mark" aria-hidden="true">
              <svg viewBox="0 0 24 24"><path d="M6 6.5h6.5c3 0 5 2 5 5v5M6 17.5h7M6 6.5v11" /><circle cx="6" cy="6.5" r="2" /><circle cx="6" cy="17.5" r="2" /><circle cx="17.5" cy="17.5" r="2" /></svg>
            </i>
            <span class="identity-copy">
              <b :title="c.name">{{ c.name }}</b>
              <small :title="'#' + c.id + ' · ' + (c.description || '暂无说明')">#{{ c.id }} · {{ c.description || '暂无说明' }}</small>
              <span class="mobile-state" :class="{ disabled: c.status !== 'active' }">{{ c.status === 'active' ? '运行中' : '已停用' }}</span>
            </span>
          </span>
          <span class="status-cell" role="cell">
            <em class="state" :class="c.status"><i></i>{{ c.status === 'active' ? '已启用' : '已停用' }}</em>
          </span>
          <span class="groups" role="cell" :title="groupNames(c)">
            <template v-if="c.group_ids?.length">
              <span v-for="(label, index) in groupLabels(c).slice(0, 2)" :key="index" class="group-chip">{{ label }}</span>
              <span v-if="c.group_ids.length > 2" class="group-overflow">+{{ c.group_ids.length - 2 }}</span>
            </template>
            <span v-else class="group-unbound">未绑定分组</span>
          </span>
          <span class="number-cell models-cell" role="cell">
            <b>{{ mappingModels(c).length }}</b><small class="stack">个映射</small>
          </span>
          <span class="number-cell pricing-cell" role="cell">
            <b v-if="pricingCount(c)">{{ pricingCount(c) }}</b>
            <span v-else class="pricing-fallback">参考价</span>
            <small class="stack">{{ pricingCount(c) ? '个显式价' : '未设显式价' }}</small>
          </span>
          <span class="billing-cell" role="cell">
            <span class="billing-pill" :title="'计费依据：' + billingSourceLabel(c.billing_model_source)">{{ billingSourceLabel(c.billing_model_source) }}</span>
          </span>
          <span class="actions" role="cell">
            <button class="toggle-action" type="button" :disabled="togglingId !== null" :title="c.status === 'active' ? '暂停渠道' : '启用渠道'" :aria-label="(c.status === 'active' ? '暂停 ' : '启用 ') + c.name" @click="toggle(c)">
              <svg v-if="c.status === 'active'" viewBox="0 0 20 20" aria-hidden="true"><path d="M6.5 5v10M13.5 5v10" /></svg>
              <svg v-else viewBox="0 0 20 20" aria-hidden="true"><path d="m7 4.5 9 5.5-9 5.5z" /></svg>
              <span>{{ togglingId === c.id ? '处理中' : c.status === 'active' ? '暂停' : '启用' }}</span>
            </button>
            <button class="edit-action" type="button" :title="'编辑 ' + c.name" :aria-label="'编辑 ' + c.name" @click="openEdit(c)">
              <svg viewBox="0 0 20 20" aria-hidden="true"><path d="m4 14.5.8-3.6 7.6-7.6 3.3 3.3-7.6 7.6-4.1.3Z" /></svg>
              <span>编辑</span>
            </button>
          </span>
        </div>

        <div v-if="loading && !channels.length" class="table-loading" aria-label="正在加载渠道">
          <i v-for="n in 3" :key="n"></i>
        </div>
        <div v-else-if="!filtered.length" class="empty">
          <span class="empty-symbol" aria-hidden="true">⌕</span>
          <strong>{{ hasFilters ? '没有符合条件的渠道' : '还没有渠道配置' }}</strong>
          <small>{{ hasFilters ? '可以更换关键词或清除筛选。' : '新增渠道后，模型映射和定价状态会显示在这里。' }}</small>
          <button v-if="hasFilters" type="button" @click="resetFilters">清除筛选</button>
          <button v-else type="button" @click="openCreate">新增渠道</button>
        </div>
      </div>
    </section>

    <div v-if="showEditor" class="overlay" @click.self="showEditor = false" @keydown.esc="showEditor = false">
      <form class="dialog" @submit.prevent="save">
        <header>
          <div><span>CHANNEL CONFIGURATION</span><h2>{{ editingId ? '编辑渠道' : '新增渠道' }}</h2></div>
          <button type="button" aria-label="关闭编辑窗口" @click="showEditor = false">×</button>
        </header>
        <div class="form">
          <label class="wide"><span>渠道名称</span><input v-model="form.name" required /></label>
          <label><span>模型协议</span>
            <UiSelect v-model="form.platform" :options="[{ label: 'OpenAI Compatible', value: 'openai' }, { label: 'Anthropic', value: 'anthropic' }, { label: 'Gemini', value: 'gemini' }, { label: 'Antigravity', value: 'antigravity' }, { label: 'xAI / Grok', value: 'grok' }]" aria-label="模型协议" fluid />
          </label>
          <label><span>计费模型来源</span>
            <UiSelect v-model="form.billing_model_source" :options="[{ label: '请求模型', value: 'requested' }, { label: '上游模型', value: 'upstream' }, { label: '渠道映射模型', value: 'channel_mapped' }]" aria-label="计费模型来源" fluid />
          </label>
          <label v-if="editingId"><span>状态</span>
            <UiSelect v-model="form.status" :options="[{ label: '启用', value: 'active' }, { label: '停用', value: 'disabled' }]" aria-label="渠道状态" fluid />
          </label>
          <label class="wide"><span>关联分组</span>
            <div class="checks"><label v-for="g in groups" :key="g.id"><input v-model="form.group_ids" type="checkbox" :value="g.id" /><span>{{ g.name }} · {{ platformLabel(g.platform) }}</span></label></div>
          </label>
          <label class="wide"><span>模型映射（每行一个，默认一一映射）</span><textarea v-model="form.models" rows="7" placeholder="gpt-5&#10;claude-sonnet-4-6&#10;gemini-..." /></label>
          <label class="wide inline"><input v-model="form.restrict_models" type="checkbox" /><span>限制为上述模型映射</span></label>
          <label class="wide"><span>说明</span><textarea v-model="form.description" rows="3" /></label>
        </div>
        <footer>
          <button type="button" class="ghost" @click="showEditor = false">取消</button>
          <button class="primary" :disabled="busy">{{ busy ? '保存中…' : '保存渠道' }}</button>
        </footer>
      </form>
    </div>
  </section>
</template>

<style scoped>.channels-page{--border:#252a31;--border2:#353b44;--text:#f4f6f8;color:var(--text);width:100%;font-size:14px}.head{display:flex;justify-content:space-between;gap:24px;padding:2px 0 24px}.head>div:first-child>span,.dialog header span{color:#65707c;font-size:.65rem;font-weight:700;letter-spacing:.12em}.head h1{margin:8px 0 0;font-size:2.05rem;line-height:1;font-weight:700;letter-spacing:-.045em}.head p{margin:11px 0 0;color:#838d98;font-size:.84rem}.head>div:last-child{display:flex;gap:8px}.ghost,.primary{height:40px;padding:0 14px;border:1px solid var(--border2);border-radius:8px;background:#13161b;color:#cbd1d7;cursor:pointer}.primary{border-color:#e0e4e8;background:#f1f3f5;color:#111318;font-weight:680}.stats{display:grid;grid-template-columns:repeat(4,1fr);border:1px solid var(--border);border-radius:12px;background:#101217;overflow:hidden}.stats article{min-height:110px;padding:20px;position:relative;display:flex;flex-direction:column;justify-content:center}.stats article+article:before{content:"";position:absolute;left:0;top:20px;bottom:20px;width:1px;background:#282d34}.stats span{color:#7a848f;font-size:.68rem}.stats strong{margin-top:8px;font-size:1.75rem}.stats small{margin-top:7px;color:#626c77;font-size:.63rem}.alert{margin:12px 0 0;padding:10px 12px;border-radius:8px;font-size:.7rem}.alert.bad{border:1px solid rgba(225,108,115,.28);background:rgba(225,108,115,.06);color:#e4a0a5}.alert.good{border:1px solid rgba(67,205,152,.22);background:rgba(67,205,152,.05);color:#8dd7b9}.panel{margin-top:14px;border:1px solid var(--border);border-radius:12px;background:#0f1115;overflow:hidden}.toolbar{min-height:68px;padding:12px 14px 12px 18px;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid var(--border)}.toolbar>div{display:flex;align-items:center;gap:8px}.toolbar span{color:#65707b;font-size:.65rem}.toolbar input,.toolbar select,.form input,.form select,.form textarea{border:1px solid #2d323a;border-radius:8px;background:#0b0d11;color:#e8ebee;outline:none}.toolbar input{width:270px;height:40px;padding:0 11px}.toolbar select{height:40px;padding:0 10px}.thead,.row{display:grid;grid-template-columns:minmax(250px,1.45fr) 90px minmax(170px,1fr) 80px 90px 120px 130px;gap:14px;align-items:center}.thead{min-height:42px;padding:0 18px;border-bottom:1px solid #22272e;color:#68727d;font-size:.62rem}.row{min-height:70px;padding:0 18px;border-bottom:1px solid #1f2329}.row:last-child{border-bottom:0}.identity{display:flex;align-items:center;gap:10px;min-width:0}.identity>i{width:34px;height:34px;border:1px solid #333942;border-radius:9px;background:#171a1f;display:grid;place-items:center;font-style:normal;font-weight:700}.identity>span{min-width:0;display:flex;flex-direction:column}.identity b{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:.74rem}.identity small{margin-top:4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:#68727d;font-size:.61rem}.state{display:inline-flex;align-items:center;gap:6px;font-style:normal;color:#b7bec6;font-size:.65rem}.state i{width:6px;height:6px;border-radius:50%;background:#7b848e}.state.active{color:#8ed8ba}.state.active i{background:#43cd98}.groups{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:#a9b1b9;font-size:.67rem}.stack{display:block;margin-top:4px;color:#68727d;font-size:.59rem}.row code{color:#98abc0;font-size:.63rem}.actions{display:flex;gap:5px}.actions button{height:30px;padding:0 8px;border:1px solid #303640;border-radius:6px;background:#14171c;color:#aeb6be;font-size:.61rem}.empty{min-height:240px;display:grid;place-items:center;color:#68727d}.overlay{position:fixed;inset:0;z-index:80;padding:22px;background:rgba(0,0,0,.7);backdrop-filter:blur(8px);display:grid;place-items:center}.dialog{width:min(690px,100%);max-height:90vh;overflow:auto;border:1px solid #343a43;border-radius:14px;background:#111318}.dialog>header{padding:19px 21px;display:flex;justify-content:space-between;border-bottom:1px solid #262b32}.dialog h2{margin:6px 0 0;font-size:1.2rem}.dialog>header button{width:34px;height:34px;border:1px solid #303640;border-radius:8px;background:#171a1f;color:#a2abb5}.form{padding:20px 21px;display:grid;grid-template-columns:1fr 1fr;gap:14px}.form>label{display:flex;flex-direction:column;gap:7px}.form>label.wide{grid-column:1/-1}.form>label>span{color:#7d8792;font-size:.67rem}.form input,.form select{height:40px;padding:0 10px}.form textarea{padding:10px;resize:vertical}.checks{max-height:150px;padding:8px;border:1px solid #292e36;border-radius:8px;background:#0b0d11;display:grid;grid-template-columns:1fr 1fr;gap:4px}.checks label{min-height:32px;padding:0 7px;border-radius:6px;display:flex;align-items:center;gap:8px;color:#b8c0c8;font-size:.65rem}.checks label:hover{background:#14181e}.checks input{height:auto}.form .inline{flex-direction:row;align-items:center}.form .inline input{height:auto}.dialog footer{padding:14px 21px;border-top:1px solid #262b32;display:flex;justify-content:flex-end;gap:8px}/* Responsive table columns are now controlled by the panel container in admin-channels-polish.css. */@media(max-width:720px){.head{flex-direction:column}.head>div:last-child{width:100%}.head button{flex:1}.stats{grid-template-columns:1fr 1fr}.stats article:nth-child(3):before{display:none}.stats article:nth-child(n+3){border-top:1px solid #282d34}.toolbar{align-items:flex-start;flex-direction:column;gap:9px}.toolbar>div:last-child{width:100%}.toolbar input{width:100%;flex:1}/* Panel container queries handle table density independently of viewport width. */.form{grid-template-columns:1fr}.form>label.wide{grid-column:auto}.checks{grid-template-columns:1fr}}</style>