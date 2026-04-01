package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
	mockrepo "gastrack/internal/repository/mock"
)

func newStatsService(t *testing.T) (*StatsService, *mockrepo.MockFuelRecordRepo, *mockrepo.MockVehicleRepo, *mockrepo.MockUserRepo, *mockrepo.MockGroupRepo) {
	ctrl := gomock.NewController(t)
	fuelRepo := mockrepo.NewMockFuelRecordRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	svc := NewStatsService(fuelRepo, vehicleRepo, userRepo, groupRepo, zap.NewNop())
	return svc, fuelRepo, vehicleRepo, userRepo, groupRepo
}

func TestStatsService_GetVehicleStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			UserID:    uid,
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
			BaseModel:          model.BaseModel{ID: uid},
			FuelEfficiencyUnit: "L/100km",
			UnitSystem:         "metric",
		}, nil)
		fuelRepo.EXPECT().GetVehicleStats(gomock.Any(), vid).Return(&repository.StatsResult{
			TotalRecords:  10,
			TotalFuel:     500,
			TotalCost:     3000,
			TotalDistance:  6000,
			AvgEfficiency: 8.33,
		}, nil)
		fuelRepo.EXPECT().GetCostByCurrency(gomock.Any(), vid).Return([]repository.CostByCurrencyResult{
			{CurrencyCode: "CNY", TotalCost: 3000},
		}, nil)

		resp, err := svc.GetVehicleStats(context.Background(), vid, uid)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
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
	t.Run("success monthly", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
			FuelEfficiencyUnit: "L/100km",
			UnitSystem:         "metric",
			CurrencyCode:       "CNY",
		}, nil)
		fuelRepo.EXPECT().GetStatsByMonth(gomock.Any(), vid, 2026).Return([]repository.PeriodStatsResult{
			{Period: "2026-01", TotalRecords: 3, TotalFuel: 150, TotalCost: 900, TotalDistance: 1800, AvgEfficiency: 8.33},
		}, nil)
		fuelRepo.EXPECT().GetStatsByMonth(gomock.Any(), vid, 2025).Return([]repository.PeriodStatsResult{}, nil)

		resp, err := svc.GetPeriodStats(context.Background(), vid, uid, "month", 2026)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("success yearly", func(t *testing.T) {
		svc, fuelRepo, vehicleRepo, userRepo, _ := newStatsService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
			FuelEfficiencyUnit: "L/100km",
			UnitSystem:         "metric",
			CurrencyCode:       "CNY",
		}, nil)
		fuelRepo.EXPECT().GetStatsByYear(gomock.Any(), vid).Return([]repository.PeriodStatsResult{}, nil)

		resp, err := svc.GetPeriodStats(context.Background(), vid, uid, "year", 2026)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
