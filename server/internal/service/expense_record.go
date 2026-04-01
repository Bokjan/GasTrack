package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/pkg/convert"
	"gastrack/internal/repository"
)

// ExpenseRecordService 开销记录业务逻辑
type ExpenseRecordService struct {
	expenseRepo  repository.ExpenseRecordRepo
	vehicleRepo  repository.VehicleRepo
	groupRepo    repository.GroupRepo
	reminderRepo repository.ReminderRepo
	logger       *zap.Logger
}

// NewExpenseRecordService 创建 ExpenseRecordService 实例
func NewExpenseRecordService(
	expenseRepo repository.ExpenseRecordRepo,
	vehicleRepo repository.VehicleRepo,
	groupRepo repository.GroupRepo,
	reminderRepo repository.ReminderRepo,
	logger *zap.Logger,
) *ExpenseRecordService {
	return &ExpenseRecordService{
		expenseRepo:  expenseRepo,
		vehicleRepo:  vehicleRepo,
		groupRepo:    groupRepo,
		reminderRepo: reminderRepo,
		logger:       logger,
	}
}

// verifyVehicleAccess 验证用户对车辆的访问权限（所有权或共享访问）
// 返回 true 表示用户是车主
func (s *ExpenseRecordService) verifyVehicleAccess(ctx context.Context, vehicleID, userID uuid.UUID) (isOwner bool, err error) {
	_, err = s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, apperror.ErrInternal("verifying vehicle ownership", err)
	}
	// 不是自己的车辆，检查是否为共享车辆
	if s.groupRepo != nil {
		shared, sharedErr := s.groupRepo.IsVehicleSharedToUser(ctx, vehicleID, userID)
		if sharedErr != nil {
			return false, apperror.ErrInternal("checking shared vehicle access", sharedErr)
		}
		if shared {
			return false, nil
		}
	}
	return false, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
}

// Create 创建开销记录
func (s *ExpenseRecordService) Create(ctx context.Context, userID, vehicleID uuid.UUID, req *dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	// 验证车辆访问权限
	if _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	// 解析日期
	expenseDate, err := time.Parse(time.RFC3339, req.ExpenseDate)
	if err != nil {
		// 尝试短格式
		expenseDate, err = time.Parse("2006-01-02", req.ExpenseDate)
		if err != nil {
			return nil, apperror.ErrBadRequest("expense.invalid_date", "invalid date format, use ISO 8601")
		}
	}

	// 默认单位
	distUnit := string(convert.UnitKm)
	if req.DistanceUnit != "" {
		distUnit = req.DistanceUnit
	}

	// 验证保养类型联动
	if req.Category == "maintenance" && req.MaintenanceCategory == "" {
		return nil, apperror.ErrBadRequest("expense.maintenance_category_required", "maintenance_category is required for maintenance expenses")
	}

	record := &model.ExpenseRecord{
		VehicleID:           vehicleID,
		UserID:              userID,
		Category:            model.ExpenseCategory(req.Category),
		MaintenanceCategory: model.MaintenanceCategory(req.MaintenanceCategory),
		Title:               req.Title,
		Amount:              req.Amount,
		CurrencyCode:        req.CurrencyCode,
		VendorName:          req.VendorName,
		Odometer:            req.Odometer,
		DistanceUnit:        distUnit,
		Note:                req.Note,
		ExpenseDate:         expenseDate,
	}

	// 处理提醒联动
	if req.ReminderID != "" {
		reminderID, err := uuid.Parse(req.ReminderID)
		if err != nil {
			return nil, apperror.ErrBadRequest("expense.invalid_reminder_id", "invalid reminder ID")
		}
		record.ReminderID = &reminderID

		// 更新提醒基准
		if err := s.updateReminderBaseline(ctx, reminderID, record.Odometer, record.DistanceUnit, expenseDate); err != nil {
			s.logger.Warn("failed to update reminder baseline", zap.Error(err))
			// 不阻塞开销记录创建
		}
	}

	if err := s.expenseRepo.Create(ctx, record); err != nil {
		return nil, apperror.ErrInternal("creating expense record", err)
	}

	resp := expenseRecordToResponse(record)
	return &resp, nil
}

// List 获取车辆的开销记录列表（分页+筛选）
func (s *ExpenseRecordService) List(ctx context.Context, userID, vehicleID uuid.UUID, filter *dto.ExpenseListFilter) ([]dto.ExpenseResponse, int64, error) {
	// 验证车辆访问权限
	if _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, 0, err
	}

	// 解析日期筛选
	var startDate, endDate *time.Time
	if filter.StartDate != "" {
		t, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			startDate = &t
		}
	}
	if filter.EndDate != "" {
		t, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			// 将结束日期设为当天末尾
			t = t.Add(24*time.Hour - time.Nanosecond)
			endDate = &t
		}
	}

	records, total, err := s.expenseRepo.ListByVehicle(
		ctx, vehicleID, filter.Page, filter.PageSize,
		filter.Category, startDate, endDate,
		filter.Keyword, filter.MinAmount, filter.MaxAmount,
	)
	if err != nil {
		return nil, 0, apperror.ErrInternal("listing expense records", err)
	}

	result := make([]dto.ExpenseResponse, len(records))
	for i, r := range records {
		result[i] = expenseRecordToResponse(&r)
	}
	return result, total, nil
}

// GetByID 获取开销记录详情
func (s *ExpenseRecordService) GetByID(ctx context.Context, recordID, vehicleID, userID uuid.UUID) (*dto.ExpenseResponse, error) {
	// 验证车辆访问权限
	if _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	record, err := s.expenseRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("expense.not_found", "expense record not found")
		}
		return nil, apperror.ErrInternal("fetching expense record", err)
	}

	resp := expenseRecordToResponse(record)
	return &resp, nil
}

// Update 更新开销记录
func (s *ExpenseRecordService) Update(ctx context.Context, recordID, vehicleID, userID uuid.UUID, req *dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error) {
	// 验证车辆访问权限
	isOwner, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	record, err := s.expenseRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("expense.not_found", "expense record not found")
		}
		return nil, apperror.ErrInternal("fetching expense record", err)
	}

	// 非车主只能编辑自己创建的记录
	if !isOwner && record.UserID != userID {
		return nil, apperror.ErrForbidden("expense.no_permission", "you can only edit your own records")
	}

	// 部分更新
	if req.Category != nil {
		record.Category = model.ExpenseCategory(*req.Category)
	}
	if req.MaintenanceCategory != nil {
		record.MaintenanceCategory = model.MaintenanceCategory(*req.MaintenanceCategory)
	}
	if req.Title != nil {
		record.Title = *req.Title
	}
	if req.Amount != nil {
		record.Amount = *req.Amount
	}
	if req.CurrencyCode != nil {
		record.CurrencyCode = *req.CurrencyCode
	}
	if req.VendorName != nil {
		record.VendorName = *req.VendorName
	}
	if req.Odometer != nil {
		record.Odometer = *req.Odometer
	}
	if req.DistanceUnit != nil {
		record.DistanceUnit = *req.DistanceUnit
	}
	if req.Note != nil {
		record.Note = *req.Note
	}
	if req.ExpenseDate != nil {
		expenseDate, err := time.Parse(time.RFC3339, *req.ExpenseDate)
		if err != nil {
			expenseDate, err = time.Parse("2006-01-02", *req.ExpenseDate)
			if err != nil {
				return nil, apperror.ErrBadRequest("expense.invalid_date", "invalid date format")
			}
		}
		record.ExpenseDate = expenseDate
	}

	// 处理提醒联动变更
	if req.ReminderID != nil {
		if *req.ReminderID == "" {
			record.ReminderID = nil
		} else {
			reminderID, err := uuid.Parse(*req.ReminderID)
			if err != nil {
				return nil, apperror.ErrBadRequest("expense.invalid_reminder_id", "invalid reminder ID")
			}
			record.ReminderID = &reminderID

			// 更新提醒基准
			if err := s.updateReminderBaseline(ctx, reminderID, record.Odometer, record.DistanceUnit, record.ExpenseDate); err != nil {
				s.logger.Warn("failed to update reminder baseline", zap.Error(err))
			}
		}
	}

	// 验证保养类型联动
	if record.Category == model.ExpenseCategoryMaintenance && record.MaintenanceCategory == "" {
		return nil, apperror.ErrBadRequest("expense.maintenance_category_required", "maintenance_category is required for maintenance expenses")
	}

	if err := s.expenseRepo.Update(ctx, record); err != nil {
		return nil, apperror.ErrInternal("updating expense record", err)
	}

	resp := expenseRecordToResponse(record)
	return &resp, nil
}

// Delete 删除开销记录
func (s *ExpenseRecordService) Delete(ctx context.Context, recordID, vehicleID, userID uuid.UUID) error {
	// 验证车辆访问权限
	isOwner, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return err
	}

	record, err := s.expenseRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("expense.not_found", "expense record not found")
		}
		return apperror.ErrInternal("fetching expense record", err)
	}

	// 只有车主或记录创建者可以删除
	if !isOwner && record.UserID != userID {
		return apperror.ErrForbidden("expense.no_permission", "you can only delete your own records")
	}

	if err := s.expenseRepo.Delete(ctx, recordID, vehicleID); err != nil {
		return apperror.ErrInternal("deleting expense record", err)
	}
	return nil
}

// GetStats 获取开销统计
func (s *ExpenseRecordService) GetStats(ctx context.Context, userID, vehicleID uuid.UUID) (*dto.VehicleExpenseStatsResponse, error) {
	// 验证车辆访问权限
	if _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	totalRecords, err := s.expenseRepo.GetTotalRecords(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("getting expense total records", err)
	}

	// 按币种汇总
	currencyTotals, err := s.expenseRepo.GetTotalsByCurrency(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("getting expense currency totals", err)
	}
	dtoTotals := make([]dto.ExpenseCurrencyTotal, len(currencyTotals))
	for i, ct := range currencyTotals {
		dtoTotals[i] = dto.ExpenseCurrencyTotal{
			CurrencyCode: ct.CurrencyCode,
			TotalAmount:  ct.TotalAmount,
			RecordCount:  ct.RecordCount,
		}
	}

	// 按分类统计
	catBreakdown, err := s.expenseRepo.GetBreakdownByCategory(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("getting expense category breakdown", err)
	}
	// 计算总额以得到百分比
	var grandTotal float64
	for _, cb := range catBreakdown {
		grandTotal += cb.TotalAmount
	}
	dtoCats := make([]dto.ExpenseCategoryBreakdown, len(catBreakdown))
	for i, cb := range catBreakdown {
		pct := float64(0)
		if grandTotal > 0 {
			pct = cb.TotalAmount / grandTotal * 100
		}
		dtoCats[i] = dto.ExpenseCategoryBreakdown{
			Category:    cb.Category,
			TotalAmount: cb.TotalAmount,
			RecordCount: cb.RecordCount,
			Percentage:  pct,
		}
	}

	// 月度趋势
	monthlyTrend, err := s.expenseRepo.GetMonthlyTrend(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("getting expense monthly trend", err)
	}
	dtoMonthly := make([]dto.ExpenseMonthlyTrend, len(monthlyTrend))
	for i, mt := range monthlyTrend {
		dtoMonthly[i] = dto.ExpenseMonthlyTrend{
			Period:      mt.Period,
			TotalAmount: mt.TotalAmount,
			RecordCount: mt.RecordCount,
		}
	}

	// 近30天开销
	last30Amount, last30Currency, err := s.expenseRepo.GetLast30DaysTotal(ctx, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("getting expense last 30 days", err)
	}

	return &dto.VehicleExpenseStatsResponse{
		VehicleID:          vehicleID.String(),
		TotalRecords:       totalRecords,
		TotalsByCurrency:   dtoTotals,
		CategoryBreakdown:  dtoCats,
		MonthlyTrend:       dtoMonthly,
		Last30DaysAmount:   last30Amount,
		Last30DaysCurrency: last30Currency,
	}, nil
}

// GetVendorSuggestions 获取商家名称建议列表
func (s *ExpenseRecordService) GetVendorSuggestions(ctx context.Context, userID, vehicleID uuid.UUID) ([]string, error) {
	// 验证车辆访问权限
	if _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	names, err := s.expenseRepo.GetDistinctVendorNames(ctx, userID, &vehicleID, 20)
	if err != nil {
		return nil, apperror.ErrInternal("fetching vendor suggestions", err)
	}
	return names, nil
}

// updateReminderBaseline 更新提醒的上次保养基准并重算下次触发
func (s *ExpenseRecordService) updateReminderBaseline(ctx context.Context, reminderID uuid.UUID, odometer float64, distUnit string, date time.Time) error {
	if s.reminderRepo == nil {
		return nil
	}

	reminder, err := s.reminderRepo.GetByID(ctx, reminderID)
	if err != nil {
		return err
	}

	// 更新上次基准
	reminder.LastDate = &date
	if odometer > 0 {
		// 如果里程单位是 mi，转为 km 存储（提醒系统以 km 为基准）
		odometerKm := odometer
		if convert.DistanceUnit(distUnit) == convert.UnitMile {
			odometerKm = odometer * convert.MileToKm
		}
		reminder.LastMileage = odometerKm
	}

	// 重算下次触发
	s.calculateReminderNext(reminder)

	return s.reminderRepo.Update(ctx, reminder)
}

// calculateReminderNext 根据上次保养基准计算下次触发里程/日期
// 复用 ReminderService 的 calculateNext 逻辑
func (s *ExpenseRecordService) calculateReminderNext(r *model.Reminder) {
	// 按里程
	if r.Trigger == model.ReminderTriggerMileage || r.Trigger == model.ReminderTriggerBoth {
		if r.MileageInterval > 0 {
			r.NextMileage = r.LastMileage + r.MileageInterval
		}
	}

	// 按时间
	if r.Trigger == model.ReminderTriggerTime || r.Trigger == model.ReminderTriggerBoth {
		if r.TimeIntervalDays > 0 && r.LastDate != nil {
			next := r.LastDate.AddDate(0, 0, r.TimeIntervalDays)
			r.NextDate = &next
		}
	}
}

// expenseRecordToResponse 将 model 转为 DTO
func expenseRecordToResponse(r *model.ExpenseRecord) dto.ExpenseResponse {
	resp := dto.ExpenseResponse{
		ID:                  r.ID.String(),
		VehicleID:           r.VehicleID.String(),
		UserID:              r.UserID.String(),
		Category:            string(r.Category),
		MaintenanceCategory: string(r.MaintenanceCategory),
		Title:               r.Title,
		Amount:              r.Amount,
		CurrencyCode:        r.CurrencyCode,
		VendorName:          r.VendorName,
		Odometer:            r.Odometer,
		DistanceUnit:        r.DistanceUnit,
		Note:                r.Note,
		ReceiptURL:          r.ReceiptURL,
		ExpenseDate:         r.ExpenseDate,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}

	if r.ReminderID != nil {
		resp.ReminderID = r.ReminderID.String()
	}

	return resp
}
