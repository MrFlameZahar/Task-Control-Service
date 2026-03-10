package auth

import (
	user "TaskControlService/internal/domain/user"
	jwt "TaskControlService/pkg/auth"
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository user.Repository
	jwtSecret      string
}

func NewService(userRepository user.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string) error {
	existing, err := a.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	id := uuid.New()

	user := &user.User{
		ID:           id,
		Email:        email,
		PasswordHash: string(hash),
	}

	err = a.userRepository.Create(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := a.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(user.ID, a.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthService) Me(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	userModel, err := a.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userModel == nil {
		return nil, ErrUserNotFound
	}

	return userModel, nil
}
