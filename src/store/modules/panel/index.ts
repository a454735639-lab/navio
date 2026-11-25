import { defineStore } from 'pinia'
import { defaultState, defaultStatePanelConfig, getLocalState, removeLocalState, setLocalState } from './helper'
import { router } from '@/router'
import type { PanelStateNetworkModeEnum } from '@/enums'
import { get as getUserConfig } from '@/api/panel/userConfig'
import { ErrorCode } from '@/enums/errorCode'
export const usePanelState = defineStore('panel', {
  state: (): Panel.State => getLocalState() || defaultState(),

  getters: {

  },

  actions: {
    setLeftSiderCollapsed(Collapsed: boolean) {
      this.leftSiderCollapsed = Collapsed
      // this.recordState()
    },

    setRightSiderCollapsed(Collapsed: boolean) {
      this.rightSiderCollapsed = Collapsed
      // this.recordState()
    },

    setNetworkMode(mode: PanelStateNetworkModeEnum) {
      this.networkMode = mode
      this.recordState()
    },

    // 获取云端（搭建的服务器）的面板配置
    async updatePanelConfigByCloud() {
      await getUserConfig<Panel.userConfig>().then((res) => {
        if (res.code === 0) {
          this.panelConfig = { ...defaultStatePanelConfig(), ...res.data.panel }
          // 旧版本默认 100/200ms，探测极易误判；升级后自动抬到一个可用值
          if (this.panelConfig.netModeChangeAuto && (this.panelConfig.netModeChangeAutoMaxPingMs ?? 0) < 300)
            this.panelConfig.netModeChangeAutoMaxPingMs = 600
          // 自动模式也保留手动切换入口：https 场景下浏览器禁止探测，用户需要能手动选
          if (this.panelConfig.netModeChangeAuto)
            this.panelConfig.netModeChangeButtonShow = true
        }
        else
          this.resetPanelConfig() // 重置恢复默认
        this.recordState()
      }).catch((res) => {
        if (res.code === ErrorCode.NoDataRecordFound) {
          this.resetPanelConfig() // 重置恢复默认
          this.recordState()
        }
      })
    },

    resetPanelConfig() {
      this.panelConfig = defaultStatePanelConfig()
      this.recordState()
    },

    resetState() {
      this.$state = defaultState()
      this.recordState()
    },

    setItemGroupCacheData(data: Panel.ItemGroupCardCacheData[]) {
      this.itemGroupCardCacheData = data
      this.recordState()
    },

    resetCacheData() {
      this.itemGroupCardCacheData = []
      this.recordState()
    },

    reset() {
      this.$state = defaultState()
      this.recordState()
    },

    // async refreshSpaceNoteList(spaceId: string) {
    //   await getListBySpaceNoteId<Common.ListResponse<SNote.InfoTree[]>>(spaceId).then((res) => {
    //     this.notesList = res.data.list
    //   })
    // },

    async reloadRoute(id?: number) {
      // this.recordState()
      await router.push({ name: 'AppletDialog', params: { aiAppletId: id } })
    },

    recordState() {
      setLocalState(this.$state)
    },

    removeState() {
      removeLocalState()
    },
  },
})
