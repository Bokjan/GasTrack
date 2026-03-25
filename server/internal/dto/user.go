package dto

import "time"

// --- 用户相关 DTO ---

// UserResponse 用户信息响应
type UserResponse struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	Nickname           string     `json:"nickname"`
	AvatarURL          string     `json:"avatar_url,omitempty"`
	Locale             string     `json:"locale"`
	Timezone           string     `json:"timezone"`
	CountryCode        string     `json:"country_code,omitempty"`
	CurrencyCode       string     `json:"currency_code"`
	UnitSystem         string     `json:"unit_system"`
	FuelEfficiencyUnit string     `json:"fuel_efficiency_unit"`
	Status             string     `json:"status"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// UpdateUserRequest 更新用户资料请求
type UpdateUserRequest struct {
	Nickname           *string `json:"nickname" validate:"omitempty,min=1,max=100"`
	AvatarURL          *string `json:"avatar_url" validate:"omitempty,url"`
	Locale             *string `json:"locale" validate:"omitempty,oneof=en-US zh-CN ja-JP"`
	Timezone           *string `json:"timezone" validate:"omitempty,max=50"`
	CountryCode        *string `json:"country_code" validate:"omitempty,len=2"`
	CurrencyCode       *string `json:"currency_code" validate:"omitempty,len=3"`
	UnitSystem         *string `json:"unit_system" validate:"omitempty,oneof=metric imperial"`
	FuelEfficiencyUnit *string `json:"fuel_efficiency_unit" validate:"omitempty,oneof=L/100km km/L MPG"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}
