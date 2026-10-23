package service

import (
	"gastrack/internal/repository"
)

// fuelTotals 合并后的加油聚合量及按原始币种分组的费用
type fuelTotals struct {
	repository.FuelAggregate
	costs map[string]float64
}

func newFuelTotals() *fuelTotals {
	return &fuelTotals{costs: make(map[string]float64)}
}

// add 合并一行按币种分组的聚合结果
func (t *fuelTotals) add(row repository.FuelAggregateRow) {
	t.Merge(row.FuelAggregate)
	t.costs[row.CurrencyCode] += row.Cost
}

// merge 合并另一份合计
func (t *fuelTotals) merge(o *fuelTotals) {
	t.Merge(o.FuelAggregate)
	for cur, cost := range o.costs {
		t.costs[cur] += cost
	}
}

// amount 指定能源类别的加注量基准值（升或 kWh）
func (t *fuelTotals) amount(kind energyKind) float64 {
	if kind == energyElectric {
		return t.Kwh
	}
	return t.Liters
}

// avgEfficiency 指定能源类别按里程加权的平均油耗/电耗（基准单位）
func (t *fuelTotals) avgEfficiency(kind energyKind) float64 {
	if kind == energyElectric {
		return t.ElectricEfficiency()
	}
	return t.LiquidEfficiency()
}

// bestWorst 指定能源类别的最佳/最差油耗（基准单位，数值越小越省）
func (t *fuelTotals) bestWorst(kind energyKind) (best, worst float64) {
	if kind == energyElectric {
		return t.ElectricBest, t.ElectricWorst
	}
	return t.LiquidBest, t.LiquidWorst
}

// efficiencyRecords 指定能源类别有有效油耗的记录数
func (t *fuelTotals) efficiencyRecords(kind energyKind) int64 {
	if kind == energyElectric {
		return t.ElectricEffRecords
	}
	return t.LiquidEffRecords
}

// totalsByVehicle 按车辆合并聚合行
func totalsByVehicle(rows []repository.FuelAggregateRow) map[string]*fuelTotals {
	out := make(map[string]*fuelTotals)
	for _, row := range rows {
		key := row.VehicleID.String()
		if out[key] == nil {
			out[key] = newFuelTotals()
		}
		out[key].add(row)
	}
	return out
}

// totalsByPeriod 按时段合并聚合行，返回合计及按时段升序排列的时段列表
func totalsByPeriod(rows []repository.FuelAggregateRow) (map[string]*fuelTotals, []string) {
	out := make(map[string]*fuelTotals)
	var order []string
	for _, row := range rows {
		if out[row.Period] == nil {
			out[row.Period] = newFuelTotals()
			order = append(order, row.Period)
		}
		out[row.Period].add(row)
	}
	return out, order
}
