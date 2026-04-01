package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
	mockrepo "gastrack/internal/repository/mock"
)

func newGroupService(t *testing.T) (*GroupService, *mockrepo.MockGroupRepo, *mockrepo.MockUserRepo, *mockrepo.MockVehicleRepo) {
	ctrl := gomock.NewController(t)
	groupRepo := mockrepo.NewMockGroupRepo(ctrl)
	userRepo := mockrepo.NewMockUserRepo(ctrl)
	vehicleRepo := mockrepo.NewMockVehicleRepo(ctrl)
	svc := NewGroupService(groupRepo, userRepo, vehicleRepo, zap.NewNop())
	return svc, groupRepo, userRepo, vehicleRepo
}

func TestGroupService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, groupRepo, userRepo, _ := newGroupService(t)
		uid := uuid.New()
		gid := uuid.New()

		groupRepo.EXPECT().GetMember(gomock.Any(), gid, uid).Return(&model.GroupMember{
			GroupID: gid,
			UserID:  uid,
			Role:    model.GroupRoleMember,
		}, nil)
		groupRepo.EXPECT().GetByID(gomock.Any(), gid).Return(&model.Group{
			BaseModel:   model.BaseModel{ID: gid},
			Name:        "Family",
			OwnerID:     uid,
			InviteCode:  "GRP-ABC",
			MaxMembers:  10,
		}, nil)
		groupRepo.EXPECT().ListMembers(gomock.Any(), gid).Return([]model.GroupMember{
			{GroupID: gid, UserID: uid, Role: model.GroupRoleOwner},
		}, nil)
		groupRepo.EXPECT().CountMembers(gomock.Any(), gid).Return(int64(1), nil)
		userRepo.EXPECT().GetByIDs(gomock.Any(), gomock.Any()).Return(map[uuid.UUID]*model.User{
			uid: {BaseModel: model.BaseModel{ID: uid}, Nickname: "Owner", Email: "owner@test.com"},
		}, nil)

		resp, err := svc.GetByID(context.Background(), gid, uid)
		assert.NoError(t, err)
		assert.Equal(t, "Family", resp.Name)
		assert.Equal(t, "owner", resp.MyRole)
	})

	t.Run("not a member", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)
		gid := uuid.New()
		uid := uuid.New()

		groupRepo.EXPECT().GetMember(gomock.Any(), gid, uid).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetByID(context.Background(), gid, uid)
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})
}

func TestGroupService_List(t *testing.T) {
	svc, groupRepo, userRepo, _ := newGroupService(t)
	uid := uuid.New()
	gid := uuid.New()

	groupRepo.EXPECT().ListGroupsByUser(gomock.Any(), uid).Return([]model.Group{
		{BaseModel: model.BaseModel{ID: gid}, Name: "Family", OwnerID: uid, InviteCode: "GRP-ABC", MaxMembers: 10},
	}, nil)
	groupRepo.EXPECT().ListMembers(gomock.Any(), gid).Return([]model.GroupMember{
		{GroupID: gid, UserID: uid, Role: model.GroupRoleOwner},
	}, nil)
	groupRepo.EXPECT().CountMembers(gomock.Any(), gid).Return(int64(1), nil)
	userRepo.EXPECT().GetByIDs(gomock.Any(), gomock.Any()).Return(map[uuid.UUID]*model.User{
		uid: {BaseModel: model.BaseModel{ID: uid}, Nickname: "Owner", Email: "o@t.com"},
	}, nil)

	result, err := svc.List(context.Background(), uid)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGroupService_Update(t *testing.T) {
	t.Run("success as owner", func(t *testing.T) {
		svc, groupRepo, userRepo, _ := newGroupService(t)
		uid := uuid.New()
		gid := uuid.New()
		newName := "Updated"

		groupRepo.EXPECT().GetMember(gomock.Any(), gid, uid).Return(&model.GroupMember{
			GroupID: gid, UserID: uid, Role: model.GroupRoleOwner,
		}, nil)
		groupRepo.EXPECT().GetByID(gomock.Any(), gid).Return(&model.Group{
			BaseModel: model.BaseModel{ID: gid}, Name: "Old", OwnerID: uid, InviteCode: "GRP", MaxMembers: 10,
		}, nil)
		groupRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		groupRepo.EXPECT().ListMembers(gomock.Any(), gid).Return([]model.GroupMember{}, nil)
		groupRepo.EXPECT().CountMembers(gomock.Any(), gid).Return(int64(1), nil)
		userRepo.EXPECT().GetByIDs(gomock.Any(), gomock.Any()).Return(map[uuid.UUID]*model.User{
			uid: {BaseModel: model.BaseModel{ID: uid}, Nickname: "Owner"},
		}, nil)

		resp, err := svc.Update(context.Background(), gid, uid, &dto.UpdateGroupRequest{Name: &newName})
		assert.NoError(t, err)
		assert.Equal(t, "Updated", resp.Name)
	})

	t.Run("forbidden - regular member", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)
		gid := uuid.New()
		uid := uuid.New()

		groupRepo.EXPECT().GetMember(gomock.Any(), gid, uid).Return(&model.GroupMember{
			GroupID: gid, UserID: uid, Role: model.GroupRoleMember,
		}, nil)

		_, err := svc.Update(context.Background(), gid, uid, &dto.UpdateGroupRequest{})
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})
}

func TestGroupService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)
		uid := uuid.New()
		gid := uuid.New()

		groupRepo.EXPECT().GetByID(gomock.Any(), gid).Return(&model.Group{
			BaseModel: model.BaseModel{ID: gid},
			OwnerID:   uid,
		}, nil)
		groupRepo.EXPECT().Delete(gomock.Any(), gid).Return(nil)

		err := svc.Delete(context.Background(), gid, uid)
		assert.NoError(t, err)
	})

	t.Run("not owner", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)
		gid := uuid.New()

		groupRepo.EXPECT().GetByID(gomock.Any(), gid).Return(&model.Group{
			BaseModel: model.BaseModel{ID: gid},
			OwnerID:   uuid.New(), // different owner
		}, nil)

		err := svc.Delete(context.Background(), gid, uuid.New())
		assert.Equal(t, 403, err.(*apperror.AppError).Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)

		groupRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		err := svc.Delete(context.Background(), uuid.New(), uuid.New())
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}

func TestGroupService_JoinByInviteCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)
		uid := uuid.New()
		gid := uuid.New()

		groupRepo.EXPECT().JoinGroupByInviteCode(gomock.Any(), "GRP-ABC", uid).Return(&model.Group{
			BaseModel: model.BaseModel{ID: gid},
			Name:      "Family",
		}, nil)

		resp, err := svc.JoinByInviteCode(context.Background(), uid, "GRP-ABC")
		assert.NoError(t, err)
		assert.Equal(t, "Family", resp.GroupName)
		assert.Equal(t, "member", resp.Role)
	})

	t.Run("already member", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)

		groupRepo.EXPECT().JoinGroupByInviteCode(gomock.Any(), "GRP-ABC", gomock.Any()).Return(nil, repository.ErrAlreadyMember)

		_, err := svc.JoinByInviteCode(context.Background(), uuid.New(), "GRP-ABC")
		assert.Equal(t, 409, err.(*apperror.AppError).Code)
	})

	t.Run("group full", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)

		groupRepo.EXPECT().JoinGroupByInviteCode(gomock.Any(), "GRP-ABC", gomock.Any()).Return(nil, repository.ErrGroupFull)

		_, err := svc.JoinByInviteCode(context.Background(), uuid.New(), "GRP-ABC")
		assert.Equal(t, 400, err.(*apperror.AppError).Code)
	})

	t.Run("invalid code", func(t *testing.T) {
		svc, groupRepo, _, _ := newGroupService(t)

		groupRepo.EXPECT().JoinGroupByInviteCode(gomock.Any(), "INVALID", gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.JoinByInviteCode(context.Background(), uuid.New(), "INVALID")
		assert.Equal(t, 404, err.(*apperror.AppError).Code)
	})
}
