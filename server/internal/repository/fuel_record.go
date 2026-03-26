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

// PeriodStatsResult 按时段聚合统计结果
type PeriodStatsResult struct {
	Period        string  `json:"period"`
	TotalRecords  int     `json:"total_records"`
	TotalFuel     float64 `json:"total_fuel"`
	TotalCost     float64 `json:"total_cost"`
	TotalDistance float64 `json:"total_distance"`
	AvgEfficiency float64 `json:"avg_efficiency"`
}

// GetStatsByMonth 按月聚合某年的统计数据
func (r *FuelRecordRepository) GetStatsByMonth(ctx context.Context, vehicleID uuid.UUID, year int) ([]PeriodStatsResult, error) {
	var results []PeriodStatsResult

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).Model(&model.FuelRecord{}).
		Select(`
			TO_CHAR(refuel_date, 'YYYY-MM') as period,
			COUNT(*) as total_records,
			COALESCE(SUM(fuel_amount), 0) as total_fuel,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(trip_distance), 0) as total_distance,
			COALESCE(AVG(NULLIF(fuel_efficiency, 0)), 0) as avg_efficiency
		`).
		Where("vehicle_id = ? AND refuel_date >= ? AND refuel_date < ?", vehicleID, start, end).
		Group("period").
		Order("period ASC").
		Scan(&results).Error
	return results, err
}

// GetStatsByYear 按年聚合全部年份的统计数据
func (r *FuelRecordRepository) GetStatsByYear(ctx context.Context, vehicleID uuid.UUID) ([]PeriodStatsResult, error) {
	var results []PeriodStatsResult

	err := r.db.WithContext(ctx).Model(&model.FuelRecord{}).
		Select(`
			TO_CHAR(refuel_date, 'YYYY') as period,
			COUNT(*) as total_records,
			COALESCE(SUM(fuel_amount), 0) as total_fuel,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(SUM(trip_distance), 0) as total_distance,
			COALESCE(AVG(NULLIF(fuel_efficiency, 0)), 0) as avg_efficiency
		`).
		Where("vehicle_id = ?", vehicleID).
		Group("period").
		Order("period ASC").
		Scan(&results).Error
	return results, err
}
