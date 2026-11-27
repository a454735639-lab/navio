<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NDivider, NFlex, NForm, NFormItem, NInput, NSelect, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useAppStore, useAuthStore } from '@/store'
import { languageOptions } from '@/utils/defaultData'
import type { Language, Theme } from '@/store/modules/app/helper'
import { AccountSwitcherManager, RoundCardModal, STip } from '@/components/common/'
import { updateInfo, updatePassword } from '@/api/system/user'
import { updateLocalUserInfo } from '@/utils/cmn'
import { t } from '@/locales'
import { themePacks } from '@/utils/themePacks'

// const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
// const panelState = usePanelState()
// const accountManagerStore = useAccountManagerStore()

const ms = useMessage()
// const dialog = useDialog()

const languageValue = ref(appStore.language)
const themeValue = ref(appStore.theme)
const nickName = ref(authStore.userInfo?.name || '')
const isEditNickNameStatus = ref(false)
const formRef = ref<FormInst | null>(null)
const themeOptions: { label: string; key: string; value: Theme }[] = [
  { label: t('apps.userInfo.themeStyle.dark'), key: 'dark', value: 'dark' },
  { label: t('apps.userInfo.themeStyle.light'), key: 'light', value: 'light' },
  { label: t('apps.userInfo.themeStyle.auto'), key: 'Auto', value: 'auto' },
]
const updatePasswordModalState = ref({
  show: false,
  loading: false,
  form: {
    password: '',
    oldPassword: '',
    confirmPassword: '',
  },
})

const updatePasswordModalFormRules: FormRules = {
  oldPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  password: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  confirmPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
}

// async function logoutApi() {
//   // await logout()
//   if (panelState.panelConfig.accountSwitcherShow) {
//     for (let i = 0; i < accountManagerStore.accounts.length; i++) {
//       const element = accountManagerStore.accounts[i]
//       await logout(element.cToken).then(() => {
//         accountManagerStore.deleteAccount(element)
//       }).catch((res) => {
//         apiRespErrMsg(res)
//       })
//     }
//     accountManagerStore.clearAccount()
//   }

//   userStore.resetUserInfo()
//   authStore.removeToken()
//   panelState.removeState()
//   // appStore.removeToken()
//   ms.success(t('settingUserInfo.logoutSuccess'))
//   // router.push({ path: '/login' })
//   location.href = '/'// 强制刷新一下页面
// }

function handleSaveInfo() {
  updateInfo(nickName.value).then(({ code, msg }) => {
    updateLocalUserInfo()
    isEditNickNameStatus.value = false
  }).catch((error) => {
    ms.error(`${t('common.editFail')}:${error.msg}`)
  })
}

function handleUpdatePassword(e: MouseEvent) {
  e.preventDefault()
  formRef.value?.validate((errors) => {
    if (errors) {
      console.log(errors)
      return
    }

    if (updatePasswordModalState.value.form.password !== updatePasswordModalState.value.form.confirmPassword) {
      ms.error(t('settingUserInfo.confirmPasswordInconsistentMsg'))
      return
    }
    updatePasswordModalState.value.loading = true
    updatePassword(updatePasswordModalState.value.form.oldPassword, updatePasswordModalState.value.form.password).then(({ code, msg }) => {
      // 成功
      updatePasswordModalState.value.show = false
      ms.success(t('common.success'))
    }).finally(() => {
      updatePasswordModalState.value.loading = false
    }).catch((error) => {
      ms.error(`${t('common.serverError')}:${error.msg}`)
    })
  })
}

// function handleLogout() {
//   dialog.warning({
//     title: t('common.warning'),
//     content: t('settingUserInfo.confirmLogoutText'),
//     positiveText: t('common.confirm'),
//     negativeText: t('common.cancel'),
//     onPositiveClick: () => {
//       logoutApi()
//     },
//   })
// }

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
  location.reload()
}

function handleChangeTheme(value: Theme) {
  themeValue.value = value
  appStore.setTheme(value)
  // location.reload()
}

function handleChangeThemePack(packId: string) {
  appStore.setThemePack(packId)
}
</script>

<template>
  <div class="app-page p-2 h-full mb-10 overflow-auto">
    <!-- 账号信息 -->
    <NCard style="border-radius:10px" size="small">
      <template #header>
        <div class="text-slate-500 font-bold">
          {{ $t('apps.userInfo.currentAccount') }}
        </div>
      </template>
      <div>
        <div class="text-slate-500 font-bold">
          {{ $t('common.username') }}
        </div>
        {{ authStore.userInfo?.username }}
      </div>

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold">
          {{ $t('common.nikeName') }}
        </div>

        <div v-if="!isEditNickNameStatus">
          {{ authStore.userInfo?.name }}

          <NButton size="small" text type="success" @click="isEditNickNameStatus = !isEditNickNameStatus">
            {{ $t('common.edit') }}
          </NButton>
        </div>

        <div v-else class="flex items-center">
          <div class="max-w-[150px]">
            <NInput v-model:value="nickName" type="text" :placeholder="$t('common.inputPlaceholder')" />
          </div>
          <NButton size="small" quaternary type="success" @click="handleSaveInfo">
            {{ $t('common.save') }}
          </NButton>
        </div>
      </div>

      <NDivider style="margin: 10px 0;" dashed />
      <div>
        <NButton size="small" text type="success" @click="updatePasswordModalState.show = !updatePasswordModalState.show">
          {{ $t('settingUserInfo.updatePassword') }}
        </NButton>
      </div>
    </NCard>

    <!-- 基本设置 -->
    <NCard style="border-radius:10px" class="my-[10px]" size="small">
      <template #header>
        <div class="text-slate-500 font-bold">
          <NFlex align="center">
            {{ $t('common.setting') }}
            <STip :content="t('apps.userInfo.settingSaveLocal')" />
          </NFlex>
        </div>
      </template>
      <div>
        <div class="text-slate-500 font-bold">
          {{ $t('common.language') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="languageValue" size="small" :options="languageOptions" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold">
          {{ $t('apps.userInfo.theme') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="themeValue" size="small" :options="themeOptions" @update-value="handleChangeTheme" />
        </div>
      </div>

      <!-- 配色主题包：与明暗模式独立组合，点击即时生效 -->
      <div class="mt-[14px]">
        <div class="text-slate-500 font-bold mb-[8px] flex items-center">
          {{ $t('apps.userInfo.themePack') }}
          <STip :content="t('apps.userInfo.themePackTip')" />
        </div>
        <div class="theme-pack-grid">
          <button
            v-for="pack in themePacks" :key="pack.id"
            type="button"
            class="theme-pack-item"
            :class="{ 'is-active': appStore.themePack === pack.id }"
            :title="t(`apps.userInfo.themePackName.${pack.i18nKey}`)"
            @click="handleChangeThemePack(pack.id)"
          >
            <span
              class="theme-pack-swatch"
              :style="{ background: `linear-gradient(135deg, ${pack.preview[0]}, ${pack.preview[1]})` }"
            />
            <span class="theme-pack-label">{{ t(`apps.userInfo.themePackName.${pack.i18nKey}`) }}</span>
          </button>
        </div>
      </div>
    </NCard>

    <!-- 账号切换 -->
    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <template #header>
        <div class="text-slate-500 font-bold">
          <NFlex align="center">
            {{ t('deskModule.accountSwitcher.loggedInAccount') }}
            <STip :content="t('apps.userInfo.settingSaveLocal')" />
          </NFlex>
        </div>
      </template>

      <div>
        <AccountSwitcherManager />
      </div>
    </NCard>
    <!--
    <NCard style="border-radius:10px;margin-bottom: 10px;" class="my-[10px]" size="small">
      <NButton size="small" text type="error" @click="handleLogout">
        <template #icon>
          <SvgIcon icon="tabler:logout" />
        </template>
        {{ $t('settingUserInfo.logout') }}
      </NButton>
    </NCard> -->

    <RoundCardModal v-model:show="updatePasswordModalState.show" size="small" preset="card" style="width: 400px" :title="$t('settingUserInfo.updatePassword')">
      <NForm ref="formRef" :model="updatePasswordModalState.form" :rules="updatePasswordModalFormRules">
        <NFormItem path="oldPassword" :label="$t('settingUserInfo.oldPassword')">
          <NInput v-model:value="updatePasswordModalState.form.oldPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.oldPassword')" />
        </NFormItem>

        <NFormItem path="password" :label="$t('settingUserInfo.newPassword')">
          <NInput v-model:value="updatePasswordModalState.form.password" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.newPassword')" />
        </NFormItem>

        <NFormItem path="confirmPassword" :label="$t('settingUserInfo.confirmPassword')">
          <NInput v-model:value="updatePasswordModalState.form.confirmPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.confirmPassword')" />
        </NFormItem>
      </NForm>

      <template #footer>
        <div class="float-right">
          <NButton type="success" size="small" :loading="updatePasswordModalState.loading" @click="handleUpdatePassword">
            {{ $t('common.save') }}
          </NButton>
        </div>
      </template>
    </RoundCardModal>
  </div>
</template>

<style scoped>
/* 配色主题包选择器 */
.theme-pack-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(92px, 1fr));
  gap: 8px;
  max-width: 560px;
}

.theme-pack-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 6px;
  border-radius: 10px;
  border: 1px solid rgba(127, 136, 155, .22);
  background: rgba(255, 255, 255, .4);
  cursor: pointer;
  transition: border-color .18s, background .18s, transform .18s;
}

.theme-pack-item:hover {
  border-color: rgba(127, 136, 155, .42);
  transform: translateY(-1px);
}

.theme-pack-item.is-active {
  border-color: var(--nv-accent, #4f46e5);
  background: var(--nv-accent-soft, rgba(79, 70, 229, .1));
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--nv-accent, #4f46e5) 22%, transparent);
}

.theme-pack-swatch {
  width: 100%;
  height: 26px;
  border-radius: 7px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, .55), 0 1px 3px rgba(15, 23, 42, .14);
}

.theme-pack-label {
  font-size: 11px;
  font-weight: 500;
  color: inherit;
  opacity: .8;
  white-space: nowrap;
}

.theme-pack-item.is-active .theme-pack-label {
  opacity: 1;
  font-weight: 600;
}
</style>
