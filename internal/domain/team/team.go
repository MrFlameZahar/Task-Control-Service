package team

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Team struct {
	ID        uuid.UUID
	Name      string
	CreatedBy   uuid.UUID
	CreatedAt time.Time
}

type Member struct {
	UserID uuid.UUID
	TeamID uuid.UUID
	Role   Role
}

func IsValidRole(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

func (r Role) CanInvite() bool {
	return r == RoleOwner || r == RoleAdmin
}