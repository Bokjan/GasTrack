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
	VehicleID   string  `json:"vehicle_id"`
	VehicleName string  `json:"vehicle_name"`
	OwnerName   string  `json:"owner_name"`
	VehicleType string  `json:"vehicle_type"`
	FuelType    string  `json:"fuel_type"`
	Records     int64   `json:"total_records"`
	TotalCost   float64 `json:"total_cost"`
	TotalFuel   float64 `json:"total_fuel"`
	AvgEff      float64 `json:"avg_efficiency"`
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
