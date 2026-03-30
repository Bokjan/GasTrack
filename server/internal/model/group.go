package model

import (
	"time"

	"github.com/google/uuid"
)

// GroupRole 群组成员角色
type GroupRole string

const (
	GroupRoleOwner  GroupRole = "owner"
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleMember GroupRole = "member"
)

// Group 家庭/群组模型
type Group struct {
	BaseModel

	Name        string    `gorm:"size:100;not null" json:"name"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	InviteCode  string    `gorm:"size:20;uniqueIndex" json:"invite_code"`
	MaxMembers  int       `gorm:"default:10;not null" json:"max_members"`
	Description string    `gorm:"size:500" json:"description,omitempty"`

	// 关联
	Owner   User          `gorm:"foreignKey:OwnerID" json:"-"`
	Members []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "groups"
}

// GroupMember 群组成员模型
type GroupMember struct {
	GroupID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role     GroupRole `gorm:"size:20;default:member;not null" json:"role"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`

	// 关联
	Group Group `gorm:"foreignKey:GroupID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (GroupMember) TableName() string {
	return "group_members"
}
