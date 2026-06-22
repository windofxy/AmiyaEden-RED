package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"strings"
)

// AOENotificationService AOE 公告业务逻辑层
type AOENotificationService struct {
	repo *repository.AOENotificationRepository
}

func NewAOENotificationService() *AOENotificationService {
	return &AOENotificationService{
		repo: repository.NewAOENotificationRepository(),
	}
}

// AOENotificationRequest 创建/更新 AOE 公告请求
type AOENotificationRequest struct {
	System string `json:"system"`
	Type   string `json:"type"`
	Remark string `json:"remark"`
}

func (s *AOENotificationService) List(page, pageSize int) ([]model.AOENotificationDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(page, pageSize)
}

func (s *AOENotificationService) Create(userID uint, req *AOENotificationRequest) (*model.AOENotification, error) {
	normalized, err := normalizeAOENotificationRequest(req)
	if err != nil {
		return nil, err
	}

	notification := &model.AOENotification{
		CreatorUserID: userID,
		System:        normalized.System,
		Type:          normalized.Type,
		Remark:        normalized.Remark,
	}
	if err := s.repo.Create(notification); err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *AOENotificationService) Update(id, userID uint, req *AOENotificationRequest) (*model.AOENotification, error) {
	normalized, err := normalizeAOENotificationRequest(req)
	if err != nil {
		return nil, err
	}

	notification, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("AOE 公告不存在")
	}
	if notification.CreatorUserID != userID {
		return nil, errors.New("只能编辑自己创建的 AOE 公告")
	}

	notification.System = normalized.System
	notification.Type = normalized.Type
	notification.Remark = normalized.Remark
	if err := s.repo.Update(notification); err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *AOENotificationService) Delete(id, userID uint) error {
	notification, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("AOE 公告不存在")
	}
	if notification.CreatorUserID != userID {
		return errors.New("只能删除自己创建的 AOE 公告")
	}
	return s.repo.Delete(id)
}

func normalizeAOENotificationRequest(req *AOENotificationRequest) (*AOENotificationRequest, error) {
	if req == nil {
		return nil, errors.New("请求参数错误")
	}

	system := strings.TrimSpace(req.System)
	notificationType := strings.TrimSpace(req.Type)
	remark := strings.TrimSpace(req.Remark)

	if system == "" {
		return nil, errors.New("星系不能为空")
	}
	if notificationType == "" {
		return nil, errors.New("类型不能为空")
	}
	if len([]rune(system)) > 128 {
		return nil, errors.New("星系不能超过128个字符")
	}
	if len([]rune(notificationType)) > 128 {
		return nil, errors.New("类型不能超过128个字符")
	}
	if len([]rune(remark)) > 1024 {
		return nil, errors.New("备注不能超过1024个字符")
	}

	return &AOENotificationRequest{
		System: system,
		Type:   notificationType,
		Remark: remark,
	}, nil
}
