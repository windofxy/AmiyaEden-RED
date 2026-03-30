package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"amiya-eden/global"
	"amiya-eden/internal/model"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  战报（Battle Report）生成
// ─────────────────────────────────────────────

const warbeaconCreateURL = "https://warbeacon.net/api/br/create"
const warbeaconReportURL = "https://warbeacon.net/api/br/report/"

// ── warbeacon 请求/响应 ──────────────────────

type brLocation struct {
	ID        int64  `json:"id"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name      string `json:"name"`
}

type brCreateReq struct {
	Locations []brLocation     `json:"locations"`
	Teams     []map[string]int `json:"teams"`
}

type brTeamMeta struct {
	TeamID           int     `json:"teamId"`
	TotalLosses      int     `json:"totalLosses"`
	TotalLossValue   float64 `json:"totalLossValue"`
	ParticipantCount int     `json:"participantCount"`
}

type brCreateResp struct {
	Success bool `json:"success"`
	Data    struct {
		UUID          string       `json:"uuid"`
		TeamsMetadata []brTeamMeta `json:"teamsMetadata"`
	} `json:"data"`
}

// ── ESI helpers ──────────────────────────────

type esiAffil struct {
	CharacterID   int64 `json:"character_id"`
	CorporationID int64 `json:"corporation_id"`
	AllianceID    int64 `json:"alliance_id"`
}

func brPostJSON(client *http.Client, url string, payload any, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}

func brGetJSON(client *http.Client, url string, dest any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}

// ─────────────────────────────────────────────

// FleetBRResult 战报生成结果（返回给 handler）
type FleetBRResult struct {
	UUID      string `json:"uuid"`
	Team0Loss int    `json:"team0_loss"`
	Team1Loss int    `json:"team1_loss"`
}

// GenerateBattleReport 生成舰队战报并入库
func (s *FleetService) GenerateBattleReport(fleetID string, userID uint, userRole string) (*FleetBRResult, error) {
	// 1. 校验权限
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, fmt.Errorf("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRole) {
		return nil, fmt.Errorf("权限不足")
	}

	global.Logger.Info("[BR] 生成战报开始", zap.String("fleet_id", fleetID), zap.Uint("user_id", userID))

	// 2. 获取 PAP 成员
	papLogs, err := s.repo.ListPapLogsByFleet(fleetID)
	if err != nil || len(papLogs) == 0 {
		return nil, fmt.Errorf("该舰队暂无 PAP 记录，无法生成战报")
	}

	charIDSet := make(map[int64]struct{}, len(papLogs))
	for _, l := range papLogs {
		charIDSet[l.CharacterID] = struct{}{}
	}
	charIDs := make([]int64, 0, len(charIDSet))
	for id := range charIDSet {
		charIDs = append(charIDs, id)
	}

	// 3. ESI 查询角色势力（Team 1 — 己方）
	affiliations := s.fetchAffilBatch(charIDs)

	team1 := make(map[string]int)
	ourAllianceIDs := make(map[int64]bool)
	ourCorpIDs := make(map[int64]bool)
	for _, a := range affiliations {
		if a.AllianceID != 0 {
			team1[fmt.Sprintf("alliance_%d", a.AllianceID)]++
			ourAllianceIDs[a.AllianceID] = true
		} else {
			team1[fmt.Sprintf("corporation_%d", a.CorporationID)]++
			ourCorpIDs[a.CorporationID] = true
		}
	}

	// 4. 从数据库获取舰队时间范围内有参与的 KM（去重）
	type kmRow struct {
		KillmailID    int64     `gorm:"column:killmail_id"`
		KillmailTime  time.Time `gorm:"column:killmail_time"`
		SolarSystemID int64     `gorm:"column:solar_system_id"`
	}
	var kmRows []kmRow
	global.DB.Table("eve_character_killmail ckm").
		Select("DISTINCT kl.kill_mail_id AS killmail_id, kl.kill_mail_time AS killmail_time, kl.solar_system_id").
		Joins("JOIN eve_killmail_list kl ON kl.kill_mail_id = ckm.killmail_id").
		Where("ckm.character_id IN ? AND kl.kill_mail_time BETWEEN ? AND ?",
			charIDs, fleet.StartAt, fleet.EndAt).
		Scan(&kmRows)

	if len(kmRows) == 0 {
		return nil, fmt.Errorf("舰队时间段内无 KM 数据")
	}

	kmIDs := make([]int64, 0, len(kmRows))
	for _, r := range kmRows {
		kmIDs = append(kmIDs, r.KillmailID)
	}

	// 5. 查攻击者 → 构建 Team 2（敌方）
	type attRow struct {
		CharacterID   int64 `gorm:"column:character_id"`
		CorporationID int64 `gorm:"column:corporation_id"`
		AllianceID    int64 `gorm:"column:alliance_id"`
	}
	var attackers []attRow
	global.DB.Table("eve_killmail_attacker").
		Select("character_id, corporation_id, alliance_id").
		Where("kill_mail_id IN ? AND character_id > 0", kmIDs).
		Scan(&attackers)

	team2 := make(map[string]int)
	seenAtk := make(map[int64]bool)
	for _, a := range attackers {
		if seenAtk[a.CharacterID] {
			continue
		}
		seenAtk[a.CharacterID] = true
		// 排除己方
		if a.AllianceID != 0 && ourAllianceIDs[a.AllianceID] {
			continue
		}
		if a.AllianceID == 0 && a.CorporationID != 0 && ourCorpIDs[a.CorporationID] {
			continue
		}
		if a.AllianceID != 0 {
			team2[fmt.Sprintf("alliance_%d", a.AllianceID)]++
		} else if a.CorporationID != 0 {
			team2[fmt.Sprintf("corporation_%d", a.CorporationID)]++
		}
	}

	// 6. 构建 Locations（同星系内间隔 > 30min 拆段）
	sysKMs := make(map[int64][]time.Time)
	for _, r := range kmRows {
		sysKMs[r.SolarSystemID] = append(sysKMs[r.SolarSystemID], r.KillmailTime)
	}

	sysNames := s.fetchSystemNames(sysKMs)

	type cluster struct {
		SysID     int64
		StartTime time.Time
		EndTime   time.Time
	}
	var clusters []cluster
	for sysID, times := range sysKMs {
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
		seg := cluster{SysID: sysID, StartTime: times[0], EndTime: times[0]}
		for i := 1; i < len(times); i++ {
			if times[i].Sub(seg.EndTime) > 30*time.Minute {
				clusters = append(clusters, seg)
				seg = cluster{SysID: sysID, StartTime: times[i], EndTime: times[i]}
			} else {
				seg.EndTime = times[i]
			}
		}
		clusters = append(clusters, seg)
	}

	locations := make([]brLocation, 0, len(clusters))
	for _, cl := range clusters {
		locations = append(locations, brLocation{
			ID:        cl.SysID,
			StartTime: cl.StartTime.UTC().Format("2006-01-02T15:04:05.000Z"),
			EndTime:   cl.EndTime.UTC().Format("2006-01-02T15:04:05.000Z"),
			Name:      sysNames[cl.SysID],
		})
	}

	// 7. 调用 warbeacon API
	reqBody := brCreateReq{
		Locations: locations,
		Teams:     []map[string]int{team1, team2},
	}
	var brResp brCreateResp
	if err := brPostJSON(s.http, warbeaconCreateURL, reqBody, &brResp); err != nil {
		return nil, fmt.Errorf("warbeacon 请求失败: %w", err)
	}
	if !brResp.Success || brResp.Data.UUID == "" {
		return nil, fmt.Errorf("warbeacon 返回失败")
	}

	uuid := brResp.Data.UUID
	global.Logger.Info("[BR] 战报生成成功", zap.String("fleet_id", fleetID), zap.Uint("user_id", userID), zap.String("br_uuid", uuid))

	// 8. 从 /api/br/report/{uuid} 获取真实战损数据（/create 返回的 teamsMetadata 可能为空）
	var brReport brCreateResp
	if err := brGetJSON(s.http, warbeaconReportURL+uuid, &brReport); err != nil {
		global.Logger.Warn("[BR] 获取战报详情失败，降级使用 create 响应", zap.Error(err))
		brReport = brResp
	}

	var team0Loss int
	var team0Value float64
	var team1Loss int
	var team1Value float64
	for _, meta := range brReport.Data.TeamsMetadata {
		global.Logger.Info("[BR] warbeacon team metadata", zap.Int("team_id", meta.TeamID), zap.Int("losses", meta.TotalLosses), zap.Float64("loss_value", meta.TotalLossValue))
		switch meta.TeamID {
		case 0:
			team0Loss = meta.TotalLosses
			team0Value = meta.TotalLossValue
		case 1:
			team1Loss = meta.TotalLosses
			team1Value = meta.TotalLossValue
		}
	}

	// 9. 写入 Fleet 表
	if err := global.DB.Model(&model.Fleet{}).
		Where("id = ?", fleetID).
		Updates(map[string]any{
			"br_uuid":        uuid,
			"br_team0_loss":  team0Loss,
			"br_team0_value": team0Value,
			"br_team1_loss":  team1Loss,
			"br_team1_value": team1Value,
		}).Error; err != nil {
		global.Logger.Warn("[BR] 战报入库失败", zap.String("uuid", uuid), zap.Error(err))
	}

	// 10. 异步检查并发放激励（幂等，失败不影响战报结果）
	go NewFleetBattleIncentiveService().CheckAndIssueIncentive(fleetID, userID)

	return &FleetBRResult{
		UUID:      uuid,
		Team0Loss: team0Loss,
		Team1Loss: team1Loss,
	}, nil
}

// fetchAffilBatch 批量获取角色势力（每批最多 1000）
func (s *FleetService) fetchAffilBatch(charIDs []int64) []esiAffil {
	const batchSize = 1000
	result := make([]esiAffil, 0, len(charIDs))
	for i := 0; i < len(charIDs); i += batchSize {
		end := i + batchSize
		if end > len(charIDs) {
			end = len(charIDs)
		}
		var batch []esiAffil
		url := esiBaseURL + "/characters/affiliation/?datasource=tranquility"
		if err := brPostJSON(s.http, url, charIDs[i:end], &batch); err != nil {
			global.Logger.Warn("[BR] ESI affiliation 失败", zap.Error(err))
			continue
		}
		result = append(result, batch...)
	}
	return result
}

// fetchSystemNames 获取星系名称（ESI，逐个查询，通常数量少）
func (s *FleetService) fetchSystemNames(sysKMs map[int64][]time.Time) map[int64]string {
	names := make(map[int64]string, len(sysKMs))
	for sysID := range sysKMs {
		var sys struct {
			Name string `json:"name"`
		}
		url := fmt.Sprintf("%s/universe/systems/%d/?datasource=tranquility", esiBaseURL, sysID)
		if err := brGetJSON(s.http, url, &sys); err == nil && sys.Name != "" {
			names[sysID] = sys.Name
		} else {
			names[sysID] = fmt.Sprintf("%d", sysID)
		}
	}
	return names
}
