package model

import (
	"github.com/google/uuid"
)

// VehicleType 车辆类型
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeOther      VehicleType = "other"
)

// FuelType 燃油类型
type FuelType string

const (
	FuelTypeGasoline FuelType = "gasoline"
	FuelTypeDiesel   FuelType = "diesel"
	FuelTypeHybrid   FuelType = "hybrid"
	FuelTypeElectric FuelType = "electric"
)

// Vehicle 车辆模型
type Vehicle struct {
	BaseModel

	UserID       uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	Name         string      `gorm:"size:100;not null" json:"name"`                       // 用户自定义名称
	VehicleType  VehicleType `gorm:"size:20;not null;default:car" json:"vehicle_type"`     // car/motorcycle/other
	Brand        string      `gorm:"size:100" json:"brand,omitempty"`                      // 品牌
	Model        string      `gorm:"size:100" json:"model,omitempty"`                      // 型号
	Year         int         `gorm:"" json:"year,omitempty"`                               // 年份
	FuelType     FuelType    `gorm:"size:20;not null" json:"fuel_type"`                    // 燃油类型
	TankCapacity    float64     `gorm:"type:decimal(6,2)" json:"tank_capacity,omitempty"`     // 油箱容量（升）
	BatteryCapacity float64     `gorm:"type:decimal(6,2)" json:"battery_capacity,omitempty"` // 电池容量（kWh），电动车使用
	EngineCC        int         `gorm:"" json:"engine_cc,omitempty"`                         // 排量(cc)，燃油/混动车辆通用
	LicensePlate string      `gorm:"size:20" json:"license_plate,omitempty"`               // 车牌号
	PhotoURL     string      `gorm:"size:500" json:"photo_url,omitempty"`                  // 照片
	IsDefault    bool        `gorm:"default:false" json:"is_default"`                      // 是否默认车辆
	IsArchived   bool        `gorm:"default:false" json:"is_archived"`                     // 是否归档

	// 关联
	User        User         `gorm:"foreignKey:UserID" json:"-"`
	FuelRecords []FuelRecord `gorm:"foreignKey:VehicleID" json:"fuel_records,omitempty"`
}

// TableName 指定表名
func (Vehicle) TableName() string {
	return "vehicles"
}
