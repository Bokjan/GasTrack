package model

import (
	"time"

	"github.com/google/uuid"
)

// ExpenseCategory 开销类别
type ExpenseCategory string

const (
	ExpenseCategoryMaintenance ExpenseCategory = "maintenance" // 保养
	ExpenseCategoryRepair      ExpenseCategory = "repair"      // 维修
	ExpenseCategoryInsurance   ExpenseCategory = "insurance"   // 保险
	ExpenseCategoryParking     ExpenseCategory = "parking"     // 停车
	ExpenseCategoryToll        ExpenseCategory = "toll"        // 路桥/过路费
	ExpenseCategoryCarWash     ExpenseCategory = "car_wash"    // 洗车
	ExpenseCategoryInspection  ExpenseCategory = "inspection"  // 年检
	ExpenseCategoryParts       ExpenseCategory = "parts"       // 配件
	ExpenseCategoryFine        ExpenseCategory = "fine"        // 罚单
	ExpenseCategoryTax         ExpenseCategory = "tax"         // 税金
	ExpenseCategoryOther       ExpenseCategory = "other"       // 其他
)

// ExpenseRecord 开销记录模型
type ExpenseRecord struct {
	BaseModel

	VehicleID uuid.UUID `gorm:"type:uuid;not null;index:idx_expense_records_vehicle_date" json:"vehicle_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_expense_records_user" json:"user_id"`

	// 开销基本信息
	Category            ExpenseCategory     `gorm:"size:20;not null;index:idx_expense_records_vehicle_category" json:"category"`
	MaintenanceCategory MaintenanceCategory `gorm:"size:500" json:"maintenance_category,omitempty"` // 逗号分隔，支持多选
	Title               string              `gorm:"size:200;not null" json:"title"`
	Amount              float64             `gorm:"type:decimal(10,2);not null" json:"amount"`
	CurrencyCode        string              `gorm:"size:3;not null" json:"currency_code"`

	// 详细信息
	VendorName   string  `gorm:"size:200" json:"vendor_name,omitempty"`            // 商家/服务商
	Odometer     float64 `gorm:"type:decimal(10,1)" json:"odometer,omitempty"`     // 里程表读数
	DistanceUnit string  `gorm:"size:5;default:km" json:"distance_unit,omitempty"` // km / mi
	Note         string  `gorm:"type:text" json:"note,omitempty"`                  // 备注
	ReceiptURL   string  `gorm:"size:500" json:"receipt_url,omitempty"`            // 凭证图片（预留）

	// 日期
	ExpenseDate time.Time `gorm:"not null;index:idx_expense_records_vehicle_date" json:"expense_date"`

	// 保养提醒联动
	ReminderID *uuid.UUID `gorm:"type:uuid;index:idx_expense_records_reminder" json:"reminder_id,omitempty"`

	// 关联
	Vehicle  Vehicle   `gorm:"foreignKey:VehicleID" json:"-"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Reminder *Reminder `gorm:"foreignKey:ReminderID" json:"-"`
}

// TableName 指定表名
func (ExpenseRecord) TableName() string {
	return "expense_records"
}
