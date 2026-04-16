package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// ServerUpdateHandler handles server self-upgrade endpoints.
type ServerUpdateHandler struct {
	svc *service.ServerUpdateService
}

func NewServerUpdateHandler() *ServerUpdateHandler {
	return &ServerUpdateHandler{svc: service.NewServerUpdateService()}
}

// CheckUpdate godoc
// GET /api/v1/system/server-update/check
// 检查是否有新版本可用
func (h *ServerUpdateHandler) CheckUpdate(c *gin.Context) {
	result, err := h.svc.CheckUpdate()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// PerformUpgrade godoc
// POST /api/v1/system/server-update/upgrade
// 下载最新二进制并替换，随后退出进程（Docker 自动重启）
func (h *ServerUpdateHandler) PerformUpgrade(c *gin.Context) {
	if err := h.svc.PerformUpgrade(); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "升级包已下载完毕，服务正在重启，请稍候..."})
}

// PerformFrontendUpgrade godoc
// POST /api/v1/system/server-update/upgrade-frontend
// 下载最新前端资源包并解压到静态文件目录
func (h *ServerUpdateHandler) PerformFrontendUpgrade(c *gin.Context) {
	if err := h.svc.PerformFrontendUpgrade(); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "前端资源已更新，刷新页面即可生效。"})
}
