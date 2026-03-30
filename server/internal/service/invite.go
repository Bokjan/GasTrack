package service

import (
	"context"
	"crypto/rand"
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
	"gastrack/internal/repository"
)

// InviteService 邀请码业务逻辑
type InviteService struct {
	inviteRepo *repository.InviteCodeRepository
	userRepo   *repository.UserRepository
	logger     *zap.Logger
}

// NewInviteService 创建 InviteService 实例
func NewInviteService(
	inviteRepo *repository.InviteCodeRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *InviteService {
	return &InviteService{
		inviteRepo: inviteRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// Create 创建邀请码
func (s *InviteService) Create(ctx context.Context, creatorID uuid.UUID, req *dto.CreateInviteRequest) (*dto.InviteCodeResponse, error) {
	// 生成唯一邀请码（最多重试 3 次）
	var code string
	for i := 0; i < 3; i++ {
		code = generateInviteCode()
		exists, err := s.inviteRepo.ExistsByCode(ctx, code)
		if err != nil {
			return nil, apperror.ErrInternal("checking invite code", err)
		}
		if !exists {
			break
		}
		if i == 2 {
			return nil, apperror.ErrInternal("generating unique invite code after 3 retries", nil)
		}
	}

	maxUses := 1
	if req.MaxUses > 0 {
		maxUses = req.MaxUses
	}

	// 默认过期时间：30 天
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	} else {
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	}

	invite := &model.InviteCode{
		Code:      code,
		CreatedBy: creatorID,
		MaxUses:   maxUses,
		ExpiresAt: expiresAt,
		Note:      req.Note,
		IsActive:  true,
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, apperror.ErrInternal("creating invite code", err)
	}

	return s.toResponse(ctx, invite), nil
}

// GetByCode 查询邀请码详情（公开验证）
func (s *InviteService) GetByCode(ctx context.Context, code string) (*dto.ValidateInviteResponse, error) {
	invite, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.ValidateInviteResponse{Valid: false}, nil
		}
		return nil, apperror.ErrInternal("finding invite code", err)
	}

	valid := invite.IsValid()
	resp := &dto.ValidateInviteResponse{
		Valid:     valid,
		ExpiresAt: invite.ExpiresAt,
	}
	if valid {
		remaining := invite.RemainingUses()
		resp.RemainingUses = remaining
	}

	return resp, nil
}

// List 查询用户创建的邀请码列表
func (s *InviteService) List(ctx context.Context, userID uuid.UUID) ([]dto.InviteCodeResponse, error) {
	invites, err := s.inviteRepo.ListByCreator(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal("listing invite codes", err)
	}

	results := make([]dto.InviteCodeResponse, 0, len(invites))
	for i := range invites {
		results = append(results, *s.toResponse(ctx, &invites[i]))
	}
	return results, nil
}

// Update 更新邀请码
func (s *InviteService) Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdateInviteRequest) (*dto.InviteCodeResponse, error) {
	invite, err := s.inviteRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound("invite.not_found", "invite code not found")
		}
		return nil, apperror.ErrInternal("finding invite code", err)
	}

	// 只有创建者可以更新
	if invite.CreatedBy != userID {
		return nil, apperror.ErrForbidden("invite.forbidden", "you can only manage your own invite codes")
	}

	if req.IsActive != nil {
		invite.IsActive = *req.IsActive
	}
	if req.Note != nil {
		invite.Note = *req.Note
	}

	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		return nil, apperror.ErrInternal("updating invite code", err)
	}

	return s.toResponse(ctx, invite), nil
}

// Delete 删除邀请码
func (s *InviteService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	invite, err := s.inviteRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperror.ErrNotFound("invite.not_found", "invite code not found")
		}
		return apperror.ErrInternal("finding invite code", err)
	}

	if invite.CreatedBy != userID {
		return apperror.ErrForbidden("invite.forbidden", "you can only manage your own invite codes")
	}

	return s.inviteRepo.Delete(ctx, id)
}

// toResponse 将 model 转为 DTO
func (s *InviteService) toResponse(ctx context.Context, invite *model.InviteCode) *dto.InviteCodeResponse {
	resp := &dto.InviteCodeResponse{
		ID:            invite.ID.String(),
		Code:          invite.Code,
		CreatedBy:     invite.CreatedBy.String(),
		MaxUses:       invite.MaxUses,
		UseCount:      invite.UseCount,
		RemainingUses: invite.RemainingUses(),
		ExpiresAt:     invite.ExpiresAt,
		Note:          invite.Note,
		IsActive:      invite.IsActive,
		IsValid:       invite.IsValid(),
		CreatedAt:     invite.CreatedAt,
	}

	// 尝试获取创建者昵称
	creator, err := s.userRepo.GetByID(ctx, invite.CreatedBy)
	if err == nil {
		resp.CreatorName = creator.Nickname
	}

	return resp
}

// --- 邀请码生成 ---

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去掉 I/O/0/1 避免混淆

// generateInviteCode 生成格式为 GT-XXXXXX 的邀请码
func generateInviteCode() string {
	var sb strings.Builder
	sb.WriteString("GT-")
	for i := 0; i < 6; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		sb.WriteByte(inviteCodeChars[idx.Int64()])
	}
	return sb.String()
}

// ValidateAndConsumeInviteCode 验证并消费邀请码（注册流程使用）
func (s *InviteService) ValidateAndConsumeInviteCode(ctx context.Context, code string, userID uuid.UUID) error {
	_, err := s.inviteRepo.ConsumeByCode(ctx, code, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperror.ErrBadRequest("invite.invalid", "invite code not found")
		}
		// ConsumeByCode 内部会检查 IsValid()，失败时返回自定义错误
		if err.Error() == "invite code is not valid" {
			return apperror.ErrBadRequest("invite.invalid", "invite code is invalid or expired")
		}
		return apperror.ErrInternal(fmt.Sprintf("consuming invite code: %v", err), err)
	}
	return nil
}
