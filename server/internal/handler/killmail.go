package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// KillmailHandler 击杀邮件处理器
type KillmailHandler struct {
	svc *service.KillmailService
}

func NewKillmailHandler() *KillmailHandler {
	return &KillmailHandler{svc: service.NewKillmailService()}
}

// GetCharacterKillmails POST /info/killmails
// 查询指定角色在指定时间段内的击杀（attacker）和损失（victim）邮件
func (h *KillmailHandler) GetCharacterKillmails(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.KillmailListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetCharacterKillmails(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
