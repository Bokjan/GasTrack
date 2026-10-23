package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gastrack/internal/model"
)

// FuelRecordRepository 加油记录数据访问
type FuelRecordRepository struct {
	db *gorm.DB
}

// NewFuelRecordRepository 创建实例
func NewFuelRecordRepository(db *gorm.DB) *FuelRecordRepository {
	return &FuelRecordRepository{db: db}
}

// Create 创建加油记录
func (r *FuelRecordRepository) Create(ctx context.Context, record *model.FuelRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByID 根据 ID 查询加油记录
func (r *FuelRecordRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FuelRecord, error) {
	var record model.FuelRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByIDAndVehicle 根据记录 ID 和车辆 ID 查询
func (r *FuelRecordRepository) GetByIDAndVehicle(ctx context.Context, id, vehicleID uuid.UUID) (*model.FuelRecord, error) {
	var record model.FuelRecord
	err := r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", id, vehicleID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByVehicle 查询车辆的加油记录（分页）
func (r *FuelRecordRepository) ListByVehicle(ctx context.Context, vehicleID uuid.UUID, page, pageSize int) ([]model.FuelRecord, int64, error) {
	var records []model.FuelRecord
	var total int64

	// 使用独立查询链分别执行 COUNT 和分页查询，避免共享 *gorm.DB 状态导致结果异常
	if err := r.db.WithContext(ctx).Model(&model.FuelRecord{}).Where("vehicle_id = ?", vehicleID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).
		Order("refuel_date DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

// RecomputeDerived 重算某车辆所有记录的派生字段（行驶距离、油耗、油耗区间里程）。
// 在事务内锁定车辆行，串行化同一车辆的并发重算；按时间升序加载全部记录交给 compute，
// compute 原地修改派生字段并返回发生变化的下标，仅这些记录会被写回（不更新 updated_at）。
// 返回重算后的全部记录（升序）。
func (r *FuelRecordRepository) RecomputeDerived(ctx context.Context, vehicleID uuid.UUID, compute func(records []model.FuelRecord) []int) ([]model.FuelRecord, error) {
	var records []model.FuelRecord

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 锁定车辆行（含已软删除车辆）
		var lockedIDs []uuid.UUID
		if err := tx.Unscoped().Model(&model.Vehicle{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", vehicleID).
			Pluck("id", &lockedIDs).Error; err != nil {
			return err
		}

		if err := tx.Where("vehicle_id = ?", vehicleID).
			Order("refuel_date ASC, created_at ASC, id ASC").
			Find(&records).Error; err != nil {
			return err
		}

		for _, i := range compute(records) {
			if err := tx.Model(&model.FuelRecord{}).
				Where("id = ?", records[i].ID).
				UpdateColumns(map[string]any{
					"trip_distance":       records[i].TripDistance,
					"fuel_efficiency":     records[i].FuelEfficiency,
					"efficiency_distance": records[i].EfficiencyDistance,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// Update 更新记录
func (r *FuelRecordRepository) Update(ctx context.Context, record *model.FuelRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// Delete 删除记录
func (r *FuelRecordRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", id, vehicleID).Delete(&model.FuelRecord{}).Error
}

// GetDistinctStationNames 获取某车辆（或用户所有车辆）去重的加油站名列表（按使用频次降序）
func (r *FuelRecordRepository) GetDistinctStationNames(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit int) ([]string, error) {
	var names []string
	query := r.db.WithContext(ctx).Model(&model.FuelRecord{}).
		Select("station_name").
		Where("user_id = ? AND station_name != ''", userID).
		Group("station_name").
		Order("COUNT(*) DESC")

	if vehicleID != nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Pluck("station_name", &names).Error
	return names, err
}

// ListAllByUser 查询与用户相关的所有加油记录（用于数据导出，按时间升序）：
// 用户本人录入的记录，以及群组成员在该用户车辆上录入的记录
func (r *FuelRecordRepository) ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.FuelRecord, error) {
	var records []model.FuelRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? OR vehicle_id IN (SELECT id FROM vehicles WHERE user_id = ? AND deleted_at IS NULL)", userID, userID).
		Order("refuel_date ASC").
		Find(&records).Error
	return records, err
}

// ListVehicleIDsWithRecords 列出所有有加油记录的车辆 ID（用于批量重算）
func (r *FuelRecordRepository) ListVehicleIDsWithRecords(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.FuelRecord{}).Distinct("vehicle_id").Pluck("vehicle_id", &ids).Error
	return ids, err
}

// GetEfficiencyTrend 获取油耗趋势（最近 limit 条有油耗的记录，按时间倒序）
func (r *FuelRecordRepository) GetEfficiencyTrend(ctx context.Context, vehicleID uuid.UUID, limit int) ([]model.FuelRecord, error) {
	var records []model.FuelRecord
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND fuel_efficiency > 0", vehicleID).
		Order("refuel_date DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}
