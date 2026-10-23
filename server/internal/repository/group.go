package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gastrack/internal/model"
	"gastrack/internal/pkg/convert"
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

// Delete 删除群组（硬删除，级联删除成员关系与共享车辆关联）
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除共享车辆关联
		if err := tx.Unscoped().Where("group_id = ?", id).Delete(&model.SharedVehicle{}).Error; err != nil {
			return err
		}
		// 再删除成员关系
		if err := tx.Where("group_id = ?", id).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		// 再删除群组
		return tx.Unscoped().Where("id = ?", id).Delete(&model.Group{}).Error
	})
}

// ExistsByInviteCode 检查邀请码是否已存在
func (r *GroupRepository) ExistsByInviteCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM groups WHERE invite_code = ? LIMIT 1)", code,
	).Scan(&exists).Error
	return exists, err
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

// RemoveMember 移除群组成员（退出或被移除），同时撤销该成员在此群组中共享的车辆
func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().
			Where("group_id = ? AND shared_by = ?", groupID, userID).
			Delete(&model.SharedVehicle{}).Error; err != nil {
			return err
		}
		return tx.Where("group_id = ? AND user_id = ?", groupID, userID).
			Delete(&model.GroupMember{}).Error
	})
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

// ListMembershipsByUser 查询用户的所有群组成员身份（用于数据导出）
func (r *GroupRepository) ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// ListSharedVehiclesByUser 查询用户共享出去的车辆记录（用于数据导出）
func (r *GroupRepository) ListSharedVehiclesByUser(ctx context.Context, userID uuid.UUID) ([]model.SharedVehicle, error) {
	var svs []model.SharedVehicle
	err := r.db.WithContext(ctx).
		Where("shared_by = ?", userID).
		Order("created_at ASC").
		Find(&svs).Error
	return svs, err
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

// --- 群组车辆 ---

// GroupVehicleRow 群组内可见的车辆：已共享到群组的车辆，以及查看者自己名下（可能尚未共享）的车辆
type GroupVehicleRow struct {
	VehicleID   uuid.UUID `gorm:"column:vehicle_id"`
	VehicleName string    `gorm:"column:vehicle_name"`
	OwnerID     uuid.UUID `gorm:"column:owner_id"`
	VehicleType string    `gorm:"column:vehicle_type"`
	FuelType    string    `gorm:"column:fuel_type"`
	IsShared    bool      `gorm:"column:is_shared"`
}

// sharedInGroupSQL 车辆 v 已共享到群组 @gid，且共享发起人仍是群组成员
const sharedInGroupSQL = `EXISTS (
	SELECT 1 FROM shared_vehicles sv
	JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
	WHERE sv.group_id = @gid AND sv.vehicle_id = v.id AND sv.deleted_at IS NULL
)`

// ListGroupVehicles 列出群组统计可见的车辆（未归档、未删除）。
// 群组统计只基于已共享的车辆；查看者自己的未共享车辆也会列出（IsShared=false），便于在群组内开启共享。
func (r *GroupRepository) ListGroupVehicles(ctx context.Context, groupID, viewerID uuid.UUID) ([]GroupVehicleRow, error) {
	var rows []GroupVehicleRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT v.id AS vehicle_id, v.name AS vehicle_name, v.user_id AS owner_id, v.vehicle_type, v.fuel_type,
			`+sharedInGroupSQL+` AS is_shared
		FROM vehicles v
		WHERE v.deleted_at IS NULL AND v.is_archived = false
		  AND (`+sharedInGroupSQL+` OR v.user_id = @viewer)
		ORDER BY v.name ASC
	`, map[string]any{"gid": groupID, "viewer": viewerID}).Scan(&rows).Error
	return rows, err
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

// --- 车辆共享 ---

// CreateSharedVehicle 创建共享车辆关联
// 旧版本取消共享为软删除，残留行会触发 (group_id, vehicle_id) 唯一索引冲突，故先清理同一对的残留行
func (r *GroupRepository) CreateSharedVehicle(ctx context.Context, sv *model.SharedVehicle) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().
			Where("group_id = ? AND vehicle_id = ? AND deleted_at IS NOT NULL", sv.GroupID, sv.VehicleID).
			Delete(&model.SharedVehicle{}).Error; err != nil {
			return err
		}
		return tx.Create(sv).Error
	})
}

// sharerIsMemberSQL 共享发起人仍是群组成员（兼容旧版本成员离开后未撤销的共享记录）
const sharerIsMemberSQL = `EXISTS (
	SELECT 1 FROM group_members sgm
	WHERE sgm.group_id = shared_vehicles.group_id AND sgm.user_id = shared_vehicles.shared_by
)`

// DeleteSharedVehicle 删除共享车辆关联（硬删除：共享关系无需保留历史，
// 软删除会使唯一索引阻止再次共享，且容易被手写 SQL 遗漏过滤）
func (r *GroupRepository) DeleteSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("group_id = ? AND vehicle_id = ?", groupID, vehicleID).
		Delete(&model.SharedVehicle{}).Error
}

// ListSharedVehiclesByGroup 获取群组内所有共享车辆
func (r *GroupRepository) ListSharedVehiclesByGroup(ctx context.Context, groupID uuid.UUID) ([]model.SharedVehicle, error) {
	var svs []model.SharedVehicle
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Where(sharerIsMemberSQL).
		Order("created_at ASC").
		Find(&svs).Error
	return svs, err
}

// ExistsSharedVehicle 检查某车辆是否已在某群组中共享
func (r *GroupRepository) ExistsSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM shared_vehicles WHERE group_id = ? AND vehicle_id = ? AND deleted_at IS NULL LIMIT 1)",
		groupID, vehicleID,
	).Scan(&exists).Error
	return exists, err
}

// GetSharedVehicle 获取共享车辆记录
func (r *GroupRepository) GetSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) (*model.SharedVehicle, error) {
	var sv model.SharedVehicle
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND vehicle_id = ?", groupID, vehicleID).
		First(&sv).Error
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

// SharedVehicleWithGroup 共享车辆及其来源群组信息
type SharedVehicleWithGroup struct {
	VehicleID uuid.UUID `gorm:"column:vehicle_id"`
	GroupID   uuid.UUID `gorm:"column:group_id"`
	GroupName string    `gorm:"column:group_name"`
	SharedBy  uuid.UUID `gorm:"column:shared_by"`
}

// ListSharedVehiclesForUser 查询用户通过群组可访问的所有共享车辆（排除自己拥有的）
func (r *GroupRepository) ListSharedVehiclesForUser(ctx context.Context, userID uuid.UUID) ([]SharedVehicleWithGroup, error) {
	var results []SharedVehicleWithGroup

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			sv.vehicle_id,
			sv.group_id,
			g.name AS group_name,
			sv.shared_by
		FROM shared_vehicles sv
		JOIN groups g ON g.id = sv.group_id AND g.deleted_at IS NULL
		JOIN vehicles v ON v.id = sv.vehicle_id AND v.deleted_at IS NULL AND v.is_archived = false
		JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
		WHERE sv.deleted_at IS NULL
		  AND v.user_id != ?
		  AND EXISTS (
			SELECT 1 FROM group_members gm
			WHERE gm.group_id = sv.group_id AND gm.user_id = ?
		  )
		ORDER BY g.name ASC, sv.created_at ASC
	`, userID, userID).Scan(&results).Error

	return results, err
}

// IsVehicleSharedToUser 检查某车辆是否通过群组共享给了某用户
func (r *GroupRepository) IsVehicleSharedToUser(ctx context.Context, vehicleID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM shared_vehicles sv
			JOIN group_members gm ON gm.group_id = sv.group_id AND gm.user_id = ?
			JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
			JOIN groups g ON g.id = sv.group_id AND g.deleted_at IS NULL
			JOIN vehicles v ON v.id = sv.vehicle_id AND v.deleted_at IS NULL
			WHERE sv.vehicle_id = ? AND sv.deleted_at IS NULL
			LIMIT 1
		)
	`, userID, vehicleID).Scan(&exists).Error
	return exists, err
}

// --- 加油站推荐共享（仅统计已共享到群组的车辆） ---

var gallonToLiterSQL = fmt.Sprintf("%v", convert.GallonToLiter)

// stationPriceSQL 单价统一为每升（gal 已换算）或每 kWh
var stationPriceSQL = "CASE WHEN fr.fuel_unit = 'gal' THEN fr.unit_price / " + gallonToLiterSQL + " ELSE fr.unit_price END"

// stationPriceUnitSQL 单价单位：充电为 kWh，其余为 L
const stationPriceUnitSQL = "CASE WHEN fr.fuel_unit = 'kWh' THEN 'kWh' ELSE 'L' END"

// StationStatsRow 加油站聚合查询结果行（按 站点 × 币种 × 单价单位 分组）
type StationStatsRow struct {
	StationName  string  `gorm:"column:station_name"`
	PriceUnit    string  `gorm:"column:price_unit"`
	AvgUnitPrice float64 `gorm:"column:avg_unit_price"`
	VisitCount   int     `gorm:"column:visit_count"`
	LatestVisit  string  `gorm:"column:latest_visit"`
	CurrencyCode string  `gorm:"column:currency_code"`
}

// GetGroupStationStats 获取群组加油站聚合数据
func (r *GroupRepository) GetGroupStationStats(ctx context.Context, groupID uuid.UUID, months int, fuelGrade, sortBy string) ([]StationStatsRow, error) {
	var results []StationStatsRow

	cutoff := time.Now().AddDate(0, -months, 0)

	// 白名单映射排序列，避免 SQL 注入
	orderClauses := map[string]string{
		"avg_price":   "avg_unit_price ASC",
		"latest_date": "latest_visit DESC",
	}
	orderClause, ok := orderClauses[sortBy]
	if !ok {
		orderClause = "visit_count DESC"
	}

	query := `
		SELECT 
			fr.station_name,
			` + stationPriceUnitSQL + ` AS price_unit,
			AVG(` + stationPriceSQL + `) AS avg_unit_price,
			COUNT(*) AS visit_count,
			TO_CHAR(MAX(fr.refuel_date), 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS latest_visit,
			fr.currency_code
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
		JOIN shared_vehicles sv ON sv.vehicle_id = v.id AND sv.group_id = ? AND sv.deleted_at IS NULL
		JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
		WHERE fr.deleted_at IS NULL
		  AND fr.station_name IS NOT NULL 
		  AND fr.station_name != ''
		  AND fr.unit_price > 0
		  AND fr.refuel_date >= ?`

	args := []any{groupID, cutoff}

	if fuelGrade != "" {
		query += ` AND fr.fuel_grade = ?`
		args = append(args, fuelGrade)
	}

	query += `
		GROUP BY fr.station_name, fr.currency_code, price_unit
		ORDER BY ` + orderClause

	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	return results, err
}

// StationVisitorRow 加油站常客查询结果行
type StationVisitorRow struct {
	StationName string    `gorm:"column:station_name"`
	UserID      uuid.UUID `gorm:"column:user_id"`
	Count       int       `gorm:"column:count"`
}

// GetStationVisitors 获取加油站常客数据
func (r *GroupRepository) GetStationVisitors(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationVisitorRow, error) {
	var results []StationVisitorRow
	if len(stationNames) == 0 {
		return results, nil
	}

	cutoff := time.Now().AddDate(0, -months, 0)

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			fr.station_name,
			v.user_id,
			COUNT(*) AS count
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
		JOIN shared_vehicles sv ON sv.vehicle_id = v.id AND sv.group_id = ? AND sv.deleted_at IS NULL
		JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
		WHERE fr.deleted_at IS NULL AND fr.station_name IN ?
		  AND fr.refuel_date >= ?
		GROUP BY fr.station_name, v.user_id
		ORDER BY fr.station_name ASC, count DESC
	`, groupID, stationNames, cutoff).Scan(&results).Error

	return results, err
}

// StationLatestPriceRow 加油站最新油价查询结果行（按 站点 × 币种 × 单价单位 分区）
type StationLatestPriceRow struct {
	StationName  string  `gorm:"column:station_name"`
	CurrencyCode string  `gorm:"column:currency_code"`
	PriceUnit    string  `gorm:"column:price_unit"`
	UnitPrice    float64 `gorm:"column:unit_price"`
	RowNum       int     `gorm:"column:row_num"`
}

// GetStationLatestPrices 获取加油站最新两次油价（用于计算趋势）
func (r *GroupRepository) GetStationLatestPrices(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationLatestPriceRow, error) {
	var results []StationLatestPriceRow
	if len(stationNames) == 0 {
		return results, nil
	}

	cutoff := time.Now().AddDate(0, -months, 0)

	err := r.db.WithContext(ctx).Raw(`
		SELECT station_name, currency_code, price_unit, unit_price, row_num FROM (
			SELECT 
				fr.station_name,
				fr.currency_code,
				`+stationPriceUnitSQL+` AS price_unit,
				`+stationPriceSQL+` AS unit_price,
				ROW_NUMBER() OVER (
					PARTITION BY fr.station_name, fr.currency_code, `+stationPriceUnitSQL+`
					ORDER BY fr.refuel_date DESC
				) AS row_num
			FROM fuel_records fr
			JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
			JOIN shared_vehicles sv ON sv.vehicle_id = v.id AND sv.group_id = ? AND sv.deleted_at IS NULL
		JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
			WHERE fr.deleted_at IS NULL AND fr.station_name IN ?
			  AND fr.refuel_date >= ?
			  AND fr.unit_price > 0
		) sub
		WHERE row_num <= 2
		ORDER BY station_name ASC, row_num ASC
	`, groupID, stationNames, cutoff).Scan(&results).Error

	return results, err
}

// StationFuelGradeRow 加油站燃油标号查询结果行
type StationFuelGradeRow struct {
	StationName string `gorm:"column:station_name"`
	FuelGrade   string `gorm:"column:fuel_grade"`
}

// GetStationFuelGrades 获取加油站出现过的燃油标号
func (r *GroupRepository) GetStationFuelGrades(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationFuelGradeRow, error) {
	var results []StationFuelGradeRow
	if len(stationNames) == 0 {
		return results, nil
	}

	cutoff := time.Now().AddDate(0, -months, 0)

	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT fr.station_name, fr.fuel_grade
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
		JOIN shared_vehicles sv ON sv.vehicle_id = v.id AND sv.group_id = ? AND sv.deleted_at IS NULL
		JOIN group_members sgm ON sgm.group_id = sv.group_id AND sgm.user_id = sv.shared_by
		WHERE fr.deleted_at IS NULL AND fr.station_name IN ?
		  AND fr.refuel_date >= ?
		  AND fr.fuel_grade IS NOT NULL
		  AND fr.fuel_grade != ''
		ORDER BY fr.station_name ASC
	`, groupID, stationNames, cutoff).Scan(&results).Error

	return results, err
}
