<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { NLayout, NLayoutContent, NLayoutSider, NSpace } from 'naive-ui'
import { getAppList } from '../common'
import { EditItem } from '@/views/home/components'
import { useAuthStore } from '@/store'
import { AppLoader, SModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { type SystemApp, systemApps } from '@/utils/cmn/builtInApps'
import { PanelCardType, PanelItemIconTypeEnum, PanelItemOpenMethodEnum } from '@/enums'
import { useTriggerUpdateItemData } from '@/hooks/useTriggerUpdateItemData'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
}>()

const componentName = ref('UserInfo')
const collapsed = ref(false)
const screenWidth = ref(0)
const isSmallScreen = ref(false)
const defaultTitle = t('appLauncher.title')
const title = ref('')
const height = ref('600px')
const maxWidth = ref('1000px')

const editItemInfoShow = ref(false)
const editItemInfo = ref<Panel.Info | null>({
  icon: {
    itemType: PanelItemIconTypeEnum.url,
    src: '',
  },
  title: '',
  cardType: PanelCardType.BuiltInApp,
  openMethod: PanelItemOpenMethodEnum.currentTab,
  url: '',
})

const apps = ref<SystemApp[]>(systemApps)

const authStore = useAuthStore()

const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
  },
})

function handleClickApp(item: SystemApp) {
  componentName.value = item.componentName
  if (isSmallScreen.value)
    collapsed.value = true
}

function getScreenWidth() {
  return window.innerWidth
}

function handleResize() {
  screenWidth.value = getScreenWidth()
  if (screenWidth.value < 640) {
    collapsed.value = true
    isSmallScreen.value = true
  }
  else {
    collapsed.value = false
    isSmallScreen.value = false
  }
}
function handleClickAppAdd(app: SystemApp) {
  editItemInfo.value = {
    icon: {
      itemType: PanelItemIconTypeEnum.online,
      text: app.icon,
    },
    title: app.name,
    cardType: PanelCardType.BuiltInApp,
    openMethod: PanelItemOpenMethodEnum.currentTab,
    url: '',
    description: t('iconItem.builtinApp'),
    expandParam: {
      systemAppName: app.componentName,
    },
  }
  editItemInfoShow.value = true
}
function handleEditSuccess() {
  useTriggerUpdateItemData().triggerUpdate()
  editItemInfoShow.value = false
}

// function handleMaxMinButton() {
//   height.value = 'calc(100vh - 70px)'
//   maxWidth.value = '100vw'
// }

// 触发更新菜单
watch(() => props.visible, (newValue) => {
  if (newValue) {
    apps.value = getAppList(apps.value, authStore.userInfo?.role || 0)
  }
}, { immediate: true })

onMounted(() => {
  window.addEventListener('resize', handleResize)
  handleResize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div>
    <SModal
      v-model:show="show"
      class="app-starter-modal"
      style="max-width: 1000px;border-radius: 1rem;"
      :style="{ maxWidth }"
      size="small"
      move
    >
      <!-- <template #header-extra>
        <NButton size="small" @click="handleMaxMinButton">
          最大化
        </NButton>
      </template> -->
      <template #header>
        <div class="flex items-center select-none">
          <div
            class="text-3xl cursor-pointer"
            style="color:var(--n-color-target)"
            @click="collapsed = !collapsed"
          >
            <SvgIcon class=" transition-all duration-500" :icon="collapsed ? 'tabler-layout-sidebar-right-collapse-filled' : 'tabler-layout-sidebar-left-collapse-filled'" />
          </div>
          <div class="ml-1" @click="collapsed = !collapsed">
            {{ title === '' ? defaultTitle : title }}
          </div>
        </div>
      </template>
      <div class="w-full h-full app-starter-modal-content">
        <NSpace vertical size="large" style="height: 100%;width: 100%;">
          <NLayout has-sider style="border-radius:0.75rem;">
            <NLayoutSider
              v-model:collapsed="collapsed"
              collapse-mode="width"
              :collapsed-width="0"
              :width="isSmallScreen ? '100%' : 240"
              style="height: 100%;"
              content-style="overflow: hidden"
            >
              <div class="w-full h-full dark:bg-[#2c2c32]">
                <div
                  class="app-side-surface p-[6px] rounded-xl overflow-auto"
                  :style="{
                    width: isSmallScreen ? '100%' : '220px',
                    minWidth: '200px',
                    height,
                  }"
                >
                  <div
                    v-for=" (item, index) in apps"
                    :key="index"
                    @click="handleClickApp(item)"
                  >
                    <div
                      class="app-menu-item p-[10px] rounded-lg mb-[3px] font-medium cursor-pointer flex items-center"
                    >
                      <div
                        class="flex items-center justify-center"
                        :style="{ color: componentName === item.componentName ? 'var(--n-color-target)' : '' }"
                      >
                        <div class="text-lg">
                          <SvgIcon :icon="item.icon" />
                        </div>
                        <span class="ml-2">{{ item.name }}</span>
                      </div>
                      <!-- 更多按钮 -->
                      <div
                        v-if="componentName === item.componentName"
                        class="ml-auto flex hover:text-[var(--n-color-target)]"
                        @click="handleClickAppAdd(item)"
                      >
                        <SvgIcon icon="fluent--square-add-16-regular" style="width: 16px;height: 16px;" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </NLayoutSider>
            <NLayoutContent :content-style="{ height }">
              <div class="app-content-surface rounded-xl overflow-auto transition-all duration-300 min-w-[300px] h-full" :class="(isSmallScreen && !collapsed) ? 'opacity-0' : 'opacity-100'">
                <AppLoader :component-name="componentName" class="h-full" />
              </div>
            </NLayoutContent>
          </NLayout>
        </NSpace>
      </div>
    </SModal>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfo" is-out-add @done="handleEditSuccess" />
  </div>
</template>

<style scoped>
.text-shadow {
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.4);
}

/* 侧栏菜单项：柔和半透明 hover / 选中，替代硬白底 */
.app-menu-item {
  color: var(--n-text-color);
  transition: background-color 0.18s, transform 0.18s;
  border: 1px solid transparent;
}
.app-menu-item:hover {
  background: rgba(127, 140, 180, 0.14);
  transform: translateX(2px);
}

/* Liquid Glass：设置弹窗内的侧栏与内容区 */
.app-side-surface {
  background: linear-gradient(155deg, rgba(255, 255, 255, .14) 0%, rgba(255, 255, 255, .06) 100%);
  border: 1px solid rgba(255, 255, 255, .18);
  backdrop-filter: blur(24px) saturate(175%);
  -webkit-backdrop-filter: blur(24px) saturate(175%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .22),
    inset 0 -1px 0 rgba(255, 255, 255, .06);
}

.app-content-surface {
  background: linear-gradient(155deg, rgba(255, 255, 255, .10) 0%, rgba(255, 255, 255, .04) 100%);
  border: 1px solid rgba(255, 255, 255, .14);
  backdrop-filter: blur(26px) saturate(180%);
  -webkit-backdrop-filter: blur(26px) saturate(180%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, .18),
    inset 0 -1px 0 rgba(255, 255, 255, .05);
}

@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
  .app-side-surface,
  .app-content-surface {
    background: rgba(24, 28, 40, .86);
  }
}
</style>

<style>
/* Liquid Glass：设置弹窗本体 —— 覆盖 Naive preset card 的纯色底
   注意：class 落在 NModal 外层容器上，n-card 在其内部，必须用后代选择器 */
.app-starter-modal .n-card {
  background: linear-gradient(155deg, rgba(255, 255, 255, .78) 0%, rgba(255, 255, 255, .58) 100%);
  border: 1px solid rgba(255, 255, 255, .85);
  backdrop-filter: blur(40px) saturate(190%);
  -webkit-backdrop-filter: blur(40px) saturate(190%);
  box-shadow:
    0 36px 80px -24px rgba(8, 14, 32, .48),
    0 10px 28px -12px rgba(8, 14, 32, .26),
    inset 0 1px 0 rgba(255, 255, 255, .95),
    inset 0 -1px 1px rgba(255, 255, 255, .45);
}

.dark .app-starter-modal .n-card {
  background: linear-gradient(155deg, rgba(255, 255, 255, .10) 0%, rgba(255, 255, 255, .045) 100%);
  border-color: rgba(255, 255, 255, .16);
  box-shadow:
    0 36px 80px -22px rgba(0, 0, 0, .80),
    0 10px 28px -12px rgba(0, 0, 0, .48),
    inset 0 1px 0 rgba(255, 255, 255, .18),
    inset 0 -1px 1px rgba(255, 255, 255, .06);
}

/* 弹窗内部布局容器透明化，让玻璃透出来 */
.app-starter-modal-content .n-layout,
.app-starter-modal-content .n-layout-sider,
.app-starter-modal-content .n-layout-content {
  background-color: transparent;
}

/* 不支持 backdrop-filter 时退回实心，保证可读 */
@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
  .app-starter-modal .n-card {
    background: #fff;
  }

  .dark .app-starter-modal .n-card {
    background: #1c202c;
  }
}
</style>
