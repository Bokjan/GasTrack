package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// UserService 用户业务逻辑
type UserService struct {
	userRepo *repository.UserRepository
	logger   *zap.Logger
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo *repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("user.not_found", "user not found")
		}
		return nil, apperror.ErrInternal("fetching user", err)
	}

	resp := userToResponse(user)
	return &resp, nil
}

// UpdateProfile 更新用户资料（部分更新）
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("user.not_found", "user not found")
		}
		return nil, apperror.ErrInternal("fetching user", err)
	}

	// 应用部分更新（只更新非 nil 字段）
	fields := make(map[string]interface{})
	if req.Nickname != nil {
		fields["nickname"] = *req.Nickname
		user.Nickname = *req.Nickname
	}
	if req.AvatarURL != nil {
		fields["avatar_url"] = *req.AvatarURL
		user.AvatarURL = *req.AvatarURL
	}
	if req.Locale != nil {
		fields["locale"] = *req.Locale
		user.Locale = *req.Locale
	}
	if req.Timezone != nil {
		fields["timezone"] = *req.Timezone
		user.Timezone = *req.Timezone
	}
	if req.CountryCode != nil {
		fields["country_code"] = *req.CountryCode
		user.CountryCode = *req.CountryCode
	}
	if req.CurrencyCode != nil {
		fields["currency_code"] = *req.CurrencyCode
		user.CurrencyCode = *req.CurrencyCode
	}
	if req.ReferenceCurrency != nil {
		fields["reference_currency"] = *req.ReferenceCurrency
		user.ReferenceCurrency = *req.ReferenceCurrency
	}
	if req.UnitSystem != nil {
		fields["unit_system"] = *req.UnitSystem
		user.UnitSystem = *req.UnitSystem
	}
	if req.FuelEfficiencyUnit != nil {
		fields["fuel_efficiency_unit"] = *req.FuelEfficiencyUnit
		user.FuelEfficiencyUnit = *req.FuelEfficiencyUnit
	}

	if len(fields) == 0 {
		resp := userToResponse(user)
		return &resp, nil
	}

	if err := s.userRepo.UpdateFields(ctx, userID, fields); err != nil {
		return nil, apperror.ErrInternal("updating user", err)
	}

	resp := userToResponse(user)
	return &resp, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.ErrInternal("fetching user", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return apperror.ErrBadRequest("user.wrong_password", "current password is incorrect")
	}

	// 生成新密码 hash
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return apperror.ErrInternal("hashing new password", err)
	}

	fields := map[string]interface{}{"password_hash": string(hash)}
	if err := s.userRepo.UpdateFields(ctx, userID, fields); err != nil {
		return apperror.ErrInternal("updating password", err)
	}

	return nil
}

// DeleteAccount 注销账号（软删除）
func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return apperror.ErrInternal("deleting user", err)
	}
	return nil
}
