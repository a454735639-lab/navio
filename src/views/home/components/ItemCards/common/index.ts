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

// 最近一次探测结果，供首页角标展示
let lastProbe: { origin: string, ok: boolean, ms: number | null, at: number } | null = null

export function getNetworkStatus() {
  const entry = window.location.hostname
  if (roleProbeCache.lan?.ok && Date.now() - roleProbeCache.lan.at < ROLE_PROBE_TTL)
    return { mode: 'lan', ms: roleProbeCache.lan.ms, at: roleProbeCache.lan.at }
  if (roleProbeCache.ts?.ok && Date.now() - roleProbeCache.ts.at < ROLE_PROBE_TTL)
    return { mode: 'tailnet', ms: roleProbeCache.ts.ms, at: roleProbeCache.ts.at }
  const mode = isTailnetHost(entry)
    ? 'tailnet'
    : (panelState.networkMode === PanelStateNetworkModeEnum.lan ? 'lan' : 'wan')
  return {
    mode,
    ms: lastProbe?.ms ?? null,
    at: lastProbe?.at ?? 0,
  }
}

/** 当前页面是否因 https→http 混合内容限制而无法探测内网 */
export function isProbeBlocked(): boolean {
  return window.location.protocol === 'https:'
}

function clampTimeout(v: number | undefined): number {
  const n = Number(v ?? 600)
  if (!Number.isFinite(n) || n <= 0)
    return 600
  return Math.min(TIMEOUT_MAX, Math.max(TIMEOUT_MIN, n))
}

/** 地址字段支持写多个，逗号/空格分隔，按顺序探测（内网、默认地址均适用） */
export function splitLanUrls(raw: string | undefined): string[] {
  return (raw || '').split(/[,\uFF0C;\uFF1B\u3001\s]+/).map(v => v.trim()).filter(Boolean)
}

/** 是否是 Tailscale 入口（MagicDNS 域名或 100.64.0.0/10 CGNAT） */
export function isTailnetHost(hostname: string): boolean {
  const h = (hostname || '').toLowerCase()
  if (!h)
    return false
  if (h.endsWith('.ts.net'))
    return true
  return /^100\.(6[4-9]|[7-9]\d|1[01]\d|12[0-7])\./.test(h)
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
/** https 探针（lan-probe / ts-probe）：镜像返回 204，用来判断当前网络位置 */
const roleProbeCache: Record<string, { ok: boolean, ms: number, at: number }> = {}
const ROLE_PROBE_TTL = 60_000

async function probeHttps(url: string, timeout = 800): Promise<number> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  const started = performance.now()
  try {
    await fetch(url, { method: 'GET', mode: 'no-cors', cache: 'no-store', signal: controller.signal })
    return Math.round(performance.now() - started)
  }
  catch {
    return -1
  }
  finally {
    clearTimeout(timer)
  }
}

export async function probeRole(role: 'lan' | 'ts', url: string | undefined, force = false): Promise<boolean> {
  if (!url)
    return false
  const cached = roleProbeCache[role]
  if (!force && cached && Date.now() - cached.at < ROLE_PROBE_TTL)
    return cached.ok
  const ms = await probeHttps(url)
  const ok = ms >= 0
  roleProbeCache[role] = { ok, ms, at: Date.now() }
  return ok
}

export function clearRoleProbeCache() {
  delete roleProbeCache.lan
  delete roleProbeCache.ts
}

export async function detectLanReachable(lanUrl: string, timeout?: number, force = false): Promise<boolean> {
  const origin = parseOrigin(lanUrl)
  if (!origin)
    return false

  if (window.location.protocol === 'https:' && origin.startsWith('http://')) {
    if (!mixedContentWarned) {
      mixedContentWarned = true
      console.warn('[network-mode] https 页面无法探测 http 内网地址（浏览器混合内容限制）；将按当前手动选择的模式打开。')
    }
    return false
  }

  const cached = probeCache.get(origin)
  if (!force && cached && Date.now() - cached.at < CACHE_TTL)
    return cached.ok

  try {
    const ms = await probeOrigin(origin, clampTimeout(timeout))
    probeCache.set(origin, { ok: true, at: Date.now() })
    lastProbe = { origin, ok: true, ms: Math.round(ms), at: Date.now() }
    return true
  }
  catch {
    probeCache.set(origin, { ok: false, at: Date.now() })
    lastProbe = { origin, ok: false, ms: null, at: Date.now() }
    return false
  }
}

function syncPanelMode(lan: boolean, mode?: PanelStateNetworkModeEnum) {
  const target = mode ?? (lan ? PanelStateNetworkModeEnum.lan : PanelStateNetworkModeEnum.wan)
  // 直接赋值：自动探测结果不必写进本地持久化，避免频繁 IO
  if (panelState.networkMode !== target)
    panelState.networkMode = target
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

  const lanList = splitLanUrls(itemInfo.lanUrl)
  const wanList = splitLanUrls(itemInfo.url)
  const fallbackWan = wanList[wanList.length - 1] || ''

  if (lanList.length === 0 && wanList.length <= 1)
    return itemInfo.url || ''

  // 手动模式：按选择直接取第一顺位
  if (!panelState.panelConfig.netModeChangeAuto) {
    if (panelState.networkMode === PanelStateNetworkModeEnum.lan)
      return lanList[0] || wanList[0] || ''
    if (panelState.networkMode === PanelStateNetworkModeEnum.tailnet) {
      // Tailscale：优先 100.x / *.ts.net 的候选，找不到就用默认地址第一顺位
      const all = [...wanList, ...lanList]
      const tailnetUrl = all.find(u => /^https?:\/\/(100\.|.*\.ts\.net)/i.test(u))
      return tailnetUrl || wanList[0] || lanList[0] || ''
    }
    return wanList[0] || ''
  }

  // Tailscale 入口：优先用「默认地址」里排在前面的地址（如 100.66），不再探测
  if (isTailnetHost(window.location.hostname)) {
    return wanList[0] || lanList[0] || ''
  }

  // https 探针：先判断在家还是在 tailnet，命中即按层返回（https 页面也可用）
  const lanProbe = panelState.panelConfig.netProbeLanUrl
  const tsProbe = panelState.panelConfig.netProbeTsUrl
  if (await probeRole('lan', lanProbe)) {
    syncPanelMode(true)
    return lanList[0] || wanList[0] || ''
  }
  if (await probeRole('ts', tsProbe)) {
    const tsUrl = wanList.find(u => /^https?:\/\/(100\.|.*\.ts\.net)/i.test(u)) || wanList[0]
    if (tsUrl) {
      // Tailscale 命中应设置 tailnet(2)，而非 lan(0)，否则右键菜单/状态显示与探测结果不一致
      syncPanelMode(true, PanelStateNetworkModeEnum.tailnet)
      return tsUrl
    }
  }

  // 入口就是内网地址：直接用内网第一顺位，0ms
  if (isPrivateHost(window.location.hostname)) {
    syncPanelMode(true)
    return lanList[0] || wanList[0] || ''
  }

  const wait = Math.min(clampTimeout(timeout), 500)
  const deadline = performance.now() + 900 // 多级探测总预算
  const mixed = window.location.protocol === 'https:'
  let mixedBlocked = false

  const tryOne = async (candidate: string): Promise<boolean> => {
    const origin = parseOrigin(candidate)
    if (!origin)
      return false

    if (mixed && origin.startsWith('http://')) {
      mixedBlocked = true
      return false
    }

    const cached = probeCache.get(origin)
    if (cached && Date.now() - cached.at < CACHE_TTL)
      return cached.ok

    const left = deadline - performance.now()
    if (left < 120)
      return false

    return detectLanReachable(candidate, Math.min(wait, left))
  }

  // 1) 内网候选（家里/局域网）
  for (const candidate of lanList) {
    if (await tryOne(candidate)) {
      syncPanelMode(true)
      return candidate
    }
  }

  // 2) 默认地址候选（如 tailnet 优先，域名兜底）
  for (const candidate of wanList) {
    if (await tryOne(candidate)) {
      syncPanelMode(true)
      return candidate
    }
  }

  // https 页面无法探测时，尊重手动选择
  if (mixedBlocked && panelState.networkMode === PanelStateNetworkModeEnum.lan && (lanList[0] || wanList[0]))
    return lanList[0] || wanList[0] || ''

  syncPanelMode(false)
  return fallbackWan || wanList[0] || ''
}

/**
 * 预热：页面加载 / 网络事件后把所有候选地址探一遍，点击时直接命中缓存。
 */
export async function warmupNetworkMode(urls: string[], timeout?: number) {
  if (!panelState.panelConfig.netModeChangeAuto)
    return
  const origins = [...new Set(urls.flatMap(u => splitLanUrls(u)).map(u => parseOrigin(u)).filter((v): v is string => !!v))].slice(0, 12)
  if (origins.length === 0)
    return
  const results = await Promise.all(origins.map(o => detectLanReachable(o, timeout, true).catch(() => false)))
  syncPanelMode(results.some(Boolean))
}
