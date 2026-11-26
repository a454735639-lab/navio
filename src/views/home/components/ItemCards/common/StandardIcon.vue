<script setup lang="ts">
import { computed, inject, onMounted, ref, watch } from 'vue'
import { NEllipsis } from 'naive-ui'
import type { PanelPanelConfigStyleEnum } from '@/enums'
import { ItemCard, ItemIcon } from '@/components/common'
import { defaultItemCardBackgroundColor } from '@/utils/defaultData'

interface Prop {
  itemInfo: Panel.ItemInfo
  backgroundColor?: string
  iconTextColor?: string
  iconTextInfoHideDescription: boolean
  iconTextIconHideTitle: boolean
  cardStyle: PanelPanelConfigStyleEnum
  sortStatus: boolean
}
const props = defineProps<Prop>()

// 卡片在线状态由首页统一探测后注入
const aliveMap = inject<import('vue').Ref<Map<number, boolean>> | null>('aliveMap', null)
const alive = computed<boolean | undefined>(() => {
  const id = props.itemInfo?.id
  if (!aliveMap || id === undefined)
    return undefined
  return aliveMap.value.get(id)
})

const defaultBackground = defaultItemCardBackgroundColor
const backgroundRef = computed(() => props.backgroundColor || defaultBackground)

const calculateLuminance = (color: string) => {
  const hex = color.replace(/^#/, '')
  const r = parseInt(hex.substring(0, 2), 16)
  const g = parseInt(hex.substring(2, 4), 16)
  const b = parseInt(hex.substring(4, 6), 16)
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255
}

const textColor = ref('')

function autoColor() {
  const luminance = calculateLuminance(backgroundRef.value || '')
  return luminance > 0.5 ? 'black' : 'white'
}

watch(() => backgroundRef.value, () => {
  textColor.value = autoColor()
})

onMounted(() => {
  textColor.value = autoColor()
})
</script>

<template>
  <div class="app-icon w-full">
    <ItemCard
      :card-type-style="cardStyle"
      :background-color="backgroundColor"
      :icon-text-icon-hide-title="iconTextIconHideTitle"
      :icon-text-color="iconTextColor"
      :icon-text="itemInfo?.title"
    >
      <template #small>
        <div
          class="app-icon-small-icon relative overflow-hidden w-[70px] h-[70px] mx-auto rounded-xl"
          :title="itemInfo?.description"
        >
          <ItemIcon :item-icon="itemInfo?.icon" :color="textColor" />
          <span v-if="alive !== undefined" class="alive-dot" :class="alive ? 'is-up' : 'is-down'" />
        </div>
        <div
          v-if="!iconTextIconHideTitle"
          class="app-icon-small-title text-center app-icon-text-shadow cursor-pointer mt-[2px]"
          :style="{ color: iconTextColor }"
        >
          <span>{{ itemInfo?.title }}</span>
        </div>
      </template>

      <template #info>
        <!-- 图标 -->
        <div class="app-icon-info-icon w-[70px] h-[70px]">
          <div class="w-[70px] h-full flex items-center justify-center">
            <ItemIcon :item-icon="itemInfo?.icon" :size="50" class="overflow-hidden rounded-xl" :color="textColor" />
          </div>
        </div>

        <!-- 文字 -->
        <!-- 如果为纯白色，将自动根据背景的明暗计算字体的黑白色 -->
        <div class="text-white flex items-center" :style="{ color: (!iconTextColor || iconTextColor === '#ffffff') ? textColor : iconTextColor, maxWidth: 'calc(100% - 80px)' }">
          <div class="app-icon-info-text-box w-full">
            <div class="app-icon-info-text-box-title font-semibold w-full flex items-center gap-1.5">
              <span v-if="alive !== undefined" class="alive-dot-inline" :class="alive ? 'is-up' : 'is-down'" />
              <NEllipsis>
                {{ itemInfo?.title }}
              </NEllipsis>
            </div>
            <div v-if="!iconTextInfoHideDescription" class="app-icon-info-text-box-description">
              <NEllipsis :line-clamp="2" class="text-xs">
                {{ itemInfo?.description }}
              </NEllipsis>
            </div>
          </div>
        </div>
      </template>
    </ItemCard>
  </div>
</template>

<style scoped>
/* Pass details: 在线状态点 —— 右上角小圆点，绿=可达，灰红=不可达 */
.alive-dot {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.25);
  pointer-events: none;
}
.alive-dot.is-up {
  background: #22c55e;
}
.alive-dot.is-down {
  background: #ef4444;
  opacity: 0.75;
}

.alive-dot-inline {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}
.alive-dot-inline.is-up {
  background: #22c55e;
}
.alive-dot-inline.is-down {
  background: #ef4444;
  opacity: 0.7;
}
</style>
