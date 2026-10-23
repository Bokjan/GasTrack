package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
	svc := NewExpenseRecordService(expenseRepo, vehicleRepo, groupRepo, reminderRepo, nil, nil, zap.NewNop())
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

	t.Run("reminder of another vehicle is rejected", func(t *testing.T) {
		svc, _, vehicleRepo, _, reminderRepo := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()
		foreignReminder := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{BaseModel: model.BaseModel{ID: vid}, UserID: uid}, nil)
		reminderRepo.EXPECT().GetByID(gomock.Any(), foreignReminder).Return(&model.Reminder{
			BaseModel: model.BaseModel{ID: foreignReminder},
			VehicleID: uuid.New(), // 别人的车
		}, nil)
		// 不允许出现 expenseRepo.Create / reminderRepo.Update 调用

		_, err := svc.Create(context.Background(), uid, vid, &dto.CreateExpenseRequest{
			Title:               "Oil Change",
			Category:            "maintenance",
			MaintenanceCategory: "oil_change",
			Amount:              350,
			CurrencyCode:        "CNY",
			ExpenseDate:         "2026-03-15",
			ReminderID:          foreignReminder.String(),
		})
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})

	t.Run("linked reminder baseline updated after create and renotification re-armed", func(t *testing.T) {
		svc, expenseRepo, vehicleRepo, _, reminderRepo := newExpenseRecordService(t)
		uid := uuid.New()
		vid := uuid.New()
		remID := uuid.New()
		notifiedAt := time.Now()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{BaseModel: model.BaseModel{ID: vid}, UserID: uid}, nil)
		reminderRepo.EXPECT().GetByID(gomock.Any(), remID).Return(&model.Reminder{
			BaseModel:       model.BaseModel{ID: remID},
			VehicleID:       vid,
			Trigger:         model.ReminderTriggerMileage,
			MileageInterval: 5000,
			LastMileage:     5000,
			NextMileage:     10000,
			DueNotifiedAt:   &notifiedAt,
		}, nil)
		gomock.InOrder(
			expenseRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil),
			reminderRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *model.Reminder) error {
				assert.Equal(t, 15500.0, r.NextMileage)
				assert.Nil(t, r.DueNotifiedAt)
				return nil
			}),
		)

		_, err := svc.Create(context.Background(), uid, vid, &dto.CreateExpenseRequest{
			Title:               "Oil Change",
			Category:            "maintenance",
			MaintenanceCategory: "oil_change",
			Amount:              350,
			CurrencyCode:        "CNY",
			Odometer:            10500,
			ExpenseDate:         "2026-03-15",
			ReminderID:          remID.String(),
		})
		assert.NoError(t, err)
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
	ctrl := gomock.NewController(t)
	expenseRepo := mockrepo.NewMockExpenseRecordRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	svc := NewExpenseRecordService(expenseRepo, vehicleRepo, nil, nil, userRepo, fakeRates{}, zap.NewNop())
	uid := uuid.New()
	vid := uuid.New()

	vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
		BaseModel: model.BaseModel{ID: vid},
	}, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{CurrencyCode: "CNY"}, nil)
	expenseRepo.EXPECT().GetTotalsByCurrency(gomock.Any(), vid).Return([]repository.ExpenseStatsByCurrency{
		{CurrencyCode: "CNY", TotalAmount: 5000, RecordCount: 10},
	}, nil)
	expenseRepo.EXPECT().GetBreakdownByCategory(gomock.Any(), vid).Return([]repository.ExpenseStatsByCategory{
		{Category: "maintenance", CurrencyCode: "CNY", TotalAmount: 300, RecordCount: 2},
		{Category: "maintenance", CurrencyCode: "USD", TotalAmount: 14, RecordCount: 1}, // = 100 CNY
		{Category: "insurance", CurrencyCode: "CNY", TotalAmount: 600, RecordCount: 1},
		{Category: "parking", CurrencyCode: "EUR", TotalAmount: 9, RecordCount: 1}, // 无汇率
	}, nil)
	expenseRepo.EXPECT().GetMonthlyTrend(gomock.Any(), vid).Return([]repository.ExpenseStatsByMonth{
		{Period: "2026-01", CurrencyCode: "CNY", TotalAmount: 100, RecordCount: 1},
		{Period: "2026-01", CurrencyCode: "JPY", TotalAmount: 2000, RecordCount: 1}, // = 100 CNY
	}, nil)
	expenseRepo.EXPECT().GetLast30DaysByCurrency(gomock.Any(), vid).Return([]repository.ExpenseStatsByCurrency{
		{CurrencyCode: "CNY", TotalAmount: 800}, {CurrencyCode: "USD", TotalAmount: 14},
	}, nil)
	expenseRepo.EXPECT().GetTotalRecords(gomock.Any(), vid).Return(int64(10), nil)

	resp, err := svc.GetStats(context.Background(), uid, vid)
	assert.NoError(t, err)
	assert.Equal(t, "CNY", resp.CurrencyCode)
	assert.Equal(t, []string{"EUR"}, resp.UnconvertedCurrencies)
	assert.Equal(t, "insurance", resp.CategoryBreakdown[0].Category) // 按换算后金额降序
	assert.InDelta(t, 400, resp.CategoryBreakdown[1].TotalAmount, 0.001)
	assert.Equal(t, 3, resp.CategoryBreakdown[1].RecordCount)
	assert.InDelta(t, 40, resp.CategoryBreakdown[1].Percentage, 0.001) // 400 / 1000
	assert.Len(t, resp.MonthlyTrend, 1)
	assert.InDelta(t, 200, resp.MonthlyTrend[0].TotalAmount, 0.001)
	assert.InDelta(t, 900, resp.Last30DaysAmount, 0.001)
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
