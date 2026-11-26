<script lang="ts">
// 模块级计数器：保证多实例渐变 id 唯一（script setup 里的变量是每实例独立的）
let uidCounter = 0
</script>

<script setup lang="ts">
import { computed } from 'vue'

interface Prop {
  data: number[]
  color: string
  height?: number
  /** Y 轴最大值；不传则用窗口内最大值（至少 1，避免除零） */
  max?: number
}

const props = withDefaults(defineProps<Prop>(), {
  height: 26,
  max: undefined,
})

// viewBox 固定宽度，靠 preserveAspectRatio="none" 横向拉伸填满容器
const W = 100
const H = props.height

// 渐变 id 需页面唯一（同色多实例会冲突）
const gradId = `sparkline-grad-${++uidCounter}`

const maxVal = computed(() => {
  if (props.max !== undefined)
    return props.max
  const m = Math.max(...props.data, 0)
  return m > 0 ? m : 1
})

const linePath = computed(() => {
  const n = props.data.length
  if (n < 2)
    return ''
  const pts = props.data.map((v, i) => {
    const x = (i / (n - 1)) * W
    const y = H - (v / maxVal.value) * H
    return `${x.toFixed(2)},${y.toFixed(2)}`
  })
  return `M${pts.join(' L')}`
})

const areaPath = computed(() => {
  const lp = linePath.value
  if (!lp)
    return ''
  return `${lp} L${W},${H} L0,${H} Z`
})
</script>

<template>
  <svg
    :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="none"
    :style="{ width: '100%', height: `${H}px`, display: 'block' }"
  >
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.32" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path v-if="areaPath" :d="areaPath" :fill="`url(#${gradId})`" />
    <path
      v-if="linePath" :d="linePath" fill="none" :stroke="color" stroke-width="1.5"
      stroke-linecap="round" stroke-linejoin="round" vector-effect="non-scaling-stroke"
    />
  </svg>
</template>
