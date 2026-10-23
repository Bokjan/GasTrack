// Package database 负责数据库连接初始化和自动迁移。
package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gastrack/internal/config"
	"gastrack/internal/model"
)

// New 创建数据库连接并执行自动迁移
func New(cfg *config.DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// GORM 日志：debug 级别输出全部 SQL，其余仅输出慢查询与错误；统一写入 zap
	gormLogLevel := logger.Warn
	if zapLogger.Core().Enabled(zap.DebugLevel) {
		gormLogLevel = logger.Info
	}
	gormLogger := logger.New(
		zap.NewStdLog(zapLogger.Named("gorm")),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true, // 查不到记录属于正常业务分支
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
		// 将数据库方言错误翻译为 gorm 通用错误（如唯一约束冲突 → gorm.ErrDuplicatedKey）
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto-migrating: %w", err)
	}
	if err := widenFuelEfficiency(db); err != nil {
		return nil, fmt.Errorf("widening fuel_records.fuel_efficiency: %w", err)
	}
	// 邮箱按小写匹配（登录、查重），需要表达式索引支撑；仅新增索引，不改动数据
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email))").Error; err != nil {
		return nil, fmt.Errorf("creating idx_users_email_lower: %w", err)
	}

	zapLogger.Info("database connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
	)

	return db, nil
}

// autoMigrate 自动迁移所有模型
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Vehicle{},
		&model.FuelRecord{},
		&model.RefreshToken{},
		&model.InviteCode{},
		&model.Reminder{},
		&model.Notification{},
		&model.Group{},
		&model.GroupMember{},
		&model.SharedVehicle{},
		&model.ExpenseRecord{},
	)
}

// widenFuelEfficiency 将 fuel_records.fuel_efficiency 从旧版 decimal(6,2) 加宽到 decimal(10,2)。
// 旧精度在里程差很小时会数值溢出导致写入失败；AutoMigrate 不会修改已有 decimal 列的精度，
// 因此显式执行。仅加宽、不改动数据，且只在精度不足时执行一次。
func widenFuelEfficiency(db *gorm.DB) error {
	var precision int
	if err := db.Raw(`
		SELECT COALESCE(numeric_precision, 0) FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'fuel_records' AND column_name = 'fuel_efficiency'
	`).Scan(&precision).Error; err != nil {
		return err
	}
	if precision == 0 || precision >= 10 {
		return nil
	}
	return db.Exec("ALTER TABLE fuel_records ALTER COLUMN fuel_efficiency TYPE decimal(10,2)").Error
}
