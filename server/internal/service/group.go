package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gastrack/internal/dto"
	"gastrack/internal/model"
	"gastrack/internal/pkg/apperror"
	"gastrack/internal/repository"
)

// GroupService 群组业务逻辑
type GroupService struct {
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
	logger    *zap.Logger
}

// NewGroupService 创建 GroupService 实例
func NewGroupService(
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
		logger:    logger,
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
			VehicleID:   row.VehicleID.String(),
			VehicleName: row.VehicleName,
			OwnerName:   ownerName,
			VehicleType: row.VehicleType,
			FuelType:    row.FuelType,
			Records:     row.Records,
			TotalCost:   row.TotalCost,
			TotalFuel:   row.TotalFuel,
			AvgEff:      row.AvgEff,
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
