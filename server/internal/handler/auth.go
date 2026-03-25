// Package handler 提供 HTTP 请求处理器。
// Handler 负责解析请求、调用 Service 层、返回 HTTP 响应。
package handler

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/dto"
	"gastrack/internal/middleware"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// AuthHandler 认证相关 HTTP 处理器
type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

// Register 处理用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// Login 处理用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// RefreshToken 处理 Token 刷新
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Logout 处理用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	if err := h.authService.Logout(r.Context(), userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// handleAppError 统一处理 AppError，将业务错误转为 HTTP 响应
func handleAppError(w http.ResponseWriter, logger *zap.Logger, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		// 5xx 错误记录日志
		if appErr.Code >= 500 {
			logger.Error("internal error",
				zap.String("message", appErr.Message),
				zap.Error(appErr.Err),
			)
		}
		respond.Error(w, appErr.Code, appErr.BizCode, appErr.Message)
		return
	}

	// 未知错误视为 500
	logger.Error("unexpected error", zap.Error(err))
	respond.InternalError(w, "internal server error")
}
