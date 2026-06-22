import request from '@/utils/http'

// ─── AOE 公告 ───

export function fetchAOENotificationList(params?: Api.PVE.AOENotificationListParams) {
  return request.get<Api.Common.PaginatedResponse<Api.PVE.AOENotification>>({
    url: '/api/v1/pve/aoe-notifications',
    params
  })
}

export function createAOENotification(data: Api.PVE.AOENotificationRequest) {
  return request.post<Api.PVE.AOENotification>({
    url: '/api/v1/pve/aoe-notifications',
    data
  })
}

export function updateAOENotification(id: number, data: Api.PVE.AOENotificationRequest) {
  return request.put<Api.PVE.AOENotification>({
    url: `/api/v1/pve/aoe-notifications/${id}`,
    data
  })
}

export function deleteAOENotification(id: number) {
  return request.del({
    url: `/api/v1/pve/aoe-notifications/${id}`
  })
}
