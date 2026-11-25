<script setup lang="ts">
import { NTag } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { get } from '@/api/system/about'
import srcSvglogo from '@/assets/logo.svg'

interface Version {
  versionName: string
  versionCode: number
}

const versionName = ref('')

onMounted(() => {
  get<Version>().then((res) => {
    if (res.code === 0)
      versionName.value = (res.data.versionName || '').trim()
  }).catch(() => {})
})
</script>

<template>
  <div class="app-background h-full overflow-auto py-8">
    <div class="flex flex-col items-center justify-center gap-3">
      <img :src="srcSvglogo" width="72" height="72" alt="">
      <div class="text-2xl font-semibold">
        {{ $t('common.appName') }}
      </div>
      <NTag type="info" :bordered="false">
        v{{ versionName }}
      </NTag>
      <p class="max-w-[560px] text-center text-sm opacity-80 px-6 leading-6">
        {{ $t('apps.about.intro') }}
      </p>
    </div>
  </div>
</template>
