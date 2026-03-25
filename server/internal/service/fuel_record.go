package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/pkg/convert"
	"gastrack/internal/repository"
)

// FuelRecordService 加油记录业务逻辑
type FuelRecordService struct {
	recordRepo  *repository.FuelRecordRepository
	vehicleRepo *repository.VehicleRepository
	logger      *zap.Logger
}

// NewFuelRecordService 创建 FuelRecordService 实例
func NewFuelRecordService(
	recordRepo *repository.FuelRecordRepository,
	vehicleRepo *repository.VehicleRepository,
	logger *zap.Logger,
) *FuelRecordService {
	return &FuelRecordService{
		recordRepo:  recordRepo,
		vehicleRepo: vehicleRepo,
		logger:      logger,
	}
}

// Create 创建加油记录
func (s *FuelRecordService) Create(ctx context.Context, userID, vehicleID uuid.UUID, req *dto.CreateFuelRecordRequest) (*dto.FuelRecordResponse, error) {
	// 验证车辆归属
	_, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, apperror.ErrInternal("verifying vehicle ownership", err)
	}

	// 解析加油日期
	refuelDate, err := time.Parse(time.RFC3339, req.RefuelDate)
	if err != nil {
		return nil, apperror.ErrBadRequest("record.invalid_date", "invalid date format, use ISO 8601")
	}

	// 默认单位
	fuelUnit := "L"
	if req.FuelUnit != "" {
		fuelUnit = req.FuelUnit
	}
	distUnit := "km"
	if req.DistanceUnit != "" {
		distUnit = req.DistanceUnit
	}

	record := &model.FuelRecord{
		VehicleID:    vehicleID,
		UserID:       userID,
		FuelAmount:   req.FuelAmount,
		FuelUnit:     fuelUnit,
		UnitPrice:    req.UnitPrice,
		TotalCost:    req.TotalCost,
		CurrencyCode: req.CurrencyCode,
		Odometer:     req.Odometer,
		DistanceUnit: distUnit,
		IsFullTank:   req.IsFullTank,
		FuelGrade:    req.FuelGrade,
		StationName:  req.StationName,
		StationLat:   req.StationLat,
		StationLng:   req.StationLng,
		Note:         req.Note,
		RefuelDate:   refuelDate,
	}

	// 计算行驶距离和油耗（基于上一条记录）
	s.calculateEfficiency(ctx, record)

	if err := s.recordRepo.Create(ctx, record); err != nil {
		return nil, apperror.ErrInternal("creating fuel record", err)
	}

	resp := fuelRecordToResponse(record)
	return &resp, nil
}

// List 获取车辆的加油记录列表（分页）
func (s *FuelRecordService) List(ctx context.Context, userID, vehicleID uuid.UUID, page, pageSize int) ([]dto.FuelRecordResponse, int64, error) {
	// 验证车辆归属
	_, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, 0, apperror.ErrInternal("verifying vehicle ownership", err)
	}

	records, total, err := s.recordRepo.ListByVehicle(ctx, vehicleID, page, pageSize)
	if err != nil {
		return nil, 0, apperror.ErrInternal("listing fuel records", err)
	}

	result := make([]dto.FuelRecordResponse, len(records))
	for i, r := range records {
		result[i] = fuelRecordToResponse(&r)
	}
	return result, total, nil
}

// GetByID 获取加油记录详情
func (s *FuelRecordService) GetByID(ctx context.Context, recordID, vehicleID uuid.UUID) (*dto.FuelRecordResponse, error) {
	record, err := s.recordRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("record.not_found", "fuel record not found")
		}
		return nil, apperror.ErrInternal("fetching fuel record", err)
	}

	resp := fuelRecordToResponse(record)
	return &resp, nil
}

// Update 更新加油记录
func (s *FuelRecordService) Update(ctx context.Context, recordID, vehicleID uuid.UUID, req *dto.UpdateFuelRecordRequest) (*dto.FuelRecordResponse, error) {
	record, err := s.recordRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("record.not_found", "fuel record not found")
		}
		return nil, apperror.ErrInternal("fetching fuel record", err)
	}

	// 部分更新
	if req.FuelAmount != nil {
		record.FuelAmount = *req.FuelAmount
	}
	if req.FuelUnit != nil {
		record.FuelUnit = *req.FuelUnit
	}
	if req.UnitPrice != nil {
		record.UnitPrice = *req.UnitPrice
	}
	if req.TotalCost != nil {
		record.TotalCost = *req.TotalCost
	}
	if req.CurrencyCode != nil {
		record.CurrencyCode = *req.CurrencyCode
	}
	if req.Odometer != nil {
		record.Odometer = *req.Odometer
	}
	if req.DistanceUnit != nil {
		record.DistanceUnit = *req.DistanceUnit
	}
	if req.IsFullTank != nil {
		record.IsFullTank = *req.IsFullTank
	}
	if req.FuelGrade != nil {
		record.FuelGrade = *req.FuelGrade
	}
	if req.StationName != nil {
		record.StationName = *req.StationName
	}
	if req.StationLat != nil {
		record.StationLat = *req.StationLat
	}
	if req.StationLng != nil {
		record.StationLng = *req.StationLng
	}
	if req.Note != nil {
		record.Note = *req.Note
	}
	if req.RefuelDate != nil {
		refuelDate, err := time.Parse(time.RFC3339, *req.RefuelDate)
		if err != nil {
			return nil, apperror.ErrBadRequest("record.invalid_date", "invalid date format")
		}
		record.RefuelDate = refuelDate
	}

	// 重新计算油耗
	s.calculateEfficiency(ctx, record)

	if err := s.recordRepo.Update(ctx, record); err != nil {
		return nil, apperror.ErrInternal("updating fuel record", err)
	}

	resp := fuelRecordToResponse(record)
	return &resp, nil
}

// Delete 删除加油记录
func (s *FuelRecordService) Delete(ctx context.Context, recordID, vehicleID uuid.UUID) error {
	if err := s.recordRepo.Delete(ctx, recordID, vehicleID); err != nil {
		return apperror.ErrInternal("deleting fuel record", err)
	}
	return nil
}

// calculateEfficiency 计算行驶距离和油耗
// 基于上一条满油记录的里程差来计算
func (s *FuelRecordService) calculateEfficiency(ctx context.Context, record *model.FuelRecord) {
	// 查找上一条记录
	prev, err := s.recordRepo.GetPreviousRecord(ctx, record.VehicleID, record.RefuelDate)
	if err != nil {
		// 第一条记录，无法计算
		return
	}

	// 计算行驶距离（先统一转为公制）
	_, currKm := convert.NormalizeToMetric(0, convert.VolumeUnit(record.FuelUnit), record.Odometer, convert.DistanceUnit(record.DistanceUnit))
	_, prevKm := convert.NormalizeToMetric(0, convert.VolumeUnit(prev.FuelUnit), prev.Odometer, convert.DistanceUnit(prev.DistanceUnit))

	tripDistance := currKm - prevKm
	if tripDistance <= 0 {
		return
	}
	record.TripDistance = tripDistance

	// 如果是加满油，计算油耗（L/100km 作为基准存储单位）
	if record.IsFullTank {
		liters, _ := convert.NormalizeToMetric(record.FuelAmount, convert.VolumeUnit(record.FuelUnit), 0, convert.UnitKm)
		record.FuelEfficiency = convert.CalcL100km(liters, tripDistance)
	}
}

// fuelRecordToResponse 将 model 转为 DTO
func fuelRecordToResponse(r *model.FuelRecord) dto.FuelRecordResponse {
	return dto.FuelRecordResponse{
		ID:             r.ID.String(),
		VehicleID:      r.VehicleID.String(),
		FuelAmount:     r.FuelAmount,
		FuelUnit:       r.FuelUnit,
		UnitPrice:      r.UnitPrice,
		TotalCost:      r.TotalCost,
		CurrencyCode:   r.CurrencyCode,
		Odometer:       r.Odometer,
		DistanceUnit:   r.DistanceUnit,
		IsFullTank:     r.IsFullTank,
		FuelGrade:      r.FuelGrade,
		StationName:    r.StationName,
		StationLat:     r.StationLat,
		StationLng:     r.StationLng,
		Note:           r.Note,
		ReceiptURL:     r.ReceiptURL,
		TripDistance:    r.TripDistance,
		FuelEfficiency: r.FuelEfficiency,
		RefuelDate:     r.RefuelDate,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
