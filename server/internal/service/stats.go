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
	"gastrack/internal/pkg/convert"
	"gastrack/internal/repository"
)

// StatsService 统计业务逻辑
type StatsService struct {
	recordRepo  repository.FuelRecordRepo
	vehicleRepo repository.VehicleRepo
	userRepo    repository.UserRepo
	groupRepo   repository.GroupRepo
	expenseRepo repository.ExpenseRecordRepo
	logger      *zap.Logger
}

// NewStatsService 创建 StatsService 实例
func NewStatsService(
	recordRepo repository.FuelRecordRepo,
	vehicleRepo repository.VehicleRepo,
	userRepo repository.UserRepo,
	groupRepo repository.GroupRepo,
	expenseRepo repository.ExpenseRecordRepo,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		recordRepo:  recordRepo,
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		expenseRepo: expenseRepo,
		logger:      logger,
	}
}

// verifyVehicleAccess 验证用户对车辆的访问权限（所有权或共享访问），返回车辆信息
func (s *StatsService) verifyVehicleAccess(ctx context.Context, vehicleID, userID uuid.UUID) (*model.Vehicle, error) {
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err == nil {
		return vehicle, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrInternal("fetching vehicle", err)
	}
	// 不是自己的车辆，检查是否为共享车辆
	if s.groupRepo != nil {
		shared, sharedErr := s.groupRepo.IsVehicleSharedToUser(ctx, vehicleID, userID)
		if sharedErr != nil {
			return nil, apperror.ErrInternal("checking shared vehicle access", sharedErr)
		}
		if shared {
			vehicle, err = s.vehicleRepo.GetByID(ctx, vehicleID)
			if err != nil {
				return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
			}
			return vehicle, nil
		}
	}
	return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
}

// GetVehicleStats 获取车辆统计
func (s *StatsService) GetVehicleStats(ctx context.Context, vehicleID, userID uuid.UUID) (*dto.VehicleStatsResponse, error) {
	// 验证车辆访问权限（所有权或共享）
	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
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

	// 获取按币种分组的费用明细
	costsByCurrency, err := s.recordRepo.GetCostByCurrency(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching costs by currency", err)
	}
	costMap := make(map[string]float64, len(costsByCurrency))
	for _, c := range costsByCurrency {
		costMap[c.CurrencyCode] = c.TotalCost
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
		CostsByCurrency: costMap,
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

	// 批量获取所有车辆的统计数据和按币种费用（2 次 SQL 替代 2N 次）
	vehicleIDs := make([]uuid.UUID, len(vehicles))
	for i, v := range vehicles {
		vehicleIDs[i] = v.ID
	}

	allStats, err := s.recordRepo.GetMultiVehicleStats(ctx, vehicleIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching vehicle stats", err)
	}

	allCosts, err := s.recordRepo.GetMultiVehicleCostByCurrency(ctx, vehicleIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching vehicle costs by currency", err)
	}

	var totalRecords int64
	var totalCost float64
	var totalFuel float64
	var totalDistance float64
	var totalExpenseCost float64
	overallCostsByCurrency := make(map[string]float64)
	overallExpenseCostsByCurrency := make(map[string]float64)
	vehicleStats := make([]dto.VehicleStatsResponse, 0, len(vehicles))

	isImperial := user.UnitSystem == "imperial"
	targetUnit := convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)

	for _, v := range vehicles {
		stats, ok := allStats[v.ID]
		if !ok {
			continue
		}

		// 从批量结果中获取该车辆的按币种分组费用
		vehicleCostMap := make(map[string]float64)
		if costs, costOk := allCosts[v.ID]; costOk {
			for _, c := range costs {
				vehicleCostMap[c.CurrencyCode] = c.TotalCost
				overallCostsByCurrency[c.CurrencyCode] += c.TotalCost
			}
		}

		// 获取该车辆的开销统计
		vehicleExpenseCostMap := make(map[string]float64)
		var vehicleExpenseTotal float64
		expenseCosts, err := s.expenseRepo.GetExpenseCostByCurrency(ctx, v.ID)
		if err == nil {
			for _, c := range expenseCosts {
				vehicleExpenseCostMap[c.CurrencyCode] = c.TotalAmount
				overallExpenseCostsByCurrency[c.CurrencyCode] += c.TotalAmount
				vehicleExpenseTotal += c.TotalAmount
			}
		}
		totalExpenseCost += vehicleExpenseTotal

		totalRecords += stats.TotalRecords
		totalCost += stats.TotalCost
		totalFuel += stats.TotalFuel
		totalDistance += stats.TotalDistance

		vFuel := stats.TotalFuel
		vDist := stats.TotalDistance
		if isImperial {
			vFuel = convert.ConvertVolume(vFuel, convert.UnitLiter, convert.UnitGallon)
			vDist = convert.ConvertDistance(vDist, convert.UnitKm, convert.UnitMile)
		}
		vehicleStats = append(vehicleStats, dto.VehicleStatsResponse{
			VehicleID:              v.ID.String(),
			VehicleName:            v.Name,
			TotalRecords:           stats.TotalRecords,
			TotalFuel:              vFuel,
			TotalCost:              stats.TotalCost,
			TotalDistance:           vDist,
			AvgEfficiency:          convert.ConvertFuelEfficiency(stats.AvgEfficiency, convert.UnitL100km, targetUnit),
			CurrencyCode:           user.CurrencyCode,
			FuelUnit:               user.FuelEfficiencyUnit,
			CostsByCurrency:        vehicleCostMap,
			TotalExpenseCost:       vehicleExpenseTotal,
			ExpenseCostsByCurrency: vehicleExpenseCostMap,
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
		TotalVehicles:          int64(len(vehicles)),
		TotalRecords:           totalRecords,
		TotalFuel:              overviewFuel,
		TotalCost:              totalCost,
		TotalDistance:           overviewDist,
		AvgConsumption:         avgConsumption,
		CurrencyCode:           user.CurrencyCode,
		CostsByCurrency:        overallCostsByCurrency,
		Vehicles:               vehicleStats,
		TotalExpenseCost:       totalExpenseCost,
		ExpenseCostsByCurrency: overallExpenseCostsByCurrency,
	}, nil
}

// GetEfficiencyTrend 获取油耗趋势
func (s *StatsService) GetEfficiencyTrend(ctx context.Context, vehicleID, userID uuid.UUID, limit int) (*dto.FuelEfficiencyTrendResponse, error) {
	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
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
	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
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

// GetExpensePeriodStats 获取按时段（月/年）聚合的开销统计数据 + 往年同比
func (s *StatsService) GetExpensePeriodStats(ctx context.Context, vehicleID, userID uuid.UUID, period string, year int) (*dto.ExpensePeriodStatsResponse, error) {
	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	convertItems := func(raw []repository.ExpensePeriodStatsResult) []dto.ExpensePeriodStatsItem {
		items := make([]dto.ExpensePeriodStatsItem, len(raw))
		for i, r := range raw {
			items[i] = dto.ExpensePeriodStatsItem{
				Period:       r.Period,
				TotalRecords: r.TotalRecords,
				TotalAmount:  r.TotalAmount,
			}
		}
		return items
	}

	var items, prevItems []dto.ExpensePeriodStatsItem

	if period == "month" {
		raw, err := s.expenseRepo.GetExpenseStatsByMonth(ctx, vehicleID, year)
		if err != nil {
			return nil, apperror.ErrInternal("fetching monthly expense stats", err)
		}
		items = convertItems(raw)

		rawPrev, err := s.expenseRepo.GetExpenseStatsByMonth(ctx, vehicleID, year-1)
		if err != nil {
			return nil, apperror.ErrInternal("fetching prev year monthly expense stats", err)
		}
		prevItems = convertItems(rawPrev)
	} else {
		raw, err := s.expenseRepo.GetExpenseStatsByYear(ctx, vehicleID)
		if err != nil {
			return nil, apperror.ErrInternal("fetching yearly expense stats", err)
		}
		items = convertItems(raw)
		prevItems = []dto.ExpensePeriodStatsItem{}
	}

	// 按币种分组统计
	costsByCurrency, err := s.expenseRepo.GetExpenseCostByCurrency(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching expense costs by currency", err)
	}
	costMap := make(map[string]float64, len(costsByCurrency))
	for _, c := range costsByCurrency {
		costMap[c.CurrencyCode] = c.TotalAmount
	}

	return &dto.ExpensePeriodStatsResponse{
		VehicleID:       vehicleID.String(),
		VehicleName:     vehicle.Name,
		Period:          period,
		Year:            year,
		CurrencyCode:    user.CurrencyCode,
		CostsByCurrency: costMap,
		Items:           items,
		PrevItems:       prevItems,
	}, nil
}
