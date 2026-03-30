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

// GroupHandler 群组相关 HTTP 处理器
type GroupHandler struct {
	groupService *service.GroupService
	logger       *zap.Logger
}

// NewGroupHandler 创建 GroupHandler 实例
func NewGroupHandler(groupService *service.GroupService, logger *zap.Logger) *GroupHandler {
	return &GroupHandler{groupService: groupService, logger: logger}
}

// Create 创建群组
// POST /api/v1/groups
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.CreateGroupRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.Create(r.Context(), userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// List 获取我所在的群组列表
// GET /api/v1/groups
func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	results, err := h.groupService.List(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, results)
}

// GetByID 获取群组详情
// GET /api/v1/groups/{id}
func (h *GroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.GetByID(r.Context(), groupID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 更新群组信息
// PATCH /api/v1/groups/{id}
func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateGroupRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.Update(r.Context(), groupID, userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除群组
// DELETE /api/v1/groups/{id}
func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.groupService.Delete(r.Context(), groupID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// Join 通过邀请码加入群组
// POST /api/v1/groups/join
func (h *GroupHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req struct {
		InviteCode string `json:"invite_code" validate:"required"`
	}
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.JoinByInviteCode(r.Context(), userID, req.InviteCode)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// RegenerateInviteCode 重新生成邀请码
// POST /api/v1/groups/{id}/regenerate-invite
func (h *GroupHandler) RegenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.RegenerateInviteCode(r.Context(), groupID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// UpdateMemberRole 更新成员角色
// PATCH /api/v1/groups/{id}/members/{uid}
func (h *GroupHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	targetUserID, err := decode.PathParamUUID(r, "uid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.groupService.UpdateMemberRole(r.Context(), groupID, targetUserID, userID, &req); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// RemoveMember 移除群组成员
// DELETE /api/v1/groups/{id}/members/{uid}
func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	targetUserID, err := decode.PathParamUUID(r, "uid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.groupService.RemoveMember(r.Context(), groupID, targetUserID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// LeaveGroup 退出群组
// POST /api/v1/groups/{id}/leave
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	// 退出群组就是移除自己
	if err := h.groupService.RemoveMember(r.Context(), groupID, userID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// GetOverview 获取群组数据汇总
// GET /api/v1/groups/{id}/overview
func (h *GroupHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	groupID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.groupService.GetOverview(r.Context(), groupID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}
