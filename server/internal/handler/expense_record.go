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

// ExpenseRecordHandler 开销记录相关 HTTP 处理器
type ExpenseRecordHandler struct {
	expenseService *service.ExpenseRecordService
	logger         *zap.Logger
}

// NewExpenseRecordHandler 创建 ExpenseRecordHandler 实例
func NewExpenseRecordHandler(expenseService *service.ExpenseRecordService, logger *zap.Logger) *ExpenseRecordHandler {
	return &ExpenseRecordHandler{expenseService: expenseService, logger: logger}
}

// List 获取车辆的开销记录列表
// GET /api/v1/vehicles/{id}/expenses
func (h *ExpenseRecordHandler) List(w http.ResponseWriter, r *http.Request) {
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
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &dto.ExpenseListFilter{
		Page:      page,
		PageSize:  pageSize,
		Category:  r.URL.Query().Get("category"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Keyword:   r.URL.Query().Get("keyword"),
		MinAmount: float64(decode.QueryInt(r, "min_amount", 0)),
		MaxAmount: float64(decode.QueryInt(r, "max_amount", 0)),
	}

	records, total, err := h.expenseService.List(r.Context(), userID, vehicleID, filter)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Paged(w, records, page, pageSize, total)
}

// Create 添加开销记录
// POST /api/v1/vehicles/{id}/expenses
func (h *ExpenseRecordHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateExpenseRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.expenseService.Create(r.Context(), userID, vehicleID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.Created(w, result)
}

// GetByID 获取开销记录详情
// GET /api/v1/vehicles/{id}/expenses/{eid}
func (h *ExpenseRecordHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	expenseID, err := decode.PathParamUUID(r, "eid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.expenseService.GetByID(r.Context(), expenseID, vehicleID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Update 编辑开销记录
// PATCH /api/v1/vehicles/{id}/expenses/{eid}
func (h *ExpenseRecordHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	expenseID, err := decode.PathParamUUID(r, "eid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	var req dto.UpdateExpenseRequest
	if err := decode.JSON(r, &req); err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	result, err := h.expenseService.Update(r.Context(), expenseID, vehicleID, userID, &req)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// Delete 删除开销记录
// DELETE /api/v1/vehicles/{id}/expenses/{eid}
func (h *ExpenseRecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	expenseID, err := decode.PathParamUUID(r, "eid")
	if err != nil {
		respond.BadRequest(w, err.Error())
		return
	}

	if err := h.expenseService.Delete(r.Context(), expenseID, vehicleID, userID); err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.NoContent(w)
}

// GetStats 获取开销统计
// GET /api/v1/vehicles/{id}/expense-stats
func (h *ExpenseRecordHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.expenseService.GetStats(r.Context(), userID, vehicleID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// GetVendorSuggestions 获取商家名称建议
// GET /api/v1/vehicles/{id}/expense-vendors
func (h *ExpenseRecordHandler) GetVendorSuggestions(w http.ResponseWriter, r *http.Request) {
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

	names, err := h.expenseService.GetVendorSuggestions(r.Context(), userID, vehicleID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, names)
}
