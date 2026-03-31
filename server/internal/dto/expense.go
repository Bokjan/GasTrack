package dto

import "time"

// --- 开销记录相关 DTO ---

// CreateExpenseRequest 创建开销记录请求
type CreateExpenseRequest struct {
	Category            string  `json:"category" validate:"required,oneof=maintenance repair insurance parking toll car_wash inspection parts fine other"`
	MaintenanceCategory string  `json:"maintenance_category" validate:"omitempty,oneof=oil_change tire_rotation brake_pads air_filter transmission coolant spark_plugs battery tire_replace inspection custom"`
	Title               string  `json:"title" validate:"required,min=1,max=200"`
	Amount              float64 `json:"amount" validate:"required,gt=0"`
	CurrencyCode        string  `json:"currency_code" validate:"required,len=3"`
	VendorName          string  `json:"vendor_name" validate:"omitempty,max=200"`
	Odometer            float64 `json:"odometer" validate:"omitempty,gte=0"`
	DistanceUnit        string  `json:"distance_unit" validate:"omitempty,oneof=km mi"`
	Note                string  `json:"note" validate:"omitempty,max=1000"`
	ExpenseDate         string  `json:"expense_date" validate:"required"` // ISO 8601
	ReminderID          string  `json:"reminder_id" validate:"omitempty,uuid"`
}

// UpdateExpenseRequest 编辑开销记录请求（指针字段实现 partial update）
type UpdateExpenseRequest struct {
	Category            *string  `json:"category" validate:"omitempty,oneof=maintenance repair insurance parking toll car_wash inspection parts fine other"`
	MaintenanceCategory *string  `json:"maintenance_category" validate:"omitempty,oneof=oil_change tire_rotation brake_pads air_filter transmission coolant spark_plugs battery tire_replace inspection custom"`
	Title               *string  `json:"title" validate:"omitempty,min=1,max=200"`
	Amount              *float64 `json:"amount" validate:"omitempty,gt=0"`
	CurrencyCode        *string  `json:"currency_code" validate:"omitempty,len=3"`
	VendorName          *string  `json:"vendor_name" validate:"omitempty,max=200"`
	Odometer            *float64 `json:"odometer" validate:"omitempty,gte=0"`
	DistanceUnit        *string  `json:"distance_unit" validate:"omitempty,oneof=km mi"`
	Note                *string  `json:"note" validate:"omitempty,max=1000"`
	ExpenseDate         *string  `json:"expense_date"`
	ReminderID          *string  `json:"reminder_id" validate:"omitempty"`
}

// ExpenseResponse 开销记录响应
type ExpenseResponse struct {
	ID                  string    `json:"id"`
	VehicleID           string    `json:"vehicle_id"`
	UserID              string    `json:"user_id"`
	Category            string    `json:"category"`
	MaintenanceCategory string    `json:"maintenance_category,omitempty"`
	Title               string    `json:"title"`
	Amount              float64   `json:"amount"`
	CurrencyCode        string    `json:"currency_code"`
	VendorName          string    `json:"vendor_name,omitempty"`
	Odometer            float64   `json:"odometer,omitempty"`
	DistanceUnit        string    `json:"distance_unit,omitempty"`
	Note                string    `json:"note,omitempty"`
	ReceiptURL          string    `json:"receipt_url,omitempty"`
	ExpenseDate         time.Time `json:"expense_date"`
	ReminderID          string    `json:"reminder_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ExpenseListFilter 列表筛选参数
type ExpenseListFilter struct {
	Page      int     `json:"page"`
	PageSize  int     `json:"page_size"`
	Category  string  `json:"category"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Keyword   string  `json:"keyword"`
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
}

// --- 开销统计相关 DTO ---

// ExpenseCurrencyTotal 按币种汇总
type ExpenseCurrencyTotal struct {
	CurrencyCode string  `json:"currency_code"`
	TotalAmount  float64 `json:"total_amount"`
	RecordCount  int     `json:"record_count"`
}

// ExpenseCategoryBreakdown 按分类汇总
type ExpenseCategoryBreakdown struct {
	Category    string  `json:"category"`
	TotalAmount float64 `json:"total_amount"`
	RecordCount int     `json:"record_count"`
	Percentage  float64 `json:"percentage"`
}

// ExpenseMonthlyTrend 月度趋势
type ExpenseMonthlyTrend struct {
	Period      string  `json:"period"` // "2026-01"
	TotalAmount float64 `json:"total_amount"`
	RecordCount int     `json:"record_count"`
}

// VehicleExpenseStatsResponse 车辆开销统计响应
type VehicleExpenseStatsResponse struct {
	VehicleID          string                     `json:"vehicle_id"`
	TotalRecords       int64                      `json:"total_records"`
	TotalsByCurrency   []ExpenseCurrencyTotal     `json:"totals_by_currency"`
	CategoryBreakdown  []ExpenseCategoryBreakdown `json:"category_breakdown"`
	MonthlyTrend       []ExpenseMonthlyTrend      `json:"monthly_trend"`
	Last30DaysAmount   float64                    `json:"last_30_days_amount"`
	Last30DaysCurrency string                     `json:"last_30_days_currency"`
}
