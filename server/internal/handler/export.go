package handler

import (
	"encoding/csv"
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

// ExportMyData 导出当前用户的全部数据为 CSV
// GET /api/v1/users/me/export
func (h *ExportHandler) ExportMyData(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	data, err := h.exportService.GatherUserData(r.Context(), userID)
	if err != nil {
		handleAppError(w, h.logger, err)
		return
	}

	// 设置 CSV 下载响应头
	filename := fmt.Sprintf("gastrack-export-%s.csv", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	// 写入 UTF-8 BOM，确保 Excel 正确识别中文
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// === Section 1: User Profile ===
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
	writer.Write([]string{}) // 空行分隔

	// === Section 2: Vehicles ===
	writer.Write([]string{"# Vehicles"})
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
	writer.Write([]string{}) // 空行分隔

	// === Section 3: Fuel / Charging Records ===
	writer.Write([]string{"# Fuel / Charging Records"})
	writer.Write([]string{
		"ID", "Vehicle ID", "Refuel Date", "Station Name",
		"Fuel Amount", "Fuel Unit", "Unit Price", "Total Cost", "Currency",
		"Odometer", "Distance Unit", "Is Full Tank", "Fuel Grade",
		"Trip Distance", "Fuel Efficiency", "Note", "Created At",
	})
	for _, rec := range data.Records {
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
