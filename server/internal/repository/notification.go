package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// NotificationRepository 通知数据访问
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建实例
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// ListByUser 查询用户的通知列表（最近 50 条）
func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	if limit <= 0 {
		limit = 50
	}
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// ListAllByUser 查询用户的所有通知（用于数据导出，不限条数）
func (r *NotificationRepository) ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// CountUnread 统计用户未读通知数
func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记指定通知为已读
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// MarkAllAsRead 标记用户所有通知为已读
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// Delete 删除通知
func (r *NotificationRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Notification{}).Error
}
