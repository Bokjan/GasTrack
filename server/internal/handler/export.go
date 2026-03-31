package handler

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"gastrack/internal/middleware"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// ExportHandler 数据导出相关 HTTP 处理器
type ExportHandler struct {
	exportService *service.ExportService
	logger        *zap.Logger
}

// NewExportHandler 创建 ExportHandler 实例
func NewExportHandler(exportService *service.ExportService, logger *zap.Logger) *ExportHandler {
	return &ExportHandler{exportService: exportService, logger: logger}
}

// ExportMyData 导出当前用户的数据
// GET /api/v1/users/me/export?format=csv|zip|json&scope=basic|full
func (h *ExportHandler) ExportMyData(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	// 解析参数
	format := service.ExportFormat(r.URL.Query().Get("format"))
	scope := service.ExportScope(r.URL.Query().Get("scope"))

	// 默认值：保持向后兼容，无参数时返回基础 CSV
	if format == "" {
		format = service.ExportFormatCSV
	}
	if scope == "" {
		scope = service.ExportScopeBasic
	}

	// 校验
	switch format {
	case service.ExportFormatCSV, service.ExportFormatZIP, service.ExportFormatJSON:
		// OK
	default:
		respond.BadRequest(w, "invalid format, must be csv, zip or json")
		return
	}
	switch scope {
	case service.ExportScopeBasic, service.ExportScopeFull:
		// OK
	default:
		respond.BadRequest(w, "invalid scope, must be basic or full")
		return
	}

	// 如果 scope=full 且 format=csv，自动升级为 zip（多模块单文件 CSV 不合适）
	if scope == service.ExportScopeFull && format == service.ExportFormatCSV {
		format = service.ExportFormatZIP
	}

	data, err := h.exportService.GatherUserData(r.Context(), userID, scope)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	timestamp := time.Now().Format("20060102-150405")

	switch format {
	case service.ExportFormatCSV:
		h.writeCSV(w, data, timestamp)
	case service.ExportFormatZIP:
		h.writeZIP(w, data, timestamp)
	case service.ExportFormatJSON:
		h.writeJSON(w, data, timestamp)
	}
}

// ==================== CSV 导出（P0 兼容）====================

func (h *ExportHandler) writeCSV(w http.ResponseWriter, data *service.UserExportData, timestamp string) {
	filename := fmt.Sprintf("gastrack-export-%s.csv", timestamp)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	// UTF-8 BOM
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writeUserProfileCSV(writer, data)
	writeVehiclesCSV(writer, data)
	writeFuelRecordsCSV(writer, data)
}

// ==================== ZIP 导出（P1 完整版）====================

func (h *ExportHandler) writeZIP(w http.ResponseWriter, data *service.UserExportData, timestamp string) {
	filename := fmt.Sprintf("gastrack-export-%s.zip", timestamp)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	zw := zip.NewWriter(w)
	defer zw.Close()

	// 1. profile.csv
	if fw, err := zw.Create("profile.csv"); err == nil {
		cw := csv.NewWriter(fw)
		writeUserProfileCSV(cw, data)
		cw.Flush()
	}

	// 2. vehicles.csv
	if fw, err := zw.Create("vehicles.csv"); err == nil {
		cw := csv.NewWriter(fw)
		writeVehiclesCSVBody(cw, data)
		cw.Flush()
	}

	// 3. fuel_records.csv
	if fw, err := zw.Create("fuel_records.csv"); err == nil {
		cw := csv.NewWriter(fw)
		writeFuelRecordsCSVBody(cw, data)
		cw.Flush()
	}

	// 4. expense_records.csv（P1）
	if len(data.ExpenseRecords) > 0 {
		if fw, err := zw.Create("expense_records.csv"); err == nil {
			cw := csv.NewWriter(fw)
			writeExpenseRecordsCSV(cw, data)
			cw.Flush()
		}
	}

	// 5. reminders.csv（P1）
	if len(data.Reminders) > 0 {
		if fw, err := zw.Create("reminders.csv"); err == nil {
			cw := csv.NewWriter(fw)
			writeRemindersCSV(cw, data)
			cw.Flush()
		}
	}

	// 6. notifications.csv（P1）
	if len(data.Notifications) > 0 {
		if fw, err := zw.Create("notifications.csv"); err == nil {
			cw := csv.NewWriter(fw)
			writeNotificationsCSV(cw, data)
			cw.Flush()
		}
	}

	// 7. invite_codes.csv（P1）
	if len(data.InviteCodes) > 0 {
		if fw, err := zw.Create("invite_codes.csv"); err == nil {
			cw := csv.NewWriter(fw)
			writeInviteCodesCSV(cw, data)
			cw.Flush()
		}
	}

	// 8. groups.csv（P1）
	if len(data.Groups) > 0 || len(data.GroupMemberships) > 0 {
		if fw, err := zw.Create("groups.csv"); err == nil {
			cw := csv.NewWriter(fw)
			writeGroupsCSV(cw, data)
			cw.Flush()
		}
	}

	// 9. manifest.json
	if fw, err := zw.Create("manifest.json"); err == nil {
		manifest := buildManifest(data, timestamp)
		json.NewEncoder(fw).Encode(manifest)
	}
}

// ==================== JSON 导出（P2）====================

func (h *ExportHandler) writeJSON(w http.ResponseWriter, data *service.UserExportData, timestamp string) {
	filename := fmt.Sprintf("gastrack-export-%s.json", timestamp)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	export := buildJSONExport(data, timestamp)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(export)
}

// ==================== CSV 写入辅助方法 ====================

func writeUserProfileCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{"# User Profile"})
	writer.Write([]string{"ID", "Email", "Nickname", "Locale", "Timezone", "Country", "Currency", "Unit System", "Fuel Efficiency Unit", "Status", "Created At"})
	writer.Write([]string{
		data.User.ID.String(),
		data.User.Email,
		data.User.Nickname,
		data.User.Locale,
		data.User.Timezone,
		data.User.CountryCode,
		data.User.CurrencyCode,
		data.User.UnitSystem,
		data.User.FuelEfficiencyUnit,
		data.User.Status,
		data.User.CreatedAt.Format(time.RFC3339),
	})
	writer.Write([]string{})
}

func writeVehiclesCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{"# Vehicles"})
	writeVehiclesCSVBody(writer, data)
	writer.Write([]string{})
}

func writeVehiclesCSVBody(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Name", "Vehicle Type", "Brand", "Model", "Year",
		"Fuel Type", "Fuel Grade", "Tank Capacity (L)", "Battery Capacity (kWh)",
		"Engine CC", "License Plate", "Is Default", "Is Archived", "Created At",
	})
	for _, v := range data.Vehicles {
		writer.Write([]string{
			v.ID.String(),
			v.Name,
			string(v.VehicleType),
			v.Brand,
			v.Model,
			fmt.Sprintf("%d", v.Year),
			string(v.FuelType),
			v.FuelGrade,
			fmt.Sprintf("%.2f", v.TankCapacity),
			fmt.Sprintf("%.2f", v.BatteryCapacity),
			fmt.Sprintf("%d", v.EngineCC),
			v.LicensePlate,
			fmt.Sprintf("%t", v.IsDefault),
			fmt.Sprintf("%t", v.IsArchived),
			v.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeFuelRecordsCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{"# Fuel / Charging Records"})
	writeFuelRecordsCSVBody(writer, data)
}

func writeFuelRecordsCSVBody(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Vehicle ID", "Refuel Date", "Station Name",
		"Fuel Amount", "Fuel Unit", "Unit Price", "Total Cost", "Currency",
		"Odometer", "Distance Unit", "Is Full Tank", "Fuel Grade",
		"Trip Distance", "Fuel Efficiency", "Note", "Created At",
	})
	for _, rec := range data.FuelRecords {
		writer.Write([]string{
			rec.ID.String(),
			rec.VehicleID.String(),
			rec.RefuelDate.Format(time.RFC3339),
			rec.StationName,
			fmt.Sprintf("%.3f", rec.FuelAmount),
			rec.FuelUnit,
			fmt.Sprintf("%.4f", rec.UnitPrice),
			fmt.Sprintf("%.2f", rec.TotalCost),
			rec.CurrencyCode,
			fmt.Sprintf("%.1f", rec.Odometer),
			rec.DistanceUnit,
			fmt.Sprintf("%t", rec.IsFullTank),
			rec.FuelGrade,
			fmt.Sprintf("%.1f", rec.TripDistance),
			fmt.Sprintf("%.2f", rec.FuelEfficiency),
			rec.Note,
			rec.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeExpenseRecordsCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Vehicle ID", "Category", "Maintenance Category", "Title", "Amount",
		"Currency", "Vendor Name", "Odometer", "Distance Unit", "Note",
		"Expense Date", "Reminder ID", "Created At",
	})
	for _, e := range data.ExpenseRecords {
		reminderID := ""
		if e.ReminderID != nil {
			reminderID = e.ReminderID.String()
		}
		writer.Write([]string{
			e.ID.String(),
			e.VehicleID.String(),
			string(e.Category),
			string(e.MaintenanceCategory),
			e.Title,
			fmt.Sprintf("%.2f", e.Amount),
			e.CurrencyCode,
			e.VendorName,
			fmt.Sprintf("%.1f", e.Odometer),
			e.DistanceUnit,
			e.Note,
			e.ExpenseDate.Format(time.RFC3339),
			reminderID,
			e.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeRemindersCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Vehicle ID", "Type", "Category", "Title", "Description",
		"Trigger", "Mileage Interval", "Time Interval Days",
		"Last Mileage", "Last Date", "Next Mileage", "Next Date",
		"Is Enabled", "Created At",
	})
	for _, rem := range data.Reminders {
		lastDate := ""
		if rem.LastDate != nil {
			lastDate = rem.LastDate.Format(time.RFC3339)
		}
		nextDate := ""
		if rem.NextDate != nil {
			nextDate = rem.NextDate.Format(time.RFC3339)
		}
		writer.Write([]string{
			rem.ID.String(),
			rem.VehicleID.String(),
			string(rem.Type),
			string(rem.Category),
			rem.Title,
			rem.Description,
			string(rem.Trigger),
			fmt.Sprintf("%.1f", rem.MileageInterval),
			fmt.Sprintf("%d", rem.TimeIntervalDays),
			fmt.Sprintf("%.1f", rem.LastMileage),
			lastDate,
			fmt.Sprintf("%.1f", rem.NextMileage),
			nextDate,
			fmt.Sprintf("%t", rem.IsEnabled),
			rem.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeNotificationsCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Type", "Title", "Message", "Is Read", "Created At",
	})
	for _, n := range data.Notifications {
		writer.Write([]string{
			n.ID.String(),
			string(n.Type),
			n.Title,
			n.Message,
			fmt.Sprintf("%t", n.IsRead),
			n.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeInviteCodesCSV(writer *csv.Writer, data *service.UserExportData) {
	writer.Write([]string{
		"ID", "Code", "Max Uses", "Use Count", "Expires At", "Note", "Is Active", "Created At",
	})
	for _, ic := range data.InviteCodes {
		expiresAt := ""
		if ic.ExpiresAt != nil {
			expiresAt = ic.ExpiresAt.Format(time.RFC3339)
		}
		writer.Write([]string{
			ic.ID.String(),
			ic.Code,
			fmt.Sprintf("%d", ic.MaxUses),
			fmt.Sprintf("%d", ic.UseCount),
			expiresAt,
			ic.Note,
			fmt.Sprintf("%t", ic.IsActive),
			ic.CreatedAt.Format(time.RFC3339),
		})
	}
}

func writeGroupsCSV(writer *csv.Writer, data *service.UserExportData) {
	// Groups
	writer.Write([]string{"# Groups"})
	writer.Write([]string{"Group ID", "Group Name", "My Role", "Joined At"})
	// Build role map from memberships
	type memberInfo struct {
		Role     string
		JoinedAt string
	}
	roleMap := make(map[string]memberInfo)
	for _, m := range data.GroupMemberships {
		roleMap[m.GroupID.String()] = memberInfo{
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		}
	}
	for _, g := range data.Groups {
		info := roleMap[g.ID.String()]
		writer.Write([]string{
			g.ID.String(),
			g.Name,
			info.Role,
			info.JoinedAt,
		})
	}

	writer.Write([]string{})

	// Shared vehicles (mine)
	if len(data.SharedVehicles) > 0 {
		writer.Write([]string{"# Shared Vehicles (by me)"})
		writer.Write([]string{"ID", "Group ID", "Vehicle ID", "Created At"})
		for _, sv := range data.SharedVehicles {
			writer.Write([]string{
				sv.ID.String(),
				sv.GroupID.String(),
				sv.VehicleID.String(),
				sv.CreatedAt.Format(time.RFC3339),
			})
		}
	}
}

// ==================== Manifest ====================

type exportManifest struct {
	ExportedAt string            `json:"exported_at"`
	Version    string            `json:"version"`
	Modules    map[string]int    `json:"modules"`
	Fields     map[string]string `json:"fields"`
}

func buildManifest(data *service.UserExportData, timestamp string) exportManifest {
	modules := map[string]int{
		"profile":      1,
		"vehicles":     len(data.Vehicles),
		"fuel_records": len(data.FuelRecords),
	}
	if len(data.ExpenseRecords) > 0 {
		modules["expense_records"] = len(data.ExpenseRecords)
	}
	if len(data.Reminders) > 0 {
		modules["reminders"] = len(data.Reminders)
	}
	if len(data.Notifications) > 0 {
		modules["notifications"] = len(data.Notifications)
	}
	if len(data.InviteCodes) > 0 {
		modules["invite_codes"] = len(data.InviteCodes)
	}
	if len(data.Groups) > 0 {
		modules["groups"] = len(data.Groups)
	}
	if len(data.SharedVehicles) > 0 {
		modules["shared_vehicles"] = len(data.SharedVehicles)
	}

	return exportManifest{
		ExportedAt: timestamp,
		Version:    "2.0",
		Modules:    modules,
		Fields: map[string]string{
			"time_format": "RFC3339 / ISO 8601",
			"encoding":    "UTF-8",
		},
	}
}

// ==================== JSON 导出结构 ====================

type jsonExport struct {
	Meta           jsonMeta           `json:"meta"`
	Profile        jsonProfile        `json:"profile"`
	Vehicles       []jsonVehicle      `json:"vehicles"`
	FuelRecords    []jsonFuelRecord   `json:"fuel_records"`
	ExpenseRecords []jsonExpense      `json:"expense_records,omitempty"`
	Reminders      []jsonReminder     `json:"reminders,omitempty"`
	Notifications  []jsonNotification `json:"notifications,omitempty"`
	InviteCodes    []jsonInviteCode   `json:"invite_codes,omitempty"`
	Groups         []jsonGroup        `json:"groups,omitempty"`
	SharedVehicles []jsonShared       `json:"shared_vehicles,omitempty"`
}

type jsonMeta struct {
	ExportedAt string `json:"exported_at"`
	Version    string `json:"version"`
	TimeFormat string `json:"time_format"`
}

type jsonProfile struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	Nickname           string `json:"nickname"`
	Locale             string `json:"locale"`
	Timezone           string `json:"timezone"`
	CountryCode        string `json:"country_code"`
	CurrencyCode       string `json:"currency_code"`
	UnitSystem         string `json:"unit_system"`
	FuelEfficiencyUnit string `json:"fuel_efficiency_unit"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
}

type jsonVehicle struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	VehicleType     string  `json:"vehicle_type"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	Year            int     `json:"year"`
	FuelType        string  `json:"fuel_type"`
	FuelGrade       string  `json:"fuel_grade"`
	TankCapacity    float64 `json:"tank_capacity"`
	BatteryCapacity float64 `json:"battery_capacity"`
	EngineCC        int     `json:"engine_cc"`
	LicensePlate    string  `json:"license_plate"`
	IsDefault       bool    `json:"is_default"`
	IsArchived      bool    `json:"is_archived"`
	CreatedAt       string  `json:"created_at"`
}

type jsonFuelRecord struct {
	ID             string  `json:"id"`
	VehicleID      string  `json:"vehicle_id"`
	RefuelDate     string  `json:"refuel_date"`
	StationName    string  `json:"station_name"`
	FuelAmount     float64 `json:"fuel_amount"`
	FuelUnit       string  `json:"fuel_unit"`
	UnitPrice      float64 `json:"unit_price"`
	TotalCost      float64 `json:"total_cost"`
	CurrencyCode   string  `json:"currency_code"`
	Odometer       float64 `json:"odometer"`
	DistanceUnit   string  `json:"distance_unit"`
	IsFullTank     bool    `json:"is_full_tank"`
	FuelGrade      string  `json:"fuel_grade"`
	TripDistance   float64 `json:"trip_distance"`
	FuelEfficiency float64 `json:"fuel_efficiency"`
	Note           string  `json:"note"`
	CreatedAt      string  `json:"created_at"`
}

type jsonExpense struct {
	ID                  string  `json:"id"`
	VehicleID           string  `json:"vehicle_id"`
	Category            string  `json:"category"`
	MaintenanceCategory string  `json:"maintenance_category,omitempty"`
	Title               string  `json:"title"`
	Amount              float64 `json:"amount"`
	CurrencyCode        string  `json:"currency_code"`
	VendorName          string  `json:"vendor_name,omitempty"`
	Odometer            float64 `json:"odometer,omitempty"`
	DistanceUnit        string  `json:"distance_unit,omitempty"`
	Note                string  `json:"note,omitempty"`
	ExpenseDate         string  `json:"expense_date"`
	ReminderID          string  `json:"reminder_id,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

type jsonReminder struct {
	ID               string  `json:"id"`
	VehicleID        string  `json:"vehicle_id"`
	Type             string  `json:"type"`
	Category         string  `json:"category"`
	Title            string  `json:"title"`
	Description      string  `json:"description,omitempty"`
	Trigger          string  `json:"trigger"`
	MileageInterval  float64 `json:"mileage_interval,omitempty"`
	TimeIntervalDays int     `json:"time_interval_days,omitempty"`
	LastMileage      float64 `json:"last_mileage,omitempty"`
	LastDate         string  `json:"last_date,omitempty"`
	NextMileage      float64 `json:"next_mileage,omitempty"`
	NextDate         string  `json:"next_date,omitempty"`
	IsEnabled        bool    `json:"is_enabled"`
	CreatedAt        string  `json:"created_at"`
}

type jsonNotification struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type jsonInviteCode struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	MaxUses   int    `json:"max_uses"`
	UseCount  int    `json:"use_count"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Note      string `json:"note,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type jsonGroup struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"my_role"`
	JoinedAt string `json:"joined_at"`
}

type jsonShared struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	VehicleID string `json:"vehicle_id"`
	CreatedAt string `json:"created_at"`
}

func buildJSONExport(data *service.UserExportData, timestamp string) jsonExport {
	export := jsonExport{
		Meta: jsonMeta{
			ExportedAt: timestamp,
			Version:    "2.0",
			TimeFormat: "RFC3339 / ISO 8601",
		},
		Profile: jsonProfile{
			ID:                 data.User.ID.String(),
			Email:              data.User.Email,
			Nickname:           data.User.Nickname,
			Locale:             data.User.Locale,
			Timezone:           data.User.Timezone,
			CountryCode:        data.User.CountryCode,
			CurrencyCode:       data.User.CurrencyCode,
			UnitSystem:         data.User.UnitSystem,
			FuelEfficiencyUnit: data.User.FuelEfficiencyUnit,
			Status:             data.User.Status,
			CreatedAt:          data.User.CreatedAt.Format(time.RFC3339),
		},
	}

	// Vehicles
	export.Vehicles = make([]jsonVehicle, len(data.Vehicles))
	for i, v := range data.Vehicles {
		export.Vehicles[i] = jsonVehicle{
			ID:              v.ID.String(),
			Name:            v.Name,
			VehicleType:     string(v.VehicleType),
			Brand:           v.Brand,
			Model:           v.Model,
			Year:            v.Year,
			FuelType:        string(v.FuelType),
			FuelGrade:       v.FuelGrade,
			TankCapacity:    v.TankCapacity,
			BatteryCapacity: v.BatteryCapacity,
			EngineCC:        v.EngineCC,
			LicensePlate:    v.LicensePlate,
			IsDefault:       v.IsDefault,
			IsArchived:      v.IsArchived,
			CreatedAt:       v.CreatedAt.Format(time.RFC3339),
		}
	}

	// Fuel records
	export.FuelRecords = make([]jsonFuelRecord, len(data.FuelRecords))
	for i, r := range data.FuelRecords {
		export.FuelRecords[i] = jsonFuelRecord{
			ID:             r.ID.String(),
			VehicleID:      r.VehicleID.String(),
			RefuelDate:     r.RefuelDate.Format(time.RFC3339),
			StationName:    r.StationName,
			FuelAmount:     r.FuelAmount,
			FuelUnit:       r.FuelUnit,
			UnitPrice:      r.UnitPrice,
			TotalCost:      r.TotalCost,
			CurrencyCode:   r.CurrencyCode,
			Odometer:       r.Odometer,
			DistanceUnit:   r.DistanceUnit,
			IsFullTank:     r.IsFullTank,
			FuelGrade:      r.FuelGrade,
			TripDistance:    r.TripDistance,
			FuelEfficiency: r.FuelEfficiency,
			Note:           r.Note,
			CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		}
	}

	// Expenses
	if len(data.ExpenseRecords) > 0 {
		export.ExpenseRecords = make([]jsonExpense, len(data.ExpenseRecords))
		for i, e := range data.ExpenseRecords {
			reminderID := ""
			if e.ReminderID != nil {
				reminderID = e.ReminderID.String()
			}
			export.ExpenseRecords[i] = jsonExpense{
				ID:                  e.ID.String(),
				VehicleID:           e.VehicleID.String(),
				Category:            string(e.Category),
				MaintenanceCategory: string(e.MaintenanceCategory),
				Title:               e.Title,
				Amount:              e.Amount,
				CurrencyCode:        e.CurrencyCode,
				VendorName:          e.VendorName,
				Odometer:            e.Odometer,
				DistanceUnit:        e.DistanceUnit,
				Note:                e.Note,
				ExpenseDate:         e.ExpenseDate.Format(time.RFC3339),
				ReminderID:          reminderID,
				CreatedAt:           e.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	// Reminders
	if len(data.Reminders) > 0 {
		export.Reminders = make([]jsonReminder, len(data.Reminders))
		for i, rem := range data.Reminders {
			lastDate := ""
			if rem.LastDate != nil {
				lastDate = rem.LastDate.Format(time.RFC3339)
			}
			nextDate := ""
			if rem.NextDate != nil {
				nextDate = rem.NextDate.Format(time.RFC3339)
			}
			export.Reminders[i] = jsonReminder{
				ID:               rem.ID.String(),
				VehicleID:        rem.VehicleID.String(),
				Type:             string(rem.Type),
				Category:         string(rem.Category),
				Title:            rem.Title,
				Description:      rem.Description,
				Trigger:          string(rem.Trigger),
				MileageInterval:  rem.MileageInterval,
				TimeIntervalDays: rem.TimeIntervalDays,
				LastMileage:      rem.LastMileage,
				LastDate:         lastDate,
				NextMileage:      rem.NextMileage,
				NextDate:         nextDate,
				IsEnabled:        rem.IsEnabled,
				CreatedAt:        rem.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	// Notifications
	if len(data.Notifications) > 0 {
		export.Notifications = make([]jsonNotification, len(data.Notifications))
		for i, n := range data.Notifications {
			export.Notifications[i] = jsonNotification{
				ID:        n.ID.String(),
				Type:      string(n.Type),
				Title:     n.Title,
				Message:   n.Message,
				IsRead:    n.IsRead,
				CreatedAt: n.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	// Invite codes
	if len(data.InviteCodes) > 0 {
		export.InviteCodes = make([]jsonInviteCode, len(data.InviteCodes))
		for i, ic := range data.InviteCodes {
			expiresAt := ""
			if ic.ExpiresAt != nil {
				expiresAt = ic.ExpiresAt.Format(time.RFC3339)
			}
			export.InviteCodes[i] = jsonInviteCode{
				ID:        ic.ID.String(),
				Code:      ic.Code,
				MaxUses:   ic.MaxUses,
				UseCount:  ic.UseCount,
				ExpiresAt: expiresAt,
				Note:      ic.Note,
				IsActive:  ic.IsActive,
				CreatedAt: ic.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	// Groups
	if len(data.Groups) > 0 {
		roleMap := make(map[string]struct{ role, joinedAt string })
		for _, m := range data.GroupMemberships {
			roleMap[m.GroupID.String()] = struct{ role, joinedAt string }{
				role:     string(m.Role),
				joinedAt: m.JoinedAt.Format(time.RFC3339),
			}
		}
		export.Groups = make([]jsonGroup, len(data.Groups))
		for i, g := range data.Groups {
			info := roleMap[g.ID.String()]
			export.Groups[i] = jsonGroup{
				ID:       g.ID.String(),
				Name:     g.Name,
				Role:     info.role,
				JoinedAt: info.joinedAt,
			}
		}
	}

	// Shared vehicles
	if len(data.SharedVehicles) > 0 {
		export.SharedVehicles = make([]jsonShared, len(data.SharedVehicles))
		for i, sv := range data.SharedVehicles {
			export.SharedVehicles[i] = jsonShared{
				ID:        sv.ID.String(),
				GroupID:   sv.GroupID.String(),
				VehicleID: sv.VehicleID.String(),
				CreatedAt: sv.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	return export
}
