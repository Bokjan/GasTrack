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
	inviteHandler *handler.InviteHandler,
	exportHandler *handler.ExportHandler,
	reminderHandler *handler.ReminderHandler,
	notificationHandler *handler.NotificationHandler,
	groupHandler *handler.GroupHandler,
	exchangeRateHandler *handler.ExchangeRateHandler,
	expenseRecordHandler *handler.ExpenseRecordHandler,
) http.Handler {
	mux := http.NewServeMux()

	// --- 公开路由（无需认证）---

	// 认证
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)

	// 邀请码验证（公开，注册前实时校验）
	mux.HandleFunc("GET /api/v1/invites/{code}", inviteHandler.Validate)

	// 健康检查
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, map[string]string{"status": "ok"})
	})

	// 注册模式查询（公开，前端根据此决定是否显示邀请码字段）
	registrationMode := cfg.Registration.Mode
	if registrationMode == "" {
		registrationMode = "invite_only"
	}
	mux.HandleFunc("GET /api/v1/auth/registration-mode", func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, map[string]string{"mode": registrationMode})
	})

	// --- 需要认证的路由 ---
	auth := middleware.Auth(cfg.JWT.Secret)

	// 登出（需要认证）
	mux.Handle("POST /api/v1/auth/logout", auth(http.HandlerFunc(authHandler.Logout)))

	// 邀请码管理（需要认证）
	mux.Handle("POST /api/v1/invites", auth(http.HandlerFunc(inviteHandler.Create)))
	mux.Handle("GET /api/v1/invites", auth(http.HandlerFunc(inviteHandler.List)))
	mux.Handle("PATCH /api/v1/invites/{id}", auth(http.HandlerFunc(inviteHandler.Update)))
	mux.Handle("DELETE /api/v1/invites/{id}", auth(http.HandlerFunc(inviteHandler.Delete)))

	// 用户
	mux.Handle("GET /api/v1/users/me", auth(http.HandlerFunc(userHandler.GetProfile)))
	mux.Handle("PATCH /api/v1/users/me", auth(http.HandlerFunc(userHandler.UpdateProfile)))
	mux.Handle("PUT /api/v1/users/me/password", auth(http.HandlerFunc(userHandler.ChangePassword)))
	mux.Handle("DELETE /api/v1/users/me", auth(http.HandlerFunc(userHandler.DeleteAccount)))

	// 数据导出（GDPR 数据可携带权）
	mux.Handle("GET /api/v1/users/me/export", auth(http.HandlerFunc(exportHandler.ExportMyData)))

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
	mux.Handle("GET /api/v1/vehicles/{id}/expense-period-stats", auth(http.HandlerFunc(statsHandler.GetExpensePeriodStats)))
	mux.Handle("GET /api/v1/stats/overview", auth(http.HandlerFunc(statsHandler.GetOverview)))

	// 汇率参考
	mux.Handle("GET /api/v1/exchange-rates", auth(http.HandlerFunc(exchangeRateHandler.GetRates)))

	// 保养提醒
	mux.Handle("GET /api/v1/reminders", auth(http.HandlerFunc(reminderHandler.List)))
	mux.Handle("POST /api/v1/reminders", auth(http.HandlerFunc(reminderHandler.Create)))
	mux.Handle("GET /api/v1/reminders/{id}", auth(http.HandlerFunc(reminderHandler.GetByID)))
	mux.Handle("PATCH /api/v1/reminders/{id}", auth(http.HandlerFunc(reminderHandler.Update)))
	mux.Handle("DELETE /api/v1/reminders/{id}", auth(http.HandlerFunc(reminderHandler.Delete)))

	// 通知
	mux.Handle("GET /api/v1/notifications", auth(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("GET /api/v1/notifications/unread-count", auth(http.HandlerFunc(notificationHandler.UnreadCount)))
	mux.Handle("PATCH /api/v1/notifications/{id}/read", auth(http.HandlerFunc(notificationHandler.MarkAsRead)))
	mux.Handle("POST /api/v1/notifications/read-all", auth(http.HandlerFunc(notificationHandler.MarkAllAsRead)))
	mux.Handle("DELETE /api/v1/notifications/{id}", auth(http.HandlerFunc(notificationHandler.Delete)))

	// 群组管理
	mux.Handle("GET /api/v1/groups", auth(http.HandlerFunc(groupHandler.List)))
	mux.Handle("POST /api/v1/groups", auth(http.HandlerFunc(groupHandler.Create)))
	mux.Handle("POST /api/v1/groups/join", auth(http.HandlerFunc(groupHandler.Join)))
	mux.Handle("GET /api/v1/groups/{id}", auth(http.HandlerFunc(groupHandler.GetByID)))
	mux.Handle("PATCH /api/v1/groups/{id}", auth(http.HandlerFunc(groupHandler.Update)))
	mux.Handle("DELETE /api/v1/groups/{id}", auth(http.HandlerFunc(groupHandler.Delete)))
	mux.Handle("POST /api/v1/groups/{id}/regenerate-invite", auth(http.HandlerFunc(groupHandler.RegenerateInviteCode)))
	mux.Handle("POST /api/v1/groups/{id}/leave", auth(http.HandlerFunc(groupHandler.LeaveGroup)))
	mux.Handle("GET /api/v1/groups/{id}/overview", auth(http.HandlerFunc(groupHandler.GetOverview)))
	mux.Handle("PATCH /api/v1/groups/{id}/members/{uid}", auth(http.HandlerFunc(groupHandler.UpdateMemberRole)))
	mux.Handle("DELETE /api/v1/groups/{id}/members/{uid}", auth(http.HandlerFunc(groupHandler.RemoveMember)))

	// 群组扩展功能
	mux.Handle("POST /api/v1/groups/{id}/shared-vehicles", auth(http.HandlerFunc(groupHandler.ShareVehicle)))
	mux.Handle("DELETE /api/v1/groups/{id}/shared-vehicles/{vid}", auth(http.HandlerFunc(groupHandler.UnshareVehicle)))
	mux.Handle("GET /api/v1/groups/{id}/shared-vehicles", auth(http.HandlerFunc(groupHandler.ListSharedVehicles)))
	mux.Handle("GET /api/v1/groups/{id}/leaderboard", auth(http.HandlerFunc(groupHandler.GetLeaderboard)))
	mux.Handle("GET /api/v1/groups/{id}/expense-stats", auth(http.HandlerFunc(groupHandler.GetExpenseStats)))
	mux.Handle("GET /api/v1/groups/{id}/stations", auth(http.HandlerFunc(groupHandler.GetStationStats)))

	// 开销记录
	mux.Handle("GET /api/v1/vehicles/{id}/expenses", auth(http.HandlerFunc(expenseRecordHandler.List)))
	mux.Handle("POST /api/v1/vehicles/{id}/expenses", auth(http.HandlerFunc(expenseRecordHandler.Create)))
	mux.Handle("GET /api/v1/vehicles/{id}/expenses/{eid}", auth(http.HandlerFunc(expenseRecordHandler.GetByID)))
	mux.Handle("PATCH /api/v1/vehicles/{id}/expenses/{eid}", auth(http.HandlerFunc(expenseRecordHandler.Update)))
	mux.Handle("DELETE /api/v1/vehicles/{id}/expenses/{eid}", auth(http.HandlerFunc(expenseRecordHandler.Delete)))
	mux.Handle("GET /api/v1/vehicles/{id}/expense-stats", auth(http.HandlerFunc(expenseRecordHandler.GetStats)))
	mux.Handle("GET /api/v1/vehicles/{id}/expense-vendors", auth(http.HandlerFunc(expenseRecordHandler.GetVendorSuggestions)))

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
