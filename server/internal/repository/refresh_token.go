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

// ConsumeByTokenHash 原子性地查找并删除指定 hash 的 refresh token（用于 Token Rotation）。
// 使用 SELECT ... FOR UPDATE 加行锁，确保并发请求只有一个能成功消费该 token。
// 返回被删除的 token 记录；如果 token 不存在或已被其他请求消费，返回 gorm.ErrRecordNotFound。
func (r *RefreshTokenRepository) ConsumeByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT ... FOR UPDATE：锁定该行，防止并发消费
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("token_hash = ? AND expires_at > ?", hash, time.Now()).
			First(&token).Error; err != nil {
			return err
		}
		// 删除该 token（一次性使用）
		if err := tx.Where("id = ?", token.ID).Delete(&model.RefreshToken{}).Error; err != nil {
			return err
		}
		return nil
	})
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
