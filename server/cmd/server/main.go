// GasTrack 后端服务入口
// 负责初始化配置、日志、数据库，组装依赖并启动 HTTP 服务器。
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gastrack/internal/config"
	"gastrack/internal/database"
	"gastrack/internal/handler"
	"gastrack/internal/repository"
	"gastrack/internal/router"
	"gastrack/internal/service"
)

func main() {
	// 1. 加载配置
	configPath := os.Getenv("GASTRACK_CONFIG")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger, err := initLogger(cfg.Log)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting GasTrack server",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// 3. 连接数据库
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	logger.Info("database connected successfully")

	// 4. 创建 Repository 层
	userRepo := repository.NewUserRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	fuelRecordRepo := repository.NewFuelRecordRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// 5. 创建 Service 层
	authService := service.NewAuthService(userRepo, refreshTokenRepo, &cfg.JWT, logger)
	userService := service.NewUserService(userRepo, logger)
	vehicleService := service.NewVehicleService(vehicleRepo, logger)
	fuelRecordService := service.NewFuelRecordService(fuelRecordRepo, vehicleRepo, logger)
	statsService := service.NewStatsService(fuelRecordRepo, vehicleRepo, userRepo, logger)

	// 6. 创建 Handler 层
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	vehicleHandler := handler.NewVehicleHandler(vehicleService, logger)
	fuelRecordHandler := handler.NewFuelRecordHandler(fuelRecordService, logger)
	statsHandler := handler.NewStatsHandler(statsService, logger)

	// 7. 注册路由
	mux := router.New(
		cfg,
		logger,
		authHandler,
		userHandler,
		vehicleHandler,
		fuelRecordHandler,
		statsHandler,
	)

	// 8. 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// 9. 优雅关闭
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", zap.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
		}
	}

	// 优雅关闭（等待进行中的请求完成）
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	// 关闭数据库连接
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	logger.Info("GasTrack server stopped gracefully")
}

// initLogger 根据配置初始化 zap logger
func initLogger(cfg config.LogConfig) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	if cfg.Format == "console" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	return zapCfg.Build()
}
