package team

import (
	"context"
	"testing"

	notificationDomain "TaskControlService/internal/domain/notification"
	teamDomain "TaskControlService/internal/domain/team"
	userDomain "TaskControlService/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockTeamRepository struct {
	createFunc            func(ctx context.Context, team *teamDomain.Team) error
	addMemberFunc         func(ctx context.Context, member *teamDomain.Member) error
	createWithCreatorFunc func(ctx context.Context, team *teamDomain.Team, member *teamDomain.Member) error
	getByIDFunc           func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error)
	getByNameFunc         func(ctx context.Context, name string) (*teamDomain.Team, error)
	getUserTeamsFunc      func(ctx context.Context, userID uuid.UUID) ([]teamDomain.Team, error)
	getTeamMembersFunc    func(ctx context.Context, teamID uuid.UUID) ([]teamDomain.Member, error)
	getMemberFunc         func(ctx context.Context, teamID, userID uuid.UUID) (*teamDomain.Member, error)
}

func (m *mockTeamRepository) Create(ctx context.Context, team *teamDomain.Team) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, team)
	}
	return nil
}

func (m *mockTeamRepository) CreateWithCreator(ctx context.Context, team *teamDomain.Team, member *teamDomain.Member) error {
	if m.createWithCreatorFunc != nil {
		return m.createWithCreatorFunc(ctx, team, member)
	}
	return nil
}

func (m *mockTeamRepository) AddMember(ctx context.Context, member *teamDomain.Member) error {
	if m.addMemberFunc != nil {
		return m.addMemberFunc(ctx, member)
	}
	return nil
}

func (m *mockTeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockTeamRepository) GetByName(ctx context.Context, name string) (*teamDomain.Team, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockTeamRepository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]teamDomain.Team, error) {
	if m.getUserTeamsFunc != nil {
		return m.getUserTeamsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockTeamRepository) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]teamDomain.Member, error) {
	if m.getTeamMembersFunc != nil {
		return m.getTeamMembersFunc(ctx, teamID)
	}
	return nil, nil
}

func (m *mockTeamRepository) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*teamDomain.Member, error) {
	if m.getMemberFunc != nil {
		return m.getMemberFunc(ctx, teamID, userID)
	}
	return nil, nil
}

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *userDomain.User) error
	getByEmailFunc func(ctx context.Context, email string) (*userDomain.User, error)
	getByIDFunc    func(ctx context.Context, id uuid.UUID) (*userDomain.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *userDomain.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

type mockEmailSender struct {
	sendTeamInvitationFunc func(ctx context.Context, email string, teamName string) error
}

func (m *mockEmailSender) SendTeamInvitation(ctx context.Context, email string, teamName string) error {
	if m.sendTeamInvitationFunc != nil {
		return m.sendTeamInvitationFunc(ctx, email, teamName)
	}
	return nil
}

var _ notificationDomain.EmailSender = (*mockEmailSender)(nil)

func TestTeamService_CreateTeam_Success(t *testing.T) {
	ctx := context.Background()
	createdBy := uuid.New()

	var createdTeam *teamDomain.Team
	var addedMember *teamDomain.Member

	teamRepo := &mockTeamRepository{
		getByNameFunc: func(ctx context.Context, name string) (*teamDomain.Team, error) {
			return nil, nil
		},
		createWithCreatorFunc: func(ctx context.Context, team *teamDomain.Team, member *teamDomain.Member) error {
			createdTeam = team
			addedMember = member
			return nil
		},
		createFunc: func(ctx context.Context, team *teamDomain.Team) error {
			createdTeam = team
			return nil
		},
		addMemberFunc: func(ctx context.Context, member *teamDomain.Member) error {
			addedMember = member
			return nil
		},
	}

	service := &TeamService{
		teamRepository: teamRepo,
		userRepository: &mockUserRepository{},
		emailSender:    &mockEmailSender{},
		logger:         zap.NewNop(),
	}

	err := service.CreateTeam(ctx, "Backend Team", createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, createdTeam)
	assert.Equal(t, "Backend Team", createdTeam.Name)
	assert.Equal(t, createdBy, createdTeam.CreatedBy)

	assert.NotNil(t, addedMember)
	assert.Equal(t, createdBy, addedMember.UserID)
	assert.Equal(t, createdTeam.ID, addedMember.TeamID)
	assert.Equal(t, teamDomain.RoleOwner, addedMember.Role)
}

func TestTeamService_CreateTeam_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	createdBy := uuid.New()

	teamRepo := &mockTeamRepository{
		getByNameFunc: func(ctx context.Context, name string) (*teamDomain.Team, error) {
			return &teamDomain.Team{
				ID:   uuid.New(),
				Name: name,
			}, nil
		},
	}

	service := &TeamService{
		teamRepository: teamRepo,
		userRepository: &mockUserRepository{},
		emailSender:    &mockEmailSender{},
		logger:         zap.NewNop(),
	}

	err := service.CreateTeam(ctx, "Backend Team", createdBy)

	assert.ErrorIs(t, err, ErrTeamExists)
}

func TestTeamService_InviteMember_Success(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	actorID := uuid.New()
	targetID := uuid.New()

	var addedMember *teamDomain.Member
	emailCalled := false

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{
				ID:   teamID,
				Name: "Backend Team",
			}, nil
		},
		getMemberFunc: func(ctx context.Context, teamIDArg, userID uuid.UUID) (*teamDomain.Member, error) {
			if userID == actorID {
				return &teamDomain.Member{
					UserID: actorID,
					TeamID: teamIDArg,
					Role:   teamDomain.RoleOwner,
				}, nil
			}

			if userID == targetID {
				return nil, nil
			}

			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, member *teamDomain.Member) error {
			addedMember = member
			return nil
		},
	}

	userRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
			return &userDomain.User{
				ID:    targetID,
				Email: "target@example.com",
			}, nil
		},
	}

	emailSender := &mockEmailSender{
		sendTeamInvitationFunc: func(ctx context.Context, email string, teamName string) error {
			emailCalled = true
			assert.Equal(t, "target@example.com", email)
			assert.Equal(t, "Backend Team", teamName)
			return nil
		},
	}

	service := &TeamService{
		teamRepository: teamRepo,
		userRepository: userRepo,
		emailSender:    emailSender,
		logger:         zap.NewNop(),
	}

	err := service.InviteMember(ctx, actorID, teamID, targetID, teamDomain.RoleMember)

	assert.NoError(t, err)
	assert.NotNil(t, addedMember)
	assert.Equal(t, targetID, addedMember.UserID)
	assert.Equal(t, teamID, addedMember.TeamID)
	assert.Equal(t, teamDomain.RoleMember, addedMember.Role)
	assert.True(t, emailCalled)
}

func TestTeamService_InviteMember_ForbiddenForMemberRole(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	actorID := uuid.New()
	targetID := uuid.New()

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{
				ID:   teamID,
				Name: "Backend Team",
			}, nil
		},
		getMemberFunc: func(ctx context.Context, teamIDArg, userID uuid.UUID) (*teamDomain.Member, error) {
			if userID == actorID {
				return &teamDomain.Member{
					UserID: actorID,
					TeamID: teamIDArg,
					Role:   teamDomain.RoleMember,
				}, nil
			}
			return nil, nil
		},
	}

	userRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
			return &userDomain.User{
				ID:    targetID,
				Email: "target@example.com",
			}, nil
		},
	}

	service := &TeamService{
		teamRepository: teamRepo,
		userRepository: userRepo,
		emailSender:    &mockEmailSender{},
		logger:         zap.NewNop(),
	}

	err := service.InviteMember(ctx, actorID, teamID, targetID, teamDomain.RoleMember)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestTeamService_InviteMember_TargetAlreadyMember(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	actorID := uuid.New()
	targetID := uuid.New()

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{
				ID:   teamID,
				Name: "Backend Team",
			}, nil
		},
		getMemberFunc: func(ctx context.Context, teamIDArg, userID uuid.UUID) (*teamDomain.Member, error) {
			switch userID {
			case actorID:
				return &teamDomain.Member{
					UserID: actorID,
					TeamID: teamIDArg,
					Role:   teamDomain.RoleAdmin,
				}, nil
			case targetID:
				return &teamDomain.Member{
					UserID: targetID,
					TeamID: teamIDArg,
					Role:   teamDomain.RoleMember,
				}, nil
			default:
				return nil, nil
			}
		},
	}

	userRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
			return &userDomain.User{
				ID:    targetID,
				Email: "target@example.com",
			}, nil
		},
	}

	service := &TeamService{
		teamRepository: teamRepo,
		userRepository: userRepo,
		emailSender:    &mockEmailSender{},
		logger:         zap.NewNop(),
	}

	err := service.InviteMember(ctx, actorID, teamID, targetID, teamDomain.RoleMember)

	assert.ErrorIs(t, err, ErrAlreadyTeamMember)
}
