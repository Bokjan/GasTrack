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
	// GORM 日志级别
	gormLogLevel := logger.Info
	if zapLogger.Core().Enabled(zap.DebugLevel) {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
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
	)
}
