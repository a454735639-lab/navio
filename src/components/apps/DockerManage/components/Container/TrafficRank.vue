<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NButton, NDataTable, NDrawer, NDrawerContent, NSpin, NTag } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getContainersStats } from '@/api/docker/container'
import { bytesToSize } from '@/utils/cmn'
import { useDockerStore } from '../../store/docker'
import { t } from '@/locales'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits(['update:visible'])

const dockerState = useDockerStore()
const loading = ref(false)
const errMsg = ref('')
const rows = ref<RankRow[]>([])

interface RankRow {
  id: string
  name: string
  rx: number // 累积下载，字节
  tx: number // 累积上传，字节
  total: number
}

const show = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
})

// 后端 networkRX/TX 单位是 KB（字节/1024），这里还原成字节用于统一格式化
function kbToBytes(kb: number | undefined): number {
  return (kb || 0) * 1024
}

async function fetchData() {
  const list = dockerState.containerList || []
  if (!list.length) {
    rows.value = []
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const ids = list.map(c => c.id)
    const { data, code } = await getContainersStats<Record<string, Docker.Container.Stats>>(ids)
    if (code !== 0 || !data) {
      errMsg.value = t('apps.dockerManage.loadFailed')
      return
    }
    const out: RankRow[] = []
    for (const c of list) {
      const s = data[c.id]
      if (!s)
        continue
      const rx = kbToBytes(s.networkRX)
      const tx = kbToBytes(s.networkTX)
      out.push({ id: c.id, name: c.name, rx, tx, total: rx + tx })
    }
    out.sort((a, b) => b.total - a.total)
    rows.value = out
  }
  catch (e: any) {
    errMsg.value = e?.msg || e?.message || t('apps.dockerManage.loadFailed')
  }
  finally {
    loading.value = false
  }
}

watch(() => props.visible, (v) => {
  if (v)
    fetchData()
})

const totals = computed(() => {
  let rx = 0
  let tx = 0
  for (const r of rows.value) {
    rx += r.rx
    tx += r.tx
  }
  return { rx, tx, total: rx + tx }
})

function fmtBytes(v: number): string {
  return bytesToSize(v)
}

const columns: DataTableColumns<RankRow> = [
  {
    title: '#',
    key: 'rank',
    width: 50,
    render(_row, index) {
      return `${index + 1}`
    },
  },
  {
    title: t('apps.dockerManage.containerName'),
    key: 'name',
    minWidth: 160,
    ellipsis: { tooltip: true },
  },
  {
    title: t('apps.dockerManage.download'),
    key: 'rx',
    minWidth: 100,
    sorter: (a, b) => a.rx - b.rx,
    render(row) {
      return fmtBytes(row.rx)
    },
  },
  {
    title: t('apps.dockerManage.upload'),
    key: 'tx',
    minWidth: 100,
    sorter: (a, b) => a.tx - b.tx,
    render(row) {
      return fmtBytes(row.tx)
    },
  },
  {
    title: t('apps.dockerManage.total'),
    key: 'total',
    minWidth: 100,
    sorter: (a, b) => a.total - b.total,
    render(row) {
      return fmtBytes(row.total)
    },
  },
]
</script>

<template>
  <NDrawer v-model:show="show" :width="720" placement="right">
    <NDrawerContent :title="t('apps.dockerManage.trafficRank')" closable>
      <template #header-extra>
        <NButton size="small" :loading="loading" @click="fetchData">
          {{ t('common.refresh') }}
        </NButton>
      </template>

      <NSpin :show="loading">
        <div v-if="errMsg" class="text-red-500 mb-3">
          {{ errMsg }}
        </div>

        <div class="mb-4 flex flex-wrap gap-2">
          <NTag type="info" size="large" round>
            {{ t('apps.dockerManage.totalDownload') }}: {{ fmtBytes(totals.rx) }}
          </NTag>
          <NTag type="warning" size="large" round>
            {{ t('apps.dockerManage.totalUpload') }}: {{ fmtBytes(totals.tx) }}
          </NTag>
          <NTag type="success" size="large" round>
            {{ t('apps.dockerManage.totalAll') }}: {{ fmtBytes(totals.total) }}
          </NTag>
          <NTag size="large" round :bordered="false">
            {{ t('apps.dockerManage.containerCount', { n: rows.length }) }}
          </NTag>
        </div>

        <NDataTable
          :columns="columns" :data="rows" :bordered="false" size="small"
          :row-key="(r: RankRow) => r.id" :max-height="480"
        />

        <div class="mt-3 text-xs opacity-60">
          {{ t('apps.dockerManage.trafficTip') }}
        </div>
      </NSpin>
    </NDrawerContent>
  </NDrawer>
</template>
