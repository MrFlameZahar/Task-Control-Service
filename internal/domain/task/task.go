package task

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          uuid.UUID 
	Title       string    
	Description string    
	Status      Status    
	TeamID      uuid.UUID 
	AssigneeID  *uuid.UUID
	CreatedBy   uuid.UUID 
	CreatedAt   time.Time 
	UpdatedAt   time.Time 
}

func (s Status) IsValid() bool {
	return s == StatusTodo || s == StatusInProgress || s == StatusDone
}

type History struct {
	ID        uuid.UUID
	TaskID    uuid.UUID
	Field     string 
	OldValue  *string 
	NewValue  *string  
	ChangedBy uuid.UUID 
	ChangedAt time.Time 
}

type ListFilter struct {
	TeamID     uuid.UUID
	Status     *Status
	AssigneeID *uuid.UUID
	Limit      int
	Offset     int
}
