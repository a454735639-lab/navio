import { ref, watch } from 'vue'

// 定义全局状态
const updateCounter = ref(0) // 使用计数器来确保每次触发都有效

// 提供触发函数
export const useTriggerUpdateItemData = () => {
  const triggerUpdate = () => {
    // console.log('主动触发更新')
    updateCounter.value++ // 每次调用递增计数器
  }

  const onTriggered = (callback: () => void) => {
    // 监听计数器变化并执行回调；返回 stop 句柄供调用方在卸载时清理，
    // 避免首页反复挂载后残留多个 watch 持有旧闭包、重复触发请求。
    const stop = watch(() => updateCounter.value, () => {
      callback()
    })
    return stop
  }

  return { triggerUpdate, onTriggered }
}
