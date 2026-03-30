package model

import (
	"time"

	"github.com/google/uuid"
)

// ReminderType 提醒类型
type ReminderType string

const (
	ReminderTypeMaintenance ReminderType = "maintenance" // 保养提醒
)

// ReminderTrigger 触发方式
type ReminderTrigger string

const (
	ReminderTriggerMileage ReminderTrigger = "mileage" // 按里程
	ReminderTriggerTime    ReminderTrigger = "time"    // 按时间
	ReminderTriggerBoth    ReminderTrigger = "both"    // 里程或时间（任一达到即触发）
)

// MaintenanceCategory 保养项目类型
type MaintenanceCategory string

const (
	MaintenanceCategoryOilChange    MaintenanceCategory = "oil_change"    // 换机油
	MaintenanceCategoryTireRotation MaintenanceCategory = "tire_rotation" // 轮胎换位
	MaintenanceCategoryBrakePads    MaintenanceCategory = "brake_pads"    // 刹车片
	MaintenanceCategoryAirFilter    MaintenanceCategory = "air_filter"    // 空气滤清器
	MaintenanceCategoryTransmission MaintenanceCategory = "transmission"  // 变速箱油
	MaintenanceCategoryCoolant      MaintenanceCategory = "coolant"       // 冷却液
	MaintenanceCategorySparkPlugs   MaintenanceCategory = "spark_plugs"   // 火花塞
	MaintenanceCategoryBattery      MaintenanceCategory = "battery"       // 电池/蓄电池
	MaintenanceCategoryTireReplace  MaintenanceCategory = "tire_replace"  // 换轮胎
	MaintenanceCategoryInspection   MaintenanceCategory = "inspection"    // 年检
	MaintenanceCategoryCustom       MaintenanceCategory = "custom"        // 自定义
)

// Reminder 提醒模型
type Reminder struct {
	BaseModel

	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`

	// 提醒基本信息
	Type        ReminderType        `gorm:"size:20;not null" json:"type"`
	Category    MaintenanceCategory `gorm:"size:30;not null" json:"category"`
	Title       string              `gorm:"size:200;not null" json:"title"`        // 用户自定义标题
	Description string              `gorm:"type:text" json:"description,omitempty"` // 备注说明

	// 触发条件
	Trigger         ReminderTrigger `gorm:"size:10;not null;default:both" json:"trigger"`
	MileageInterval float64         `gorm:"type:decimal(10,1)" json:"mileage_interval,omitempty"` // 每隔多少公里（存 km）
	TimeIntervalDays int            `gorm:"" json:"time_interval_days,omitempty"`                   // 每隔多少天

	// 上次执行基准
	LastMileage    float64    `gorm:"type:decimal(10,1)" json:"last_mileage,omitempty"`   // 上次保养时的里程
	LastDate       *time.Time `json:"last_date,omitempty"`                                 // 上次保养日期

	// 下次预计触发
	NextMileage    float64    `gorm:"type:decimal(10,1)" json:"next_mileage,omitempty"`   // 下次保养里程
	NextDate       *time.Time `json:"next_date,omitempty"`                                 // 下次保养日期

	// 状态
	IsEnabled bool `gorm:"default:true;not null" json:"is_enabled"`

	// 关联
	User    User    `gorm:"foreignKey:UserID" json:"-"`
	Vehicle Vehicle `gorm:"foreignKey:VehicleID" json:"-"`
}

// TableName 指定表名
func (Reminder) TableName() string {
	return "reminders"
}
