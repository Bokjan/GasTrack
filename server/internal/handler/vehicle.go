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

// VehicleHandler 车辆相关 HTTP 处理器
type VehicleHandler struct {
	vehicleService *service.VehicleService
	logger         *zap.Logger
}

// NewVehicleHandler 创建 VehicleHandler 实例
func NewVehicleHandler(vehicleService *service.VehicleService, logger *zap.Logger) *VehicleHandler {
	return &VehicleHandler{vehicleService: vehicleService, logger: logger}
}

// List 获取车辆列表
// GET /api/v1/vehicles
func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	includeArchived := decode.QueryString(r, "include_archived", "false") == "true"

	result, err := h.vehicleService.List(r.Context(), userID, includeArchived)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Create 添加车辆
// POST /api/v1/vehicles
func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	var req dto.CreateVehicleRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.vehicleService.Create(r.Context(), userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// GetByID 获取车辆详情
// GET /api/v1/vehicles/{id}
func (h *VehicleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.vehicleService.GetByID(r.Context(), vehicleID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 编辑车辆
// PATCH /api/v1/vehicles/{id}
func (h *VehicleHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateVehicleRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.vehicleService.Update(r.Context(), vehicleID, userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除车辆
// DELETE /api/v1/vehicles/{id}
func (h *VehicleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.vehicleService.Delete(r.Context(), vehicleID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
