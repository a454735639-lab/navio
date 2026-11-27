<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { NPopover } from 'naive-ui'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { getNetworkState } from '@/api/system/systemMonitor'

interface Prop {
  cardTypeStyle: PanelPanelConfigStyleEnum
  refreshInterval: number
  textColor: string
  progressColor: string
  progressRailColor: string
  /** 速度警告阈值（KB/s），上下行任一超过即告警；0 表示不启用 */
  warningThreshold?: number
  warningColor?: string
}

const props = withDefaults(defineProps<Prop>(), {
  warningThreshold: 0,
  warningColor: '#ef4444',
})
let timer: any
const netState = ref<SystemMonitor.NetworkInfo | null>(null)
const speed = ref({ rx: 0, tx: 0 })
let lastRx = 0
let lastTx = 0
let lastTime = 0

// 近期速度峰值（滑动窗口），用于环形仪表的相对刻度，让低速时也有可见反馈
const speedHistory = ref<number[]>([])
const HISTORY_SIZE = 20

function formatSpeed(bytesPerSec: number): string {
  const p = speedParts(bytesPerSec)
  return `${p.value} ${p.unit}`
}

function speedParts(bytesPerSec: number): { value: string, unit: string } {
  if (bytesPerSec < 1024)
    return { value: bytesPerSec.toFixed(0), unit: 'B/s' }
  if (bytesPerSec < 1024 * 1024)
    return { value: (bytesPerSec / 1024).toFixed(1), unit: 'KB/s' }
  return { value: (bytesPerSec / 1024 / 1024).toFixed(1), unit: 'MB/s' }
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
    recordPeak(0)
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
  recordPeak(Math.max(speed.value.rx, speed.value.tx))
}

function recordPeak(v: number) {
  const list = speedHistory.value
  list.push(v)
  if (list.length > HISTORY_SIZE)
    list.shift()
}

const maxSpeed = computed(() => Math.max(speed.value.rx, speed.value.tx))
const mainParts = computed(() => speedParts(maxSpeed.value))

// 超过阈值（KB/s）即告警，整卡切警告色
const isWarning = computed(() => props.warningThreshold > 0 && maxSpeed.value / 1024 >= props.warningThreshold)
const activeColor = computed(() => isWarning.value ? props.warningColor : props.progressColor)

// 环形仪表：相对近期峰值的占比；有阈值时超阈值直接充满，直观提示过载
const R = 42
const CIRC = 2 * Math.PI * R
const ringPercentage = computed(() => {
  if (isWarning.value)
    return 100
  const peak = Math.max(...speedHistory.value, 0)
  if (peak <= 0 || maxSpeed.value <= 0)
    return 0
  return Math.min(100, Math.max(4, (maxSpeed.value / peak) * 100))
})
const dashOffset = computed(() => CIRC * (1 - ringPercentage.value / 100))

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
      <!-- 详情风格：与其他环形仪表保持同构（环 + 标题 + 数值） -->
      <div v-if="cardTypeStyle === PanelPanelConfigStyleEnum.info" class="flex flex-col items-center justify-center w-full py-1">
        <div class="relative" style="width: 56px; height: 56px;">
          <svg viewBox="0 0 100 100" style="width: 100%; height: 100%;">
            <circle cx="50" cy="50" :r="R" fill="none" :stroke="progressRailColor" stroke-width="8" stroke-opacity="0.3" />
            <circle
              cx="50" cy="50" :r="R" fill="none" :stroke="activeColor" stroke-width="8"
              stroke-linecap="round" :stroke-dasharray="CIRC" :stroke-dashoffset="dashOffset"
              transform="rotate(-90 50 50)"
            />
          </svg>
          <div class="absolute inset-0 flex flex-col items-center justify-center" :style="{ color: isWarning ? activeColor : textColor }">
            <span style="font-size: 14px; font-weight: 600; line-height: 1;">{{ mainParts.value }}</span>
            <span style="font-size: 8px; line-height: 1.2; opacity: 0.75;">{{ mainParts.unit }}</span>
          </div>
        </div>
        <div class="mt-1 text-sm font-medium text-center w-full px-1" :style="{ color: isWarning ? activeColor : textColor }">
          {{ $t('deskModule.systemMonitor.network') }}
        </div>
        <!-- 上下行速率并排（同色系，避免蓝橙混杂） -->
        <div class="flex items-center justify-center w-full gap-2 mt-[2px]" style="font-size: 11px; line-height: 1.3;">
          <span class="truncate" :style="{ color: activeColor }" :title="`${$t('deskModule.systemMonitor.upload')} ${formatSpeed(speed.tx)}`">
            ↑{{ formatSpeed(speed.tx) }}
          </span>
          <span class="truncate opacity-70" :style="{ color: activeColor }" :title="`${$t('deskModule.systemMonitor.download')} ${formatSpeed(speed.rx)}`">
            ↓{{ formatSpeed(speed.rx) }}
          </span>
        </div>
      </div>
      <!-- 图标风格：竖排数字 -->
      <div v-else class="w-full flex flex-col justify-center items-center h-full mt-3 gap-[2px]">
        <div style="font-size: 9px;" :style="{ color: activeColor }">
          ↑{{ formatSpeed(speed.tx) }}
        </div>
        <div style="font-size: 9px; opacity: 0.75;" :style="{ color: activeColor }">
          ↓{{ formatSpeed(speed.rx) }}
        </div>
      </div>
    </template>

    <div class="text-sm min-w-[200px]">
      <div class="mb-2 flex justify-between">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.upload') }}</span>
        <span class="font-bold" :style="{ color: activeColor }">{{ formatSpeed(speed.tx) }}</span>
      </div>
      <div class="mb-2 flex justify-between">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.download') }}</span>
        <span class="font-bold" :style="{ color: activeColor }">{{ formatSpeed(speed.rx) }}</span>
      </div>
      <div v-if="warningThreshold > 0" class="mb-2 flex justify-between opacity-80">
        <span class="opacity-70">{{ $t('deskModule.systemMonitor.warningThreshold') }}</span>
        <span>{{ warningThreshold }} KB/s</span>
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
