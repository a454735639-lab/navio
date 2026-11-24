import { post } from '@/utils/request'

export function get<T>() {
  return post<T>({
    url: '/about',
  })
}

export interface RunningInfo {
  configPath: string
  webPath: string
  customStylePath: string
  sourcePath: string
  sourceTempPath: string
  binaryFilePath: string
}

export function runningInfo<T>() {
  return post<T>({
    url: '/about/version/runningInfo',
  })
}

export function checkNewVersion<T>(isCheckBeta: boolean) {
  return post<T>({
    url: '/about/version/checkNewVersion',
    data: { isCheckBeta },
  })
}

export function setAutoCheckNewVersion<T>(data: About.AutoCheckNewVersionConfig) {
  return post<T>({
    url: '/about/version/setAutoCheckNewVersion',
    data,
  })
}

export function getAutoCheckNewVersion<T>(isCheckBeta: boolean) {
  return post<T>({
    url: '/about/version/getAutoCheckNewVersion',
  })
}
