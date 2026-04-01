package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gastrack/internal/model"
)

// VehicleRepository 车辆数据访问
type VehicleRepository struct {
	db *gorm.DB
}

// NewVehicleRepository 创建实例
func NewVehicleRepository(db *gorm.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

// Create 创建车辆
func (r *VehicleRepository) Create(ctx context.Context, vehicle *model.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

// GetByID 根据 ID 查询车辆（须验证归属用户）
func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// GetByIDAndUser 根据 ID 和用户 ID 查询车辆
func (r *VehicleRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// ListByUser 查询用户的所有车辆（可选是否包含归档）
func (r *VehicleRepository) ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]model.Vehicle, error) {
	var vehicles []model.Vehicle
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}
	err := query.Order("is_default DESC, created_at DESC").Find(&vehicles).Error
	return vehicles, err
}

// Update 更新车辆
func (r *VehicleRepository) Update(ctx context.Context, vehicle *model.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

// Delete 删除车辆（硬删除，会级联删除加油记录）
func (r *VehicleRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Vehicle{}).Error
}

// ClearDefault 清除用户所有车辆的默认标记
func (r *VehicleRepository) ClearDefault(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Vehicle{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

// DB 返回底层 *gorm.DB 实例（用于 Service 层执行事务）
func (r *VehicleRepository) DB() *gorm.DB {
	return r.db
}

// ClearDefaultTx 在指定事务中清除用户所有车辆的默认标记
func (r *VehicleRepository) ClearDefaultTx(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	return tx.WithContext(ctx).Model(&model.Vehicle{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

// CreateTx 在指定事务中创建车辆
func (r *VehicleRepository) CreateTx(ctx context.Context, tx *gorm.DB, vehicle *model.Vehicle) error {
	return tx.WithContext(ctx).Create(vehicle).Error
}

// UpdateTx 在指定事务中更新车辆
func (r *VehicleRepository) UpdateTx(ctx context.Context, tx *gorm.DB, vehicle *model.Vehicle) error {
	return tx.WithContext(ctx).Save(vehicle).Error
}

// CountByUser 统计用户的车辆数量
func (r *VehicleRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Vehicle{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetByIDs 根据多个 ID 批量查询车辆，返回 ID→Vehicle 映射
func (r *VehicleRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*model.Vehicle, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*model.Vehicle), nil
	}

	var vehicles []model.Vehicle
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&vehicles).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*model.Vehicle, len(vehicles))
	for i := range vehicles {
		result[vehicles[i].ID] = &vehicles[i]
	}
	return result, nil
}
