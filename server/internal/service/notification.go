package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// AnomalyThreshold 异常油耗阈值：偏离历史均值超过 30% 视为异常
const AnomalyThreshold = 0.30

// NotificationService 通知业务逻辑
type NotificationService struct {
	notificationRepo repository.NotificationRepo
	fuelRecordRepo   repository.FuelRecordRepo
	reminderRepo     repository.ReminderRepo
	logger           *zap.Logger
}

// NewNotificationService 创建 NotificationService 实例
func NewNotificationService(
	notificationRepo repository.NotificationRepo,
	fuelRecordRepo repository.FuelRecordRepo,
	reminderRepo repository.ReminderRepo,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		fuelRecordRepo:   fuelRecordRepo,
		reminderRepo:     reminderRepo,
		logger:           logger,
	}
}

// List 获取用户的通知列表
func (s *NotificationService) List(ctx context.Context, userID uuid.UUID) ([]dto.NotificationResponse, error) {
	notifications, err := s.notificationRepo.ListByUser(ctx, userID, 50)
	if err != nil {
		return nil, apperror.ErrInternal("listing notifications", err)
	}

	result := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = notificationToResponse(&n)
	}
	return result, nil
}

// UnreadCount 获取未读通知数
func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.notificationRepo.CountUnread(ctx, userID)
	if err != nil {
		return 0, apperror.ErrInternal("counting unread notifications", err)
	}
	return count, nil
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.notificationRepo.MarkAsRead(ctx, id, userID); err != nil {
		return apperror.ErrInternal("marking notification as read", err)
	}
	return nil
}

// MarkAllAsRead 标记所有通知为已读
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return apperror.ErrInternal("marking all notifications as read", err)
	}
	return nil
}

// Delete 删除通知
func (s *NotificationService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.notificationRepo.Delete(ctx, id, userID); err != nil {
		return apperror.ErrInternal("deleting notification", err)
	}
	return nil
}

// CheckFuelAnomaly 检查油耗异常并生成通知
// 在创建加油记录后调用，检测本次油耗是否偏离历史均值超过阈值
func (s *NotificationService) CheckFuelAnomaly(ctx context.Context, userID, vehicleID uuid.UUID, recordID uuid.UUID, currentEfficiency float64) {
	// 获取车辆统计数据
	stats, err := s.fuelRecordRepo.GetVehicleStats(ctx, vehicleID)
	if err != nil || stats.TotalRecords < 3 {
		// 至少 3 条记录才能做有意义的异常检测
		return
	}

	avgEfficiency := stats.AvgEfficiency
	if avgEfficiency <= 0 || currentEfficiency <= 0 {
		return
	}

	// 计算偏离百分比
	deviation := math.Abs(currentEfficiency-avgEfficiency) / avgEfficiency
	if deviation < AnomalyThreshold {
		return
	}

	// 偏离超过阈值，生成异常通知
	deviationPct := int(deviation * 100)
	direction := "higher"
	if currentEfficiency < avgEfficiency {
		direction = "lower"
	}

	// 存储结构化数据，前端根据 type 和 message 参数渲染本地化文本
	title := "notification.anomalyFuel"
	message := fmt.Sprintf(
		"current=%.1f;avg=%.1f;direction=%s;pct=%d",
		currentEfficiency, avgEfficiency, direction, deviationPct,
	)

	notification := &model.Notification{
		UserID:    userID,
		VehicleID: &vehicleID,
		Type:      model.NotificationTypeAnomalyFuel,
		Title:     title,
		Message:   message,
		RecordID:  &recordID,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		s.logger.Error("failed to create fuel anomaly notification",
			zap.Error(err),
			zap.String("vehicle_id", vehicleID.String()),
		)
	}
}

// CheckMaintenanceReminders 检查保养提醒（根据当前里程触发）
// 在创建加油记录后调用（因为加油时会更新里程）
func (s *NotificationService) CheckMaintenanceReminders(ctx context.Context, userID, vehicleID uuid.UUID, currentOdometer float64) {
	reminders, err := s.reminderRepo.ListEnabledByVehicle(ctx, vehicleID)
	if err != nil {
		s.logger.Error("failed to list reminders for maintenance check", zap.Error(err))
		return
	}

	for _, r := range reminders {
		if r.Trigger == model.ReminderTriggerMileage || r.Trigger == model.ReminderTriggerBoth {
			if r.NextMileage > 0 && currentOdometer >= r.NextMileage {
				// 里程已达到，生成保养提醒通知（结构化数据）
				title := "notification.maintenanceDue"
				message := fmt.Sprintf(
					"reminderTitle=%s;currentOdo=%.0f;targetOdo=%.0f",
					r.Title, currentOdometer, r.NextMileage,
				)

				notification := &model.Notification{
					UserID:     userID,
					VehicleID:  &vehicleID,
					Type:       model.NotificationTypeMaintenance,
					Title:      title,
					Message:    message,
					ReminderID: &r.ID,
				}

				if err := s.notificationRepo.Create(ctx, notification); err != nil {
					s.logger.Error("failed to create maintenance notification",
						zap.Error(err),
						zap.String("reminder_id", r.ID.String()),
					)
				}
			}
		}
	}
}

// NotifyInviteUsed 邀请码被使用时通知创建者
func (s *NotificationService) NotifyInviteUsed(ctx context.Context, inviteCreatorID uuid.UUID, inviteCode string, newUserNickname string) {
	// 结构化数据，前端根据 type 渲染本地化文本
	title := "notification.inviteUsed"
	message := fmt.Sprintf("nickname=%s;code=%s", newUserNickname, inviteCode)

	notification := &model.Notification{
		UserID:  inviteCreatorID,
		Type:    model.NotificationTypeInviteUsed,
		Title:   title,
		Message: message,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		s.logger.Error("failed to create invite used notification",
			zap.Error(err),
			zap.String("invite_code", inviteCode),
			zap.String("creator_id", inviteCreatorID.String()),
		)
	}
}

// notificationToResponse 将 model 转为 DTO
func notificationToResponse(n *model.Notification) dto.NotificationResponse {
	resp := dto.NotificationResponse{
		ID:        n.ID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
	if n.VehicleID != nil {
		s := n.VehicleID.String()
		resp.VehicleID = &s
	}
	if n.ReminderID != nil {
		s := n.ReminderID.String()
		resp.ReminderID = &s
	}
	if n.RecordID != nil {
		s := n.RecordID.String()
		resp.RecordID = &s
	}
	return resp
}
