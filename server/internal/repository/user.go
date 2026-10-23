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

// GetByEmail 根据邮箱查询用户（大小写不敏感；email 参数应已规范化为小写）
// 历史数据中可能存在仅大小写不同的多个账号，按注册时间取最早的一个
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("LOWER(email) = ?", email).Order("created_at ASC").First(&user).Error
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

// PurgeAccount 在一个事务中彻底删除用户及其数据（硬删除，含历史软删除记录）：
//   - 本人车辆及其加油/开销记录、提醒、相关通知、共享关系
//   - 本人在他人车辆上录入的加油/开销记录改挂到车主名下（保持车主车辆里程与油耗链完整）
//   - 本人创建的提醒（含他人车辆上的）；引用这些提醒的开销记录解除关联
//   - 本人为群主的群组：仍有其他成员时随机转让给其中一位，否则解散；退出其他群组并撤销共享
//   - 本人创建的邀请码；本人使用过的邀请码清空 used_by
//   - refresh token、通知、用户本身
func (r *UserRepository) PurgeAccount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Session 使后续多条链式调用互不共享条件；Unscoped 使删除/查询覆盖历史软删除记录
		tx = tx.Unscoped().Session(&gorm.Session{})
		ownVehicles := "SELECT id FROM vehicles WHERE user_id = @uid"
		ownedGroups := "SELECT id FROM groups WHERE owner_id = @uid"
		reminders := "SELECT id FROM reminders WHERE user_id = @uid OR vehicle_id IN (" + ownVehicles + ")"
		args := map[string]any{"uid": id}

		// 群主转让：每个本人为群主的群组随机选一名其他成员接任；没有其他成员的群组保留给下面的解散步骤
		var heirs []struct {
			GroupID uuid.UUID
			UserID  uuid.UUID
		}
		if err := tx.Raw(`
			SELECT DISTINCT ON (gm.group_id) gm.group_id, gm.user_id
			FROM group_members gm
			JOIN groups g ON g.id = gm.group_id
			WHERE g.owner_id = @uid AND gm.user_id <> @uid
			ORDER BY gm.group_id, random()`, args).Scan(&heirs).Error; err != nil {
			return err
		}
		for _, h := range heirs {
			if err := tx.Model(&model.Group{}).Where("id = ?", h.GroupID).Update("owner_id", h.UserID).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.GroupMember{}).
				Where("group_id = ? AND user_id = ?", h.GroupID, h.UserID).
				Update("role", model.GroupRoleOwner).Error; err != nil {
				return err
			}
		}

		// 以下 ownedGroups 只剩没有其他成员、需要解散的群组
		steps := []string{
			// 他人车辆上由本人录入的记录改挂到车主名下
			`UPDATE fuel_records fr SET user_id = v.user_id FROM vehicles v
				WHERE fr.vehicle_id = v.id AND fr.user_id = @uid AND v.user_id <> @uid`,
			`UPDATE expense_records er SET user_id = v.user_id FROM vehicles v
				WHERE er.vehicle_id = v.id AND er.user_id = @uid AND v.user_id <> @uid`,
			// 提醒及其引用
			"UPDATE expense_records SET reminder_id = NULL WHERE reminder_id IN (" + reminders + ")",
			"DELETE FROM notifications WHERE user_id = @uid OR vehicle_id IN (" + ownVehicles + ") OR reminder_id IN (" + reminders + ")",
			"DELETE FROM reminders WHERE id IN (" + reminders + ")",
			// 本人车辆的记录与共享
			"DELETE FROM expense_records WHERE vehicle_id IN (" + ownVehicles + ")",
			"DELETE FROM fuel_records WHERE vehicle_id IN (" + ownVehicles + ")",
			"DELETE FROM shared_vehicles WHERE vehicle_id IN (" + ownVehicles + ") OR shared_by = @uid OR group_id IN (" + ownedGroups + ")",
			// 群组：解散无人接任的群组，退出其他群组
			"DELETE FROM group_members WHERE user_id = @uid OR group_id IN (" + ownedGroups + ")",
			"DELETE FROM groups WHERE owner_id = @uid",
			"DELETE FROM vehicles WHERE user_id = @uid",
			// 邀请码与令牌
			"UPDATE invite_codes SET used_by = NULL WHERE used_by = @uid",
			"DELETE FROM invite_codes WHERE created_by = @uid",
			"DELETE FROM refresh_tokens WHERE user_id = @uid",
			"DELETE FROM users WHERE id = @uid",
		}
		for _, sql := range steps {
			if err := tx.Exec(sql, args).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListSoftDeletedIDs 列出旧版本软删除（未彻底清除）的用户 ID；email 非空时只匹配该邮箱（大小写不敏感）
func (r *UserRepository) ListSoftDeletedIDs(ctx context.Context, email string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	q := r.db.WithContext(ctx).Unscoped().Model(&model.User{}).Where("deleted_at IS NOT NULL")
	if email != "" {
		q = q.Where("LOWER(email) = ?", email)
	}
	err := q.Pluck("id", &ids).Error
	return ids, err
}

// ExistsByEmail 检查邮箱是否已注册
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = ? AND deleted_at IS NULL LIMIT 1)", email,
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
