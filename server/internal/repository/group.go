package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gastrack/internal/model"
)

// GroupRepository 群组数据访问
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository 创建 GroupRepository 实例
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// DB 返回底层 *gorm.DB 实例（用于 Service 层执行事务）
func (r *GroupRepository) DB() *gorm.DB {
	return r.db
}

// --- 群组 CRUD ---

// Create 创建群组
func (r *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetByID 根据 ID 查询群组
func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByIDWithMembers 根据 ID 查询群组（包含成员列表）
func (r *GroupRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).
		Preload("Members").
		Where("id = ?", id).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByInviteCode 根据邀请码查询群组
func (r *GroupRepository) GetByInviteCode(ctx context.Context, code string) (*model.Group, error) {
	var group model.Group
	err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Update 更新群组
func (r *GroupRepository) Update(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// Delete 删除群组（硬删除，级联删除成员关系）
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除成员关系
		if err := tx.Where("group_id = ?", id).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		// 再删除群组
		return tx.Unscoped().Where("id = ?", id).Delete(&model.Group{}).Error
	})
}

// ExistsByInviteCode 检查邀请码是否已存在
func (r *GroupRepository) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Group{}).Where("invite_code = ?", code).Count(&count).Error
	return count > 0, err
}

// --- 群组成员管理 ---

// AddMember 添加群组成员
func (r *GroupRepository) AddMember(ctx context.Context, member *model.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember 查询群组成员
func (r *GroupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*model.GroupMember, error) {
	var member model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// ListMembers 查询群组的所有成员
func (r *GroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// UpdateMemberRole 更新成员角色
func (r *GroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role model.GroupRole) error {
	return r.db.WithContext(ctx).
		Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

// RemoveMember 移除群组成员
func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{}).Error
}

// CountMembers 统计群组成员数量
func (r *GroupRepository) CountMembers(ctx context.Context, groupID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

// ListGroupsByUser 查询用户所在的所有群组
func (r *GroupRepository) ListGroupsByUser(ctx context.Context, userID uuid.UUID) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.WithContext(ctx).
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error
	return groups, err
}

// JoinGroupByInviteCode 通过邀请码加入群组（SELECT FOR UPDATE 保证并发安全）
func (r *GroupRepository) JoinGroupByInviteCode(ctx context.Context, code string, userID uuid.UUID) (*model.Group, error) {
	var group model.Group

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE 锁定群组行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("invite_code = ?", code).
			First(&group).Error; err != nil {
			return err
		}

		// 检查是否已是成员
		var count int64
		if err := tx.Model(&model.GroupMember{}).
			Where("group_id = ? AND user_id = ?", group.ID, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrAlreadyMember
		}

		// 检查成员上限
		var memberCount int64
		if err := tx.Model(&model.GroupMember{}).
			Where("group_id = ?", group.ID).
			Count(&memberCount).Error; err != nil {
			return err
		}
		if int(memberCount) >= group.MaxMembers {
			return ErrGroupFull
		}

		// 添加成员
		member := &model.GroupMember{
			GroupID: group.ID,
			UserID:  userID,
			Role:    model.GroupRoleMember,
		}
		return tx.Create(member).Error
	})

	if err != nil {
		return nil, err
	}
	return &group, nil
}

// --- 群组车辆数据汇总 ---

// VehicleSummaryRow 车辆汇总查询结果行
type VehicleSummaryRow struct {
	VehicleID   uuid.UUID `gorm:"column:vehicle_id"`
	VehicleName string    `gorm:"column:vehicle_name"`
	OwnerID     uuid.UUID `gorm:"column:owner_id"`
	VehicleType string    `gorm:"column:vehicle_type"`
	FuelType    string    `gorm:"column:fuel_type"`
	Records     int64     `gorm:"column:total_records"`
	TotalCost   float64   `gorm:"column:total_cost"`
	TotalFuel   float64   `gorm:"column:total_fuel"`
	AvgEff      float64   `gorm:"column:avg_efficiency"`
}

// GetGroupVehicleSummary 获取群组内所有成员的车辆数据汇总
func (r *GroupRepository) GetGroupVehicleSummary(ctx context.Context, groupID uuid.UUID) ([]VehicleSummaryRow, error) {
	var results []VehicleSummaryRow

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			v.id AS vehicle_id,
			v.name AS vehicle_name,
			v.user_id AS owner_id,
			v.vehicle_type,
			v.fuel_type,
			COUNT(fr.id) AS total_records,
			COALESCE(SUM(fr.total_cost), 0) AS total_cost,
			COALESCE(SUM(fr.fuel_amount), 0) AS total_fuel,
			CASE WHEN COUNT(fr.id) > 0 
				THEN COALESCE(AVG(fr.fuel_efficiency), 0)
				ELSE 0 
			END AS avg_efficiency
		FROM vehicles v
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		LEFT JOIN fuel_records fr ON fr.vehicle_id = v.id
		WHERE v.deleted_at IS NULL AND v.is_archived = false
		GROUP BY v.id, v.name, v.user_id, v.vehicle_type, v.fuel_type
		ORDER BY v.name ASC
	`, groupID).Scan(&results).Error

	return results, err
}

// --- 自定义错误 ---

// 群组相关错误
var (
	ErrAlreadyMember = &groupError{msg: "already a member of this group"}
	ErrGroupFull     = &groupError{msg: "group has reached maximum members"}
)

type groupError struct {
	msg string
}

func (e *groupError) Error() string {
	return e.msg
}
