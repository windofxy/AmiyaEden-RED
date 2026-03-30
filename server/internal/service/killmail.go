package service

import (
	"fmt"
	"time"

	"amiya-eden/internal/repository"
)

// KillmailService 击杀邮件查询业务层
type KillmailService struct {
	kmRepo   *repository.KillmailRepository
	charRepo *repository.EveCharacterRepository
	sdeRepo  *repository.SdeRepository
}

func NewKillmailService() *KillmailService {
	return &KillmailService{
		kmRepo:   repository.NewKillmailRepository(),
		charRepo: repository.NewEveCharacterRepository(),
		sdeRepo:  repository.NewSdeRepository(),
	}
}

// ─────────────────────────────────────────────
//  请求 & 响应
// ─────────────────────────────────────────────

// KillmailListRequest 查询击杀/损失邮件列表请求
type KillmailListRequest struct {
	CharacterID int64  `json:"character_id" binding:"required"`
	StartDate   string `json:"start_date"`              // 格式: 2006-01-02，缺省近 30 天
	EndDate     string `json:"end_date"`                // 格式: 2006-01-02，缺省今天
	Language    string `json:"language"`                // 默认 zh
	Page        int    `json:"page"    binding:"min=0"` // 0 = 不分页
	PageSize    int    `json:"page_size" binding:"min=0"`
}

// KillmailListItem 列表单条记录
type KillmailListItem struct {
	KillmailID    int64     `json:"killmail_id"`
	KillmailTime  time.Time `json:"killmail_time"`
	ShipTypeID    int64     `json:"ship_type_id"`
	ShipName      string    `json:"ship_name"`
	SolarSystemID int64     `json:"solar_system_id"`
	SystemName    string    `json:"system_name"`
	IsVictim      bool      `json:"is_victim"` // true=损失, false=击杀
}

// KillmailListResponse 分页响应
type KillmailListResponse struct {
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Items    []KillmailListItem `json:"items"`
}

// ─────────────────────────────────────────────
//  业务方法
// ─────────────────────────────────────────────

// GetCharacterKillmails 查询指定角色在指定时间段内的击杀/损失邮件
func (s *KillmailService) GetCharacterKillmails(userID uint, req *KillmailListRequest) (*KillmailListResponse, error) {
	// 1. 校验角色归属
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	owned := false
	for _, c := range chars {
		if c.CharacterID == req.CharacterID {
			owned = true
			break
		}
	}
	if !owned {
		return nil, fmt.Errorf("character not owned by current user")
	}

	// 2. 解析时间范围
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	start := end.AddDate(0, 0, -29)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	if req.StartDate != "" {
		if t, e := time.ParseInLocation("2006-01-02", req.StartDate, time.UTC); e == nil {
			start = t
		}
	}
	if req.EndDate != "" {
		if t, e := time.ParseInLocation("2006-01-02", req.EndDate, time.UTC); e == nil {
			end = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		}
	}

	page, pageSize := req.Page, req.PageSize
	if page == 0 {
		pageSize = 0 // 不分页
	}
	if pageSize == 0 {
		page = 0
	}

	// 3. 查询数据库
	rows, total, err := s.kmRepo.ListByCharacter(req.CharacterID, start, end, page, pageSize)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &KillmailListResponse{Total: 0, Page: page, PageSize: pageSize, Items: []KillmailListItem{}}, nil
	}

	// 4. 收集 typeID（舰船）和 systemID，批量查 SDE 名称
	shipIDs := make([]int, 0, len(rows))
	sysIDs := make([]int, 0, len(rows))
	shipSet := make(map[int64]bool)
	sysSet := make(map[int64]bool)
	for _, r := range rows {
		if !shipSet[r.ShipTypeID] {
			shipSet[r.ShipTypeID] = true
			shipIDs = append(shipIDs, int(r.ShipTypeID))
		}
		if !sysSet[r.SolarSystemID] {
			sysSet[r.SolarSystemID] = true
			sysIDs = append(sysIDs, int(r.SolarSystemID))
		}
	}

	nameMap, _ := s.sdeRepo.GetNames(map[string][]int{
		"type":         shipIDs,
		"solar_system": sysIDs,
	}, lang)

	// 5. 组装响应
	items := make([]KillmailListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, KillmailListItem{
			KillmailID:    r.KillmailID,
			KillmailTime:  r.KillmailTime,
			ShipTypeID:    r.ShipTypeID,
			ShipName:      nameMap[int(r.ShipTypeID)],
			SolarSystemID: r.SolarSystemID,
			SystemName:    nameMap[int(r.SolarSystemID)],
			IsVictim:      r.IsVictim,
		})
	}

	return &KillmailListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}
