package task

import (
	"context"
	"testing"
	"time"

	taskDomain "TaskControlService/internal/domain/task"
	teamDomain "TaskControlService/internal/domain/team"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockTeamRepository struct {
	createFunc        func(ctx context.Context, team *teamDomain.Team) error
	addMemberFunc     func(ctx context.Context, member *teamDomain.Member) error
	createWithCreatorFunc func(ctx context.Context, team *teamDomain.Team, member *teamDomain.Member) error
	getByIDFunc       func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error)
	getByNameFunc     func(ctx context.Context, name string) (*teamDomain.Team, error)
	getUserTeamsFunc  func(ctx context.Context, userID uuid.UUID) ([]teamDomain.Team, error)
	getTeamMembersFunc func(ctx context.Context, teamID uuid.UUID) ([]teamDomain.Member, error)
	getMemberFunc     func(ctx context.Context, teamID, userID uuid.UUID) (*teamDomain.Member, error)
}

func (m *mockTeamRepository) Create(ctx context.Context, team *teamDomain.Team) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, team)
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

type mockTaskRepository struct {
	createFunc     func(ctx context.Context, task *taskDomain.Task) error
	getByIDFunc    func(ctx context.Context, taskID uuid.UUID) (*taskDomain.Task, error)
	listFunc       func(ctx context.Context, filter taskDomain.ListFilter) ([]taskDomain.Task, error)
	updateFunc     func(ctx context.Context, task *taskDomain.Task) error
	addHistoryFunc func(ctx context.Context, entry *taskDomain.History) error
	getHistoryFunc func(ctx context.Context, taskID uuid.UUID) ([]taskDomain.History, error)
}

func (m *mockTaskRepository) Create(ctx context.Context, task *taskDomain.Task) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, task)
	}
	return nil
}

func (m *mockTeamRepository) CreateWithCreator(ctx context.Context, team *teamDomain.Team, member *teamDomain.Member) error {
	if m.createWithCreatorFunc != nil {
		return m.createWithCreatorFunc(ctx, team, member)
	}
	return nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, taskID uuid.UUID) (*taskDomain.Task, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, taskID)
	}
	return nil, nil
}

func (m *mockTaskRepository) List(ctx context.Context, filter taskDomain.ListFilter) ([]taskDomain.Task, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *taskDomain.Task) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, task)
	}
	return nil
}

func (m *mockTaskRepository) AddHistory(ctx context.Context, entry *taskDomain.History) error {
	if m.addHistoryFunc != nil {
		return m.addHistoryFunc(ctx, entry)
	}
	return nil
}

func (m *mockTaskRepository) GetHistory(ctx context.Context, taskID uuid.UUID) ([]taskDomain.History, error) {
	if m.getHistoryFunc != nil {
		return m.getHistoryFunc(ctx, taskID)
	}
	return nil, nil
}

type mockTaskCache struct {
	getTasksFunc            func(key string) ([]taskDomain.Task, bool, error)
	setTasksFunc            func(key string, tasks []taskDomain.Task) error
	invalidateTeamTasksFunc func(teamID uuid.UUID) error
}

func (m *mockTaskCache) GetTasks(key string) ([]taskDomain.Task, bool, error) {
	if m.getTasksFunc != nil {
		return m.getTasksFunc(key)
	}
	return nil, false, nil
}

func (m *mockTaskCache) SetTasks(key string, tasks []taskDomain.Task) error {
	if m.setTasksFunc != nil {
		return m.setTasksFunc(key, tasks)
	}
	return nil
}

func (m *mockTaskCache) InvalidateTeamTasks(teamID uuid.UUID) error {
	if m.invalidateTeamTasksFunc != nil {
		return m.invalidateTeamTasksFunc(teamID)
	}
	return nil
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	userID := uuid.New()

	var createdTask *taskDomain.Task
	invalidated := false

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{ID: teamID}, nil
		},
		
		getMemberFunc: func(ctx context.Context, teamIDArg, userIDArg uuid.UUID) (*teamDomain.Member, error) {
			return &teamDomain.Member{
				UserID: userIDArg,
				TeamID: teamIDArg,
				Role:   teamDomain.RoleMember,
			}, nil
		},
	}

	taskRepo := &mockTaskRepository{
		createFunc: func(ctx context.Context, task *taskDomain.Task) error {
			createdTask = task
			return nil
		},
	}

	cache := &mockTaskCache{
		invalidateTeamTasksFunc: func(teamIDArg uuid.UUID) error {
			invalidated = true
			assert.Equal(t, teamID, teamIDArg)
			return nil
		},
	}

	service := &TaskService{
		taskRepository: taskRepo,
		teamRepository: teamRepo,
		cache:          cache,
	}

	err := service.CreateTask(ctx, CreateTaskInput{
		Title:       "Test task",
		Description: "Test description",
		TeamID:      teamID,
		CreatedBy:   userID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "Test task", createdTask.Title)
	assert.Equal(t, taskDomain.StatusTodo, createdTask.Status)
	assert.True(t, invalidated)
}

func TestTaskService_CreateTask_InvalidTitle(t *testing.T) {
	ctx := context.Background()

	service := &TaskService{}

	err := service.CreateTask(ctx, CreateTaskInput{
		Title: "",
	})

	assert.ErrorIs(t, err, ErrInvalidTitle)
}

func TestTaskService_CreateTask_AssigneeNotInTeam(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	creatorID := uuid.New()
	assigneeID := uuid.New()

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{ID: teamID}, nil
		},
		getMemberFunc: func(ctx context.Context, teamIDArg, userIDArg uuid.UUID) (*teamDomain.Member, error) {
			if userIDArg == creatorID {
				return &teamDomain.Member{
					UserID: creatorID,
					TeamID: teamIDArg,
					Role:   teamDomain.RoleMember,
				}, nil
			}
			return nil, nil
		},
	}

	service := &TaskService{
		taskRepository: &mockTaskRepository{},
		teamRepository: teamRepo,
		cache:          &mockTaskCache{},
	}

	err := service.CreateTask(ctx, CreateTaskInput{
		Title:      "Test task",
		TeamID:     teamID,
		CreatedBy:  creatorID,
		AssigneeID: &assigneeID,
	})

	assert.ErrorIs(t, err, ErrAssigneeNotInTeam)
}

func TestTaskService_GetTasks_UsesCache(t *testing.T) {
	ctx := context.Background()

	teamID := uuid.New()
	userID := uuid.New()

	expectedTasks := []taskDomain.Task{
		{
			ID:        uuid.New(),
			Title:     "Cached task",
			Status:    taskDomain.StatusTodo,
			TeamID:    teamID,
			CreatedBy: userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	teamRepo := &mockTeamRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*teamDomain.Team, error) {
			return &teamDomain.Team{ID: teamID}, nil
		},
		getMemberFunc: func(ctx context.Context, teamIDArg, userIDArg uuid.UUID) (*teamDomain.Member, error) {
			return &teamDomain.Member{
				UserID: userIDArg,
				TeamID: teamIDArg,
				Role:   teamDomain.RoleMember,
			}, nil
		},
	}

	taskRepoCalled := false
	taskRepo := &mockTaskRepository{
		listFunc: func(ctx context.Context, filter taskDomain.ListFilter) ([]taskDomain.Task, error) {
			taskRepoCalled = true
			return nil, nil
		},
	}

	cache := &mockTaskCache{
		getTasksFunc: func(key string) ([]taskDomain.Task, bool, error) {
			return expectedTasks, true, nil
		},
	}

	service := &TaskService{
		taskRepository: taskRepo,
		teamRepository: teamRepo,
		cache:          cache,
	}

	result, err := service.GetTasks(ctx, userID, taskDomain.ListFilter{
		TeamID: teamID,
		Limit:  20,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedTasks, result)
	assert.False(t, taskRepoCalled)
}

func TestTaskService_UpdateTask_SuccessWritesHistory(t *testing.T) {
	ctx := context.Background()

	taskID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()

	existingTask := &taskDomain.Task{
		ID:          taskID,
		Title:       "Old title",
		Description: "Old description",
		Status:      taskDomain.StatusTodo,
		TeamID:      teamID,
		CreatedBy:   actorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newTitle := "New title"
	newStatus := taskDomain.StatusDone

	var updatedTask *taskDomain.Task
	historyCalls := 0
	invalidated := false

	taskRepo := &mockTaskRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*taskDomain.Task, error) {
			return existingTask, nil
		},
		updateFunc: func(ctx context.Context, task *taskDomain.Task) error {
			updatedTask = task
			return nil
		},
		addHistoryFunc: func(ctx context.Context, entry *taskDomain.History) error {
			historyCalls++
			return nil
		},
	}

	teamRepo := &mockTeamRepository{
		getMemberFunc: func(ctx context.Context, teamIDArg, userIDArg uuid.UUID) (*teamDomain.Member, error) {
			return &teamDomain.Member{
				UserID: userIDArg,
				TeamID: teamIDArg,
				Role:   teamDomain.RoleMember,
			}, nil
		},
	}

	cache := &mockTaskCache{
		invalidateTeamTasksFunc: func(teamIDArg uuid.UUID) error {
			invalidated = true
			return nil
		},
	}

	service := &TaskService{
		taskRepository: taskRepo,
		teamRepository: teamRepo,
		cache:          cache,
	}

	err := service.UpdateTask(ctx, UpdateTaskInput{
		TaskID:  taskID,
		ActorID: actorID,
		Title:   &newTitle,
		Status:  &newStatus,
	})

	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, "New title", updatedTask.Title)
	assert.Equal(t, taskDomain.StatusDone, updatedTask.Status)
	assert.Equal(t, 2, historyCalls)
	assert.True(t, invalidated)
}