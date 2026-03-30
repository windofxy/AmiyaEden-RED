package model

import "time"

// FleetBattleIncentive 舰队激励配置（按舰队类型，每种类型一条记录）
type FleetBattleIncentive struct {
	ID            uint      `gorm:"primarykey"                         json:"id"`
	FleetType     string    `gorm:"size:32;uniqueIndex;not null"       json:"fleet_type"`      // strat_op / cta / other
	Enabled       bool      `gorm:"default:false"                      json:"enabled"`         // 是否启用战报激励
	Multiplier    float64   `gorm:"default:1.5"                        json:"multiplier"`      // 触发条件：敌方损失/我方损失 >= 此倍数（0 = 无条件发放）
	MemberReward  float64   `gorm:"default:0"                          json:"member_reward"`   // 成员获得的钱包奖励
	FCReward      float64   `gorm:"default:0"                          json:"fc_reward"`       // FC 获得的钱包奖励（战报激励）
	FCLeadEnabled bool      `gorm:"default:false"                      json:"fc_lead_enabled"` // 是否启用 FC 带队奖励（PAP 发放时触发）
	FCLeadReward  float64   `gorm:"default:0"                          json:"fc_lead_reward"`  // FC 带队奖励金额
	CreatedAt     time.Time `gorm:"autoCreateTime"                     json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"                     json:"updated_at"`
}

func (FleetBattleIncentive) TableName() string { return "fleet_battle_incentive" }
