<script setup lang="ts">
import { computed } from 'vue'
import { NEllipsis } from 'naive-ui'
import { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
  textColor: string
  progressColor: string
  progressRailColor: string
  progressHeight?: number
  percentage: number
  cardTypeStyle: PanelPanelConfigStyleEnum
  infoCardLeftText?: string
  infoCardRightText?: string
}

const props = defineProps<Prop>()

// 1Panel 风格环形仪表：圆头进度弧 + 中心大号百分比 + 下方标题/数值
const R = 42
const CIRC = 2 * Math.PI * R

const dashOffset = computed(() => {
  const p = Math.max(0, Math.min(props.percentage, 100))
  return CIRC * (1 - p / 100)
})

const displayInt = computed(() => {
  const v = props.percentage
  // 10% 以下保留一位小数，其余取整，避免 <10% 时看起来像 0
  return v > 0 && v < 10 ? v.toFixed(1) : Math.round(v).toString()
})
</script>

<template>
  <template v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info">
    <div class="flex flex-col items-center justify-center w-full py-1">
      <div class="relative" style="width: 74px; height: 74px;">
        <svg viewBox="0 0 100 100" style="width: 100%; height: 100%;">
          <circle cx="50" cy="50" :r="R" fill="none" :stroke="progressRailColor" stroke-width="8" stroke-opacity="0.3" />
          <circle
            cx="50" cy="50" :r="R" fill="none" :stroke="progressColor" stroke-width="8"
            stroke-linecap="round" :stroke-dasharray="CIRC" :stroke-dashoffset="dashOffset"
            transform="rotate(-90 50 50)"
          />
        </svg>
        <div class="absolute inset-0 flex items-center justify-center" :style="{ color: textColor }">
          <span style="font-size: 19px; font-weight: 600; line-height: 1;">{{ displayInt }}</span>
          <span style="font-size: 11px; margin-left: 1px; align-self: flex-start; margin-top: 3px;">%</span>
        </div>
      </div>
      <div class="mt-2 text-sm font-medium text-center w-full px-1" :style="{ color: textColor }">
        <NEllipsis>{{ infoCardLeftText }}</NEllipsis>
      </div>
      <div class="text-xs opacity-60 text-center w-full px-1 mt-[2px]" :style="{ color: textColor }">
        <NEllipsis>{{ infoCardRightText }}</NEllipsis>
      </div>
    </div>
  </template>
  <template v-else>
    <div class="w-full flex justify-center h-full mt-3">
      <div class="relative" style="width: 50px; height: 50px;">
        <svg viewBox="0 0 100 100" style="width: 100%; height: 100%;">
          <circle cx="50" cy="50" :r="R" fill="none" :stroke="progressRailColor" stroke-width="10" stroke-opacity="0.3" />
          <circle
            cx="50" cy="50" :r="R" fill="none" :stroke="progressColor" stroke-width="10"
            stroke-linecap="round" :stroke-dasharray="CIRC" :stroke-dashoffset="dashOffset"
            transform="rotate(-90 50 50)"
          />
        </svg>
        <div class="absolute inset-0 flex items-center justify-center" style="font-size: 11px;" :style="{ color: textColor }">
          {{ Math.round(percentage) }}%
        </div>
      </div>
    </div>
  </template>
</template>
