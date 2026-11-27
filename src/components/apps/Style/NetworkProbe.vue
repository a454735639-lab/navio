<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  NAlert, NButton, NCard, NFlex, NInput, NInputNumber, NSelect,
  NSpace, NSwitch, NTag, useMessage,
} from 'naive-ui'
import { getProbeConfig, saveProbeConfig, testProbe } from '@/api/panel/network'
import type { ProbeConfig, ProbeConfigResp } from '@/api/panel/network'
import { usePanelState } from '@/store'
import { t } from '@/locales'

const ms = useMessage()
const panelState = usePanelState()

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const form = ref<ProbeConfig>({
  enabled: false,
  port: 8443,
  provider: 'path',
  certPath: '',
  keyPath: '',
  acmeEmail: '',
  acmeDomains: [],
  dnsProvider: 'cloudflare',
  staging: false,
})
const dnsToken = ref('')
const hasToken = ref(false)
const cert = ref<ProbeConfigResp['cert']>({})
const probeResult = ref<{ lan: string, ts: string }>({ lan: '-', ts: '-' })

const providerOptions = [
  { label: 'path（已有证书文件）', value: 'path' },
  { label: 'upload（上传到 conf/probe/）', value: 'upload' },
  { label: 'acme（自动申请）', value: 'acme' },
]

const domainsText = computed({
  get: () => (form.value.acmeDomains || []).join(', '),
  set: (v: string) => { form.value.acmeDomains = v.split(/[,\s]+/).filter(Boolean) },
})

const certExpireText = computed(() => {
  if (!cert.value?.notAfter)
    return '-'
  const d = new Date(cert.value.notAfter)
  if (Number.isNaN(d.getTime()))
    return '-'
  const days = Math.floor((d.getTime() - Date.now()) / 86400000)
  return `${cert.value.notAfter.slice(0, 10)}（剩 ${days} 天）`
})

async function load() {
  loading.value = true
  try {
    const { data } = await getProbeConfig<ProbeConfigResp>()
    form.value = { ...form.value, ...data.config }
    hasToken.value = data.hasToken
    cert.value = data.cert || {}
  }
  catch (e: any) {
    ms.error(e?.msg || String(e))
  }
  finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    await saveProbeConfig({ ...form.value, dnsToken: dnsToken.value || undefined })
    dnsToken.value = ''
    ms.success(t('common.success'))
    await load()
  }
  catch (e: any) {
    ms.error(e?.msg || String(e))
  }
  finally {
    saving.value = false
  }
}

async function runTest() {
  testing.value = true
  try {
    const { data } = await testProbe<any>()
    cert.value = data.cert || {}
    // 后端不再回传证书/私钥文件路径（避免泄露文件系统布局），仅显示 provider
    ms.success(`${data.provider}`)
  }
  catch (e: any) {
    ms.error(e?.msg || String(e))
  }
  finally {
    testing.value = false
  }
}

async function fetchLatency(url?: string) {
  if (!url)
    return '-'
  const started = performance.now()
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), 3000)
  try {
    await fetch(url, { mode: 'no-cors', cache: 'no-store', signal: controller.signal })
    return `${Math.round(performance.now() - started)} ms`
  }
  catch {
    return t('apps.networkProbe.unreachable')
  }
  finally {
    clearTimeout(timer)
  }
}

async function probeNetwork() {
  probeResult.value = {
    lan: await fetchLatency(panelState.panelConfig.netProbeLanUrl),
    ts: await fetchLatency(panelState.panelConfig.netProbeTsUrl),
  }
}

onMounted(load)
</script>

<template>
  <NCard style="border-radius:10px" class="mt-[10px]" size="small">
    <template #header>
      <div class="text-slate-500 font-bold flex items-center gap-2">
        {{ $t('apps.networkProbe.title') }}
        <NTag size="small" :type="form.enabled ? 'success' : 'default'">
          {{ form.enabled ? 'on' : 'off' }}
        </NTag>
      </div>
    </template>

    <NAlert type="info" :bordered="false" class="mb-2">
      {{ $t('apps.networkProbe.tip') }}
    </NAlert>

    <NSpace vertical :size="8">
      <NFlex align="center" :size="8">
        <NSwitch v-model:value="form.enabled" />
        <span>{{ $t('apps.networkProbe.enable') }}</span>
        <span class="ml-4">{{ $t('apps.networkProbe.port') }}</span>
        <NInputNumber v-model:value="form.port" size="small" :min="1" :max="65535" style="width:100px" />
      </NFlex>

      <NFlex align="center" :size="8">
        <span>{{ $t('apps.networkProbe.provider') }}</span>
        <NSelect v-model:value="form.provider" :options="providerOptions" size="small" style="width:240px" />
      </NFlex>

      <template v-if="form.provider !== 'acme'">
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.certPath') }}</span>
          <NInput v-model:value="form.certPath" size="small" placeholder="./conf/probe/fullchain.pem" />
        </NFlex>
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.keyPath') }}</span>
          <NInput v-model:value="form.keyPath" size="small" placeholder="./conf/probe/key.pem" />
        </NFlex>
      </template>

      <template v-else>
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.acmeEmail') }}</span>
          <NInput v-model:value="form.acmeEmail" size="small" placeholder="you@example.com" />
        </NFlex>
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.acmeDomains') }}</span>
          <NInput v-model:value="domainsText" size="small" placeholder="lan-probe.example.com, ts-probe.example.com" />
        </NFlex>
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.dnsProvider') }}</span>
          <NInput v-model:value="form.dnsProvider" size="small" placeholder="cloudflare" style="width:200px" />
          <NSwitch v-model:value="form.staging" />
          <span>{{ $t('apps.networkProbe.staging') }}</span>
        </NFlex>
        <NFlex align="center" :size="8">
          <span class="w-[92px] text-right">{{ $t('apps.networkProbe.dnsToken') }}</span>
          <NInput
            v-model:value="dnsToken" type="password" show-password-on="click" size="small"
            :placeholder="hasToken ? $t('apps.networkProbe.tokenSaved') : 'Cloudflare API Token'"
          />
        </NFlex>
      </template>

      <NFlex :size="8">
        <NButton size="small" type="primary" :loading="saving" @click="save">{{ $t('common.save') }}</NButton>
        <NButton size="small" :loading="testing" @click="runTest">{{ $t('apps.networkProbe.testCert') }}</NButton>
        <NButton size="small" @click="probeNetwork">{{ $t('apps.networkProbe.testProbe') }}</NButton>
      </NFlex>

      <NAlert v-if="cert?.lastError" type="warning" :bordered="false">
        {{ cert.lastError }}
      </NAlert>
      <NSpace :size="8" align="center">
        <NTag size="small">{{ $t('apps.networkProbe.certExpire') }}: {{ certExpireText }}</NTag>
        <NTag v-if="cert?.issuer" size="small">{{ cert.issuer }}</NTag>
      </NSpace>

      <NSpace :size="8">
        <NTag size="small">LAN {{ probeResult.lan }}</NTag>
        <NTag size="small">TS {{ probeResult.ts }}</NTag>
      </NSpace>
    </NSpace>
  </NCard>
</template>
