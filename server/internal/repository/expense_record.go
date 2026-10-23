package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// ExpenseRecordRepository 开销记录数据访问
type ExpenseRecordRepository struct {
	db *gorm.DB
}

// NewExpenseRecordRepository 创建实例
func NewExpenseRecordRepository(db *gorm.DB) *ExpenseRecordRepository {
	return &ExpenseRecordRepository{db: db}
}

// Create 创建开销记录
func (r *ExpenseRecordRepository) Create(ctx context.Context, record *model.ExpenseRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByID 根据 ID 查询开销记录
func (r *ExpenseRecordRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ExpenseRecord, error) {
	var record model.ExpenseRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByIDAndVehicle 根据记录 ID 和车辆 ID 查询
func (r *ExpenseRecordRepository) GetByIDAndVehicle(ctx context.Context, id, vehicleID uuid.UUID) (*model.ExpenseRecord, error) {
	var record model.ExpenseRecord
	err := r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", id, vehicleID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByVehicle 查询车辆的开销记录（带筛选和分页）
func (r *ExpenseRecordRepository) ListByVehicle(
	ctx context.Context,
	vehicleID uuid.UUID,
	page, pageSize int,
	category string,
	startDate, endDate *time.Time,
	keyword string,
	minAmount, maxAmount float64,
) ([]model.ExpenseRecord, int64, error) {
	var records []model.ExpenseRecord
	var total int64

	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID)

	// 分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 日期范围筛选
	if startDate != nil {
		query = query.Where("expense_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("expense_date <= ?", *endDate)
	}

	// 关键词搜索（标题或商家）
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(title ILIKE ? OR vendor_name ILIKE ? OR note ILIKE ?)", like, like, like)
	}

	// 金额区间筛选
	if minAmount > 0 {
		query = query.Where("amount >= ?", minAmount)
	}
	if maxAmount > 0 {
		query = query.Where("amount <= ?", maxAmount)
	}

	// 使用独立查询链执行 COUNT，避免共享 *gorm.DB 状态导致分页查询异常
	countQuery := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).Where("vehicle_id = ?", vehicleID)
	if category != "" {
		countQuery = countQuery.Where("category = ?", category)
	}
	if startDate != nil {
		countQuery = countQuery.Where("expense_date >= ?", *startDate)
	}
	if endDate != nil {
		countQuery = countQuery.Where("expense_date <= ?", *endDate)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.Where("(title ILIKE ? OR vendor_name ILIKE ? OR note ILIKE ?)", like, like, like)
	}
	if minAmount > 0 {
		countQuery = countQuery.Where("amount >= ?", minAmount)
	}
	if maxAmount > 0 {
		countQuery = countQuery.Where("amount <= ?", maxAmount)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("expense_date DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

// Update 更新记录
func (r *ExpenseRecordRepository) Update(ctx context.Context, record *model.ExpenseRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// Delete 删除记录
func (r *ExpenseRecordRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", id, vehicleID).Delete(&model.ExpenseRecord{}).Error
}

// ListAllByUser 查询与用户相关的所有开销记录（用于数据导出，按日期升序）：
// 用户本人录入的记录，以及群组成员在该用户车辆上录入的记录
func (r *ExpenseRecordRepository) ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.ExpenseRecord, error) {
	var records []model.ExpenseRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? OR vehicle_id IN (SELECT id FROM vehicles WHERE user_id = ? AND deleted_at IS NULL)", userID, userID).
		Order("expense_date ASC, created_at ASC").
		Find(&records).Error
	return records, err
}

// --- 统计查询 ---

// ExpenseStatsByCurrency 按币种汇总
type ExpenseStatsByCurrency struct {
	CurrencyCode string  `json:"currency_code"`
	TotalAmount  float64 `json:"total_amount"`
	RecordCount  int     `json:"record_count"`
}

// GetTotalsByCurrency 按币种统计总开销
func (r *ExpenseRecordRepository) GetTotalsByCurrency(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCurrency, error) {
	var results []ExpenseStatsByCurrency
	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			currency_code,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as record_count
		`).
		Where("vehicle_id = ?", vehicleID).
		Group("currency_code").
		Order("total_amount DESC").
		Scan(&results).Error
	return results, err
}

// ExpenseStatsByCategory 按分类 × 币种汇总
type ExpenseStatsByCategory struct {
	Category     string  `json:"category"`
	CurrencyCode string  `json:"currency_code"`
	TotalAmount  float64 `json:"total_amount"`
	RecordCount  int     `json:"record_count"`
}

// GetBreakdownByCategory 按分类统计
func (r *ExpenseRecordRepository) GetBreakdownByCategory(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCategory, error) {
	var results []ExpenseStatsByCategory
	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			category,
			currency_code,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as record_count
		`).
		Where("vehicle_id = ?", vehicleID).
		Group("category, currency_code").
		Order("total_amount DESC").
		Scan(&results).Error
	return results, err
}

// ExpenseStatsByMonth 按月 × 币种汇总
type ExpenseStatsByMonth struct {
	Period       string  `json:"period"`
	CurrencyCode string  `json:"currency_code"`
	TotalAmount  float64 `json:"total_amount"`
	RecordCount  int     `json:"record_count"`
}

// GetMonthlyTrend 获取月度趋势（最近12个月）
func (r *ExpenseRecordRepository) GetMonthlyTrend(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByMonth, error) {
	var results []ExpenseStatsByMonth

	start := time.Now().AddDate(-1, 0, 0) // 近12个月

	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			TO_CHAR(expense_date, 'YYYY-MM') as period,
			currency_code,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as record_count
		`).
		Where("vehicle_id = ? AND expense_date >= ?", vehicleID, start).
		Group("period, currency_code").
		Order("period ASC").
		Scan(&results).Error
	return results, err
}

// GetLast30DaysByCurrency 按币种统计最近 30 天开销
func (r *ExpenseRecordRepository) GetLast30DaysByCurrency(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCurrency, error) {
	var results []ExpenseStatsByCurrency
	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			currency_code,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as record_count
		`).
		Where("vehicle_id = ? AND expense_date >= ?", vehicleID, time.Now().AddDate(0, 0, -30)).
		Group("currency_code").
		Scan(&results).Error
	return results, err
}

// GetTotalRecords 获取车辆开销记录总数
func (r *ExpenseRecordRepository) GetTotalRecords(ctx context.Context, vehicleID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Where("vehicle_id = ?", vehicleID).
		Count(&total).Error
	return total, err
}

// ExpensePeriodStatsResult 按时段 × 币种聚合开销统计结果
type ExpensePeriodStatsResult struct {
	Period       string  `json:"period"`
	CurrencyCode string  `json:"currency_code"`
	TotalRecords int     `json:"total_records"`
	TotalAmount  float64 `json:"total_amount"`
}

// GetExpenseStatsByMonth 按月聚合某年的开销统计数据
func (r *ExpenseRecordRepository) GetExpenseStatsByMonth(ctx context.Context, vehicleID uuid.UUID, year int) ([]ExpensePeriodStatsResult, error) {
	var results []ExpensePeriodStatsResult

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			TO_CHAR(expense_date, 'YYYY-MM') as period,
			currency_code,
			COUNT(*) as total_records,
			COALESCE(SUM(amount), 0) as total_amount
		`).
		Where("vehicle_id = ? AND expense_date >= ? AND expense_date < ?", vehicleID, start, end).
		Group("period, currency_code").
		Order("period ASC").
		Scan(&results).Error
	return results, err
}

// GetExpenseStatsByYear 按年聚合全部年份的开销统计数据
func (r *ExpenseRecordRepository) GetExpenseStatsByYear(ctx context.Context, vehicleID uuid.UUID) ([]ExpensePeriodStatsResult, error) {
	var results []ExpensePeriodStatsResult

	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			TO_CHAR(expense_date, 'YYYY') as period,
			currency_code,
			COUNT(*) as total_records,
			COALESCE(SUM(amount), 0) as total_amount
		`).
		Where("vehicle_id = ?", vehicleID).
		Group("period, currency_code").
		Order("period ASC").
		Scan(&results).Error
	return results, err
}

// GetExpenseCostByCurrency 按币种分组统计某车辆的总开销
func (r *ExpenseRecordRepository) GetExpenseCostByCurrency(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCurrency, error) {
	var results []ExpenseStatsByCurrency
	err := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select(`
			currency_code,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as record_count
		`).
		Where("vehicle_id = ?", vehicleID).
		Group("currency_code").
		Order("total_amount DESC").
		Scan(&results).Error
	return results, err
}

// GetDistinctVendorNames 获取某车辆去重的商家名列表（按使用频次降序）
func (r *ExpenseRecordRepository) GetDistinctVendorNames(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit int) ([]string, error) {
	var names []string
	query := r.db.WithContext(ctx).Model(&model.ExpenseRecord{}).
		Select("vendor_name").
		Where("user_id = ? AND vendor_name != ''", userID).
		Group("vendor_name").
		Order("COUNT(*) DESC")

	if vehicleID != nil {
		query = query.Where("vehicle_id = ?", *vehicleID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Pluck("vendor_name", &names).Error
	return names, err
}
