package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// AOENotificationRepository AOE 公告数据访问层
type AOENotificationRepository struct{}

func NewAOENotificationRepository() *AOENotificationRepository {
	return &AOENotificationRepository{}
}

// Create 创建 AOE 公告
func (r *AOENotificationRepository) Create(notification *model.AOENotification) error {
	return global.DB.Create(notification).Error
}

// Update 更新 AOE 公告
func (r *AOENotificationRepository) Update(notification *model.AOENotification) error {
	return global.DB.Save(notification).Error
}

// Delete 删除 AOE 公告
func (r *AOENotificationRepository) Delete(id uint) error {
	return global.DB.Delete(&model.AOENotification{}, id).Error
}

// GetByID 根据 ID 查询 AOE 公告
func (r *AOENotificationRepository) GetByID(id uint) (*model.AOENotification, error) {
	var notification model.AOENotification
	err := global.DB.First(&notification, id).Error
	return &notification, err
}

// List 查询 AOE 公告列表
func (r *AOENotificationRepository) List(page, pageSize int) ([]model.AOENotificationDTO, int64, error) {
	var list []model.AOENotificationDTO
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Table("aoe_notification AS an").
		Select(`an.id, an.creator_user_id, COALESCE(NULLIF(u.nickname, ''), CONCAT('Capsuleer#', an.creator_user_id)) AS creator_nickname, an.created_at, an.updated_at, an.system, an.type, an.remark`).
		Joins("LEFT JOIN `user` AS u ON u.id = an.creator_user_id AND u.deleted_at IS NULL").
		Where("an.deleted_at IS NULL")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("an.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
