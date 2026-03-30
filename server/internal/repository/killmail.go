package repository

import (
	"amiya-eden/global"
	"time"
)

// KillmailRepository 击杀邮件数据访问层
type KillmailRepository struct{}

func NewKillmailRepository() *KillmailRepository { return &KillmailRepository{} }

// KillmailListRow 查询结果行（击杀邮件 + 角色关联信息）
type KillmailListRow struct {
	KillmailID    int64     `gorm:"column:kill_mail_id"`
	KillmailTime  time.Time `gorm:"column:kill_mail_time"`
	SolarSystemID int64     `gorm:"column:solar_system_id"`
	ShipTypeID    int64     `gorm:"column:ship_type_id"`
	VictimCharID  int64     `gorm:"column:victim_char_id"` // eve_killmail_list.character_id（受害者）
	IsVictim      bool      `gorm:"column:is_victim"`      // eve_character_killmail.victim（当前角色是否受害者）
}

// ListByCharacter 按角色 + 时间段查询击杀/损失邮件（分页）
// page <= 0 时返回全部
func (r *KillmailRepository) ListByCharacter(
	characterID int64,
	start, end time.Time,
	page, pageSize int,
) ([]KillmailListRow, int64, error) {
	db := global.DB.Table("eve_character_killmail ckm").
		Select(`kl.kill_mail_id,
			kl.kill_mail_time,
			kl.solar_system_id,
			kl.ship_type_id,
			kl.character_id AS victim_char_id,
			ckm.victim AS is_victim`).
		Joins("JOIN eve_killmail_list kl ON kl.kill_mail_id = ckm.killmail_id").
		Where("ckm.character_id = ? AND kl.kill_mail_time BETWEEN ? AND ?", characterID, start, end)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := db.Order("kl.kill_mail_time DESC")
	if page > 0 && pageSize > 0 {
		q = q.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	var rows []KillmailListRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
