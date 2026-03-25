package handler

import (
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/dto"
	"gastrack/internal/middleware"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// UserHandler 用户相关 HTTP 处理器
type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

// GetProfile 获取当前用户资料
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	result, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// UpdateProfile 更新当前用户资料
// PATCH /api/v1/users/me
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.UpdateUserRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.userService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// ChangePassword 修改密码
// PUT /api/v1/users/me/password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.ChangePasswordRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.userService.ChangePassword(r.Context(), userID, &req); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// DeleteAccount 注销账号
// DELETE /api/v1/users/me
func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	if err := h.userService.DeleteAccount(r.Context(), userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
