package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type FleetBattleIncentiveRepository struct{}

func NewFleetBattleIncentiveRepository() *FleetBattleIncentiveRepository {
	return &FleetBattleIncentiveRepository{}
}

// ListAll 返回全部激励配置（最多 3 条：cta/strat_op/other）
func (r *FleetBattleIncentiveRepository) ListAll() ([]model.FleetBattleIncentive, error) {
	var list []model.FleetBattleIncentive
	err := global.DB.Order("fleet_type").Find(&list).Error
	return list, err
}

// GetByFleetType 按舰队类型获取配置，不存在时返回默认值（不写库）
func (r *FleetBattleIncentiveRepository) GetByFleetType(fleetType string) (*model.FleetBattleIncentive, error) {
	var cfg model.FleetBattleIncentive
	err := global.DB.Where("fleet_type = ?", fleetType).First(&cfg).Error
	if err != nil {
		// 记录不存在时返回默认结构
		return &model.FleetBattleIncentive{FleetType: fleetType, Enabled: false}, nil
	}
	return &cfg, nil
}

// Upsert 创建或更新指定类型的激励配置
func (r *FleetBattleIncentiveRepository) Upsert(cfg *model.FleetBattleIncentive) error {
	return global.DB.
		Where(model.FleetBattleIncentive{FleetType: cfg.FleetType}).
		Assign(*cfg).
		FirstOrCreate(cfg).Error
}

// Save 保存（已有 ID 则更新）
func (r *FleetBattleIncentiveRepository) Save(cfg *model.FleetBattleIncentive) error {
	return global.DB.Save(cfg).Error
}
