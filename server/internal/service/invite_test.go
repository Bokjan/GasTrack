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
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	mockrepo "gastrack/internal/repository/mock"
)

func newInviteService(t *testing.T) (*InviteService, *mockrepo.MockInviteCodeRepo, *mockrepo.MockUserRepo) {
	ctrl := gomock.NewController(t)
	inviteRepo := mockrepo.NewMockInviteCodeRepo(ctrl)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	svc := NewInviteService(inviteRepo, userRepo, zap.NewNop())
	return svc, inviteRepo, userRepo
}

func TestInviteService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, inviteRepo, userRepo := newInviteService(t)
		creatorID := uuid.New()

		inviteRepo.EXPECT().ExistsByCode(gomock.Any(), gomock.Any()).Return(false, nil)
		inviteRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		userRepo.EXPECT().GetByID(gomock.Any(), creatorID).Return(&model.User{
			BaseModel: model.BaseModel{ID: creatorID},
			Nickname:  "Creator",
		}, nil)

		resp, err := svc.Create(context.Background(), creatorID, &dto.CreateInviteRequest{MaxUses: 5})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 5, resp.MaxUses)
		assert.True(t, resp.IsActive)
	})

	t.Run("code collision resolved", func(t *testing.T) {
		svc, inviteRepo, userRepo := newInviteService(t)
		creatorID := uuid.New()

		// First call returns exists=true, second returns false
		gomock.InOrder(
			inviteRepo.EXPECT().ExistsByCode(gomock.Any(), gomock.Any()).Return(true, nil),
			inviteRepo.EXPECT().ExistsByCode(gomock.Any(), gomock.Any()).Return(false, nil),
		)
		inviteRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		userRepo.EXPECT().GetByID(gomock.Any(), creatorID).Return(&model.User{
			BaseModel: model.BaseModel{ID: creatorID},
		}, nil)

		resp, err := svc.Create(context.Background(), creatorID, &dto.CreateInviteRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestInviteService_GetByCode(t *testing.T) {
	t.Run("valid code", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		future := time.Now().Add(24 * time.Hour)
		inviteRepo.EXPECT().GetByCode(gomock.Any(), "GT-ABC123").Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: uuid.New()},
			Code:      "GT-ABC123",
			MaxUses:   5,
			UseCount:  2,
			ExpiresAt: &future,
			IsActive:  true,
		}, nil)

		resp, err := svc.GetByCode(context.Background(), "GT-ABC123")
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, 3, resp.RemainingUses)
	})

	t.Run("not found", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteRepo.EXPECT().GetByCode(gomock.Any(), "INVALID").Return(nil, gorm.ErrRecordNotFound)

		resp, err := svc.GetByCode(context.Background(), "INVALID")
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})
}

func TestInviteService_List(t *testing.T) {
	svc, inviteRepo, userRepo := newInviteService(t)
	uid := uuid.New()
	future := time.Now().Add(24 * time.Hour)

	inviteRepo.EXPECT().ListByCreator(gomock.Any(), uid).Return([]model.InviteCode{
		{BaseModel: model.BaseModel{ID: uuid.New()}, Code: "GT-AAA", CreatedBy: uid, MaxUses: 1, IsActive: true, ExpiresAt: &future},
	}, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
		BaseModel: model.BaseModel{ID: uid},
		Nickname:  "Test",
	}, nil)

	result, err := svc.List(context.Background(), uid)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestInviteService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, inviteRepo, userRepo := newInviteService(t)
		uid := uuid.New()
		inviteID := uuid.New()
		isActive := false

		inviteRepo.EXPECT().GetByID(gomock.Any(), inviteID).Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: inviteID},
			CreatedBy: uid,
			IsActive:  true,
		}, nil)
		inviteRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		userRepo.EXPECT().GetByID(gomock.Any(), uid).Return(&model.User{
			BaseModel: model.BaseModel{ID: uid},
		}, nil)

		resp, err := svc.Update(context.Background(), inviteID, uid, &dto.UpdateInviteRequest{IsActive: &isActive})
		assert.NoError(t, err)
		assert.False(t, resp.IsActive)
	})

	t.Run("forbidden - not owner", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteID := uuid.New()

		inviteRepo.EXPECT().GetByID(gomock.Any(), inviteID).Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: inviteID},
			CreatedBy: uuid.New(), // different user
		}, nil)

		_, err := svc.Update(context.Background(), inviteID, uuid.New(), &dto.UpdateInviteRequest{})
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), &dto.UpdateInviteRequest{})
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestInviteService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		uid := uuid.New()
		inviteID := uuid.New()

		inviteRepo.EXPECT().GetByID(gomock.Any(), inviteID).Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: inviteID},
			CreatedBy: uid,
		}, nil)
		inviteRepo.EXPECT().Delete(gomock.Any(), inviteID).Return(nil)

		err := svc.Delete(context.Background(), inviteID, uid)
		assert.NoError(t, err)
	})

	t.Run("forbidden", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteID := uuid.New()

		inviteRepo.EXPECT().GetByID(gomock.Any(), inviteID).Return(&model.InviteCode{
			BaseModel: model.BaseModel{ID: inviteID},
			CreatedBy: uuid.New(),
		}, nil)

		err := svc.Delete(context.Background(), inviteID, uuid.New())
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})
}

func TestInviteService_ValidateAndConsumeInviteCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		uid := uuid.New()
		invite := &model.InviteCode{
			BaseModel: model.BaseModel{ID: uuid.New()},
			Code:      "GT-AAA",
			CreatedBy: uuid.New(),
		}
		inviteRepo.EXPECT().ConsumeByCode(gomock.Any(), "GT-AAA", uid).Return(invite, nil)

		result, err := svc.ValidateAndConsumeInviteCode(context.Background(), "GT-AAA", uid)
		assert.NoError(t, err)
		assert.Equal(t, "GT-AAA", result.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteRepo.EXPECT().ConsumeByCode(gomock.Any(), "INVALID", gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.ValidateAndConsumeInviteCode(context.Background(), "INVALID", uuid.New())
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})

	t.Run("invalid code", func(t *testing.T) {
		svc, inviteRepo, _ := newInviteService(t)
		inviteRepo.EXPECT().ConsumeByCode(gomock.Any(), "GT-EXP", gomock.Any()).Return(nil, errors.New("invite code is not valid"))

		_, err := svc.ValidateAndConsumeInviteCode(context.Background(), "GT-EXP", uuid.New())
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})
}

// Ensure InviteService satisfies InviteServicer interface
var _ InviteServicer = (*InviteService)(nil)
