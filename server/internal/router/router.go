// Package router 负责注册所有 API 路由。
// 使用 Go 1.22 的 net/http.ServeMux 增强路由功能（支持 HTTP 方法匹配和路径参数）。
package router

import (
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/config"
	"gastrack/internal/handler"
	"gastrack/internal/middleware"
	"gastrack/internal/pkg/respond"
)

// New 创建并配置路由
func New(
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	vehicleHandler *handler.VehicleHandler,
	fuelRecordHandler *handler.FuelRecordHandler,
	statsHandler *handler.StatsHandler,
) http.Handler {
	mux := http.NewServeMux()

	// --- 公开路由（无需认证）---

	// 认证
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)

	// 健康检查
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, map[string]string{"status": "ok"})
	})

	// --- 需要认证的路由 ---
	auth := middleware.Auth(cfg.JWT.Secret)

	// 登出（需要认证）
	mux.Handle("POST /api/v1/auth/logout", auth(http.HandlerFunc(authHandler.Logout)))

	// 用户
	mux.Handle("GET /api/v1/users/me", auth(http.HandlerFunc(userHandler.GetProfile)))
	mux.Handle("PATCH /api/v1/users/me", auth(http.HandlerFunc(userHandler.UpdateProfile)))
	mux.Handle("PUT /api/v1/users/me/password", auth(http.HandlerFunc(userHandler.ChangePassword)))
	mux.Handle("DELETE /api/v1/users/me", auth(http.HandlerFunc(userHandler.DeleteAccount)))

	// 车辆
	mux.Handle("GET /api/v1/vehicles", auth(http.HandlerFunc(vehicleHandler.List)))
	mux.Handle("POST /api/v1/vehicles", auth(http.HandlerFunc(vehicleHandler.Create)))
	mux.Handle("GET /api/v1/vehicles/{id}", auth(http.HandlerFunc(vehicleHandler.GetByID)))
	mux.Handle("PATCH /api/v1/vehicles/{id}", auth(http.HandlerFunc(vehicleHandler.Update)))
	mux.Handle("DELETE /api/v1/vehicles/{id}", auth(http.HandlerFunc(vehicleHandler.Delete)))

	// 加油记录
	mux.Handle("GET /api/v1/vehicles/{id}/records", auth(http.HandlerFunc(fuelRecordHandler.List)))
	mux.Handle("POST /api/v1/vehicles/{id}/records", auth(http.HandlerFunc(fuelRecordHandler.Create)))
	mux.Handle("GET /api/v1/vehicles/{id}/records/{rid}", auth(http.HandlerFunc(fuelRecordHandler.GetByID)))
	mux.Handle("PATCH /api/v1/vehicles/{id}/records/{rid}", auth(http.HandlerFunc(fuelRecordHandler.Update)))
	mux.Handle("DELETE /api/v1/vehicles/{id}/records/{rid}", auth(http.HandlerFunc(fuelRecordHandler.Delete)))
	mux.Handle("GET /api/v1/vehicles/{id}/stations", auth(http.HandlerFunc(fuelRecordHandler.GetStationSuggestions)))

	// 统计
	mux.Handle("GET /api/v1/vehicles/{id}/stats", auth(http.HandlerFunc(statsHandler.GetVehicleStats)))
	mux.Handle("GET /api/v1/vehicles/{id}/efficiency-trend", auth(http.HandlerFunc(statsHandler.GetEfficiencyTrend)))
	mux.Handle("GET /api/v1/vehicles/{id}/period-stats", auth(http.HandlerFunc(statsHandler.GetPeriodStats)))
	mux.Handle("GET /api/v1/stats/overview", auth(http.HandlerFunc(statsHandler.GetOverview)))

	// --- 应用全局中间件 ---
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.Server.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept-Language"},
		MaxAge:         "86400",
	}

	// 中间件链：Recovery → Logger → CORS → RateLimit → 路由
	global := middleware.Chain(
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.CORS(corsConfig),
		middleware.RateLimit(100, 200), // 每秒 100 请求，突发 200
	)

	return global(mux)
}
