<script setup lang="ts">
import type { SelectOption } from 'naive-ui'
import { computed, nextTick, ref, watch } from 'vue'
import { NButton, NDrawer, NDrawerContent, NSelect, NSwitch, NTabPane, NTabs, useMessage } from 'naive-ui'
import { streamContainerLogs } from '@/api/docker/container'
import ResourceChart from './ResourceChart.vue'
import { t } from '@/locales'

const props = defineProps<{
  visible: boolean
  containerName: string
  containerId: string
}>()
const emit = defineEmits(['update:visible'])

// 前端只保留最近 MAX_LINES 行，避免长时间跟随导致 DOM 无限增长
const MAX_LINES = 2000

const ms = useMessage()
const logs = ref<string[]>([])
const follow = ref(true)
const loading = ref(false)
const connected = ref(false)
const errMsg = ref('')
const logBox = ref<HTMLElement | null>(null)
const tail = ref('200')
const activeTab = ref('logs')
let abortFn: (() => void) | null = null
let streamGen = 0

const drawerWidth = ref<number | string>(720)

const show = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
})

const tailOptions: SelectOption[] = ['100', '200', '500', '1000', '2000', 'all'].map(v => ({
  label: v,
  value: v,
}))

function scrollToBottom() {
  nextTick(() => {
    const el = logBox.value
    if (el)
      el.scrollTop = el.scrollHeight
  })
}

function handleLine(line: string) {
  logs.value.push(line)
  if (logs.value.length > MAX_LINES)
    logs.value.splice(0, logs.value.length - MAX_LINES)
  if (follow.value)
    scrollToBottom()
}

function start() {
  stop()
  logs.value = []
  errMsg.value = ''
  if (!props.containerId) {
    errMsg.value = t('apps.dockerManage.noContainerId')
    return
  }
  loading.value = true
  connected.value = true
  const gen = ++streamGen
  abortFn = streamContainerLogs(
    { id: props.containerId, follow: true, tail: tail.value },
    handleLine,
    (msg) => {
      if (gen !== streamGen)
        return
      errMsg.value = msg
      connected.value = false
      loading.value = false
    },
    () => {
      // 仅当仍是当前流时才更新状态；被 abort 的旧流回调直接忽略
      if (gen !== streamGen)
        return
      connected.value = false
      loading.value = false
    },
  )
}

function stop() {
  streamGen++ // 使旧流回调失效，避免被 abort 的旧流误置 connected=false
  if (abortFn) {
    abortFn()
    abortFn = null
  }
}

function handleClear() {
  logs.value = []
}

function handleTailChange() {
  // 重新连接以应用新的 tail 行数
  if (props.visible)
    start()
}

function handleDownload() {
  if (logs.value.length === 0) {
    ms.warning(t('apps.dockerManage.emptyLog'))
    return
  }
  const blob = new Blob([logs.value.join('\n')], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.containerName || 'container'}-logs.txt`
  a.click()
  URL.revokeObjectURL(url)
}

function updateWidth() {
  drawerWidth.value = window.innerWidth < 640 ? '100%' : 720
}

watch(() => [props.visible, props.containerId], ([visible]) => {
  updateWidth()
  if (visible && props.containerId) {
    activeTab.value = 'logs'
    start()
  }
  else if (!visible)
    stop()
})
</script>

<template>
  <NDrawer v-model:show="show" :width="drawerWidth" placement="right">
    <NDrawerContent :title="`${containerName} ${t('apps.dockerManage.monitor')}`" closable>
      <NTabs v-model:value="activeTab" type="line" size="small" animated>
        <NTabPane name="logs" :tab="t('apps.dockerManage.logs')">
      <div class="log-toolbar">
        <div class="toolbar-item">
          <span class="toolbar-label">{{ t('apps.dockerManage.follow') }}</span>
          <NSwitch v-model:value="follow" size="small" />
        </div>
        <div class="toolbar-item">
          <span class="toolbar-label">{{ t('apps.dockerManage.tailLines') }}</span>
          <NSelect v-model:value="tail" size="small" :options="tailOptions" style="width: 100px;" @update:value="handleTailChange" />
        </div>
        <div class="toolbar-item">
          <NButton size="small" @click="handleClear">
            {{ t('apps.dockerManage.clearLog') }}
          </NButton>
          <NButton size="small" @click="handleDownload">
            {{ t('apps.dockerManage.downloadLog') }}
          </NButton>
        </div>
        <div class="toolbar-item status">
          <span class="status-dot" :class="connected ? 'on' : 'off'" />
          <span class="status-text">{{ connected ? t('apps.dockerManage.connected') : t('apps.dockerManage.disconnected') }}</span>
        </div>
      </div>

      <div v-if="errMsg" class="log-err">
        {{ t('apps.dockerManage.logError') }}: {{ errMsg }}
      </div>

      <div ref="logBox" class="log-box">
        <div v-if="logs.length === 0 && !loading && !errMsg" class="log-empty">
          {{ t('apps.dockerManage.emptyLog') }}
        </div>
        <div v-for="(line, i) in logs" :key="i" class="log-line">
          {{ line }}
        </div>
      </div>
          </NTabPane>
        <NTabPane name="resource" :tab="t('apps.dockerManage.resourceChart')">
          <ResourceChart :container-id="containerId" :active="visible && activeTab === 'resource'" />
        </NTabPane>
      </NTabs>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped>
.log-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.toolbar-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-label {
  font-size: 13px;
  opacity: 0.8;
}

.status {
  margin-left: auto;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.status-dot.on {
  background: #3fb950;
}

.status-dot.off {
  background: #8b949e;
}

.status-text {
  font-size: 12px;
  opacity: 0.7;
}

.log-err {
  color: #f85149;
  margin-bottom: 8px;
  font-size: 12px;
}

.log-box {
  height: calc(100vh - 190px);
  overflow: auto;
  background: #0d1117;
  color: #c9d1d9;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  padding: 12px;
  border-radius: 6px;
}

.log-line {
  white-space: pre-wrap;
  word-break: break-all;
}

.log-empty {
  opacity: 0.5;
  text-align: center;
  padding: 40px 0;
}
</style>
