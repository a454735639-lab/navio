<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { NPopover } from 'naive-ui'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { getNetworkState } from '@/api/system/systemMonitor'

interface Prop {
  cardTypeStyle: PanelPanelConfigStyleEnum
  refreshInterval: number
  textColor: string
  progressColor: string
  progressRailColor: string
}

const props = defineProps<Prop>()
let timer: any
const netState = ref<SystemMonitor.NetworkInfo | null>(null)
const speed = ref({ rx: 0, tx: 0 })
let lastRx = 0
let lastTx = 0
let lastTime = 0

function formatSpeed(bytesPerSec: number): string {
  if (bytesPerSec < 1024)
    return `${bytesPerSec.toFixed(0)} B/s`
  if (bytesPerSec < 1024 * 1024)
    return `${(bytesPerSec / 1024).toFixed(1)} KB/s`
  return `${(bytesPerSec / 1024 / 1024).toFixed(1)} MB/s`
}

function formatBytes(bytes: number): string {
  if (bytes < 1024 * 1024)
    return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

// 网速 = (当前累计 - 上次累计) / 时间差；前端保存上次值，后端无状态
function calcSpeed(curRx: number, curTx: number) {
  const now = Date.now()
  if (lastTime === 0 || curRx < lastRx || curTx < lastTx) {
    // 首次采样或累计值归零（系统重启），重新基准
    lastRx = curRx
    lastTx = curTx
    lastTime = now
    speed.value = { rx: 0, tx: 0 }
    return
  }
  const dt = (now - lastTime) / 1000
  if (dt <= 0)
    return
  speed.value = {
    rx: (curRx - lastRx) / dt,
    tx: (curTx - lastTx) / dt,
  }
  lastRx = curRx
  lastTx = curTx
  lastTime = now
}

async function getData() {
  try {
    const { data, code } = await getNetworkState<SystemMonitor.NetworkInfo>()
    if (code === 0 && data) {
      netState.value = data
      calcSpeed(data.rxBytes, data.txBytes)
    }
  }
  catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  getData()
  timer = setInterval(getData, (!props.refreshInterval || props.refreshInterval <= 2000) ? 2000 : props.refreshInterval)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<template>
  <NPopover trigger="click">
    <template #trigger>
      <!-- 详情风格 -->
      <div v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info" class="w-full">
        <div class="mb-1 text-xs flex justify-between" :style="{ color: textColor }">
          <span>NET</span>
          <span>↑{{ formatSpeed(speed.tx) }} ↓{{ formatSpeed(speed.rx) }}</span>
        </div>
      </div>
      <!-- 图标风格 -->
      <div v-else class="w-full flex flex-col justify-center items-center h-full mt-3 gap-[2px]">
        <div style="font-size: 9px;" :style="{ color: textColor }">
          ↑{{ formatSpeed(speed.tx) }}
        </div>
        <div style="font-size: 9px;" :style="{ color: textColor }">
          ↓{{ formatSpeed(speed.rx) }}
        </div>
      </div>
    </template>

    <div class="text-sm min-w-[160px]">
      <div class="mb-2">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.upload') }}</span>
        <span class="ml-2 font-bold">{{ formatSpeed(speed.tx) }}</span>
      </div>
      <div class="mb-2">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.download') }}</span>
        <span class="ml-2 font-bold">{{ formatSpeed(speed.rx) }}</span>
      </div>
      <div class="border-t border-current/10 pt-2 mt-2 opacity-80">
        <div>
          <span class="opacity-70">{{ $t('deskModule.systemMonitor.totalSent') }}</span>
          <span class="ml-2">{{ formatBytes(netState?.txBytes || 0) }}</span>
        </div>
        <div>
          <span class="opacity-70">{{ $t('deskModule.systemMonitor.totalRecv') }}</span>
          <span class="ml-2">{{ formatBytes(netState?.rxBytes || 0) }}</span>
        </div>
      </div>
    </div>
  </NPopover>
</template>
