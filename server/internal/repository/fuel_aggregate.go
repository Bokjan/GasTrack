package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
	"gastrack/internal/pkg/convert"
)

// FuelAggregate 加油/充电记录的可加性聚合量。
// 只包含可以跨分组直接相加的量（计数、求和、加权和、最值），平均值由调用方合并后再计算，
// 从而可以先按 车辆 × 币种 ×（时段）分组查询，再在 Go 中按任意维度汇总与换汇。
// 燃油与电分开统计：燃油量统一为升（gal 已换算），电量为 kWh；油耗/电耗按区间里程加权。
type FuelAggregate struct {
	Records  int64   `gorm:"column:records"`
	Liters   float64 `gorm:"column:liters"`   // 燃油量（L）
	Kwh      float64 `gorm:"column:kwh"`      // 电量（kWh）
	Distance float64 `gorm:"column:distance"` // 行驶距离（km）

	LiquidEffRecords int64   `gorm:"column:liquid_eff_records"`
	LiquidEffSum     float64 `gorm:"column:liquid_eff_sum"`    // Σ(L/100km × 区间里程)
	LiquidEffWeight  float64 `gorm:"column:liquid_eff_weight"` // Σ区间里程
	LiquidBest       float64 `gorm:"column:liquid_best"`
	LiquidWorst      float64 `gorm:"column:liquid_worst"`

	ElectricEffRecords int64   `gorm:"column:electric_eff_records"`
	ElectricEffSum     float64 `gorm:"column:electric_eff_sum"`
	ElectricEffWeight  float64 `gorm:"column:electric_eff_weight"`
	ElectricBest       float64 `gorm:"column:electric_best"`
	ElectricWorst      float64 `gorm:"column:electric_worst"`
}

// Merge 累加另一组聚合量
func (a *FuelAggregate) Merge(b FuelAggregate) {
	a.Records += b.Records
	a.Liters += b.Liters
	a.Kwh += b.Kwh
	a.Distance += b.Distance

	a.LiquidEffRecords += b.LiquidEffRecords
	a.LiquidEffSum += b.LiquidEffSum
	a.LiquidEffWeight += b.LiquidEffWeight
	a.LiquidBest = minPositive(a.LiquidBest, b.LiquidBest)
	a.LiquidWorst = max(a.LiquidWorst, b.LiquidWorst)

	a.ElectricEffRecords += b.ElectricEffRecords
	a.ElectricEffSum += b.ElectricEffSum
	a.ElectricEffWeight += b.ElectricEffWeight
	a.ElectricBest = minPositive(a.ElectricBest, b.ElectricBest)
	a.ElectricWorst = max(a.ElectricWorst, b.ElectricWorst)
}

// LiquidEfficiency 按里程加权的平均油耗（L/100km），即 总油量 / 总里程
func (a FuelAggregate) LiquidEfficiency() float64 {
	if a.LiquidEffWeight <= 0 {
		return 0
	}
	return a.LiquidEffSum / a.LiquidEffWeight
}

// ElectricEfficiency 按里程加权的平均电耗（kWh/100km）
func (a FuelAggregate) ElectricEfficiency() float64 {
	if a.ElectricEffWeight <= 0 {
		return 0
	}
	return a.ElectricEffSum / a.ElectricEffWeight
}

func minPositive(a, b float64) float64 {
	switch {
	case a <= 0:
		return b
	case b <= 0:
		return a
	default:
		return min(a, b)
	}
}

// FuelAggregateRow 按 车辆 × 币种 ×（时段）分组的聚合结果；Cost 为该币种下的费用合计
type FuelAggregateRow struct {
	VehicleID     uuid.UUID `gorm:"column:vehicle_id"`
	Period        string    `gorm:"column:period"`
	CurrencyCode  string    `gorm:"column:currency_code"`
	Cost          float64   `gorm:"column:cost"`
	FuelAggregate `gorm:"embedded"`
}

// FuelAggregateQuery 聚合查询条件
type FuelAggregateQuery struct {
	VehicleIDs   []uuid.UUID
	PeriodFormat string     // 为空则不按时段分组；否则为 TO_CHAR 格式，如 "YYYY-MM"、"YYYY"
	Start, End   *time.Time // 加油日期范围 [Start, End)，均可为空
}

// PeriodMonth / PeriodYear 常用时段格式
const (
	PeriodMonth = "YYYY-MM"
	PeriodYear  = "YYYY"
)

// fuelAggregateColumns 聚合列。区间里程优先使用 efficiency_distance（满油到满油区间），
// 历史数据（该列为空）回退到 trip_distance，与旧算法计算油耗时使用的里程一致。
var fuelAggregateColumns = strings.NewReplacer(
	"{liquid}", "fuel_efficiency > 0 AND fuel_unit IS DISTINCT FROM 'kWh'",
	"{electric}", "fuel_efficiency > 0 AND fuel_unit = 'kWh'",
	"{weight}", "COALESCE(NULLIF(efficiency_distance, 0), trip_distance)",
	"{gal}", fmt.Sprintf("%v", convert.GallonToLiter),
).Replace(`
	COUNT(*) AS records,
	COALESCE(SUM(CASE WHEN fuel_unit = 'gal' THEN fuel_amount * {gal} WHEN fuel_unit = 'kWh' THEN 0 ELSE fuel_amount END), 0) AS liters,
	COALESCE(SUM(CASE WHEN fuel_unit = 'kWh' THEN fuel_amount ELSE 0 END), 0) AS kwh,
	COALESCE(SUM(trip_distance), 0) AS distance,
	COUNT(*) FILTER (WHERE {liquid}) AS liquid_eff_records,
	COALESCE(SUM(fuel_efficiency * {weight}) FILTER (WHERE {liquid}), 0) AS liquid_eff_sum,
	COALESCE(SUM({weight}) FILTER (WHERE {liquid}), 0) AS liquid_eff_weight,
	COALESCE(MIN(fuel_efficiency) FILTER (WHERE {liquid}), 0) AS liquid_best,
	COALESCE(MAX(fuel_efficiency) FILTER (WHERE {liquid}), 0) AS liquid_worst,
	COUNT(*) FILTER (WHERE {electric}) AS electric_eff_records,
	COALESCE(SUM(fuel_efficiency * {weight}) FILTER (WHERE {electric}), 0) AS electric_eff_sum,
	COALESCE(SUM({weight}) FILTER (WHERE {electric}), 0) AS electric_eff_weight,
	COALESCE(MIN(fuel_efficiency) FILTER (WHERE {electric}), 0) AS electric_best,
	COALESCE(MAX(fuel_efficiency) FILTER (WHERE {electric}), 0) AS electric_worst,
	COALESCE(SUM(total_cost), 0) AS cost`)

// Aggregate 按 车辆 × 币种 ×（时段）分组聚合加油记录
func (r *FuelRecordRepository) Aggregate(ctx context.Context, q FuelAggregateQuery) ([]FuelAggregateRow, error) {
	var rows []FuelAggregateRow
	if len(q.VehicleIDs) == 0 {
		return rows, nil
	}

	db := r.db.WithContext(ctx).Model(&model.FuelRecord{})
	if q.PeriodFormat != "" {
		db = db.Select("vehicle_id, currency_code, TO_CHAR(refuel_date, ?) AS period,"+fuelAggregateColumns, q.PeriodFormat)
	} else {
		db = db.Select("vehicle_id, currency_code, '' AS period," + fuelAggregateColumns)
	}
	db = db.Where("vehicle_id IN ?", q.VehicleIDs)
	db = applyDateRange(db, "refuel_date", q.Start, q.End)

	err := db.Group("vehicle_id, currency_code, period").Order("period ASC").Scan(&rows).Error
	return rows, err
}

// applyDateRange 追加 [start, end) 日期范围条件
func applyDateRange(db *gorm.DB, column string, start, end *time.Time) *gorm.DB {
	if start != nil {
		db = db.Where(column+" >= ?", *start)
	}
	if end != nil {
		db = db.Where(column+" < ?", *end)
	}
	return db
}
