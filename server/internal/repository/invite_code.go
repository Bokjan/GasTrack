package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gastrack/internal/model"
)

// InviteCodeRepository 邀请码数据访问
type InviteCodeRepository struct {
	db *gorm.DB
}

// NewInviteCodeRepository 创建 InviteCodeRepository 实例
func NewInviteCodeRepository(db *gorm.DB) *InviteCodeRepository {
	return &InviteCodeRepository{db: db}
}

// Create 创建邀请码
func (r *InviteCodeRepository) Create(ctx context.Context, invite *model.InviteCode) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

// GetByID 根据 ID 查询邀请码
func (r *InviteCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InviteCode, error) {
	var invite model.InviteCode
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// GetByCode 根据邀请码字符串查询
func (r *InviteCodeRepository) GetByCode(ctx context.Context, code string) (*model.InviteCode, error) {
	var invite model.InviteCode
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// ListByCreator 查询某用户创建的邀请码列表
func (r *InviteCodeRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID) ([]model.InviteCode, error) {
	var invites []model.InviteCode
	err := r.db.WithContext(ctx).
		Where("created_by = ?", creatorID).
		Order("created_at DESC").
		Find(&invites).Error
	return invites, err
}

// Update 更新邀请码
func (r *InviteCodeRepository) Update(ctx context.Context, invite *model.InviteCode) error {
	return r.db.WithContext(ctx).Save(invite).Error
}

// Delete 删除邀请码（硬删除）
func (r *InviteCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&model.InviteCode{}).Error
}

// ConsumeByCode 原子性消费邀请码（SELECT FOR UPDATE + UPDATE use_count）
// 返回消费后的邀请码，如果邀请码无效则返回错误
func (r *InviteCodeRepository) ConsumeByCode(ctx context.Context, code string, usedByID uuid.UUID) (*model.InviteCode, error) {
	var invite model.InviteCode

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE 锁定行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", code).
			First(&invite).Error; err != nil {
			return err
		}

		// 校验有效性
		if !invite.IsValid() {
			return fmt.Errorf("invite code is not valid")
		}

		// 更新使用计数和使用者
		invite.UseCount++
		invite.UsedBy = &usedByID

		return tx.Model(&invite).Updates(map[string]any{
			"use_count": invite.UseCount,
			"used_by":   usedByID,
		}).Error
	})

	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// ExistsByCode 检查邀请码是否存在
func (r *InviteCodeRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM invite_codes WHERE code = ? LIMIT 1)", code,
	).Scan(&exists).Error
	return exists, err
}
