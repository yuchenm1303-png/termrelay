<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'

export type UiSelectValue = any

export type UiSelectOption = {
  label: string
  value: UiSelectValue
  disabled?: boolean
  hint?: string
}

const props = withDefaults(defineProps<{
  modelValue: UiSelectValue
  options: UiSelectOption[]
  placeholder?: string
  ariaLabel?: string
  disabled?: boolean
  fluid?: boolean
  minWidth?: string
  menuMinWidth?: string
}>(), {
  placeholder: '请选择',
  ariaLabel: '',
  disabled: false,
  fluid: false,
  minWidth: '132px',
  menuMinWidth: '168px',
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: UiSelectValue): void
  (event: 'change', value: UiSelectValue): void
  (event: 'open'): void
  (event: 'close'): void
}>()

const uid = `ui-select-${Math.random().toString(36).slice(2, 9)}`
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const open = ref(false)
const highlightedIndex = ref(-1)
const placement = ref<'bottom' | 'top'>('bottom')
const menuStyle = ref<Record<string, string>>({})

const selectedIndex = computed(() =>
  props.options.findIndex((option) => Object.is(option.value, props.modelValue)),
)

const selectedOption = computed(() =>
  selectedIndex.value >= 0 ? props.options[selectedIndex.value] : null,
)

const triggerLabel = computed(() => selectedOption.value?.label || props.placeholder)

function firstEnabledIndex() {
  return props.options.findIndex((option) => !option.disabled)
}

function lastEnabledIndex() {
  for (let index = props.options.length - 1; index >= 0; index -= 1) {
    if (!props.options[index]?.disabled) return index
  }
  return -1
}

function updatePosition() {
  const trigger = triggerRef.value
  if (!trigger) return

  const rect = trigger.getBoundingClientRect()
  const gap = 8
  const viewportPadding = 10
  const estimatedHeight = Math.min(300, Math.max(52, props.options.length * 44 + 12))
  const roomBelow = window.innerHeight - rect.bottom - viewportPadding
  const roomAbove = rect.top - viewportPadding
  const shouldOpenAbove = roomBelow < estimatedHeight && roomAbove > roomBelow
  const maxHeight = Math.max(120, Math.min(300, (shouldOpenAbove ? roomAbove : roomBelow) - gap))
  const desiredWidth = Math.max(rect.width, Number.parseFloat(props.menuMinWidth) || 168)
  const width = Math.min(desiredWidth, window.innerWidth - viewportPadding * 2)
  const left = Math.min(
    Math.max(viewportPadding, rect.left),
    Math.max(viewportPadding, window.innerWidth - viewportPadding - width),
  )

  placement.value = shouldOpenAbove ? 'top' : 'bottom'
  menuStyle.value = {
    left: `${Math.round(left)}px`,
    width: `${Math.round(width)}px`,
    maxHeight: `${Math.round(maxHeight)}px`,
    ...(shouldOpenAbove
      ? { bottom: `${Math.round(window.innerHeight - rect.top + gap)}px`, top: 'auto' }
      : { top: `${Math.round(rect.bottom + gap)}px`, bottom: 'auto' }),
  }
}

function addViewportListeners() {
  window.addEventListener('resize', updatePosition)
  window.addEventListener('scroll', updatePosition, true)
}

function removeViewportListeners() {
  window.removeEventListener('resize', updatePosition)
  window.removeEventListener('scroll', updatePosition, true)
}

async function openMenu() {
  if (props.disabled || open.value || !props.options.length) return
  highlightedIndex.value = selectedIndex.value >= 0 ? selectedIndex.value : firstEnabledIndex()
  open.value = true
  emit('open')
  await nextTick()
  updatePosition()
  addViewportListeners()
  scrollHighlightedIntoView()
}

function closeMenu({ focusTrigger = false } = {}) {
  if (!open.value) return
  open.value = false
  highlightedIndex.value = -1
  removeViewportListeners()
  emit('close')
  if (focusTrigger) {
    void nextTick(() => triggerRef.value?.focus())
  }
}

function toggleMenu() {
  if (open.value) closeMenu()
  else void openMenu()
}

function selectOption(index: number) {
  const option = props.options[index]
  if (!option || option.disabled) return
  emit('update:modelValue', option.value)
  emit('change', option.value)
  closeMenu({ focusTrigger: true })
}

function scrollHighlightedIntoView() {
  void nextTick(() => {
    const menu = menuRef.value
    if (!menu || highlightedIndex.value < 0) return
    menu.querySelector<HTMLElement>(`[data-option-index="${highlightedIndex.value}"]`)?.scrollIntoView({
      block: 'nearest',
    })
  })
}

function moveHighlight(direction: 1 | -1) {
  if (!props.options.length) return
  let index = highlightedIndex.value

  if (index < 0) {
    index = direction > 0 ? firstEnabledIndex() : lastEnabledIndex()
    highlightedIndex.value = index
    scrollHighlightedIntoView()
    return
  }

  for (let count = 0; count < props.options.length; count += 1) {
    index = (index + direction + props.options.length) % props.options.length
    if (!props.options[index]?.disabled) {
      highlightedIndex.value = index
      scrollHighlightedIntoView()
      return
    }
  }
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (props.disabled) return

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (!open.value) void openMenu()
    else moveHighlight(1)
    return
  }

  if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (!open.value) void openMenu()
    else moveHighlight(-1)
    return
  }

  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    if (!open.value) void openMenu()
    else if (highlightedIndex.value >= 0) selectOption(highlightedIndex.value)
    return
  }

  if (event.key === 'Escape' && open.value) {
    event.preventDefault()
    closeMenu({ focusTrigger: true })
    return
  }

  if (event.key === 'Home' && open.value) {
    event.preventDefault()
    highlightedIndex.value = firstEnabledIndex()
    scrollHighlightedIntoView()
    return
  }

  if (event.key === 'End' && open.value) {
    event.preventDefault()
    highlightedIndex.value = lastEnabledIndex()
    scrollHighlightedIntoView()
  }
}

function onMenuKeydown(event: KeyboardEvent) {
  onTriggerKeydown(event)
}

function onPointerDown(event: PointerEvent) {
  const target = event.target as Node | null
  if (!target) return
  if (triggerRef.value?.contains(target) || menuRef.value?.contains(target)) return
  closeMenu()
}

function onFocusIn(event: FocusEvent) {
  const target = event.target as Node | null
  if (!open.value || !target) return
  if (triggerRef.value?.contains(target) || menuRef.value?.contains(target)) return
  closeMenu()
}

watch(
  () => props.modelValue,
  () => {
    if (open.value && selectedIndex.value >= 0) {
      highlightedIndex.value = selectedIndex.value
    }
  },
)

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) closeMenu()
  },
)

if (typeof document !== 'undefined') {
  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('focusin', onFocusIn)
}

onBeforeUnmount(() => {
  removeViewportListeners()
  if (typeof document !== 'undefined') {
    document.removeEventListener('pointerdown', onPointerDown)
    document.removeEventListener('focusin', onFocusIn)
  }
})
</script>

<template>
  <div
    class="ui-select"
    :class="{ 'is-open': open, 'is-disabled': disabled, 'is-fluid': fluid }"
    :style="{ minWidth: fluid ? undefined : minWidth }"
  >
    <button
      :id="`${uid}-trigger`"
      ref="triggerRef"
      class="ui-select__trigger"
      type="button"
      :disabled="disabled"
      :aria-label="ariaLabel || triggerLabel"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-controls="`${uid}-menu`"
      @click="toggleMenu"
      @keydown="onTriggerKeydown"
    >
      <span class="ui-select__value" :class="{ 'is-placeholder': !selectedOption }">
        {{ triggerLabel }}
      </span>
      <span class="ui-select__chevron" aria-hidden="true">
        <svg viewBox="0 0 16 16">
          <path d="m4 6 4 4 4-4" />
        </svg>
      </span>
    </button>

    <Teleport to="body">
      <Transition :name="placement === 'top' ? 'ui-select-popover-top' : 'ui-select-popover'">
        <div
          v-if="open"
          :id="`${uid}-menu`"
          ref="menuRef"
          class="ui-select__menu"
          :class="{ 'opens-top': placement === 'top' }"
          :style="menuStyle"
          role="listbox"
          :aria-labelledby="`${uid}-trigger`"
          tabindex="-1"
          @keydown="onMenuKeydown"
        >
          <button
            v-for="(option, index) in options"
            :key="`${String(option.value)}-${index}`"
            type="button"
            class="ui-select__option"
            :class="{
              'is-selected': Object.is(modelValue, option.value),
              'is-highlighted': highlightedIndex === index,
              'is-disabled': option.disabled,
            }"
            role="option"
            :aria-selected="Object.is(modelValue, option.value)"
            :disabled="option.disabled"
            :data-option-index="index"
            @mouseenter="highlightedIndex = option.disabled ? highlightedIndex : index"
            @click="selectOption(index)"
          >
            <span class="ui-select__option-copy">
              <strong>{{ option.label }}</strong>
              <small v-if="option.hint">{{ option.hint }}</small>
            </span>
            <span class="ui-select__check" aria-hidden="true">
              <svg viewBox="0 0 16 16"><path d="m3.6 8.2 2.8 2.8 6-6" /></svg>
            </span>
          </button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.ui-select {
  --ui-select-height: 40px;
  --ui-select-radius: 9px;
  --ui-select-bg: #0b0d11;
  --ui-select-bg-hover: #101319;
  --ui-select-bg-open: #10151b;
  --ui-select-border: #2a2f37;
  --ui-select-border-hover: #39414b;
  --ui-select-border-open: #477694;
  --ui-select-text: #cbd1d8;
  --ui-select-muted: #7a8591;
  --ui-select-ring: rgba(71, 143, 193, .10);
  --ui-select-menu-bg: rgba(13, 15, 19, .985);
  --ui-select-menu-border: #2b3038;
  --ui-select-menu-shadow: 0 22px 52px rgba(0, 0, 0, .42), 0 5px 16px rgba(0, 0, 0, .22), inset 0 1px rgba(255, 255, 255, .025);
  --ui-select-option: #b9c1ca;
  --ui-select-option-hover: #171b21;
  --ui-select-option-selected: #132331;
  --ui-select-option-selected-hover: #182a39;
  --ui-select-check: #67bff2;

  position: relative;
  display: inline-block;
  vertical-align: middle;
  max-width: 100%;
}

.ui-select.is-fluid {
  display: block;
  width: 100%;
}

.ui-select__trigger {
  width: 100%;
  height: var(--ui-select-height);
  min-height: var(--ui-select-height);
  padding: 0 11px 0 13px;
  border: 1px solid var(--ui-select-border);
  border-radius: var(--ui-select-radius);
  background: var(--ui-select-bg);
  color: var(--ui-select-text);
  box-shadow: inset 0 1px rgba(255, 255, 255, .015);
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 10px;
  text-align: left;
  outline: none;
  cursor: pointer;
  transition:
    border-color .16s ease,
    background-color .16s ease,
    box-shadow .16s ease,
    transform .16s ease;
}

.ui-select__trigger:hover:not(:disabled) {
  border-color: var(--ui-select-border-hover);
  background: var(--ui-select-bg-hover);
}

.ui-select__trigger:focus-visible,
.is-open .ui-select__trigger {
  border-color: var(--ui-select-border-open);
  background: var(--ui-select-bg-open);
  box-shadow: 0 0 0 3px var(--ui-select-ring), inset 0 1px rgba(255, 255, 255, .02);
}

.ui-select__value {
  min-width: 0;
  height: 100%;
  display: flex;
  align-items: center;
  overflow: hidden;
  color: var(--ui-select-text);
  font-size: .78rem;
  font-weight: 580;
  line-height: 1;
  letter-spacing: -.006em;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ui-select__value.is-placeholder {
  color: var(--ui-select-muted);
}

.ui-select__chevron {
  width: 18px;
  height: 18px;
  display: grid;
  place-items: center;
  color: var(--ui-select-muted);
  transform-origin: 50% 50%;
  transition: transform .2s cubic-bezier(.2, .78, .2, 1), color .16s ease;
}

.ui-select__chevron svg {
  width: 14px;
  height: 14px;
  display: block;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.55;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ui-select__trigger:hover .ui-select__chevron,
.is-open .ui-select__chevron {
  color: #aab3bd;
}

.is-open .ui-select__chevron {
  transform: rotate(180deg);
}

.is-disabled {
  opacity: .52;
}

.is-disabled .ui-select__trigger {
  cursor: not-allowed;
}

.ui-select__menu {
  --ui-select-menu-bg: rgba(13, 15, 19, .985);
  --ui-select-menu-border: #2b3038;
  --ui-select-menu-shadow: 0 22px 52px rgba(0, 0, 0, .42), 0 5px 16px rgba(0, 0, 0, .22), inset 0 1px rgba(255, 255, 255, .025);
  --ui-select-option: #b9c1ca;
  --ui-select-option-hover: #171b21;
  --ui-select-option-selected: #132331;
  --ui-select-option-selected-hover: #182a39;
  --ui-select-check: #67bff2;

  position: fixed;
  z-index: 1200;
  padding: 6px;
  overflow: auto;
  overscroll-behavior: contain;
  border: 1px solid var(--ui-select-menu-border);
  border-radius: 12px;
  background: var(--ui-select-menu-bg);
  color: var(--ui-select-option);
  box-shadow: var(--ui-select-menu-shadow);
  backdrop-filter: blur(18px) saturate(120%);
  -webkit-backdrop-filter: blur(18px) saturate(120%);
  scrollbar-width: thin;
  scrollbar-color: #353c46 transparent;
  transform-origin: top center;
}

.ui-select__menu.opens-top {
  transform-origin: bottom center;
}

.ui-select__option {
  width: 100%;
  min-height: 42px;
  margin: 1px 0;
  padding: 7px 9px 7px 11px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--ui-select-option);
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 10px;
  text-align: left;
  cursor: pointer;
  outline: none;
  transition: background-color .12s ease, color .12s ease, transform .12s ease;
}

.ui-select__option:hover:not(:disabled),
.ui-select__option.is-highlighted:not(:disabled) {
  background: var(--ui-select-option-hover);
  color: #f1f4f6;
}

.ui-select__option.is-selected {
  background: var(--ui-select-option-selected);
  color: #eff8fd;
}

.ui-select__option.is-selected:hover,
.ui-select__option.is-selected.is-highlighted {
  background: var(--ui-select-option-selected-hover);
}

.ui-select__option:active:not(:disabled) {
  transform: scale(.995);
}

.ui-select__option.is-disabled {
  opacity: .42;
  cursor: not-allowed;
}

.ui-select__option-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
}

.ui-select__option-copy strong {
  min-width: 0;
  overflow: hidden;
  font-size: .76rem;
  font-weight: 590;
  line-height: 1.2;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ui-select__option-copy small {
  overflow: hidden;
  color: #69737d;
  font-size: .64rem;
  line-height: 1.2;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ui-select__check {
  width: 18px;
  height: 18px;
  display: grid;
  place-items: center;
  color: var(--ui-select-check);
  opacity: 0;
  transform: scale(.84);
  transition: opacity .14s ease, transform .16s cubic-bezier(.2, .78, .2, 1);
}

.ui-select__check svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ui-select__option.is-selected .ui-select__check {
  opacity: 1;
  transform: scale(1);
}

.ui-select__menu::-webkit-scrollbar {
  width: 6px;
}

.ui-select__menu::-webkit-scrollbar-track {
  background: transparent;
}

.ui-select__menu::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: #353c46;
}

.ui-select-popover-enter-active,
.ui-select-popover-leave-active,
.ui-select-popover-top-enter-active,
.ui-select-popover-top-leave-active {
  transition:
    opacity .15s ease,
    transform .18s cubic-bezier(.2, .78, .2, 1),
    filter .18s ease;
}

.ui-select-popover-enter-from,
.ui-select-popover-leave-to {
  opacity: 0;
  transform: translateY(-5px) scale(.985);
  filter: blur(1px);
}

.ui-select-popover-top-enter-from,
.ui-select-popover-top-leave-to {
  opacity: 0;
  transform: translateY(5px) scale(.985);
  filter: blur(1px);
}

:global(html.smirel-app[data-theme='light']) .ui-select {
  --ui-select-bg: #ffffff;
  --ui-select-bg-hover: #f8fafc;
  --ui-select-bg-open: #ffffff;
  --ui-select-border: #d8e0e7;
  --ui-select-border-hover: #becbd6;
  --ui-select-border-open: #7fb1d2;
  --ui-select-text: #33414e;
  --ui-select-muted: #7b8996;
  --ui-select-ring: rgba(54, 135, 192, .10);
  --ui-select-menu-bg: rgba(255, 255, 255, .985);
  --ui-select-menu-border: #d8e1e8;
  --ui-select-menu-shadow: 0 20px 46px rgba(29, 45, 59, .14), 0 4px 12px rgba(29, 45, 59, .06), inset 0 1px rgba(255, 255, 255, .95);
  --ui-select-option: #52606d;
  --ui-select-option-hover: #f4f7fa;
  --ui-select-option-selected: #eaf4fb;
  --ui-select-option-selected-hover: #e1eff8;
  --ui-select-check: #2f91cf;
}

:global(html.smirel-app[data-theme='light']) .ui-select__menu {
  --ui-select-menu-bg: rgba(255, 255, 255, .985);
  --ui-select-menu-border: #d8e1e8;
  --ui-select-menu-shadow: 0 20px 46px rgba(29, 45, 59, .14), 0 4px 12px rgba(29, 45, 59, .06), inset 0 1px rgba(255, 255, 255, .95);
  --ui-select-option: #52606d;
  --ui-select-option-hover: #f4f7fa;
  --ui-select-option-selected: #eaf4fb;
  --ui-select-option-selected-hover: #e1eff8;
  --ui-select-check: #2f91cf;
  scrollbar-color: #c3ced7 transparent;
}

:global(html.smirel-app[data-theme='light']) .ui-select__menu::-webkit-scrollbar-thumb {
  background: #c3ced7;
}

:global(html.smirel-app[data-theme='light']) .ui-select__trigger:hover .ui-select__chevron,
:global(html.smirel-app[data-theme='light']) .ui-select.is-open .ui-select__chevron {
  color: #536b7e;
}

:global(html.smirel-app[data-theme='light']) .ui-select__option:hover:not(:disabled),
:global(html.smirel-app[data-theme='light']) .ui-select__option.is-highlighted:not(:disabled) {
  color: #202b35;
}

:global(html.smirel-app[data-theme='light']) .ui-select__option.is-selected {
  color: #236f9f;
}

@media (prefers-reduced-motion: reduce) {
  .ui-select__trigger,
  .ui-select__chevron,
  .ui-select__option,
  .ui-select__check,
  .ui-select-popover-enter-active,
  .ui-select-popover-leave-active,
  .ui-select-popover-top-enter-active,
  .ui-select-popover-top-leave-active {
    transition: none !important;
  }
}
</style>
