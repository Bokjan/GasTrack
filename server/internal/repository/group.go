package repository

import (
	"context"
	"time"

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

// --- 群组车辆数据汇总 ---

// VehicleSummaryRow 车辆汇总查询结果行
type VehicleSummaryRow struct {
	VehicleID    uuid.UUID `gorm:"column:vehicle_id"`
	VehicleName  string    `gorm:"column:vehicle_name"`
	OwnerID      uuid.UUID `gorm:"column:owner_id"`
	VehicleType  string    `gorm:"column:vehicle_type"`
	FuelType     string    `gorm:"column:fuel_type"`
	CurrencyCode string    `gorm:"column:currency_code"`
	Records      int64     `gorm:"column:total_records"`
	TotalCost    float64   `gorm:"column:total_cost"`
	TotalFuel    float64   `gorm:"column:total_fuel"`
	AvgEff       float64   `gorm:"column:avg_efficiency"`
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
			COALESCE(top_cur.currency_code, u.currency_code) AS currency_code,
			COUNT(fr.id) AS total_records,
			COALESCE(SUM(fr.total_cost), 0) AS total_cost,
			COALESCE(SUM(fr.fuel_amount), 0) AS total_fuel,
			CASE WHEN COUNT(fr.id) > 0 
				THEN COALESCE(AVG(fr.fuel_efficiency), 0)
				ELSE 0 
			END AS avg_efficiency
		FROM vehicles v
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		JOIN users u ON u.id = v.user_id
		LEFT JOIN fuel_records fr ON fr.vehicle_id = v.id
		LEFT JOIN LATERAL (
			SELECT fr2.currency_code
			FROM fuel_records fr2
			WHERE fr2.vehicle_id = v.id
			GROUP BY fr2.currency_code
			ORDER BY COUNT(*) DESC
			LIMIT 1
		) top_cur ON true
		WHERE v.deleted_at IS NULL AND v.is_archived = false
		GROUP BY v.id, v.name, v.user_id, v.vehicle_type, v.fuel_type, u.currency_code, top_cur.currency_code
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

// --- 车辆共享 ---

// CreateSharedVehicle 创建共享车辆关联
func (r *GroupRepository) CreateSharedVehicle(ctx context.Context, sv *model.SharedVehicle) error {
	return r.db.WithContext(ctx).Create(sv).Error
}

// DeleteSharedVehicle 删除共享车辆关联
func (r *GroupRepository) DeleteSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ? AND vehicle_id = ?", groupID, vehicleID).
		Delete(&model.SharedVehicle{}).Error
}

// ListSharedVehiclesByGroup 获取群组内所有共享车辆
func (r *GroupRepository) ListSharedVehiclesByGroup(ctx context.Context, groupID uuid.UUID) ([]model.SharedVehicle, error) {
	var svs []model.SharedVehicle
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at ASC").
		Find(&svs).Error
	return svs, err
}

// ExistsSharedVehicle 检查某车辆是否已在某群组中共享
func (r *GroupRepository) ExistsSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM shared_vehicles WHERE group_id = ? AND vehicle_id = ? LIMIT 1)",
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
		JOIN groups g ON g.id = sv.group_id
		JOIN vehicles v ON v.id = sv.vehicle_id AND v.deleted_at IS NULL AND v.is_archived = false
		WHERE v.user_id != ?
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
			WHERE sv.vehicle_id = ?
			LIMIT 1
		)
	`, userID, vehicleID).Scan(&exists).Error
	return exists, err
}

// --- 排行榜 ---

// LeaderboardRow 排行榜查询结果行
type LeaderboardRow struct {
	VehicleID     uuid.UUID `gorm:"column:vehicle_id"`
	VehicleName   string    `gorm:"column:vehicle_name"`
	UserID        uuid.UUID `gorm:"column:user_id"`
	RecordCount   int       `gorm:"column:record_count"`
	AvgEfficiency float64   `gorm:"column:avg_efficiency"`
	TotalCost     float64   `gorm:"column:total_cost"`
	TotalDistance float64   `gorm:"column:total_distance"`
}

// GetLeaderboard 获取群组排行榜数据
func (r *GroupRepository) GetLeaderboard(ctx context.Context, groupID uuid.UUID, metric string, startDate, endDate time.Time) ([]LeaderboardRow, error) {
	var results []LeaderboardRow

	// 白名单映射排序列，避免 SQL 注入
	orderClauses := map[string]string{
		"cost":      "total_cost DESC",
		"distance":  "total_distance DESC",
		"frequency": "record_count DESC",
	}
	orderClause, ok := orderClauses[metric]
	if !ok {
		orderClause = "avg_efficiency ASC" // 默认油耗排行，越低越好
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			v.id AS vehicle_id,
			v.name AS vehicle_name,
			v.user_id,
			COUNT(fr.id) AS record_count,
			COALESCE(AVG(NULLIF(fr.fuel_efficiency, 0)), 0) AS avg_efficiency,
			COALESCE(SUM(fr.total_cost), 0) AS total_cost,
			COALESCE(SUM(fr.trip_distance), 0) AS total_distance
		FROM vehicles v
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		JOIN fuel_records fr ON fr.vehicle_id = v.id
		WHERE v.deleted_at IS NULL 
		  AND v.is_archived = false
		  AND fr.refuel_date >= ?
		  AND fr.refuel_date < ?
		  AND fr.fuel_efficiency > 0
		GROUP BY v.id, v.name, v.user_id
		HAVING COUNT(fr.id) >= 2
		ORDER BY `+orderClause, groupID, startDate, endDate).Scan(&results).Error

	return results, err
}

// --- 费用统计看板 ---

// GroupExpenseRow 群组费用按月/年聚合行
type GroupExpenseRow struct {
	PeriodLabel   string    `gorm:"column:period_label"`
	UserID        uuid.UUID `gorm:"column:user_id"`
	CurrencyCode  string    `gorm:"column:currency_code"`
	Cost          float64   `gorm:"column:cost"`
	Fuel          float64   `gorm:"column:fuel"`
	Distance      float64   `gorm:"column:distance"`
	AvgEfficiency float64   `gorm:"column:avg_eff"`
	Records       int       `gorm:"column:records"`
}

// GetGroupExpenseByMonth 按月聚合群组费用
func (r *GroupRepository) GetGroupExpenseByMonth(ctx context.Context, groupID uuid.UUID, year int) ([]GroupExpenseRow, error) {
	var results []GroupExpenseRow

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			TO_CHAR(fr.refuel_date, 'YYYY-MM') AS period_label,
			gm.user_id,
			fr.currency_code,
			COALESCE(SUM(fr.total_cost), 0) AS cost,
			COALESCE(SUM(fr.fuel_amount), 0) AS fuel,
			COALESCE(SUM(fr.trip_distance), 0) AS distance,
			COALESCE(AVG(NULLIF(fr.fuel_efficiency, 0)), 0) AS avg_eff,
			COUNT(fr.id) AS records
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL AND v.is_archived = false
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		WHERE fr.refuel_date >= ? AND fr.refuel_date < ?
		GROUP BY period_label, gm.user_id, fr.currency_code
		ORDER BY period_label ASC, gm.user_id ASC
	`, groupID, startDate, endDate).Scan(&results).Error

	return results, err
}

// GetGroupExpenseByYear 按年聚合群组费用
func (r *GroupRepository) GetGroupExpenseByYear(ctx context.Context, groupID uuid.UUID) ([]GroupExpenseRow, error) {
	var results []GroupExpenseRow

	// 限制最近 10 年的数据，避免全表扫描
	cutoff := time.Now().AddDate(-10, 0, 0)

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			TO_CHAR(fr.refuel_date, 'YYYY') AS period_label,
			gm.user_id,
			fr.currency_code,
			COALESCE(SUM(fr.total_cost), 0) AS cost,
			COALESCE(SUM(fr.fuel_amount), 0) AS fuel,
			COALESCE(SUM(fr.trip_distance), 0) AS distance,
			COALESCE(AVG(NULLIF(fr.fuel_efficiency, 0)), 0) AS avg_eff,
			COUNT(fr.id) AS records
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL AND v.is_archived = false
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		WHERE fr.refuel_date >= ?
		GROUP BY period_label, gm.user_id, fr.currency_code
		ORDER BY period_label ASC, gm.user_id ASC
	`, groupID, cutoff).Scan(&results).Error

	return results, err
}

// --- 加油站推荐共享 ---

// StationStatsRow 加油站聚合查询结果行
type StationStatsRow struct {
	StationName  string  `gorm:"column:station_name"`
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
			AVG(fr.unit_price) AS avg_unit_price,
			COUNT(*) AS visit_count,
			TO_CHAR(MAX(fr.refuel_date), 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS latest_visit,
			fr.currency_code
		FROM fuel_records fr
		JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		WHERE fr.station_name IS NOT NULL 
		  AND fr.station_name != ''
		  AND fr.unit_price > 0
		  AND fr.refuel_date >= ?`

	args := []any{groupID, cutoff}

	if fuelGrade != "" {
		query += ` AND fr.fuel_grade = ?`
		args = append(args, fuelGrade)
	}

	query += `
		GROUP BY fr.station_name, fr.currency_code
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
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		WHERE fr.station_name IN ?
		  AND fr.refuel_date >= ?
		GROUP BY fr.station_name, v.user_id
		ORDER BY fr.station_name ASC, count DESC
	`, groupID, stationNames, cutoff).Scan(&results).Error

	return results, err
}

// StationLatestPriceRow 加油站最新油价查询结果行
type StationLatestPriceRow struct {
	StationName string  `gorm:"column:station_name"`
	UnitPrice   float64 `gorm:"column:unit_price"`
	RowNum      int     `gorm:"column:row_num"`
}

// GetStationLatestPrices 获取加油站最新两次油价（用于计算趋势）
func (r *GroupRepository) GetStationLatestPrices(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationLatestPriceRow, error) {
	var results []StationLatestPriceRow
	if len(stationNames) == 0 {
		return results, nil
	}

	cutoff := time.Now().AddDate(0, -months, 0)

	err := r.db.WithContext(ctx).Raw(`
		SELECT station_name, unit_price, row_num FROM (
			SELECT 
				fr.station_name,
				fr.unit_price,
				ROW_NUMBER() OVER (PARTITION BY fr.station_name ORDER BY fr.refuel_date DESC) AS row_num
			FROM fuel_records fr
			JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
			JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
			WHERE fr.station_name IN ?
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
		JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = ?
		WHERE fr.station_name IN ?
		  AND fr.refuel_date >= ?
		  AND fr.fuel_grade IS NOT NULL
		  AND fr.fuel_grade != ''
		ORDER BY fr.station_name ASC
	`, groupID, stationNames, cutoff).Scan(&results).Error

	return results, err
}
