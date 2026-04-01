package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"

	"gastrack/internal/config"
	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	mockrepo "gastrack/internal/repository/mock"
	mocksvc "gastrack/internal/service/mock"
)

func mustHashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4) // cost=4 for fast tests
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func newAuthService(t *testing.T, mode string) (*AuthService, *mockrepo.MockUserRepo, *mockrepo.MockRefreshTokenRepo, *mocksvc.MockInviteServicer, *mocksvc.MockNotificationServicer) {
	ctrl := gomock.NewController(t)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	tokenRepo := mockrepo.NewMockRefreshTokenRepo(ctrl)
	inviteSvc := mocksvc.NewMockInviteServicer(ctrl)
	notifSvc := mocksvc.NewMockNotificationServicer(ctrl)

	jwtCfg := &config.JWTConfig{
		Secret:            "test-secret-key-32-bytes-long!!!", // 32 bytes
		Issuer:            "gastrack-test",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, tokenRepo, inviteSvc, notifSvc, jwtCfg, mode, zap.NewNop())
	return svc, userRepo, tokenRepo, inviteSvc, notifSvc
}

func TestAuthService_Register(t *testing.T) {
	t.Run("open mode success", func(t *testing.T) {
		svc, userRepo, tokenRepo, _, _ := newAuthService(t, "open")

		userRepo.EXPECT().ExistsByEmail(gomock.Any(), "new@example.com").Return(false, nil)
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		tokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Email:    "new@example.com",
			Password: "password123",
			Nickname: "Newbie",
			Locale:   "zh-CN",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, "new@example.com", resp.User.Email)
	})

	t.Run("closed mode rejects", func(t *testing.T) {
		svc, _, _, _, _ := newAuthService(t, "closed")

		_, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Email:    "new@example.com",
			Password: "password123",
			Nickname: "Newbie",
		})
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})

	t.Run("invite_only mode without code", func(t *testing.T) {
		svc, _, _, _, _ := newAuthService(t, "invite_only")

		_, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Email:    "new@example.com",
			Password: "password123",
			Nickname: "Newbie",
		})
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})

	t.Run("invite_only mode with valid code", func(t *testing.T) {
		svc, userRepo, tokenRepo, inviteSvc, notifSvc := newAuthService(t, "invite_only")
		creatorID := uuid.New()

		userRepo.EXPECT().ExistsByEmail(gomock.Any(), "invited@example.com").Return(false, nil)
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		inviteSvc.EXPECT().ValidateAndConsumeInviteCode(gomock.Any(), "GT-ABC123", gomock.Any()).Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: uuid.New()},
			Code:      "GT-ABC123",
			CreatedBy: creatorID,
		}, nil)
		notifSvc.EXPECT().NotifyInviteUsed(gomock.Any(), creatorID, "GT-ABC123", "Invited").Return()
		tokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Email:      "invited@example.com",
			Password:   "password123",
			Nickname:   "Invited",
			InviteCode: "GT-ABC123",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Give goroutine time to execute
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("email already exists", func(t *testing.T) {
		svc, userRepo, _, _, _ := newAuthService(t, "open")
		userRepo.EXPECT().ExistsByEmail(gomock.Any(), "exists@example.com").Return(true, nil)

		_, err := svc.Register(context.Background(), &dto.RegisterRequest{
			Email:    "exists@example.com",
			Password: "password123",
			Nickname: "Exists",
		})
		assert.Equal(t, 409, err.(*apperror.AppError).Code)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo, tokenRepo, _, _ := newAuthService(t, "open")
		uid := uuid.New()
		hash := mustHashPassword("password123")

		userRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.User{
			BaseModel:    model.BaseModel{ID: uid},
			Email:        "test@example.com",
			PasswordHash: hash,
			Status:       "active",
		}, nil)
		userRepo.EXPECT().UpdateLastLogin(gomock.Any(), uid).Return(nil)
		tokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Login(context.Background(), &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, userRepo, _, _, _ := newAuthService(t, "open")
		userRepo.EXPECT().GetByEmail(gomock.Any(), "noone@example.com").Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.Login(context.Background(), &dto.LoginRequest{
			Email:    "noone@example.com",
			Password: "password123",
		})
		assert.Equal(t, 401, err.(*apperror.AppError).Code)
	})

	t.Run("wrong password", func(t *testing.T) {
		svc, userRepo, _, _, _ := newAuthService(t, "open")
		hash := mustHashPassword("password123")

		userRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.User{
			BaseModel:    model.BaseModel{ID: uuid.New()},
			PasswordHash: hash,
			Status:       "active",
		}, nil)

		_, err := svc.Login(context.Background(), &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpass",
		})
		assert.Equal(t, 401, err.(*apperror.AppError).Code)
	})

	t.Run("account disabled", func(t *testing.T) {
		svc, userRepo, _, _, _ := newAuthService(t, "open")
		hash := mustHashPassword("password123")

		userRepo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(&model.User{
			BaseModel:    model.BaseModel{ID: uuid.New()},
			PasswordHash: hash,
			Status:       "suspended",
		}, nil)

		_, err := svc.Login(context.Background(), &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, userRepo, tokenRepo, _, _ := newAuthService(t, "open")
		uid := uuid.New()

		tokenRepo.EXPECT().ConsumeByTokenHash(gomock.Any(), gomock.Any()).Return(&model.RefreshToken{
			BaseModel: model.BaseModel{ID: uuid.New()},
			UserID:    uid,
		}, nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
			BaseModel: model.BaseModel{ID: uid},
			Email:     "test@example.com",
			Status:    "active",
		}, nil)
		tokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
			RefreshToken: "some-refresh-token",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("invalid token", func(t *testing.T) {
		svc, _, tokenRepo, _, _ := newAuthService(t, "open")
		tokenRepo.EXPECT().ConsumeByTokenHash(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		})
		assert.Equal(t, 401, err.(*apperror.AppError).Code)
	})
}

func TestAuthService_Logout(t *testing.T) {
	svc, _, tokenRepo, _, _ := newAuthService(t, "open")
	uid := uuid.New()
	tokenRepo.EXPECT().DeleteByUserID(gomock.Any(), uid).Return(nil)

	err := svc.Logout(context.Background(), uid)
	assert.NoError(t, err)
}
