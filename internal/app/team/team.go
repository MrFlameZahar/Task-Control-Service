package team

import (
	"TaskControlService/internal/domain/notification"
	teamDomain "TaskControlService/internal/domain/team"
	userDomain "TaskControlService/internal/domain/user"
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TeamService struct {
	teamRepository teamDomain.Repository
	userRepository userDomain.Repository
	emailSender    notification.EmailSender
	logger         *zap.Logger
}

func NewService(
	teamRepository teamDomain.Repository,
	userRepository userDomain.Repository,
	emailSender notification.EmailSender,
	logger *zap.Logger,
) *TeamService {
	return &TeamService{
		teamRepository: teamRepository,
		userRepository: userRepository,
		emailSender:    emailSender,
		logger:         logger,
	}
}

func (t *TeamService) CreateTeam(ctx context.Context, name string, createdBy uuid.UUID) error {
	existing, err := t.teamRepository.GetByName(ctx, name)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrTeamExists
	}

	team := &teamDomain.Team{
		ID:        uuid.New(),
		Name:      name,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	member := &teamDomain.Member{
		UserID: createdBy,
		TeamID: team.ID,
		Role:   teamDomain.RoleOwner,
	}

	return t.teamRepository.CreateWithCreator(ctx, team, member)
}

func (t *TeamService) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]teamDomain.Team, error) {
	teams, err := t.teamRepository.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (t *TeamService) GetTeam(ctx context.Context, teamID uuid.UUID) (*teamDomain.Team, error) {
	teamEntity, err := t.teamRepository.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if teamEntity == nil {
		return nil, ErrTeamNotFound
	}
	return teamEntity, nil
}

func (t *TeamService) InviteMember(
	ctx context.Context,
	actorID uuid.UUID,
	teamID uuid.UUID,
	targetUserID uuid.UUID,
	role teamDomain.Role,
) error {
	if role != teamDomain.RoleAdmin && role != teamDomain.RoleMember {
		return ErrInvalidInviteRole
	}

	teamEntity, err := t.teamRepository.GetByID(ctx, teamID)
	if err != nil {
		return err
	}
	if teamEntity == nil {
		return ErrTeamNotFound
	}

	actorMember, err := t.teamRepository.GetMember(ctx, teamID, actorID)
	if err != nil {
		return err
	}
	if actorMember == nil {
		return ErrForbidden
	}
	if !actorMember.Role.CanInvite() {
		return ErrForbidden
	}

	targetUser, err := t.userRepository.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return ErrUserNotFound
	}

	existingMember, err := t.teamRepository.GetMember(ctx, teamID, targetUserID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return ErrAlreadyTeamMember
	}

	member := &teamDomain.Member{
		UserID: targetUserID,
		TeamID: teamID,
		Role:   role,
	}

	if err := t.teamRepository.AddMember(ctx, member); err != nil {
		return err
	}

	if t.emailSender != nil {
		if err := t.emailSender.SendTeamInvitation(ctx, targetUser.Email, teamEntity.Name); err != nil {
			t.logger.Warn(
				"failed to send invitation email",
				zap.String("user_id", targetUserID.String()),
				zap.String("email", targetUser.Email),
				zap.String("team_id", teamID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (t *TeamService) GetTeamMembers(
	ctx context.Context,
	actorID uuid.UUID,
	teamID uuid.UUID,
) ([]teamDomain.Member, error) {
	teamEntity, err := t.teamRepository.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if teamEntity == nil {
		return nil, ErrTeamNotFound
	}

	actorMember, err := t.teamRepository.GetMember(ctx, teamID, actorID)
	if err != nil {
		return nil, err
	}
	if actorMember == nil {
		return nil, ErrForbidden
	}

	members, err := t.teamRepository.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return members, nil
}
