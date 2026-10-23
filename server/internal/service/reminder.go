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

// ReminderService 提醒业务逻辑
type ReminderService struct {
	reminderRepo repository.ReminderRepo
	vehicleRepo  repository.VehicleRepo
	groupRepo    repository.GroupRepo
	logger       *zap.Logger
}

// NewReminderService 创建 ReminderService 实例
func NewReminderService(
	reminderRepo repository.ReminderRepo,
	vehicleRepo repository.VehicleRepo,
	groupRepo repository.GroupRepo,
	logger *zap.Logger,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		vehicleRepo:  vehicleRepo,
		groupRepo:    groupRepo,
		logger:       logger,
	}
}

// verifyVehicleAccess 验证用户对车辆的访问权限（所有权或共享访问）
func (s *ReminderService) verifyVehicleAccess(ctx context.Context, vehicleID, userID uuid.UUID) (*model.Vehicle, error) {
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err == nil {
		return vehicle, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrInternal("verifying vehicle ownership", err)
	}
	// 不是自己的车辆，检查是否为共享车辆
	if s.groupRepo != nil {
		shared, sharedErr := s.groupRepo.IsVehicleSharedToUser(ctx, vehicleID, userID)
		if sharedErr != nil {
			return nil, apperror.ErrInternal("checking shared vehicle access", sharedErr)
		}
		if shared {
			// 共享车辆，通过 GetByID 获取车辆信息
			v, getErr := s.vehicleRepo.GetByID(ctx, vehicleID)
			if getErr != nil {
				return nil, apperror.ErrInternal("fetching shared vehicle", getErr)
			}
			return v, nil
		}
	}
	return nil, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
}

// Create 创建提醒
func (s *ReminderService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateReminderRequest) (*dto.ReminderResponse, error) {
	// 验证车辆归属（支持共享车辆）
	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		return nil, apperror.ErrBadRequest("reminder.invalid_vehicle", "invalid vehicle ID")
	}

	vehicle, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	// 验证触发条件参数
	if err := validateTriggerIntervals(model.ReminderTrigger(req.Trigger), req.MileageInterval, req.TimeIntervalDays); err != nil {
		return nil, err
	}

	reminder := &model.Reminder{
		UserID:           userID,
		VehicleID:        vehicleID,
		Type:             model.ReminderTypeMaintenance,
		Category:         model.MaintenanceCategory(req.Category),
		Title:            req.Title,
		Description:      req.Description,
		Trigger:          model.ReminderTrigger(req.Trigger),
		MileageInterval:  req.MileageInterval,
		TimeIntervalDays: req.TimeIntervalDays,
		LastMileage:      req.LastMileage,
		IsEnabled:        true,
	}

	// 解析上次保养日期
	if req.LastDate != "" {
		t, err := time.Parse("2006-01-02", req.LastDate)
		if err != nil {
			return nil, apperror.ErrBadRequest("reminder.invalid_date", "invalid last date format")
		}
		reminder.LastDate = &t
	}

	// 计算下次触发
	s.calculateNext(reminder)

	if err := s.reminderRepo.Create(ctx, reminder); err != nil {
		return nil, apperror.ErrInternal("creating reminder", err)
	}

	resp := s.reminderToResponse(reminder, vehicle.Name)
	return &resp, nil
}

// List 获取用户的所有提醒
func (s *ReminderService) List(ctx context.Context, userID uuid.UUID) ([]dto.ReminderResponse, error) {
	reminders, err := s.reminderRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing reminders", err)
	}

	// 获取所有关联车辆名称
	// 批量查询车辆名称（1 次 SQL 替代 N 次）
	vehicleIDs := make([]uuid.UUID, 0, len(reminders))
	seen := make(map[uuid.UUID]bool, len(reminders))
	for _, r := range reminders {
		if !seen[r.VehicleID] {
			seen[r.VehicleID] = true
			vehicleIDs = append(vehicleIDs, r.VehicleID)
		}
	}
	vehicleNames := make(map[uuid.UUID]string, len(vehicleIDs))
	if vehicleMap, err := s.vehicleRepo.GetByIDs(ctx, vehicleIDs); err == nil {
		for id, v := range vehicleMap {
			vehicleNames[id] = v.Name
		}
	} else {
		s.logger.Warn("failed to batch fetch reminder vehicles", zap.Error(err))
	}

	result := make([]dto.ReminderResponse, len(reminders))
	for i, r := range reminders {
		result[i] = s.reminderToResponse(&r, vehicleNames[r.VehicleID])
	}
	return result, nil
}

// GetByID 获取提醒详情
func (s *ReminderService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.ReminderResponse, error) {
	reminder, err := s.reminderRepo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("reminder.not_found", "reminder not found")
		}
		return nil, apperror.ErrInternal("fetching reminder", err)
	}

	vehicleName := ""
	if v, err := s.vehicleRepo.GetByID(ctx, reminder.VehicleID); err == nil {
		vehicleName = v.Name
	}

	resp := s.reminderToResponse(reminder, vehicleName)
	return &resp, nil
}

// Update 更新提醒
func (s *ReminderService) Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdateReminderRequest) (*dto.ReminderResponse, error) {
	reminder, err := s.reminderRepo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("reminder.not_found", "reminder not found")
		}
		return nil, apperror.ErrInternal("fetching reminder", err)
	}

	// 部分更新
	if req.Category != nil {
		reminder.Category = model.MaintenanceCategory(*req.Category)
	}
	if req.Title != nil {
		reminder.Title = *req.Title
	}
	if req.Description != nil {
		reminder.Description = *req.Description
	}
	if req.Trigger != nil {
		reminder.Trigger = model.ReminderTrigger(*req.Trigger)
	}
	if req.MileageInterval != nil {
		reminder.MileageInterval = *req.MileageInterval
	}
	if req.TimeIntervalDays != nil {
		reminder.TimeIntervalDays = *req.TimeIntervalDays
	}
	if req.LastMileage != nil {
		reminder.LastMileage = *req.LastMileage
	}
	if req.LastDate != nil {
		if *req.LastDate == "" {
			reminder.LastDate = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.LastDate)
			if err != nil {
				return nil, apperror.ErrBadRequest("reminder.invalid_date", "invalid last date format")
			}
			reminder.LastDate = &t
		}
	}
	if req.IsEnabled != nil {
		reminder.IsEnabled = *req.IsEnabled
	}

	// 修改触发条件或间隔时，合并后的触发类型必须有对应的间隔
	if req.Trigger != nil || req.MileageInterval != nil || req.TimeIntervalDays != nil {
		if err := validateTriggerIntervals(reminder.Trigger, reminder.MileageInterval, reminder.TimeIntervalDays); err != nil {
			return nil, err
		}
	}

	// 重新计算下次触发
	s.calculateNext(reminder)

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		return nil, apperror.ErrInternal("updating reminder", err)
	}

	vehicleName := ""
	if v, err := s.vehicleRepo.GetByID(ctx, reminder.VehicleID); err == nil {
		vehicleName = v.Name
	}

	resp := s.reminderToResponse(reminder, vehicleName)
	return &resp, nil
}

// Delete 删除提醒
func (s *ReminderService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := s.reminderRepo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("reminder.not_found", "reminder not found")
		}
		return apperror.ErrInternal("fetching reminder", err)
	}

	if err := s.reminderRepo.Delete(ctx, id, userID); err != nil {
		return apperror.ErrInternal("deleting reminder", err)
	}
	return nil
}

// validateTriggerIntervals 校验触发类型所需的里程/时间间隔均大于 0
func validateTriggerIntervals(trigger model.ReminderTrigger, mileageInterval float64, timeIntervalDays int) error {
	if trigger == model.ReminderTriggerMileage || trigger == model.ReminderTriggerBoth {
		if mileageInterval <= 0 {
			return apperror.ErrBadRequest("reminder.mileage_required", "mileage interval is required")
		}
	}
	if trigger == model.ReminderTriggerTime || trigger == model.ReminderTriggerBoth {
		if timeIntervalDays <= 0 {
			return apperror.ErrBadRequest("reminder.time_required", "time interval is required")
		}
	}
	return nil
}

// calculateNext 根据上次保养基准计算下次触发里程/日期
func (s *ReminderService) calculateNext(r *model.Reminder) {
	prevMileage, prevDate := r.NextMileage, r.NextDate
	defer resetDueNotificationIfRescheduled(r, prevMileage, prevDate)

	// 按里程
	if r.Trigger == model.ReminderTriggerMileage || r.Trigger == model.ReminderTriggerBoth {
		if r.MileageInterval > 0 {
			r.NextMileage = r.LastMileage + r.MileageInterval
		}
	}

	// 按时间
	if r.Trigger == model.ReminderTriggerTime || r.Trigger == model.ReminderTriggerBoth {
		if r.TimeIntervalDays > 0 && r.LastDate != nil {
			next := r.LastDate.AddDate(0, 0, r.TimeIntervalDays)
			r.NextDate = &next
		} else if r.TimeIntervalDays > 0 && r.LastDate == nil {
			// 没有上次日期，以当前时间为基准
			now := time.Now()
			r.LastDate = &now
			next := now.AddDate(0, 0, r.TimeIntervalDays)
			r.NextDate = &next
		}
	}
}

// reminderToResponse 将 model 转为 DTO
// resetDueNotificationIfRescheduled 下次触发里程/日期发生变化（如完成保养、修改间隔）时开启新的通知周期
func resetDueNotificationIfRescheduled(r *model.Reminder, prevMileage float64, prevDate *time.Time) {
	sameDate := (prevDate == nil && r.NextDate == nil) ||
		(prevDate != nil && r.NextDate != nil && prevDate.Equal(*r.NextDate))
	if r.NextMileage != prevMileage || !sameDate {
		r.DueNotifiedAt = nil
	}
}

func (s *ReminderService) reminderToResponse(r *model.Reminder, vehicleName string) dto.ReminderResponse {
	resp := dto.ReminderResponse{
		ID:               r.ID.String(),
		VehicleID:        r.VehicleID.String(),
		VehicleName:      vehicleName,
		Type:             string(r.Type),
		Category:         string(r.Category),
		Title:            r.Title,
		Description:      r.Description,
		Trigger:          string(r.Trigger),
		MileageInterval:  r.MileageInterval,
		TimeIntervalDays: r.TimeIntervalDays,
		LastMileage:      r.LastMileage,
		LastDate:         r.LastDate,
		NextMileage:      r.NextMileage,
		NextDate:         r.NextDate,
		IsEnabled:        r.IsEnabled,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}

	// 判断是否过期
	now := time.Now()
	if r.IsEnabled {
		if r.NextDate != nil && now.After(*r.NextDate) {
			resp.IsOverdue = true
		}
		// 里程过期需要在前端根据当前里程判断
	}

	return resp
}
