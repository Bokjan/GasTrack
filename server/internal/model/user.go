package model

import (
	"time"

	"github.com/google/uuid"
)

// User 用户模型
type User struct {
	BaseModel

	Email              string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash       string     `gorm:"size:255;not null" json:"-"` // 不暴露给客户端
	Nickname           string     `gorm:"size:100;not null" json:"nickname"`
	AvatarURL          string     `gorm:"size:500" json:"avatar_url,omitempty"`
	Locale             string     `gorm:"size:10;default:en-US" json:"locale"`         // 偏好语言: zh-CN/en-US/ja-JP
	Timezone           string     `gorm:"size:50;default:UTC" json:"timezone"`          // 时区
	CountryCode        string     `gorm:"size:5" json:"country_code,omitempty"`         // ISO 3166-1 alpha-2
	CurrencyCode       string     `gorm:"size:3;default:USD" json:"currency_code"`      // ISO 4217
	ReferenceCurrency  string     `gorm:"size:3;default:''" json:"reference_currency"`  // 参考换算币种（空表示自动推导）
	UnitSystem         string     `gorm:"size:10;default:metric" json:"unit_system"`    // metric / imperial
	FuelEfficiencyUnit string     `gorm:"size:10;default:L/100km" json:"fuel_efficiency_unit"` // L/100km / km/L / MPG
	Status             string     `gorm:"size:20;default:active" json:"status"`         // active/suspended/deleted
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`

	// 关联
	Vehicles      []Vehicle      `gorm:"foreignKey:UserID" json:"vehicles,omitempty"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	BaseModel

	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash  string    `gorm:"size:255;not null" json:"-"`
	DeviceInfo string    `gorm:"size:255" json:"device_info,omitempty"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
