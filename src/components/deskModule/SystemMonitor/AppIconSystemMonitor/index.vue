<script setup lang="ts">
import { computed } from 'vue'
import { MonitorType } from '../typings'
import type { CardStyle } from '../typings'
import GenericMonitorCard from '../components/GenericMonitorCard/index.vue'
import CardCPU from './CPU.vue'
import Memory from './Memory.vue'
import Disk from './Disk.vue'
import Network from './Network.vue'
import { SvgIcon } from '@/components/common'
import { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
  extendParam?: any
  iconTextColor?: string
  iconTextIconHideTitle: boolean
  cardTypeStyle: PanelPanelConfigStyleEnum
  monitorType: string
  cardStyle: CardStyle
}

const props = withDefaults(defineProps<Prop>(), {})

const iconText = computed(() => {
  switch (props.monitorType) {
    case MonitorType.cpu:
      return 'CPU'
    case MonitorType.disk:
      return props.extendParam.path
    case MonitorType.memory:
      return 'RAM'
    case MonitorType.network:
      return 'NET'
  }
  return ''
})
const refreshInterval = 5000

// 1Panel 浅色卡片风格：白底 + 蓝色环 + 深灰字 + 浅灰轨道
const LIGHT_BG = '#ffffff'
const LIGHT_TEXT = '#2b3245'
const LIGHT_PROGRESS = '#3b82f6'
const LIGHT_RAIL = '#e8edf5'

// 阈值配色（兼容旧配置未设置警告色的情况）
const warningThreshold = computed(() => props.extendParam?.warningThreshold ?? 0)
const warningColor = computed(() => props.extendParam?.warningColor || '#ef4444')

</script>

<template>
  <div class="w-full">
    <GenericMonitorCard
      :icon-text-icon-hide-title="false"
      :card-type-style="cardTypeStyle"
      :icon-text="extendParam?.title || iconText"
      :icon-text-color="iconTextColor"
      :background-color="LIGHT_BG"
    >
      <template #icon>
        <!-- 图标 -->
        <div class="w-[60px] h-[70px]">
          <div class="app-icon w-[60px] h-full flex items-center justify-center text-white">
            <SvgIcon v-if="monitorType === MonitorType.cpu" icon="solar-cpu-bold" :style="{ color: extendParam.color }" style="width:35px;height:35px" />
            <SvgIcon v-if="monitorType === MonitorType.memory" icon="material-symbols-memory-alt-rounded" :style="{ color: extendParam.color }" style="width:35px;height:35px" />
            <SvgIcon v-if="monitorType === MonitorType.disk" icon="clarity-hard-disk-solid" :style="{ color: extendParam.color }" style="width:35px;height:35px" />
            <SvgIcon v-if="monitorType === MonitorType.network" icon="solar-transfer-horizontal-bold" :style="{ color: extendParam.color }" style="width:35px;height:35px" />
          </div>
        </div>
      </template>
      <template #info>
        <!-- 如果为纯白色，将自动根据背景的明暗计算字体的黑白色 -->
        <div
          class=" w-full text-white flex items-center"
        >
          <CardCPU
            v-if="monitorType === MonitorType.cpu"
            :card-type-style="PanelPanelConfigStyleEnum.info"
            :progress-color="LIGHT_PROGRESS"
            :progress-rail-color="LIGHT_RAIL"
            :text-color="LIGHT_TEXT"
            :refresh-interval="refreshInterval"
            :warning-threshold="warningThreshold"
            :warning-color="warningColor"
          />
          <Memory
            v-else-if="monitorType === MonitorType.memory"
            :card-type-style="PanelPanelConfigStyleEnum.info"
            :progress-color="LIGHT_PROGRESS"
            :progress-rail-color="LIGHT_RAIL"
            :text-color="LIGHT_TEXT"
            :refresh-interval="refreshInterval"
            :warning-threshold="warningThreshold"
            :warning-color="warningColor"
          />
          <Disk
            v-else-if="monitorType === MonitorType.disk"
            :card-type-style="PanelPanelConfigStyleEnum.info"
            :progress-color="LIGHT_PROGRESS"
            :progress-rail-color="LIGHT_RAIL"
            :text-color="LIGHT_TEXT"
            :path="extendParam?.path"
            :title="extendParam?.title"
            :refresh-interval="refreshInterval"
            :warning-threshold="warningThreshold"
            :warning-color="warningColor"
          />
          <Network
            v-else-if="monitorType === MonitorType.network"
            :card-type-style="PanelPanelConfigStyleEnum.info"
            :progress-color="LIGHT_PROGRESS"
            :progress-rail-color="LIGHT_RAIL"
            :text-color="LIGHT_TEXT"
            :refresh-interval="refreshInterval"
            :warning-threshold="warningThreshold"
            :warning-color="warningColor"
          />
        </div>
      </template>
      <template #small>
        <CardCPU
          v-if="monitorType === MonitorType.cpu"
          :card-type-style="PanelPanelConfigStyleEnum.icon"
          :progress-color="LIGHT_PROGRESS"
          :progress-rail-color="LIGHT_RAIL"
          :text-color="LIGHT_TEXT"
          :refresh-interval="refreshInterval"
          :warning-threshold="warningThreshold"
          :warning-color="warningColor"
        />
        <Memory
          v-else-if="monitorType === MonitorType.memory"
          :card-type-style="PanelPanelConfigStyleEnum.icon"
          :progress-color="LIGHT_PROGRESS"
          :progress-rail-color="LIGHT_RAIL"
          :text-color="LIGHT_TEXT"
          :refresh-interval="refreshInterval"
          :warning-threshold="warningThreshold"
          :warning-color="warningColor"
        />
        <Disk
          v-else-if="monitorType === MonitorType.disk"
          :card-type-style="PanelPanelConfigStyleEnum.icon"
          :progress-color="LIGHT_PROGRESS"
          :progress-rail-color="LIGHT_RAIL"
          :text-color="LIGHT_TEXT"
          :path="extendParam?.path"
          :title="extendParam?.title"
          :refresh-interval="refreshInterval"
          :warning-threshold="warningThreshold"
          :warning-color="warningColor"
        />
        <Network
          v-else-if="monitorType === MonitorType.network"
          :card-type-style="PanelPanelConfigStyleEnum.icon"
          :progress-color="LIGHT_PROGRESS"
          :progress-rail-color="LIGHT_RAIL"
          :text-color="LIGHT_TEXT"
          :refresh-interval="refreshInterval"
          :warning-threshold="warningThreshold"
          :warning-color="warningColor"
        />
      </template>
    </GenericMonitorCard>
  </div>
</template>
