<script setup lang="ts">
import UiSelect from '../components/ui/UiSelect.vue'
// Admin Payment Plans —— 订阅计划 CRUD
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  paymentAdminApi,
  type AdminSubscriptionPlan,
  type CreatePlanRequest,
  type UpdatePlanRequest,
} from '../api/payment'
import { api, getErrorMessage } from '../core/api'

const { t, locale } = useI18n()
const isZh = computed(() => String(locale.value || '').toLowerCase().startsWith('zh'))
const text = (zh: string, en: string) => isZh.value ? zh : en

const loading = ref(false)
const error = ref('')
const plans = ref<AdminSubscriptionPlan[]>([])

type ListResponse<T> = { items?: T[]; total?: number }
type GroupOption = {
  id: number
  name: string
  platform?: string
  subscription_type?: string
  status?: string
}
const groups = ref<GroupOption[]>([])

const sortedGroups = computed(() => {
  const rank = (g: GroupOption): number => {
    if (g.subscription_type === 'subscription' && g.status === 'active') return 0
    if (g.subscription_type === 'subscription') return 1
    if (g.status === 'active') return 2
    return 3
  }
  return [...groups.value].sort((a, b) => rank(a) - rank(b) || a.id - b.id)
})

function groupOptionLabel(g: GroupOption): string {
  return g.name ? g.name + ' (#' + g.id + ')' : '#' + g.id
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const [planList, groupResp] = await Promise.all([
      paymentAdminApi.listPlans(),
      api
        .get<ListResponse<GroupOption>>('/admin/groups', {
          params: { page: 1, page_size: 200, sort_by: 'sort_order', sort_order: 'asc' },
        })
        .catch(() => null),
    ])
    plans.value = planList || []
    const items = groupResp?.data?.items
    groups.value = Array.isArray(items) ? items : []
  } catch (e) {
    error.value = getErrorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)

interface PlanDraft {
  id?: number
  group_id: number
  name: string
  description: string
  price: number
  original_price: number | null
  currency: string
  validity_days: number
  validity_unit: string
  features: string
  product_name: string
  for_sale: boolean
  sort_order: number
}

function emptyDraft(): PlanDraft {
  return {
    group_id: 0,
    name: '',
    description: '',
    price: 0,
    original_price: null,
    currency: 'CNY',
    validity_days: 30,
    validity_unit: 'day',
    features: '',
    product_name: '',
    for_sale: true,
    sort_order: 0,
  }
}

const editorOpen = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const draft = ref<PlanDraft>(emptyDraft())
const saving = ref(false)

function openCreate(): void {
  editorMode.value = 'create'
  draft.value = emptyDraft()
  editorOpen.value = true
}

function openEdit(p: AdminSubscriptionPlan): void {
  editorMode.value = 'edit'
  draft.value = {
    id: p.id,
    group_id: p.group_id,
    name: p.name,
    description: p.description,
    price: p.price,
    original_price: p.original_price ?? null,
    currency: p.currency || 'CNY',
    validity_days: p.validity_days,
    validity_unit: p.validity_unit,
    features: p.features || '',
    product_name: p.product_name,
    for_sale: p.for_sale,
    sort_order: p.sort_order,
  }
  editorOpen.value = true
}

function closeEditor(): void {
  editorOpen.value = false
}

const isEdit = computed(() => editorMode.value === 'edit')

async function save(): Promise<void> {
  if (!draft.value.group_id) {
    error.value = t('payment.adminPlans.groupRequired')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload: CreatePlanRequest = {
      group_id: draft.value.group_id,
      name: draft.value.name,
      description: draft.value.description,
      price: draft.value.price,
      original_price: draft.value.original_price,
      currency: draft.value.currency,
      validity_days: draft.value.validity_days,
      validity_unit: draft.value.validity_unit,
      features: draft.value.features,
      product_name: draft.value.product_name,
      for_sale: draft.value.for_sale,
      sort_order: draft.value.sort_order,
    }
    if (editorMode.value === 'create') {
      await paymentAdminApi.createPlan(payload)
    } else {
      const editPayload: UpdatePlanRequest = { ...payload }
      delete editPayload.group_id
      await paymentAdminApi.updatePlan(draft.value.id!, editPayload)
    }
    closeEditor()
    await load()
  } catch (e) {
    error.value = getErrorMessage(e)
  } finally {
    saving.value = false
  }
}

const relationOpen = ref(false)
const relationPlan = ref<AdminSubscriptionPlan | null>(null)
const relationGroupId = ref(0)
const relationSaving = ref(false)

function openRelation(p: AdminSubscriptionPlan): void {
  relationPlan.value = p
  relationGroupId.value = p.group_bound ? p.group_id : (sortedGroups.value.find((g) => g.subscription_type === 'subscription' && g.status === 'active')?.id || p.group_id || 0)
  relationOpen.value = true
}

function closeRelation(): void {
  if (relationSaving.value) return
  relationOpen.value = false
  relationPlan.value = null
}

function relationGroupHint(g: GroupOption): string {
  const parts = [g.platform ? g.platform.toUpperCase() : '']
  parts.push(g.subscription_type === 'subscription' ? text('订阅计费', 'Subscription') : text('余额计费', 'Balance'))
  parts.push(g.status === 'active' ? text('启用', 'Active') : text('停用', 'Inactive'))
  return parts.filter(Boolean).join(' · ')
}

async function saveRelation(): Promise<void> {
  if (!relationPlan.value || !relationGroupId.value) return
  relationSaving.value = true
  error.value = ''
  try {
    await paymentAdminApi.bindPlanGroup(relationPlan.value.id, relationGroupId.value)
    closeRelation()
    relationOpen.value = false
    relationPlan.value = null
    await load()
  } catch (e) {
    error.value = getErrorMessage(e)
  } finally {
    relationSaving.value = false
  }
}

async function unbind(p: AdminSubscriptionPlan): Promise<void> {
  if (!p.group_bound) return
  const ok = confirm(text(
    `解除“${p.name}”与当前分组的关联？\n\n套餐会自动停止销售；已经购买的用户订阅不会被删除或迁移。`,
    `Unbind “${p.name}” from its current group?\n\nThe plan will be taken off sale automatically. Existing subscriptions will not be deleted or migrated.`,
  ))
  if (!ok) return
  error.value = ''
  try {
    await paymentAdminApi.unbindPlanGroup(p.id)
    await load()
  } catch (e) {
    error.value = getErrorMessage(e)
  }
}

async function remove(p: AdminSubscriptionPlan): Promise<void> {
  if (!confirm(t('payment.adminPlans.deleteConfirm'))) return
  try {
    await paymentAdminApi.deletePlan(p.id)
    await load()
  } catch (e) {
    error.value = getErrorMessage(e)
  }
}

function formatPrice(p: AdminSubscriptionPlan): string {
  const c = p.currency || 'CNY'
  return c === 'CNY'
    ? `¥${(p.price || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
    : `${c} ${(p.price || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatValidity(p: AdminSubscriptionPlan): string {
  const n = p.validity_days
  const u = p.validity_unit
  if (u === 'year' || u === 'years') return t('payment.planValidityYear', { n })
  if (u === 'month' || u === 'months') return t('payment.planValidityMonth', { n })
  return t('payment.planValidityDays', { n })
}
</script>

<template>
  <section class="workspace-page payment-plans">
    <header class="plans-heading">
      <div>
        <div class="plans-eyebrow"><i></i><span>PLANS</span></div>
        <h1>{{ t('payment.adminPlans.title') }}</h1>
        <p>{{ t('payment.adminPlans.description') }}</p>
      </div>
      <div class="plans-heading-actions">
        <button type="button" class="primary-btn" @click="openCreate">
          + {{ t('payment.adminPlans.addPlan') }}
        </button>
      </div>
    </header>

    <p v-if="error" class="error-banner">{{ error }}</p>

    <div v-if="!plans.length && !loading" class="empty-state">
      <strong>{{ t('payment.adminPlans.empty') }}</strong>
    </div>

    <table v-else class="plans-table">
      <thead>
        <tr>
          <th>{{ t('payment.adminPlans.colName') }}</th>
          <th>{{ t('payment.adminPlans.colGroup') }}</th>
          <th>{{ t('payment.adminPlans.colPrice') }}</th>
          <th>{{ t('payment.adminPlans.colValidity') }}</th>
          <th>{{ t('payment.adminPlans.colForSale') }}</th>
          <th>{{ t('payment.adminPlans.colSort') }}</th>
          <th>{{ t('payment.adminPlans.colAction') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in plans" :key="p.id">
          <td>
            <strong>{{ p.name }}</strong>
            <div class="plan-product">{{ p.product_name }}</div>
          </td>
          <td>
            <div v-if="p.group_bound" class="group-binding">
              <span class="binding-dot"></span>
              <div>
                <strong>{{ p.group_name || `#${p.group_id}` }}</strong>
                <small>{{ p.group_platform ? p.group_platform.toUpperCase() : text('已关联', 'Bound') }}</small>
              </div>
            </div>
            <div v-else class="group-binding unbound">
              <span class="binding-dot"></span>
              <div>
                <strong>{{ text('未关联分组', 'No group bound') }}</strong>
                <small>{{ text('套餐已自动下架', 'Plan is automatically off sale') }}</small>
              </div>
            </div>
          </td>
          <td>{{ formatPrice(p) }}<div v-if="p.original_price" class="plan-original">{{ formatPrice({ ...p, price: p.original_price }) }}</div></td>
          <td>{{ formatValidity(p) }}</td>
          <td>
            <span :class="['sale-badge', p.for_sale ? 'on' : 'off']">
              {{ p.for_sale ? t('payment.planForSale') : t('payment.planSoldOut') }}
            </span>
          </td>
          <td>{{ p.sort_order }}</td>
          <td class="action-cell">
            <button type="button" class="action relation" @click="openRelation(p)">
              {{ p.group_bound ? text('更换分组', 'Change group') : text('绑定分组', 'Bind group') }}
            </button>
            <button v-if="p.group_bound" type="button" class="action" @click="unbind(p)">{{ text('解除关联', 'Unbind') }}</button>
            <button type="button" class="action" @click="openEdit(p)">{{ t('payment.edit') }}</button>
            <button type="button" class="action danger" @click="remove(p)">{{ t('payment.delete') }}</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="relationOpen && relationPlan" class="modal-mask" @click.self="closeRelation">
      <div class="modal-card relation-modal">
        <header class="relation-head">
          <div>
            <span class="modal-eyebrow">{{ text('套餐资源关系', 'PLAN ROUTING') }}</span>
            <h3>{{ relationPlan.group_bound ? text('更换关联分组', 'Change group') : text('绑定分组', 'Bind group') }}</h3>
            <p>{{ relationPlan.name }}</p>
          </div>
          <button type="button" class="modal-close" :disabled="relationSaving" @click="closeRelation">×</button>
        </header>

        <div v-if="relationPlan.group_bound" class="current-binding">
          <span>{{ text('当前关联', 'Current group') }}</span>
          <strong>{{ relationPlan.group_name || '#' + relationPlan.group_id }}</strong>
        </div>

        <div class="relation-options">
          <button
            v-for="g in sortedGroups"
            :key="g.id"
            type="button"
            :class="['group-option', { selected: relationGroupId === g.id, risky: g.subscription_type !== 'subscription' || g.status !== 'active' }]"
            @click="relationGroupId = g.id"
          >
            <span class="group-radio"></span>
            <span class="group-option-copy">
              <strong>{{ groupOptionLabel(g) }}</strong>
              <small>{{ relationGroupHint(g) }}</small>
            </span>
            <span v-if="g.subscription_type !== 'subscription' || g.status !== 'active'" class="group-warning">
              {{ text('绑定后保持下架', 'Stays off sale') }}
            </span>
          </button>
        </div>

        <div class="relation-note">
          <strong>{{ text('更换关系不会迁移历史订阅', 'Existing subscriptions stay unchanged') }}</strong>
          <span>{{ text('新关联只决定之后可销售套餐的路由目标；已经购买的用户继续保留其原订阅分组。', 'The new binding controls future purchases. Existing users keep the group stored on their current subscription.') }}</span>
        </div>

        <footer class="modal-actions">
          <button type="button" class="btn ghost" :disabled="relationSaving" @click="closeRelation">{{ t('payment.cancel') }}</button>
          <button type="button" class="btn primary" :disabled="relationSaving || !relationGroupId" @click="saveRelation">
            {{ relationSaving ? t('payment.saving') : text('确认绑定', 'Confirm binding') }}
          </button>
        </footer>
      </div>
    </div>

    <div v-if="editorOpen" class="modal-mask" @click.self="closeEditor">
      <div class="modal-card">
        <header><h3>{{ isEdit ? t('payment.adminPlans.editTitle') : t('payment.adminPlans.createTitle') }}</h3></header>
        <div class="form-grid">
          <label class="field"><span>{{ t('payment.adminPlans.formName') }}</span><input v-model="draft.name" type="text" /></label>
          <label v-if="!isEdit" class="field">
            <span>{{ t('payment.adminPlans.formGroup') }}</span>
            <UiSelect
              v-if="groups.length"
              v-model="draft.group_id"
              :options="[{ label: t('payment.adminPlans.groupPlaceholder'), value: 0, disabled: true }, ...sortedGroups.map((g) => ({ label: groupOptionLabel(g), value: g.id }))]"
              :aria-label="t('payment.adminPlans.formGroup')"
              fluid
            />
            <input v-else v-model.number="draft.group_id" type="number" />
            <small class="field-hint">{{ text('创建后可在套餐列表中单独更换或解除关联。', 'After creation, manage the group relationship separately from the plan list.') }}</small>
          </label>
          <div v-else class="field binding-summary">
            <span>{{ t('payment.adminPlans.formGroup') }}</span>
            <strong>{{ plans.find((p) => p.id === draft.id)?.group_bound ? (plans.find((p) => p.id === draft.id)?.group_name || '#' + draft.group_id) : text('未关联分组', 'No group bound') }}</strong>
            <small>{{ text('分组关系请使用套餐列表中的“更换分组 / 解除关联”。', 'Use Change group / Unbind in the plan list to manage this relationship.') }}</small>
          </div>
          <label class="field"><span>{{ t('payment.adminPlans.formPrice') }}</span><input v-model.number="draft.price" type="number" step="0.01" /></label>
          <label class="field"><span>{{ t('payment.adminPlans.formOriginalPrice') }}</span><input v-model.number="draft.original_price" type="number" step="0.01" /></label>
          <label class="field"><span>{{ t('payment.adminPlans.formCurrency') }}</span><input v-model="draft.currency" type="text" maxlength="8" /></label>
          <label class="field"><span>{{ t('payment.adminPlans.formValidityDays') }}</span><input v-model.number="draft.validity_days" type="number" min="1" /></label>
          <label class="field"><span>{{ t('payment.adminPlans.formValidityUnit') }}</span>
            <UiSelect
              v-model="draft.validity_unit"
              :options="[{ label: 'day', value: 'day' }, { label: 'month', value: 'month' }, { label: 'year', value: 'year' }]"
              :aria-label="t('payment.adminPlans.formValidityUnit')"
              fluid
            />
          </label>
          <label class="field"><span>{{ t('payment.adminPlans.formProductName') }}</span><input v-model="draft.product_name" type="text" /></label>
          <label class="field full"><span>{{ t('payment.adminPlans.formFeatures') }}</span><textarea v-model="draft.features" rows="4" /></label>
          <label class="field full"><span>Description</span><textarea v-model="draft.description" rows="3" /></label>
          <label class="field"><span>{{ t('payment.adminPlans.formSort') }}</span><input v-model.number="draft.sort_order" type="number" /></label>
          <label class="checkbox"><input v-model="draft.for_sale" type="checkbox" /><span>{{ t('payment.adminPlans.formForSale') }}</span></label>
        </div>
        <footer class="modal-actions">
          <button type="button" class="btn ghost" :disabled="saving" @click="closeEditor">{{ t('payment.cancel') }}</button>
          <button type="button" class="btn primary" :disabled="saving" @click="save">
            {{ saving ? t('payment.saving') : t('payment.save') }}
          </button>
        </footer>
      </div>
    </div>
  </section>
</template>

<style scoped>
.payment-plans { width: 100%; max-width: 1280px; margin: 0 auto; padding: 12px 0 44px; }
.plans-heading { display: flex; align-items: center; justify-content: space-between; min-height: 92px; margin-bottom: 22px; gap: 24px; }
.plans-heading h1 { margin: 0; color: #f7f8fa; font-size: clamp(1.9rem, 2.4vw, 2.35rem); line-height: 1.08; font-weight: 680; letter-spacing: -.043em; }
.plans-heading p { max-width: 640px; margin: 10px 0 0; color: #858d97; font-size: .88rem; line-height: 1.6; }
.plans-eyebrow { display: inline-flex; align-items: center; gap: 8px; color: #6ec0f5; font: 700 .67rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .13em; }
.plans-eyebrow i { width: 5px; height: 5px; border-radius: 50%; background: #6ec0f5; }
.primary-btn { padding: 8px 16px; border-radius: 8px; background: #4a93c5; border: 1px solid #4a93c5; color: #0d0f12; font: 600 .82rem/1 ui-sans-serif, system-ui, sans-serif; cursor: pointer; }
.primary-btn:hover { background: #5fa3d5; border-color: #5fa3d5; }
.error-banner { margin: 0 0 16px; padding: 10px 14px; border-radius: 8px; background: rgba(239, 68, 68, .12); border: 1px solid rgba(239, 68, 68, .35); color: #fca5a5; font-size: .85rem; }
.empty-state { padding: 56px 0; text-align: center; color: #6c727b; font-size: .85rem; }
.plans-table { width: 100%; border-collapse: collapse; padding: 0; border-radius: 14px; background: #11141a; border: 1px solid #1d2128; overflow: hidden; }
.plans-table th { text-align: left; padding: 12px 14px; color: #6c727b; font: 500 .72rem/1 ui-sans-serif, system-ui, sans-serif; text-transform: uppercase; letter-spacing: .08em; border-bottom: 1px solid #1d2128; background: #0f1217; }
.plans-table td { padding: 14px; color: #d6dbe1; font-size: .85rem; border-bottom: 1px solid #161a20; vertical-align: top; }
.plans-table tr:last-child td { border-bottom: 0; }
.plans-table strong { display: block; color: #f7f8fa; font-weight: 600; }
.plan-product { color: #6c727b; font-size: .72rem; margin-top: 2px; }
.plan-original { color: #6c727b; font-size: .72rem; text-decoration: line-through; }
.sale-badge { display: inline-block; padding: 3px 8px; border-radius: 12px; font: 500 .72rem/1 ui-sans-serif, system-ui, sans-serif; }
.sale-badge.on { background: rgba(72, 187, 153, .12); color: #48bb99; }
.sale-badge.off { background: rgba(140, 145, 152, .12); color: #8c9198; }
.action-cell { white-space: nowrap; }
.action { padding: 5px 10px; margin-right: 6px; border-radius: 6px; background: #16191f; border: 1px solid #2a2f37; color: #d6dbe1; font: 500 .72rem/1 ui-sans-serif, system-ui, sans-serif; cursor: pointer; }
.action:hover { border-color: #3d4754; }
.action.danger { color: #fca5a5; border-color: rgba(239, 68, 68, .35); }
.action.danger:hover { background: rgba(239, 68, 68, .08); }
.modal-mask { position: fixed; inset: 0; background: rgba(8, 10, 14, .7); display: flex; align-items: flex-start; justify-content: center; z-index: 1000; overflow-y: auto; padding: 40px 16px; }
.modal-card { width: min(720px, 100%); padding: 24px 26px; background: #11141a; border: 1px solid #1d2128; border-radius: 14px; }
.modal-card h3 { margin: 0 0 18px; color: #f7f8fa; font-size: 1.1rem; font-weight: 600; }
.form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.field { display: block; }
.field.full { grid-column: 1 / -1; }
.field > span { display: block; margin-bottom: 5px; color: #b8bfc7; font-size: .78rem; }
.field input, .field textarea, .field select { width: 100%; padding: 8px 12px; border-radius: 7px; background: #0d0f12; border: 1px solid #2a2f37; color: #f7f8fa; font: 400 .82rem/1.4 ui-sans-serif, system-ui, sans-serif; box-sizing: border-box; }
.field input:focus, .field textarea:focus, .field select:focus { outline: none; border-color: #4a93c5; }
.checkbox { display: inline-flex; align-items: center; gap: 8px; margin: 18px 0 0; color: #b8bfc7; font-size: .82rem; cursor: pointer; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
.btn { padding: 9px 20px; border-radius: 7px; border: 1px solid transparent; font: 500 .85rem/1 ui-sans-serif, system-ui, sans-serif; cursor: pointer; }
.btn.ghost { background: #16191f; border-color: #2a2f37; color: #d6dbe1; }
.btn.ghost:hover:not(:disabled) { border-color: #3d4754; }
.btn.primary { background: #4a93c5; color: #0d0f12; }
.btn.primary:hover:not(:disabled) { background: #5fa3d5; }
.btn:disabled { opacity: .6; cursor: not-allowed; }

.group-binding { display: flex; align-items: center; gap: 9px; min-width: 170px; }
.group-binding > div { min-width: 0; }
.group-binding strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.group-binding small { display: block; margin-top: 3px; color: #6c727b; font-size: .67rem; }
.binding-dot { width: 7px; height: 7px; flex: 0 0 auto; border-radius: 50%; background: #48bb99; box-shadow: 0 0 0 4px rgba(72,187,153,.08); }
.group-binding.unbound strong { color: #9ca3ad; }
.group-binding.unbound .binding-dot { background: #79818b; box-shadow: 0 0 0 4px rgba(121,129,139,.08); }
.action.relation { color: #8fd0fb; border-color: rgba(74,147,197,.45); background: rgba(74,147,197,.08); }
.field-hint, .binding-summary small { display: block; margin-top: 6px; color: #717a84; font-size: .68rem; line-height: 1.5; }
.binding-summary { padding: 8px 0; }
.binding-summary strong { color: #e8edf2; font-size: .86rem; }
.relation-modal { width: min(640px, 100%); }
.relation-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; }
.relation-head h3 { margin: 5px 0 0; }
.relation-head p { margin: 6px 0 0; color: #858d97; font-size: .78rem; }
.modal-eyebrow { color: #6ec0f5; font: 700 .63rem/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .12em; }
.modal-close { width: 34px; height: 34px; border: 1px solid #2a2f37; border-radius: 9px; background: #16191f; color: #9ca3ad; font-size: 1.15rem; cursor: pointer; }
.current-binding { display: flex; align-items: center; justify-content: space-between; gap: 14px; margin: 18px 0 12px; padding: 12px 14px; border: 1px solid #242a32; border-radius: 10px; background: #0d1015; }
.current-binding span { color: #737d88; font-size: .72rem; }
.current-binding strong { color: #e8edf2; font-size: .82rem; }
.relation-options { display: grid; gap: 8px; max-height: 330px; overflow-y: auto; padding: 2px; }
.group-option { width: 100%; display: grid; grid-template-columns: auto minmax(0,1fr) auto; align-items: center; gap: 11px; min-height: 62px; padding: 11px 13px; border: 1px solid #262c34; border-radius: 11px; background: #101319; color: #dbe1e7; text-align: left; cursor: pointer; }
.group-option:hover { border-color: #3a4652; background: #13171d; }
.group-option.selected { border-color: #4a93c5; background: rgba(74,147,197,.09); box-shadow: inset 0 0 0 1px rgba(74,147,197,.08); }
.group-radio { width: 14px; height: 14px; border: 1.5px solid #56616d; border-radius: 50%; }
.group-option.selected .group-radio { border: 4px solid #65b1e3; background: #0f1217; }
.group-option-copy { min-width: 0; }
.group-option-copy strong { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #eef2f5; font-size: .82rem; }
.group-option-copy small { display: block; margin-top: 4px; color: #737d88; font-size: .68rem; }
.group-warning { padding: 4px 7px; border-radius: 999px; background: rgba(245,158,11,.1); color: #f0b85a; font-size: .62rem; white-space: nowrap; }
.relation-note { display: grid; gap: 5px; margin-top: 14px; padding: 12px 14px; border-left: 2px solid #4a93c5; background: rgba(74,147,197,.06); }
.relation-note strong { color: #cbd4dc; font-size: .74rem; }
.relation-note span { color: #75808b; font-size: .69rem; line-height: 1.55; }

:global(html.smirel-app[data-theme='light']) .group-binding strong,
:global(html.smirel-app[data-theme='light']) .binding-summary strong,
:global(html.smirel-app[data-theme='light']) .current-binding strong,
:global(html.smirel-app[data-theme='light']) .group-option-copy strong { color: #25313c; }
:global(html.smirel-app[data-theme='light']) .modal-close,
:global(html.smirel-app[data-theme='light']) .current-binding,
:global(html.smirel-app[data-theme='light']) .group-option { border-color: #dce4eb; background: #fff; color: #374553; }
:global(html.smirel-app[data-theme='light']) .group-option:hover { border-color: #bfd2e1; background: #f8fbfd; }
:global(html.smirel-app[data-theme='light']) .group-option.selected { border-color: #6ba9d3; background: #f0f8fd; }
:global(html.smirel-app[data-theme='light']) .group-option.selected .group-radio { background: #fff; }
:global(html.smirel-app[data-theme='light']) .current-binding { background: #f8fafc; }

@media (max-width: 720px) { .plans-heading { flex-direction: column; align-items: flex-start; } .form-grid { grid-template-columns: 1fr; } .plans-table th, .plans-table td { padding: 8px; font-size: .72rem; } .group-warning { display: none; } }
</style>
