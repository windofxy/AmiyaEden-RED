package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// FleetBattleIncentiveHandler 舰队激励配置 HTTP 处理器
type FleetBattleIncentiveHandler struct {
	svc *service.FleetBattleIncentiveService
}

func NewFleetBattleIncentiveHandler() *FleetBattleIncentiveHandler {
	return &FleetBattleIncentiveHandler{svc: service.NewFleetBattleIncentiveService()}
}

// ListAll GET /corp/battle-incentives
func (h *FleetBattleIncentiveHandler) ListAll(c *gin.Context) {
	list, err := h.svc.ListAll()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, list)
}

// Update PUT /corp/battle-incentives/:fleet_type
func (h *FleetBattleIncentiveHandler) Update(c *gin.Context) {
	fleetType := c.Param("fleet_type")
	var req service.UpdateIncentiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	if err := h.svc.Update(fleetType, req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// IssueFCLeadReward POST /corp/fleets/:id/lead-reward
// 手动补发 FC 带队奖励（force=true，忽略幂等标记）
func (h *FleetBattleIncentiveHandler) IssueFCLeadReward(c *gin.Context) {
	fleetID := c.Param("id")
	if err := h.svc.IssueFCLeadReward(fleetID, true); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
