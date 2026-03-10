package task

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	List(ctx context.Context, filter ListFilter) ([]Task, error)
	Update(ctx context.Context, task *Task) error

	AddHistory(ctx context.Context, entry *History) error
	GetHistory(ctx context.Context, taskID uuid.UUID) ([]History, error)
}