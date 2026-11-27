/**
 * 主题包（Theme Pack）
 *
 * 设计思路：把「明暗模式」和「配色主题」拆成两个独立维度。
 * - 明暗模式：light / dark / auto（沿用原有 appStore.theme）
 * - 配色主题：这里的 themePack，只负责色相，不负责明暗
 *
 * 这样用户可以用「深色 + 极光」或「浅色 + 极光」，组合自由。
 *
 * 每个主题只定义一套 CSS 变量，运行时注入到 documentElement，
 * 所有玻璃/按钮/高光都从变量取值，因此新增主题不需要改组件样式。
 */

/** 单个主题在某一明暗模式下的色板 */
export interface ThemePalette {
  /** 主强调色（按钮、focus 环、选中态） */
  accent: string
  /** 强调色 hover */
  accentHover: string
  /** 强调色柔和底（chip、浅底按钮） */
  accentSoft: string
  /** 登录页 logo 徽章渐变 */
  logoGradient: string
  /** 玻璃底色渐变（上 -> 下） */
  glassTop: string
  glassBottom: string
  /** 玻璃边框 */
  glassBorder: string
  /** 玻璃顶部高光 */
  glassSheen: string
  /** 玻璃边缘折射（亮端） */
  glassRim: string
  /** 极光光斑四色（左上、右上、右下、左下） */
  auroraA: string
  auroraB: string
  auroraC: string
  auroraD: string
  /** 页面底色 */
  pageBg: string
  /** 正文强色 / 弱色 */
  textStrong: string
  textMuted: string
}

export interface ThemePack {
  /** 唯一标识，写入 localStorage */
  id: string
  /** i18n 文案键，如 themePack.sakura */
  i18nKey: string
  /** 选择器里的预览色（两个色块） */
  preview: [string, string]
  light: ThemePalette
  dark: ThemePalette
}

/* ── 1. 靛蓝（默认，接近 Apple 系统色） ───────────────────── */
const indigo: ThemePack = {
  id: 'indigo',
  i18nKey: 'indigo',
  preview: ['#6366f1', '#a5b4fc'],
  light: {
    accent: '#4f46e5',
    accentHover: '#4338ca',
    accentSoft: 'rgba(79, 70, 229, .10)',
    logoGradient: 'linear-gradient(150deg, #6366f1, #8b5cf6 55%, #4f46e5)',
    glassTop: 'rgba(255, 255, 255, .34)',
    glassBottom: 'rgba(255, 255, 255, .14)',
    glassBorder: 'rgba(255, 255, 255, .48)',
    glassSheen: 'rgba(255, 255, 255, .62)',
    glassRim: 'rgba(255, 255, 255, .95)',
    auroraA: 'rgba(120, 170, 255, .62)',
    auroraB: 'rgba(186, 146, 255, .58)',
    auroraC: 'rgba(104, 226, 214, .50)',
    auroraD: 'rgba(255, 168, 200, .48)',
    pageBg: '#e9edf5',
    textStrong: '#0f172a',
    textMuted: 'rgba(71, 85, 105, .78)',
  },
  dark: {
    accent: '#818cf8',
    accentHover: '#a5b4fc',
    accentSoft: 'rgba(129, 140, 248, .16)',
    logoGradient: 'linear-gradient(150deg, #6366f1, #8b5cf6 55%, #4f46e5)',
    glassTop: 'rgba(255, 255, 255, .12)',
    glassBottom: 'rgba(255, 255, 255, .045)',
    glassBorder: 'rgba(255, 255, 255, .20)',
    glassSheen: 'rgba(255, 255, 255, .26)',
    glassRim: 'rgba(255, 255, 255, .48)',
    auroraA: 'rgba(58, 92, 190, .70)',
    auroraB: 'rgba(112, 76, 190, .66)',
    auroraC: 'rgba(28, 132, 140, .58)',
    auroraD: 'rgba(150, 62, 116, .52)',
    pageBg: '#070b16',
    textStrong: '#f1f5ff',
    textMuted: 'rgba(203, 213, 225, .72)',
  },
}

/* ── 2. 樱花（暖粉） ──────────────────────────────────────── */
const sakura: ThemePack = {
  id: 'sakura',
  i18nKey: 'sakura',
  preview: ['#ec4899', '#fbcfe8'],
  light: {
    accent: '#db2777',
    accentHover: '#be185d',
    accentSoft: 'rgba(219, 39, 119, .10)',
    logoGradient: 'linear-gradient(150deg, #f472b6, #ec4899 55%, #db2777)',
    glassTop: 'rgba(255, 255, 255, .38)',
    glassBottom: 'rgba(255, 245, 250, .18)',
    glassBorder: 'rgba(255, 255, 255, .52)',
    glassSheen: 'rgba(255, 255, 255, .66)',
    glassRim: 'rgba(255, 255, 255, .96)',
    auroraA: 'rgba(255, 182, 213, .62)',
    auroraB: 'rgba(255, 214, 165, .58)',
    auroraC: 'rgba(214, 168, 255, .50)',
    auroraD: 'rgba(255, 148, 176, .52)',
    pageBg: '#f7edf2',
    textStrong: '#2a121b',
    textMuted: 'rgba(120, 80, 98, .78)',
  },
  dark: {
    accent: '#f472b6',
    accentHover: '#f9a8d4',
    accentSoft: 'rgba(244, 114, 182, .16)',
    logoGradient: 'linear-gradient(150deg, #f472b6, #ec4899 55%, #db2777)',
    glassTop: 'rgba(255, 255, 255, .13)',
    glassBottom: 'rgba(255, 240, 248, .05)',
    glassBorder: 'rgba(255, 255, 255, .20)',
    glassSheen: 'rgba(255, 255, 255, .28)',
    glassRim: 'rgba(255, 255, 255, .50)',
    auroraA: 'rgba(190, 60, 130, .68)',
    auroraB: 'rgba(150, 70, 190, .60)',
    auroraC: 'rgba(200, 110, 150, .56)',
    auroraD: 'rgba(120, 50, 100, .54)',
    pageBg: '#160810',
    textStrong: '#fff1f7',
    textMuted: 'rgba(235, 200, 215, .72)',
  },
}

/* ── 3. 森野（青绿） ──────────────────────────────────────── */
const forest: ThemePack = {
  id: 'forest',
  i18nKey: 'forest',
  preview: ['#10b981', '#a7f3d0'],
  light: {
    accent: '#059669',
    accentHover: '#047857',
    accentSoft: 'rgba(5, 150, 105, .10)',
    logoGradient: 'linear-gradient(150deg, #34d399, #10b981 55%, #059669)',
    glassTop: 'rgba(255, 255, 255, .36)',
    glassBottom: 'rgba(240, 255, 250, .16)',
    glassBorder: 'rgba(255, 255, 255, .50)',
    glassSheen: 'rgba(255, 255, 255, .64)',
    glassRim: 'rgba(255, 255, 255, .95)',
    auroraA: 'rgba(120, 226, 180, .60)',
    auroraB: 'rgba(150, 220, 255, .54)',
    auroraC: 'rgba(200, 240, 150, .50)',
    auroraD: 'rgba(110, 210, 200, .52)',
    pageBg: '#e9f2ee',
    textStrong: '#0a1f18',
    textMuted: 'rgba(60, 100, 85, .78)',
  },
  dark: {
    accent: '#34d399',
    accentHover: '#6ee7b7',
    accentSoft: 'rgba(52, 211, 153, .16)',
    logoGradient: 'linear-gradient(150deg, #34d399, #10b981 55%, #059669)',
    glassTop: 'rgba(255, 255, 255, .11)',
    glassBottom: 'rgba(235, 255, 248, .04)',
    glassBorder: 'rgba(255, 255, 255, .18)',
    glassSheen: 'rgba(255, 255, 255, .24)',
    glassRim: 'rgba(255, 255, 255, .46)',
    auroraA: 'rgba(20, 120, 100, .68)',
    auroraB: 'rgba(30, 130, 150, .60)',
    auroraC: 'rgba(80, 140, 60, .56)',
    auroraD: 'rgba(15, 90, 90, .54)',
    pageBg: '#06120e',
    textStrong: '#eafff5',
    textMuted: 'rgba(180, 220, 205, .72)',
  },
}

/* ── 4. 暮光（紫橙撞色） ──────────────────────────────────── */
const dusk: ThemePack = {
  id: 'dusk',
  i18nKey: 'dusk',
  preview: ['#f97316', '#c084fc'],
  light: {
    accent: '#ea580c',
    accentHover: '#c2410c',
    accentSoft: 'rgba(234, 88, 12, .10)',
    logoGradient: 'linear-gradient(150deg, #fb923c, #f97316 45%, #a855f7)',
    glassTop: 'rgba(255, 255, 255, .34)',
    glassBottom: 'rgba(255, 248, 240, .16)',
    glassBorder: 'rgba(255, 255, 255, .50)',
    glassSheen: 'rgba(255, 255, 255, .64)',
    glassRim: 'rgba(255, 255, 255, .95)',
    auroraA: 'rgba(255, 170, 110, .60)',
    auroraB: 'rgba(200, 150, 255, .58)',
    auroraC: 'rgba(255, 120, 150, .50)',
    auroraD: 'rgba(150, 140, 255, .52)',
    pageBg: '#f4eee9',
    textStrong: '#241408',
    textMuted: 'rgba(110, 85, 65, .78)',
  },
  dark: {
    accent: '#fb923c',
    accentHover: '#fdba74',
    accentSoft: 'rgba(251, 146, 60, .16)',
    logoGradient: 'linear-gradient(150deg, #fb923c, #f97316 45%, #a855f7)',
    glassTop: 'rgba(255, 255, 255, .12)',
    glassBottom: 'rgba(255, 245, 235, .045)',
    glassBorder: 'rgba(255, 255, 255, .19)',
    glassSheen: 'rgba(255, 255, 255, .26)',
    glassRim: 'rgba(255, 255, 255, .48)',
    auroraA: 'rgba(180, 90, 30, .68)',
    auroraB: 'rgba(120, 60, 170, .64)',
    auroraC: 'rgba(160, 50, 80, .56)',
    auroraD: 'rgba(70, 60, 150, .54)',
    pageBg: '#120a06',
    textStrong: '#fff4ea',
    textMuted: 'rgba(225, 200, 180, .72)',
  },
}

/* ── 5. 冰川（冷蓝白，接近 macOS 原生） ───────────────────── */
const glacier: ThemePack = {
  id: 'glacier',
  i18nKey: 'glacier',
  preview: ['#0ea5e9', '#bae6fd'],
  light: {
    accent: '#0284c7',
    accentHover: '#0369a1',
    accentSoft: 'rgba(2, 132, 199, .10)',
    logoGradient: 'linear-gradient(150deg, #38bdf8, #0ea5e9 55%, #0284c7)',
    glassTop: 'rgba(255, 255, 255, .42)',
    glassBottom: 'rgba(245, 251, 255, .20)',
    glassBorder: 'rgba(255, 255, 255, .58)',
    glassSheen: 'rgba(255, 255, 255, .70)',
    glassRim: 'rgba(255, 255, 255, .97)',
    auroraA: 'rgba(150, 210, 255, .58)',
    auroraB: 'rgba(190, 225, 255, .54)',
    auroraC: 'rgba(130, 230, 240, .48)',
    auroraD: 'rgba(200, 215, 255, .52)',
    pageBg: '#eaf1f8',
    textStrong: '#0a1a2b',
    textMuted: 'rgba(65, 90, 115, .78)',
  },
  dark: {
    accent: '#38bdf8',
    accentHover: '#7dd3fc',
    accentSoft: 'rgba(56, 189, 248, .16)',
    logoGradient: 'linear-gradient(150deg, #38bdf8, #0ea5e9 55%, #0284c7)',
    glassTop: 'rgba(255, 255, 255, .13)',
    glassBottom: 'rgba(240, 250, 255, .05)',
    glassBorder: 'rgba(255, 255, 255, .21)',
    glassSheen: 'rgba(255, 255, 255, .28)',
    glassRim: 'rgba(255, 255, 255, .50)',
    auroraA: 'rgba(30, 100, 180, .68)',
    auroraB: 'rgba(50, 120, 200, .60)',
    auroraC: 'rgba(20, 130, 160, .56)',
    auroraD: 'rgba(40, 80, 140, .54)',
    pageBg: '#050b14',
    textStrong: '#eef7ff',
    textMuted: 'rgba(185, 210, 235, .72)',
  },
}

/* ── 6. 石墨（中性灰，最克制） ────────────────────────────── */
const graphite: ThemePack = {
  id: 'graphite',
  i18nKey: 'graphite',
  preview: ['#52525b', '#d4d4d8'],
  light: {
    accent: '#3f3f46',
    accentHover: '#27272a',
    accentSoft: 'rgba(63, 63, 70, .08)',
    logoGradient: 'linear-gradient(150deg, #71717a, #52525b 55%, #3f3f46)',
    glassTop: 'rgba(255, 255, 255, .40)',
    glassBottom: 'rgba(250, 250, 250, .18)',
    glassBorder: 'rgba(255, 255, 255, .56)',
    glassSheen: 'rgba(255, 255, 255, .68)',
    glassRim: 'rgba(255, 255, 255, .96)',
    auroraA: 'rgba(190, 195, 210, .50)',
    auroraB: 'rgba(210, 200, 215, .46)',
    auroraC: 'rgba(180, 200, 200, .42)',
    auroraD: 'rgba(205, 205, 215, .46)',
    pageBg: '#eceef1',
    textStrong: '#18181b',
    textMuted: 'rgba(82, 82, 91, .78)',
  },
  dark: {
    accent: '#a1a1aa',
    accentHover: '#d4d4d8',
    accentSoft: 'rgba(161, 161, 170, .14)',
    logoGradient: 'linear-gradient(150deg, #a1a1aa, #71717a 55%, #52525b)',
    glassTop: 'rgba(255, 255, 255, .10)',
    glassBottom: 'rgba(255, 255, 255, .04)',
    glassBorder: 'rgba(255, 255, 255, .16)',
    glassSheen: 'rgba(255, 255, 255, .20)',
    glassRim: 'rgba(255, 255, 255, .40)',
    auroraA: 'rgba(90, 95, 110, .60)',
    auroraB: 'rgba(110, 100, 120, .54)',
    auroraC: 'rgba(80, 100, 105, .50)',
    auroraD: 'rgba(70, 70, 85, .52)',
    pageBg: '#0c0c0e',
    textStrong: '#f4f4f5',
    textMuted: 'rgba(190, 190, 195, .70)',
  },
}

/* ── 7. 赛博（霓虹青紫，高对比） ──────────────────────────── */
const cyber: ThemePack = {
  id: 'cyber',
  i18nKey: 'cyber',
  preview: ['#22d3ee', '#d946ef'],
  light: {
    accent: '#0891b2',
    accentHover: '#0e7490',
    accentSoft: 'rgba(8, 145, 178, .10)',
    logoGradient: 'linear-gradient(150deg, #22d3ee, #a855f7 55%, #d946ef)',
    glassTop: 'rgba(255, 255, 255, .36)',
    glassBottom: 'rgba(240, 250, 255, .16)',
    glassBorder: 'rgba(255, 255, 255, .52)',
    glassSheen: 'rgba(255, 255, 255, .66)',
    glassRim: 'rgba(255, 255, 255, .96)',
    auroraA: 'rgba(120, 230, 255, .62)',
    auroraB: 'rgba(216, 130, 255, .60)',
    auroraC: 'rgba(100, 255, 220, .50)',
    auroraD: 'rgba(180, 140, 255, .54)',
    pageBg: '#eaf4f8',
    textStrong: '#08191f',
    textMuted: 'rgba(55, 95, 110, .78)',
  },
  dark: {
    accent: '#22d3ee',
    accentHover: '#67e8f9',
    accentSoft: 'rgba(34, 211, 238, .16)',
    logoGradient: 'linear-gradient(150deg, #22d3ee, #a855f7 55%, #d946ef)',
    glassTop: 'rgba(255, 255, 255, .11)',
    glassBottom: 'rgba(230, 250, 255, .045)',
    glassBorder: 'rgba(255, 255, 255, .19)',
    glassSheen: 'rgba(255, 255, 255, .26)',
    glassRim: 'rgba(255, 255, 255, .48)',
    auroraA: 'rgba(20, 150, 180, .70)',
    auroraB: 'rgba(140, 50, 190, .68)',
    auroraC: 'rgba(15, 170, 150, .58)',
    auroraD: 'rgba(110, 60, 200, .60)',
    pageBg: '#050a12',
    textStrong: '#e8fbff',
    textMuted: 'rgba(170, 215, 230, .72)',
  },
}

/* ── 8. 莫兰迪（低饱和灰调，最耐看） ──────────────────────── */
const morandi: ThemePack = {
  id: 'morandi',
  i18nKey: 'morandi',
  preview: ['#8b7d6b', '#c4b5a5'],
  light: {
    accent: '#8a7563',
    accentHover: '#6f5d4e',
    accentSoft: 'rgba(138, 117, 99, .10)',
    logoGradient: 'linear-gradient(150deg, #c4b5a5, #a8998a 55%, #8a7563)',
    glassTop: 'rgba(255, 255, 255, .40)',
    glassBottom: 'rgba(252, 249, 245, .18)',
    glassBorder: 'rgba(255, 255, 255, .56)',
    glassSheen: 'rgba(255, 255, 255, .68)',
    glassRim: 'rgba(255, 255, 255, .96)',
    auroraA: 'rgba(212, 196, 178, .52)',
    auroraB: 'rgba(196, 200, 190, .48)',
    auroraC: 'rgba(206, 190, 200, .46)',
    auroraD: 'rgba(200, 205, 195, .48)',
    pageBg: '#f2efea',
    textStrong: '#2b2620',
    textMuted: 'rgba(110, 100, 88, .78)',
  },
  dark: {
    accent: '#c4b5a5',
    accentHover: '#ddd0c2',
    accentSoft: 'rgba(196, 181, 165, .14)',
    logoGradient: 'linear-gradient(150deg, #c4b5a5, #a8998a 55%, #8a7563)',
    glassTop: 'rgba(255, 255, 255, .10)',
    glassBottom: 'rgba(255, 250, 245, .04)',
    glassBorder: 'rgba(255, 255, 255, .15)',
    glassSheen: 'rgba(255, 255, 255, .20)',
    glassRim: 'rgba(255, 255, 255, .40)',
    auroraA: 'rgba(110, 95, 80, .58)',
    auroraB: 'rgba(90, 95, 85, .54)',
    auroraC: 'rgba(100, 88, 95, .50)',
    auroraD: 'rgba(95, 98, 90, .52)',
    pageBg: '#100e0c',
    textStrong: '#f5f1ea',
    textMuted: 'rgba(200, 190, 178, .70)',
  },
}

/* ── 9. 深海（沉静深蓝） ──────────────────────────────────── */
const abyss: ThemePack = {
  id: 'abyss',
  i18nKey: 'abyss',
  preview: ['#1e40af', '#93c5fd'],
  light: {
    accent: '#1d4ed8',
    accentHover: '#1e3a8a',
    accentSoft: 'rgba(29, 78, 216, .10)',
    logoGradient: 'linear-gradient(150deg, #60a5fa, #3b82f6 50%, #1e40af)',
    glassTop: 'rgba(255, 255, 255, .38)',
    glassBottom: 'rgba(243, 248, 255, .18)',
    glassBorder: 'rgba(255, 255, 255, .54)',
    glassSheen: 'rgba(255, 255, 255, .66)',
    glassRim: 'rgba(255, 255, 255, .96)',
    auroraA: 'rgba(130, 175, 255, .60)',
    auroraB: 'rgba(150, 200, 240, .54)',
    auroraC: 'rgba(110, 200, 230, .50)',
    auroraD: 'rgba(160, 175, 235, .52)',
    pageBg: '#eaeff8',
    textStrong: '#0a1533',
    textMuted: 'rgba(58, 78, 120, .78)',
  },
  dark: {
    accent: '#60a5fa',
    accentHover: '#93c5fd',
    accentSoft: 'rgba(96, 165, 250, .16)',
    logoGradient: 'linear-gradient(150deg, #60a5fa, #3b82f6 50%, #1e40af)',
    glassTop: 'rgba(255, 255, 255, .12)',
    glassBottom: 'rgba(235, 244, 255, .045)',
    glassBorder: 'rgba(255, 255, 255, .20)',
    glassSheen: 'rgba(255, 255, 255, .26)',
    glassRim: 'rgba(255, 255, 255, .48)',
    auroraA: 'rgba(30, 70, 170, .70)',
    auroraB: 'rgba(40, 100, 180, .62)',
    auroraC: 'rgba(25, 120, 160, .56)',
    auroraD: 'rgba(45, 60, 140, .58)',
    pageBg: '#040814',
    textStrong: '#eaf2ff',
    textMuted: 'rgba(175, 200, 235, .72)',
  },
}

export const themePacks: ThemePack[] = [indigo, sakura, forest, dusk, glacier, graphite, cyber, morandi, abyss]

export const DEFAULT_THEME_PACK = 'indigo'

export function getThemePack(id: string | undefined): ThemePack {
  return themePacks.find(p => p.id === id) || themePacks[0]
}

/**
 * 把主题色板转成 CSS 变量对象，注入 documentElement。
 * 变量名统一以 --nv- 开头，避免和 Naive / Tailwind 冲突。
 */
export function paletteToCssVars(p: ThemePalette): Record<string, string> {
  return {
    '--nv-accent': p.accent,
    '--nv-accent-hover': p.accentHover,
    '--nv-accent-soft': p.accentSoft,
    '--nv-logo-gradient': p.logoGradient,
    '--nv-glass-top': p.glassTop,
    '--nv-glass-bottom': p.glassBottom,
    '--nv-glass-border': p.glassBorder,
    '--nv-glass-sheen': p.glassSheen,
    '--nv-glass-rim': p.glassRim,
    '--nv-aurora-a': p.auroraA,
    '--nv-aurora-b': p.auroraB,
    '--nv-aurora-c': p.auroraC,
    '--nv-aurora-d': p.auroraD,
    '--nv-page-bg': p.pageBg,
    '--nv-text-strong': p.textStrong,
    '--nv-text-muted': p.textMuted,
  }
}
