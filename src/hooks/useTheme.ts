import type { GlobalThemeOverrides } from 'naive-ui'
import { computed, watch } from 'vue'
import { darkTheme, useOsTheme } from 'naive-ui'
import { useAppStore } from '@/store'
import { getThemePack, paletteToCssVars } from '@/utils/themePacks'

/**
 * 把主题包的色板注入到 documentElement 上的 CSS 变量。
 * 明暗模式与配色主题是两个独立维度，这里按当前明暗取对应色板。
 */
function applyThemeVars(packId: string, isDark: boolean) {
  const pack = getThemePack(packId)
  const palette = isDark ? pack.dark : pack.light
  const vars = paletteToCssVars(palette)
  const root = document.documentElement
  for (const [k, v] of Object.entries(vars))
    root.style.setProperty(k, v)
  // 供 CSS 做主题特判（例如某些主题需要单独调整光斑强度）
  root.dataset.themePack = pack.id
}

export function useTheme() {
  const appStore = useAppStore()

  const OsTheme = useOsTheme()

  const isDark = computed(() => {
    if (appStore.theme === 'auto')
      return OsTheme.value === 'dark'
    else
      return appStore.theme === 'dark'
  })

  const theme = computed(() => {
    return isDark.value ? darkTheme : undefined
  })

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    if (isDark.value) {
      return {
        common: {},
      }
    }
    return {}
  })

  watch(
    () => isDark.value,
    (dark) => {
      if (dark)
        document.documentElement.classList.add('dark')
      else
        document.documentElement.classList.remove('dark')
      // 明暗变化时重新注入对应色板
      applyThemeVars(appStore.themePack, dark)
    },
    { immediate: true },
  )

  // 配色主题切换时立即生效，无需刷新
  watch(
    () => appStore.themePack,
    (packId) => {
      applyThemeVars(packId, isDark.value)
    },
  )

  return { theme, themeOverrides }
}
