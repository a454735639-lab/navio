<script setup lang="ts">
import { NButton, NCard, NCheckbox, NForm, NFormItem, NInput, NSelect, NSwitch, useMessage, useOsTheme } from 'naive-ui'
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

// 记住账号：勾选时把用户名存 localStorage，下次自动带出
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
  <div class="login-page">
    <div class="background" :style="{ backgroundImage: `url('${appStore.siteSetting.loginBackgroundUrl}')` }" />

    <!-- 顶部工具条：主题 + 语言，不再占用卡片内部空间 -->
    <div class="toolbar">
      <div class="toolbar-item">
        <NSwitch v-model:value="isDark" @update:value="handleTheme">
          <template #checked-icon>
            <SvgIcon icon="mingcute:moon-fog-line" />
          </template>
          <SvgIcon icon="tabler-sun-filled" />
          <template #unchecked-icon>
            <SvgIcon icon="mingcute:sun-fog-line" />
          </template>
        </NSwitch>
      </div>
      <div class="toolbar-item flex items-center gap-2">
        <SvgIcon icon="flowbite-language-outline" class="language-text-style" />
        <NSelect
          v-model:value="languageValue" size="small" :options="languageOptions"
          class="!w-[100px]" @update:value="handleChangeLanuage"
        />
      </div>
    </div>

    <div class="login-container">
      <NCard class="login-card" style="border-radius: 24px">
        <!-- 品牌区 -->
        <div class="brand">
          <div class="brand-icon">
            <SvgIcon icon="mdi:compass-outline" />
          </div>
          <div class="brand-text">
            <div class="brand-name">{{ t('common.appName') }}</div>
            <div class="brand-sub">{{ t('login.brandSubtitle') }}</div>
          </div>
        </div>

        <div class="login-title">{{ appStore.siteSetting.title }}</div>

        <NForm :model="form" @keydown.enter="handleEnterDown">
          <NFormItem>
            <NInput
              v-model:value="form.username" size="large" :placeholder="$t('login.usernamePlaceholder')" clearable
            >
              <template #prefix>
                <SvgIcon icon="ph:user-bold" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem>
            <NInput
              ref="inputpasswordRef" v-model:value="form.password" :type="showPassword ? 'text' : 'password'"
              :placeholder="$t('login.passwordPlaceholder')"
            >
              <template #prefix>
                <SvgIcon icon="mdi:password-outline" />
              </template>
              <template #suffix>
                <SvgIcon
                  :icon="showPassword ? 'mdi:eye-outline' : 'mdi:eye-off-outline'" class="cursor-pointer"
                  @click="showPassword = !showPassword"
                />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem v-if="appStore.siteSetting.isEnableLoginCaptcha">
            <div class="captcha-row">
              <div class="captcha-box">
                <Captcha ref="captchaRef" src="/api/captcha/getImage" />
              </div>
              <NInput
                v-model:value="form.vcode" size="large" :placeholder="$t('login.captchaPlaceholder')"
                class="captcha-input"
              >
                <template #prefix>
                  <SvgIcon icon="mdi:shield-lock-outline" />
                </template>
              </NInput>
            </div>
          </NFormItem>

          <NFormItem :show-feedback="false">
            <NCheckbox v-if="route.query.type !== 'addLogin'" :checked="form.rememberMe" size="small" @update:checked="handleRememberChange">
              {{ $t('login.rememberAccount') }}
            </NCheckbox>
          </NFormItem>

          <NButton
            type="primary" size="large" block :loading="loading" class="login-submit"
            @click="loginPost"
          >
            {{ $t('login.loginBtn') }}
          </NButton>
        </NForm>
      </NCard>
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
}

.background {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  filter: blur(12px) brightness(0.7);
}

.toolbar {
  position: absolute;
  top: 24px;
  right: 28px;
  display: flex;
  gap: 18px;
  z-index: 2;
}

.toolbar-item {
  display: flex;
  align-items: center;
}

.language-text-style {
  width: 22px;
  height: 22px;
  color: #fff;
  opacity: 0.85;
}

.login-container {
  position: relative;
  z-index: 1;
  width: min(400px, 92vw);
  padding: 8px;
}

.login-card {
  padding: 34px 30px 28px !important;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 22px;
}

.brand-icon {
  width: 46px;
  height: 46px;
  border-radius: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.35);
  flex-shrink: 0;
}

.brand-name {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.brand-sub {
  font-size: 12px;
  opacity: 0.6;
  margin-top: 2px;
}

.login-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 22px;
  text-align: center;
}

.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;
  align-items: center;
}

.captcha-box {
  width: 124px;
  height: 40px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  cursor: pointer;
}

.captcha-input {
  flex: 1;
}

.login-submit {
  margin-top: 6px;
  font-weight: 600;
}
</style>
