package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// ReminderRepository 提醒数据访问
type ReminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository 创建实例
func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

// Create 创建提醒
func (r *ReminderRepository) Create(ctx context.Context, reminder *model.Reminder) error {
	return r.db.WithContext(ctx).Create(reminder).Error
}

// GetByID 根据 ID 查询提醒
func (r *ReminderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reminder, error) {
	var reminder model.Reminder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&reminder).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

// GetByIDAndUser 根据 ID 和用户 ID 查询提醒
func (r *ReminderRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Reminder, error) {
	var reminder model.Reminder
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&reminder).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

// ListByUser 查询用户所有提醒
func (r *ReminderRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_enabled DESC, created_at DESC").
		Find(&reminders).Error
	return reminders, err
}

// ListByVehicle 查询车辆的所有提醒
func (r *ReminderRepository) ListByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Order("is_enabled DESC, created_at DESC").
		Find(&reminders).Error
	return reminders, err
}

// ListEnabledByVehicle 查询车辆已启用的提醒
func (r *ReminderRepository) ListEnabledByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND is_enabled = ?", vehicleID, true).
		Find(&reminders).Error
	return reminders, err
}

// ListDueByDate 查询按时间触发且已到期、本周期尚未通知的提醒（车辆未删除）
func (r *ReminderRepository) ListDueByDate(ctx context.Context, now time.Time) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.WithContext(ctx).
		Joins("JOIN vehicles v ON v.id = reminders.vehicle_id AND v.deleted_at IS NULL").
		Where("reminders.is_enabled = ? AND reminders.trigger IN ?", true,
			[]model.ReminderTrigger{model.ReminderTriggerTime, model.ReminderTriggerBoth}).
		Where("reminders.next_date IS NOT NULL AND reminders.next_date <= ?", now).
		Where("reminders.due_notified_at IS NULL").
		Find(&reminders).Error
	return reminders, err
}

// MarkDueNotified 原子地标记提醒本周期已通知；返回 false 表示已被其他流程标记过（不应重复通知）
func (r *ReminderRepository) MarkDueNotified(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.Reminder{}).
		Where("id = ? AND due_notified_at IS NULL", id).
		UpdateColumn("due_notified_at", at)
	return res.RowsAffected == 1, res.Error
}

// Update 更新提醒
func (r *ReminderRepository) Update(ctx context.Context, reminder *model.Reminder) error {
	return r.db.WithContext(ctx).Save(reminder).Error
}

// Delete 删除提醒（硬删除）
func (r *ReminderRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Reminder{}).Error
}
