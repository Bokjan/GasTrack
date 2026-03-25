package dto

// --- 统计相关 DTO ---

// VehicleStatsResponse 车辆统计响应
type VehicleStatsResponse struct {
	VehicleID       string  `json:"vehicle_id"`
	VehicleName     string  `json:"vehicle_name"`
	TotalRecords    int64   `json:"total_records"`        // 总加油次数
	TotalFuel       float64 `json:"total_fuel"`           // 总加油量(L)
	TotalCost       float64 `json:"total_cost"`           // 总费用
	TotalDistance   float64 `json:"total_distance"`       // 总行驶里程(km)
	AvgEfficiency   float64 `json:"avg_efficiency"`       // 平均油耗(L/100km)
	BestEfficiency  float64 `json:"best_efficiency"`      // 最佳油耗
	WorstEfficiency float64 `json:"worst_efficiency"`     // 最差油耗
	AvgCostPerKm    float64 `json:"avg_cost_per_km"`      // 每公里平均费用
	AvgCostPerFill  float64 `json:"avg_cost_per_fill"`    // 每次平均费用
	CurrencyCode    string  `json:"currency_code"`        // 费用币种
	FuelUnit        string  `json:"fuel_efficiency_unit"` // 用户偏好的油耗单位
}

// OverviewStatsResponse 全局统计总览响应
type OverviewStatsResponse struct {
	TotalVehicles  int64                  `json:"total_vehicles"`
	TotalRecords   int64                  `json:"total_records"`
	TotalFuel      float64                `json:"total_fuel"`
	TotalCost      float64                `json:"total_cost"`
	TotalDistance  float64                `json:"total_distance"`
	AvgConsumption float64                `json:"avg_consumption"`
	CurrencyCode   string                 `json:"currency_code"`
	Vehicles       []VehicleStatsResponse `json:"vehicles"`
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
	EfficiencyUnit string                    `json:"efficiency_unit"` // L/100km / km/L / MPG
	Items          []FuelEfficiencyTrendItem `json:"items"`
}
