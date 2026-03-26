package dto

import "time"

// --- 车辆相关 DTO ---

// CreateVehicleRequest 添加车辆请求
type CreateVehicleRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=100"`
	VehicleType  string  `json:"vehicle_type" validate:"required,oneof=car motorcycle other"`
	Brand        string  `json:"brand" validate:"omitempty,max=100"`
	Model        string  `json:"model" validate:"omitempty,max=100"`
	Year         int     `json:"year" validate:"omitempty,min=1900,max=2100"`
	FuelType     string  `json:"fuel_type" validate:"required,oneof=gasoline diesel hybrid electric"`
	FuelGrade    string  `json:"fuel_grade" validate:"omitempty,max=20"`
	TankCapacity    float64 `json:"tank_capacity" validate:"omitempty,gt=0"`
	BatteryCapacity float64 `json:"battery_capacity" validate:"omitempty,gt=0"` // 电池容量(kWh)
	EngineCC        int     `json:"engine_cc" validate:"omitempty,gt=0"`        // 排量(cc)
	LicensePlate string  `json:"license_plate" validate:"omitempty,max=20"`
	IsDefault    bool    `json:"is_default"`
}

// UpdateVehicleRequest 编辑车辆请求
type UpdateVehicleRequest struct {
	Name         *string  `json:"name" validate:"omitempty,min=1,max=100"`
	VehicleType  *string  `json:"vehicle_type" validate:"omitempty,oneof=car motorcycle other"`
	Brand        *string  `json:"brand" validate:"omitempty,max=100"`
	Model        *string  `json:"model" validate:"omitempty,max=100"`
	Year         *int     `json:"year" validate:"omitempty,min=1900,max=2100"`
	FuelType     *string  `json:"fuel_type" validate:"omitempty,oneof=gasoline diesel hybrid electric"`
	FuelGrade    *string  `json:"fuel_grade" validate:"omitempty,max=20"`
	TankCapacity    *float64 `json:"tank_capacity" validate:"omitempty,gt=0"`
	BatteryCapacity *float64 `json:"battery_capacity" validate:"omitempty,gt=0"`
	EngineCC        *int     `json:"engine_cc" validate:"omitempty,gt=0"`
	LicensePlate *string  `json:"license_plate" validate:"omitempty,max=20"`
	IsDefault    *bool    `json:"is_default"`
	IsArchived   *bool    `json:"is_archived"`
}

// VehicleResponse 车辆响应
type VehicleResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	VehicleType  string    `json:"vehicle_type"`
	Brand        string    `json:"brand,omitempty"`
	Model        string    `json:"model,omitempty"`
	Year         int       `json:"year,omitempty"`
	FuelType     string    `json:"fuel_type"`
	FuelGrade    string    `json:"fuel_grade,omitempty"`
	TankCapacity    float64   `json:"tank_capacity,omitempty"`
	BatteryCapacity float64   `json:"battery_capacity,omitempty"`
	EngineCC        int       `json:"engine_cc,omitempty"`
	LicensePlate string    `json:"license_plate,omitempty"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	IsDefault    bool      `json:"is_default"`
	IsArchived   bool      `json:"is_archived"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
