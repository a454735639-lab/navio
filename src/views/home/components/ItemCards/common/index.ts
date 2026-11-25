import { usePanelState } from '@/store'
import { PanelStateNetworkModeEnum } from '@/enums'

const panelState = usePanelState()

/** 探测结果缓存：key = 内网地址的 origin */
const probeCache = new Map<string, { ok: boolean; at: number }>()
const CACHE_TTL = 60_000

/** 超时下限/上限，避免用户把值设得太小导致误判，或太大导致点击卡顿 */
const TIMEOUT_MIN = 200
const TIMEOUT_MAX = 3000

let mixedContentWarned = false

function clampTimeout(v: number | undefined): number {
  const n = Number(v ?? 600)
  if (!Number.isFinite(n) || n <= 0)
    return 600
  return Math.min(TIMEOUT_MAX, Math.max(TIMEOUT_MIN, n))
}

/** 是否是内网/本机地址（用于入口快速判定，0ms 得出结果） */
export function isPrivateHost(hostname: string): boolean {
  const h = (hostname || '').toLowerCase().replace(/^\[|\]$/g, '')
  if (!h)
    return false
  if (h === 'localhost' || h.endsWith('.local') || h.endsWith('.lan') || h.endsWith('.home') || h.endsWith('.internal'))
    return true
  if (/^127\./.test(h) || /^10\./.test(h) || /^192\.168\./.test(h) || /^169\.254\./.test(h))
    return true
  // 172.16.0.0 ~ 172.31.255.255
  if (/^172\.(1[6-9]|2\d|3[01])\./.test(h))
    return true
  // IPv6 唯一本地地址 fc00::/7 与链路本地 fe80::/10
  if (/^(fc|fd|fe80)/.test(h))
    return true
  return false
}

function parseOrigin(url: string): string | null {
  try {
    return new URL(url, window.location.href).origin
  }
  catch {
    return null
  }
}

/**
 * 探测一个 origin 是否可达。
 * 用 no-cors GET：只要拿到响应（哪怕 404/403）就视为可达；网络错误、DNS 失败、超时才失败。
 * 不用 HEAD：部分网关/反代对 HEAD 直接断开，会误判成不可达。
 */
async function probeOnce(url: string, timeout: number): Promise<number> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  const started = performance.now()
  try {
    await fetch(url, {
      method: 'GET',
      mode: 'no-cors',
      cache: 'no-store',
      signal: controller.signal,
    })
    return performance.now() - started
  }
  finally {
    clearTimeout(timer)
  }
}

/**
 * 探测 origin 是否可达：先探根路径（多数服务响应更快），失败再探 favicon。
 * 两次共用一个超时预算，避免总耗时翻倍。
 */
export async function probeOrigin(origin: string, timeout: number): Promise<number> {
  const started = performance.now()
  try {
    return await probeOnce(`${origin}/?_=${Date.now()}`, timeout)
  }
  catch (e) {
    const left = timeout - (performance.now() - started)
    if (left < 80)
      throw e
    return await probeOnce(`${origin}/favicon.ico?_=${Date.now()}`, left)
  }
}

/** 清空缓存（网络切换、页面回到前台时调用） */
export function clearNetworkCache() {
  probeCache.clear()
}

/**
 * 判断内网地址当前是否可达。
 * - https 页面探测 http 内网会被浏览器按混合内容直接拦截，这里提前短路，避免无谓等待
 * - 结果缓存 60s，避免每次点击卡片都探测
 */
export async function detectLanReachable(lanUrl: string, timeout?: number, force = false): Promise<boolean> {
  const origin = parseOrigin(lanUrl)
  if (!origin)
    return false

  if (window.location.protocol === 'https:' && origin.startsWith('http://')) {
    if (!mixedContentWarned) {
      mixedContentWarned = true
      console.warn('[network-mode] https 页面无法探测 http 内网地址，已跳过探测；如需内网直连，请给内网地址也配置 https。')
    }
    return false
  }

  const cached = probeCache.get(origin)
  if (!force && cached && Date.now() - cached.at < CACHE_TTL)
    return cached.ok

  try {
    await probeOrigin(origin, clampTimeout(timeout))
    probeCache.set(origin, { ok: true, at: Date.now() })
    return true
  }
  catch {
    probeCache.set(origin, { ok: false, at: Date.now() })
    return false
  }
}

function syncPanelMode(lan: boolean) {
  const mode = lan ? PanelStateNetworkModeEnum.lan : PanelStateNetworkModeEnum.wan
  // 直接赋值：自动探测结果不必写进本地持久化，避免频繁 IO
  if (panelState.networkMode !== mode)
    panelState.networkMode = mode
}

/** 兼容旧调用的 ping 实现（保留导出，内部逻辑走 probeOrigin） */
export function pingUrl(url: string, attempts: number, timeout: number): Promise<number> {
  const origin = parseOrigin(url)
  if (!origin)
    return Promise.reject(new Error('invalid url'))
  const once = (i: number, latencies: number[]): Promise<number> => {
    if (i >= attempts) {
      if (latencies.length === 0)
        return Promise.reject(new Error('All ping attempts failed'))
      return Promise.resolve(latencies.reduce((a, b) => a + b, 0) / latencies.length)
    }
    return probeOrigin(origin, clampTimeout(timeout))
      .then(ms => once(i + 1, [...latencies, ms]))
      .catch(() => once(i + 1, latencies))
  }
  return once(0, [])
}

/**
 * 获取项目跳转地址。
 * 自动模式下的判定顺序：
 *   1. 入口快速判定：当前页面本身就是用内网地址打开的 → 直接内网（0ms，最准）
 *   2. 探测缓存：60s 内已知内网可达 → 内网
 *   3. 主动探测：no-cors GET 内网 origin，超时失败则回落外网
 */
export async function getItemUrl(itemInfo: Panel.ItemInfo, timeout?: number) {
  if (!itemInfo)
    return ''
  if (!itemInfo.lanUrl)
    return itemInfo.url

  if (!panelState.panelConfig.netModeChangeAuto)
    return (panelState.networkMode === PanelStateNetworkModeEnum.lan ? itemInfo.lanUrl : itemInfo.url) as string

  if (isPrivateHost(window.location.hostname)) {
    syncPanelMode(true)
    return itemInfo.lanUrl
  }

  // 缓存命中就直接用，点击零等待
  const origin = parseOrigin(itemInfo.lanUrl)
  if (origin) {
    const cached = probeCache.get(origin)
    if (cached && Date.now() - cached.at < CACHE_TTL) {
      syncPanelMode(cached.ok)
      return (cached.ok ? itemInfo.lanUrl : itemInfo.url) as string
    }
  }

  // 未命中才探测，且点击路径上的等待上限压到 500ms
  const wait = Math.min(clampTimeout(timeout), 500)
  const ok = await detectLanReachable(itemInfo.lanUrl, wait)
  syncPanelMode(ok)
  return (ok ? itemInfo.lanUrl : itemInfo.url) as string
}

/**
 * 预热：页面加载 / 网络事件后把所有内网 origin 探一遍，点击时直接命中缓存。
 */
export async function warmupNetworkMode(lanUrls: string[], timeout?: number) {
  if (!panelState.panelConfig.netModeChangeAuto)
    return
  const origins = [...new Set(lanUrls.map(u => parseOrigin(u)).filter((v): v is string => !!v))].slice(0, 12)
  if (origins.length === 0)
    return
  const results = await Promise.all(origins.map(o => detectLanReachable(o, timeout, true).catch(() => false)))
  syncPanelMode(results.some(Boolean))
}
