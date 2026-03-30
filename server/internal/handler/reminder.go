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

// ReminderHandler 提醒相关 HTTP 处理器
type ReminderHandler struct {
	reminderService *service.ReminderService
	logger          *zap.Logger
}

// NewReminderHandler 创建 ReminderHandler 实例
func NewReminderHandler(reminderService *service.ReminderService, logger *zap.Logger) *ReminderHandler {
	return &ReminderHandler{reminderService: reminderService, logger: logger}
}

// List 获取提醒列表
// GET /api/v1/reminders
func (h *ReminderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	result, err := h.reminderService.List(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Create 创建提醒
// POST /api/v1/reminders
func (h *ReminderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.CreateReminderRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.reminderService.Create(r.Context(), userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// GetByID 获取提醒详情
// GET /api/v1/reminders/{id}
func (h *ReminderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	reminderID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.reminderService.GetByID(r.Context(), reminderID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 更新提醒
// PATCH /api/v1/reminders/{id}
func (h *ReminderHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	reminderID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateReminderRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.reminderService.Update(r.Context(), reminderID, userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除提醒
// DELETE /api/v1/reminders/{id}
func (h *ReminderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	reminderID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.reminderService.Delete(r.Context(), reminderID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
