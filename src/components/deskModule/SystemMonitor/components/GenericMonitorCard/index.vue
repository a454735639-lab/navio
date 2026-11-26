<script setup lang="ts">
// -------------------
// 系统监控图标临时使用（与 AppIcon/index.vue 一致）
// 如果确定这种方案将 AppIcon/index.vue 封装成通用组件
// -------------------

import { ref } from 'vue'
import { ItemCard } from '@/components/common'
import type { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
//   size?: number // 默认70
  extendParam?: any
  iconTextColor?: string
  iconTextIconHideTitle?: boolean
  iconText?: string
  cardTypeStyle: PanelPanelConfigStyleEnum
  class?: string
  backgroundColor?: string
}

const props = withDefaults(defineProps<Prop>(), {})
const propClass = ref(props.class)
</script>

<template>
  <div class="generic-monitor-card w-full">
    <ItemCard
      :card-type-style="cardTypeStyle"
      :icon-text="iconText"
      :icon-text-color="iconTextColor"
      :class="propClass"
      :background-color="backgroundColor"
      :icon-text-icon-hide-title="iconTextIconHideTitle"
    >
      <template #info>
        <!-- info 模式：环形仪表居中，不预留左侧图标空间（1Panel 环形瓦片以环形为视觉主体） -->
        <div class="w-full flex items-center justify-center">
          <slot name="info" />
        </div>
      </template>
      <template #small>
        <slot name="small" />
      </template>
    </ItemCard>
  </div>
</template>
