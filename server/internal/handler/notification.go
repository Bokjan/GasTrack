package handler

import (
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/middleware"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// NotificationHandler 通知相关 HTTP 处理器
type NotificationHandler struct {
	notificationService *service.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler 创建 NotificationHandler 实例
func NewNotificationHandler(notificationService *service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService, logger: logger}
}

// List 获取通知列表
// GET /api/v1/notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	result, err := h.notificationService.List(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// UnreadCount 获取未读通知数
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	count, err := h.notificationService.UnreadCount(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, map[string]int64{"count": count})
}

// MarkAsRead 标记通知为已读
// PATCH /api/v1/notifications/{id}/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	notificationID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), notificationID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// MarkAllAsRead 标记所有通知为已读
// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// Delete 删除通知
// DELETE /api/v1/notifications/{id}
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	notificationID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.notificationService.Delete(r.Context(), notificationID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
