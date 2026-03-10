package user

import (
	domain "TaskControlService/internal/domain/user"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type userDB struct {
	ID           string `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, email, password_hash)
        VALUES (?, ?, ?)
    `
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE email = ?
	`

	var row userDB

	err := r.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           id,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	query := `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE id = ?
	`

	var row userDB

	err := r.db.GetContext(ctx, &row, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           id,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}, nil
}
