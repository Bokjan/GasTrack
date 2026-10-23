package service

import (
	"context"
	"errors"
	"math"
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

// FuelRecordService 加油记录业务逻辑
type FuelRecordService struct {
	recordRepo          repository.FuelRecordRepo
	vehicleRepo         repository.VehicleRepo
	userRepo            repository.UserRepo
	groupRepo           repository.GroupRepo
	logger              *zap.Logger
	notificationService NotificationServicer
}

// NewFuelRecordService 创建 FuelRecordService 实例
func NewFuelRecordService(
	recordRepo repository.FuelRecordRepo,
	vehicleRepo repository.VehicleRepo,
	userRepo repository.UserRepo,
	groupRepo repository.GroupRepo,
	logger *zap.Logger,
	notificationService NotificationServicer,
) *FuelRecordService {
	return &FuelRecordService{
		recordRepo:          recordRepo,
		vehicleRepo:         vehicleRepo,
		userRepo:            userRepo,
		groupRepo:           groupRepo,
		logger:              logger,
		notificationService: notificationService,
	}
}

// Create 创建加油记录
func (s *FuelRecordService) Create(ctx context.Context, userID, vehicleID uuid.UUID, req *dto.CreateFuelRecordRequest) (*dto.FuelRecordResponse, error) {
	// 验证车辆访问权限
	vehicle, _, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	// 解析加油日期
	refuelDate, err := time.Parse(time.RFC3339, req.RefuelDate)
	if err != nil {
		return nil, apperror.ErrBadRequest("record.invalid_date", "invalid date format, use ISO 8601")
	}

	// 默认单位
	fuelUnit := string(convert.UnitLiter)
	if req.FuelUnit != "" {
		fuelUnit = req.FuelUnit
	}
	distUnit := string(convert.UnitKm)
	if req.DistanceUnit != "" {
		distUnit = req.DistanceUnit
	}

	record := &model.FuelRecord{
		VehicleID:    vehicleID,
		UserID:       userID,
		FuelAmount:   req.FuelAmount,
		FuelUnit:     fuelUnit,
		UnitPrice:    req.UnitPrice,
		TotalCost:    req.TotalCost,
		CurrencyCode: req.CurrencyCode,
		Odometer:     req.Odometer,
		DistanceUnit: distUnit,
		IsFullTank:   req.IsFullTank,
		FuelGrade:    req.FuelGrade,
		StationName:  req.StationName,
		StationLat:   req.StationLat,
		StationLng:   req.StationLng,
		Note:         req.Note,
		RefuelDate:   refuelDate,
	}

	if err := s.recordRepo.Create(ctx, record); err != nil {
		return nil, apperror.ErrInternal("creating fuel record", err)
	}

	// 新记录可能插入到历史中间，重算整车的行驶距离与油耗
	s.recomputeDerived(ctx, vehicleID, record)

	// 异步检查：异常油耗预警 & 保养里程提醒（通知发给车主）
	// 注意：使用 context.WithoutCancel 避免 HTTP 请求结束后 context 被取消导致异步操作失败
	if s.notificationService != nil {
		asyncCtx := context.WithoutCancel(ctx)
		if record.FuelEfficiency > 0 {
			go s.notificationService.CheckFuelAnomaly(asyncCtx, vehicle.UserID, vehicleID, record.ID, record.FuelEfficiency)
		}
		// 将里程转为 km 后检查保养提醒
		odometerKm := record.Odometer
		if convert.DistanceUnit(record.DistanceUnit) == convert.UnitMile {
			odometerKm = record.Odometer * convert.MileToKm
		}
		go s.notificationService.CheckMaintenanceReminders(asyncCtx, vehicleID, odometerKm)
	}

	prefs := s.getUserUnits(ctx, userID)
	resp := fuelRecordToResponse(record, prefs)
	return &resp, nil
}

// List 获取车辆的加油记录列表（分页）
func (s *FuelRecordService) List(ctx context.Context, userID, vehicleID uuid.UUID, page, pageSize int) ([]dto.FuelRecordResponse, int64, error) {
	// 验证车辆访问权限
	if _, _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, 0, err
	}

	records, total, err := s.recordRepo.ListByVehicle(ctx, vehicleID, page, pageSize)
	if err != nil {
		return nil, 0, apperror.ErrInternal("listing fuel records", err)
	}

	prefs := s.getUserUnits(ctx, userID)
	result := make([]dto.FuelRecordResponse, len(records))
	for i, r := range records {
		result[i] = fuelRecordToResponse(&r, prefs)
	}
	return result, total, nil
}

// verifyVehicleAccess 验证用户对车辆的访问权限（所有权或共享访问）
// 返回车辆信息；isOwner 为 true 表示用户是车主
func (s *FuelRecordService) verifyVehicleAccess(ctx context.Context, vehicleID, userID uuid.UUID) (vehicle *model.Vehicle, isOwner bool, err error) {
	vehicle, err = s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err == nil {
		return vehicle, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, apperror.ErrInternal("verifying vehicle ownership", err)
	}
	// 不是自己的车辆，检查是否为共享车辆
	if s.groupRepo != nil {
		shared, sharedErr := s.groupRepo.IsVehicleSharedToUser(ctx, vehicleID, userID)
		if sharedErr != nil {
			return nil, false, apperror.ErrInternal("checking shared vehicle access", sharedErr)
		}
		if shared {
			vehicle, err = s.vehicleRepo.GetByID(ctx, vehicleID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, false, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
				}
				return nil, false, apperror.ErrInternal("fetching shared vehicle", err)
			}
			return vehicle, false, nil
		}
	}
	return nil, false, apperror.ErrNotFound("vehicle.not_found", "vehicle not found")
}

// GetByID 获取加油记录详情
func (s *FuelRecordService) GetByID(ctx context.Context, recordID, vehicleID, userID uuid.UUID) (*dto.FuelRecordResponse, error) {
	// 验证车辆访问权限
	if _, _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	record, err := s.recordRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("record.not_found", "fuel record not found")
		}
		return nil, apperror.ErrInternal("fetching fuel record", err)
	}

	prefs := s.getUserUnits(ctx, userID)
	resp := fuelRecordToResponse(record, prefs)
	return &resp, nil
}

// Update 更新加油记录
func (s *FuelRecordService) Update(ctx context.Context, recordID, vehicleID, userID uuid.UUID, req *dto.UpdateFuelRecordRequest) (*dto.FuelRecordResponse, error) {
	// 验证车辆访问权限
	_, isOwner, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}

	record, err := s.recordRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("record.not_found", "fuel record not found")
		}
		return nil, apperror.ErrInternal("fetching fuel record", err)
	}

	// 非车主只能编辑自己创建的记录
	if !isOwner && record.UserID != userID {
		return nil, apperror.ErrForbidden("record.no_permission", "you can only edit your own records")
	}

	// 部分更新
	if req.FuelAmount != nil {
		record.FuelAmount = *req.FuelAmount
	}
	if req.FuelUnit != nil {
		record.FuelUnit = *req.FuelUnit
	}
	if req.UnitPrice != nil {
		record.UnitPrice = *req.UnitPrice
	}
	if req.TotalCost != nil {
		record.TotalCost = *req.TotalCost
	}
	if req.CurrencyCode != nil {
		record.CurrencyCode = *req.CurrencyCode
	}
	if req.Odometer != nil {
		record.Odometer = *req.Odometer
	}
	if req.DistanceUnit != nil {
		record.DistanceUnit = *req.DistanceUnit
	}
	if req.IsFullTank != nil {
		record.IsFullTank = *req.IsFullTank
	}
	if req.FuelGrade != nil {
		record.FuelGrade = *req.FuelGrade
	}
	if req.StationName != nil {
		record.StationName = *req.StationName
	}
	if req.StationLat != nil {
		record.StationLat = *req.StationLat
	}
	if req.StationLng != nil {
		record.StationLng = *req.StationLng
	}
	if req.Note != nil {
		record.Note = *req.Note
	}
	if req.RefuelDate != nil {
		refuelDate, err := time.Parse(time.RFC3339, *req.RefuelDate)
		if err != nil {
			return nil, apperror.ErrBadRequest("record.invalid_date", "invalid date format")
		}
		record.RefuelDate = refuelDate
	}

	if err := s.recordRepo.Update(ctx, record); err != nil {
		return nil, apperror.ErrInternal("updating fuel record", err)
	}

	// 里程、油量、日期、是否加满的变化都会影响前后记录，重算整车
	s.recomputeDerived(ctx, vehicleID, record)

	prefs := s.getUserUnits(ctx, userID)
	resp := fuelRecordToResponse(record, prefs)
	return &resp, nil
}

// Delete 删除加油记录
func (s *FuelRecordService) Delete(ctx context.Context, recordID, vehicleID, userID uuid.UUID) error {
	// 验证车辆访问权限
	_, isOwner, err := s.verifyVehicleAccess(ctx, vehicleID, userID)
	if err != nil {
		return err
	}

	// 获取记录详情以检查创建者
	record, err := s.recordRepo.GetByIDAndVehicle(ctx, recordID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("record.not_found", "fuel record not found")
		}
		return apperror.ErrInternal("fetching fuel record", err)
	}

	// 只有车主或记录创建者可以删除
	if !isOwner && record.UserID != userID {
		return apperror.ErrForbidden("record.no_permission", "you can only delete your own records")
	}

	if err := s.recordRepo.Delete(ctx, recordID, vehicleID); err != nil {
		return apperror.ErrInternal("deleting fuel record", err)
	}

	// 删除后下一条记录的行驶距离、下一次满油的油耗都需要重算
	s.recomputeDerived(ctx, vehicleID, nil)
	return nil
}

// GetStationSuggestions 获取加油站/充电站名称建议列表
func (s *FuelRecordService) GetStationSuggestions(ctx context.Context, userID, vehicleID uuid.UUID) ([]string, error) {
	// 验证车辆访问权限
	if _, _, err := s.verifyVehicleAccess(ctx, vehicleID, userID); err != nil {
		return nil, err
	}

	names, err := s.recordRepo.GetDistinctStationNames(ctx, userID, &vehicleID, 20)
	if err != nil {
		return nil, apperror.ErrInternal("fetching station suggestions", err)
	}

	return names, nil
}

// userUnitPrefs 用户的单位偏好
type userUnitPrefs struct {
	efficiencyUnit convert.FuelEfficiencyUnit
	volumeUnit     convert.VolumeUnit
	distanceUnit   convert.DistanceUnit
}

// getUserUnits 获取用户的完整单位偏好
func (s *FuelRecordService) getUserUnits(ctx context.Context, userID uuid.UUID) userUnitPrefs {
	prefs := userUnitPrefs{
		efficiencyUnit: convert.UnitL100km,
		volumeUnit:     convert.UnitLiter,
		distanceUnit:   convert.UnitKm,
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return prefs
	}

	if user.FuelEfficiencyUnit != "" {
		prefs.efficiencyUnit = convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
	}
	if user.UnitSystem == "imperial" {
		prefs.volumeUnit = convert.UnitGallon
		prefs.distanceUnit = convert.UnitMile
	}

	return prefs
}

// recomputeDerived 重算车辆全部记录的派生字段，并把 target（如有）的结果回填到调用方持有的对象上。
// 记录本身已保存成功，重算失败只记录日志：派生值会在该车下一次增删改时被修正。
func (s *FuelRecordService) recomputeDerived(ctx context.Context, vehicleID uuid.UUID, target *model.FuelRecord) {
	records, err := s.recordRepo.RecomputeDerived(ctx, vehicleID, computeDerivedFields)
	if err != nil {
		s.logger.Error("failed to recompute fuel record derived fields",
			zap.String("vehicle_id", vehicleID.String()),
			zap.Error(err),
		)
		return
	}
	if target == nil {
		return
	}
	for i := range records {
		if records[i].ID == target.ID {
			target.TripDistance = records[i].TripDistance
			target.FuelEfficiency = records[i].FuelEfficiency
			return
		}
	}
}

// computeDerivedFields 按时间升序计算每条记录的派生字段，返回发生变化的记录下标。
//   - 行驶距离：与上一条记录（不论是否加满）的里程差，统一为 km
//   - 油耗：满油到满油法。当前记录加满时，取上一次加满以来（不含上次加满本身）所有加油量之和，
//     除以两次加满之间的里程差，得到 L/100km（电车为 kWh/100km）；未加满或无上一次加满时为 0
func computeDerivedFields(records []model.FuelRecord) []int {
	var changed []int
	lastFullKm := -1.0 // 上一次加满时的里程（km），-1 表示尚未出现加满记录
	fuelSinceFull := 0.0

	for i := range records {
		r := &records[i]
		liters, km := convert.NormalizeToMetric(r.FuelAmount, convert.VolumeUnit(r.FuelUnit), r.Odometer, convert.DistanceUnit(r.DistanceUnit))

		tripDistance := 0.0
		if i > 0 {
			prev := records[i-1]
			_, prevKm := convert.NormalizeToMetric(0, convert.VolumeUnit(prev.FuelUnit), prev.Odometer, convert.DistanceUnit(prev.DistanceUnit))
			if d := km - prevKm; d > 0 {
				tripDistance = math.Round(d*10) / 10
			}
		}

		efficiency := 0.0
		fuelSinceFull += liters
		if r.IsFullTank {
			if lastFullKm >= 0 {
				if dist := km - lastFullKm; dist > 0 {
					efficiency = convert.CalcL100km(fuelSinceFull, dist)
				}
			}
			lastFullKm = km
			fuelSinceFull = 0
		}

		if math.Abs(r.TripDistance-tripDistance) > 1e-6 || math.Abs(r.FuelEfficiency-efficiency) > 1e-6 {
			r.TripDistance = tripDistance
			r.FuelEfficiency = efficiency
			changed = append(changed, i)
		}
	}
	return changed
}

// fuelRecordToResponse 将 model 转为 DTO，按用户偏好转换所有单位
func fuelRecordToResponse(r *model.FuelRecord, prefs userUnitPrefs) dto.FuelRecordResponse {
	// 油耗转换
	efficiency := r.FuelEfficiency
	if efficiency > 0 && prefs.efficiencyUnit != "" {
		efficiency = convert.ConvertFuelEfficiency(efficiency, convert.UnitL100km, prefs.efficiencyUnit)
	}

	// 加油量转换（从记录原始单位 → 用户偏好单位）
	fuelAmount := r.FuelAmount
	unitPrice := r.UnitPrice
	srcFuelUnit := convert.VolumeUnit(r.FuelUnit)
	// kWh 不参与容量换算
	if srcFuelUnit == convert.UnitLiter || srcFuelUnit == convert.UnitGallon {
		fuelAmount = convert.ConvertVolume(r.FuelAmount, srcFuelUnit, prefs.volumeUnit)
		// 单价也需同步转换：单价是 "货币/容量单位"，容量单位变了单价也要变
		// 例：$3.50/gal → ¥0.92/L（单价 × 反向容量比率）
		if srcFuelUnit != prefs.volumeUnit && unitPrice > 0 {
			// 1 单位目标 = N 单位源 → 单价(目标) = 单价(源) × N
			unitPrice = unitPrice * convert.ConvertVolume(1, prefs.volumeUnit, srcFuelUnit)
		}
	}

	// 里程 / 行驶距离转换（从记录原始单位 → 用户偏好单位）
	srcDistUnit := convert.DistanceUnit(r.DistanceUnit)
	odometer := convert.ConvertDistance(r.Odometer, srcDistUnit, prefs.distanceUnit)
	tripDistance := r.TripDistance
	if tripDistance > 0 {
		// TripDistance 在 computeDerivedFields 中已统一转为 km 存储
		tripDistance = convert.ConvertDistance(tripDistance, convert.UnitKm, prefs.distanceUnit)
	}

	// 返回时使用用户偏好的单位标识
	respFuelUnit := r.FuelUnit
	if srcFuelUnit == convert.UnitLiter || srcFuelUnit == convert.UnitGallon {
		respFuelUnit = string(prefs.volumeUnit)
	}
	respDistUnit := string(prefs.distanceUnit)

	return dto.FuelRecordResponse{
		ID:             r.ID.String(),
		VehicleID:      r.VehicleID.String(),
		FuelAmount:     fuelAmount,
		FuelUnit:       respFuelUnit,
		UnitPrice:      unitPrice,
		TotalCost:      r.TotalCost,
		CurrencyCode:   r.CurrencyCode,
		Odometer:       odometer,
		DistanceUnit:   respDistUnit,
		IsFullTank:     r.IsFullTank,
		FuelGrade:      r.FuelGrade,
		StationName:    r.StationName,
		StationLat:     r.StationLat,
		StationLng:     r.StationLng,
		Note:           r.Note,
		ReceiptURL:     r.ReceiptURL,
		TripDistance:    tripDistance,
		FuelEfficiency: efficiency,
		RefuelDate:     r.RefuelDate,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
