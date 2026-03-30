package handler

import (
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gastrack/internal/dto"
	"gastrack/internal/middleware"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// InviteHandler 邀请码相关 HTTP 处理器
type InviteHandler struct {
	inviteService *service.InviteService
	logger        *zap.Logger
}

// NewInviteHandler 创建 InviteHandler 实例
func NewInviteHandler(inviteService *service.InviteService, logger *zap.Logger) *InviteHandler {
	return &InviteHandler{inviteService: inviteService, logger: logger}
}

// Create 创建邀请码
// POST /api/v1/invites
func (h *InviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.CreateInviteRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.inviteService.Create(r.Context(), userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// List 查询我创建的邀请码列表
// GET /api/v1/invites
func (h *InviteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	results, err := h.inviteService.List(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, results)
}

// Validate 验证邀请码是否有效（公开接口）
// GET /api/v1/invites/{code}
func (h *InviteHandler) Validate(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		respond.BadRequest(w, "invite code is required")
		return
	}

	result, err := h.inviteService.GetByCode(r.Context(), code)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 更新邀请码
// PATCH /api/v1/invites/{id}
func (h *InviteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respond.BadRequest(w, "invalid invite id")
		return
	}

	var req dto.UpdateInviteRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.inviteService.Update(r.Context(), id, userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除邀请码
// DELETE /api/v1/invites/{id}
func (h *InviteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respond.BadRequest(w, "invalid invite id")
		return
	}

	if err := h.inviteService.Delete(r.Context(), id, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
