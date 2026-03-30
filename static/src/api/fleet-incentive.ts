import request from '@/utils/http'

/** 获取所有舰队激励配置 */
export function fetchBattleIncentives() {
  return request.get<Api.FleetIncentive.BattleIncentive[]>({
    url: '/api/v1/corp/battle-incentives'
  })
}

/** 更新指定舰队类型的激励配置 */
export function updateBattleIncentive(
  fleetType: string,
  data: Api.FleetIncentive.UpdateBattleIncentiveParams
) {
  return request.put<Api.FleetIncentive.BattleIncentive>({
    url: `/api/v1/corp/battle-incentives/${fleetType}`,
    data
  })
}

/** 手动补发指定舰队的 FC 带队奖励 */
export function issueFCLeadReward(fleetId: string) {
  return request.post<null>({
    url: `/api/v1/corp/fleets/${fleetId}/lead-reward`
  })
}
