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

	// 根据用户偏好换算容量和距离单位
	isImperial := user.UnitSystem == "imperial"
	totalFuel := stats.TotalFuel
	totalDist := stats.TotalDistance
	if isImperial {
		totalFuel = convert.ConvertVolume(totalFuel, convert.UnitLiter, convert.UnitGallon)
		totalDist = convert.ConvertDistance(totalDist, convert.UnitKm, convert.UnitMile)
	}

	// 计算每公里/英里费用
	var avgCostPerKm, avgCostPerFill float64
	if totalDist > 0 {
		avgCostPerKm = stats.TotalCost / totalDist
	}
	if stats.TotalRecords > 0 {
		avgCostPerFill = stats.TotalCost / float64(stats.TotalRecords)
	}

	return &dto.VehicleStatsResponse{
		VehicleID:       vehicleID.String(),
		VehicleName:     vehicle.Name,
		TotalRecords:    stats.TotalRecords,
		TotalFuel:       totalFuel,
		TotalCost:       stats.TotalCost,
		TotalDistance:   totalDist,
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
	var totalFuel float64
	var totalDistance float64
	vehicleStats := make([]dto.VehicleStatsResponse, 0, len(vehicles))

	isImperial := user.UnitSystem == "imperial"

	for _, v := range vehicles {
		stats, err := s.recordRepo.GetVehicleStats(ctx, v.ID)
		if err != nil {
			continue
		}

		totalRecords += stats.TotalRecords
		totalCost += stats.TotalCost
		totalFuel += stats.TotalFuel
		totalDistance += stats.TotalDistance

		targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
		vFuel := stats.TotalFuel
		vDist := stats.TotalDistance
		if isImperial {
			vFuel = convert.ConvertVolume(vFuel, convert.UnitLiter, convert.UnitGallon)
			vDist = convert.ConvertDistance(vDist, convert.UnitKm, convert.UnitMile)
		}
		vehicleStats = append(vehicleStats, dto.VehicleStatsResponse{
			VehicleID:     v.ID.String(),
			VehicleName:   v.Name,
			TotalRecords:  stats.TotalRecords,
			TotalFuel:     vFuel,
			TotalCost:     stats.TotalCost,
			TotalDistance:  vDist,
			AvgEfficiency:  convert.ConvertFuelEfficiency(stats.AvgEfficiency, convert.UnitL100km, targetUnit),
			CurrencyCode:  user.CurrencyCode,
			FuelUnit:      user.FuelEfficiencyUnit,
		})
	}

	// 计算平均油耗 (L/100km 基准，再转为用户偏好单位)
	var avgConsumption float64
	if totalDistance > 0 {
		avgConsumption = totalFuel / totalDistance * 100
		targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
		avgConsumption = convert.ConvertFuelEfficiency(avgConsumption, convert.UnitL100km, targetUnit)
	}

	// 转换总计的容量和距离
	overviewFuel := totalFuel
	overviewDist := totalDistance
	if isImperial {
		overviewFuel = convert.ConvertVolume(overviewFuel, convert.UnitLiter, convert.UnitGallon)
		overviewDist = convert.ConvertDistance(overviewDist, convert.UnitKm, convert.UnitMile)
	}

	return &dto.OverviewStatsResponse{
		TotalVehicles:  int64(len(vehicles)),
		TotalRecords:   totalRecords,
		TotalFuel:      overviewFuel,
		TotalCost:      totalCost,
		TotalDistance:  overviewDist,
		AvgConsumption: avgConsumption,
		CurrencyCode:   user.CurrencyCode,
		Vehicles:       vehicleStats,
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
	isImperial := user.UnitSystem == "imperial"
	items := make([]dto.FuelEfficiencyTrendItem, len(records))
	for i, r := range records {
		tripDist := r.TripDistance
		if isImperial {
			tripDist = convert.ConvertDistance(tripDist, convert.UnitKm, convert.UnitMile)
		}
		items[i] = dto.FuelEfficiencyTrendItem{
			Date:           r.RefuelDate.Format("2006-01-02"),
			FuelEfficiency: convert.ConvertFuelEfficiency(r.FuelEfficiency, convert.UnitL100km, targetUnit),
			TripDistance:   tripDist,
		}
	}

	return &dto.FuelEfficiencyTrendResponse{
		VehicleID:      vehicleID.String(),
		VehicleName:    vehicle.Name,
		EfficiencyUnit: user.FuelEfficiencyUnit,
		Items:          items,
	}, nil
}

// GetPeriodStats 获取按时段（月/年）聚合的统计数据 + 往年同比
func (s *StatsService) GetPeriodStats(ctx context.Context, vehicleID, userID uuid.UUID, period string, year int) (*dto.PeriodStatsResponse, error) {
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

	targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
	isImperial := user.UnitSystem == "imperial"

	convertItems := func(raw []repository.PeriodStatsResult) []dto.PeriodStatsItem {
		items := make([]dto.PeriodStatsItem, len(raw))
		for i, r := range raw {
			totalFuel := r.TotalFuel
			totalDist := r.TotalDistance
			if isImperial {
				totalFuel = convert.ConvertVolume(totalFuel, convert.UnitLiter, convert.UnitGallon)
				totalDist = convert.ConvertDistance(totalDist, convert.UnitKm, convert.UnitMile)
			}
			items[i] = dto.PeriodStatsItem{
				Period:        r.Period,
				TotalRecords:  r.TotalRecords,
				TotalFuel:     totalFuel,
				TotalCost:     r.TotalCost,
				TotalDistance:  totalDist,
				AvgEfficiency: convert.ConvertFuelEfficiency(r.AvgEfficiency, convert.UnitL100km, targetUnit),
			}
		}
		return items
	}

	var items, prevItems []dto.PeriodStatsItem

	if period == "month" {
		// 按月聚合：当年 + 上一年同比
		raw, err := s.recordRepo.GetStatsByMonth(ctx, vehicleID, year)
		if err != nil {
			return nil, apperror.ErrInternal("fetching monthly stats", err)
		}
		items = convertItems(raw)

		rawPrev, err := s.recordRepo.GetStatsByMonth(ctx, vehicleID, year-1)
		if err != nil {
			return nil, apperror.ErrInternal("fetching prev year monthly stats", err)
		}
		prevItems = convertItems(rawPrev)
	} else {
		// 按年聚合：全部年份
		raw, err := s.recordRepo.GetStatsByYear(ctx, vehicleID)
		if err != nil {
			return nil, apperror.ErrInternal("fetching yearly stats", err)
		}
		items = convertItems(raw)
		// 按年模式没有 "上一周期"，prevItems 为空
		prevItems = []dto.PeriodStatsItem{}
	}

	return &dto.PeriodStatsResponse{
		VehicleID:    vehicleID.String(),
		VehicleName:  vehicle.Name,
		Period:       period,
		Year:         year,
		CurrencyCode: user.CurrencyCode,
		FuelUnit:     user.FuelEfficiencyUnit,
		Items:        items,
		PrevItems:    prevItems,
	}, nil
}
