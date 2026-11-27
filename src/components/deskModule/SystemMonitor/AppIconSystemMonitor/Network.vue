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
const RX_COLOR = '#f0883e'
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
      <!-- 详情风格：竖向大号 ↑↓，和环形仪表卡布局协调 -->
      <div v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info" class="flex flex-col items-center justify-center w-full py-1">
        <!-- 主视觉区与环形仪表同为 56px 高，保证四张卡等高对齐 -->
        <div class="flex flex-col items-center justify-center" style="width: 100%; height: 56px;">
          <div class="text-center" :style="{ color: textColor }">
            <div style="font-size: 14px; font-weight: 600; line-height: 1.3;">
              <span :style="{ color: progressColor }">↑{{ formatSpeed(speed.tx) }}</span>
            </div>
            <div style="font-size: 14px; font-weight: 600; line-height: 1.3; margin-top: 4px;">
              <span :style="{ color: RX_COLOR }">↓{{ formatSpeed(speed.rx) }}</span>
            </div>
          </div>
        </div>
        <div class="mt-1 text-sm font-medium text-center w-full px-1" :style="{ color: textColor }">
          {{ $t('deskModule.systemMonitor.network') }}
        </div>
        <!-- 累积上传/下载并排显示（对齐环形仪表副标题区） -->
        <div class="flex items-center justify-center w-full gap-2 mt-[2px]" style="font-size: 11px; line-height: 1.3;">
          <span class="truncate" :style="{ color: progressColor }" :title="`${$t('deskModule.systemMonitor.totalSent')} ${formatBytes(netState?.txBytes || 0)}`">
            ↑{{ formatBytes(netState?.txBytes || 0) }}
          </span>
          <span class="truncate" :style="{ color: RX_COLOR }" :title="`${$t('deskModule.systemMonitor.totalRecv')} ${formatBytes(netState?.rxBytes || 0)}`">
            ↓{{ formatBytes(netState?.rxBytes || 0) }}
          </span>
        </div>
      </div>
      <!-- 图标风格：竖排数字 -->
      <div v-else class="w-full flex flex-col justify-center items-center h-full mt-3 gap-[2px]">
        <div style="font-size: 9px;" :style="{ color: progressColor }">
          ↑{{ formatSpeed(speed.tx) }}
        </div>
        <div style="font-size: 9px;" :style="{ color: RX_COLOR }">
          ↓{{ formatSpeed(speed.rx) }}
        </div>
      </div>
    </template>

    <div class="text-sm min-w-[200px]">
      <div class="mb-2 flex justify-between">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.upload') }}</span>
        <span class="font-bold" :style="{ color: progressColor }">{{ formatSpeed(speed.tx) }}</span>
      </div>
      <div class="mb-2 flex justify-between">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.download') }}</span>
        <span class="font-bold" :style="{ color: RX_COLOR }">{{ formatSpeed(speed.rx) }}</span>
      </div>
      <div class="border-t border-current/10 pt-2 mt-2 opacity-80">
        <div class="flex justify-between">
          <span class="opacity-70">{{ $t('deskModule.systemMonitor.totalSent') }}</span>
          <span>{{ formatBytes(netState?.txBytes || 0) }}</span>
        </div>
        <div class="flex justify-between">
          <span class="opacity-70">{{ $t('deskModule.systemMonitor.totalRecv') }}</span>
          <span>{{ formatBytes(netState?.rxBytes || 0) }}</span>
        </div>
      </div>
    </div>
  </NPopover>
</template>
