<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  duration?: number
  disabled?: boolean
  loading?: boolean
}>(), {
  duration: 1200,
  disabled: false,
  loading: false,
})
const emit = defineEmits<{ (e: 'confirm'): void }>()

const RADIUS = 13
const CIRCUMFERENCE = 2 * Math.PI * RADIUS
const progress = ref(0)
const holding = ref(false)
const done = ref(false)
let raf = 0
let startedAt = 0

const size = 30
const dashOffset = computed(() => CIRCUMFERENCE * (1 - progress.value))
function reducedMotion() {
  return typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
}

function frame(now: number) {
  if (!holding.value)
    return
  progress.value = Math.min(1, (now - startedAt) / props.duration)
  if (progress.value >= 1) {
    holding.value = false
    done.value = true
    emit('confirm')
    // brief success state then reset
    window.setTimeout(reset, reducedMotion() ? 0 : 500)
    return
  }
  raf = requestAnimationFrame(frame)
}

function start(e: PointerEvent) {
  if (props.disabled || props.loading || holding.value)
    return
  e.preventDefault()
  ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
  holding.value = true
  progress.value = 0
  startedAt = performance.now()
  raf = requestAnimationFrame(frame)
}

function cancel() {
  if (!holding.value)
    return
  holding.value = false
  cancelAnimationFrame(raf)
  // quick rewind
  const from = progress.value
  if (reducedMotion() || from <= 0) { progress.value = 0; return }
  const t0 = performance.now()
  const step = (now: number) => {
    const k = Math.min(1, (now - t0) / 220)
    progress.value = from * (1 - k)
    if (k < 1) raf = requestAnimationFrame(step)
    else progress.value = 0
  }
  raf = requestAnimationFrame(step)
}

function reset() {
  cancelAnimationFrame(raf)
  holding.value = false
  done.value = false
  progress.value = 0
}

function activate() {
  // keyboard: one-shot confirm via Enter/Space
  if (props.disabled || props.loading)
    return
  done.value = true
  emit('confirm')
  window.setTimeout(reset, reducedMotion() ? 0 : 500)
}

onUnmounted(() => cancelAnimationFrame(raf))
</script>

<template>
  <button
    type="button"
    class="hold-confirm"
    :class="{ 'is-holding': holding, 'is-done': done, 'is-disabled': disabled || loading }"
    :disabled="disabled || loading"
    role="button"
    tabindex="0"
    :aria-label="$attrs['aria-label']"
    @pointerdown="start"
    @pointerup="cancel"
    @pointercancel="cancel"
    @pointerleave="cancel"
    @keydown.enter.prevent="activate"
    @keydown.space.prevent="activate"
    @keyup.enter="reset"
  >
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`" class="hold-ring" aria-hidden="true">
      <circle :cx="size / 2" :cy="size / 2" :r="RADIUS" class="hold-track" />
      <circle
        :cx="size / 2" :cy="size / 2" :r="RADIUS" class="hold-fill"
        :stroke-dasharray="CIRCUMFERENCE" :stroke-dashoffset="dashOffset"
      />
      <path v-if="done" d="M9 15l3 3 6-7" class="hold-check" />
    </svg>
    <span class="hold-label"><slot /></span>
  </button>
</template>

<style scoped>
.hold-confirm {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  position: relative;
  border: 1px solid rgba(250, 173, 20, 0.45);
  background: rgba(250, 173, 20, 0.08);
  color: #f0a020;
  border-radius: 6px;
  padding: 2px 7px 2px 4px;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
  touch-action: none;
  transition: transform 160ms ease, background-color 160ms ease;
  -webkit-tap-highlight-color: transparent;
}
.hold-confirm:hover { background: rgba(250, 173, 20, 0.16); }
.hold-confirm.is-holding { transform: scale(0.96); }
.hold-confirm.is-done { border-color: #18a058; color: #18a058; background: rgba(24, 160, 88, 0.12); }
.hold-confirm.is-disabled { opacity: 0.5; cursor: not-allowed; }
.hold-confirm:focus-visible { outline: 2px solid #f0a020; outline-offset: 2px; }
.hold-ring { flex: none; transform: rotate(-90deg); }
.hold-track { fill: none; stroke: rgba(255, 255, 255, 0.15); stroke-width: 2.5; }
.hold-fill { fill: none; stroke: currentColor; stroke-width: 2.5; stroke-linecap: round; }
.hold-check { fill: none; stroke: #18a058; stroke-width: 2.2; stroke-linecap: round; stroke-linejoin: round; }
.hold-label { white-space: nowrap; }
@media (prefers-reduced-motion: reduce) {
  .hold-confirm { transition: none; }
}
</style>
