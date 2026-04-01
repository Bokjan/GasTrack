package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	mockrepo "gastrack/internal/repository/mock"
)

func newExportService(t *testing.T) (*ExportService, *mockrepo.MockUserRepo, *mockrepo.MockVehicleRepo, *mockrepo.MockFuelRecordRepo, *mockrepo.MockExpenseRecordRepo, *mockrepo.MockReminderRepo, *mockrepo.MockNotificationRepo, *mockrepo.MockInviteCodeRepo, *mockrepo.MockGroupRepo) {
	ctrl := gomock.NewController(t)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	fuelRepo := mockrepo.NewMockFuelRecordRepo(ctrl)
	expenseRepo := mockrepo.NewMockExpenseRecordRepo(ctrl)
	reminderRepo := mockrepo.NewMockReminderRepo(ctrl)
	notifRepo := mockrepo.NewMockNotificationRepo(ctrl)
	inviteRepo := mockrepo.NewMockInviteCodeRepo(ctrl)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	svc := NewExportService(userRepo, vehicleRepo, fuelRepo, expenseRepo, reminderRepo, notifRepo, inviteRepo, groupRepo, zap.NewNop())
	return svc, userRepo, vehicleRepo, fuelRepo, expenseRepo, reminderRepo, notifRepo, inviteRepo, groupRepo
}

func TestExportService_GatherUserData_Basic(t *testing.T) {
	svc, userRepo, vehicleRepo, fuelRepo, _, _, _, _, _ := newExportService(t)
	uid := uuid.New()

	userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
		BaseModel: model.BaseModel{ID: uid},
		Email:     "test@example.com",
	}, nil)
	vehicleRepo.EXPECT().ListByUser(gomock.Any(), uid, true).Return([]model.Vehicle{
		{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "Car1"},
	}, nil)
	fuelRepo.EXPECT().ListAllByUser(gomock.Any(), uid).Return([]model.FuelRecord{}, nil)

	data, err := svc.GatherUserData(context.Background(), uid, ExportScopeBasic)
	assert.NoError(t, err)
	assert.Equal(t, uid, data.User.ID)
	assert.Len(t, data.Vehicles, 1)
	assert.Empty(t, data.ExpenseRecords)
}

func TestExportService_GatherUserData_Full(t *testing.T) {
	svc, userRepo, vehicleRepo, fuelRepo, expenseRepo, reminderRepo, notifRepo, inviteRepo, groupRepo := newExportService(t)
	uid := uuid.New()

	userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{BaseModel: model.BaseModel{ID: uid}}, nil)
	vehicleRepo.EXPECT().ListByUser(gomock.Any(), uid, true).Return([]model.Vehicle{}, nil)
	fuelRepo.EXPECT().ListAllByUser(gomock.Any(), uid).Return([]model.FuelRecord{}, nil)
	expenseRepo.EXPECT().ListAllByUser(gomock.Any(), uid).Return([]model.ExpenseRecord{}, nil)
	reminderRepo.EXPECT().ListByUser(gomock.Any(), uid).Return([]model.Reminder{}, nil)
	notifRepo.EXPECT().ListAllByUser(gomock.Any(), uid).Return([]model.Notification{}, nil)
	inviteRepo.EXPECT().ListByCreator(gomock.Any(), uid).Return([]model.InviteCode{}, nil)
	groupRepo.EXPECT().ListGroupsByUser(gomock.Any(), uid).Return([]model.Group{}, nil)
	groupRepo.EXPECT().ListMembershipsByUser(gomock.Any(), uid).Return([]model.GroupMember{}, nil)
	groupRepo.EXPECT().ListSharedVehiclesByUser(gomock.Any(), uid).Return([]model.SharedVehicle{}, nil)

	data, err := svc.GatherUserData(context.Background(), uid, ExportScopeFull)
	assert.NoError(t, err)
	assert.NotNil(t, data.User)
}

func TestExportService_GatherUserData_UserError(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _, _ := newExportService(t)
	uid := uuid.New()
	userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(nil, assert.AnError)

	_, err := svc.GatherUserData(context.Background(), uid, ExportScopeBasic)
	assert.Equal(t, 500, err.(*apperror.AppError).Code)
}
