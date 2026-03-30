package model

import (
	"github.com/google/uuid"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeAnomalyFuel NotificationType = "anomaly_fuel"    // 异常油耗预警
	NotificationTypeMaintenance NotificationType = "maintenance_due" // 保养到期提醒
	NotificationTypeInviteUsed  NotificationType = "invite_used"     // 邀请码被使用
)

// Notification 通知模型
type Notification struct {
	BaseModel

	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	VehicleID *uuid.UUID `gorm:"type:uuid" json:"vehicle_id,omitempty"` // 可选，邀请码通知无关联车辆

	// 通知内容
	Type    NotificationType `gorm:"size:30;not null" json:"type"`
	Title   string           `gorm:"size:200;not null" json:"title"`
	Message string           `gorm:"type:text;not null" json:"message"`

	// 关联实体（可选）
	ReminderID *uuid.UUID `gorm:"type:uuid" json:"reminder_id,omitempty"` // 关联的提醒
	RecordID   *uuid.UUID `gorm:"type:uuid" json:"record_id,omitempty"`   // 关联的加油记录

	// 状态
	IsRead bool `gorm:"default:false;not null" json:"is_read"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}
