package task

import (
	taskDomain "TaskControlService/internal/domain/task"
	teamDomain "TaskControlService/internal/domain/team"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CreateTaskInput struct {
	Title       string
	Description string
	Status      *taskDomain.Status
	TeamID      uuid.UUID
	AssigneeID  *uuid.UUID
	CreatedBy   uuid.UUID
}

type UpdateTaskInput struct {
	TaskID      uuid.UUID
	ActorID     uuid.UUID
	Title       *string
	Description *string
	Status      *taskDomain.Status
	AssigneeID  *uuid.UUID
	ClearAssign bool
}

type TaskService struct {
	taskRepository taskDomain.Repository
	teamRepository teamDomain.Repository
	cache          taskDomain.Cache
}

func NewService(
	taskRepository taskDomain.Repository,
	teamRepository teamDomain.Repository,
	cache taskDomain.Cache,
) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		teamRepository: teamRepository,
		cache:          cache,
	}
}

func (t *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) error {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return ErrInvalidTitle
	}

	teamEntity, err := t.teamRepository.GetByID(ctx, input.TeamID)
	if err != nil {
		return err
	}
	if teamEntity == nil {
		return ErrTeamNotFound
	}

	creatorMember, err := t.teamRepository.GetMember(ctx, input.TeamID, input.CreatedBy)
	if err != nil {
		return err
	}
	if creatorMember == nil {
		return ErrCreatorNotInTeam
	}

	status := taskDomain.StatusTodo
	if input.Status != nil {
		if !input.Status.IsValid() {
			return ErrInvalidStatus
		}
		status = *input.Status
	}

	if input.AssigneeID != nil {
		assigneeMember, err := t.teamRepository.GetMember(ctx, input.TeamID, *input.AssigneeID)
		if err != nil {
			return err
		}
		if assigneeMember == nil {
			return ErrAssigneeNotInTeam
		}
	}

	now := time.Now()

	taskEntity := &taskDomain.Task{
		ID:          uuid.New(),
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
		TeamID:      input.TeamID,
		AssigneeID:  input.AssigneeID,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := t.taskRepository.Create(ctx, taskEntity); err != nil {
		return err
	}

	if t.cache != nil {
		_ = t.cache.InvalidateTeamTasks(input.TeamID)
	}

	return nil
}

func (t *TaskService) GetTasks(
	ctx context.Context,
	actorID uuid.UUID,
	filter taskDomain.ListFilter,
) ([]taskDomain.Task, error) {
	teamEntity, err := t.teamRepository.GetByID(ctx, filter.TeamID)
	if err != nil {
		return nil, err
	}
	if teamEntity == nil {
		return nil, ErrTeamNotFound
	}

	actorMember, err := t.teamRepository.GetMember(ctx, filter.TeamID, actorID)
	if err != nil {
		return nil, err
	}
	if actorMember == nil {
		return nil, ErrForbidden
	}

	if filter.Status != nil && !filter.Status.IsValid() {
		return nil, ErrInvalidStatus
	}

	cacheKey := buildTasksCacheKey(filter)

	if t.cache != nil {
		cachedTasks, found, err := t.cache.GetTasks(cacheKey)
		if err == nil && found {
			return cachedTasks, nil
		}
	}

	tasks, err := t.taskRepository.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if t.cache != nil {
		_ = t.cache.SetTasks(cacheKey, tasks)
	}

	return tasks, nil
}

func (t *TaskService) GetTaskHistory(
	ctx context.Context,
	actorID uuid.UUID,
	taskID uuid.UUID,
) ([]taskDomain.History, error) {
	taskEntity, err := t.taskRepository.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if taskEntity == nil {
		return nil, ErrTaskNotFound
	}

	actorMember, err := t.teamRepository.GetMember(ctx, taskEntity.TeamID, actorID)
	if err != nil {
		return nil, err
	}
	if actorMember == nil {
		return nil, ErrForbidden
	}

	history, err := t.taskRepository.GetHistory(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (t *TaskService) UpdateTask(ctx context.Context, input UpdateTaskInput) error {
	taskEntity, err := t.taskRepository.GetByID(ctx, input.TaskID)
	if err != nil {
		return err
	}
	if taskEntity == nil {
		return ErrTaskNotFound
	}

	actorMember, err := t.teamRepository.GetMember(ctx, taskEntity.TeamID, input.ActorID)
	if err != nil {
		return err
	}
	if actorMember == nil {
		return ErrForbidden
	}

	updatedTask := *taskEntity

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return ErrInvalidTitle
		}
		updatedTask.Title = title
	}

	if input.Description != nil {
		updatedTask.Description = strings.TrimSpace(*input.Description)
	}

	if input.Status != nil {
		if !input.Status.IsValid() {
			return ErrInvalidStatus
		}
		updatedTask.Status = *input.Status
	}

	if input.ClearAssign {
		updatedTask.AssigneeID = nil
	} else if input.AssigneeID != nil {
		assigneeMember, err := t.teamRepository.GetMember(ctx, taskEntity.TeamID, *input.AssigneeID)
		if err != nil {
			return err
		}
		if assigneeMember == nil {
			return ErrAssigneeNotInTeam
		}
		updatedTask.AssigneeID = input.AssigneeID
	}

	updatedTask.UpdatedAt = time.Now()

	if err := t.taskRepository.Update(ctx, &updatedTask); err != nil {
		return err
	}

	historyEntries := buildTaskHistory(taskEntity, &updatedTask, input.ActorID)
	for _, entry := range historyEntries {
		if err := t.taskRepository.AddHistory(ctx, &entry); err != nil {
			return err
		}
	}

	if t.cache != nil {
		_ = t.cache.InvalidateTeamTasks(taskEntity.TeamID)
	}

	return nil
}

func buildTaskHistory(oldTask, newTask *taskDomain.Task, actorID uuid.UUID) []taskDomain.History {
	result := make([]taskDomain.History, 0, 4)

	if oldTask.Title != newTask.Title {
		oldValue := oldTask.Title
		newValue := newTask.Title

		result = append(result, taskDomain.History{
			ID:        uuid.New(),
			TaskID:    oldTask.ID,
			Field:     "title",
			OldValue:  &oldValue,
			NewValue:  &newValue,
			ChangedBy: actorID,
			ChangedAt: newTask.UpdatedAt,
		})
	}

	if oldTask.Description != newTask.Description {
		oldValue := oldTask.Description
		newValue := newTask.Description

		result = append(result, taskDomain.History{
			ID:        uuid.New(),
			TaskID:    oldTask.ID,
			Field:     "description",
			OldValue:  &oldValue,
			NewValue:  &newValue,
			ChangedBy: actorID,
			ChangedAt: newTask.UpdatedAt,
		})
	}

	if oldTask.Status != newTask.Status {
		oldValue := string(oldTask.Status)
		newValue := string(newTask.Status)

		result = append(result, taskDomain.History{
			ID:        uuid.New(),
			TaskID:    oldTask.ID,
			Field:     "status",
			OldValue:  &oldValue,
			NewValue:  &newValue,
			ChangedBy: actorID,
			ChangedAt: newTask.UpdatedAt,
		})
	}

	if !equalUUIDPointers(oldTask.AssigneeID, newTask.AssigneeID) {
		var oldValue *string
		if oldTask.AssigneeID != nil {
			value := oldTask.AssigneeID.String()
			oldValue = &value
		}

		var newValue *string
		if newTask.AssigneeID != nil {
			value := newTask.AssigneeID.String()
			newValue = &value
		}

		result = append(result, taskDomain.History{
			ID:        uuid.New(),
			TaskID:    oldTask.ID,
			Field:     "assignee_id",
			OldValue:  oldValue,
			NewValue:  newValue,
			ChangedBy: actorID,
			ChangedAt: newTask.UpdatedAt,
		})
	}

	return result
}

func equalUUIDPointers(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
