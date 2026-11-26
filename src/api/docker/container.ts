import { useAuthStore } from '@/store'
import { post } from '@/utils/request'

export function getList<T>() {
  return post<T>({
    url: '/docker/container/getList',
  })
}

export function execSwitchAction<T>(param: Docker.Container.ExecSwitchActionReq) {
  // oss 原生端点，与 Pro 端点行为一致（后端 /docker/container/switchState）
  return post<T>({
    url: '/docker/container/switchState',
    data: param,
  })
}

export function restartContainer<T>(param: { id: string }) {
  return post<T>({
    url: '/docker/container/restartContainer',
    data: param,
  })
}

export function getContainerStates<T>(containerIds: string[], ts: number) {
  return post<T>({
    url: '/docker/container/getContainerStates',
    data: { containerIds, ts },
  })
}

export function getContainerStatus<T>(containerId: string) {
  return post<T>({
    url: '/docker/container/getContainerStatus',
    data: { id: containerId },
  })
}

export function getContainersStats<T>(containerIds: string[]) {
  return post<T>({
    url: '/docker/container/getContainersStats',
    data: containerIds,
  })
}
export function getContainerIDByName(name: string) {
  return post<Docker.Container.Info>({
    url: '/docker/container/getContainerIDByName',
    data: { name },
  })
}

export interface ContainerLogParams {
  id: string
  follow?: boolean
  tail?: string
  since?: string
  timestamps?: boolean
}

/**
 * 流式拉取容器日志（后端 SSE）。
 * EventSource 无法携带 token header，故用 fetch + ReadableStream 手动解析 SSE 事件流，
 * 以复用现有 token 登录鉴权。
 *
 * @returns abort 函数，调用后中断日志流
 */
export function streamContainerLogs(
  params: ContainerLogParams,
  onLine: (line: string) => void,
  onError: (msg: string) => void,
  onClose: () => void,
): () => void {
  const authStore = useAuthStore()
  const query = new URLSearchParams()
  query.set('id', params.id)
  query.set('follow', String(params.follow ?? true))
  query.set('tail', params.tail ?? '200')
  query.set('timestamps', String(params.timestamps ?? false))
  if (params.since)
    query.set('since', params.since)

  const controller = new AbortController()
  ;(async () => {
    try {
      const resp = await fetch(`${import.meta.env.VITE_GLOB_API_URL}/docker/container/logs?${query.toString()}`, {
        method: 'GET',
        headers: { token: authStore.token || '' },
        signal: controller.signal,
      })

      const contentType = resp.headers.get('content-type')?.toLowerCase() || ''
      if (contentType.includes('application/json')) {
        let msg = `HTTP ${resp.status}`
        try {
          const j = await resp.json()
          msg = j?.msg || j?.message || msg
        }
        catch {}
        onError(msg)
        return
      }
      if (!resp.ok || !resp.body) {
        let msg = `HTTP ${resp.status}`
        try {
          const j = await resp.json()
          msg = j?.msg || j?.message || msg
        }
        catch {}
        onError(msg)
        return
      }

      const reader = resp.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done)
          break
        buffer += decoder.decode(value, { stream: true })
        // SSE 事件以空行分隔
        const events = buffer.split('\n\n')
        buffer = events.pop() || ''
        for (const ev of events) {
          for (const line of ev.split('\n')) {
            if (line.startsWith('data: '))
              onLine(line.slice(6))
          }
        }
      }
    }
    catch (e: any) {
      if (e?.name !== 'AbortError')
        onError(e?.message || 'stream error')
    }
    finally {
      onClose()
    }
  })()

  return () => controller.abort()
}
