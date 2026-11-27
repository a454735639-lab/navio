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

// ── 卡片风格：glass（液态玻璃）/ soft（柔和实心）──────────────
type LoginCardStyle = 'glass' | 'soft'
const CARD_STYLE_KEY = 'navio_login_card_style'

function readCardStyle(): LoginCardStyle {
  try {
    return localStorage.getItem(CARD_STYLE_KEY) === 'soft' ? 'soft' : 'glass'
  }
  catch {
    return 'glass'
  }
}

const cardStyle = ref<LoginCardStyle>(readCardStyle())
const isGlass = computed(() => cardStyle.value === 'glass')

function toggleCardStyle() {
  cardStyle.value = isGlass.value ? 'soft' : 'glass'
  try {
    localStorage.setItem(CARD_STYLE_KEY, cardStyle.value)
  }
  catch {
    // 存储不可用时仅本次会话生效
  }
}

const form = ref<Login.LoginReqest>({ username: defaultUsername.value, password: '', rememberMe: false })
const REMEMBER_KEY = 'navio_remember_username'
// 隐私模式/禁用站点数据时 localStorage 访问会抛 SecurityError，需容错避免登录页白屏
let savedUsername: string | null = null
try {
  savedUsername = localStorage.getItem(REMEMBER_KEY)
}
catch {
  savedUsername = null
}
if (savedUsername && !defaultUsername.value) {
  form.value.username = savedUsername
  form.value.rememberMe = true
}

// 站点标题只在「与品牌名不同」时展示，避免出现 Navio / Navio 这种重复分割线
const siteTitle = computed(() => {
  const raw = (appStore.siteSetting.title || '').trim()
  if (!raw) return ''
  const brand = t('common.appName').trim()
  if (brand && raw.toLowerCase() === brand.toLowerCase()) return ''
  return raw
})

function handleRememberChange(checked: boolean) {
  form.value.rememberMe = checked
  try {
    if (checked && form.value.username) localStorage.setItem(REMEMBER_KEY, form.value.username)
    else localStorage.removeItem(REMEMBER_KEY)
  }
  catch {
    // 存储不可用时忽略，不影响登录
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
  <div class="root" :class="[isDark ? 'dark' : 'light', `card-style-${cardStyle}`]">
    <!-- 背景图 -->
    <div class="bg" :style="{ backgroundImage: siteSetting.loginBackgroundUrl ? `url('${siteSetting.loginBackgroundUrl}')` : undefined }" />

    <!-- 彩色光斑：玻璃需要有东西可折射，否则玻璃质感出不来 -->
    <div class="bg-aurora" />

    <!-- 压暗遮罩，保证卡片内文字对比度 -->
    <div class="bg-veil" />

    <!-- 工具栏 -->
    <div class="toolbar">
      <button
        type="button"
        class="style-toggle"
        :title="$t('login.switchCardStyle')"
        @click="toggleCardStyle"
      >
        <SvgIcon :icon="isGlass ? 'mdi:blur' : 'mdi:blur-linear'" class="style-icon" />
        <span>{{ isGlass ? $t('login.styleGlass') : $t('login.styleSoft') }}</span>
      </button>
      <span class="toolbar-sep" />
      <NSwitch v-model:value="isDark" size="small" @update:value="handleTheme">
        <template #checked-icon><SvgIcon icon="mingcute:moon-fog-line" /></template>
        <template #unchecked-icon><SvgIcon icon="mingcute:sun-fog-line" /></template>
      </NSwitch>
      <span class="toolbar-sep" />
      <NSelect
        v-model:value="languageValue" size="tiny" :options="languageOptions"
        class="lang-select" :show-arrow="false" @update:value="handleChangeLanuage"
      />
    </div>

    <div class="card-wrap">
      <!-- 卡片背后的柔光，制造「悬浮玻璃」错觉 -->
      <div class="card-glow" />

      <div class="card">
        <!-- 顶部弧形高光 + 边缘折射描边 -->
        <span class="card-sheen" aria-hidden="true" />
        <span class="card-rim" aria-hidden="true" />

        <div class="card-inner">
          <!-- 品牌区 -->
          <div class="brand">
            <div class="logo-badge">
              <svg class="logo" viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="16" cy="16" r="12.5" />
                <path d="M22 10l-4 9-9 4 4-9z" />
              </svg>
            </div>
            <div class="brand-text">
              <div class="name">{{ $t('common.appName') }}</div>
              <div class="tagline">{{ $t('login.brandSubtitle') }}</div>
            </div>
          </div>

          <!-- 站点自定义标题：仅在独立设置了标题时展示 -->
          <div v-if="siteTitle" class="site-chip">
            <span>{{ siteTitle }}</span>
          </div>

          <NForm :model="form" class="form" @keydown.enter="handleEnterDown">
            <NFormItem :show-feedback="false" class="field">
              <NInput v-model:value="form.username" size="large" :placeholder="$t('login.usernamePlaceholder')" clearable>
                <template #prefix><SvgIcon class="field-icon" icon="ph:user-bold" /></template>
              </NInput>
            </NFormItem>

            <NFormItem :show-feedback="false" class="field">
              <NInput
                ref="inputpasswordRef" v-model:value="form.password" size="large"
                :type="showPassword ? 'text' : 'password'"
                :placeholder="$t('login.passwordPlaceholder')"
              >
                <template #prefix><SvgIcon class="field-icon" icon="mdi:password-outline" /></template>
                <template #suffix>
                  <span class="eye" role="button" tabindex="0" @click="showPassword = !showPassword">
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
                  <template #prefix><SvgIcon class="field-icon" icon="mdi:shield-lock-outline" /></template>
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
  </div>
</template>

<style scoped>
/* ══ 主题变量 ══════════════════════════════════════════ */
.root {
  --accent: #4f46e5;
  --accent-hover: #4338ca;
  --accent-soft: rgba(79, 70, 229, .10);
  --text-strong: #0f172a;
  --text-muted: rgba(71, 85, 105, .78);
  --field-text: #1e293b;
  --page-bg: #e9edf5;
}

.root.dark {
  --accent: #818cf8;
  --accent-hover: #a5b4fc;
  --accent-soft: rgba(129, 140, 248, .16);
  --text-strong: #f1f5ff;
  --text-muted: rgba(203, 213, 225, .72);
  --field-text: #e8eefc;
  --page-bg: #070b16;
}

.root {
  min-height: 100vh;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 24px;
  background: var(--page-bg);
}

/* ══ 背景层 ════════════════════════════════════════════ */
.bg {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  transform: scale(1.06);
  opacity: .55;
}
.root.dark .bg { opacity: .26; filter: brightness(.7) saturate(.85); }

/* 彩色光斑 —— Liquid Glass 的关键：玻璃背后必须有流动色块 */
.bg-aurora {
  position: absolute;
  inset: -10%;
  pointer-events: none;
  background:
    radial-gradient(46% 38% at 16% 20%, rgba(120, 170, 255, .62), transparent 64%),
    radial-gradient(42% 34% at 84% 14%, rgba(186, 146, 255, .58), transparent 62%),
    radial-gradient(52% 44% at 74% 84%, rgba(104, 226, 214, .50), transparent 66%),
    radial-gradient(40% 34% at 22% 80%, rgba(255, 168, 200, .48), transparent 62%);
  filter: blur(46px) saturate(150%);
  animation: auroraDrift 26s ease-in-out infinite alternate;
}
.root.dark .bg-aurora {
  background:
    radial-gradient(46% 38% at 16% 20%, rgba(58, 92, 190, .70), transparent 64%),
    radial-gradient(42% 34% at 84% 14%, rgba(112, 76, 190, .66), transparent 62%),
    radial-gradient(52% 44% at 74% 84%, rgba(28, 132, 140, .58), transparent 66%),
    radial-gradient(40% 34% at 22% 80%, rgba(150, 62, 116, .52), transparent 62%);
  filter: blur(52px) saturate(135%);
}

@keyframes auroraDrift {
  0%   { transform: translate3d(-2%, -1%, 0) scale(1.02); }
  50%  { transform: translate3d(2%, 2%, 0) scale(1.06); }
  100% { transform: translate3d(-1%, 3%, 0) scale(1.03); }
}

.bg-veil {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(120% 90% at 50% 0%, rgba(255, 255, 255, .18), transparent 58%),
    linear-gradient(180deg, rgba(15, 23, 42, .04), rgba(15, 23, 42, .22));
}
.root.dark .bg-veil {
  background:
    radial-gradient(120% 90% at 50% 0%, rgba(99, 102, 241, .10), transparent 58%),
    linear-gradient(180deg, rgba(0, 0, 0, .30), rgba(0, 0, 0, .52));
}

/* ══ 工具栏 ════════════════════════════════════════════ */
.toolbar {
  position: absolute;
  top: 22px;
  right: 24px;
  z-index: 4;
  display: flex;
  gap: 9px;
  align-items: center;
  background: rgba(255, 255, 255, .40);
  border: 1px solid rgba(255, 255, 255, .62);
  border-radius: 999px;
  padding: 5px 12px 5px 7px;
  backdrop-filter: blur(22px) saturate(180%);
  -webkit-backdrop-filter: blur(22px) saturate(180%);
  box-shadow: 0 6px 20px rgba(15, 23, 42, .10), inset 0 1px 0 rgba(255, 255, 255, .8);
}
.root.dark .toolbar {
  background: rgba(255, 255, 255, .08);
  border-color: rgba(255, 255, 255, .14);
  box-shadow: 0 6px 20px rgba(0, 0, 0, .34), inset 0 1px 0 rgba(255, 255, 255, .12);
}
.toolbar-sep { width: 1px; height: 14px; background: currentColor; opacity: .16; color: var(--text-muted); }
.lang-select { width: 78px; }

.style-toggle {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px 3px 7px;
  border: none;
  border-radius: 999px;
  background: rgba(255, 255, 255, .34);
  color: var(--text-strong);
  font-size: 11.5px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color .18s;
}
.root.dark .style-toggle { background: rgba(255, 255, 255, .10); }
.style-toggle:hover { background: rgba(255, 255, 255, .58); }
.root.dark .style-toggle:hover { background: rgba(255, 255, 255, .18); }
.style-icon { font-size: 14px; opacity: .8; }

/* ══ 卡片容器 ══════════════════════════════════════════ */
.card-wrap { position: relative; z-index: 2; width: min(400px, 100%); }

/* 卡片背后的柔光，制造「悬浮玻璃」错觉 */
.card-glow {
  position: absolute;
  inset: -14% -10%;
  border-radius: 50%;
  background: radial-gradient(closest-side, rgba(120, 160, 255, .45), rgba(180, 140, 255, .22) 55%, transparent 76%);
  filter: blur(38px);
  pointer-events: none;
  opacity: 0;
  transition: opacity .4s;
}
.card-style-glass .card-glow { opacity: 1; }

.card {
  position: relative;
  border-radius: 22px;
  padding: 34px 30px 30px;
  transition: border-radius .35s, background .35s, box-shadow .35s;
}

.card-inner { position: relative; z-index: 1; }

/* ══ 风格 A：Liquid Glass ═══════════════════════════════ */
.card-style-glass .card {
  border-radius: 30px;
  background: linear-gradient(
    155deg,
    rgba(255, 255, 255, .34) 0%,
    rgba(255, 255, 255, .14) 44%,
    rgba(255, 255, 255, .22) 100%
  );
  border: 1px solid rgba(255, 255, 255, .48);
  backdrop-filter: blur(46px) saturate(195%) brightness(1.05);
  -webkit-backdrop-filter: blur(46px) saturate(195%) brightness(1.05);
  box-shadow:
    0 34px 74px -22px rgba(8, 14, 32, .52),
    0 10px 28px -10px rgba(8, 14, 32, .26),
    inset 0 1px 0 rgba(255, 255, 255, .88),
    inset 0 -1px 1px rgba(255, 255, 255, .26),
    inset 1px 0 0 rgba(255, 255, 255, .34),
    inset -1px 0 0 rgba(255, 255, 255, .16);
}

.card-style-glass.dark .card {
  background: linear-gradient(
    155deg,
    rgba(255, 255, 255, .12) 0%,
    rgba(255, 255, 255, .045) 46%,
    rgba(255, 255, 255, .085) 100%
  );
  border-color: rgba(255, 255, 255, .20);
  backdrop-filter: blur(46px) saturate(170%) brightness(1.14);
  -webkit-backdrop-filter: blur(46px) saturate(170%) brightness(1.14);
  box-shadow:
    0 34px 74px -20px rgba(0, 0, 0, .78),
    0 10px 28px -10px rgba(0, 0, 0, .44),
    inset 0 1px 0 rgba(255, 255, 255, .30),
    inset 0 -1px 1px rgba(255, 255, 255, .10),
    inset 1px 0 0 rgba(255, 255, 255, .16),
    inset -1px 0 0 rgba(255, 255, 255, .08);
}

/* 顶部弧形高光：玻璃受光面 */
.card-sheen {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  opacity: 0;
  transition: opacity .35s;
  background: radial-gradient(118% 58% at 50% -12%, rgba(255, 255, 255, .62), transparent 58%);
}
.card-style-glass .card-sheen { opacity: .9; }
.card-style-glass.dark .card-sheen {
  background: radial-gradient(118% 58% at 50% -12%, rgba(255, 255, 255, .26), transparent 58%);
}

/* 边缘渐变折射描边 */
.card-rim {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  pointer-events: none;
  opacity: 0;
  transition: opacity .35s;
  background: linear-gradient(
    142deg,
    rgba(255, 255, 255, .95) 0%,
    rgba(255, 255, 255, .14) 28%,
    rgba(255, 255, 255, .04) 52%,
    rgba(255, 255, 255, .20) 74%,
    rgba(255, 255, 255, .70) 100%
  );
  -webkit-mask: linear-gradient(#000 0 0) content-box, linear-gradient(#000 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#000 0 0) content-box, linear-gradient(#000 0 0);
  mask-composite: exclude;
}
.card-style-glass .card-rim { opacity: 1; }
.card-style-glass.dark .card-rim {
  background: linear-gradient(
    142deg,
    rgba(255, 255, 255, .48) 0%,
    rgba(255, 255, 255, .08) 30%,
    rgba(255, 255, 255, .02) 54%,
    rgba(255, 255, 255, .10) 76%,
    rgba(255, 255, 255, .34) 100%
  );
}

/* ══ 风格 B：柔和实心 ═══════════════════════════════════ */
.card-style-soft .card {
  background: rgba(255, 255, 255, .84);
  border: 1px solid rgba(255, 255, 255, .9);
  backdrop-filter: blur(22px) saturate(150%);
  -webkit-backdrop-filter: blur(22px) saturate(150%);
  box-shadow:
    0 24px 60px rgba(31, 39, 64, .14),
    0 1px 0 rgba(255, 255, 255, .95) inset;
}
.card-style-soft.dark .card {
  background: rgba(18, 24, 38, .82);
  border-color: rgba(255, 255, 255, .09);
  box-shadow:
    0 24px 60px rgba(0, 0, 0, .46),
    0 1px 0 rgba(255, 255, 255, .05) inset;
}
/* 柔和风格不需要光斑和光晕，背景更干净 */
.card-style-soft .bg-aurora { opacity: 0; }
.card-style-soft .card-glow { display: none; }

/* ══ 品牌区 ════════════════════════════════════════════ */
.brand { display: flex; align-items: center; gap: 13px; margin-bottom: 24px; }

.logo-badge {
  width: 46px;
  height: 46px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  color: #fff;
  background: linear-gradient(150deg, #6366f1, #8b5cf6 55%, #4f46e5);
  box-shadow:
    0 8px 20px -6px rgba(79, 70, 229, .58),
    inset 0 1px 0 rgba(255, 255, 255, .45),
    inset 0 -1px 0 rgba(0, 0, 0, .12);
}
.card-style-soft .logo-badge {
  background: var(--accent-soft);
  border: 1px solid var(--accent-soft);
  color: var(--accent);
  box-shadow: none;
}
.logo { width: 26px; height: 26px; }

.brand-text { min-width: 0; }
.name {
  font-size: 21px;
  font-weight: 700;
  letter-spacing: -.3px;
  line-height: 1.2;
  color: var(--text-strong);
}
.tagline {
  margin-top: 3px;
  font-size: 12px;
  line-height: 1.3;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 站点自定义标题胶囊 */
.site-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  margin: -8px 0 20px;
  padding: 4px 11px;
  border-radius: 999px;
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: .2px;
  color: var(--accent);
  background: var(--accent-soft);
}
.card-style-glass .site-chip {
  color: var(--text-strong);
  background: rgba(255, 255, 255, .34);
  border: 1px solid rgba(255, 255, 255, .45);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}
.card-style-glass.dark .site-chip {
  background: rgba(255, 255, 255, .12);
  border-color: rgba(255, 255, 255, .18);
}
.site-chip > span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ══ 表单 ══════════════════════════════════════════════ */
.form { margin-top: 4px; }
.field { margin-bottom: 15px; }

/* 基础输入框 */
.field :deep(.n-input) {
  border-radius: 13px;
  height: 47px;
  transition: border-color .18s, box-shadow .18s, background-color .18s;
}
.field :deep(.n-input__state-border) { display: none; }
.field :deep(.n-input .n-input__border),
.field :deep(.n-input .n-input__state-border) { box-shadow: none; }
.field :deep(.n-input__input-el) { font-size: 14px; color: var(--field-text); }
.field :deep(.n-input__placeholder) { color: var(--text-muted); }
.field-icon { opacity: .5; font-size: 17px; }

/* 柔和风格输入框 */
.card-style-soft .field :deep(.n-input) {
  background-color: rgba(255, 255, 255, .72);
  border: 1px solid rgba(148, 163, 184, .38);
}
.card-style-soft.dark .field :deep(.n-input) {
  background-color: rgba(255, 255, 255, .05);
  border-color: rgba(255, 255, 255, .12);
}
.card-style-soft .field :deep(.n-input:hover) { border-color: rgba(79, 70, 229, .45); }
.card-style-soft .field :deep(.n-input.n-input--focus) {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

/* 玻璃风格输入框：半透明 + 内高光，像嵌在玻璃里的凹槽 */
.card-style-glass .field :deep(.n-input) {
  background-color: rgba(255, 255, 255, .18);
  border: 1px solid rgba(255, 255, 255, .32);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .50),
    inset 0 -1px 0 rgba(255, 255, 255, .08),
    0 1px 2px rgba(15, 23, 42, .05);
}
.card-style-glass.dark .field :deep(.n-input) {
  background-color: rgba(255, 255, 255, .07);
  border-color: rgba(255, 255, 255, .16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .16),
    0 1px 2px rgba(0, 0, 0, .18);
}
.card-style-glass .field :deep(.n-input:hover) {
  border-color: rgba(255, 255, 255, .55);
  background-color: rgba(255, 255, 255, .26);
}
.card-style-glass.dark .field :deep(.n-input:hover) {
  border-color: rgba(255, 255, 255, .26);
  background-color: rgba(255, 255, 255, .11);
}
.card-style-glass .field :deep(.n-input.n-input--focus) {
  border-color: rgba(255, 255, 255, .85);
  background-color: rgba(255, 255, 255, .30);
  box-shadow:
    0 0 0 3.5px rgba(255, 255, 255, .26),
    inset 0 1px 0 rgba(255, 255, 255, .6);
}
.card-style-glass.dark .field :deep(.n-input.n-input--focus) {
  border-color: rgba(255, 255, 255, .45);
  background-color: rgba(255, 255, 255, .14);
  box-shadow:
    0 0 0 3.5px rgba(255, 255, 255, .12),
    inset 0 1px 0 rgba(255, 255, 255, .22);
}

/* 覆盖浏览器自动填充的蓝底 */
.field :deep(input:-webkit-autofill),
.field :deep(input:-webkit-autofill:hover),
.field :deep(input:-webkit-autofill:focus) {
  -webkit-text-fill-color: var(--field-text);
  transition: background-color 99999s ease-out 0s;
  caret-color: var(--field-text);
}
.card-style-soft .field :deep(input:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px rgba(255, 255, 255, .72) inset;
  box-shadow: 0 0 0 1000px rgba(255, 255, 255, .72) inset;
}
.card-style-soft.dark .field :deep(input:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px rgba(255, 255, 255, .05) inset;
  box-shadow: 0 0 0 1000px rgba(255, 255, 255, .05) inset;
}
.card-style-glass .field :deep(input:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px rgba(255, 255, 255, .20) inset;
  box-shadow: 0 0 0 1000px rgba(255, 255, 255, .20) inset;
}
.card-style-glass.dark .field :deep(input:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px rgba(255, 255, 255, .09) inset;
  box-shadow: 0 0 0 1000px rgba(255, 255, 255, .09) inset;
}

.eye {
  display: inline-flex;
  cursor: pointer;
  opacity: .5;
  font-size: 17px;
  transition: opacity .15s;
  outline: none;
}
.eye:hover, .eye:focus-visible { opacity: 1; }

.captcha-row { display: flex; gap: 12px; align-items: center; width: 100%; }
.captcha-box {
  width: 112px;
  height: 47px;
  flex-shrink: 0;
  border-radius: 13px;
  overflow: hidden;
  cursor: pointer;
  border: 1px solid rgba(148, 163, 184, .38);
  background: rgba(255, 255, 255, .72);
}
.card-style-glass .captcha-box {
  border-color: rgba(255, 255, 255, .32);
  background: rgba(255, 255, 255, .18);
}
.card-style-glass.dark .captcha-box {
  border-color: rgba(255, 255, 255, .16);
  background: rgba(255, 255, 255, .07);
}
.card-style-soft.dark .captcha-box {
  border-color: rgba(255, 255, 255, .12);
  background: rgba(255, 255, 255, .05);
}

/* ══ 底部行 + 按钮 ═════════════════════════════════════ */
.footrow {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin: 2px 0 18px;
  min-height: 24px;
  font-size: 13px;
}
.hint {
  color: var(--text-muted);
  white-space: nowrap;
  font-size: 12px;
}

.submit {
  height: 48px;
  border-radius: 14px;
  border: none;
  font-weight: 600;
  font-size: 14.5px;
  letter-spacing: .3px;
  transition: transform .16s, box-shadow .16s, background-color .16s;
}

/* 柔和风格按钮：实心品牌色 */
.card-style-soft .submit {
  color: #fff;
  background: var(--accent);
  box-shadow: 0 10px 24px -6px color-mix(in srgb, var(--accent) 60%, transparent);
}
.card-style-soft .submit:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 14px 28px -8px color-mix(in srgb, var(--accent) 65%, transparent);
}

/* 玻璃风格按钮：白色液态玻璃，深色文字 */
.card-style-glass .submit {
  color: #1b2334;
  background: linear-gradient(180deg, rgba(255, 255, 255, .96), rgba(255, 255, 255, .76));
  border: 1px solid rgba(255, 255, 255, .95);
  backdrop-filter: blur(24px) saturate(170%);
  -webkit-backdrop-filter: blur(24px) saturate(170%);
  box-shadow:
    0 12px 30px -10px rgba(8, 14, 32, .48),
    inset 0 1px 0 rgba(255, 255, 255, 1),
    inset 0 -2px 8px rgba(120, 140, 180, .22);
}
.card-style-glass .submit:hover {
  transform: translateY(-1px);
  background: linear-gradient(180deg, #fff, rgba(255, 255, 255, .88));
  box-shadow:
    0 16px 36px -12px rgba(8, 14, 32, .55),
    inset 0 1px 0 rgba(255, 255, 255, 1),
    inset 0 -2px 8px rgba(120, 140, 180, .26);
}
.card-style-glass.dark .submit {
  color: #f4f7ff;
  background: linear-gradient(180deg, rgba(129, 140, 248, .95), rgba(99, 102, 241, .82));
  border-color: rgba(255, 255, 255, .26);
  box-shadow:
    0 12px 30px -10px rgba(49, 46, 129, .78),
    inset 0 1px 0 rgba(255, 255, 255, .40),
    inset 0 -2px 8px rgba(30, 27, 90, .34);
}
.card-style-glass.dark .submit:hover {
  background: linear-gradient(180deg, rgba(165, 180, 252, .98), rgba(129, 140, 248, .9));
}
.submit:active { transform: translateY(0); }

/* ══ 响应式 / 降级 ═════════════════════════════════════ */
@media (max-width: 480px) {
  .card { padding: 28px 22px 24px; }
  .card-style-glass .card { border-radius: 26px; }
  .toolbar { top: 14px; right: 14px; gap: 7px; padding: 4px 10px 4px 6px; }
  .style-toggle span { display: none; }
  .brand { margin-bottom: 20px; }
}

@media (prefers-reduced-motion: reduce) {
  .submit, .field :deep(.n-input), .card { transition: none; }
  .bg-aurora { animation: none; }
}

/* 不支持 backdrop-filter 时退回不透明背景，避免文字看不清 */
@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
  .card-style-glass .card { background: rgba(255, 255, 255, .94); }
  .card-style-glass.dark .card { background: rgba(18, 24, 38, .95); }
  .card-style-glass .field :deep(.n-input) { background-color: rgba(255, 255, 255, .85); }
  .card-style-glass.dark .field :deep(.n-input) { background-color: rgba(255, 255, 255, .12); }
}
</style>
