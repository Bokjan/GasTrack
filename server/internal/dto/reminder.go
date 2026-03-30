package dto

import "time"

// --- 提醒相关 DTO ---

// CreateReminderRequest 创建提醒请求
type CreateReminderRequest struct {
	VehicleID        string  `json:"vehicle_id" validate:"required,uuid"`
	Category         string  `json:"category" validate:"required,oneof=oil_change tire_rotation brake_pads air_filter transmission coolant spark_plugs battery tire_replace inspection custom"`
	Title            string  `json:"title" validate:"required,min=1,max=200"`
	Description      string  `json:"description" validate:"omitempty,max=1000"`
	Trigger          string  `json:"trigger" validate:"required,oneof=mileage time both"`
	MileageInterval  float64 `json:"mileage_interval" validate:"omitempty,gt=0"`
	TimeIntervalDays int     `json:"time_interval_days" validate:"omitempty,gt=0"`
	LastMileage      float64 `json:"last_mileage" validate:"omitempty,gte=0"`
	LastDate         string  `json:"last_date" validate:"omitempty"` // ISO 8601
}

// UpdateReminderRequest 更新提醒请求
type UpdateReminderRequest struct {
	Category         *string  `json:"category" validate:"omitempty,oneof=oil_change tire_rotation brake_pads air_filter transmission coolant spark_plugs battery tire_replace inspection custom"`
	Title            *string  `json:"title" validate:"omitempty,min=1,max=200"`
	Description      *string  `json:"description" validate:"omitempty,max=1000"`
	Trigger          *string  `json:"trigger" validate:"omitempty,oneof=mileage time both"`
	MileageInterval  *float64 `json:"mileage_interval" validate:"omitempty,gt=0"`
	TimeIntervalDays *int     `json:"time_interval_days" validate:"omitempty,gt=0"`
	LastMileage      *float64 `json:"last_mileage" validate:"omitempty,gte=0"`
	LastDate         *string  `json:"last_date"`
	IsEnabled        *bool    `json:"is_enabled"`
}

// ReminderResponse 提醒响应
type ReminderResponse struct {
	ID               string     `json:"id"`
	VehicleID        string     `json:"vehicle_id"`
	VehicleName      string     `json:"vehicle_name"`
	Type             string     `json:"type"`
	Category         string     `json:"category"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	Trigger          string     `json:"trigger"`
	MileageInterval  float64    `json:"mileage_interval,omitempty"`
	TimeIntervalDays int        `json:"time_interval_days,omitempty"`
	LastMileage      float64    `json:"last_mileage,omitempty"`
	LastDate         *time.Time `json:"last_date,omitempty"`
	NextMileage      float64    `json:"next_mileage,omitempty"`
	NextDate         *time.Time `json:"next_date,omitempty"`
	IsEnabled        bool       `json:"is_enabled"`
	IsOverdue        bool       `json:"is_overdue"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
