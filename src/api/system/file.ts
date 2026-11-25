import { post } from '@/utils/request'

export function getList<T>(params?: { type?: number }) {
  return post<T>({
    url: '/file/getList',
    data: params || {},
  })
}

export function deletes<T>(ids: number[]) {
  return post<T>({
    url: '/file/deletes',
    data: { ids },
  })
}
