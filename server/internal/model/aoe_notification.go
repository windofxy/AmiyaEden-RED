package model

// AOENotification PVE AOE 公告
type AOENotification struct {
	BaseModel
	CreatorUserID uint   `gorm:"not null;index" json:"creator_user_id"`
	System        string `gorm:"size:128;not null" json:"system"`
	Type          string `gorm:"size:128;not null" json:"type"`
	Remark        string `gorm:"size:1024" json:"remark"`
}

func (AOENotification) TableName() string {
	return "aoe_notification"
}

// AOENotificationDTO AOE 公告前端展示数据
type AOENotificationDTO struct {
	ID              uint   `json:"id"`
	CreatorUserID   uint   `json:"creator_user_id"`
	CreatorNickname string `json:"creator_nickname"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	System          string `json:"system"`
	Type            string `json:"type"`
	Remark          string `json:"remark"`
}
