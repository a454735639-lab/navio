import { get, post } from '@/utils/request'

export interface ProbeConfig {
  enabled: boolean
  port: number
  provider: 'path' | 'upload' | 'acme'
  certPath: string
  keyPath: string
  acmeEmail?: string
  acmeDomains?: string[]
  dnsProvider?: string
  staging?: boolean
}

export interface ProbeConfigResp {
  config: ProbeConfig
  hasToken: boolean
  cert: {
    provider?: string
    notAfter?: string
    issuer?: string
    lastError?: string
  }
}

export function getProbeConfig<T>() {
  return get<T>({ url: '/panel/network/probe' })
}

export function saveProbeConfig<T>(data: ProbeConfig & { dnsToken?: string }) {
  return post<T>({ url: '/panel/network/probe', data })
}

export function testProbe<T>() {
  return post<T>({ url: '/panel/network/probe/test' })
}
