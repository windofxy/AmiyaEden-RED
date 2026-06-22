package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AOENotificationHandler AOE 公告 HTTP 处理器
type AOENotificationHandler struct {
	svc *service.AOENotificationService
}

func NewAOENotificationHandler() *AOENotificationHandler {
	return &AOENotificationHandler{
		svc: service.NewAOENotificationService(),
	}
}

func (h *AOENotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	list, total, err := h.svc.List(page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, page, size)
}

func (h *AOENotificationHandler) Create(c *gin.Context) {
	var req service.AOENotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	result, err := h.svc.Create(middleware.GetUserID(c), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *AOENotificationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的公告ID")
		return
	}

	var req service.AOENotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	result, err := h.svc.Update(uint(id), middleware.GetUserID(c), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *AOENotificationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的公告ID")
		return
	}

	if err := h.svc.Delete(uint(id), middleware.GetUserID(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
