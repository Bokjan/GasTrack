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

// FuelRecordHandler 加油记录相关 HTTP 处理器
type FuelRecordHandler struct {
	recordService *service.FuelRecordService
	logger        *zap.Logger
}

// NewFuelRecordHandler 创建 FuelRecordHandler 实例
func NewFuelRecordHandler(recordService *service.FuelRecordService, logger *zap.Logger) *FuelRecordHandler {
	return &FuelRecordHandler{recordService: recordService, logger: logger}
}

// List 获取车辆的加油记录列表
// GET /api/v1/vehicles/{id}/records
func (h *FuelRecordHandler) List(w http.ResponseWriter, r *http.Request) {
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

	page := decode.QueryInt(r, "page", 1)
	pageSize := decode.QueryInt(r, "page_size", 20)

	// 限制 page_size 范围
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	records, total, err := h.recordService.List(r.Context(), userID, vehicleID, page, pageSize)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Paged(w, records, page, pageSize, total)
}

// Create 添加加油记录
// POST /api/v1/vehicles/{id}/records
func (h *FuelRecordHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateFuelRecordRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.recordService.Create(r.Context(), userID, vehicleID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// GetByID 获取加油记录详情
// GET /api/v1/vehicles/{id}/records/{rid}
func (h *FuelRecordHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	recordID, err := decode.PathParamUUID(r, "rid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.recordService.GetByID(r.Context(), recordID, vehicleID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 编辑加油记录
// PATCH /api/v1/vehicles/{id}/records/{rid}
func (h *FuelRecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	recordID, err := decode.PathParamUUID(r, "rid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateFuelRecordRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.recordService.Update(r.Context(), recordID, vehicleID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除加油记录
// DELETE /api/v1/vehicles/{id}/records/{rid}
func (h *FuelRecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := decode.PathParamUUID(r, "id")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	recordID, err := decode.PathParamUUID(r, "rid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.recordService.Delete(r.Context(), recordID, vehicleID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}
