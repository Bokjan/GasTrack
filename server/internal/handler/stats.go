package handler

import (
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/middleware"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// StatsHandler 统计相关 HTTP 处理器
type StatsHandler struct {
	statsService *service.StatsService
	logger       *zap.Logger
}

// NewStatsHandler 创建 StatsHandler 实例
func NewStatsHandler(statsService *service.StatsService, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{statsService: statsService, logger: logger}
}

// GetVehicleStats 获取车辆统计
// GET /api/v1/vehicles/{id}/stats
func (h *StatsHandler) GetVehicleStats(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.statsService.GetVehicleStats(r.Context(), vehicleID, userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// GetOverview 获取全局统计总览
// GET /api/v1/stats/overview
func (h *StatsHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	result, err := h.statsService.GetOverview(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}

// GetEfficiencyTrend 获取油耗趋势
// GET /api/v1/vehicles/{id}/efficiency-trend
func (h *StatsHandler) GetEfficiencyTrend(w http.ResponseWriter, r *http.Request) {
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

	limit := decode.QueryInt(r, "limit", 30)
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	result, err := h.statsService.GetEfficiencyTrend(r.Context(), vehicleID, userID, limit)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	respond.OK(w, result)
}
