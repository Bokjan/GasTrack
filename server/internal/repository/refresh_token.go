package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// RefreshTokenRepository 刷新令牌数据访问
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository 创建实例
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create 创建刷新令牌
func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByTokenHash 根据 token hash 查询
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND expires_at > ?", hash, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteByID 删除指定 token
func (r *RefreshTokenRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.RefreshToken{}).Error
}

// DeleteByUserID 删除用户的所有 refresh token（登出所有设备）
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

// DeleteExpired 清除过期 token（定时任务调用）
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{}).Error
}
