package model

import (
	"time"

	"github.com/google/uuid"
)

// FuelRecord 加油记录模型
type FuelRecord struct {
	BaseModel

	VehicleID uuid.UUID `gorm:"type:uuid;not null;index:idx_fuel_records_vehicle" json:"vehicle_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_fuel_records_user" json:"user_id"`

	// 加油数据（存原始值）
	FuelAmount   float64 `gorm:"type:decimal(8,3);not null" json:"fuel_amount"`    // 加油量
	FuelUnit     string  `gorm:"size:5;default:L" json:"fuel_unit"`                // L / gal / kWh
	UnitPrice    float64 `gorm:"type:decimal(10,4)" json:"unit_price,omitempty"`   // 单价
	TotalCost    float64 `gorm:"type:decimal(10,2);not null" json:"total_cost"`    // 总费用
	CurrencyCode string  `gorm:"size:3;not null" json:"currency_code"`             // 币种

	// 里程数据
	Odometer     float64 `gorm:"type:decimal(10,1);not null" json:"odometer"`      // 里程表读数
	DistanceUnit string  `gorm:"size:5;default:km" json:"distance_unit"`           // km / mi

	// 加油详情
	// 是否加满。不能声明 default:true：GORM 创建时会把零值 false 替换为默认值，导致"未加满"无法保存
	IsFullTank  bool    `json:"is_full_tank"`
	FuelGrade   string  `gorm:"size:20" json:"fuel_grade,omitempty"`               // 92/95/98/diesel
	StationName string  `gorm:"size:200;index:idx_fuel_records_station" json:"station_name,omitempty"` // 加油站
	StationLat  float64 `gorm:"type:decimal(10,7)" json:"station_lat,omitempty"`   // 纬度
	StationLng  float64 `gorm:"type:decimal(10,7)" json:"station_lng,omitempty"`   // 经度
	Note        string  `gorm:"type:text" json:"note,omitempty"`                   // 备注
	ReceiptURL  string  `gorm:"size:500" json:"receipt_url,omitempty"`             // 小票照片

	// 计算字段（冗余存储，提高查询性能）
	TripDistance   float64 `gorm:"type:decimal(10,1)" json:"trip_distance,omitempty"`   // 本次行驶距离
	FuelEfficiency float64 `gorm:"type:decimal(10,2)" json:"fuel_efficiency,omitempty"` // 油耗/电耗（L/100km 或 kWh/100km 存储基准，满油到满油区间计算）
	// 油耗区间里程（km）：本次加满与上一次加满之间的里程，用于按里程加权求平均油耗
	EfficiencyDistance float64 `gorm:"type:decimal(10,1)" json:"-"`

	RefuelDate time.Time `gorm:"not null;index:idx_fuel_records_vehicle;index:idx_fuel_records_user" json:"refuel_date"`

	// 关联
	Vehicle Vehicle `gorm:"foreignKey:VehicleID" json:"-"`
	User    User    `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (FuelRecord) TableName() string {
	return "fuel_records"
}
