package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
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
	groupRepo   *repository.GroupRepository
	userRepo    *repository.UserRepository
	vehicleRepo *repository.VehicleRepository
	logger      *zap.Logger
}

// NewGroupService 创建 GroupService 实例
func NewGroupService(
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
	vehicleRepo *repository.VehicleRepository,
	logger *zap.Logger,
) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		logger:      logger,
	}
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

	rows, err := s.groupRepo.GetGroupVehicleSummary(ctx, groupID)
	if err != nil {
		return nil, apperror.ErrInternal("fetching group vehicle summary", err)
	}

	vehicles := make([]dto.GroupVehicleSummary, 0, len(rows))
	for _, row := range rows {
		ownerName := ""
		owner, oErr := s.userRepo.GetByID(ctx, row.OwnerID)
		if oErr == nil {
			ownerName = owner.Nickname
		}

		vehicles = append(vehicles, dto.GroupVehicleSummary{
			VehicleID:    row.VehicleID.String(),
			VehicleName:  row.VehicleName,
			OwnerID:      row.OwnerID.String(),
			OwnerName:    ownerName,
			VehicleType:  row.VehicleType,
			FuelType:     row.FuelType,
			CurrencyCode: row.CurrencyCode,
			Records:      row.Records,
			TotalCost:    row.TotalCost,
			TotalFuel:    row.TotalFuel,
			AvgEff:       row.AvgEff,
		})
	}

	return &dto.GroupOverviewResponse{
		GroupID:      group.ID.String(),
		GroupName:    group.Name,
		MemberCount:  int(memberCount),
		VehicleCount: len(vehicles),
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

	// 获取 owner 昵称
	ownerName := ""
	owner, oErr := s.userRepo.GetByID(ctx, group.OwnerID)
	if oErr == nil {
		ownerName = owner.Nickname
	}

	// 获取当前用户在此群组的角色
	myRole := ""
	member, mErr := s.groupRepo.GetMember(ctx, group.ID, currentUserID)
	if mErr == nil {
		myRole = string(member.Role)
	}

	// 获取成员详情列表
	members, err := s.groupRepo.ListMembers(ctx, group.ID)
	if err != nil {
		return nil, apperror.ErrInternal("listing members", err)
	}

	memberDetails := make([]dto.GroupMemberDetail, 0, len(members))
	for _, m := range members {
		user, uErr := s.userRepo.GetByID(ctx, m.UserID)
		nickname := ""
		email := ""
		if uErr == nil {
			nickname = user.Nickname
			email = user.Email
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

	results := make([]dto.SharedVehicleResponse, 0, len(svs))
	for _, sv := range svs {
		vehicleName := ""
		vehicle, vErr := s.vehicleRepo.GetByID(ctx, sv.VehicleID)
		if vErr == nil {
			vehicleName = vehicle.Name
		}

		ownerName := ""
		owner, oErr := s.userRepo.GetByID(ctx, sv.SharedBy)
		if oErr == nil {
			ownerName = owner.Nickname
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
func (s *GroupService) GetLeaderboard(ctx context.Context, groupID, userID uuid.UUID, metric, period string) (*dto.LeaderboardResponse, error) {
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

	rows, err := s.groupRepo.GetLeaderboard(ctx, groupID, metric, startDate, endDate)
	if err != nil {
		return nil, apperror.ErrInternal("fetching leaderboard", err)
	}

	// 计算群组平均值和单位
	var groupAvg float64
	unit := string(convert.UnitL100km)
	if len(rows) > 0 {
		var sum float64
		for _, row := range rows {
			switch metric {
			case "cost":
				sum += row.TotalCost
			case "distance":
				sum += row.TotalDistance
			case "frequency":
				sum += float64(row.RecordCount)
			default:
				sum += row.AvgEfficiency
			}
		}
		groupAvg = sum / float64(len(rows))
	}

	switch metric {
	case "cost":
		unit = ""
	case "distance":
		unit = string(convert.UnitKm)
	case "frequency":
		unit = ""
	}

	rankings := make([]dto.LeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		var value float64
		switch metric {
		case "cost":
			value = row.TotalCost
		case "distance":
			value = row.TotalDistance
		case "frequency":
			value = float64(row.RecordCount)
		default:
			value = row.AvgEfficiency
		}

		diffFromAvg := 0.0
		if groupAvg > 0 {
			diffFromAvg = ((value - groupAvg) / groupAvg) * 100
		}

		nickname := ""
		u, uErr := s.userRepo.GetByID(ctx, row.UserID)
		if uErr == nil {
			nickname = u.Nickname
		}

		rankings = append(rankings, dto.LeaderboardEntry{
			Rank:        i + 1,
			UserID:      row.UserID.String(),
			Nickname:    nickname,
			VehicleID:   row.VehicleID.String(),
			VehicleName: row.VehicleName,
			Value:       value,
			DiffFromAvg: diffFromAvg,
			RecordCount: row.RecordCount,
			IsSelf:      row.UserID == userID,
		})
	}

	return &dto.LeaderboardResponse{
		GroupID:           groupID.String(),
		GroupName:         group.Name,
		Metric:            metric,
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

	// 获取当前时段和上一时段数据
	var rows, prevRows []repository.GroupExpenseRow

	if period == "year" {
		rows, err = s.groupRepo.GetGroupExpenseByYear(ctx, groupID)
		if err != nil {
			return nil, apperror.ErrInternal("fetching expense stats by year", err)
		}
	} else {
		// 月维度
		if year == 0 {
			year = time.Now().Year()
		}
		rows, err = s.groupRepo.GetGroupExpenseByMonth(ctx, groupID, year)
		if err != nil {
			return nil, apperror.ErrInternal("fetching expense stats by month", err)
		}
		// 上一年同期数据
		prevRows, err = s.groupRepo.GetGroupExpenseByMonth(ctx, groupID, year-1)
		if err != nil {
			return nil, apperror.ErrInternal("fetching prev expense stats", err)
		}
	}

	// 构建成员昵称映射
	members, _ := s.groupRepo.ListMembers(ctx, groupID)
	nicknameMap := make(map[string]string)
	for _, m := range members {
		u, uErr := s.userRepo.GetByID(ctx, m.UserID)
		if uErr == nil {
			nicknameMap[m.UserID.String()] = u.Nickname
		}
	}

	// 聚合当前时段
	trendItems, memberTotals := s.buildTrendItems(rows, nicknameMap)

	// 聚合上一时段
	prevTrendItems, prevMemberTotals := s.buildTrendItems(prevRows, nicknameMap)

	// 计算汇总
	summary := s.calculateSummary(memberTotals, prevMemberTotals)

	// 计算成员费用占比（用原始费用总和计算百分比，而非 summary.TotalCost 的占位 0）
	var rawTotalCost float64
	for _, mt := range memberTotals {
		for _, cost := range mt.CostByCurrency {
			rawTotalCost += cost
		}
	}
	memberBreakdown := s.buildMemberBreakdown(memberTotals, nicknameMap, rawTotalCost)

	return &dto.GroupExpenseStatsResponse{
		GroupID:         groupID.String(),
		GroupName:       group.Name,
		Period:          period,
		Year:            year,
		Summary:         summary,
		MemberBreakdown: memberBreakdown,
		TrendItems:      trendItems,
		PrevTrendItems:  prevTrendItems,
	}, nil
}

// memberTotal 成员总计
type memberTotal struct {
	CostByCurrency map[string]float64 // currency_code -> cost
	Fuel           float64
	Distance       float64
	Efficiency     float64
	Count          int
}

// buildTrendItems 从查询行构建趋势项
func (s *GroupService) buildTrendItems(rows []repository.GroupExpenseRow, nicknameMap map[string]string) ([]dto.GroupTrendItem, map[string]*memberTotal) {
	// period -> list of rows
	periodData := make(map[string][]*repository.GroupExpenseRow)
	periodOrder := make([]string, 0)
	memberTotals := make(map[string]*memberTotal)

	for i := range rows {
		row := &rows[i]
		if _, ok := periodData[row.PeriodLabel]; !ok {
			periodOrder = append(periodOrder, row.PeriodLabel)
		}
		periodData[row.PeriodLabel] = append(periodData[row.PeriodLabel], row)

		uid := row.UserID.String()
		if _, ok := memberTotals[uid]; !ok {
			memberTotals[uid] = &memberTotal{CostByCurrency: make(map[string]float64)}
		}
		mt := memberTotals[uid]
		mt.CostByCurrency[row.CurrencyCode] += row.Cost
		mt.Fuel += row.Fuel
		mt.Distance += row.Distance
		if row.AvgEfficiency > 0 {
			mt.Efficiency += row.AvgEfficiency
			mt.Count++
		}
	}

	trendItems := make([]dto.GroupTrendItem, 0, len(periodOrder))
	for _, pl := range periodOrder {
		rowList := periodData[pl]
		var totalFuel, totalDist, totalEff float64
		var effCount int
		byMember := make([]dto.MemberCostItem, 0)

		for _, row := range rowList {
			totalFuel += row.Fuel
			totalDist += row.Distance
			if row.AvgEfficiency > 0 {
				totalEff += row.AvgEfficiency
				effCount++
			}
			// 每个 (user, currency) 组合生成一个 MemberCostItem
			byMember = append(byMember, dto.MemberCostItem{
				UserID:       row.UserID.String(),
				Nickname:     nicknameMap[row.UserID.String()],
				Cost:         row.Cost,
				CurrencyCode: row.CurrencyCode,
			})
		}

		avgEff := 0.0
		if effCount > 0 {
			avgEff = totalEff / float64(effCount)
		}

		trendItems = append(trendItems, dto.GroupTrendItem{
			PeriodLabel:   pl,
			TotalCost:     0, // 前端按 by_member 各币种换算后汇总
			TotalFuel:     totalFuel,
			TotalDistance:  totalDist,
			AvgEfficiency: avgEff,
			ByMember:      byMember,
		})
	}

	return trendItems, memberTotals
}

// calculateSummary 计算费用摘要及环比变化
func (s *GroupService) calculateSummary(current, prev map[string]*memberTotal) dto.GroupExpenseSummary {
	var totalCost, totalFuel, totalDist, totalEff float64
	var effCount int

	for _, mt := range current {
		for _, cost := range mt.CostByCurrency {
			totalCost += cost
		}
		totalFuel += mt.Fuel
		totalDist += mt.Distance
		if mt.Count > 0 {
			totalEff += mt.Efficiency
			effCount += mt.Count
		}
	}

	avgEff := 0.0
	if effCount > 0 {
		avgEff = totalEff / float64(effCount)
	}

	// 上一时段汇总
	var prevCost, prevFuel, prevDist, prevEff float64
	var prevEffCount int
	for _, mt := range prev {
		for _, cost := range mt.CostByCurrency {
			prevCost += cost
		}
		prevFuel += mt.Fuel
		prevDist += mt.Distance
		if mt.Count > 0 {
			prevEff += mt.Efficiency
			prevEffCount += mt.Count
		}
	}
	prevAvgEff := 0.0
	if prevEffCount > 0 {
		prevAvgEff = prevEff / float64(prevEffCount)
	}

	// 计算环比变化百分比
	pctChange := func(curr, prev float64) float64 {
		if prev == 0 {
			return 0
		}
		return ((curr - prev) / prev) * 100
	}

	return dto.GroupExpenseSummary{
		TotalCost:           0, // 前端根据 member_breakdown 各币种换算后汇总
		TotalFuel:           totalFuel,
		TotalDistance:        totalDist,
		AvgEfficiency:       avgEff,
		CostChangePct:       pctChange(totalCost, prevCost),
		FuelChangePct:       pctChange(totalFuel, prevFuel),
		DistanceChangePct:   pctChange(totalDist, prevDist),
		EfficiencyChangePct: pctChange(avgEff, prevAvgEff),
	}
}

// buildMemberBreakdown 构建成员费用占比
func (s *GroupService) buildMemberBreakdown(memberTotals map[string]*memberTotal, nicknameMap map[string]string, totalCost float64) []dto.MemberCostBreakdown {
	result := make([]dto.MemberCostBreakdown, 0)
	for uid, mt := range memberTotals {
		// 每个币种生成一条记录
		for cur, cost := range mt.CostByCurrency {
			pct := 0.0
			if totalCost > 0 {
				pct = (cost / totalCost) * 100
			}
			result = append(result, dto.MemberCostBreakdown{
				UserID:       uid,
				Nickname:     nicknameMap[uid],
				TotalCost:    cost,
				CurrencyCode: cur,
				TotalFuel:    mt.Fuel,
				Percentage:   pct,
			})
		}
	}
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

	// 构建常客映射
	visitorMap := make(map[string][]dto.StationVisitor)
	for _, v := range visitors {
		nickname := ""
		u, uErr := s.userRepo.GetByID(ctx, v.UserID)
		if uErr == nil {
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
	priceMap := make(map[string]*priceInfo)
	for _, p := range latestPrices {
		if _, ok := priceMap[p.StationName]; !ok {
			priceMap[p.StationName] = &priceInfo{latest: p.UnitPrice, trend: "stable"}
		} else {
			// row_num=1 是最新的，row_num=2 是上一次的
			pi := priceMap[p.StationName]
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
		if pi, ok := priceMap[row.StationName]; ok {
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
