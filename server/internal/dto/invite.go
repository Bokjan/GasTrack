package dto

import "time"

// --- 邀请码相关 DTO ---

// CreateInviteRequest 创建邀请码请求
type CreateInviteRequest struct {
	MaxUses   int        `json:"max_uses" validate:"omitempty,min=1"`    // 最大使用次数，默认 1
	ExpiresAt *time.Time `json:"expires_at" validate:"omitempty"`        // 过期时间
	Note      string     `json:"note" validate:"omitempty,max=255"`      // 备注
}

// UpdateInviteRequest 更新邀请码请求
type UpdateInviteRequest struct {
	IsActive *bool   `json:"is_active" validate:"omitempty"`          // 激活/禁用
	Note     *string `json:"note" validate:"omitempty,max=255"`       // 备注
}

// InviteCodeResponse 邀请码响应
type InviteCodeResponse struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	CreatedBy     string     `json:"created_by"`
	CreatorName   string     `json:"creator_name,omitempty"`
	MaxUses       int        `json:"max_uses"`
	UseCount      int        `json:"use_count"`
	RemainingUses int        `json:"remaining_uses"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Note          string     `json:"note,omitempty"`
	IsActive      bool       `json:"is_active"`
	IsValid       bool       `json:"is_valid"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ValidateInviteResponse 邀请码验证响应（公开接口）
type ValidateInviteResponse struct {
	Valid         bool       `json:"valid"`
	RemainingUses int       `json:"remaining_uses,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}
