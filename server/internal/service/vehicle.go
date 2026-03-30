package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// VehicleService 车辆业务逻辑
type VehicleService struct {
	vehicleRepo *repository.VehicleRepository
	logger      *zap.Logger
}

// NewVehicleService 创建 VehicleService 实例
func NewVehicleService(vehicleRepo *repository.VehicleRepository, logger *zap.Logger) *VehicleService {
	return &VehicleService{vehicleRepo: vehicleRepo, logger: logger}
}

// Create 创建车辆
func (s *VehicleService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateVehicleRequest) (*dto.VehicleResponse, error) {
	vehicle := &model.Vehicle{
		UserID:          userID,
		Name:            req.Name,
		VehicleType:     model.VehicleType(req.VehicleType),
		Brand:           req.Brand,
		Model:           req.Model,
		Year:            req.Year,
		FuelType:        model.FuelType(req.FuelType),
		FuelGrade:       req.FuelGrade,
		TankCapacity:    req.TankCapacity,
		BatteryCapacity: req.BatteryCapacity,
		EngineCC:        req.EngineCC,
		LicensePlate:    req.LicensePlate,
		IsDefault:       req.IsDefault,
	}

	if req.IsDefault {
		// 如果设为默认车辆，用事务保证 ClearDefault + Create 原子性
		err := s.vehicleRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.vehicleRepo.ClearDefaultTx(ctx, tx, userID); err != nil {
				return err
			}
			return s.vehicleRepo.CreateTx(ctx, tx, vehicle)
		})
		if err != nil {
			return nil, apperror.ErrInternal("creating default vehicle", err)
		}
	} else {
		if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
			return nil, apperror.ErrInternal("creating vehicle", err)
		}
	}

	resp := vehicleToResponse(vehicle)
	return &resp, nil
}

// List 获取用户的车辆列表
func (s *VehicleService) List(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]dto.VehicleResponse, error) {
	vehicles, err := s.vehicleRepo.ListByUser(ctx, userID, includeArchived)
	if err != nil {
		return nil, apperror.ErrInternal("listing vehicles", err)
	}

	result := make([]dto.VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		result[i] = vehicleToResponse(&v)
	}
	return result, nil
}

// GetByID 获取车辆详情
func (s *VehicleService) GetByID(ctx context.Context, vehicleID, userID uuid.UUID) (*dto.VehicleResponse, error) {
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, apperror.ErrInternal("fetching vehicle", err)
	}

	resp := vehicleToResponse(vehicle)
	return &resp, nil
}

// Update 更新车辆（部分更新）
func (s *VehicleService) Update(ctx context.Context, vehicleID, userID uuid.UUID, req *dto.UpdateVehicleRequest) (*dto.VehicleResponse, error) {
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, apperror.ErrInternal("fetching vehicle", err)
	}

	// 应用部分更新
	if req.Name != nil {
		vehicle.Name = *req.Name
	}
	if req.VehicleType != nil {
		vehicle.VehicleType = model.VehicleType(*req.VehicleType)
	}
	if req.Brand != nil {
		vehicle.Brand = *req.Brand
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.FuelType != nil {
		vehicle.FuelType = model.FuelType(*req.FuelType)
	}
	if req.FuelGrade != nil {
		vehicle.FuelGrade = *req.FuelGrade
	}
	if req.TankCapacity != nil {
		vehicle.TankCapacity = *req.TankCapacity
	}
	if req.BatteryCapacity != nil {
		vehicle.BatteryCapacity = *req.BatteryCapacity
	}
	if req.EngineCC != nil {
		vehicle.EngineCC = *req.EngineCC
	}
	if req.LicensePlate != nil {
		vehicle.LicensePlate = *req.LicensePlate
	}
	if req.IsArchived != nil {
		vehicle.IsArchived = *req.IsArchived
	}

	// 设为默认
	if req.IsDefault != nil && *req.IsDefault {
		// 用事务保证 ClearDefault + Update 原子性
		err := s.vehicleRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.vehicleRepo.ClearDefaultTx(ctx, tx, userID); err != nil {
				return err
			}
			vehicle.IsDefault = true
			return s.vehicleRepo.UpdateTx(ctx, tx, vehicle)
		})
		if err != nil {
			return nil, apperror.ErrInternal("updating default vehicle", err)
		}
	} else {
		if req.IsDefault != nil && !*req.IsDefault {
			vehicle.IsDefault = false
		}
		if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
			return nil, apperror.ErrInternal("updating vehicle", err)
		}
	}

	resp := vehicleToResponse(vehicle)
	return &resp, nil
}

// Delete 删除车辆
func (s *VehicleService) Delete(ctx context.Context, vehicleID, userID uuid.UUID) error {
	// 验证车辆归属
	_, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return apperror.ErrInternal("fetching vehicle", err)
	}

	if err := s.vehicleRepo.Delete(ctx, vehicleID, userID); err != nil {
		return apperror.ErrInternal("deleting vehicle", err)
	}
	return nil
}

// vehicleToResponse 将 model 转为 DTO
func vehicleToResponse(v *model.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID:              v.ID.String(),
		Name:            v.Name,
		VehicleType:     string(v.VehicleType),
		Brand:           v.Brand,
		Model:           v.Model,
		Year:            v.Year,
		FuelType:        string(v.FuelType),
		FuelGrade:       v.FuelGrade,
		TankCapacity:    v.TankCapacity,
		BatteryCapacity: v.BatteryCapacity,
		EngineCC:        v.EngineCC,
		LicensePlate:    v.LicensePlate,
		PhotoURL:        v.PhotoURL,
		IsDefault:       v.IsDefault,
		IsArchived:      v.IsArchived,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}
