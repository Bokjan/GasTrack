package service

import (
	"context"
	"errors"
	"fmt"

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
	userRepo repository.UserRepo
	tokens   TokenIssuer
	logger   *zap.Logger
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo repository.UserRepo, tokens TokenIssuer, logger *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, tokens: tokens, logger: logger}
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
	fields := make(map[string]any)
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
// 成功后吊销该用户所有 refresh token（其他设备需重新登录），并为当前设备签发新的 token 对
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) (*dto.AuthResponse, error) {
	if len(req.NewPassword) > maxPasswordBytes {
		return nil, apperror.ErrBadRequest("auth.password_too_long", "password must be at most 72 bytes")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching user", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return nil, apperror.ErrBadRequest("user.wrong_password", "current password is incorrect")
	}

	// 生成新密码 hash
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		return nil, apperror.ErrInternal("hashing new password", err)
	}

	fields := map[string]any{"password_hash": string(hash)}
	if err := s.userRepo.UpdateFields(ctx, userID, fields); err != nil {
		return nil, apperror.ErrInternal("updating password", err)
	}
	user.PasswordHash = string(hash)

	if err := s.tokens.RevokeAllTokens(ctx, userID); err != nil {
		return nil, err
	}
	return s.tokens.IssueTokenPair(ctx, user)
}

// DeleteAccount 注销账号：彻底删除用户及其数据（含 refresh token），详见 UserRepo.PurgeAccount
func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.PurgeAccount(ctx, userID); err != nil {
		return apperror.ErrInternal("deleting user", err)
	}
	return nil
}

// PurgeSoftDeletedUsers 彻底清除旧版本软删除遗留的账号（运维命令调用），返回清除数量
func (s *UserService) PurgeSoftDeletedUsers(ctx context.Context) (int, error) {
	ids, err := s.userRepo.ListSoftDeletedIDs(ctx, "")
	if err != nil {
		return 0, err
	}
	for i, id := range ids {
		if err := s.userRepo.PurgeAccount(ctx, id); err != nil {
			return i, fmt.Errorf("purging user %s: %w", id, err)
		}
	}
	return len(ids), nil
}
