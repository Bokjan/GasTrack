package service

import (
	"context"

	"github.com/google/uuid"

	"gastrack/internal/dto"
	"gastrack/internal/model"
)

// InviteServicer 邀请码业务接口（供 AuthService 依赖注入）
type InviteServicer interface {
	Create(ctx context.Context, creatorID uuid.UUID, req *dto.CreateInviteRequest) (*dto.InviteCodeResponse, error)
	GetByCode(ctx context.Context, code string) (*dto.ValidateInviteResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.InviteCodeResponse, error)
	Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdateInviteRequest) (*dto.InviteCodeResponse, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ValidateAndConsumeInviteCode(ctx context.Context, code string, userID uuid.UUID) (*model.InviteCode, error)
}

// NotificationServicer 通知业务接口（供 AuthService、FuelRecordService 依赖注入）
type NotificationServicer interface {
	List(ctx context.Context, userID uuid.UUID) ([]dto.NotificationResponse, error)
	UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	CheckFuelAnomaly(ctx context.Context, userID, vehicleID uuid.UUID, recordID uuid.UUID, currentEfficiency float64)
	CheckMaintenanceReminders(ctx context.Context, userID, vehicleID uuid.UUID, currentOdometer float64)
	NotifyInviteUsed(ctx context.Context, inviteCreatorID uuid.UUID, inviteCode string, newUserNickname string)
}
