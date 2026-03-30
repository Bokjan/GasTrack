// Package service 包含应用的业务逻辑层。
// Service 层协调 Repository 和外部依赖，实现业务规则。
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gastrack/internal/config"
	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// AuthService 认证业务逻辑
type AuthService struct {
	userRepo            *repository.UserRepository
	tokenRepo           *repository.RefreshTokenRepository
	inviteService       *InviteService
	notificationService *NotificationService
	jwtCfg              *config.JWTConfig
	registrationMode    string // open / invite_only / closed
	logger              *zap.Logger
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.RefreshTokenRepository,
	inviteService *InviteService,
	notificationService *NotificationService,
	jwtCfg *config.JWTConfig,
	registrationMode string,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:            userRepo,
		tokenRepo:           tokenRepo,
		inviteService:       inviteService,
		notificationService: notificationService,
		jwtCfg:              jwtCfg,
		registrationMode:    registrationMode,
		logger:              logger,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 1. 注册模式校验
	switch s.registrationMode {
	case "closed":
		return nil, apperror.ErrForbidden("auth.registration_closed", "registration is currently closed")
	case "invite_only":
		if req.InviteCode == "" {
			return nil, apperror.ErrForbidden("auth.invite_required", "an invite code is required to register")
		}
	case "open":
		// 公开注册，邀请码可选
	default:
		// 未知模式视为邀请制
		if req.InviteCode == "" {
			return nil, apperror.ErrForbidden("auth.invite_required", "an invite code is required to register")
		}
	}

	// 2. 检查邮箱是否已注册
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal("checking email existence", err)
	}
	if exists {
		return nil, apperror.ErrConflict("auth.email_exists", "email already registered")
	}

	// 3. 密码哈希（bcrypt, cost=12）
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, apperror.ErrInternal("hashing password", err)
	}

	// 4. 默认语言
	locale := "en-US"
	if req.Locale != "" {
		locale = req.Locale
	}

	// 根据语言推断默认单位偏好
	fuelUnit, unitSystem, currencyCode := defaultPreferences(locale)

	user := &model.User{
		Email:              req.Email,
		PasswordHash:       string(hash),
		Nickname:           req.Nickname,
		Locale:             locale,
		CurrencyCode:       currencyCode,
		UnitSystem:         unitSystem,
		FuelEfficiencyUnit: fuelUnit,
		Status:             "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		// 捕获数据库 unique violation（并发注册同一邮箱的兜底处理）
		if isDuplicateKeyError(err) {
			return nil, apperror.ErrConflict("auth.email_exists", "email already registered")
		}
		return nil, apperror.ErrInternal("creating user", err)
	}

	// 5. 消费邀请码（在用户创建成功后）
	if req.InviteCode != "" && s.inviteService != nil {
		invite, err := s.inviteService.ValidateAndConsumeInviteCode(ctx, req.InviteCode, user.ID)
		if err != nil {
			// 邀请码消费失败不影响已创建的用户（已进入系统）
			// 仅记录日志告警
			s.logger.Warn("failed to consume invite code after user creation",
				zap.String("invite_code", req.InviteCode),
				zap.String("user_id", user.ID.String()),
				zap.Error(err),
			)
		} else if s.notificationService != nil && invite != nil {
			// 邀请码消费成功，异步通知邀请码创建者
			creatorID := invite.CreatedBy
			code := invite.Code
			nickname := user.Nickname
			go s.notificationService.NotifyInviteUsed(ctx, creatorID, code, nickname)
		}
	}

	// 6. 生成 Token 对
	return s.generateTokenPair(ctx, user)
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized("auth.invalid_credentials", "invalid email or password")
		}
		return nil, apperror.ErrInternal("finding user", err)
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrUnauthorized("auth.invalid_credentials", "invalid email or password")
	}

	// 检查账户状态
	if user.Status != "active" {
		return nil, apperror.ErrForbidden("auth.account_disabled", "account is disabled")
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.generateTokenPair(ctx, user)
}

// RefreshToken 刷新 access token
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// 计算 token hash
	hash := hashToken(req.RefreshToken)

	// 原子性地查找并删除 refresh token（SELECT FOR UPDATE + DELETE）
	// 确保同一个 refresh token 只能被消费一次，防止并发 rotation 竞态
	tokenRecord, err := s.tokenRepo.ConsumeByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized("auth.invalid_refresh_token", "invalid or expired refresh token")
		}
		return nil, apperror.ErrInternal("consuming refresh token", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, apperror.ErrInternal("finding user", err)
	}

	return s.generateTokenPair(ctx, user)
}

// Logout 登出（吊销当前设备的 refresh token）
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	// 删除该用户的所有 refresh token
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

// generateTokenPair 生成 access token + refresh token
func (s *AuthService) generateTokenPair(ctx context.Context, user *model.User) (*dto.AuthResponse, error) {
	now := time.Now()

	// 生成 Access Token (JWT)
	accessClaims := jwt.MapClaims{
		"sub":    user.ID.String(),
		"email":  user.Email,
		"locale": user.Locale,
		"iss":    s.jwtCfg.Issuer,
		"iat":    now.Unix(),
		"exp":    now.Add(s.jwtCfg.AccessExpiration).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, apperror.ErrInternal("signing access token", err)
	}

	// 生成 Refresh Token（随机字符串，非 JWT）
	refreshTokenRaw, err := generateRandomToken(32)
	if err != nil {
		return nil, apperror.ErrInternal("generating refresh token", err)
	}

	// 存储 refresh token 的 hash
	refreshTokenHash := hashToken(refreshTokenRaw)
	refreshRecord := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(s.jwtCfg.RefreshExpiration),
	}

	if err := s.tokenRepo.Create(ctx, refreshRecord); err != nil {
		return nil, apperror.ErrInternal("storing refresh token", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int64(s.jwtCfg.AccessExpiration.Seconds()),
		User:         userToResponse(user),
	}, nil
}

// --- 辅助函数 ---

// generateRandomToken 生成指定长度的随机 hex 字符串
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken 对 token 做 SHA-256 哈希（数据库只存 hash）
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// userToResponse 将 model.User 转为 dto.UserResponse
func userToResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:                 user.ID.String(),
		Email:              user.Email,
		Nickname:           user.Nickname,
		AvatarURL:          user.AvatarURL,
		Locale:             user.Locale,
		Timezone:           user.Timezone,
		CountryCode:        user.CountryCode,
		CurrencyCode:       user.CurrencyCode,
		UnitSystem:         user.UnitSystem,
		FuelEfficiencyUnit: user.FuelEfficiencyUnit,
		Status:             user.Status,
		LastLoginAt:        user.LastLoginAt,
		CreatedAt:          user.CreatedAt,
	}
}

// defaultPreferences 根据语言推断默认偏好
func defaultPreferences(locale string) (fuelUnit, unitSystem, currency string) {
	switch locale {
	case "zh-CN":
		return "L/100km", "metric", "CNY"
	case "ja-JP":
		return "km/L", "metric", "JPY"
	default:
		return "L/100km", "metric", "USD"
	}
}

// isDuplicateKeyError 判断是否为数据库唯一约束冲突错误（PostgreSQL error code 23505）
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "23505") ||
		strings.Contains(errMsg, "unique constraint")
}
