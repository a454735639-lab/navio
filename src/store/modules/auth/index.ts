import { defineStore } from 'pinia'
import { getStorage, removeToken as hRemoveToken, setStorage } from './helper'
import { VisitMode } from '@/enums/auth'

// interface SessionResponse {
//   auth: boolean
// }

export interface AuthState {
  token: string | null
  userInfo: User.Info | null
  // session: SessionResponse | null
  visitMode: VisitMode
}

const defaultState: AuthState = {
  token: null,
  userInfo: null,
  visitMode: VisitMode.VISIT_MODE_LOGIN,
}

export const useAuthStore = defineStore('auth-store', {
  // 必须返回 defaultState 的副本：直接返回引用会让 Pinia 的响应式代理把
  // 模块级 defaultState 本身改成真实 token，导致 removeToken 无法复位。
  state: (): AuthState => {
    const stored = getStorage() as AuthState | null
    return stored ? { ...stored } : { ...defaultState }
  },

  actions: {
    setToken(token: string) {
      this.token = token
      this.saveStorage()
    },

    setUserInfo(userInfo: User.Info) {
      this.userInfo = userInfo
      this.saveStorage()
    },

    setVisitMode(visitMode: VisitMode) {
      this.visitMode = visitMode
      this.saveStorage()
    },

    saveStorage() {
      setStorage(this.$state)
    },

    removeToken() {
      // 逐字段复位，避免把被污染的对象引用整体拷回 state
      this.token = null
      this.userInfo = null
      this.visitMode = VisitMode.VISIT_MODE_LOGIN
      hRemoveToken()
    },
  },

})
