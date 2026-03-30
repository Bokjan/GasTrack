// Package dto 定义请求和响应的数据传输对象。
// DTO 负责 API 层的数据序列化/反序列化与校验，与 model 层解耦。
package dto

// --- 认证相关 DTO ---

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8,max=72"`
	Nickname   string `json:"nickname" validate:"required,min=1,max=100"`
	Locale     string `json:"locale" validate:"omitempty,oneof=en-US zh-CN ja-JP"`
	InviteCode string `json:"invite_code" validate:"omitempty"` // 邀请码（invite_only 模式下必填）
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse 认证响应（登录/注册成功后返回）
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // 秒
	User         UserResponse `json:"user"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}
