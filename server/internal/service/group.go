package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/pkg/convert"
	"gastrack/internal/repository"
)

// GroupService 群组业务逻辑
type GroupService struct {
	groupRepo   repository.GroupRepo
	userRepo    repository.UserRepo
	vehicleRepo repository.VehicleRepo
	recordRepo  repository.FuelRecordRepo
	rates       RateProvider
	logger      *zap.Logger
}

// NewGroupService 创建 GroupService 实例
func NewGroupService(
	groupRepo repository.GroupRepo,
	userRepo repository.UserRepo,
	vehicleRepo repository.VehicleRepo,
	recordRepo repository.FuelRecordRepo,
	rates RateProvider,
	logger *zap.Logger,
) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		recordRepo:  recordRepo,
		rates:       rates,
		logger:      logger,
	}
}

// viewerConverter 按查看者的偏好币种创建金额换算器
func (s *GroupService) viewerConverter(ctx context.Context, viewerID uuid.UUID) *moneyConverter {
	currency := ""
	if u, err := s.userRepo.GetByID(ctx, viewerID); err == nil {
		currency = u.CurrencyCode
	}
	return newMoneyConverter(s.rates, currency)
}

// sharedVehicles 群组中已共享的车辆（群组统计的数据范围）
func sharedVehicles(rows []repository.GroupVehicleRow) []repository.GroupVehicleRow {
	out := make([]repository.GroupVehicleRow, 0, len(rows))
	for _, r := range rows {
		if r.IsShared {
			out = append(out, r)
		}
	}
	return out
}

func vehicleIDsOf(rows []repository.GroupVehicleRow) []uuid.UUID {
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.VehicleID
	}
	return ids
}

func groupRowEnergyKind(r repository.GroupVehicleRow) energyKind {
	if model.FuelType(r.FuelType) == model.FuelTypeElectric {
		return energyElectric
	}
	return energyLiquid
}

// Create 创建群组
func (s *GroupService) Create(ctx context.Context, ownerID uuid.UUID, req *dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	// 生成唯一邀请码（最多重试 3 次）
	var code string
	for i := 0; i < 3; i++ {
		code = generateGroupInviteCode()
		exists, err := s.groupRepo.ExistsByInviteCode(ctx, code)
		if err != nil {
			return nil, apperror.ErrInternal("checking group invite code", err)
		}
		if !exists {
			break
		}
		if i == 2 {
			return nil, apperror.ErrInternal("generating unique group invite code after 3 retries", nil)
		}
	}

	maxMembers := 10 // 默认上限
	if req.MaxMembers > 0 {
		maxMembers = req.MaxMembers
	}

	group := &model.Group{
		Name:        req.Name,
		OwnerID:     ownerID,
		InviteCode:  code,
		MaxMembers:  maxMembers,
		Description: req.Description,
	}

	// 使用事务：创建群组 + 将创建者作为 owner 成员加入
	err := s.groupRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		member := &model.GroupMember{
			GroupID: group.ID,
			UserID:  ownerID,
			Role:    model.GroupRoleOwner,
		}
		return tx.Create(member).Error
	})
	if err != nil {
		return nil, apperror.ErrInternal("creating group", err)
	}

	return s.buildGroupResponse(ctx, group, ownerID)
}

// GetByID 获取群组详情（仅成员可查看）
func (s *GroupService) GetByID(ctx context.Context, groupID, userID uuid.UUID) (*dto.GroupResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("group.not_found", "group not found")
		}
		return nil, apperror.ErrInternal("fetching group", err)
	}

	return s.buildGroupResponse(ctx, group, userID)
}

// List 获取用户所在的所有群组
func (s *GroupService) List(ctx context.Context, userID uuid.UUID) ([]dto.GroupResponse, error) {
	groups, err := s.groupRepo.ListGroupsByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing groups", err)
	}

	results := make([]dto.GroupResponse, 0, len(groups))
	for i := range groups {
		resp, err := s.buildGroupResponse(ctx, &groups[i], userID)
		if err != nil {
			return nil, err
		}
		results = append(results, *resp)
	}
	return results, nil
}

// Update 更新群组信息（仅 owner/admin 可操作）
func (s *GroupService) Update(ctx context.Context, groupID, userID uuid.UUID, req *dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	// 权限检查
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}
	if member.Role != model.GroupRoleOwner && member.Role != model.GroupRoleAdmin {
		return nil, apperror.ErrForbidden("group.no_permission", "only owner or admin can update group")
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("group.not_found", "group not found")
		}
		return nil, apperror.ErrInternal("fetching group", err)
	}

	// 应用部分更新
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.MaxMembers != nil {
		group.MaxMembers = *req.MaxMembers
	}
	if req.Description != nil {
		group.Description = *req.Description
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, apperror.ErrInternal("updating group", err)
	}

	return s.buildGroupResponse(ctx, group, userID)
}

// Delete 删除群组（仅 owner 可操作）
func (s *GroupService) Delete(ctx context.Context, groupID, userID uuid.UUID) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("group.not_found", "group not found")
		}
		return apperror.ErrInternal("fetching group", err)
	}

	if group.OwnerID != userID {
		return apperror.ErrForbidden("group.not_owner", "only group owner can delete group")
	}

	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		return apperror.ErrInternal("deleting group", err)
	}
	return nil
}

// JoinByInviteCode 通过邀请码加入群组
func (s *GroupService) JoinByInviteCode(ctx context.Context, userID uuid.UUID, code string) (*dto.JoinGroupResponse, error) {
	group, err := s.groupRepo.JoinGroupByInviteCode(ctx, code, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("group.invite_invalid", "invalid invite code")
		}
		if errors.Is(err, repository.ErrAlreadyMember) {
			return nil, apperror.ErrConflict("group.already_member", "you are already a member of this group")
		}
		if errors.Is(err, repository.ErrGroupFull) {
			return nil, apperror.ErrBadRequest("group.full", "group has reached maximum members")
		}
		return nil, apperror.ErrInternal("joining group", err)
	}

	return &dto.JoinGroupResponse{
		GroupID:   group.ID.String(),
		GroupName: group.Name,
		Role:      string(model.GroupRoleMember),
	}, nil
}

// RegenerateInviteCode 重新生成邀请码（仅 owner/admin 可操作）
func (s *GroupService) RegenerateInviteCode(ctx context.Context, groupID, userID uuid.UUID) (*dto.GroupResponse, error) {
	// 权限检查
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}
	if member.Role != model.GroupRoleOwner && member.Role != model.GroupRoleAdmin {
		return nil, apperror.ErrForbidden("group.no_permission", "only owner or admin can regenerate invite code")
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching group", err)
	}

	// 生成新邀请码
	var code string
	for i := 0; i < 3; i++ {
		code = generateGroupInviteCode()
		exists, err := s.groupRepo.ExistsByInviteCode(ctx, code)
		if err != nil {
			return nil, apperror.ErrInternal("checking group invite code", err)
		}
		if !exists {
			break
		}
		if i == 2 {
			return nil, apperror.ErrInternal("generating unique group invite code after 3 retries", nil)
		}
	}

	group.InviteCode = code
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, apperror.ErrInternal("updating group invite code", err)
	}

	return s.buildGroupResponse(ctx, group, userID)
}

// UpdateMemberRole 更新成员角色（仅 owner 可操作）
func (s *GroupService) UpdateMemberRole(ctx context.Context, groupID, targetUserID, operatorID uuid.UUID, req *dto.UpdateMemberRoleRequest) error {
	// 权限检查：只有 owner 可以变更角色
	operatorMember, err := s.groupRepo.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return apperror.ErrInternal("checking group membership", err)
	}
	if operatorMember.Role != model.GroupRoleOwner {
		return apperror.ErrForbidden("group.not_owner", "only group owner can change member roles")
	}

	// 不能修改自己的角色
	if targetUserID == operatorID {
		return apperror.ErrBadRequest("group.self_role", "cannot change your own role")
	}

	// 验证目标成员存在
	_, err = s.groupRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("group.member_not_found", "member not found")
		}
		return apperror.ErrInternal("fetching member", err)
	}

	// 不允许设为 owner（owner 不可转让）
	newRole := model.GroupRole(req.Role)
	if newRole == model.GroupRoleOwner {
		return apperror.ErrBadRequest("group.invalid_role", "cannot assign owner role")
	}

	if err := s.groupRepo.UpdateMemberRole(ctx, groupID, targetUserID, newRole); err != nil {
		return apperror.ErrInternal("updating member role", err)
	}
	return nil
}

// RemoveMember 移除群组成员（owner/admin 可移除，成员可自行退出）
func (s *GroupService) RemoveMember(ctx context.Context, groupID, targetUserID, operatorID uuid.UUID) error {
	// 自行退出
	if targetUserID == operatorID {
		member, err := s.groupRepo.GetMember(ctx, groupID, targetUserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.ErrNotFound("group.not_member", "you are not a member of this group")
			}
			return apperror.ErrInternal("checking membership", err)
		}
		// Owner 不能退出自己的群组（需要先转让或删除群组）
		if member.Role == model.GroupRoleOwner {
			return apperror.ErrBadRequest("group.owner_leave", "owner cannot leave group, transfer ownership or delete the group")
		}
		if err := s.groupRepo.RemoveMember(ctx, groupID, targetUserID); err != nil {
			return apperror.ErrInternal("leaving group", err)
		}
		return nil
	}

	// 移除他人：权限检查
	operatorMember, err := s.groupRepo.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return apperror.ErrInternal("checking membership", err)
	}
	if operatorMember.Role != model.GroupRoleOwner && operatorMember.Role != model.GroupRoleAdmin {
		return apperror.ErrForbidden("group.no_permission", "only owner or admin can remove members")
	}

	// admin 不能移除 owner
	targetMember, err := s.groupRepo.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("group.member_not_found", "member not found")
		}
		return apperror.ErrInternal("fetching member", err)
	}
	if targetMember.Role == model.GroupRoleOwner {
		return apperror.ErrForbidden("group.cannot_remove_owner", "cannot remove group owner")
	}
	// admin 不能移除其他 admin（只有 owner 可以）
	if operatorMember.Role == model.GroupRoleAdmin && targetMember.Role == model.GroupRoleAdmin {
		return apperror.ErrForbidden("group.admin_remove_admin", "admin cannot remove other admins")
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, targetUserID); err != nil {
		return apperror.ErrInternal("removing member", err)
	}
	return nil
}

// GetOverview 获取群组数据汇总
func (s *GroupService) GetOverview(ctx context.Context, groupID, userID uuid.UUID) (*dto.GroupOverviewResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound("group.not_found", "group not found")
		}
		return nil, apperror.ErrInternal("fetching group", err)
	}

	memberCount, err := s.groupRepo.CountMembers(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("counting members", err)
	}

	// 统计范围：已共享到本群组的车辆；查看者自己的未共享车辆也列出，便于开启共享
	rows, err := s.groupRepo.ListGroupVehicles(ctx, groupID, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing group vehicles", err)
	}
	aggRows, err := s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: vehicleIDsOf(rows)})
	if err != nil {
		return nil, apperror.ErrInternal("fetching group vehicle stats", err)
	}
	byVehicle := totalsByVehicle(aggRows)

	// 批量查询所有车辆 owner 信息（1 次 SQL 替代 N 次）
	ownerIDSet := make(map[uuid.UUID]struct{})
	for _, row := range rows {
		ownerIDSet[row.OwnerID] = struct{}{}
	}
	ownerIDs := make([]uuid.UUID, 0, len(ownerIDSet))
	for id := range ownerIDSet {
		ownerIDs = append(ownerIDs, id)
	}
	ownerMap, err := s.userRepo.GetByIDs(ctx, ownerIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching vehicle owners", err)
	}

	conv := s.viewerConverter(ctx, userID)
	sharedCount := 0
	vehicles := make([]dto.GroupVehicleSummary, 0, len(rows))
	for _, row := range rows {
		if row.IsShared {
			sharedCount++
		}
		ownerName := ""
		if owner, ok := ownerMap[row.OwnerID]; ok {
			ownerName = owner.Nickname
		}
		t, ok := byVehicle[row.VehicleID.String()]
		if !ok {
			t = newFuelTotals()
		}
		kind := groupRowEnergyKind(row)

		vehicles = append(vehicles, dto.GroupVehicleSummary{
			VehicleID:    row.VehicleID.String(),
			VehicleName:  row.VehicleName,
			OwnerID:      row.OwnerID.String(),
			OwnerName:    ownerName,
			VehicleType:  row.VehicleType,
			FuelType:     row.FuelType,
			IsShared:     row.IsShared,
			CurrencyCode: conv.base,
			Records:      t.Records,
			TotalCost:    conv.sum(t.costs),
			TotalFuel:    t.amount(kind),
			AvgEff:       t.avgEfficiency(kind),
		})
	}

	return &dto.GroupOverviewResponse{
		GroupID:      group.ID.String(),
		GroupName:    group.Name,
		MemberCount:  int(memberCount),
		VehicleCount: sharedCount,
		Vehicles:     vehicles,
	}, nil
}

// --- 辅助方法 ---

// buildGroupResponse 构建群组响应 DTO
func (s *GroupService) buildGroupResponse(ctx context.Context, group *model.Group, currentUserID uuid.UUID) (*dto.GroupResponse, error) {
	memberCount, err := s.groupRepo.CountMembers(ctx, group.ID)
	if err != nil {
		return nil, apperror.ErrInternal("counting members", err)
	}

	// 获取成员列表
	members, err := s.groupRepo.ListMembers(ctx, group.ID)
	if err != nil {
		return nil, apperror.ErrInternal("listing members", err)
	}

	// 批量查询所有成员的用户信息（1 次 SQL 替代 N 次）
	userIDs := make([]uuid.UUID, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	userMap, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching users", err)
	}

	// 获取 owner 昵称
	ownerName := ""
	if owner, ok := userMap[group.OwnerID]; ok {
		ownerName = owner.Nickname
	}

	// 获取当前用户在此群组的角色
	myRole := ""
	memberDetails := make([]dto.GroupMemberDetail, 0, len(members))
	for _, m := range members {
		if m.UserID == currentUserID {
			myRole = string(m.Role)
		}

		nickname := ""
		email := ""
		if u, ok := userMap[m.UserID]; ok {
			nickname = u.Nickname
			email = u.Email
		}
		memberDetails = append(memberDetails, dto.GroupMemberDetail{
			UserID:   m.UserID.String(),
			Nickname: nickname,
			Email:    email,
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt,
		})
	}

	return &dto.GroupResponse{
		ID:          group.ID.String(),
		Name:        group.Name,
		OwnerID:     group.OwnerID.String(),
		OwnerName:   ownerName,
		InviteCode:  group.InviteCode,
		MaxMembers:  group.MaxMembers,
		Description: group.Description,
		MemberCount: int(memberCount),
		MyRole:      myRole,
		Members:     memberDetails,
		CreatedAt:   group.CreatedAt,
	}, nil
}

// --- 邀请码生成 ---

const groupInviteChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generateGroupInviteCode 生成格式为 GF-XXXXXX 的群组邀请码
func generateGroupInviteCode() string {
	var sb strings.Builder
	sb.WriteString("GF-")
	for i := 0; i < 6; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(groupInviteChars))))
		sb.WriteByte(groupInviteChars[idx.Int64()])
	}
	return sb.String()
}

// ===================== 功能①: 车辆共享标记 =====================

// ShareVehicle 共享车辆到群组
func (s *GroupService) ShareVehicle(ctx context.Context, groupID, userID uuid.UUID, req *dto.ShareVehicleRequest) (*dto.SharedVehicleResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		return nil, apperror.ErrBadRequest("vehicle.invalid_id", "invalid vehicle ID")
	}

	// 验证车辆归属：只有车主才能共享
	vehicle, err := s.vehicleRepo.GetByIDAndUser(ctx, vehicleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("vehicle.not_owner", "you can only share your own vehicles")
		}
		return nil, apperror.ErrInternal("verifying vehicle ownership", err)
	}

	// 检查车辆是否已归档
	if vehicle.IsArchived {
		return nil, apperror.ErrBadRequest("vehicle.archived", "cannot share an archived vehicle")
	}

	// 检查是否已共享
	exists, err := s.groupRepo.ExistsSharedVehicle(ctx, groupID, vehicleID)
	if err != nil {
		return nil, apperror.ErrInternal("checking shared vehicle", err)
	}
	if exists {
		return nil, apperror.ErrConflict("group.vehicle_already_shared", "vehicle is already shared in this group")
	}

	sv := &model.SharedVehicle{
		GroupID:   groupID,
		VehicleID: vehicleID,
		SharedBy:  userID,
	}

	if err := s.groupRepo.CreateSharedVehicle(ctx, sv); err != nil {
		return nil, apperror.ErrInternal("sharing vehicle", err)
	}

	ownerName := ""
	owner, oErr := s.userRepo.GetByID(ctx, userID)
	if oErr == nil {
		ownerName = owner.Nickname
	}

	return &dto.SharedVehicleResponse{
		ID:          sv.ID.String(),
		GroupID:     groupID.String(),
		VehicleID:   vehicleID.String(),
		VehicleName: vehicle.Name,
		OwnerName:   ownerName,
		SharedAt:    sv.CreatedAt.Format(time.RFC3339),
	}, nil
}

// UnshareVehicle 取消车辆共享
func (s *GroupService) UnshareVehicle(ctx context.Context, groupID, vehicleID, userID uuid.UUID) error {
	// 验证成员身份
	member, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return apperror.ErrInternal("checking group membership", err)
	}

	// 获取共享记录
	sv, err := s.groupRepo.GetSharedVehicle(ctx, groupID, vehicleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound("group.shared_vehicle_not_found", "shared vehicle not found")
		}
		return apperror.ErrInternal("fetching shared vehicle", err)
	}

	// 权限：只有车主或群主可以取消共享
	if sv.SharedBy != userID && member.Role != model.GroupRoleOwner {
		return apperror.ErrForbidden("group.no_permission", "only vehicle owner or group owner can unshare")
	}

	if err := s.groupRepo.DeleteSharedVehicle(ctx, groupID, vehicleID); err != nil {
		return apperror.ErrInternal("unsharing vehicle", err)
	}
	return nil
}

// ListSharedVehicles 获取群组内共享车辆列表
func (s *GroupService) ListSharedVehicles(ctx context.Context, groupID, userID uuid.UUID) ([]dto.SharedVehicleResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	svs, err := s.groupRepo.ListSharedVehiclesByGroup(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("listing shared vehicles", err)
	}

	if len(svs) == 0 {
		return []dto.SharedVehicleResponse{}, nil
	}

	// 批量查询所有共享车辆信息和分享者信息（各 1 次 SQL 替代 N 次）
	vehicleIDSet := make(map[uuid.UUID]struct{})
	userIDSet := make(map[uuid.UUID]struct{})
	for _, sv := range svs {
		vehicleIDSet[sv.VehicleID] = struct{}{}
		userIDSet[sv.SharedBy] = struct{}{}
	}
	vehicleIDs := make([]uuid.UUID, 0, len(vehicleIDSet))
	for id := range vehicleIDSet {
		vehicleIDs = append(vehicleIDs, id)
	}
	sharedUserIDs := make([]uuid.UUID, 0, len(userIDSet))
	for id := range userIDSet {
		sharedUserIDs = append(sharedUserIDs, id)
	}
	vehicleMap, err := s.vehicleRepo.GetByIDs(ctx, vehicleIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching shared vehicles", err)
	}
	ownerMap, err := s.userRepo.GetByIDs(ctx, sharedUserIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching shared vehicle owners", err)
	}

	results := make([]dto.SharedVehicleResponse, 0, len(svs))
	for _, sv := range svs {
		vehicleName := ""
		if v, ok := vehicleMap[sv.VehicleID]; ok {
			vehicleName = v.Name
		}

		ownerName := ""
		if u, ok := ownerMap[sv.SharedBy]; ok {
			ownerName = u.Nickname
		}

		results = append(results, dto.SharedVehicleResponse{
			ID:          sv.ID.String(),
			GroupID:     sv.GroupID.String(),
			VehicleID:   sv.VehicleID.String(),
			VehicleName: vehicleName,
			OwnerName:   ownerName,
			SharedAt:    sv.CreatedAt.Format(time.RFC3339),
		})
	}

	return results, nil
}

// ===================== 功能②: 群组油耗排行榜 =====================

// GetLeaderboard 获取群组排行榜
// 只统计已共享到群组的车辆；费用按查看者币种换算；油耗排行只在同一能源类别（energy）内比较
func (s *GroupService) GetLeaderboard(ctx context.Context, groupID, userID uuid.UUID, metric, period, energy string) (*dto.LeaderboardResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching group", err)
	}

	// 计算时间范围
	now := time.Now()
	var startDate, endDate time.Time
	var periodLabel string

	switch period {
	case "last_month":
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		startDate = firstOfThisMonth.AddDate(0, -1, 0)
		endDate = firstOfThisMonth
		periodLabel = startDate.Format("2006-01")
	case "last_3_months":
		endDate = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		startDate = endDate.AddDate(0, -3, 0)
		periodLabel = fmt.Sprintf("%s ~ %s", startDate.Format("2006-01"), now.Format("2006-01"))
	case "current_year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
		periodLabel = fmt.Sprintf("%d", now.Year())
	default: // current_month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0)
		periodLabel = now.Format("2006-01")
	}

	groupVehicles, err := s.groupRepo.ListGroupVehicles(ctx, groupID, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing group vehicles", err)
	}
	shared := sharedVehicles(groupVehicles)
	aggRows, err := s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{
		VehicleIDs: vehicleIDsOf(shared), Start: &startDate, End: &endDate,
	})
	if err != nil {
		return nil, apperror.ErrInternal("fetching leaderboard", err)
	}
	byVehicle := totalsByVehicle(aggRows)

	wantKind := energyLiquid
	if energy == "electric" {
		wantKind = energyElectric
	}
	conv := s.viewerConverter(ctx, userID)

	type entry struct {
		vehicle repository.GroupVehicleRow
		records int64
		value   float64
	}
	entries := make([]entry, 0, len(shared))
	for _, v := range shared {
		t, ok := byVehicle[v.VehicleID.String()]
		// 至少 2 条记录才参与排行
		if !ok || t.Records < 2 {
			continue
		}
		var value float64
		switch metric {
		case "cost":
			value = conv.sum(t.costs)
		case "distance":
			value = t.Distance
		case "frequency":
			value = float64(t.Records)
		default: // efficiency：只比较同一能源类别、且有有效油耗的车辆
			if groupRowEnergyKind(v) != wantKind || t.avgEfficiency(wantKind) <= 0 {
				continue
			}
			value = t.avgEfficiency(wantKind)
		}
		entries = append(entries, entry{vehicle: v, records: t.Records, value: value})
	}

	// 油耗越低越好（升序），其余指标越高排名越前（降序）
	sort.SliceStable(entries, func(i, j int) bool {
		if metric == "cost" || metric == "distance" || metric == "frequency" {
			return entries[i].value > entries[j].value
		}
		return entries[i].value < entries[j].value
	})

	// 计算群组平均值和单位
	var groupAvg float64
	if len(entries) > 0 {
		var sum float64
		for _, e := range entries {
			sum += e.value
		}
		groupAvg = sum / float64(len(entries))
	}

	unit := string(convert.UnitL100km)
	if wantKind == energyElectric {
		unit = "kWh/100km"
	}
	switch metric {
	case "cost", "frequency":
		unit = ""
	case "distance":
		unit = string(convert.UnitKm)
	}

	// 批量查询排行榜所有用户昵称（1 次 SQL 替代 N 次）
	leaderUserIDs := make([]uuid.UUID, len(entries))
	for i, e := range entries {
		leaderUserIDs[i] = e.vehicle.OwnerID
	}
	leaderUserMap, err := s.userRepo.GetByIDs(ctx, leaderUserIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching leaderboard users", err)
	}

	rankings := make([]dto.LeaderboardEntry, 0, len(entries))
	for i, e := range entries {
		diffFromAvg := 0.0
		if groupAvg > 0 {
			diffFromAvg = ((e.value - groupAvg) / groupAvg) * 100
		}

		nickname := ""
		if u, ok := leaderUserMap[e.vehicle.OwnerID]; ok {
			nickname = u.Nickname
		}

		rankings = append(rankings, dto.LeaderboardEntry{
			Rank:        i + 1,
			UserID:      e.vehicle.OwnerID.String(),
			Nickname:    nickname,
			VehicleID:   e.vehicle.VehicleID.String(),
			VehicleName: e.vehicle.VehicleName,
			Value:       e.value,
			DiffFromAvg: diffFromAvg,
			RecordCount: int(e.records),
			IsSelf:      e.vehicle.OwnerID == userID,
		})
	}

	return &dto.LeaderboardResponse{
		GroupID:           groupID.String(),
		GroupName:         group.Name,
		Metric:            metric,
		Energy:            energy,
		Period:            period,
		PeriodLabel:       periodLabel,
		GroupAvg:          groupAvg,
		Unit:              unit,
		Rankings:          rankings,
		TotalParticipants: len(rankings),
	}, nil
}

// ===================== 功能③: 群组费用统计看板 =====================

// GetExpenseStats 获取群组费用统计
func (s *GroupService) GetExpenseStats(ctx context.Context, groupID, userID uuid.UUID, period string, year int) (*dto.GroupExpenseStatsResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching group", err)
	}

	// 统计范围：已共享到本群组的车辆；费用归属车主
	groupVehicles, err := s.groupRepo.ListGroupVehicles(ctx, groupID, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing group vehicles", err)
	}
	shared := sharedVehicles(groupVehicles)
	ownerOf := make(map[uuid.UUID]uuid.UUID, len(shared))
	for _, v := range shared {
		ownerOf[v.VehicleID] = v.OwnerID
	}
	ids := vehicleIDsOf(shared)

	// 获取当前时段和上一时段数据
	var rows, prevRows []repository.FuelAggregateRow
	if period == "year" {
		// 限制最近 10 年的数据，避免全表扫描
		cutoff := time.Now().AddDate(-10, 0, 0)
		rows, err = s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: ids, PeriodFormat: repository.PeriodYear, Start: &cutoff})
		if err != nil {
			return nil, apperror.ErrInternal("fetching expense stats by year", err)
		}
	} else {
		// 月维度
		if year == 0 {
			year = time.Now().Year()
		}
		start, end := yearRange(year)
		rows, err = s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: ids, PeriodFormat: repository.PeriodMonth, Start: start, End: end})
		if err != nil {
			return nil, apperror.ErrInternal("fetching expense stats by month", err)
		}
		// 上一年同期数据
		start, end = yearRange(year - 1)
		prevRows, err = s.recordRepo.Aggregate(ctx, repository.FuelAggregateQuery{VehicleIDs: ids, PeriodFormat: repository.PeriodMonth, Start: start, End: end})
		if err != nil {
			return nil, apperror.ErrInternal("fetching prev expense stats", err)
		}
	}

	// 构建成员昵称映射（批量查询，1 次 SQL 替代 N 次）
	members, _ := s.groupRepo.ListMembers(ctx, groupID)
	memberUserIDs := make([]uuid.UUID, len(members))
	for i, m := range members {
		memberUserIDs[i] = m.UserID
	}
	memberUserMap, _ := s.userRepo.GetByIDs(ctx, memberUserIDs)
	nicknameMap := make(map[string]string, len(members))
	for _, m := range members {
		if u, ok := memberUserMap[m.UserID]; ok {
			nicknameMap[m.UserID.String()] = u.Nickname
		}
	}

	conv := s.viewerConverter(ctx, userID)
	trendItems, memberTotals := buildGroupTrend(rows, ownerOf, nicknameMap, conv)
	prevTrendItems, prevMemberTotals := buildGroupTrend(prevRows, ownerOf, nicknameMap, conv)

	return &dto.GroupExpenseStatsResponse{
		GroupID:               groupID.String(),
		GroupName:             group.Name,
		Period:                period,
		Year:                  year,
		Summary:               groupExpenseSummary(memberTotals, prevMemberTotals, conv),
		CurrencyCode:          conv.base,
		UnconvertedCurrencies: conv.unconvertedCurrencies(),
		MemberBreakdown:       groupMemberBreakdown(memberTotals, nicknameMap, conv),
		TrendItems:            trendItems,
		PrevTrendItems:        prevTrendItems,
	}, nil
}

// buildGroupTrend 将按 车辆 × 币种 × 时段 的聚合行汇总为趋势项，并返回按成员（车主）汇总的合计
func buildGroupTrend(rows []repository.FuelAggregateRow, ownerOf map[uuid.UUID]uuid.UUID, nicknameMap map[string]string, conv *moneyConverter) ([]dto.GroupTrendItem, map[string]*fuelTotals) {
	periodTotals := make(map[string]*fuelTotals)
	periodMembers := make(map[string]map[string]*fuelTotals)
	memberTotals := make(map[string]*fuelTotals)
	var order []string

	for _, row := range rows {
		uid := ownerOf[row.VehicleID].String()
		if periodTotals[row.Period] == nil {
			periodTotals[row.Period] = newFuelTotals()
			periodMembers[row.Period] = make(map[string]*fuelTotals)
			order = append(order, row.Period)
		}
		if periodMembers[row.Period][uid] == nil {
			periodMembers[row.Period][uid] = newFuelTotals()
		}
		if memberTotals[uid] == nil {
			memberTotals[uid] = newFuelTotals()
		}
		periodTotals[row.Period].add(row)
		periodMembers[row.Period][uid].add(row)
		memberTotals[uid].add(row)
	}

	items := make([]dto.GroupTrendItem, 0, len(order))
	for _, p := range order {
		t := periodTotals[p]
		uids := make([]string, 0, len(periodMembers[p]))
		for uid := range periodMembers[p] {
			uids = append(uids, uid)
		}
		sort.Strings(uids)
		byMember := make([]dto.MemberCostItem, 0, len(uids))
		for _, uid := range uids {
			byMember = append(byMember, dto.MemberCostItem{
				UserID:       uid,
				Nickname:     nicknameMap[uid],
				Cost:         conv.sum(periodMembers[p][uid].costs),
				CurrencyCode: conv.base,
			})
		}
		items = append(items, dto.GroupTrendItem{
			PeriodLabel:         p,
			TotalCost:           conv.sum(t.costs),
			TotalFuel:           t.Liters,
			TotalEnergy:         t.Kwh,
			TotalDistance:       t.Distance,
			AvgEfficiency:       t.LiquidEfficiency(),
			AvgEnergyEfficiency: t.ElectricEfficiency(),
			ByMember:            byMember,
		})
	}
	return items, memberTotals
}

// groupExpenseSummary 计算费用摘要及同比变化（金额已换算为同一币种，油耗按里程加权）
func groupExpenseSummary(current, prev map[string]*fuelTotals, conv *moneyConverter) dto.GroupExpenseSummary {
	cur, old := newFuelTotals(), newFuelTotals()
	for _, t := range current {
		cur.merge(t)
	}
	for _, t := range prev {
		old.merge(t)
	}

	pctChange := func(curr, prev float64) float64 {
		if prev == 0 {
			return 0
		}
		return ((curr - prev) / prev) * 100
	}
	curCost, oldCost := conv.sum(cur.costs), conv.sum(old.costs)

	return dto.GroupExpenseSummary{
		TotalCost:           curCost,
		TotalFuel:           cur.Liters,
		TotalEnergy:         cur.Kwh,
		TotalDistance:       cur.Distance,
		AvgEfficiency:       cur.LiquidEfficiency(),
		AvgEnergyEfficiency: cur.ElectricEfficiency(),
		CostChangePct:       pctChange(curCost, oldCost),
		FuelChangePct:       pctChange(cur.Liters, old.Liters),
		DistanceChangePct:   pctChange(cur.Distance, old.Distance),
		EfficiencyChangePct: pctChange(cur.LiquidEfficiency(), old.LiquidEfficiency()),
	}
}

// groupMemberBreakdown 构建成员费用占比（按换算后金额降序）
func groupMemberBreakdown(memberTotals map[string]*fuelTotals, nicknameMap map[string]string, conv *moneyConverter) []dto.MemberCostBreakdown {
	var total float64
	costs := make(map[string]float64, len(memberTotals))
	for uid, t := range memberTotals {
		costs[uid] = conv.sum(t.costs)
		total += costs[uid]
	}

	result := make([]dto.MemberCostBreakdown, 0, len(memberTotals))
	for uid, t := range memberTotals {
		pct := 0.0
		if total > 0 {
			pct = costs[uid] / total * 100
		}
		result = append(result, dto.MemberCostBreakdown{
			UserID:       uid,
			Nickname:     nicknameMap[uid],
			TotalCost:    costs[uid],
			CurrencyCode: conv.base,
			TotalFuel:    t.Liters,
			TotalEnergy:  t.Kwh,
			Percentage:   pct,
		})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].TotalCost > result[j].TotalCost })
	return result
}

// ===================== 功能④: 加油站推荐共享 =====================

// GetStationStats 获取群组加油站推荐共享数据
func (s *GroupService) GetStationStats(ctx context.Context, groupID, userID uuid.UUID, months int, fuelGrade, sortBy string) (*dto.GroupStationStatsResponse, error) {
	// 验证成员身份
	_, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrForbidden("group.not_member", "you are not a member of this group")
		}
		return nil, apperror.ErrInternal("checking group membership", err)
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching group", err)
	}

	// 参数校验
	if months <= 0 || months > 24 {
		months = 6
	}

	// 获取加油站聚合数据
	stationRows, err := s.groupRepo.GetGroupStationStats(ctx, groupID, months, fuelGrade, sortBy)
	if err != nil {
		return nil, apperror.ErrInternal("fetching station stats", err)
	}

	if len(stationRows) == 0 {
		return &dto.GroupStationStatsResponse{
			GroupID:       groupID.String(),
			GroupName:     group.Name,
			TotalStations: 0,
			Stations:      []dto.StationInfo{},
		}, nil
	}

	// 收集站名列表
	stationNames := make([]string, len(stationRows))
	for i, row := range stationRows {
		stationNames[i] = row.StationName
	}

	// 并行获取常客、最新油价和燃油标号
	visitors, err := s.groupRepo.GetStationVisitors(ctx, groupID, stationNames, months)
	if err != nil {
		return nil, apperror.ErrInternal("fetching station visitors", err)
	}

	latestPrices, err := s.groupRepo.GetStationLatestPrices(ctx, groupID, stationNames, months)
	if err != nil {
		return nil, apperror.ErrInternal("fetching station latest prices", err)
	}

	fuelGrades, err := s.groupRepo.GetStationFuelGrades(ctx, groupID, stationNames, months)
	if err != nil {
		return nil, apperror.ErrInternal("fetching station fuel grades", err)
	}

	// 批量查询所有 visitor 的用户信息（1 次 SQL 替代 N 次）
	visitorUserIDSet := make(map[uuid.UUID]struct{})
	for _, v := range visitors {
		visitorUserIDSet[v.UserID] = struct{}{}
	}
	visitorUserIDs := make([]uuid.UUID, 0, len(visitorUserIDSet))
	for id := range visitorUserIDSet {
		visitorUserIDs = append(visitorUserIDs, id)
	}
	visitorUserMap, err := s.userRepo.GetByIDs(ctx, visitorUserIDs)
	if err != nil {
		return nil, apperror.ErrInternal("batch fetching visitor users", err)
	}

	// 构建常客映射
	visitorMap := make(map[string][]dto.StationVisitor)
	for _, v := range visitors {
		nickname := ""
		if u, ok := visitorUserMap[v.UserID]; ok {
			nickname = u.Nickname
		}
		visitorMap[v.StationName] = append(visitorMap[v.StationName], dto.StationVisitor{
			UserID:   v.UserID.String(),
			Nickname: nickname,
			Count:    v.Count,
		})
	}

	// 构建最新油价和趋势映射
	type priceInfo struct {
		latest float64
		trend  string
	}
	// 同一站点可能有多个币种、燃油/充电两种单价，按 站点 × 币种 × 单位 区分
	priceKey := func(station, currency, unit string) string { return station + "\x00" + currency + "\x00" + unit }
	priceMap := make(map[string]*priceInfo)
	for _, p := range latestPrices {
		key := priceKey(p.StationName, p.CurrencyCode, p.PriceUnit)
		if _, ok := priceMap[key]; !ok {
			priceMap[key] = &priceInfo{latest: p.UnitPrice, trend: "stable"}
		} else {
			// row_num=1 是最新的，row_num=2 是上一次的
			pi := priceMap[key]
			if p.UnitPrice > 0 {
				if pi.latest > p.UnitPrice {
					pi.trend = "up"
				} else if pi.latest < p.UnitPrice {
					pi.trend = "down"
				}
			}
		}
	}

	// 构建燃油标号映射
	gradeMap := make(map[string][]string)
	for _, fg := range fuelGrades {
		gradeMap[fg.StationName] = append(gradeMap[fg.StationName], fg.FuelGrade)
	}

	// 组装最终响应
	stations := make([]dto.StationInfo, 0, len(stationRows))
	for _, row := range stationRows {
		latestUnitPrice := row.AvgUnitPrice
		priceTrend := "stable"
		if pi, ok := priceMap[priceKey(row.StationName, row.CurrencyCode, row.PriceUnit)]; ok {
			latestUnitPrice = pi.latest
			priceTrend = pi.trend
		}

		stationVisitors := visitorMap[row.StationName]
		if stationVisitors == nil {
			stationVisitors = []dto.StationVisitor{}
		}

		stationGrades := gradeMap[row.StationName]
		if stationGrades == nil {
			stationGrades = []string{}
		}

		stations = append(stations, dto.StationInfo{
			StationName:     row.StationName,
			AvgUnitPrice:    row.AvgUnitPrice,
			LatestUnitPrice: latestUnitPrice,
			PriceUnit:       row.PriceUnit,
			PriceTrend:      priceTrend,
			CurrencyCode:    row.CurrencyCode,
			VisitCount:      row.VisitCount,
			Visitors:        stationVisitors,
			LatestVisit:     row.LatestVisit,
			FuelGradesSeen:  stationGrades,
		})
	}

	return &dto.GroupStationStatsResponse{
		GroupID:       groupID.String(),
		GroupName:     group.Name,
		TotalStations: len(stations),
		Stations:      stations,
	}, nil
}
