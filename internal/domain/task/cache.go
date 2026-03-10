package task

import (
	"github.com/google/uuid"
)

type Cache interface {
	GetTasks(key string) ([]Task, bool, error)
	SetTasks(key string, tasks []Task) error
	InvalidateTeamTasks(teamID uuid.UUID) error
}