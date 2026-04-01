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
	mockrepo "gastrack/internal/repository/mock"
)

func newReminderService(t *testing.T) (*ReminderService, *mockrepo.MockReminderRepo, *mockrepo.MockVehicleRepo, *mockrepo.MockGroupRepo) {
	ctrl := gomock.NewController(t)
	reminderRepo := mockrepo.NewMockReminderRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	svc := NewReminderService(reminderRepo, vehicleRepo, groupRepo, zap.NewNop())
	return svc, reminderRepo, vehicleRepo, groupRepo
}

func TestReminderService_Create(t *testing.T) {
	t.Run("success with mileage trigger", func(t *testing.T) {
		svc, reminderRepo, vehicleRepo, _ := newReminderService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			Name:      "My Car",
			UserID:    uid,
		}, nil)
		reminderRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), uid, &dto.CreateReminderRequest{
			VehicleID:       vid.String(),
			Category:        "oil_change",
			Title:           "Oil Change",
			Trigger:         "mileage",
			MileageInterval: 5000,
			LastMileage:     10000,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Oil Change", resp.Title)
		assert.Equal(t, float64(15000), resp.NextMileage)
	})

	t.Run("invalid vehicle ID", func(t *testing.T) {
		svc, _, _, _ := newReminderService(t)

		_, err := svc.Create(context.Background(), uuid.New(), &dto.CreateReminderRequest{
			VehicleID: "not-a-uuid",
			Category:  "oil_change",
			Title:     "Test",
			Trigger:   "mileage",
		})
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})

	t.Run("vehicle not found", func(t *testing.T) {
		svc, _, vehicleRepo, groupRepo := newReminderService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(nil, gorm.ErrRecordNotFound)
		groupRepo.EXPECT().IsVehicleSharedToUser(gomock.Any(), vid, uid).Return(false, nil)

		_, err := svc.Create(context.Background(), uid, &dto.CreateReminderRequest{
			VehicleID: vid.String(),
			Category:  "oil_change",
			Title:     "Test",
			Trigger:   "mileage",
		})
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})

	t.Run("mileage interval required", func(t *testing.T) {
		svc, _, vehicleRepo, _ := newReminderService(t)
		uid := uuid.New()
		vid := uuid.New()

		vehicleRepo.EXPECT().GetByIDAndUser(gomock.Any(), vid, uid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			Name:      "Car",
		}, nil)

		_, err := svc.Create(context.Background(), uid, &dto.CreateReminderRequest{
			VehicleID:       vid.String(),
			Category:        "oil_change",
			Title:           "Test",
			Trigger:         "mileage",
			MileageInterval: 0, // missing
		})
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})
}

func TestReminderService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, reminderRepo, vehicleRepo, _ := newReminderService(t)
		uid := uuid.New()
		vid := uuid.New()

		reminderRepo.EXPECT().ListByUser(gomock.Any(), uid).Return([]model.Reminder{
			{BaseModel: model.BaseModel{ID: uuid.New()}, VehicleID: vid, Title: "R1"},
		}, nil)
		vehicleRepo.EXPECT().GetByID(gomock.Any(), vid).Return(&model.Vehicle{
			BaseModel: model.BaseModel{ID: vid},
			Name:      "Car",
		}, nil)

		result, err := svc.List(context.Background(), uid)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Car", result[0].VehicleName)
	})

	t.Run("error", func(t *testing.T) {
		svc, reminderRepo, _, _ := newReminderService(t)
		uid := uuid.New()
		reminderRepo.EXPECT().ListByUser(gomock.Any(), uid).Return(nil, errors.New("db error"))

		_, err := svc.List(context.Background(), uid)
		assert.Equal(t, 500, err.(*apperror.AppError).Code)
	})
}

func TestReminderService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, reminderRepo, vehicleRepo, _ := newReminderService(t)
		uid := uuid.New()
		rid := uuid.New()
		vid := uuid.New()

		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), rid, uid).Return(&model.Reminder{
			BaseModel: model.BaseModel{ID: rid},
			VehicleID: vid,
			Title:     "Oil Change",
		}, nil)
		vehicleRepo.EXPECT().GetByID(gomock.Any(), vid).Return(&model.Vehicle{Name: "Car"}, nil)

		resp, err := svc.GetByID(context.Background(), rid, uid)
		assert.NoError(t, err)
		assert.Equal(t, "Oil Change", resp.Title)
	})

	t.Run("not found", func(t *testing.T) {
		svc, reminderRepo, _, _ := newReminderService(t)
		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestReminderService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, reminderRepo, vehicleRepo, _ := newReminderService(t)
		uid := uuid.New()
		rid := uuid.New()
		vid := uuid.New()
		newTitle := "Updated Title"

		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), rid, uid).Return(&model.Reminder{
			BaseModel: model.BaseModel{ID: rid},
			VehicleID: vid,
			Title:     "Old Title",
			Trigger:   model.ReminderTriggerMileage,
		}, nil)
		reminderRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		vehicleRepo.EXPECT().GetByID(gomock.Any(), vid).Return(&model.Vehicle{Name: "Car"}, nil)

		resp, err := svc.Update(context.Background(), rid, uid, &dto.UpdateReminderRequest{Title: &newTitle})
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", resp.Title)
	})

	t.Run("not found", func(t *testing.T) {
		svc, reminderRepo, _, _ := newReminderService(t)
		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), &dto.UpdateReminderRequest{})
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestReminderService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, reminderRepo, _, _ := newReminderService(t)
		uid := uuid.New()
		rid := uuid.New()

		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), rid, uid).Return(&model.Reminder{
			BaseModel: model.BaseModel{ID: rid},
		}, nil)
		reminderRepo.EXPECT().Delete(gomock.Any(), rid, uid).Return(nil)

		err := svc.Delete(context.Background(), rid, uid)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		svc, reminderRepo, _, _ := newReminderService(t)
		reminderRepo.EXPECT().GetByIDAndUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		err := svc.Delete(context.Background(), uuid.New(), uuid.New())
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}
