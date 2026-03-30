package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// UserExportData 用户导出数据的聚合结构
type UserExportData struct {
	User     *model.User
	Vehicles []model.Vehicle
	Records  []model.FuelRecord
}

// ExportService 数据导出业务逻辑
type ExportService struct {
	userRepo       *repository.UserRepository
	vehicleRepo    *repository.VehicleRepository
	fuelRecordRepo *repository.FuelRecordRepository
	logger         *zap.Logger
}

// NewExportService 创建 ExportService 实例
func NewExportService(
	userRepo *repository.UserRepository,
	vehicleRepo *repository.VehicleRepository,
	fuelRecordRepo *repository.FuelRecordRepository,
	logger *zap.Logger,
) *ExportService {
	return &ExportService{
		userRepo:       userRepo,
		vehicleRepo:    vehicleRepo,
		fuelRecordRepo: fuelRecordRepo,
		logger:         logger,
	}
}

// GatherUserData 收集当前用户的所有数据用于导出
func (s *ExportService) GatherUserData(ctx context.Context, userID uuid.UUID) (*UserExportData, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user for export", err)
	}

	// 2. 获取所有车辆（含归档）
	vehicles, err := s.vehicleRepo.ListByUser(ctx, userID, true)
	if err != nil {
		return nil, apperror.ErrInternal("fetching vehicles for export", err)
	}

	// 3. 获取所有加油记录
	records, err := s.fuelRecordRepo.ListAllByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching fuel records for export", err)
	}

	return &UserExportData{
		User:     user,
		Vehicles: vehicles,
		Records:  records,
	}, nil
}
