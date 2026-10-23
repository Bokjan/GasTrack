package dto

import "time"

// --- 群组相关 DTO ---

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	MaxMembers  int    `json:"max_members" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateGroupRequest 更新群组请求
type UpdateGroupRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	MaxMembers  *int    `json:"max_members" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// UpdateMemberRoleRequest 更新成员角色请求
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

// GroupResponse 群组响应
type GroupResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	OwnerID     string              `json:"owner_id"`
	OwnerName   string              `json:"owner_name,omitempty"`
	InviteCode  string              `json:"invite_code"`
	MaxMembers  int                 `json:"max_members"`
	Description string              `json:"description,omitempty"`
	MemberCount int                 `json:"member_count"`
	MyRole      string              `json:"my_role,omitempty"`
	Members     []GroupMemberDetail `json:"members,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}

// GroupMemberDetail 群组成员详情
type GroupMemberDetail struct {
	UserID   string    `json:"user_id"`
	Nickname string    `json:"nickname"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// GroupVehicleSummary 群组车辆汇总
type GroupVehicleSummary struct {
	VehicleID    string  `json:"vehicle_id"`
	VehicleName  string  `json:"vehicle_name"`
	OwnerID      string  `json:"owner_id"`
	OwnerName    string  `json:"owner_name"`
	VehicleType  string  `json:"vehicle_type"`
	FuelType     string  `json:"fuel_type"`
	IsShared     bool    `json:"is_shared"`     // 是否已共享到本群组（查看者自己的未共享车辆也会列出）
	CurrencyCode string  `json:"currency_code"` // total_cost 的币种（查看者偏好币种，按参考汇率换算）
	Records      int64   `json:"total_records"`
	TotalCost    float64 `json:"total_cost"`
	TotalFuel    float64 `json:"total_fuel"`     // 燃油车为升，电动车为 kWh（见 fuel_type）
	AvgEff       float64 `json:"avg_efficiency"` // 燃油车为 L/100km，电动车为 kWh/100km；按里程加权
}

// GroupOverviewResponse 群组数据汇总响应
type GroupOverviewResponse struct {
	GroupID      string                `json:"group_id"`
	GroupName    string                `json:"group_name"`
	MemberCount  int                  `json:"member_count"`
	VehicleCount int                  `json:"vehicle_count"`
	Vehicles     []GroupVehicleSummary `json:"vehicles"`
}

// JoinGroupResponse 加入群组响应
type JoinGroupResponse struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Role      string `json:"role"`
}

// --- 车辆共享相关 DTO ---

// ShareVehicleRequest 共享车辆请求
type ShareVehicleRequest struct {
	VehicleID string `json:"vehicle_id" validate:"required"`
}

// SharedVehicleResponse 共享车辆响应
type SharedVehicleResponse struct {
	ID          string `json:"id"`
	GroupID     string `json:"group_id"`
	VehicleID   string `json:"vehicle_id"`
	VehicleName string `json:"vehicle_name"`
	OwnerName   string `json:"owner_name"`
	SharedAt    string `json:"shared_at"`
}

// --- 排行榜相关 DTO ---

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Nickname    string  `json:"nickname"`
	VehicleID   string  `json:"vehicle_id"`
	VehicleName string  `json:"vehicle_name"`
	Value       float64 `json:"value"`
	DiffFromAvg float64 `json:"diff_from_avg"`
	RecordCount int     `json:"record_count"`
	IsSelf      bool    `json:"is_self"`
}

// LeaderboardResponse 排行榜响应
type LeaderboardResponse struct {
	GroupID           string             `json:"group_id"`
	GroupName         string             `json:"group_name"`
	Metric            string             `json:"metric"`
	Energy            string             `json:"energy"` // fuel / electric：油耗排行只比较同一能源类别
	Period            string             `json:"period"`
	PeriodLabel       string             `json:"period_label"`
	GroupAvg          float64            `json:"group_avg"`
	Unit              string             `json:"unit"` // 油耗为基准单位 L/100km 或 kWh/100km；费用为空（已换算为查看者币种）
	Rankings          []LeaderboardEntry `json:"rankings"`
	TotalParticipants int                `json:"total_participants"`
}

// --- 群组费用统计看板相关 DTO ---

// GroupExpenseSummary 群组费用统计摘要
type GroupExpenseSummary struct {
	TotalCost           float64 `json:"total_cost"`     // 换算为查看者币种
	TotalFuel           float64 `json:"total_fuel"`     // 燃油（L），不含充电
	TotalEnergy         float64 `json:"total_energy"`   // 充电（kWh）
	TotalDistance       float64 `json:"total_distance"` // km
	AvgEfficiency       float64 `json:"avg_efficiency"` // 燃油车平均油耗 L/100km（按里程加权）
	AvgEnergyEfficiency float64 `json:"avg_energy_efficiency"` // 电动车平均电耗 kWh/100km
	CostChangePct       float64 `json:"cost_change_pct"`
	FuelChangePct       float64 `json:"fuel_change_pct"`
	DistanceChangePct   float64 `json:"distance_change_pct"`
	EfficiencyChangePct float64 `json:"efficiency_change_pct"`
}

// MemberCostBreakdown 成员费用分解
type MemberCostBreakdown struct {
	UserID       string  `json:"user_id"`
	Nickname     string  `json:"nickname"`
	TotalCost    float64 `json:"total_cost"`    // 换算为 currency_code（查看者币种）
	CurrencyCode string  `json:"currency_code"`
	TotalFuel    float64 `json:"total_fuel"`    // 燃油（L）
	TotalEnergy  float64 `json:"total_energy"`  // 充电（kWh）
	Percentage   float64 `json:"percentage"`
}

// MemberCostItem 成员费用项（用于趋势图中的 by_member）
type MemberCostItem struct {
	UserID       string  `json:"user_id"`
	Nickname     string  `json:"nickname"`
	Cost         float64 `json:"cost"`
	CurrencyCode string  `json:"currency_code"`
}

// GroupTrendItem 群组趋势项
type GroupTrendItem struct {
	PeriodLabel         string           `json:"period_label"`
	TotalCost           float64          `json:"total_cost"` // 换算为查看者币种
	TotalFuel           float64          `json:"total_fuel"` // L
	TotalEnergy         float64          `json:"total_energy"` // kWh
	TotalDistance       float64          `json:"total_distance"`
	AvgEfficiency       float64          `json:"avg_efficiency"`        // L/100km
	AvgEnergyEfficiency float64          `json:"avg_energy_efficiency"` // kWh/100km
	ByMember            []MemberCostItem `json:"by_member,omitempty"`
}

// GroupExpenseStatsResponse 群组费用统计看板响应
type GroupExpenseStatsResponse struct {
	GroupID         string                `json:"group_id"`
	GroupName       string                `json:"group_name"`
	Period          string                `json:"period"`
	Year            int                   `json:"year,omitempty"`
	Summary         GroupExpenseSummary    `json:"summary"`
	CurrencyCode    string                 `json:"currency_code"` // 所有金额换算后的币种（查看者偏好币种）
	UnconvertedCurrencies []string         `json:"unconverted_currencies,omitempty"`
	MemberBreakdown []MemberCostBreakdown `json:"member_breakdown"`
	TrendItems      []GroupTrendItem       `json:"trend_items"`
	PrevTrendItems  []GroupTrendItem       `json:"prev_trend_items,omitempty"`
}

// --- 加油站推荐共享相关 DTO ---

// StationVisitor 加油站常客
type StationVisitor struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Count    int    `json:"count"`
}

// StationInfo 加油站信息
type StationInfo struct {
	StationName     string           `json:"station_name"`
	AvgUnitPrice    float64          `json:"avg_unit_price"`
	LatestUnitPrice float64          `json:"latest_unit_price"`
	PriceUnit       string           `json:"price_unit"` // L（gal 已换算为每升）或 kWh
	PriceTrend      string           `json:"price_trend"` // "up" / "down" / "stable"
	CurrencyCode    string           `json:"currency_code"`
	VisitCount      int              `json:"visit_count"`
	Visitors        []StationVisitor `json:"visitors"`
	LatestVisit     string           `json:"latest_visit"`
	FuelGradesSeen  []string         `json:"fuel_grades_seen"`
}

// GroupStationStatsResponse 加油站推荐共享响应
type GroupStationStatsResponse struct {
	GroupID       string        `json:"group_id"`
	GroupName     string        `json:"group_name"`
	TotalStations int           `json:"total_stations"`
	Stations      []StationInfo `json:"stations"`
}
