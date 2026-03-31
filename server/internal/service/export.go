package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// UserExportData 用户导出数据的聚合结构（P0 ~ P1 全量）
type UserExportData struct {
	User            *model.User
	Vehicles        []model.Vehicle
	FuelRecords     []model.FuelRecord
	ExpenseRecords  []model.ExpenseRecord
	Reminders       []model.Reminder
	Notifications   []model.Notification
	InviteCodes     []model.InviteCode
	Groups          []model.Group
	GroupMemberships []model.GroupMember
	SharedVehicles  []model.SharedVehicle
}

// ExportFormat 导出格式枚举
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatZIP  ExportFormat = "zip"
	ExportFormatJSON ExportFormat = "json"
)

// ExportScope 导出范围枚举
type ExportScope string

const (
	ExportScopeBasic ExportScope = "basic" // P0：用户 + 车辆 + 加油记录
	ExportScopeFull  ExportScope = "full"  // P1：全量数据
)

// ExportService 数据导出业务逻辑
type ExportService struct {
	userRepo           *repository.UserRepository
	vehicleRepo        *repository.VehicleRepository
	fuelRecordRepo     *repository.FuelRecordRepository
	expenseRecordRepo  *repository.ExpenseRecordRepository
	reminderRepo       *repository.ReminderRepository
	notificationRepo   *repository.NotificationRepository
	inviteCodeRepo     *repository.InviteCodeRepository
	groupRepo          *repository.GroupRepository
	logger             *zap.Logger
}

// NewExportService 创建 ExportService 实例
func NewExportService(
	userRepo *repository.UserRepository,
	vehicleRepo *repository.VehicleRepository,
	fuelRecordRepo *repository.FuelRecordRepository,
	expenseRecordRepo *repository.ExpenseRecordRepository,
	reminderRepo *repository.ReminderRepository,
	notificationRepo *repository.NotificationRepository,
	inviteCodeRepo *repository.InviteCodeRepository,
	groupRepo *repository.GroupRepository,
	logger *zap.Logger,
) *ExportService {
	return &ExportService{
		userRepo:          userRepo,
		vehicleRepo:       vehicleRepo,
		fuelRecordRepo:    fuelRecordRepo,
		expenseRecordRepo: expenseRecordRepo,
		reminderRepo:      reminderRepo,
		notificationRepo:  notificationRepo,
		inviteCodeRepo:    inviteCodeRepo,
		groupRepo:         groupRepo,
		logger:            logger,
	}
}

// GatherUserData 收集当前用户的所有数据用于导出
// scope=basic: 仅 P0 数据（用户 + 车辆 + 加油记录）
// scope=full: 全量 P1 数据
func (s *ExportService) GatherUserData(ctx context.Context, userID uuid.UUID, scope ExportScope) (*UserExportData, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user for export", err)
	}

	// 2. 获取所有车辆（含归档）
	vehicles, err := s.vehicleRepo.ListByUser(ctx, userID, true)
	if err != nil {
		return nil, apperror.ErrInternal("fetching vehicles for export", err)
	}

	// 3. 获取所有加油记录
	fuelRecords, err := s.fuelRecordRepo.ListAllByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching fuel records for export", err)
	}

	data := &UserExportData{
		User:        user,
		Vehicles:    vehicles,
		FuelRecords: fuelRecords,
	}

	// P0 基础版到此为止
	if scope == ExportScopeBasic {
		return data, nil
	}

	// === P1: 全量数据 ===

	// 4. 获取开销记录（全量按用户）
	expenseRecords, err := s.expenseRecordRepo.ListAllByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching expense records for export failed, skipping", zap.Error(err))
	} else {
		data.ExpenseRecords = expenseRecords
	}

	// 5. 获取保养提醒
	reminders, err := s.reminderRepo.ListByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching reminders for export failed, skipping", zap.Error(err))
	} else {
		data.Reminders = reminders
	}

	// 6. 获取通知（全量）
	notifications, err := s.notificationRepo.ListAllByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching notifications for export failed, skipping", zap.Error(err))
	} else {
		data.Notifications = notifications
	}

	// 7. 获取邀请码
	inviteCodes, err := s.inviteCodeRepo.ListByCreator(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching invite codes for export failed, skipping", zap.Error(err))
	} else {
		data.InviteCodes = inviteCodes
	}

	// 8. 获取群组关系
	groups, err := s.groupRepo.ListGroupsByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching groups for export failed, skipping", zap.Error(err))
	} else {
		data.Groups = groups
	}

	// 9. 获取群组成员身份
	memberships, err := s.groupRepo.ListMembershipsByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching group memberships for export failed, skipping", zap.Error(err))
	} else {
		data.GroupMemberships = memberships
	}

	// 10. 获取共享车辆关系（我共享出去的）
	sharedVehicles, err := s.groupRepo.ListSharedVehiclesByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("fetching shared vehicles for export failed, skipping", zap.Error(err))
	} else {
		data.SharedVehicles = sharedVehicles
	}

	return data, nil
}
