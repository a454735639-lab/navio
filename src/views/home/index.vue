<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, useDialog, useMessage } from 'naive-ui'
import { onMounted, onUnmounted, ref } from 'vue'
import { AppStarter, EditItem, ItemCards } from './components'
import { AccountSwitcher, Clock, SearchBox, SystemMonitor } from '@/components/deskModule'
import { Loading, SvgIcon } from '@/components/common'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { apiRespErrMsg, setTitle, updateLocalUserInfo } from '@/utils/cmn'
import { useTriggerUpdateItemData } from '@/hooks/useTriggerUpdateItemData'
import { getSiteStting } from '@/api/openness'
import { clearNetworkCache, clearRoleProbeCache, getNetworkStatus, probeRole, warmupNetworkMode } from './components/ItemCards/common'
import { defaultSetting } from '@/store/modules/app/helper'

import { useAccountManagerStore, useAppStore, useAuthStore, usePanelState } from '@/store'
import { PanelCardType, PanelPanelConfigStyleEnum, PanelStateNetworkModeEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { t } from '@/locales'

interface ItemGroup extends Panel.ItemGroupCardCacheData {

}

const ms = useMessage()
const dialog = useDialog()
const panelState = usePanelState()
const appStore = useAppStore()
const authStore = useAuthStore()
const accountManagerStore = useAccountManagerStore()
const triggerUpdateItemDatas = useTriggerUpdateItemData()

const scrollContainerRef = ref<HTMLElement | undefined>(undefined)

const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)

const itemCardsLoading = ref(false)

// const dropdownShow = ref(false)
const currentAddItenIconGroupId = ref<number | undefined>()
const settingModalShow = ref(false)

const items = ref<ItemGroup[]>([])
const filterItems = ref<ItemGroup[]>([])

let hasRunFooterJS = false

async function getGlobalConfig() {
  try {
    const { data } = await getSiteStting<GlobalSetting.SiteSetting>()
    const dcfg = defaultSetting().siteSetting
    if (data.title === '')
      data.title = dcfg.title
    if (data.loginBackgroundUrl === '')
      data.loginBackgroundUrl = dcfg.loginBackgroundUrl

    appStore.setSiteSetting(data)
    // console.log('全局配置', data)
  }
  catch (error) {
    apiRespErrMsg(error)
  }
}

let requestGeneration = 0
let activeRequests = 0

async function getList() {
  const generation = ++requestGeneration
  activeRequests++
  itemCardsLoading.value = true
  try {
    if (appStore.siteSetting.isDisableItemCardCache === false && panelState.itemGroupCardCacheData.length)
      filterItems.value = panelState.itemGroupCardCacheData

    const { data } = await getGroupList<Common.ListResponse<ItemGroup[]>>()
    if (generation !== requestGeneration)
      return
    items.value = data.list
    if (filterItems.value.length === 0)
      filterItems.value = items.value
    warmupNetworkMode(collectLanUrls(data.list), panelState.panelConfig.netModeChangeAutoMaxPingMs)

    // Bounded workers make each completed group visible without request storms.
    let cursor = 0
    const worker = async () => {
      while (cursor < data.list.length) {
        const index = cursor++
        const group = data.list[index]
        if (!group.id)
          continue
        try {
          const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(group.id)
          if (generation === requestGeneration && res.code === 0) {
            items.value[index].items = res.data.list
            if (filterItems.value === items.value || filterItems.value.length === 0)
              filterItems.value = items.value
            if (!appStore.siteSetting.isDisableItemCardCache)
              panelState.setItemGroupCacheData(items.value)
          }
        }
        catch (error) { apiRespErrMsg(error) }
      }
    }
    await Promise.all(Array.from({ length: Math.min(4, data.list.length) }, worker))
    if (generation === requestGeneration) {
      if (!appStore.siteSetting.isDisableItemCardCache)
        panelState.setItemGroupCacheData(items.value)
      else if (panelState.itemGroupCardCacheData.length)
        panelState.setItemGroupCacheData([])
    }
  }
  catch (error) { if (generation === requestGeneration) apiRespErrMsg(error) }
  finally {
    activeRequests--
    if (activeRequests === 0)
      itemCardsLoading.value = false
  }
}

function replaceItemLocally(item: Panel.ItemInfo) {
  const target = items.value.find(g => g.id === item.itemIconGroupId)
  for (const group of items.value) {
    const index = group.items?.findIndex(i => i.id === item.id) ?? -1
    if (index >= 0) group.items!.splice(index, 1)
  }
  if (target) {
    target.items ||= []
    const oldIndex = target.items.findIndex(i => i.id === item.id)
    if (oldIndex >= 0) target.items.splice(oldIndex, 1, item)
    else target.items.push(item)
  }
  // Keep the current filtered/sorted array and its group objects intact.
  for (const group of filterItems.value) {
    const index = group.items?.findIndex(i => i.id === item.id) ?? -1
    if (index >= 0) {
      if (group.id === item.itemIconGroupId) group.items!.splice(index, 1, item)
      else group.items!.splice(index, 1)
    }
  }
  panelState.setItemGroupCacheData(items.value)
}

function handleEditSuccess(item: Panel.ItemInfo) {
  replaceItemLocally(item)
}

// Keep drag-sort refresh scoped to the affected group.
async function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  try {
    const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId)
    if (res.code === 0 && items.value[itemIconGroupIndex]) {
      items.value[itemIconGroupIndex].items = res.data.list
      if (!appStore.siteSetting.isDisableItemCardCache)
        panelState.setItemGroupCacheData(items.value)
    }
  }
  catch (error) { apiRespErrMsg(error) }
}


  for (const group of items.value) {
    const index = group.items?.findIndex(i => i.id === item.id) ?? -1
    if (index >= 0) group.items!.splice(index, 1)
  }
  for (const group of filterItems.value) {
    const index = group.items?.findIndex(i => i.id === item.id) ?? -1
    if (index >= 0) group.items!.splice(index, 1)
  }
  panelState.setItemGroupCacheData(items.value)
}


function handleChangeNetwork(mode: PanelStateNetworkModeEnum) {
  panelState.setNetworkMode(mode)
  if (mode === PanelStateNetworkModeEnum.lan)
    ms.success(t('panelHome.changeToLanModelSuccess'))

  else
    ms.success(t('panelHome.changeToWanModelSuccess'))
}

// 删除项目
function deleteItem(itemInfo: Panel.ItemInfo) {
  dialog.warning({
    title: t('common.warning'),
    content: t('common.deleteConfirmByName', { name: itemInfo.title }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      deletes([itemInfo.id as number]).then(({ code, msg }) => {
        if (code === 0) {
          ms.success(t('common.deleteSuccess'))
          removeItemLocally(itemInfo)
        }
        else {
          ms.error(`${t('common.deleteFail')}:${msg}`)
        }
      }).catch((error) => {
        apiRespErrMsg(error)
      })
    },
  })
}

// 结束拖拽
// function handleEndDrag(event: any, itemIconGroup: Panel.ItemGroupInfo) {
//   // console.log(event)
//   // console.log(items.value)
// }

function handleSaveSort(itemGroup: ItemGroup) {
  const saveItems: Common.SortItemRequest[] = []
  if (itemGroup.items) {
    for (let i = 0; i < itemGroup.items.length; i++) {
      const element = itemGroup.items[i]
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
      })
    }

    saveSort({ itemIconGroupId: itemGroup.id as number, sortItems: saveItems }).then(({ code, msg }) => {
      if (code === 0) {
        ms.success(t('common.saveSuccess'))
        itemGroup.sortStatus = false
      }
      else {
        ms.error(`${t('common.saveFail')}:${msg}`)
      }
    }).catch((error) => {
      apiRespErrMsg(error)
    })
  }
}

function runFooterJSCode() {
  if (hasRunFooterJS)
    return

  // 获取类名为 'footer' 的元素
  const footerElements = document.querySelectorAll('.custom-footer')
  footerElements.forEach((footer) => {
    const scriptElements = footer.querySelectorAll('script')

    scriptElements.forEach((script) => {
      // 创建一个新的 script 标签
      const newScript = document.createElement('script')

      if (script.src) {
        // 如果 script 标签有 src 属性
        newScript.src = script.src
        newScript.type = script.type
        // newScript.onload = () => console.log(`Loaded script: ${script.src}`)
        newScript.onerror = () => console.error(`Failed to load script: ${script.src}`)
      }
      else {
        // 如果 script 标签是内联脚本
        newScript.textContent = script.textContent
      }

      // 将新的 script 标签添加到文档中
      document.body.appendChild(newScript)

      // 如果不需要保留原有的 script 标签，可以移除它
      if (script.parentNode) {
        script.parentNode.removeChild(script)
      }
    })
  })

  hasRunFooterJS = true
}


function collectLanUrls(list: ItemGroup[] | undefined): string[] {
  const out: string[] = []
  ;(list || []).forEach((g) => {
    ;(g.items || []).forEach((it: Panel.ItemInfo) => {
      if (it && it.lanUrl)
        out.push(it.lanUrl)
      if (it && it.url)
        out.push(it.url)
    })
  })
  return out
}


function handleVisibilityChange() {
  if (document.visibilityState === 'visible')
    handleNetworkChange()
}

// 网络状态变化 / 页面回到前台时，清缓存并重新判定
let networkTimer: ReturnType<typeof setInterval> | undefined
let statusTimer: ReturnType<typeof setInterval> | undefined

const netStatus = ref(getNetworkStatus())

function refreshNetStatus() {
  netStatus.value = getNetworkStatus()
}

function labelNetwork() {
  if (netStatus.value.mode === 'tailnet')
    return netStatus.value.ms !== null ? `TS ${netStatus.value.ms}ms` : 'TS'
  if (netStatus.value.mode === 'lan')
    return netStatus.value.ms !== null ? `LAN ${netStatus.value.ms}ms` : 'LAN'
  return 'WAN'
}

async function handleNetworkChange() {
  clearNetworkCache()
  clearRoleProbeCache()
  await Promise.all([
    probeRole('lan', panelState.panelConfig.netProbeLanUrl, true),
    probeRole('ts', panelState.panelConfig.netProbeTsUrl, true),
  ]).catch(() => {})
  refreshNetStatus()
  warmupNetworkMode(collectLanUrls(items.value), panelState.panelConfig.netModeChangeAutoMaxPingMs)
}

onMounted(async () => {
  itemCardsLoading.value = true // 加载中
  // 获取全局配置
  await getGlobalConfig()

  // 更新用户信息
  await updateLocalUserInfo()
  getList()

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)

  // 更新同步云端配置
  await panelState.updatePanelConfigByCloud()

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)

  // 首屏就探一次网络位置（lan/ts 探针），点击卡片零等待
  handleNetworkChange()

  // 后台静默刷新：页面可见时每 30s 复探，网络切换后 30s 内必纠正
  networkTimer = setInterval(() => {
    if (document.visibilityState === 'visible') {
      handleNetworkChange()
      refreshNetStatus()
    }
  }, 30_000)

  statusTimer = setInterval(refreshNetStatus, 2_000)

  // 监听触发更新列表
  triggerUpdateItemDatas.onTriggered(getList)
  runFooterJSCode()

  // 内外网自动识别：网络恢复、页面回到前台时重新探测
  window.addEventListener('online', handleNetworkChange)
  window.addEventListener('offline', handleNetworkChange)
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  keyword = keyword?.trim()
  if (keyword !== '' && panelState.panelConfig.searchBoxSearchIcon) {
    const filteredData: ItemGroup[] = []
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        return (
          item.title.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.url.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.description?.toLowerCase().includes(keyword?.toLowerCase() ?? '')
        )
      })
      if (element && element.length > 0)
        filteredData.push({ items: element, hoverStatus: false, cardStyle: items.value[i].cardStyle, title: items.value[i].title })
    }
    filterItems.value = filteredData
    // console.log('搜索模式开启')
  }
  else {
    // console.log('搜索模式关闭')
    filterItems.value = items.value
  }
}

function handleSetHoverStatus(groupIndex: number, hoverStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].hoverStatus = hoverStatus
}

function handleSetSortStatus(groupIndex: number, sortStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus) {
    // 单独更新组
    if (items.value[groupIndex] && items.value[groupIndex].id)
      updateItemIconGroupByNet(groupIndex, items.value[groupIndex].id as number)
  }
}

function toggleEditPanel(item: Panel.ItemInfo | null, groupId?: number) {
  editItemInfoData.value = item || null
  currentAddItenIconGroupId.value = groupId
  editItemInfoShow.value = true
}
onUnmounted(() => {
  if (networkTimer)
    clearInterval(networkTimer)
  if (statusTimer)
    clearInterval(statusTimer)
  window.removeEventListener('online', handleNetworkChange)
  window.removeEventListener('offline', handleNetworkChange)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

</script>

<template>
  <div class="w-full h-full sun-main">
    <div
      class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(0,0,0,${panelState.panelConfig.backgroundMaskNumber})` }" />
    <div ref="scrollContainerRef" class="scroll-container absolute w-full h-full overflow-auto">
      <div
        class="p-2.5 mx-auto"
        :style="{
          marginTop: `calc(${panelState.panelConfig.marginTop}% + 70px)`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <!-- 头 -->
        <div
          v-if="panelState.panelConfig.clockShow || panelState.panelConfig.logoShow || panelState.panelConfig.searchBoxShow"
          class="mx-[auto] w-[80%]"
        >
          <!-- LOGO及时钟组件 -->
          <div v-if="panelState.panelConfig.clockShow || panelState.panelConfig.logoShow" class="flex mx-[auto] items-center justify-center text-white">
            <div v-if="panelState.panelConfig.logoShow" class="logo">
              <span class="text-2xl md:text-6xl font-bold text-shadow">
                {{ panelState.panelConfig.logoText }}
              </span>
            </div>
            <div v-if="panelState.panelConfig.clockShow && panelState.panelConfig.logoShow" class="divider text-base lg:text-2xl mx-[10px]">
              |
            </div>
            <div v-if="panelState.panelConfig.clockShow" class="text-shadow">
              <Clock :hide-second="!panelState.panelConfig.clockShowSecond" />
            </div>
          </div>

          <!-- 搜索栏组件 -->
          <div v-if="panelState.panelConfig.searchBoxShow" class="flex mt-[20px] mx-auto sm:w-full lg:w-[80%]">
            <SearchBox
              :background="panelState.panelConfig.searchBoxBackground"
              :text-color="panelState.panelConfig.searchBoxTextColor"
              :border-color="panelState.panelConfig.searchBoxBorderColor"
              @item-search="itemFrontEndSearch"
            />
          </div>
        </div>

        <!-- 应用盒子 -->
        <div id="item-card-box" :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="item-list group-glass mt-[28px] group-enter"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <div v-show="itemCardsLoading" class="mt-20">
            <Loading />
          </div>

          <!-- 组纵向排列 -->
          <div
            v-for="(itemGroup, itemGroupIndex) in filterItems" :id="`item-group-${itemGroup.onlyName}`"
            :key="itemGroupIndex"
            class="item-list group-glass mt-[28px] group-enter"
            :style="{ animationDelay: `${Math.min(itemGroupIndex, 8) * 60}ms` }"
            :class="`item-group-${itemGroup.onlyName} item-group-index-${itemGroupIndex} ${itemGroup.sortStatus ? (`border shadow-[0_8px_30px_rgba(0,0,0,0.3)] p-[16px] rounded-2xl `) : ''}`"
            @mouseenter="handleSetHoverStatus(itemGroupIndex, true)"
            @mouseleave="handleSetHoverStatus(itemGroupIndex, false)"
          >
            <!-- 分组标题 -->
            <div class="text-white text-[15px] font-semibold mb-[16px] ml-[10px] flex items-center tracking-wide">
              <span class="group-title text-shadow">
                {{ itemGroup.title }}
              </span>
              <div
                v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
                class="group-buttons ml-2 delay-100 transition-opacity flex"
                :class="itemGroup.hoverStatus ? 'opacity-100' : 'opacity-0'"
              >
                <span class="mr-2 cursor-pointer" :title="t('common.add')" @click="toggleEditPanel(null, itemGroup.id)">
                  <SvgIcon class="text-white font-xl" icon="typcn:plus" />
                </span>
                <span class="mr-2 cursor-pointer " :title="t('common.sort')" @click="handleSetSortStatus(itemGroupIndex, !itemGroup.sortStatus)">
                  <SvgIcon class="text-white font-xl" icon="ri:drag-drop-line" />
                </span>
              </div>
            </div>

            <!-- 编辑栏 -->
            <div v-if="itemGroup.sortStatus" class="flex my-4">
              <NButton color="#2a2a2a6b" size="small" @click="handleSaveSort(itemGroup)">
                <template #icon>
                  <SvgIcon class="text-white font-xl" icon="material-symbols:save" />
                </template>
                <div class="text-white">
                  {{ $t('common.saveSort') }}
                </div>
              </NButton>
            </div>

            <!-- 详情图标 -->
            <template v-if="itemGroup.items && (!itemGroup.cardStyle || itemGroup.cardStyle.style === PanelPanelConfigStyleEnum.info)">
              <VueDraggable
                v-model="itemGroup.items" item-key="sort" :animation="300"
                class="icon-info-box"
                :disabled="!itemGroup.sortStatus"
              >
                <div
                  v-for="item, index in itemGroup.items"
                  :id="item.onlyName ? `item-card-name-${item.onlyName}` : ''"
                  :key="index"
                  :class="item.onlyName ? `item-card-name-${item.onlyName}` : ''"
                >
                  <ItemCards
                    :sort-status="itemGroup.sortStatus || false"
                    :class="itemGroup.sortStatus ? 'cursor-move' : (item.url !== '' || item.lanUrl !== '' ? 'cursor-pointer' : '')"
                    :card-info="item"
                    :icon-text-color="itemGroup.cardStyle.textColor"
                    :icon-text-info-hide-description="itemGroup.cardStyle.textInfoHideDescription || false"
                    :icon-text-icon-hide-title="itemGroup.cardStyle.textIconHideTitle || false"
                    :card-style="PanelPanelConfigStyleEnum.info"
                    :card-type="PanelCardType.Normal"
                    @delete="() => deleteItem(item)"
                    @edit="() => toggleEditPanel(item)"
                  />
                </div>

                <div v-if="itemGroup.items.length === 0" class="not-drag">
                  <ItemCards
                    :sort-status="itemGroup.sortStatus || false"
                    class="cursor-pointer"
                    :card-info="{ icon: { itemType: 3, text: 'subway:add' }, title: t('common.add'), url: '', openMethod: 0, cardType: PanelCardType.Normal }"
                    :icon-text-color="itemGroup.cardStyle.textColor"
                    :icon-text-info-hide-description="itemGroup.cardStyle.textInfoHideDescription || false"
                    :icon-text-icon-hide-title="itemGroup.cardStyle.textIconHideTitle || false"
                    :card-style="PanelPanelConfigStyleEnum.info"
                    :card-type="PanelCardType.Normal"
                    @click="toggleEditPanel(null, itemGroup.id)"
                  />
                </div>
              </VueDraggable>
            </template>

            <!-- APP图标宫型盒子 -->
            <template v-if="itemGroup.items && itemGroup.cardStyle.style === PanelPanelConfigStyleEnum.icon">
              <VueDraggable
                v-model="itemGroup.items" item-key="sort" :animation="300"
                class="icon-small-box"
                filter=".not-drag"
                :disabled="!itemGroup.sortStatus"
              >
                <div
                  v-for="item, index in itemGroup.items"
                  :id="item.onlyName ? `item-card-name-${item.onlyName}` : ''"
                  :key="index"
                  :class="item.onlyName ? `item-card-name-${item.onlyName}` : ''"
                >
                  <ItemCards
                    :sort-status="itemGroup.sortStatus || false"
                    :class="itemGroup.sortStatus ? 'cursor-move' : (item.url !== '' || item.lanUrl !== '' ? 'cursor-pointer' : '')"
                    :card-info="item"
                    :icon-text-color="itemGroup.cardStyle.textColor"
                    :icon-text-info-hide-description="itemGroup.cardStyle.textInfoHideDescription || false"
                    :icon-text-icon-hide-title="itemGroup.cardStyle.textIconHideTitle || false"
                    :card-style="PanelPanelConfigStyleEnum.small"
                    :card-type="PanelCardType.Normal"
                    @delete="() => deleteItem(item)"
                    @edit="() => toggleEditPanel(item)"
                  />
                </div>

                <div v-if="itemGroup.items.length === 0" class="not-drag">
                  <ItemCards
                    :sort-status="itemGroup.sortStatus || false"
                    class="cursor-pointer"
                    :card-info="{ icon: { itemType: 3, text: 'subway:add' }, title: t('common.add'), url: '', openMethod: 0, cardType: PanelCardType.Normal }"
                    :icon-text-color="itemGroup.cardStyle.textColor"
                    :icon-text-info-hide-description="itemGroup.cardStyle.textInfoHideDescription || false"
                    :icon-text-icon-hide-title="itemGroup.cardStyle.textIconHideTitle || false"
                    :card-style="PanelPanelConfigStyleEnum.small"
                    :card-type="PanelCardType.Normal"
                    @click="toggleEditPanel(null, itemGroup.id)"
                  />
                </div>
              </vuedraggable>
            </template>
          </div>
        </div>
        <div class="mt-5 footer custom-footer" v-html="panelState.panelConfig.footerHtml" />
      </div>
    </div>

    <!-- 悬浮按钮 -->
    <div
      class="fixed-element bg-[rgba(30,34,48,0.55)] rounded-xl shadow-[0_8px_28px_rgba(0,0,0,0.28)] p-[3px] flex backdrop-blur-xl border border-white/10"
    >
      <!-- 自动模式：延迟显示（与系统入口同一排，点击立即重新探测） -->
      <div
        v-if="panelState.panelConfig.netModeChangeAuto"
        class="float-btn flex items-center justify-center text-[11px] font-semibold select-none !w-auto px-3 whitespace-nowrap"
        :title="t('apps.baseSettings.netModeChangeAuto')"
        @click="handleNetworkChange(); refreshNetStatus()"
      >
        <span :class="netStatus.mode === 'lan' ? 'text-green-400' : (netStatus.mode === 'tailnet' ? 'text-blue-400' : 'text-gray-300')">
          {{ labelNetwork() }}
        </span>
      </div>

      <div
        v-if="panelState.networkMode === PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow && !panelState.panelConfig.netModeChangeAuto"
        class="float-btn flex items-center justify-center"
        :title="t('panelHome.changeToWanModel')"
        @click="handleChangeNetwork(PanelStateNetworkModeEnum.wan)"
      >
        <SvgIcon class="text-white font-xl float-btn-icon" icon="material-symbols:lan-outline-rounded" />
      </div>

      <div
        v-if="panelState.networkMode !== PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow && !panelState.panelConfig.netModeChangeAuto"
        class="float-btn flex items-center justify-center"
        :title="t('panelHome.changeToLanModel')"
        @click="handleChangeNetwork(PanelStateNetworkModeEnum.lan)"
      >
        <SvgIcon class="text-white font-xl float-btn-icon" icon="mdi:wan" />
      </div>

      <!-- Tailscale 模式：https 页面无法探测时，一键走 100.x 地址 -->
      <div
        v-if="panelState.networkMode !== PanelStateNetworkModeEnum.tailnet && panelState.panelConfig.netModeChangeButtonShow && !panelState.panelConfig.netModeChangeAuto"
        class="float-btn flex items-center justify-center text-[11px] font-bold text-white"
        :title="t('panelHome.useTailnet')"
        @click="handleChangeNetwork(PanelStateNetworkModeEnum.tailnet)"
      >
        TS
      </div>

      <!-- 账号管理器切换 -->
      <template
        v-if="accountManagerStore.accounts.length > 1 && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
      >
        <AccountSwitcher>
          <div
            class="float-btn flex items-center justify-center"
          >
            <SvgIcon class="text-white font-xl float-btn-icon" icon="ph--user-switch-bold" />
          </div>
        </AccountSwitcher>
      </template>

      <!-- 应用启动器 -->
      <div
        v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
        class="float-btn flex items-center justify-center"
        @click="settingModalShow = !settingModalShow"
      >
        <SvgIcon class="text-white font-xl float-btn-icon" icon="majesticons-applications" />
      </div>

      <div
        v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC"
        class="float-btn flex items-center justify-center"
        :title="$t('panelHome.goToLogin')"
        @click="$router.push('/login')"
      >
        <SvgIcon class="text-white font-xl float-btn-icon" icon="material-symbols:account-circle" />
      </div>
      <AppStarter v-model:visible="settingModalShow" />
      <!-- <Setting v-model:visible="settingModalShow" /> -->
    </div>

    <NBackTop
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div class="shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
        <NButton color="#2a2a2a6b">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />
  </div>
</template>

<style>
body,
html {
  overflow: hidden;
  background-color: rgb(54, 54, 54);
}
</style>

<style scoped>
.mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.sun-main {
  user-select: none;
}

.cover {
  position:fixed;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background: url(@/assets/start_sky.jpg) no-repeat; */

  transform: scale(1.05);
}

.text-shadow {
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.45);
}

.app-icon-text-shadow {
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.5);
}

.fixed-element {
  position: fixed;
  /* left:10px; */
  top:20px;
  /* 将元素固定在屏幕上 */
  right: 20px;
  /* 距离屏幕顶部的距离 */
  /* bottom: 50px; */
  /* 距离屏幕左侧的距离 */
}

/* Pass visuals: 分组毛玻璃底板 —— 给每个分组一层可见但不抢眼的容器 */
.group-glass {
  position: relative;
  padding: 18px 20px 20px;
  margin-left: -20px;
  margin-right: -20px;
  border-radius: 18px;
  background: rgba(24, 28, 40, 0.28);
  border: 1px solid rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(16px) saturate(115%);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.16);
}

.item-list {
  transition: background-color 0.25s, box-shadow 0.25s;
}

/* Pass motion: 分组依次淡入，仅首次加载，不干扰拖拽 */
.group-enter {
  animation: groupFadeIn 0.42s cubic-bezier(0.22, 0.61, 0.36, 1) both;
}

@keyframes groupFadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .group-enter {
    animation: none;
  }
}

.icon-info-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.icon-small-box {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(76px, 1fr));
  gap: 20px;

}

.float-btn {
  width: 42px;
  height: 42px;
  cursor: pointer;
  border-radius: 8px;
  transition: background-color 0.2s, transform 0.2s;
}

.float-btn:hover {
  background-color: rgba(255, 255, 255, 0.16);
  transform: translateY(-1px);
}

.float-btn-icon {
  width: 21px;
  height: 21px;
}

@media (max-width: 500px) {
  .icon-info-box{
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}
</style>
