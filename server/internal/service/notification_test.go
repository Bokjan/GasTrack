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

	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
	mockrepo "gastrack/internal/repository/mock"
)

func newNotificationService(t *testing.T) (*NotificationService, *mockrepo.MockNotificationRepo, *mockrepo.MockFuelRecordRepo, *mockrepo.MockReminderRepo) {
	ctrl := gomock.NewController(t)
	notifRepo := mockrepo.NewMockNotificationRepo(ctrl)
	fuelRepo := mockrepo.NewMockFuelRecordRepo(ctrl)
	reminderRepo := mockrepo.NewMockReminderRepo(ctrl)
	svc := NewNotificationService(notifRepo, fuelRepo, reminderRepo, zap.NewNop())
	return svc, notifRepo, fuelRepo, reminderRepo
}

func TestNotificationService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, notifRepo, _, _ := newNotificationService(t)
		uid := uuid.New()
		notifRepo.EXPECT().ListByUser(gomock.Any(), uid, 50).Return([]model.Notification{
			{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: uid, Type: model.NotificationTypeInviteUsed, Title: "t", Message: "m"},
		}, nil)

		result, err := svc.List(context.Background(), uid)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("error", func(t *testing.T) {
		svc, notifRepo, _, _ := newNotificationService(t)
		uid := uuid.New()
		notifRepo.EXPECT().ListByUser(gomock.Any(), uid, 50).Return(nil, errors.New("db error"))

		result, err := svc.List(context.Background(), uid)
		assert.Nil(t, result)
		assert.Equal(t, 500, err.(*apperror.AppError).Code)
	})
}

func TestNotificationService_UnreadCount(t *testing.T) {
	svc, notifRepo, _, _ := newNotificationService(t)
	uid := uuid.New()
	notifRepo.EXPECT().CountUnread(gomock.Any(), uid).Return(int64(5), nil)

	count, err := svc.UnreadCount(context.Background(), uid)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	svc, notifRepo, _, _ := newNotificationService(t)
	uid := uuid.New()
	nid := uuid.New()
	notifRepo.EXPECT().MarkAsRead(gomock.Any(), nid, uid).Return(nil)

	err := svc.MarkAsRead(context.Background(), nid, uid)
	assert.NoError(t, err)
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	svc, notifRepo, _, _ := newNotificationService(t)
	uid := uuid.New()
	notifRepo.EXPECT().MarkAllAsRead(gomock.Any(), uid).Return(nil)

	err := svc.MarkAllAsRead(context.Background(), uid)
	assert.NoError(t, err)
}

func TestNotificationService_Delete(t *testing.T) {
	svc, notifRepo, _, _ := newNotificationService(t)
	uid := uuid.New()
	nid := uuid.New()
	notifRepo.EXPECT().Delete(gomock.Any(), nid, uid).Return(nil)

	err := svc.Delete(context.Background(), nid, uid)
	assert.NoError(t, err)
}

func TestNotificationService_CheckFuelAnomaly(t *testing.T) {
	t.Run("anomaly detected - creates notification", func(t *testing.T) {
		svc, notifRepo, fuelRepo, _ := newNotificationService(t)
		uid := uuid.New()
		vid := uuid.New()
		rid := uuid.New()

		fuelRepo.EXPECT().GetVehicleStats(gomock.Any(), vid).Return(&repository.StatsResult{
			TotalRecords:  10,
			AvgEfficiency: 8.0,
		}, nil)
		notifRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		// 12 L/100km is 50% above 8.0 avg => anomaly
		svc.CheckFuelAnomaly(context.Background(), uid, vid, rid, 12.0)
	})

	t.Run("no anomaly - within threshold", func(t *testing.T) {
		svc, _, fuelRepo, _ := newNotificationService(t)
		vid := uuid.New()

		fuelRepo.EXPECT().GetVehicleStats(gomock.Any(), vid).Return(&repository.StatsResult{
			TotalRecords:  10,
			AvgEfficiency: 8.0,
		}, nil)
		// No Create call expected

		svc.CheckFuelAnomaly(context.Background(), uuid.New(), vid, uuid.New(), 8.5)
	})

	t.Run("too few records - skips", func(t *testing.T) {
		svc, _, fuelRepo, _ := newNotificationService(t)
		vid := uuid.New()

		fuelRepo.EXPECT().GetVehicleStats(gomock.Any(), vid).Return(&repository.StatsResult{
			TotalRecords:  2,
			AvgEfficiency: 8.0,
		}, nil)

		svc.CheckFuelAnomaly(context.Background(), uuid.New(), vid, uuid.New(), 20.0)
	})
}

func TestNotificationService_CheckMaintenanceReminders(t *testing.T) {
	t.Run("mileage trigger reached", func(t *testing.T) {
		svc, notifRepo, _, reminderRepo := newNotificationService(t)
		uid := uuid.New()
		vid := uuid.New()
		remID := uuid.New()

		reminderRepo.EXPECT().ListEnabledByVehicle(gomock.Any(), vid).Return([]model.Reminder{
			{
				BaseModel:   model.BaseModel{ID: remID},
				Title:       "Oil Change",
				Trigger:     model.ReminderTriggerMileage,
				NextMileage: 10000,
			},
		}, nil)
		notifRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		svc.CheckMaintenanceReminders(context.Background(), uid, vid, 10500)
	})

	t.Run("mileage not reached - no notification", func(t *testing.T) {
		svc, _, _, reminderRepo := newNotificationService(t)
		vid := uuid.New()

		reminderRepo.EXPECT().ListEnabledByVehicle(gomock.Any(), vid).Return([]model.Reminder{
			{
				BaseModel:   model.BaseModel{ID: uuid.New()},
				Trigger:     model.ReminderTriggerMileage,
				NextMileage: 10000,
			},
		}, nil)

		svc.CheckMaintenanceReminders(context.Background(), uuid.New(), vid, 5000)
	})
}

func TestNotificationService_NotifyInviteUsed(t *testing.T) {
	svc, notifRepo, _, _ := newNotificationService(t)
	creatorID := uuid.New()
	notifRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, n *model.Notification) error {
		assert.Equal(t, creatorID, n.UserID)
		assert.Equal(t, model.NotificationTypeInviteUsed, n.Type)
		return nil
	})

	svc.NotifyInviteUsed(context.Background(), creatorID, "GT-ABC123", "NewUser")
}

// Ensure NotificationService satisfies NotificationServicer interface
var _ NotificationServicer = (*NotificationService)(nil)

// Suppress unused import
var _ = time.Now
