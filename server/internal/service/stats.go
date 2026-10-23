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
	"gastrack/internal/repository"
)

// StatsService 统计业务逻辑
type StatsService struct {
	recordRepo  repository.FuelRecordRepo
	vehicleRepo repository.VehicleRepo
	userRepo    repository.UserRepo
	groupRepo   repository.GroupRepo
	expenseRepo repository.ExpenseRecordRepo
	rates       RateProvider
	logger      *zap.Logger
}

// NewStatsService 创建 StatsService 实例
func NewStatsService(
	recordRepo repository.FuelRecordRepo,
	vehicleRepo repository.VehicleRepo,
	userRepo repository.UserRepo,
	groupRepo repository.GroupRepo,
	expenseRepo repository.ExpenseRecordRepo,
	rates RateProvider,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		recordRepo:  recordRepo,
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		expenseRepo: expenseRepo,
		rates:       rates,
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

// buildVehicleStats 将车辆的加油合计转为展示用统计（单位换算、费用换汇）
// 电动车只统计充电记录，燃油/混动车只统计燃油记录，两类能源不混算
func buildVehicleStats(v *model.Vehicle, t *fuelTotals, prefs unitPrefs, conv *moneyConverter) dto.VehicleStatsResponse {
	kind := vehicleEnergyKind(v)
	fuel, volumeUnit := prefs.amount(t.amount(kind), kind)
	avgEff, effUnit := prefs.efficiency(t.avgEfficiency(kind), kind)
	bestBase, worstBase := t.bestWorst(kind)
	best, _ := prefs.efficiency(bestBase, kind)
	worst, _ := prefs.efficiency(worstBase, kind)
	distance := prefs.distance(t.Distance)

	totalCost := conv.sum(t.costs)
	var avgCostPerKm, avgCostPerFill float64
	if distance > 0 {
		avgCostPerKm = totalCost / distance
	}
	if t.Records > 0 {
		avgCostPerFill = totalCost / float64(t.Records)
	}

	energyType := "fuel"
	if kind == energyElectric {
		energyType = "electric"
	}

	return dto.VehicleStatsResponse{
		VehicleID:       v.ID.String(),
		VehicleName:     v.Name,
		EnergyType:      energyType,
		TotalRecords:    t.Records,
		TotalFuel:       fuel,
		VolumeUnit:      volumeUnit,
		TotalCost:       totalCost,
		TotalDistance:   distance,
		AvgEfficiency:   avgEff,
		BestEfficiency:  best,
		WorstEfficiency: worst,
		AvgCostPerKm:    avgCostPerKm,
		AvgCostPerFill:  avgCostPerFill,
		CurrencyCode:    conv.base,
		FuelUnit:        effUnit,
		CostsByCurrency: t.costs,
	}
}

// GetVehicleStats 获取车辆统计
func (s *StatsService) GetVehicleStats(ctx context.Context, vehicleID, userID uuid.UUID) (*dto.VehicleStatsResponse, error) {
	// 验证车辆访问权限（所有权或共享）
	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	// 获取用户偏好（用于单位换算与换汇）
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	rows, err := s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: []uuid.UUID{vehicleID}})
	if err != nil {
		return nil, apperror.ErrInternal("fetching vehicle stats", err)
	}
	totals := newFuelTotals()
	for _, row := range rows {
		totals.add(row)
	}

	conv := newMoneyConverter(s.rates, user.CurrencyCode)
	resp := buildVehicleStats(vehicle, totals, prefsOf(user), conv)
	resp.UnconvertedCurrencies = conv.unconvertedCurrencies()
	return &resp, nil
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

	// 1 次 SQL 获取所有车辆按币种分组的聚合
	vehicleIDs := make([]uuid.UUID, len(vehicles))
	for i, v := range vehicles {
		vehicleIDs[i] = v.ID
	}
	rows, err := s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: vehicleIDs})
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching vehicle stats", err)
	}
	byVehicle := totalsByVehicle(rows)

	prefs := prefsOf(user)
	conv := newMoneyConverter(s.rates, user.CurrencyCode)
	overall := newFuelTotals()
	var totalExpenseCost float64
	overallExpenseCostsByCurrency := make(map[string]float64)
	vehicleStats := make([]dto.VehicleStatsResponse, 0, len(vehicles))

	for i := range vehicles {
		v := &vehicles[i]
		totals, ok := byVehicle[v.ID.String()]
		if !ok {
			totals = newFuelTotals()
		}

		// 获取该车辆的开销统计
		vehicleExpenseCostMap := make(map[string]float64)
		expenseCosts, err := s.expenseRepo.GetExpenseCostByCurrency(ctx, v.ID)
		if err == nil {
			for _, c := range expenseCosts {
				vehicleExpenseCostMap[c.CurrencyCode] = c.TotalAmount
				overallExpenseCostsByCurrency[c.CurrencyCode] += c.TotalAmount
			}
		}

		// 既无加油记录也无开销的车辆不出现在总览中
		if totals.Records == 0 && len(vehicleExpenseCostMap) == 0 {
			continue
		}
		overall.merge(totals)

		vs := buildVehicleStats(v, totals, prefs, conv)
		vs.TotalExpenseCost = conv.sum(vehicleExpenseCostMap)
		vs.ExpenseCostsByCurrency = vehicleExpenseCostMap
		totalExpenseCost += vs.TotalExpenseCost
		vehicleStats = append(vehicleStats, vs)
	}

	// 燃油与电分别汇总：总量、按里程加权的平均油耗/电耗
	totalFuel, volumeUnit := prefs.amount(overall.Liters, energyLiquid)
	avgConsumption, fuelUnit := prefs.efficiency(overall.LiquidEfficiency(), energyLiquid)
	avgEnergy, energyUnit := prefs.efficiency(overall.ElectricEfficiency(), energyElectric)

	return &dto.OverviewStatsResponse{
		TotalVehicles:          int64(len(vehicles)),
		TotalRecords:           overall.Records,
		TotalFuel:              totalFuel,
		VolumeUnit:             volumeUnit,
		TotalEnergy:            overall.Kwh,
		TotalCost:              conv.sum(overall.costs),
		TotalDistance:          prefs.distance(overall.Distance),
		AvgConsumption:         avgConsumption,
		FuelUnit:               fuelUnit,
		AvgEnergyConsumption:   avgEnergy,
		EnergyUnit:             energyUnit,
		CurrencyCode:           conv.base,
		CostsByCurrency:        overall.costs,
		Vehicles:               vehicleStats,
		TotalExpenseCost:       totalExpenseCost,
		ExpenseCostsByCurrency: overallExpenseCostsByCurrency,
		UnconvertedCurrencies:  conv.unconvertedCurrencies(),
	}, nil
}

// GetEfficiencyTrend 获取油耗/电耗趋势（只包含与车辆能源类别一致的记录）
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

	prefs := prefsOf(user)
	kind := vehicleEnergyKind(vehicle)
	items := make([]dto.FuelEfficiencyTrendItem, 0, len(records))
	for _, r := range records {
		if recordEnergyKind(r.FuelUnit) != kind {
			continue
		}
		eff, _ := prefs.efficiency(r.FuelEfficiency, kind)
		items = append(items, dto.FuelEfficiencyTrendItem{
			Date:           r.RefuelDate.Format("2006-01-02"),
			FuelEfficiency: eff,
			TripDistance:   prefs.distance(r.TripDistance),
		})
	}

	return &dto.FuelEfficiencyTrendResponse{
		VehicleID:      vehicleID.String(),
		VehicleName:    vehicle.Name,
		EfficiencyUnit: prefs.efficiencyUnit(kind),
		Items:          items,
	}, nil
}

// yearRange 返回某年的 [1 月 1 日, 次年 1 月 1 日)
func yearRange(year int) (*time.Time, *time.Time) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	return &start, &end
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

	prefs := prefsOf(user)
	conv := newMoneyConverter(s.rates, user.CurrencyCode)
	kind := vehicleEnergyKind(vehicle)

	load := func(q repository.FuelAggregateQuery) ([]dto.PeriodStatsItem, error) {
		q.VehicleIDs = []uuid.UUID{vehicleID}
		rows, err := s.recordRepo.Aggregate(ctx, q)
		if err != nil {
			return nil, err
		}
		byPeriod, order := totalsByPeriod(rows)
		items := make([]dto.PeriodStatsItem, 0, len(order))
		for _, p := range order {
			t := byPeriod[p]
			fuel, _ := prefs.amount(t.amount(kind), kind)
			eff, _ := prefs.efficiency(t.avgEfficiency(kind), kind)
			items = append(items, dto.PeriodStatsItem{
				Period:        p,
				TotalRecords:  int(t.Records),
				TotalFuel:     fuel,
				TotalCost:     conv.sum(t.costs),
				TotalDistance: prefs.distance(t.Distance),
				AvgEfficiency: eff,
			})
		}
		return items, nil
	}

	var items, prevItems []dto.PeriodStatsItem
	if period == "month" {
		// 按月聚合：当年 + 上一年同比
		start, end := yearRange(year)
		if items, err = load(repository.FuelAggregateQuery{PeriodFormat: repository.PeriodMonth, Start: start, End: end}); err != nil {
			return nil, apperror.ErrInternal("fetching monthly stats", err)
		}
		start, end = yearRange(year - 1)
		if prevItems, err = load(repository.FuelAggregateQuery{PeriodFormat: repository.PeriodMonth, Start: start, End: end}); err != nil {
			return nil, apperror.ErrInternal("fetching prev year monthly stats", err)
		}
	} else {
		// 按年聚合：全部年份；按年模式没有 "上一周期"，prevItems 为空
		if items, err = load(repository.FuelAggregateQuery{PeriodFormat: repository.PeriodYear}); err != nil {
			return nil, apperror.ErrInternal("fetching yearly stats", err)
		}
		prevItems = []dto.PeriodStatsItem{}
	}

	_, volumeUnit := prefs.amount(0, kind)
	return &dto.PeriodStatsResponse{
		VehicleID:             vehicleID.String(),
		VehicleName:           vehicle.Name,
		Period:                period,
		Year:                  year,
		CurrencyCode:          conv.base,
		FuelUnit:              prefs.efficiencyUnit(kind),
		VolumeUnit:            volumeUnit,
		UnconvertedCurrencies: conv.unconvertedCurrencies(),
		Items:                 items,
		PrevItems:             prevItems,
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
	conv := newMoneyConverter(s.rates, user.CurrencyCode)

	// 同一时段的不同币种换算为用户币种后合并
	convertItems := func(raw []repository.ExpensePeriodStatsResult) []dto.ExpensePeriodStatsItem {
		index := make(map[string]int)
		items := make([]dto.ExpensePeriodStatsItem, 0, len(raw))
		for _, r := range raw {
			i, ok := index[r.Period]
			if !ok {
				i = len(items)
				index[r.Period] = i
				items = append(items, dto.ExpensePeriodStatsItem{Period: r.Period})
			}
			items[i].TotalRecords += r.TotalRecords
			if amount, ok := conv.convert(r.TotalAmount, r.CurrencyCode); ok {
				items[i].TotalAmount += amount
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
		VehicleID:             vehicleID.String(),
		VehicleName:           vehicle.Name,
		Period:                period,
		Year:                  year,
		CurrencyCode:          conv.base,
		CostsByCurrency:       costMap,
		UnconvertedCurrencies: conv.unconvertedCurrencies(),
		Items:                 items,
		PrevItems:             prevItems,
	}, nil
}
