<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { NPopover } from 'naive-ui'
import GenericProgress from '../components/GenericProgress/index.vue'
import Sparkline from '../components/Sparkline/index.vue'
import { correctionNumber, correctionNumberByCardStyle, sizePairText } from './common'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { bytesToSize } from '@/utils/cmn'
import { getDiskStateByPath } from '@/api/system/systemMonitor'

interface Prop {
  cardTypeStyle: PanelPanelConfigStyleEnum
  refreshInterval: number
  textColor: string
  progressColor: string
  progressRailColor: string
  path: string
  title: string
}

const props = defineProps<Prop>()
let timer: any // NodeJS.Timer
const diskState = ref<SystemMonitor.DiskInfo | null>(null)
const diskHistory = ref<number[]>([])

function formatdiskSize(v: number): string {
  return bytesToSize(v)
}

function formatdiskToByte(v: number): number {
  return v
}

async function getData() {
  try {
    const { data, code } = await getDiskStateByPath<SystemMonitor.DiskInfo>(props.path)
    if (code === 0) {
      diskState.value = data
      diskHistory.value.push(data.usedPercent)
      if (diskHistory.value.length > 30)
        diskHistory.value.shift()
    }
  }
  catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  getData()
  timer = setInterval(() => {
    getData()
  }, (!props.refreshInterval || props.refreshInterval <= 2000) ? 2000 : props.refreshInterval)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<template>
  <NPopover trigger="click">
    <template #trigger>
      <div class="w-full">
        <GenericProgress
          :progress-color="progressColor" :progress-rail-color="progressRailColor" :progress-height="5"
          :percentage="correctionNumberByCardStyle(diskState?.usedPercent || 0, cardTypeStyle)"
          :card-type-style="cardTypeStyle"
          :info-card-right-text="sizePairText(diskState?.used || 0, diskState?.total || 0)"
          :info-card-left-text="title || diskState?.mountpoint" :text-color="textColor" class="cursor-pointer"
        />
        <Sparkline
          v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info" :data="diskHistory" :color="progressColor"
          :max="100" :height="26" class="mt-[6px]"
        />
      </div>
    </template>

    <div>
      <div class="overflow-auto max-h-[200px]">
        <div>
          <span> {{ $t('deskModule.systemMonitor.free') }}</span>
          <span class="ml-2">{{ formatdiskSize(formatdiskToByte(diskState?.free || 0)) }}</span>
          <span class="ml-2">({{ 100 - correctionNumber(diskState?.usedPercent || 0) }}%)</span>
        </div>
        <div>
          <span> {{ $t('deskModule.systemMonitor.used') }}</span>
          <span class="ml-2">{{ formatdiskSize(formatdiskToByte(diskState?.used || 0)) }}</span>
          <span class="ml-2">({{ correctionNumber(diskState?.usedPercent || 0) }}%)</span>
        </div>
        <div>
          <span>{{ $t('deskModule.systemMonitor.total') }}</span>
          <span class="ml-2">{{ formatdiskSize(formatdiskToByte(diskState?.total || 0)) }}</span>
        </div>
      </div>
    </div>
  </NPopover>
</template>
