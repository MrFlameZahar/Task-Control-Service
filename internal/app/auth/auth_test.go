package auth

import (
	"context"
	"testing"

	userDomain "TaskControlService/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

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

func TestService_Register_Success(t *testing.T) {
	ctx := context.Background()

	var createdUser *userDomain.User

	repo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*userDomain.User, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, user *userDomain.User) error {
			createdUser = user
			return nil
		},
	}

	service := NewService(repo, "test-secret")

	err := service.Register(ctx, "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, "test@example.com", createdUser.Email)
	assert.NotEqual(t, "", createdUser.PasswordHash)
	assert.NotEqual(t, "password123", createdUser.PasswordHash)
	assert.NotEqual(t, uuid.Nil, createdUser.ID)

	hashErr := bcrypt.CompareHashAndPassword([]byte(createdUser.PasswordHash), []byte("password123"))
	assert.NoError(t, hashErr)
}

func TestService_Register_UserAlreadyExists(t *testing.T) {
	ctx := context.Background()

	repo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*userDomain.User, error) {
			return &userDomain.User{
				ID:    uuid.New(),
				Email: email,
			}, nil
		},
	}

	service := NewService(repo, "test-secret")

	err := service.Register(ctx, "test@example.com", "password123")

	assert.ErrorIs(t, err, ErrUserExists)
}

func TestService_Login_Success(t *testing.T) {
	ctx := context.Background()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()

	repo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*userDomain.User, error) {
			return &userDomain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: string(passwordHash),
			}, nil
		},
	}

	service := NewService(repo, "test-secret")

	token, err := service.Login(ctx, "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestService_Login_InvalidCredentials_WhenUserNotFound(t *testing.T) {
	ctx := context.Background()

	repo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*userDomain.User, error) {
			return nil, nil
		},
	}

	service := NewService(repo, "test-secret")

	token, err := service.Login(ctx, "test@example.com", "password123")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, token)
}

func TestService_Login_InvalidCredentials_WhenPasswordMismatch(t *testing.T) {
	ctx := context.Background()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	repo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*userDomain.User, error) {
			return &userDomain.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: string(passwordHash),
			}, nil
		},
	}

	service := NewService(repo, "test-secret")

	token, err := service.Login(ctx, "test@example.com", "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, token)
}