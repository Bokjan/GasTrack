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

func newExpenseRecordService(t *testing.T) (*ExpenseRecordService, *mockrepo.MockExpenseRecordRepo, *mockrepo.MockVehicleRepo, *mockrepo.MockGroupRepo, *mockrepo.MockReminderRepo) {
	ctrl := gomock.NewController(t)
	expenseRepo := mockrepo.NewMockExpenseRecordRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	reminderRepo := mockrepo.NewMockReminderRepo(ctrl)
	svc := NewExpenseRecordService(expenseRepo, vehicleRepo, groupRepo, reminderRepo, zap.NewNop())
	return svc, expenseRepo, vehicleRepo, groupRepo, reminderRepo
}

func TestExpenseRecordService_Create(t *testing.T) {
	t.Run("success as owner", func(t *testing.T) {
		svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			UserID:    uid,
		}, nil)
		expenseRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), uid, vid, &dto.CreateExpenseRequest{
			Title:               "Oil Change",
			Category:            "maintenance",
			MaintenanceCategory: "oil_change",
			Amount:              350,
			CurrencyCode:        "CNY",
			ExpenseDate:         "2026-03-15",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Oil Change", resp.Title)
	})

	t.Run("vehicle not found", func(t *testing.T) {
		svc, _, vehicleRepo, groupRepo, _ := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(nil, gorm.ErrRecordNotFound)
		groupRepo.EXPECT().IsVehicleSharedToUser(gomock.Any(), vid, uid).Return(false, nil)

		_, err := svc.Create(context.Background(), uid, vid, &dto.CreateExpenseRequest{
			Title:       "Test",
			Category:    "repair",
			Amount:      100,
			ExpenseDate: "2026-03-15",
		})
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestExpenseRecordService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()
		eid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
		}, nil)
		expenseRepo.EXPECT().GetByIDAndVehicle(gomock.Any(), eid, vid).Return(&model.ExpenseRecord{
			BaseModel:    model.BaseModel{ID: eid},
			VehicleID:    vid,
			Title:        "Test",
			CurrencyCode: "CNY",
		}, nil)

		resp, err := svc.GetByID(context.Background(), eid, vid, uid)
		assert.NoError(t, err)
		assert.Equal(t, "Test", resp.Title)
	})

	t.Run("not found", func(t *testing.T) {
		svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{}, nil)
		expenseRepo.EXPECT().GetByIDAndVehicle(gomock.Any(), gomock.Any(), vid).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetByID(context.Background(), uuid.New(), vid, uid)
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestExpenseRecordService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()
		eid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			UserID:    uid,
		}, nil)
		expenseRepo.EXPECT().GetByIDAndVehicle(gomock.Any(), eid, vid).Return(&model.ExpenseRecord{
			BaseModel: model.BaseModel{ID: eid},
			VehicleID: vid,
		}, nil)
		expenseRepo.EXPECT().Delete(gomock.Any(), eid, vid).Return(nil)

		err := svc.Delete(context.Background(), eid, vid, uid)
		assert.NoError(t, err)
	})
}

func TestExpenseRecordService_GetStats(t *testing.T) {
	svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
	uid := uuid.New()
	vid := uuid.New()

	vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
		BaseModel: model.BaseModel{ID: vid},
	}, nil)
	expenseRepo.EXPECT().GetTotalsByCurrency(gomock.Any(), vid).Return([]repository.ExpenseStatsByCurrency{
		{CurrencyCode: "CNY", TotalAmount: 5000, RecordCount: 10},
	}, nil)
	expenseRepo.EXPECT().GetBreakdownByCategory(gomock.Any(), vid).Return([]repository.ExpenseStatsByCategory{
		{Category: "maintenance", TotalAmount: 3000, RecordCount: 5},
	}, nil)
	expenseRepo.EXPECT().GetMonthlyTrend(gomock.Any(), vid).Return([]repository.ExpenseStatsByMonth{}, nil)
	expenseRepo.EXPECT().GetLast30DaysTotal(gomock.Any(), vid).Return(float64(800), "CNY", nil)
	expenseRepo.EXPECT().GetTotalRecords(gomock.Any(), vid).Return(int64(10), nil)

	resp, err := svc.GetStats(context.Background(), uid, vid)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestExpenseRecordService_GetVendorSuggestions(t *testing.T) {
	svc, expenseRepo, vehicleRepo, _, _ := newExpenseRecordService(t)
	uid := uuid.New()
	vid := uuid.New()

	vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
		BaseModel: model.BaseModel{ID: vid},
	}, nil)
	expenseRepo.EXPECT().GetDistinctVendorNames(gomock.Any(), uid, gomock.Any(), 20).Return([]string{"AutoShop", "TireKing"}, nil)

	names, err := svc.GetVendorSuggestions(context.Background(), uid, vid)
	assert.NoError(t, err)
	assert.Equal(t, []string{"AutoShop", "TireKing"}, names)
}

// Suppress unused import
var _ = errors.New
