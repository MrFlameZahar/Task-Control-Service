package team

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, team *Team) error
	AddMember(ctx context.Context, member *Member) error
	CreateWithCreator(ctx context.Context, team *Team, member *Member) error

	GetByID(ctx context.Context, teamID uuid.UUID) (*Team, error)
	GetByName(ctx context.Context, name string) (*Team, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]Member, error)
	GetMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (*Member, error)
}