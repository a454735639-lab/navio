<script setup lang="ts">
import { NButton, NPopconfirm, NSwitch, useMessage } from 'naive-ui'
import { computed, ref } from 'vue'
import { useDockerStore } from '../../../store/docker'
import { DockerContainerInfoState } from '@/enums/docker'
import { execSwitchAction, restartContainer } from '@/api/docker/container'
import { apiRespErrMsg } from '@/utils/cmn/apiMessage'
import { t } from '@/locales'

const props = defineProps<{
  value: Docker.Container.Info
}>()

const emit = defineEmits(['openLogs'])

const dockerState = useDockerStore()
const data = ref(props.value)
const ms = useMessage()
const switchLoading = ref(false)
const restartLoading = ref(false)

const restartShow = computed(() => {
  return data.value.state === DockerContainerInfoState.running
    || data.value.state === DockerContainerInfoState.paused
    || data.value.state === DockerContainerInfoState.exited
    || data.value.state === DockerContainerInfoState.created
})

function doRestart() {
  restartLoading.value = true
  restartContainer({ id: data.value.id }).then(({ code, msg }) => {
    restartLoading.value = false
    if (code !== 0) {
      ms.warning(msg)
      return
    }
    ms.success(t('iconItem.dockerAction.restartContainer'))
    dockerState.refreshContainerList()
  }).catch((res) => {
    restartLoading.value = false
    if (res.code === -1) {
      ms.error(`${t('common.failed')}:${res.msg}`, {
        duration: 5000,
        closable: true,
        keepAliveOnHover: true,
      })
      return
    }
    apiRespErrMsg(res)
  })
}

const switchShow = computed(() => {
  switch (data.value.state) {
    case DockerContainerInfoState.created:
    case DockerContainerInfoState.running:
    case DockerContainerInfoState.paused:
    case DockerContainerInfoState.dead:
    case DockerContainerInfoState.exited:
      return true

    case DockerContainerInfoState.restarting:
    case DockerContainerInfoState.removing:
      return true

    default:
      return false
  }
})

const switchValue = computed({
  get() {
    if (props.value.state === DockerContainerInfoState.running)
      return true
    return false
  },
  set() {
    // 更新状态
    switchLoading.value = true
    const acticonStatus = !switchValue.value
    execSwitchAction({ id: data.value.id, action: acticonStatus }).then(({ code, msg }) => {
      if (code !== 0)
        ms.warning(msg)
      else
        data.value.state = acticonStatus ? DockerContainerInfoState.running : DockerContainerInfoState.exited

      switchLoading.value = false
      dockerState.refreshContainerList()
    }).catch((res) => {
      switchLoading.value = false
      if (res.code === -1) {
        ms.error(`${t('common.failed')}:${res.msg}`, {
          duration: 5000,
          closable: true,
          keepAliveOnHover: true,
        })
        return
      }

      apiRespErrMsg(res)
    })
  },
})
</script>

<template>
  <div class="w-full flex items-center gap-2">
    <div v-if="switchShow">
      <NSwitch v-model:value="switchValue" :loading="switchLoading" />
    </div>
    <div v-else>
      <NSwitch v-model:value="switchValue" :loading="switchLoading" disabled />
    </div>
    <NPopconfirm v-if="restartShow" @positive-click="doRestart">
      <template #trigger>
        <NButton size="tiny" type="warning" :loading="restartLoading">
          {{ t('iconItem.dockerAction.restartContainer') }}
        </NButton>
      </template>
      {{ t('iconItem.dockerAction.restartConfirm') }}
    </NPopconfirm>
    <NButton size="tiny" @click="emit('openLogs')">
      {{ t('apps.dockerManage.monitor') }}
    </NButton>
  </div>
</template>
