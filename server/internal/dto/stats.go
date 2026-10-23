package dto

// --- 统计相关 DTO ---

// VehicleStatsResponse 车辆统计响应
type VehicleStatsResponse struct {
	VehicleID       string             `json:"vehicle_id"`
	VehicleName     string             `json:"vehicle_name"`
	EnergyType      string             `json:"energy_type"`          // fuel / electric（电动车只统计充电记录）
	TotalRecords    int64              `json:"total_records"`        // 总加油/充电次数
	TotalFuel       float64            `json:"total_fuel"`           // 总加油量/充电量（单位见 volume_unit）
	VolumeUnit      string             `json:"volume_unit"`          // L / gal / kWh
	TotalCost       float64            `json:"total_cost"`           // 总费用，按参考汇率换算为 currency_code
	TotalDistance   float64            `json:"total_distance"`       // 总行驶里程（km 或 mi，随用户单位制）
	AvgEfficiency   float64            `json:"avg_efficiency"`       // 平均油耗/电耗 = 总油量 / 总里程（按区间里程加权）
	BestEfficiency  float64            `json:"best_efficiency"`      // 最佳油耗
	WorstEfficiency float64            `json:"worst_efficiency"`     // 最差油耗
	AvgCostPerKm    float64            `json:"avg_cost_per_km"`      // 每公里（英制为每英里）平均费用
	AvgCostPerFill  float64            `json:"avg_cost_per_fill"`    // 每次平均费用
	CurrencyCode    string             `json:"currency_code"`        // 用户偏好币种（换算目标）
	FuelUnit        string             `json:"fuel_efficiency_unit"` // 油耗/电耗单位（电动车为 kWh/100km、km/kWh、mi/kWh）
	CostsByCurrency map[string]float64 `json:"costs_by_currency"`   // 按原始入账币种分组的费用明细
	// UnconvertedCurrencies 无参考汇率、未计入 total_cost 等换算金额的币种
	UnconvertedCurrencies []string `json:"unconverted_currencies,omitempty"`
	// 开销统计
	TotalExpenseCost       float64            `json:"total_expense_cost"`        // 其他开销总额（换算为 currency_code）
	ExpenseCostsByCurrency map[string]float64 `json:"expense_costs_by_currency"` // 按原始入账币种分组的开销明细
}

// OverviewStatsResponse 全局统计总览响应
type OverviewStatsResponse struct {
	TotalVehicles  int64   `json:"total_vehicles"`
	TotalRecords   int64   `json:"total_records"`
	TotalFuel      float64 `json:"total_fuel"`      // 燃油总量（单位见 volume_unit：L / gal），不含充电
	VolumeUnit     string  `json:"volume_unit"`     // L / gal
	TotalEnergy    float64 `json:"total_energy"`    // 充电总量（kWh）
	TotalCost      float64 `json:"total_cost"`      // 加油/充电总费用，按参考汇率换算为 currency_code
	TotalDistance  float64 `json:"total_distance"`  // km 或 mi
	AvgConsumption float64 `json:"avg_consumption"` // 燃油车平均油耗（总油量 / 总里程），单位见 fuel_efficiency_unit
	FuelUnit       string  `json:"fuel_efficiency_unit"`
	// AvgEnergyConsumption 电动车平均电耗（总电量 / 总里程），单位见 energy_efficiency_unit
	AvgEnergyConsumption float64 `json:"avg_energy_consumption"`
	EnergyUnit           string  `json:"energy_efficiency_unit"`
	CurrencyCode    string                 `json:"currency_code"`
	CostsByCurrency map[string]float64     `json:"costs_by_currency"` // 按原始入账币种分组的总费用明细
	Vehicles        []VehicleStatsResponse `json:"vehicles"`
	// 开销统计
	TotalExpenseCost       float64            `json:"total_expense_cost"`        // 全局其他开销总额（换算为 currency_code）
	ExpenseCostsByCurrency map[string]float64 `json:"expense_costs_by_currency"` // 全局按币种分组的开销明细
	UnconvertedCurrencies  []string           `json:"unconverted_currencies,omitempty"`
}

// ExpenseStatsRequest 费用统计请求参数
type ExpenseStatsRequest struct {
	VehicleID string `json:"vehicle_id"` // 可选，不传则查全部车辆
	Period    string `json:"period"`     // month/quarter/year
	StartDate string `json:"start_date"` // ISO 8601
	EndDate   string `json:"end_date"`   // ISO 8601
}

// ExpenseStatsItem 费用统计单项
type ExpenseStatsItem struct {
	Period    string  `json:"period"`     // 如 "2024-01", "2024-Q1"
	TotalCost float64 `json:"total_cost"` // 该时段总费用
	FuelCount int     `json:"fuel_count"` // 加油次数
	TotalFuel float64 `json:"total_fuel"` // 总加油量
}

// ExpenseStatsResponse 费用统计响应
type ExpenseStatsResponse struct {
	CurrencyCode string             `json:"currency_code"`
	Items        []ExpenseStatsItem `json:"items"`
}

// PeriodStatsItem 时段统计单项（按月或按年）
type PeriodStatsItem struct {
	Period        string  `json:"period"`         // "2026-01" 或 "2026"
	TotalRecords  int     `json:"total_records"`  // 加油次数
	TotalFuel     float64 `json:"total_fuel"`     // 总加油量/充电量（单位见 volume_unit）
	TotalCost     float64 `json:"total_cost"`     // 总费用（换算为 currency_code）
	TotalDistance float64 `json:"total_distance"` // 总行驶距离
	AvgEfficiency float64 `json:"avg_efficiency"` // 平均油耗/电耗（按里程加权）
}

// PeriodStatsResponse 时段统计响应（含往年同比数据）
type PeriodStatsResponse struct {
	VehicleID      string            `json:"vehicle_id"`
	VehicleName    string            `json:"vehicle_name"`
	Period         string            `json:"period"`          // "month" 或 "year"
	Year           int               `json:"year,omitempty"`  // 查询年份（按月时有值）
	CurrencyCode   string            `json:"currency_code"`
	FuelUnit       string            `json:"fuel_efficiency_unit"`
	VolumeUnit     string            `json:"volume_unit"`
	UnconvertedCurrencies []string   `json:"unconverted_currencies,omitempty"`
	Items          []PeriodStatsItem `json:"items"`           // 当前周期数据
	PrevItems      []PeriodStatsItem `json:"prev_items"`      // 上一周期同比数据
}

// FuelEfficiencyTrendItem 油耗趋势单项
type FuelEfficiencyTrendItem struct {
	Date           string  `json:"date"`            // ISO date
	FuelEfficiency float64 `json:"fuel_efficiency"` // 油耗值
	TripDistance   float64 `json:"trip_distance"`   // 行驶距离
}

// FuelEfficiencyTrendResponse 油耗趋势响应
type FuelEfficiencyTrendResponse struct {
	VehicleID      string                    `json:"vehicle_id"`
	VehicleName    string                    `json:"vehicle_name"`
	EfficiencyUnit string                    `json:"efficiency_unit"` // L/100km / km/L / MPG；电动车为 kWh/100km / km/kWh / mi/kWh
	Items          []FuelEfficiencyTrendItem `json:"items"`
}

// --- 开销时段统计 ---

// ExpensePeriodStatsItem 开销时段统计单项
type ExpensePeriodStatsItem struct {
	Period       string  `json:"period"`        // "2026-01" 或 "2026"
	TotalRecords int     `json:"total_records"`
	TotalAmount  float64 `json:"total_amount"` // 换算为 currency_code
}

// ExpensePeriodStatsResponse 开销时段统计响应（含往年同比数据）
type ExpensePeriodStatsResponse struct {
	VehicleID       string                   `json:"vehicle_id"`
	VehicleName     string                   `json:"vehicle_name"`
	Period          string                   `json:"period"`          // "month" 或 "year"
	Year            int                      `json:"year,omitempty"`
	CurrencyCode    string                   `json:"currency_code"`
	CostsByCurrency map[string]float64       `json:"costs_by_currency"` // 按原始入账币种分组的开销明细
	UnconvertedCurrencies []string           `json:"unconverted_currencies,omitempty"`
	Items           []ExpensePeriodStatsItem `json:"items"`
	PrevItems       []ExpensePeriodStatsItem `json:"prev_items"`
}
