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
  if (appStore.theme === 'auto') return osTheme.value === 'dark'
  else return appStore.theme === 'dark'
})
const languageValue = ref<Language>(appStore.language)
const siteSetting = ref<GlobalSetting.SiteSetting>(defaultSetting().siteSetting)
const captchaRef = ref()

const form = ref<Login.LoginReqest>({ username: defaultUsername.value, password: '', rememberMe: false })
const REMEMBER_KEY = 'navio_remember_username'
const savedUsername = localStorage.getItem(REMEMBER_KEY)
if (savedUsername && !defaultUsername.value) {
  form.value.username = savedUsername
  form.value.rememberMe = true
}

function handleRememberChange(checked: boolean) {
  form.value.rememberMe = checked
  if (checked && form.value.username) localStorage.setItem(REMEMBER_KEY, form.value.username)
  else localStorage.removeItem(REMEMBER_KEY)
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
        name: res.data.name || '', id: res.data.id || 0,
        cToken: res.data.token, loginStatus: true,
      })
      setTimeout(() => {
        ms.success(`Hi ${res.data.name || form.value.username},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 400)
    } else { loading.value = false }
  } catch (error) {
    if (appStore.siteSetting.isEnableLoginCaptcha) captchaRef.value.refresh()
    apiRespErrMsg(error); loading.value = false
  }
}

function handleEnterDown(e: any) {
  if (e.keyCode === 229) return
  if (e.key === 'Enter' && !e.isComposing) loginPost()
}
function handleChangeLanuage(v: Language) { languageValue.value = v; appStore.setLanguage(v); history.go(0) }

async function getConfig() {
  try {
    const { data } = await getSiteStting<GlobalSetting.SiteSetting>()
    const dcfg = defaultSetting().siteSetting
    if (data.title === '') data.title = dcfg.title
    if (data.loginBackgroundUrl === '') data.loginBackgroundUrl = dcfg.loginBackgroundUrl
    appStore.setSiteSetting(data); siteSetting.value = data; setTitle(data.title)
  } catch (error) { apiRespErrMsg(error) }
}
function handleTheme(v: boolean) { appStore.setTheme(v ? 'dark' : 'light') }

onMounted(() => {
  setTitle(appStore.siteSetting.title); getConfig()
  if (defaultUsername.value !== '') { form.value.username = defaultUsername.value; inputpasswordRef.value?.focus() }
  if (route.query.type !== 'addLogin' && (authStore.token !== null && authStore.token !== '')) router.push({ path: '/' })
})
</script>

<template>
  <div class="root" :class="isDark ? 'dark' : 'light'">
    <!-- 背景：浅色用极淡的纯色，深色用深靛蓝，不用模糊大图避免抢视线 -->
    <div class="bg" :style="{ backgroundImage: siteSetting.loginBackgroundUrl ? `url('${siteSetting.loginBackgroundUrl}')` : undefined }" />

    <!-- 工具栏：主题 + 语言，极简胶囊 -->
    <div class="toolbar">
      <NSwitch v-model:value="isDark" size="small" @update:value="handleTheme">
        <template #checked-icon><SvgIcon icon="mingcute:moon-fog-line" /></template>
        <template #unchecked-icon><SvgIcon icon="mingcute:sun-fog-line" /></template>
      </NSwitch>
      <NSelect v-model:value="languageValue" size="tiny" :options="languageOptions" class="lang-select" @update:value="handleChangeLanuage" />
    </div>

    <div class="card-wrap">
      <div class="card">
        <!-- 品牌：只保留名字，副标题收进辅助行，不再重复显示 siteTitle -->
        <div class="brand">
          <svg class="logo" viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="16" cy="16" r="13"/>
            <path d="M22 10l-4 9-9 4 4-9z"/>
          </svg>
          <div class="name">{{ $t('common.appName') }}</div>
        </div>

        <!-- 分割线只在有站点标题时显示 -->
        <div v-if="appStore.siteSetting.title" class="divider">
          <span>{{ appStore.siteSetting.title }}</span>
        </div>

        <NForm :model="form" @keydown.enter="handleEnterDown">
          <NFormItem :show-feedback="false" class="field">
            <NInput v-model:value="form.username" size="large" :placeholder="$t('login.usernamePlaceholder')" clearable>
              <template #prefix><SvgIcon icon="ph:user-bold" /></template>
            </NInput>
          </NFormItem>

          <NFormItem :show-feedback="false" class="field">
            <NInput
              ref="inputpasswordRef" v-model:value="form.password" size="large"
              :type="showPassword ? 'text' : 'password'"
              :placeholder="$t('login.passwordPlaceholder')"
            >
              <template #prefix><SvgIcon icon="mdi:password-outline" /></template>
              <template #suffix>
                <span class="eye" @click="showPassword = !showPassword">
                  <SvgIcon :icon="showPassword ? 'mdi:eye-outline' : 'mdi:eye-off-outline'" />
                </span>
              </template>
            </NInput>
          </NFormItem>

          <NFormItem v-if="appStore.siteSetting.isEnableLoginCaptcha" :show-feedback="false" class="field">
            <div class="captcha-row">
              <div class="captcha-box" @click="captchaRef?.refresh?.()">
                <Captcha ref="captchaRef" src="/api/captcha/getImage" />
              </div>
              <NInput v-model:value="form.vcode" size="large" :placeholder="$t('login.captchaPlaceholder')">
                <template #prefix><SvgIcon icon="mdi:shield-lock-outline" /></template>
              </NInput>
            </div>
          </NFormItem>

          <div class="footrow">
            <NCheckbox v-if="route.query.type !== 'addLogin'" :checked="form.rememberMe" size="small" @update:checked="handleRememberChange">
              {{ $t('login.rememberAccount') }}
            </NCheckbox>
            <span class="hint">{{ $t('login.selfHosted') }}</span>
          </div>

          <NButton
            type="primary" size="large" block :loading="loading"
            class="submit" @click="loginPost"
          >
            {{ $t('login.loginButton') }}
          </NButton>
        </NForm>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── unslop Pass 1: spacing ───────────────────────────── */
.root { min-height: 100vh; position: relative; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.bg { position: absolute; inset: 0; background-size: cover; background-position: center; opacity: .55; }
.dark .bg { opacity: .28; filter: brightness(.7) saturate(.8); }
.toolbar { position: absolute; top: 20px; right: 22px; z-index: 3; display: flex; gap: 10px; align-items: center; background: rgba(127,136,155,.18); border: 1px solid rgba(127,136,155,.25); border-radius: 999px; padding: 4px 10px 4px 6px; backdrop-filter: blur(10px); }
.dark .toolbar { background: rgba(255,255,255,.06); border-color: rgba(255,255,255,.08); }
.lang-select { width: 84px; }

.card-wrap { position: relative; z-index: 2; width: min(380px, 90vw); }

.card {
  /* Pass 6: layout / container — single card, no extra chrome */
  padding: 36px 32px 32px;
  border-radius: 20px;
  background: rgba(255,255,255,.78);
  border: 1px solid rgba(255,255,255,.88);
  box-shadow: 0 20px 50px rgba(31,39,64,.10), 0 1px 0 rgba(255,255,255,.9) inset;
  backdrop-filter: blur(20px) saturate(140%);
}
.dark .card {
  background: rgba(18,24,38,.76);
  border-color: rgba(255,255,255,.07);
  box-shadow: 0 20px 50px rgba(0,0,0,.38), 0 1px 0 rgba(255,255,255,.04) inset;
}

/* ── Pass 2: color — neutral brand, one accent only ──── */
.brand { display: flex; align-items: center; gap: 12px; margin-bottom: 22px; }
.logo { width: 36px; height: 36px; flex-shrink: 0; color: var(--accent); }
/* Accent defined once per theme context via CSS vars injected by n-config-provider */
.name { font-size: 22px; font-weight: 700; letter-spacing: -.2px; line-height: 1.1; }
.dark .name { color: #f0f4ff; }
.light .name { color: #0f172a; }

.divider { display: flex; align-items: center; gap: 14px; margin: -2px 0 24px; color: rgba(100,116,139,.7); font-size: 12px; font-weight: 500; }
.divider::before, .divider::after { content: ''; flex: 1; height: 1px; background: currentColor; opacity: .22; }

.field { margin-bottom: 18px; }
.field :deep(.n-input) { border-radius: 11px; height: 42px; }
.eye { display: inline-flex; cursor: pointer; opacity: .55; transition: opacity .15s; }
.eye:hover { opacity: 1; }
.captcha-row { display: flex; gap: 12px; align-items: center; }
.captcha-box { width: 110px; height: 42px; flex-shrink: 0; border-radius: 11px; overflow: hidden; cursor: pointer; border: 1px solid rgba(127,136,155,.35); }

/* ── Pass 4: copy — keep one action, no redundant labels ─ */
.footrow { display: flex; align-items: center; justify-content: space-between; height: 26px; margin: 2px 0 16px; font-size: 13px; }
.hint { color: rgba(100,116,139,.65); white-space: nowrap; }

.submit {
  height: 44px; border-radius: 11px;
  background: var(--accent);
  box-shadow: 0 8px 20px color-mix(in srgb, var(--accent) 55%, transparent);
  font-weight: 600; border: none;
}

/* Dark/light override for accent */
.dark .submit { background: #6366f1; box-shadow: 0 8px 20px rgba(99,102,241,.35); }
.light .submit { background: #4f46e5; box-shadow: 0 8px 20px rgba(79,70,229,.28); }

@media (max-width: 480px) {
  .card { padding: 28px 22px 26px; }
  .toolbar { top: 14px; right: 14px; }
}
</style>
