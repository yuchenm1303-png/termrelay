<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

export interface SmirelSelectOption {
  value: string
  label: string
  hint?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: SmirelSelectOption[]
  label?: string
  ariaLabel?: string
  disabled?: boolean
}>(), {
  label: '',
  ariaLabel: '选择',
  disabled: false,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
  (event: 'change', value: string): void
}>()

const open = ref(false)
const highlightedIndex = ref(-1)
const trigger = ref<HTMLButtonElement | null>(null)
const menu = ref<HTMLElement | null>(null)
const menuStyle = ref<Record<string, string>>({})
const uid = `smirel-select-${Math.random().toString(36).slice(2, 9)}`

const selectedIndex = computed(() => props.options.findIndex((option) => option.value === props.modelValue))
const selectedOption = computed(() => props.options[selectedIndex.value] || props.options[0] || null)
const activeOptionId = computed(() => open.value && highlightedIndex.value >= 0
  ? `${uid}-option-${highlightedIndex.value}`
  : undefined)

function firstEnabledIndex() {
  return props.options.findIndex((option) => !option.disabled)
}

function lastEnabledIndex() {
  for (let index = props.options.length - 1; index >= 0; index -= 1) {
    if (!props.options[index]?.disabled) return index
  }
  return -1
}

function nextEnabledIndex(start: number, direction: 1 | -1) {
  if (!props.options.length) return -1
  let index = start
  for (let count = 0; count < props.options.length; count += 1) {
    index = (index + direction + props.options.length) % props.options.length
    if (!props.options[index]?.disabled) return index
  }
  return -1
}

function positionMenu() {
  if (!trigger.value) return
  const rect = trigger.value.getBoundingClientRect()
  const viewportPadding = 8
  const estimatedHeight = Math.min(286, Math.max(52, props.options.length * 42 + 12))
  const spaceBelow = window.innerHeight - rect.bottom
  const openAbove = spaceBelow < estimatedHeight + 12 && rect.top > spaceBelow
  const width = Math.min(Math.max(rect.width, 156), window.innerWidth - viewportPadding * 2)
  const left = Math.min(
    Math.max(viewportPadding, rect.left),
    Math.max(viewportPadding, window.innerWidth - width - viewportPadding),
  )

  menuStyle.value = {
    left: `${left}px`,
    width: `${width}px`,
    top: openAbove ? 'auto' : `${rect.bottom + 6}px`,
    bottom: openAbove ? `${window.innerHeight - rect.top + 6}px` : 'auto',
  }
}

function scrollHighlightedIntoView() {
  if (highlightedIndex.value < 0) return
  nextTick(() => document.getElementById(`${uid}-option-${highlightedIndex.value}`)?.scrollIntoView({ block: 'nearest' }))
}

async function openMenu(preferredIndex?: number) {
  if (props.disabled || !props.options.length) return
  open.value = true
  highlightedIndex.value = preferredIndex ?? (selectedIndex.value >= 0 ? selectedIndex.value : firstEnabledIndex())
  await nextTick()
  positionMenu()
  scrollHighlightedIntoView()
}

function closeMenu(focusTrigger = false) {
  open.value = false
  highlightedIndex.value = -1
  if (focusTrigger) nextTick(() => trigger.value?.focus())
}

function choose(index: number) {
  const option = props.options[index]
  if (!option || option.disabled) return
  emit('update:modelValue', option.value)
  emit('change', option.value)
  closeMenu(true)
}

function toggleMenu() {
  if (open.value) closeMenu()
  else void openMenu()
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (props.disabled) return

  if (!open.value) {
    if (event.key === 'ArrowDown' || event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      void openMenu()
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      void openMenu(lastEnabledIndex())
    }
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    highlightedIndex.value = nextEnabledIndex(highlightedIndex.value, 1)
    scrollHighlightedIntoView()
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    highlightedIndex.value = nextEnabledIndex(highlightedIndex.value, -1)
    scrollHighlightedIntoView()
  } else if (event.key === 'Home') {
    event.preventDefault()
    highlightedIndex.value = firstEnabledIndex()
    scrollHighlightedIntoView()
  } else if (event.key === 'End') {
    event.preventDefault()
    highlightedIndex.value = lastEnabledIndex()
    scrollHighlightedIntoView()
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    choose(highlightedIndex.value)
  } else if (event.key === 'Escape') {
    event.preventDefault()
    closeMenu(true)
  } else if (event.key === 'Tab') {
    closeMenu()
  }
}

function onDocumentPointerDown(event: PointerEvent) {
  const target = event.target as Node
  if (trigger.value?.contains(target) || menu.value?.contains(target)) return
  closeMenu()
}

function onViewportChange() {
  if (open.value) positionMenu()
}

watch(() => props.options, () => {
  if (open.value) nextTick(positionMenu)
}, { deep: true })

watch(() => props.disabled, (disabled) => {
  if (disabled) closeMenu()
})

onMounted(() => {
  document.addEventListener('pointerdown', onDocumentPointerDown, true)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown, true)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<template>
  <span class="smirel-select" :class="{ 'is-open': open, 'is-disabled': disabled }">
    <button
      ref="trigger"
      class="smirel-select-trigger"
      type="button"
      role="combobox"
      aria-haspopup="listbox"
      :aria-label="ariaLabel"
      :aria-expanded="open"
      :aria-controls="`${uid}-listbox`"
      :aria-activedescendant="activeOptionId"
      :disabled="disabled"
      @click="toggleMenu"
      @keydown="onTriggerKeydown"
    >
      <span v-if="label" class="smirel-select-label">{{ label }}</span>
      <span class="smirel-select-value">{{ selectedOption?.label || '—' }}</span>
      <svg class="smirel-select-chevron" viewBox="0 0 16 16" aria-hidden="true">
        <path d="m4 6 4 4 4-4" />
      </svg>
    </button>

    <Teleport to="body">
      <Transition name="smirel-select-menu">
        <div
          v-if="open"
          :id="`${uid}-listbox`"
          ref="menu"
          class="smirel-select-menu"
          role="listbox"
          :aria-label="ariaLabel"
          :style="menuStyle"
          @keydown="onTriggerKeydown"
        >
          <button
            v-for="(option, index) in options"
            :id="`${uid}-option-${index}`"
            :key="`${option.value}-${index}`"
            class="smirel-select-option"
            :class="{
              selected: option.value === modelValue,
              highlighted: index === highlightedIndex,
            }"
            type="button"
            role="option"
            :aria-selected="option.value === modelValue"
            :disabled="option.disabled"
            @pointerenter="highlightedIndex = index"
            @click="choose(index)"
          >
            <span class="smirel-select-option-copy">
              <strong>{{ option.label }}</strong>
              <small v-if="option.hint">{{ option.hint }}</small>
            </span>
            <svg v-if="option.value === modelValue" class="smirel-select-check" viewBox="0 0 16 16" aria-hidden="true">
              <path d="m3 8 3 3 7-7" />
            </svg>
          </button>
        </div>
      </Transition>
    </Teleport>
  </span>
</template>

<style scoped>
.smirel-select {
  position: relative;
  min-width: 0;
  display: inline-flex;
  vertical-align: middle;
}

.smirel-select-trigger {
  width: 100%;
  height: 40px;
  padding: 0 11px 0 12px;
  border: 1px solid #252a32;
  border-radius: 9px;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) 16px;
  align-items: center;
  gap: 8px;
  background: #0b0d11;
  color: #d9dde2;
  font: inherit;
  cursor: pointer;
  outline: none;
  box-shadow: inset 0 1px rgba(255, 255, 255, .012);
  transition: border-color .14s ease, background-color .14s ease, box-shadow .14s ease;
}

.smirel-select-trigger:hover:not(:disabled) {
  border-color: #343a44;
  background: #101318;
}

.smirel-select.is-open .smirel-select-trigger,
.smirel-select-trigger:focus-visible {
  border-color: #3d596d;
  background: #101318;
  box-shadow: 0 0 0 3px rgba(91, 188, 245, .07), inset 0 1px rgba(255, 255, 255, .018);
}

.smirel-select-trigger:active:not(:disabled) {
  background: #0d1014;
}

.smirel-select-trigger:disabled {
  opacity: .46;
  cursor: default;
}

.smirel-select-label {
  color: #77818c;
  font-size: .72rem;
  font-weight: 540;
  white-space: nowrap;
}

.smirel-select-value {
  min-width: 0;
  overflow: hidden;
  color: #e8ebef;
  font-size: .76rem;
  font-weight: 610;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smirel-select-chevron {
  width: 15px;
  height: 15px;
  color: #75808c;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: transform .16s ease, color .16s ease;
}

.smirel-select.is-open .smirel-select-chevron {
  color: #9ba6b2;
  transform: rotate(180deg);
}

.smirel-select-menu {
  position: fixed;
  z-index: 1000;
  max-height: 286px;
  padding: 5px;
  overflow: auto;
  border: 1px solid #2a3038;
  border-radius: 10px;
  background: #0d0f13;
  box-shadow: 0 18px 42px rgba(0, 0, 0, .38), inset 0 1px rgba(255, 255, 255, .02);
  scrollbar-width: thin;
  scrollbar-color: #353b45 transparent;
}

.smirel-select-menu::-webkit-scrollbar { width: 6px; }
.smirel-select-menu::-webkit-scrollbar-track { background: transparent; }
.smirel-select-menu::-webkit-scrollbar-thumb { border-radius: 99px; background: #353b45; }

.smirel-select-option {
  width: 100%;
  min-height: 40px;
  padding: 7px 9px 7px 10px;
  border: 0;
  border-radius: 7px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 16px;
  align-items: center;
  gap: 10px;
  background: transparent;
  color: #b9c0c8;
  font: inherit;
  text-align: left;
  cursor: pointer;
  outline: none;
  transition: background-color .1s ease, color .1s ease;
}

.smirel-select-option + .smirel-select-option { margin-top: 1px; }

.smirel-select-option:hover:not(:disabled),
.smirel-select-option.highlighted:not(:disabled) {
  background: #15191f;
  color: #f2f4f6;
}

.smirel-select-option.selected {
  background: #111c26;
  color: #eef6fb;
}

.smirel-select-option.selected.highlighted,
.smirel-select-option.selected:hover:not(:disabled) {
  background: #142331;
}

.smirel-select-option:focus-visible {
  box-shadow: inset 0 0 0 1px #3b566b;
}

.smirel-select-option:disabled {
  opacity: .38;
  cursor: default;
}

.smirel-select-option-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.smirel-select-option-copy strong {
  overflow: hidden;
  font-size: .76rem;
  line-height: 1.25;
  font-weight: 570;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smirel-select-option-copy small {
  overflow: hidden;
  color: #6f7984;
  font-size: .64rem;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.smirel-select-check {
  width: 15px;
  height: 15px;
  justify-self: end;
  color: #67bff2;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.smirel-select-menu-enter-active,
.smirel-select-menu-leave-active {
  transition: opacity .12s ease, transform .12s ease;
  transform-origin: top center;
}

.smirel-select-menu-enter-from,
.smirel-select-menu-leave-to {
  opacity: 0;
  transform: translateY(-3px) scale(.99);
}

@media (prefers-reduced-motion: reduce) {
  .smirel-select-trigger,
  .smirel-select-chevron,
  .smirel-select-option,
  .smirel-select-menu-enter-active,
  .smirel-select-menu-leave-active {
    transition: none;
  }
}
</style>
