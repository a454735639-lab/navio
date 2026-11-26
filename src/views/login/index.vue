<script setup lang="ts">
import { NButton, NCheckbox, NForm, NFormItem, NInput, NSelect, NSwitch, useMessage, useOsTheme } from 'naive-ui'
import type { InputInst } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Captcha, SvgIcon } from '@/components/common'
import { login } from '@/api'
import { useAccountManagerStore, useAppStore, useAuthStore, usePanelState } from '@/store'
import { router } from '@/router'
import { t } from '@/locales'
import { languageOptions } from '@/utils/defaultData'
import type { Language } from '@/store/modules/app/helper'
import { getSiteStting } from '@/api/openness'
import { apiRespErrMsg, setTitle } from '@/utils/cmn'
import { defaultSetting } from '@/store/modules/app/helper'

const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()
const panelState = usePanelState()
const accountManagerStore = useAccountManagerStore()
const osTheme = useOsTheme()
const ms = useMessage()
const loading = ref(false)
const defaultUsername = ref<string>(route.query.username as string || '')
const inputpasswordRef = ref<InputInst | null>(null)
const showPassword = ref(false)
const isDark = computed(() => {
  if (appStore.theme === 'auto')
    return osTheme.value === 'dark'
  else
    return appStore.theme === 'dark'
})
const languageValue = ref<Language>(appStore.language)
const siteSetting = ref<GlobalSetting.SiteSetting>(defaultSetting().siteSetting)
const captchaRef = ref()

const form = ref<Login.LoginReqest>({
  username: defaultUsername.value,
  password: '',
  rememberMe: false,
})

const REMEMBER_KEY = 'navio_remember_username'
const savedUsername = localStorage.getItem(REMEMBER_KEY)
if (savedUsername && !defaultUsername.value) {
  form.value.username = savedUsername
  form.value.rememberMe = true
}

function handleRememberChange(checked: boolean) {
  form.value.rememberMe = checked
  if (checked) {
    if (form.value.username)
      localStorage.setItem(REMEMBER_KEY, form.value.username)
  }
  else {
    localStorage.removeItem(REMEMBER_KEY)
  }
}

const loginPost = async () => {
  loading.value = true
  try {
    const res = await login<Login.LoginResponse>(form.value)
    if (res.code === 0) {
      authStore.setToken(res.data.token)
      authStore.setUserInfo(res.data)
      panelState.reset()

      accountManagerStore.addAccount({
        username: res.data.username || form.value.username,
        name: res.data.name || '',
        id: res.data.id || 0,
        cToken: res.data.token,
        loginStatus: true,
      })

      setTimeout(() => {
        ms.success(`Hi ${res.data.name},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 500)
    }
    else {
      loading.value = false
    }
  }
  catch (error) {
    if (appStore.siteSetting.isEnableLoginCaptcha)
      captchaRef.value.refresh()

    apiRespErrMsg(error)
    loading.value = false
  }
}

function handleEnterDown(e: any) {
  if (e.keyCode === 229)
    return
  if (e.key === 'Enter' && !e.isComposing) {
    loginPost()
  }
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
  history.go(0)
}

async function getConfig() {
  try {
    const { data } = await getSiteStting<GlobalSetting.SiteSetting>()
    const dcfg = defaultSetting().siteSetting
    if (data.title === '')
      data.title = dcfg.title
    if (data.loginBackgroundUrl === '')
      data.loginBackgroundUrl = dcfg.loginBackgroundUrl

    appStore.setSiteSetting(data)
    siteSetting.value = data
    setTitle(data.title)
  }
  catch (error) {
    apiRespErrMsg(error)
  }
}

function handleTheme(value: boolean) {
  appStore.setTheme(value ? 'dark' : 'light')
}

onMounted(() => {
  setTitle(appStore.siteSetting.title)
  getConfig()
  if (defaultUsername.value !== '') {
    form.value.username = defaultUsername.value
    inputpasswordRef.value?.focus()
  }

  if (route.query.type !== 'addLogin' && (authStore.token !== null && authStore.token !== '')) {
    router.push({ path: '/' })
  }
})
</script>

<template>
  <div class="login-page" :class="{ 'is-dark': isDark }">
    <!-- 背景 -->
    <div class="bg-image" :style="{ backgroundImage: `url('${appStore.siteSetting.loginBackgroundUrl}')` }" />
    <div class="bg-veil" />
    <div class="bg-blob blob-a" />
    <div class="bg-blob blob-b" />

    <!-- 工具栏 -->
    <div class="toolbar">
      <NSwitch v-model:value="isDark" size="small" @update:value="handleTheme">
        <template #checked-icon>
          <SvgIcon icon="mingcute:moon-fog-line" />
        </template>
        <SvgIcon icon="tabler-sun-filled" />
        <template #unchecked-icon>
          <SvgIcon icon="mingcute:sun-fog-line" />
        </template>
      </NSwitch>
      <NSelect
        v-model:value="languageValue" size="tiny" :options="languageOptions"
        class="lang-select" @update:value="handleChangeLanuage"
      />
    </div>

    <!-- 登录卡片 -->
    <div class="card-wrap">
      <div class="card">
        <div class="brand">
          <div class="logo">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="9" />
              <path d="m15.5 8.5-2.1 5-5 2.1 2.1-5 5-2.1Z" />
            </svg>
          </div>
          <div class="brand-name">{{ t('common.appName') }}</div>
          <div class="brand-sub">{{ t('login.brandSubtitle') }}</div>
        </div>

        <div class="rule"><span>{{ t('login.signInTitle') }}</span></div>

        <NForm :model="form" @keydown.enter="handleEnterDown">
          <NFormItem :show-label="false" class="field">
            <NInput
              v-model:value="form.username" size="large" :placeholder="$t('login.usernamePlaceholder')" clearable
              class="field-input"
            >
              <template #prefix>
                <SvgIcon icon="ph:user-bold" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem :show-label="false" class="field">
            <NInput
              ref="inputpasswordRef" v-model:value="form.password" size="large"
              :type="showPassword ? 'text' : 'password'" :placeholder="$t('login.passwordPlaceholder')"
              class="field-input"
            >
              <template #prefix>
                <SvgIcon icon="mdi:password-outline" />
              </template>
              <template #suffix>
                <span class="eye" @click="showPassword = !showPassword">
                  <SvgIcon :icon="showPassword ? 'mdi:eye-outline' : 'mdi:eye-off-outline'" />
                </span>
              </template>
            </NInput>
          </NFormItem>

          <NFormItem v-if="appStore.siteSetting.isEnableLoginCaptcha" :show-label="false" class="field">
            <div class="captcha-row">
              <div class="captcha-box" @click="captchaRef?.refresh?.()">
                <Captcha ref="captchaRef" src="/api/captcha/getImage" />
              </div>
              <NInput
                v-model:value="form.vcode" size="large" :placeholder="$t('login.captchaPlaceholder')"
                class="field-input"
              >
                <template #prefix>
                  <SvgIcon icon="mdi:shield-lock-outline" />
                </template>
              </NInput>
            </div>
          </NFormItem>

          <div class="options">
            <NCheckbox v-if="route.query.type !== 'addLogin'" :checked="form.rememberMe" size="small" @update:checked="handleRememberChange">
              {{ $t('login.rememberAccount') }}
            </NCheckbox>
          </div>

          <NButton
            type="primary" size="large" block :loading="loading" class="submit"
            @click="loginPost"
          >
            <span class="submit-label">{{ $t('login.loginButton') }}</span>
            <SvgIcon icon="mdi:arrow-right" class="submit-arrow" />
          </NButton>
        </NForm>

        <div class="card-foot">{{ $t('login.selfHosted') }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: #eef2f9;
  color: #1f2740;
}
.login-page.is-dark {
  background: #0d1220;
  color: #e7ecf6;
}

.bg-image {
  position: absolute;
  inset: -24px;
  background-size: cover;
  background-position: center;
  filter: blur(18px) saturate(120%);
  transform: scale(1.06);
  opacity: .55;
}
.is-dark .bg-image { opacity: .3; filter: blur(18px) saturate(120%) brightness(.6); }

.bg-veil {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(99,102,241,.28), rgba(56,189,248,.18) 45%, rgba(255,255,255,.5));
}
.is-dark .bg-veil {
  background: linear-gradient(135deg, rgba(49,46,129,.55), rgba(8,47,73,.5) 45%, rgba(13,18,32,.85));
}

.bg-blob { position: absolute; border-radius: 50%; filter: blur(70px); opacity: .45; }
.blob-a { width: 380px; height: 380px; background: #7c8cff; top: -120px; left: -100px; }
.blob-b { width: 320px; height: 320px; background: #67e8f9; bottom: -120px; right: -80px; }
.is-dark .bg-blob { opacity: .22; }

.toolbar {
  position: absolute;
  top: 22px;
  right: 26px;
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(255,255,255,.55);
  backdrop-filter: blur(14px);
  box-shadow: 0 6px 22px rgba(31,39,64,.08);
}
.is-dark .toolbar { background: rgba(255,255,255,.08); box-shadow: 0 6px 22px rgba(0,0,0,.25); }
.lang-select { width: 96px; }

.card-wrap { position: relative; z-index: 2; width: min(420px, 92vw); }

.card {
  padding: 40px 36px 28px;
  border-radius: 24px;
  background: rgba(255,255,255,.72);
  border: 1px solid rgba(255,255,255,.85);
  backdrop-filter: blur(22px) saturate(150%);
  box-shadow: 0 24px 60px rgba(31,39,64,.16), inset 0 1px 0 rgba(255,255,255,.9);
}
.is-dark .card {
  background: rgba(24,31,51,.66);
  border-color: rgba(255,255,255,.1);
  box-shadow: 0 24px 60px rgba(0,0,0,.45), inset 0 1px 0 rgba(255,255,255,.06);
}

.brand { text-align: center; margin-bottom: 22px; }
.logo {
  width: 58px; height: 58px; margin: 0 auto 14px;
  border-radius: 18px;
  display: flex; align-items: center; justify-content: center;
  color: #fff;
  background: linear-gradient(140deg, #6d7cff, #4f46e5 60%, #38bdf8);
  box-shadow: 0 12px 26px rgba(79,70,229,.38), inset 0 1px 0 rgba(255,255,255,.4);
}
.logo svg { width: 30px; height: 30px; }
.brand-name { font-size: 24px; font-weight: 800; letter-spacing: .5px; }
.brand-sub { font-size: 12px; opacity: .62; margin-top: 5px; }

.rule {
  display: flex; align-items: center; gap: 12px;
  margin: 4px 0 22px;
  font-size: 12px; opacity: .55;
}
.rule::before, .rule::after { content: ''; flex: 1; height: 1px; background: currentColor; opacity: .28; }

.field { margin-bottom: 16px; }
.field-input :deep(.n-input) { border-radius: 12px; }
.eye { display: inline-flex; cursor: pointer; opacity: .6; }
.eye:hover { opacity: 1; }

.captcha-row { display: flex; gap: 12px; width: 100%; align-items: center; }
.captcha-box {
  width: 116px; height: 40px; flex-shrink: 0; cursor: pointer;
  border-radius: 12px; overflow: hidden;
  border: 1px solid rgba(127,140,180,.35);
}
.captcha-row .field-input { flex: 1; }

.options { display: flex; align-items: center; height: 24px; margin: 2px 0 14px; }

.submit {
  height: 46px;
  border-radius: 12px;
  font-weight: 700;
  background: linear-gradient(120deg, #5b6cff, #4f46e5);
  box-shadow: 0 10px 22px rgba(79,70,229,.3);
  position: relative;
}
.submit-label { transition: transform .15s ease; }
.submit-arrow { position: absolute; right: 18px; font-size: 18px; opacity: .85; }

.card-foot { text-align: center; font-size: 11px; opacity: .45; margin-top: 20px; }

@media (max-width: 480px) {
  .card { padding: 32px 24px 24px; }
  .toolbar { top: 14px; right: 14px; }
}
</style>
