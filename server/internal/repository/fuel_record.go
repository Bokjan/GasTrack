package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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

	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID)
	query.Model(&model.FuelRecord{}).Count(&total)

	err := query.Order("refuel_date DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

// GetPreviousRecord 获取指定车辆在某个日期之前的最后一条记录（用于计算行驶距离）
func (r *FuelRecordRepository) GetPreviousRecord(ctx context.Context, vehicleID uuid.UUID, beforeDate time.Time) (*model.FuelRecord, error) {
	var record model.FuelRecord
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND refuel_date < ?", vehicleID, beforeDate).
		Order("refuel_date DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新记录
func (r *FuelRecordRepository) Update(ctx context.Context, record *model.FuelRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// Delete 删除记录
func (r *FuelRecordRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", id, vehicleID).Delete(&model.FuelRecord{}).Error
}

// --- 统计查询 ---

// StatsResult 统计查询结果
type StatsResult struct {
	TotalRecords    int64   `json:"total_records"`
	TotalFuel       float64 `json:"total_fuel"`
	TotalCost       float64 `json:"total_cost"`
	TotalDistance   float64 `json:"total_distance"`
	AvgEfficiency   float64 `json:"avg_efficiency"`
	BestEfficiency  float64 `json:"best_efficiency"`
	WorstEfficiency float64 `json:"worst_efficiency"`
}

// GetVehicleStats 获取车辆的汇总统计
func (r *FuelRecordRepository) GetVehicleStats(ctx context.Context, vehicleID uuid.UUID) (*StatsResult, error) {
	var result StatsResult

	err := r.db.WithContext(ctx).Model(&model.FuelRecord{}).
		Select(`
			COUNT(*) as total_records,
			COALESCE(SUM(fuel_amount), 0) as total_fuel,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(trip_distance), 0) as total_distance,
			COALESCE(AVG(NULLIF(fuel_efficiency, 0)), 0) as avg_efficiency,
			COALESCE(MIN(NULLIF(fuel_efficiency, 0)), 0) as best_efficiency,
			COALESCE(MAX(NULLIF(fuel_efficiency, 0)), 0) as worst_efficiency
		`).
		Where("vehicle_id = ? AND fuel_efficiency > 0", vehicleID).
		Scan(&result).Error

	return &result, err
}

// ExpenseByPeriod 按时间段统计费用
type ExpenseByPeriod struct {
	Period    string  `json:"period"`
	TotalCost float64 `json:"total_cost"`
	FuelCount int     `json:"fuel_count"`
	TotalFuel float64 `json:"total_fuel"`
}

// GetExpensesByMonth 按月统计费用
func (r *FuelRecordRepository) GetExpensesByMonth(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, start, end time.Time) ([]ExpenseByPeriod, error) {
	var results []ExpenseByPeriod

	query := r.db.WithContext(ctx).Model(&model.FuelRecord{}).
		Select(`
			TO_CHAR(refuel_date, 'YYYY-MM') as period,
			SUM(total_cost) as total_cost,
			COUNT(*) as fuel_count,
			SUM(fuel_amount) as total_fuel
		`).
		Where("user_id = ? AND refuel_date BETWEEN ? AND ?", userID, start, end)

	if vehicleID != nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}

	err := query.Group("period").Order("period ASC").Scan(&results).Error
	return results, err
}

// GetEfficiencyTrend 获取油耗趋势
func (r *FuelRecordRepository) GetEfficiencyTrend(ctx context.Context, vehicleID uuid.UUID, limit int) ([]model.FuelRecord, error) {
	var records []model.FuelRecord
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND fuel_efficiency > 0", vehicleID).
		Order("refuel_date DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}
