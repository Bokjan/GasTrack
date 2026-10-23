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
	svc := NewGroupService(groupRepo, userRepo, vehicleRepo, mockrepo.NewMockFuelRecordRepo(ctrl), fakeRates{}, zap.NewNop())
	return svc, groupRepo, userRepo, vehicleRepo
}

type groupStatsMocks struct {
	svc    *GroupService
	group  *mockrepo.MockGroupRepo
	user   *mockrepo.MockUserRepo
	record *mockrepo.MockFuelRecordRepo
}

func newGroupStatsService(t *testing.T, gid, viewer uuid.UUID) groupStatsMocks {
	ctrl := gomock.NewController(t)
	m := groupStatsMocks{
		group:  mockrepo.NewMockGroupRepo(ctrl),
		user:   mockrepo.NewMockUserRepo(ctrl),
		record: mockrepo.NewMockFuelRecordRepo(ctrl),
	}
	m.svc = NewGroupService(m.group, m.user, mockrepo.NewMockVehicleRepo(ctrl), m.record, fakeRates{}, zap.NewNop())
	m.group.EXPECT().GetMember(gomock.Any(), gid, viewer).Return(&model.GroupMember{GroupID: gid, UserID: viewer}, nil)
	m.group.EXPECT().GetByID(gomock.Any(), gid).Return(&model.Group{BaseModel: model.BaseModel{ID: gid}, Name: "g"}, nil)
	m.user.EXPECT().GetByID(gomock.Any(), viewer).Return(&model.User{CurrencyCode: "CNY"}, nil).AnyTimes()
	m.user.EXPECT().GetByIDs(gomock.Any(), gomock.Any()).Return(map[uuid.UUID]*model.User{}, nil).AnyTimes()
	return m
}

func TestGroupService_GetLeaderboard(t *testing.T) {
	gid, viewer, other := uuid.New(), uuid.New(), uuid.New()
	gas1, gas2, ev, unshared := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	vehicles := []repository.GroupVehicleRow{
		{VehicleID: gas1, OwnerID: viewer, FuelType: "gasoline", IsShared: true},
		{VehicleID: gas2, OwnerID: other, FuelType: "gasoline", IsShared: true},
		{VehicleID: ev, OwnerID: other, FuelType: "electric", IsShared: true},
		{VehicleID: unshared, OwnerID: viewer, FuelType: "gasoline", IsShared: false},
	}
	agg := func(vid uuid.UUID, cur string, cost float64, a repository.FuelAggregate) repository.FuelAggregateRow {
		return repository.FuelAggregateRow{VehicleID: vid, CurrencyCode: cur, Cost: cost, FuelAggregate: a}
	}
	rows := []repository.FuelAggregateRow{
		agg(gas1, "CNY", 500, repository.FuelAggregate{Records: 3, LiquidEffSum: 800, LiquidEffWeight: 100}),   // 8 L/100km
		agg(gas2, "USD", 14, repository.FuelAggregate{Records: 3, LiquidEffSum: 600, LiquidEffWeight: 100}),    // 6 L/100km, 100 CNY
		agg(ev, "CNY", 300, repository.FuelAggregate{Records: 3, ElectricEffSum: 1500, ElectricEffWeight: 100}), // 15 kWh/100km
	}

	expectData := func(m groupStatsMocks) {
		m.group.EXPECT().ListGroupVehicles(gomock.Any(), gid, viewer).Return(vehicles, nil)
		m.record.EXPECT().Aggregate(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, q repository.FuelAggregateQuery) ([]repository.FuelAggregateRow, error) {
			assert.ElementsMatch(t, []uuid.UUID{gas1, gas2, ev}, q.VehicleIDs, "only shared vehicles are aggregated")
			return rows, nil
		})
	}

	t.Run("efficiency ranks only the requested energy type", func(t *testing.T) {
		m := newGroupStatsService(t, gid, viewer)
		expectData(m)
		resp, err := m.svc.GetLeaderboard(context.Background(), gid, viewer, "efficiency", "current_year", "fuel")
		assert.NoError(t, err)
		assert.Len(t, resp.Rankings, 2)
		assert.Equal(t, gas2.String(), resp.Rankings[0].VehicleID) // 6 < 8，越低越好
		assert.Equal(t, "L/100km", resp.Unit)
	})

	t.Run("electric efficiency", func(t *testing.T) {
		m := newGroupStatsService(t, gid, viewer)
		expectData(m)
		resp, err := m.svc.GetLeaderboard(context.Background(), gid, viewer, "efficiency", "current_year", "electric")
		assert.NoError(t, err)
		assert.Len(t, resp.Rankings, 1)
		assert.Equal(t, "kWh/100km", resp.Unit)
		assert.Equal(t, 15.0, resp.Rankings[0].Value)
	})

	t.Run("cost converted to viewer currency before ranking", func(t *testing.T) {
		m := newGroupStatsService(t, gid, viewer)
		expectData(m)
		resp, err := m.svc.GetLeaderboard(context.Background(), gid, viewer, "cost", "current_year", "fuel")
		assert.NoError(t, err)
		assert.Len(t, resp.Rankings, 3)
		assert.Equal(t, gas1.String(), resp.Rankings[0].VehicleID) // 500 CNY
		assert.InDelta(t, 100, resp.Rankings[2].Value, 0.001)      // 14 USD = 100 CNY
	})
}

func TestGroupService_GetExpenseStats(t *testing.T) {
	gid, viewer, other := uuid.New(), uuid.New(), uuid.New()
	v1, v2 := uuid.New(), uuid.New()
	m := newGroupStatsService(t, gid, viewer)
	m.group.EXPECT().ListGroupVehicles(gomock.Any(), gid, viewer).Return([]repository.GroupVehicleRow{
		{VehicleID: v1, OwnerID: viewer, IsShared: true},
		{VehicleID: v2, OwnerID: other, IsShared: true},
	}, nil)
	m.group.EXPECT().ListMembers(gomock.Any(), gid).Return(nil, nil)
	m.record.EXPECT().Aggregate(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, q repository.FuelAggregateQuery) ([]repository.FuelAggregateRow, error) {
		if q.Start.Year() == 2025 {
			return []repository.FuelAggregateRow{
				{VehicleID: v1, Period: "2025-01", CurrencyCode: "CNY", Cost: 100, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 10}},
			}, nil
		}
		return []repository.FuelAggregateRow{
			{VehicleID: v1, Period: "2026-01", CurrencyCode: "CNY", Cost: 100, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 20, Distance: 200, LiquidEffSum: 1600, LiquidEffWeight: 200}},
			{VehicleID: v2, Period: "2026-01", CurrencyCode: "USD", Cost: 14, FuelAggregate: repository.FuelAggregate{Records: 1, Kwh: 30, Distance: 200, ElectricEffSum: 3000, ElectricEffWeight: 200}},
			{VehicleID: v1, Period: "2026-02", CurrencyCode: "JPY", Cost: 2000, FuelAggregate: repository.FuelAggregate{Records: 1, Liters: 10}},
		}, nil
	}).Times(2)

	resp, err := m.svc.GetExpenseStats(context.Background(), gid, viewer, "month", 2026)
	assert.NoError(t, err)
	assert.Equal(t, "CNY", resp.CurrencyCode)
	assert.InDelta(t, 300, resp.Summary.TotalCost, 0.001) // 100 CNY + 14 USD + 2000 JPY
	assert.InDelta(t, 200, resp.Summary.CostChangePct, 0.001)
	assert.Equal(t, 30.0, resp.Summary.TotalFuel)
	assert.Equal(t, 30.0, resp.Summary.TotalEnergy)
	assert.InDelta(t, 8, resp.Summary.AvgEfficiency, 0.001)
	assert.InDelta(t, 15, resp.Summary.AvgEnergyEfficiency, 0.001)
	assert.Len(t, resp.TrendItems, 2)
	assert.InDelta(t, 200, resp.TrendItems[0].TotalCost, 0.001)
	assert.Len(t, resp.MemberBreakdown, 2)
	assert.Equal(t, viewer.String(), resp.MemberBreakdown[0].UserID) // 200 CNY > 100 CNY
	assert.InDelta(t, 66.667, resp.MemberBreakdown[0].Percentage, 0.01)
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
