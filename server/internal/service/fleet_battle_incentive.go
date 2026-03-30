package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// FleetBattleIncentiveService 舰队激励业务层
type FleetBattleIncentiveService struct {
	repo      *repository.FleetBattleIncentiveRepository
	fleetRepo *repository.FleetRepository
	walletSvc *SysWalletService
}

func NewFleetBattleIncentiveService() *FleetBattleIncentiveService {
	return &FleetBattleIncentiveService{
		repo:      repository.NewFleetBattleIncentiveRepository(),
		fleetRepo: repository.NewFleetRepository(),
		walletSvc: NewSysWalletService(),
	}
}

// ─── 配置管理 ─────────────────────────────────

// ListAll 获取全部激励配置（前端展示用，确保三个类型都有值）
func (s *FleetBattleIncentiveService) ListAll() ([]model.FleetBattleIncentive, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}

	// 补全未配置的类型（保证前端始终拿到 3 条）
	existing := make(map[string]bool)
	for _, v := range list {
		existing[v.FleetType] = true
	}
	defaults := []string{model.FleetImportanceCTA, model.FleetImportanceStratOp, model.FleetImportanceOther}
	for _, ft := range defaults {
		if !existing[ft] {
			list = append(list, model.FleetBattleIncentive{
				FleetType:  ft,
				Enabled:    false,
				Multiplier: 1.5,
			})
		}
	}
	return list, nil
}

// UpdateIncentiveRequest 更新激励配置的请求体
type UpdateIncentiveRequest struct {
	Enabled       bool    `json:"enabled"`
	Multiplier    float64 `json:"multiplier"`
	MemberReward  float64 `json:"member_reward"`
	FCReward      float64 `json:"fc_reward"`
	FCLeadEnabled bool    `json:"fc_lead_enabled"`
	FCLeadReward  float64 `json:"fc_lead_reward"`
}

// Update 更新或创建指定类型的激励配置
func (s *FleetBattleIncentiveService) Update(fleetType string, req UpdateIncentiveRequest) error {
	validTypes := map[string]bool{
		model.FleetImportanceCTA:     true,
		model.FleetImportanceStratOp: true,
		model.FleetImportanceOther:   true,
	}
	if !validTypes[fleetType] {
		return errors.New("无效的舰队类型")
	}
	if req.Multiplier < 0 {
		return errors.New("倍数不能为负数")
	}
	if req.MemberReward < 0 || req.FCReward < 0 {
		return errors.New("奖励金额不能为负数")
	}

	// 先取已有记录（或默认），再更新字段
	cfg, err := s.repo.GetByFleetType(fleetType)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}
	cfg.FleetType = fleetType
	cfg.Enabled = req.Enabled
	cfg.Multiplier = req.Multiplier
	cfg.MemberReward = req.MemberReward
	cfg.FCReward = req.FCReward
	cfg.FCLeadEnabled = req.FCLeadEnabled
	cfg.FCLeadReward = req.FCLeadReward

	return s.repo.Save(cfg)
}

// ─── 自动发放 ──────────────────────────────────

// CheckAndIssueIncentive 检查舰队战报并自动发放激励（幂等，重复调用无效）
// fleetID: 舰队 ID；fcUserID: FC 的用户 ID
func (s *FleetBattleIncentiveService) CheckAndIssueIncentive(fleetID string, fcUserID uint) {
	fleet, err := s.fleetRepo.GetByID(fleetID)
	if err != nil {
		global.Logger.Warn("[Incentive] 获取舰队失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	// 幂等：已发放过则跳过
	if fleet.IncentiveIssued {
		global.Logger.Info("[Incentive] 已发放过，跳过", zap.String("fleet_id", fleetID))
		return
	}

	// 获取激励配置
	cfg, err := s.repo.GetByFleetType(fleet.Importance)
	if err != nil || !cfg.Enabled {
		return
	}
	if cfg.MemberReward <= 0 && cfg.FCReward <= 0 {
		return
	}

	// 我方损失为 0 时无法计算比值；当 Multiplier=0 表示无条件发放
	if cfg.Multiplier > 0 {
		if fleet.BrTeam0Value <= 0 {
			global.Logger.Info("[Incentive] 我方损失为0，无法计算倍数，跳过", zap.String("fleet_id", fleetID))
			return
		}
		ratio := fleet.BrTeam1Value / fleet.BrTeam0Value
		if ratio < cfg.Multiplier {
			global.Logger.Info("[Incentive] 倍数未达标，不发放",
				zap.String("fleet_id", fleetID),
				zap.Float64("ratio", ratio),
				zap.Float64("required", cfg.Multiplier),
			)
			return
		}
	}

	global.Logger.Info("[Incentive] 条件满足，开始发放激励",
		zap.String("fleet_id", fleetID),
		zap.Float64("member_reward", cfg.MemberReward),
		zap.Float64("fc_reward", cfg.FCReward),
	)

	// 获取 PAP 成员列表
	papLogs, err := s.fleetRepo.ListPapLogsByFleet(fleetID)
	if err != nil || len(papLogs) == 0 {
		global.Logger.Warn("[Incentive] 无 PAP 成员，跳过", zap.String("fleet_id", fleetID))
		return
	}

	// 去重成员 userID
	seen := make(map[uint]bool)
	for _, log := range papLogs {
		uid := log.UserID
		if seen[uid] {
			continue
		}
		seen[uid] = true

		reward := cfg.MemberReward
		reason := fmt.Sprintf("舰队激励奖励 [%s]", fleet.Title)

		// FC 使用 fc_reward
		if uid == fcUserID && cfg.FCReward > 0 {
			reward = cfg.FCReward
			reason = fmt.Sprintf("舰队激励奖励 (FC) [%s]", fleet.Title)
		}
		if reward <= 0 {
			continue
		}

		if err := s.walletSvc.CreditUser(uid, reward, reason, model.WalletRefBrIncentive, fleetID); err != nil {
			global.Logger.Warn("[Incentive] 发放钱包失败",
				zap.Uint("user_id", uid),
				zap.Error(err),
			)
		}
	}

	// 标记已发放（避免重复）
	if err := global.DB.Model(&model.Fleet{}).
		Where("id = ?", fleetID).
		Update("incentive_issued", true).Error; err != nil {
		global.Logger.Warn("[Incentive] 标记 incentive_issued 失败", zap.String("fleet_id", fleetID), zap.Error(err))
	}

	global.Logger.Info("[Incentive] 激励发放完成", zap.String("fleet_id", fleetID), zap.Int("member_count", len(seen)))
}

// ─── FC 带队奖励 ─────────────────────────────────

// IssueFCLeadReward 发放 FC 带队奖励（PAP 发放时自动调用，也可手动触发）
// force=true 时忽略幂等标记，强制重新发放（手动补发场景）
func (s *FleetBattleIncentiveService) IssueFCLeadReward(fleetID string, force bool) error {
	fleet, err := s.fleetRepo.GetByID(fleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}

	// 幂等检查（非强制时）
	if !force && fleet.LeadRewardIssued {
		return errors.New("带队奖励已发放过，如需重新发放请使用手动补发")
	}

	// 获取该舰队类型的激励配置
	cfg, err := s.repo.GetByFleetType(fleet.Importance)
	if err != nil || !cfg.FCLeadEnabled || cfg.FCLeadReward <= 0 {
		return errors.New("该舰队类型未启用带队奖励或奖励金额为 0")
	}

	reason := fmt.Sprintf("FC 带队奖励 [%s]", fleet.Title)
	if err := s.walletSvc.CreditUser(fleet.FCUserID, cfg.FCLeadReward, reason, model.WalletRefFCLeadReward, fleetID); err != nil {
		return fmt.Errorf("发放钱包失败: %w", err)
	}

	// 标记已发放
	if err := global.DB.Model(&model.Fleet{}).
		Where("id = ?", fleetID).
		Update("lead_reward_issued", true).Error; err != nil {
		global.Logger.Warn("[LeadReward] 标记 lead_reward_issued 失败",
			zap.String("fleet_id", fleetID),
			zap.Error(err),
		)
	}

	global.Logger.Info("[LeadReward] FC 带队奖励发放完成",
		zap.String("fleet_id", fleetID),
		zap.Uint("fc_user_id", fleet.FCUserID),
		zap.Float64("amount", cfg.FCLeadReward),
	)
	return nil
}

// tryIssueFCLeadReward 异步尝试发放 FC 带队奖励（忽略错误，用于 PAP 发放后自动触发）
func (s *FleetBattleIncentiveService) tryIssueFCLeadReward(fleetID string) {
	if err := s.IssueFCLeadReward(fleetID, false); err != nil {
		global.Logger.Info("[LeadReward] 自动发放跳过",
			zap.String("fleet_id", fleetID),
			zap.String("reason", err.Error()),
		)
	}
}
