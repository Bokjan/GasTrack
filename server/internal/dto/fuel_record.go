package dto

import "time"

// --- 加油记录相关 DTO ---

// CreateFuelRecordRequest 添加加油记录请求
type CreateFuelRecordRequest struct {
	FuelAmount   float64 `json:"fuel_amount" validate:"required,gt=0"`
	FuelUnit     string  `json:"fuel_unit" validate:"omitempty,oneof=L gal kWh"`
	UnitPrice    float64 `json:"unit_price" validate:"omitempty,gte=0"`
	TotalCost    float64 `json:"total_cost" validate:"required,gt=0"`
	CurrencyCode string  `json:"currency_code" validate:"required,len=3"`
	Odometer     float64 `json:"odometer" validate:"required,gt=0"`
	DistanceUnit string  `json:"distance_unit" validate:"omitempty,oneof=km mi"`
	IsFullTank   bool    `json:"is_full_tank"`
	FuelGrade    string  `json:"fuel_grade" validate:"omitempty,max=20"`
	StationName  string  `json:"station_name" validate:"omitempty,max=200"`
	StationLat   float64 `json:"station_lat" validate:"omitempty"`
	StationLng   float64 `json:"station_lng" validate:"omitempty"`
	Note         string  `json:"note" validate:"omitempty,max=1000"`
	RefuelDate   string  `json:"refuel_date" validate:"required"` // ISO 8601
}

// UpdateFuelRecordRequest 编辑加油记录请求
type UpdateFuelRecordRequest struct {
	FuelAmount   *float64 `json:"fuel_amount" validate:"omitempty,gt=0"`
	FuelUnit     *string  `json:"fuel_unit" validate:"omitempty,oneof=L gal kWh"`
	UnitPrice    *float64 `json:"unit_price" validate:"omitempty,gte=0"`
	TotalCost    *float64 `json:"total_cost" validate:"omitempty,gt=0"`
	CurrencyCode *string  `json:"currency_code" validate:"omitempty,len=3"`
	Odometer     *float64 `json:"odometer" validate:"omitempty,gt=0"`
	DistanceUnit *string  `json:"distance_unit" validate:"omitempty,oneof=km mi"`
	IsFullTank   *bool    `json:"is_full_tank"`
	FuelGrade    *string  `json:"fuel_grade" validate:"omitempty,max=20"`
	StationName  *string  `json:"station_name" validate:"omitempty,max=200"`
	StationLat   *float64 `json:"station_lat"`
	StationLng   *float64 `json:"station_lng"`
	Note         *string  `json:"note" validate:"omitempty,max=1000"`
	RefuelDate   *string  `json:"refuel_date"` // ISO 8601
}

// FuelRecordResponse 加油记录响应
type FuelRecordResponse struct {
	ID             string    `json:"id"`
	VehicleID      string    `json:"vehicle_id"`
	FuelAmount     float64   `json:"fuel_amount"`
	FuelUnit       string    `json:"fuel_unit"`
	UnitPrice      float64   `json:"unit_price,omitempty"`
	TotalCost      float64   `json:"total_cost"`
	CurrencyCode   string    `json:"currency_code"`
	Odometer       float64   `json:"odometer"`
	DistanceUnit   string    `json:"distance_unit"`
	IsFullTank     bool      `json:"is_full_tank"`
	FuelGrade      string    `json:"fuel_grade,omitempty"`
	StationName    string    `json:"station_name,omitempty"`
	StationLat     float64   `json:"station_lat,omitempty"`
	StationLng     float64   `json:"station_lng,omitempty"`
	Note           string    `json:"note,omitempty"`
	ReceiptURL     string    `json:"receipt_url,omitempty"`
	TripDistance   float64   `json:"trip_distance,omitempty"`
	FuelEfficiency float64   `json:"fuel_efficiency,omitempty"`
	RefuelDate     time.Time `json:"refuel_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
