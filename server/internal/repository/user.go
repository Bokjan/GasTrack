// Package repository 提供数据访问层，封装数据库操作。
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// UserRepository 用户数据访问
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据 ID 查询用户
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱查询用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields 更新用户指定字段
func (r *UserRepository) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login_at", now).Error
}

// Delete 软删除用户
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

// ExistsByEmail 检查邮箱是否已注册
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1)", email,
	).Scan(&exists).Error
	return exists, err
}

// GetByIDs 根据多个 ID 批量查询用户，返回 ID→User 映射
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*model.User, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*model.User), nil
	}

	var users []model.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*model.User, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result, nil
}
