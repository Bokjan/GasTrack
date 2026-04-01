package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	mockrepo "gastrack/internal/repository/mock"
)

func newUserService(t *testing.T) (*UserService, *mockrepo.MockUserRepo) {
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockUserRepo(ctrl)
	svc := NewUserService(mockRepo, zap.NewNop())
	return svc, mockRepo
}

func TestUserService_GetProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		user := &model.User{
			BaseModel: model.BaseModel{ID: uid, CreatedAt: time.Now()},
			Email:     "test@example.com",
			Nickname:  "Test",
			Status:    "active",
		}
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(user, nil)

		resp, err := svc.GetProfile(context.Background(), uid)
		assert.NoError(t, err)
		assert.Equal(t, uid.String(), resp.ID)
		assert.Equal(t, "test@example.com", resp.Email)
	})

	t.Run("not found", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(nil, gorm.ErrRecordNotFound)

		resp, err := svc.GetProfile(context.Background(), uid)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})

	t.Run("internal error", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(nil, errors.New("db error"))

		resp, err := svc.GetProfile(context.Background(), uid)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, 500, err.(*apperror.AppError).Code)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		user := &model.User{
			BaseModel: model.BaseModel{ID: uid},
			Email:     "test@example.com",
			Nickname:  "Old",
			Status:    "active",
		}
		newNickname := "New"
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(user, nil)
		mockRepo.EXPECT().UpdateFields(gomock.Any(), uid, gomock.Any()).Return(nil)

		resp, err := svc.UpdateProfile(context.Background(), uid, &dto.UpdateUserRequest{Nickname: &newNickname})
		assert.NoError(t, err)
		assert.Equal(t, "New", resp.Nickname)
	})

	t.Run("no fields to update", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		user := &model.User{
			BaseModel: model.BaseModel{ID: uid},
			Nickname:  "Same",
			Status:    "active",
		}
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(user, nil)

		resp, err := svc.UpdateProfile(context.Background(), uid, &dto.UpdateUserRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "Same", resp.Nickname)
	})

	t.Run("user not found", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(nil, gorm.ErrRecordNotFound)

		resp, err := svc.UpdateProfile(context.Background(), uid, &dto.UpdateUserRequest{})
		assert.Nil(t, resp)
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass123"), 12)
		user := &model.User{
			BaseModel:    model.BaseModel{ID: uid},
			PasswordHash: string(hash),
		}
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(user, nil)
		mockRepo.EXPECT().UpdateFields(gomock.Any(), uid, gomock.Any()).Return(nil)

		err := svc.ChangePassword(context.Background(), uid, &dto.ChangePasswordRequest{
			OldPassword: "oldpass123",
			NewPassword: "newpass456",
		})
		assert.NoError(t, err)
	})

	t.Run("wrong old password", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass123"), 12)
		user := &model.User{
			BaseModel:    model.BaseModel{ID: uid},
			PasswordHash: string(hash),
		}
		mockRepo.EXPECT().GetByID(gomock.Any(), uid).Return(user, nil)

		err := svc.ChangePassword(context.Background(), uid, &dto.ChangePasswordRequest{
			OldPassword: "wrongpass",
			NewPassword: "newpass456",
		})
		assert.Error(t, err)
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})
}

func TestUserService_DeleteAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		mockRepo.EXPECT().Delete(gomock.Any(), uid).Return(nil)

		err := svc.DeleteAccount(context.Background(), uid)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		svc, mockRepo := newUserService(t)
		uid := uuid.New()
		mockRepo.EXPECT().Delete(gomock.Any(), uid).Return(errors.New("db error"))

		err := svc.DeleteAccount(context.Background(), uid)
		assert.Error(t, err)
		assert.Equal(t, 500, err.(*apperror.AppError).Code)
	})
}
