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
	"gopkg.in/natefinch/lumberjack.v2"

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
	if cfg.UsesSampleJWTSecret() {
		logger.Warn("jwt.secret is still the sample value from config.yaml; never use it in production")
	}

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
	inviteCodeRepo := repository.NewInviteCodeRepository(db)
	reminderRepo := repository.NewReminderRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	expenseRecordRepo := repository.NewExpenseRecordRepository(db)

	// 5. 创建 Service 层
	// 汇率参考服务（无需数据库，纯 API 缓存）：统计中的多币种金额按其换算为用户币种
	exchangeRateService := service.NewExchangeRateService(cfg.ExchangeRate, logger)
	inviteService := service.NewInviteService(inviteCodeRepo, userRepo, logger)
	reminderService := service.NewReminderService(reminderRepo, vehicleRepo, groupRepo, logger)
	notificationService := service.NewNotificationService(notificationRepo, fuelRecordRepo, reminderRepo, logger)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, inviteService, notificationService, &cfg.JWT, cfg.Registration.Mode, logger)
	userService := service.NewUserService(userRepo, authService, logger)
	vehicleService := service.NewVehicleService(vehicleRepo, groupRepo, logger)
	fuelRecordService := service.NewFuelRecordService(fuelRecordRepo, vehicleRepo, userRepo, groupRepo, logger, notificationService)
	statsService := service.NewStatsService(fuelRecordRepo, vehicleRepo, userRepo, groupRepo, expenseRecordRepo, exchangeRateService, logger)
	exportService := service.NewExportService(userRepo, vehicleRepo, fuelRecordRepo, expenseRecordRepo, reminderRepo, notificationRepo, inviteCodeRepo, groupRepo, logger)
	groupService := service.NewGroupService(groupRepo, userRepo, vehicleRepo, fuelRecordRepo, exchangeRateService, logger)
	expenseRecordService := service.NewExpenseRecordService(expenseRecordRepo, vehicleRepo, groupRepo, reminderRepo, userRepo, exchangeRateService, logger)

	// 运维命令：gastrack-server <command>，执行完毕即退出，不启动 HTTP 服务
	if len(os.Args) > 1 {
		if err := runMaintenance(context.Background(), os.Args[1], userService, fuelRecordService, logger); err != nil {
			logger.Fatal("maintenance command failed", zap.String("command", os.Args[1]), zap.Error(err))
		}
		return
	}

	// 5b. 启动汇率定时刷新
	exchangeRateService.Start()
	defer exchangeRateService.Stop()

	// 6. 创建 Handler 层
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	vehicleHandler := handler.NewVehicleHandler(vehicleService, logger)
	fuelRecordHandler := handler.NewFuelRecordHandler(fuelRecordService, logger)
	statsHandler := handler.NewStatsHandler(statsService, logger)
	inviteHandler := handler.NewInviteHandler(inviteService, logger)
	exportHandler := handler.NewExportHandler(exportService, logger)
	reminderHandler := handler.NewReminderHandler(reminderService, logger)
	notificationHandler := handler.NewNotificationHandler(notificationService, logger)
	groupHandler := handler.NewGroupHandler(groupService, logger)
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateService, logger)
	expenseRecordHandler := handler.NewExpenseRecordHandler(expenseRecordService, logger)

	// 5c. 后台定时任务：清理过期 refresh token
	bgCtx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	go runPeriodically(bgCtx, time.Hour, func(ctx context.Context) {
		if err := authService.PurgeExpiredTokens(ctx); err != nil {
			logger.Warn("failed to purge expired refresh tokens", zap.Error(err))
		}
	})
	// 按时间触发的保养提醒：每小时检查一次到期提醒并生成通知
	go runPeriodically(bgCtx, time.Hour, notificationService.CheckTimeReminders)

	// 7. 注册路由
	mux, err := router.New(
		cfg,
		logger,
		authHandler,
		userHandler,
		vehicleHandler,
		fuelRecordHandler,
		statsHandler,
		inviteHandler,
		exportHandler,
		reminderHandler,
		notificationHandler,
		groupHandler,
		exchangeRateHandler,
		expenseRecordHandler,
	)
	if err != nil {
		logger.Fatal("failed to build router", zap.Error(err))
	}

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

// runMaintenance 执行一次性运维命令（需由运维人员手动触发）
func runMaintenance(ctx context.Context, command string, userService *service.UserService, fuelRecordService *service.FuelRecordService, logger *zap.Logger) error {
	switch command {
	case "purge-deleted-users":
		// 彻底清除旧版本"软删除"注销遗留的账号及其数据
		n, err := userService.PurgeSoftDeletedUsers(ctx)
		if err != nil {
			return err
		}
		logger.Info("purged soft-deleted users", zap.Int("count", n))
		return nil
	case "recompute-fuel-stats":
		// 按当前算法（满油到满油、按里程加权）重算所有加油记录的行驶距离、油耗与区间里程
		n, err := fuelRecordService.RecomputeAllVehicles(ctx)
		if err != nil {
			return err
		}
		logger.Info("recomputed fuel record derived fields", zap.Int("vehicles", n))
		return nil
	default:
		return fmt.Errorf("unknown command %q (available: purge-deleted-users, recompute-fuel-stats)", command)
	}
}

// runPeriodically 立即执行一次 fn，之后每隔 interval 执行一次，直到 ctx 取消
func runPeriodically(ctx context.Context, interval time.Duration, fn func(ctx context.Context)) {
	fn(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fn(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// initLogger 根据配置初始化 zap logger。
// 当 cfg.FilePath 非空时，日志同时写入文件（带 lumberjack 自动轮转）和 stderr；
// 否则仅输出到 stderr，行为与之前一致。
func initLogger(cfg config.LogConfig) (*zap.Logger, error) {
	// 1. 解析日志级别
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

	// 2. 选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderCfg := zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	// 3. 构建输出目标
	cores := []zapcore.Core{
		// 始终保留 stderr 输出（容器 / systemd 场景友好）
		zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), level),
	}

	if cfg.FilePath != "" {
		// 使用 lumberjack 实现日志文件自动轮转
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // MB
			MaxAge:     cfg.MaxAge,     // 天
			MaxBackups: cfg.MaxBackups, // 备份文件数
			Compress:   cfg.Compress,   // gzip 压缩
			LocalTime:  true,           // 使用本地时间命名备份文件
		}
		fileSyncer := zapcore.AddSync(fileWriter)
		cores = append(cores, zapcore.NewCore(encoder, fileSyncer, level))
	}

	// 4. 合并 core 并构建 logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
