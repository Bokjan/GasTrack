package model

import (
	"time"

	"github.com/google/uuid"
)

// InviteCode 邀请码模型
type InviteCode struct {
	BaseModel

	Code       string     `gorm:"uniqueIndex;size:20;not null" json:"code"`          // 邀请码，如 GT-A3X7K9
	CreatedBy  uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by"`        // 创建者
	UsedBy     *uuid.UUID `gorm:"type:uuid" json:"used_by,omitempty"`                // 单次码使用者（多次码此字段为最后一个）
	MaxUses    int        `gorm:"default:1;not null" json:"max_uses"`                // 最大使用次数
	UseCount   int        `gorm:"default:0;not null" json:"use_count"`               // 已使用次数
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`                              // 过期时间（NULL 永不过期）
	Note       string     `gorm:"size:255" json:"note,omitempty"`                    // 备注
	IsActive   bool       `gorm:"default:true;not null" json:"is_active"`            // 是否激活

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"-"`
}

// TableName 指定表名
func (InviteCode) TableName() string {
	return "invite_codes"
}

// IsValid 判断邀请码是否可用
func (ic *InviteCode) IsValid() bool {
	if !ic.IsActive {
		return false
	}
	if ic.MaxUses > 0 && ic.UseCount >= ic.MaxUses {
		return false
	}
	if ic.ExpiresAt != nil && time.Now().After(*ic.ExpiresAt) {
		return false
	}
	return true
}

// RemainingUses 剩余可用次数（0 表示无限制）
func (ic *InviteCode) RemainingUses() int {
	if ic.MaxUses <= 0 {
		return -1 // 无限制
	}
	remaining := ic.MaxUses - ic.UseCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
