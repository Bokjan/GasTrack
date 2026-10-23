package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
	mockrepo "gastrack/internal/repository/mock"
)

// fakeRates 固定汇率：1 CNY = 0.14 USD = 20 JPY
type fakeRates struct{ fail bool }

func (f fakeRates) GetRates(base string) (*dto.ExchangeRateResponse, error) {
	if f.fail || base != "CNY" {
		return nil, errors.New("unavailable")
	}
	return &dto.ExchangeRateResponse{Base: "CNY", Rates: map[string]float64{"USD": 0.14, "JPY": 20}}, nil
}

func newStatsService(t *testing.T) (*StatsService, *mockrepo.MockFuelRecordRepo, *mockrepo.MockVehicleRepo, *mockrepo.MockUserRepo, *mockrepo.MockGroupRepo) {
	ctrl := gomock.NewController(t)
	fuelRepo := mockrepo.NewMockFuelRecordRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	expenseRepo := mockrepo.NewMockExpenseRecordRepo(ctrl)
	svc := NewStatsService(fuelRepo, vehicleRepo, userRepo, groupRepo, expenseRepo, fakeRates{}, zap.NewNop())
	return svc, fuelRepo, vehicleRepo, userRepo, groupRepo
}

func cnyUser(uid uuid.UUID, effUnit, system string) *model.User {
	return &model.User{BaseModel: model.BaseModel{ID: uid}, FuelEfficiencyUnit: effUnit, UnitSystem: system, CurrencyCode: "CNY"}
}

func TestStatsService_GetVehicleStats(t *testing.T) {
	t.Run("costs converted, fuel normalized, efficiency weighted by distance", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid, vid := uuid.New(), uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid}, UserID: uid, FuelType: model.FuelTypeGasoline,
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(cnyUser(uid, "L/100km", "metric"), nil)
		// 区间 1：6 L/100km × 100 km；区间 2：10 L/100km × 300 km → 加权平均 9（简单平均会得到 8）
		fuelRepo.EXPECT().Aggregate(gomock.Any(), repository.FuelAggregateQuery{VehicleIDs: []uuid.UUID{vid}}).Return([]repository.FuelAggregateRow{
			{VehicleID: vid, CurrencyCode: "CNY", Cost: 300, FuelAggregate: repository.FuelAggregate{
				Records: 2, Liters: 40, Distance: 100, LiquidEffRecords: 1, LiquidEffSum: 600, LiquidEffWeight: 100, LiquidBest: 6, LiquidWorst: 6,
			}},
			{VehicleID: vid, CurrencyCode: "USD", Cost: 14, FuelAggregate: repository.FuelAggregate{
				Records: 1, Liters: 30, Distance: 300, LiquidEffRecords: 1, LiquidEffSum: 3000, LiquidEffWeight: 300, LiquidBest: 10, LiquidWorst: 10,
			}},
			{VehicleID: vid, CurrencyCode: "EUR", Cost: 5, FuelAggregate: repository.FuelAggregate{Records: 1}},
		}, nil)

		resp, err := svc.GetVehicleStats(context.Background(), vid, uid)
		assert.NoError(t, err)
		assert.Equal(t, int64(4), resp.TotalRecords)
		assert.InDelta(t, 400, resp.TotalCost, 0.001) // 300 CNY + 14 USD(=100 CNY)，EUR 无汇率不计入
		assert.Equal(t, []string{"EUR"}, resp.UnconvertedCurrencies)
		assert.Equal(t, 70.0, resp.TotalFuel)
		assert.Equal(t, "L", resp.VolumeUnit)
		assert.InDelta(t, 9.0, resp.AvgEfficiency, 0.001)
		assert.Equal(t, 6.0, resp.BestEfficiency)
		assert.Equal(t, 10.0, resp.WorstEfficiency)
		assert.InDelta(t, 1.0, resp.AvgCostPerKm, 0.001)
		assert.Equal(t, "fuel", resp.EnergyType)
	})

	t.Run("electric vehicle uses energy units", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid, vid := uuid.New(), uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid}, UserID: uid, FuelType: model.FuelTypeElectric,
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(cnyUser(uid, "MPG", "imperial"), nil)
		fuelRepo.EXPECT().Aggregate(gomock.Any(), gomock.Any()).Return([]repository.FuelAggregateRow{
			{VehicleID: vid, CurrencyCode: "CNY", Cost: 100, FuelAggregate: repository.FuelAggregate{
				Records: 2, Kwh: 45, Distance: 300, ElectricEffRecords: 1, ElectricEffSum: 4500, ElectricEffWeight: 300, ElectricBest: 15, ElectricWorst: 15,
			}},
		}, nil)

		resp, err := svc.GetVehicleStats(context.Background(), vid, uid)
		assert.NoError(t, err)
		assert.Equal(t, "electric", resp.EnergyType)
		assert.Equal(t, 45.0, resp.TotalFuel)
		assert.Equal(t, "kWh", resp.VolumeUnit)
		assert.Equal(t, "mi/kWh", resp.FuelUnit)
		assert.Equal(t, 4.14, resp.AvgEfficiency) // 15 kWh/100km → 4.14 mi/kWh
	})

	t.Run("rate service unavailable: only same-currency costs counted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fuelRepo := mockrepo.NewMockFuelRecordRepo(ctrl)
		vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
		userRepo := mockrepo.NewMockUserRepo(ctrl)
		svc := NewStatsService(fuelRepo, vehicleRepo, userRepo, nil, nil, fakeRates{fail: true}, zap.NewNop())
		uid, vid := uuid.New(), uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{BaseModel: model.BaseModel{ID: vid}}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(cnyUser(uid, "L/100km", "metric"), nil)
		fuelRepo.EXPECT().Aggregate(gomock.Any(), gomock.Any()).Return([]repository.FuelAggregateRow{
			{VehicleID: vid, CurrencyCode: "CNY", Cost: 300, FuelAggregate: repository.FuelAggregate{Records: 1}},
			{VehicleID: vid, CurrencyCode: "USD", Cost: 14, FuelAggregate: repository.FuelAggregate{Records: 1}},
		}, nil)

		resp, err := svc.GetVehicleStats(context.Background(), vid, uid)
		assert.NoError(t, err)
		assert.Equal(t, 300.0, resp.TotalCost)
		assert.Equal(t, []string{"USD"}, resp.UnconvertedCurrencies)
		assert.Equal(t, map[string]float64{"CNY": 300, "USD": 14}, resp.CostsByCurrency)
	})

	t.Run("vehicle not found", func(t *testing.T) {
		svc, _, vehicleRepo, _, groupRepo := newStatsService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(nil, gorm.ErrRecordNotFound)
		groupRepo.EXPECT().IsVehicleSharedToUser(gomock.Any(), vid, uid).Return(false, nil)

		_, err := svc.GetVehicleStats(context.Background(), vid, uid)
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestStatsService_GetPeriodStats(t *testing.T) {
	t.Run("monthly merges currencies per period", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid, vid := uuid.New(), uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{BaseModel: model.BaseModel{ID: vid}}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(cnyUser(uid, "L/100km", "metric"), nil)
		fuelRepo.EXPECT().Aggregate(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, q repository.FuelAggregateQuery) ([]repository.FuelAggregateRow, error) {
			assert.Equal(t, repository.PeriodMonth, q.PeriodFormat)
			if q.Start.Year() == 2025 {
				return nil, nil
			}
			return []repository.FuelAggregateRow{
				{Period: "2026-01", CurrencyCode: "CNY", Cost: 100, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 30}},
				{Period: "2026-01", CurrencyCode: "JPY", Cost: 2000, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 20}},
				{Period: "2026-02", CurrencyCode: "CNY", Cost: 50, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 10}},
			}, nil
		}).Times(2)

		resp, err := svc.GetPeriodStats(context.Background(), vid, uid, "month", 2026)
		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, "2026-01", resp.Items[0].Period)
		assert.Equal(t, 2, resp.Items[0].TotalRecords)
		assert.InDelta(t, 200, resp.Items[0].TotalCost, 0.001) // 100 CNY + 2000 JPY(=100 CNY)
		assert.Equal(t, 50.0, resp.Items[0].TotalFuel)
		assert.Empty(t, resp.PrevItems)
	})

	t.Run("yearly", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid, vid := uuid.New(), uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{BaseModel: model.BaseModel{ID: vid}}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(cnyUser(uid, "L/100km", "metric"), nil)
		fuelRepo.EXPECT().Aggregate(gomock.Any(), repository.FuelAggregateQuery{VehicleIDs: []uuid.UUID{vid}, PeriodFormat: repository.PeriodYear}).Return(nil, nil)

		resp, err := svc.GetPeriodStats(context.Background(), vid, uid, "year", 2026)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
