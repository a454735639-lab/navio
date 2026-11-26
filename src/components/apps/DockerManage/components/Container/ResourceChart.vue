<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { getContainerStatus } from '@/api/docker/container'
import { t } from '@/locales'

interface Stats {
  cpuPercent: number
  memory: number
  memoryPercent: number
  memoryLimit: number
  runStatus: string
}

const props = defineProps<{ containerId: string, active: boolean }>()

// 前端滑动窗口：只保留最近 MAX_POINTS 个采样点（无后端历史存储）
const MAX_POINTS = 30
const INTERVAL = 3000

const W = 700
const H = 260
const padL = 44
const padR = 18
const padT = 18
const padB = 26
const plotW = W - padL - padR
const plotH = H - padT - padB

const cpuData = ref<number[]>([])
const memData = ref<number[]>([])
const lastMem = ref<{ usage: number, limit: number }>({ usage: 0, limit: 0 })
const err = ref('')
let timer: any = null

const gridLines = [0, 25, 50, 75, 100].map(v => ({ v, y: padT + (1 - v / 100) * plotH }))

function xy(v: number, i: number, n: number) {
  const x = n <= 1 ? padL + plotW : padL + i * (plotW / (n - 1))
  const y = padT + (1 - Math.max(0, Math.min(100, v)) / 100) * plotH
  return { x, y }
}

function points(arr: number[]) {
  const n = arr.length
  if (!n)
    return ''
  return arr.map((v, i) => { const p = xy(v, i, n); return `${p.x.toFixed(1)},${p.y.toFixed(1)}` }).join(' ')
}

function lastPt(arr: number[]) {
  const n = arr.length
  if (!n)
    return null
  return xy(arr[n - 1], n - 1, n)
}

const cpuPoints = computed(() => points(cpuData.value))
const memPoints = computed(() => points(memData.value))
const cpuLast = computed(() => lastPt(cpuData.value))
const memLast = computed(() => lastPt(memData.value))
const cpuNow = computed(() => cpuData.value.length ? cpuData.value[cpuData.value.length - 1] : null)
const memNow = computed(() => memData.value.length ? memData.value[memData.value.length - 1] : null)

async function sample() {
  if (!props.containerId)
    return
  try {
    const { data } = await getContainerStatus<Stats>(props.containerId)
    err.value = ''
    cpuData.value.push(data.cpuPercent || 0)
    memData.value.push(data.memoryPercent || 0)
    lastMem.value = { usage: data.memory || 0, limit: data.memoryLimit || 0 }
    if (cpuData.value.length > MAX_POINTS) {
      cpuData.value.shift()
      memData.value.shift()
    }
  }
  catch (e: any) {
    err.value = e?.msg || e?.message || 'error'
  }
}

function start() {
  stop()
  cpuData.value = []
  memData.value = []
  sample()
  timer = setInterval(sample, INTERVAL)
}

function stop() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

watch(() => [props.active, props.containerId], ([active]) => {
  if (active && props.containerId)
    start()
  else
    stop()
}, { immediate: true })

onUnmounted(stop)
</script>

<template>
  <div class="rc-wrap">
    <div class="rc-legend">
      <div class="rc-item">
        <span class="rc-dot rc-dot-cpu" />
        <span>CPU</span>
        <b>{{ cpuNow !== null ? cpuNow.toFixed(1) : '--' }}%</b>
      </div>
      <div class="rc-item">
        <span class="rc-dot rc-dot-mem" />
        <span>{{ t('apps.dockerManage.memory') }}</span>
        <b>{{ memNow !== null ? memNow.toFixed(1) : '--' }}%</b>
        <span v-if="lastMem.limit" class="rc-mem-detail">({{ lastMem.usage.toFixed(0) }} / {{ lastMem.limit.toFixed(0) }} MB)</span>
      </div>
      <div class="rc-note">
        {{ t('apps.dockerManage.realtime') }} · {{ INTERVAL / 1000 }}s
      </div>
    </div>

    <div v-if="err" class="rc-err">
      {{ t('apps.dockerManage.logError') }}: {{ err }}
    </div>

    <svg :viewBox="`0 0 ${W} ${H}`" class="rc-svg" preserveAspectRatio="xMidYMid meet">
      <line v-for="g in gridLines" :key="`g${g.v}`" :x1="padL" :x2="W - padR" :y1="g.y" :y2="g.y" class="rc-grid" />
      <text v-for="g in gridLines" :key="`t${g.v}`" :x="padL - 8" :y="g.y + 4" class="rc-ytick" text-anchor="end">
        {{ g.v }}%
      </text>

      <polyline v-if="cpuPoints" :points="cpuPoints" class="rc-line rc-line-cpu" />
      <polyline v-if="memPoints" :points="memPoints" class="rc-line rc-line-mem" />

      <circle v-if="cpuLast" :cx="cpuLast.x" :cy="cpuLast.y" r="3.5" class="rc-dot-cpu" />
      <circle v-if="memLast" :cx="memLast.x" :cy="memLast.y" r="3.5" class="rc-dot-mem" />
      <text v-if="cpuLast" :x="cpuLast.x - 6" :y="cpuLast.y - 8" class="rc-vlabel" text-anchor="end">
        {{ cpuNow?.toFixed(1) }}%
      </text>
      <text v-if="memLast" :x="memLast.x - 6" :y="memLast.y + 14" class="rc-vlabel rc-vlabel-mem" text-anchor="end">
        {{ memNow?.toFixed(1) }}%
      </text>
    </svg>
  </div>
</template>

<style scoped>
.rc-wrap {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rc-legend {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.rc-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.rc-item b {
  font-size: 16px;
  margin-left: 2px;
}

.rc-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.rc-dot-cpu {
  background: #f0883e;
}

.rc-dot-mem {
  background: #58a6ff;
}

.rc-mem-detail {
  font-size: 12px;
  opacity: 0.6;
}

.rc-note {
  margin-left: auto;
  font-size: 12px;
  opacity: 0.5;
}

.rc-err {
  color: #f85149;
  font-size: 12px;
}

.rc-svg {
  width: 100%;
  height: auto;
  background: #0d1117;
  border-radius: 6px;
}

.rc-grid {
  stroke: rgba(255, 255, 255, 0.08);
  stroke-width: 1;
}

.rc-ytick {
  fill: rgba(255, 255, 255, 0.45);
  font-size: 11px;
}

.rc-line {
  fill: none;
  stroke-width: 2;
  stroke-linejoin: round;
  stroke-linecap: round;
}

.rc-line-cpu {
  stroke: #f0883e;
}

.rc-line-mem {
  stroke: #58a6ff;
}

.rc-vlabel {
  fill: #f0883e;
  font-size: 11px;
}

.rc-vlabel-mem {
  fill: #58a6ff;
}
</style>
