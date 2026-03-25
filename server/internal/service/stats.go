package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/pkg/convert"
	"gastrack/internal/repository"
)

// StatsService 统计业务逻辑
type StatsService struct {
	recordRepo  *repository.FuelRecordRepository
	vehicleRepo *repository.VehicleRepository
	userRepo    *repository.UserRepository
	logger      *zap.Logger
}

// NewStatsService 创建 StatsService 实例
func NewStatsService(
	recordRepo *repository.FuelRecordRepository,
	vehicleRepo *repository.VehicleRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		recordRepo:  recordRepo,
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// GetVehicleStats 获取车辆统计
func (s *StatsService) GetVehicleStats(ctx context.Context, vehicleID, userID uuid.UUID) (*dto.VehicleStatsResponse, error) {
	// 验证车辆归属
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, apperror.ErrInternal("fetching vehicle", err)
	}

	// 获取用户偏好（用于单位换算）
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	stats, err := s.recordRepo.GetVehicleStats(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching vehicle stats", err)
	}

	// 根据用户偏好换算油耗单位
	targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
	avgEff := convert.ConvertFuelEfficiency(stats.AvgEfficiency, convert.UnitL100km, targetUnit)
	bestEff := convert.ConvertFuelEfficiency(stats.BestEfficiency, convert.UnitL100km, targetUnit)
	worstEff := convert.ConvertFuelEfficiency(stats.WorstEfficiency, convert.UnitL100km, targetUnit)

	// 计算每公里费用
	var avgCostPerKm, avgCostPerFill float64
	if stats.TotalDistance > 0 {
		avgCostPerKm = stats.TotalCost / stats.TotalDistance
	}
	if stats.TotalRecords > 0 {
		avgCostPerFill = stats.TotalCost / float64(stats.TotalRecords)
	}

	return &dto.VehicleStatsResponse{
		VehicleID:       vehicleID.String(),
		VehicleName:     vehicle.Name,
		TotalRecords:    stats.TotalRecords,
		TotalFuel:       stats.TotalFuel,
		TotalCost:       stats.TotalCost,
		TotalDistance:   stats.TotalDistance,
		AvgEfficiency:   avgEff,
		BestEfficiency:  bestEff,
		WorstEfficiency: worstEff,
		AvgCostPerKm:    avgCostPerKm,
		AvgCostPerFill:  avgCostPerFill,
		CurrencyCode:    user.CurrencyCode,
		FuelUnit:        user.FuelEfficiencyUnit,
	}, nil
}

// GetOverview 获取全局统计总览
func (s *StatsService) GetOverview(ctx context.Context, userID uuid.UUID) (*dto.OverviewStatsResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	vehicles, err := s.vehicleRepo.ListByUser(ctx, userID, false)
	if err != nil {
		return nil, apperror.ErrInternal("listing vehicles", err)
	}

	var totalRecords int64
	var totalCost float64
	vehicleStats := make([]dto.VehicleStatsResponse, 0, len(vehicles))

	for _, v := range vehicles {
		stats, err := s.recordRepo.GetVehicleStats(ctx, v.ID)
		if err != nil {
			continue
		}

		totalRecords += stats.TotalRecords
		totalCost += stats.TotalCost

		targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
		vehicleStats = append(vehicleStats, dto.VehicleStatsResponse{
			VehicleID:     v.ID.String(),
			VehicleName:   v.Name,
			TotalRecords:  stats.TotalRecords,
			TotalFuel:     stats.TotalFuel,
			TotalCost:     stats.TotalCost,
			TotalDistance:  stats.TotalDistance,
			AvgEfficiency:  convert.ConvertFuelEfficiency(stats.AvgEfficiency, convert.UnitL100km, targetUnit),
			CurrencyCode:  user.CurrencyCode,
			FuelUnit:      user.FuelEfficiencyUnit,
		})
	}

	return &dto.OverviewStatsResponse{
		TotalVehicles: int64(len(vehicles)),
		TotalRecords:  totalRecords,
		TotalCost:     totalCost,
		CurrencyCode:  user.CurrencyCode,
		Vehicles:      vehicleStats,
	}, nil
}

// GetEfficiencyTrend 获取油耗趋势
func (s *StatsService) GetEfficiencyTrend(ctx context.Context, vehicleID, userID uuid.UUID, limit int) (*dto.FuelEfficiencyTrendResponse, error) {
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
		}
		return nil, apperror.ErrInternal("fetching vehicle", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	records, err := s.recordRepo.GetEfficiencyTrend(ctx, vehicleID, limit)
	if err != nil {
		return nil, apperror.ErrInternal("fetching efficiency trend", err)
	}

	targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
	items := make([]dto.FuelEfficiencyTrendItem, len(records))
	for i, r := range records {
		items[i] = dto.FuelEfficiencyTrendItem{
			Date:           r.RefuelDate.Format("2006-01-02"),
			FuelEfficiency: convert.ConvertFuelEfficiency(r.FuelEfficiency, convert.UnitL100km, targetUnit),
			TripDistance:   r.TripDistance,
		}
	}

	return &dto.FuelEfficiencyTrendResponse{
		VehicleID:      vehicleID.String(),
		VehicleName:    vehicle.Name,
		EfficiencyUnit: user.FuelEfficiencyUnit,
		Items:          items,
	}, nil
}
