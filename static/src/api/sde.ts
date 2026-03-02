import request from '@/utils/http'

/** 批量查询 ID → 名称映射 */
export function fetchNames(data: {
  language?: string
  ids?: Record<string, number[]>
  esi?: number[]
}) {
  return request.post<Record<number, string>>({
    url: '/api/v1/sde/names',
    data
  })
}

/** 模糊搜索物品/成员名称 */
export function fuzzySearch(data: Api.Sde.FuzzySearchRequest) {
  return request.post<Api.Sde.FuzzySearchItem[]>({
    url: '/api/v1/sde/search',
    data
  })
}
