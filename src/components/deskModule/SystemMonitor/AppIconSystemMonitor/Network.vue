<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { NPopover } from 'naive-ui'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { getNetworkState } from '@/api/system/systemMonitor'
import Sparkline from '../components/Sparkline/index.vue'

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
// 双线趋势：上传(progressColor) + 下载(橙)，统一刻度归一化
const txHistory = ref<number[]>([])
const rxHistory = ref<number[]>([])
const RX_COLOR = '#f0883e'
let lastRx = 0
let lastTx = 0
let lastTime = 0

const netMax = computed(() => {
  const m = Math.max(...txHistory.value, ...rxHistory.value, 0)
  return m > 0 ? m : 1
})

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
  // 记录趋势（30 点滑动窗口）
  txHistory.value.push(speed.value.tx)
  rxHistory.value.push(speed.value.rx)
  if (txHistory.value.length > 30) {
    txHistory.value.shift()
    rxHistory.value.shift()
  }
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
      <!-- 详情风格：大号 ↑↓ + 双线趋势 -->
      <div v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info" class="w-full">
        <div class="flex justify-between items-baseline mb-[4px]" :style="{ color: textColor }">
          <span class="text-xs font-medium">NET</span>
          <span class="text-xs">
            <span :style="{ color: progressColor }">↑{{ formatSpeed(speed.tx) }}</span>
            <span class="ml-2" :style="{ color: RX_COLOR }">↓{{ formatSpeed(speed.rx) }}</span>
          </span>
        </div>
        <div class="relative w-full" style="height: 30px;">
          <Sparkline :data="txHistory" :color="progressColor" :max="netMax" :height="30" class="absolute inset-0" />
          <Sparkline :data="rxHistory" :color="RX_COLOR" :max="netMax" :height="30" class="absolute inset-0" />
        </div>
      </div>
      <!-- 图标风格：竖排数字（空间小不加趋势） -->
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
