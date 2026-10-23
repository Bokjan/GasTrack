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
	"sync"
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

// maxPasswordBytes bcrypt 支持的最大密码字节数
const maxPasswordBytes = 72

// bcryptCost 密码哈希成本
const bcryptCost = 12

// AuthService 认证业务逻辑
type AuthService struct {
	userRepo            repository.UserRepo
	tokenRepo           repository.RefreshTokenRepo
	inviteService       InviteServicer
	notificationService NotificationServicer
	jwtCfg              *config.JWTConfig
	registrationMode    string // open / invite_only / closed
	logger              *zap.Logger
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(
	userRepo repository.UserRepo,
	tokenRepo repository.RefreshTokenRepo,
	inviteService InviteServicer,
	notificationService NotificationServicer,
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

	// bcrypt 只处理前 72 字节，超长会报错；validator 的 max=72 按字符计，多字节字符需额外校验
	if len(req.Password) > maxPasswordBytes {
		return nil, apperror.ErrBadRequest("auth.password_too_long", "password must be at most 72 bytes")
	}

	// 邮箱统一去空白并小写，避免大小写不同被视为两个账号
	req.Email = NormalizeEmail(req.Email)

	// 2. 检查邮箱是否已注册
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal("checking email existence", err)
	}
	if exists {
		return nil, apperror.ErrConflict("auth.email_exists", "email already registered")
	}

	// 3. 密码哈希（bcrypt, cost=12）
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
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

	// 5. 创建用户；携带邀请码时，校验 + 建用户 + 消费邀请码在同一事务内完成
	var invite *model.InviteCode
	if req.InviteCode != "" {
		if s.inviteService == nil {
			return nil, apperror.ErrInternal("invite service not configured", nil)
		}
		invite, err = s.inviteService.CreateUserWithInvite(ctx, user, req.InviteCode)
	} else {
		err = s.userRepo.Create(ctx, user)
	}
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		// 捕获数据库 unique violation（并发注册同一邮箱的兜底处理）
		if isDuplicateKeyError(err) {
			return nil, apperror.ErrConflict("auth.email_exists", "email already registered")
		}
		return nil, apperror.ErrInternal("creating user", err)
	}

	// 邀请码消费成功，异步通知邀请码创建者（脱离请求 context，避免响应结束后被取消）
	if invite != nil && s.notificationService != nil {
		go s.notificationService.NotifyInviteUsed(context.WithoutCancel(ctx), invite.CreatedBy, invite.Code, user.Nickname)
	}

	// 6. 生成 Token 对
	return s.IssueTokenPair(ctx, user)
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, NormalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在时同样执行一次 bcrypt 比较，使响应耗时与"密码错误"一致，防止通过耗时枚举邮箱
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash(), []byte(req.Password))
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

	return s.IssueTokenPair(ctx, user)
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

	// 获取用户（已注销的用户查不到，视为 token 无效）
	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized("auth.invalid_refresh_token", "invalid or expired refresh token")
		}
		return nil, apperror.ErrInternal("finding user", err)
	}

	// 被停用的账户不允许续期
	if user.Status != "active" {
		return nil, apperror.ErrForbidden("auth.account_disabled", "account is disabled")
	}

	return s.IssueTokenPair(ctx, user)
}

// Logout 登出（吊销该用户所有设备的 refresh token）
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.RevokeAllTokens(ctx, userID)
}

// RevokeAllTokens 吊销用户所有 refresh token（登出、改密码、注销账号时使用）
func (s *AuthService) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	if err := s.tokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return apperror.ErrInternal("revoking refresh tokens", err)
	}
	return nil
}

// PurgeExpiredTokens 清理过期的 refresh token（由后台定时任务调用）
func (s *AuthService) PurgeExpiredTokens(ctx context.Context) error {
	return s.tokenRepo.DeleteExpired(ctx)
}

// IssueTokenPair 生成 access token + refresh token
func (s *AuthService) IssueTokenPair(ctx context.Context, user *model.User) (*dto.AuthResponse, error) {
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

// NormalizeEmail 邮箱规范化：去除首尾空白并转为小写
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var (
	dummyHashOnce sync.Once
	dummyHash     []byte
)

// dummyPasswordHash 返回与真实密码同 cost 的占位 hash（首次使用时生成）
func dummyPasswordHash() []byte {
	dummyHashOnce.Do(func() {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("gastrack-timing-equalizer"), bcryptCost)
	})
	return dummyHash
}

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
		ReferenceCurrency:  user.ReferenceCurrency,
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

// isDuplicateKeyError 判断是否为数据库唯一约束冲突错误
// 依赖 gorm.Config.TranslateError，PostgreSQL 23505 会被翻译为 gorm.ErrDuplicatedKey
func isDuplicateKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
