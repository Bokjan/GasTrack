package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// UserRepo 用户数据访问接口
type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]any) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*model.User, error)
}

// VehicleRepo 车辆数据访问接口
type VehicleRepo interface {
	Create(ctx context.Context, vehicle *model.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Vehicle, error)
	GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Vehicle, error)
	ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]model.Vehicle, error)
	Update(ctx context.Context, vehicle *model.Vehicle) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ClearDefault(ctx context.Context, userID uuid.UUID) error
	DB() *gorm.DB
	ClearDefaultTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	CreateTx(ctx context.Context, tx *gorm.DB, vehicle *model.Vehicle) error
	UpdateTx(ctx context.Context, tx *gorm.DB, vehicle *model.Vehicle) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*model.Vehicle, error)
}

// FuelRecordRepo 加油记录数据访问接口
type FuelRecordRepo interface {
	Create(ctx context.Context, record *model.FuelRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.FuelRecord, error)
	GetByIDAndVehicle(ctx context.Context, id, vehicleID uuid.UUID) (*model.FuelRecord, error)
	ListByVehicle(ctx context.Context, vehicleID uuid.UUID, page, pageSize int) ([]model.FuelRecord, int64, error)
	GetPreviousRecord(ctx context.Context, vehicleID uuid.UUID, beforeDate time.Time) (*model.FuelRecord, error)
	Update(ctx context.Context, record *model.FuelRecord) error
	Delete(ctx context.Context, id, vehicleID uuid.UUID) error
	GetDistinctStationNames(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit int) ([]string, error)
	ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.FuelRecord, error)
	GetVehicleStats(ctx context.Context, vehicleID uuid.UUID) (*StatsResult, error)
	GetCostByCurrency(ctx context.Context, vehicleID uuid.UUID) ([]CostByCurrencyResult, error)
	GetExpensesByMonth(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, start, end time.Time) ([]ExpenseByPeriod, error)
	GetEfficiencyTrend(ctx context.Context, vehicleID uuid.UUID, limit int) ([]model.FuelRecord, error)
	GetStatsByMonth(ctx context.Context, vehicleID uuid.UUID, year int) ([]PeriodStatsResult, error)
	GetStatsByYear(ctx context.Context, vehicleID uuid.UUID) ([]PeriodStatsResult, error)
	GetMultiVehicleStats(ctx context.Context, vehicleIDs []uuid.UUID) (map[uuid.UUID]*MultiVehicleStatsResult, error)
	GetMultiVehicleCostByCurrency(ctx context.Context, vehicleIDs []uuid.UUID) (map[uuid.UUID][]CostByCurrencyResult, error)
}

// GroupRepo 群组数据访问接口
type GroupRepo interface {
	DB() *gorm.DB
	Create(ctx context.Context, group *model.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Group, error)
	GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*model.Group, error)
	GetByInviteCode(ctx context.Context, code string) (*model.Group, error)
	Update(ctx context.Context, group *model.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByInviteCode(ctx context.Context, code string) (bool, error)
	AddMember(ctx context.Context, member *model.GroupMember) error
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*model.GroupMember, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]model.GroupMember, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role model.GroupRole) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	CountMembers(ctx context.Context, groupID uuid.UUID) (int64, error)
	ListGroupsByUser(ctx context.Context, userID uuid.UUID) ([]model.Group, error)
	ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]model.GroupMember, error)
	ListSharedVehiclesByUser(ctx context.Context, userID uuid.UUID) ([]model.SharedVehicle, error)
	JoinGroupByInviteCode(ctx context.Context, code string, userID uuid.UUID) (*model.Group, error)
	GetGroupVehicleSummary(ctx context.Context, groupID uuid.UUID) ([]VehicleSummaryRow, error)
	CreateSharedVehicle(ctx context.Context, sv *model.SharedVehicle) error
	DeleteSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) error
	ListSharedVehiclesByGroup(ctx context.Context, groupID uuid.UUID) ([]model.SharedVehicle, error)
	ExistsSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) (bool, error)
	GetSharedVehicle(ctx context.Context, groupID, vehicleID uuid.UUID) (*model.SharedVehicle, error)
	ListSharedVehiclesForUser(ctx context.Context, userID uuid.UUID) ([]SharedVehicleWithGroup, error)
	IsVehicleSharedToUser(ctx context.Context, vehicleID, userID uuid.UUID) (bool, error)
	GetLeaderboard(ctx context.Context, groupID uuid.UUID, metric string, startDate, endDate time.Time) ([]LeaderboardRow, error)
	GetGroupExpenseByMonth(ctx context.Context, groupID uuid.UUID, year int) ([]GroupExpenseRow, error)
	GetGroupExpenseByYear(ctx context.Context, groupID uuid.UUID) ([]GroupExpenseRow, error)
	GetGroupStationStats(ctx context.Context, groupID uuid.UUID, months int, fuelGrade, sortBy string) ([]StationStatsRow, error)
	GetStationVisitors(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationVisitorRow, error)
	GetStationLatestPrices(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationLatestPriceRow, error)
	GetStationFuelGrades(ctx context.Context, groupID uuid.UUID, stationNames []string, months int) ([]StationFuelGradeRow, error)
}

// ExpenseRecordRepo 开销记录数据访问接口
type ExpenseRecordRepo interface {
	Create(ctx context.Context, record *model.ExpenseRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ExpenseRecord, error)
	GetByIDAndVehicle(ctx context.Context, id, vehicleID uuid.UUID) (*model.ExpenseRecord, error)
	ListByVehicle(ctx context.Context, vehicleID uuid.UUID, page, pageSize int, category string, startDate, endDate *time.Time, keyword string, minAmount, maxAmount float64) ([]model.ExpenseRecord, int64, error)
	Update(ctx context.Context, record *model.ExpenseRecord) error
	Delete(ctx context.Context, id, vehicleID uuid.UUID) error
	ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.ExpenseRecord, error)
	GetTotalsByCurrency(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCurrency, error)
	GetBreakdownByCategory(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByCategory, error)
	GetMonthlyTrend(ctx context.Context, vehicleID uuid.UUID) ([]ExpenseStatsByMonth, error)
	GetLast30DaysTotal(ctx context.Context, vehicleID uuid.UUID) (float64, string, error)
	GetTotalRecords(ctx context.Context, vehicleID uuid.UUID) (int64, error)
	GetDistinctVendorNames(ctx context.Context, userID uuid.UUID, vehicleID *uuid.UUID, limit int) ([]string, error)
}

// RefreshTokenRepo 刷新令牌数据访问接口
type RefreshTokenRepo interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	ConsumeByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// InviteCodeRepo 邀请码数据访问接口
type InviteCodeRepo interface {
	Create(ctx context.Context, invite *model.InviteCode) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InviteCode, error)
	GetByCode(ctx context.Context, code string) (*model.InviteCode, error)
	ListByCreator(ctx context.Context, creatorID uuid.UUID) ([]model.InviteCode, error)
	Update(ctx context.Context, invite *model.InviteCode) error
	Delete(ctx context.Context, id uuid.UUID) error
	ConsumeByCode(ctx context.Context, code string, usedByID uuid.UUID) (*model.InviteCode, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
}

// ReminderRepo 提醒数据访问接口
type ReminderRepo interface {
	Create(ctx context.Context, reminder *model.Reminder) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Reminder, error)
	GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Reminder, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Reminder, error)
	ListByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]model.Reminder, error)
	ListEnabledByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]model.Reminder, error)
	Update(ctx context.Context, reminder *model.Reminder) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

// NotificationRepo 通知数据访问接口
type NotificationRepo interface {
	Create(ctx context.Context, notification *model.Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]model.Notification, error)
	ListAllByUser(ctx context.Context, userID uuid.UUID) ([]model.Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
