package integration

import (
	"context"
	"testing"
	"time"

	taskmysql "TaskControlService/internal/adapters/mysql/task"
	taskDomain "TaskControlService/internal/domain/task"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTaskRepository_CreateAndGetByID(t *testing.T) {
	tdb := setupMySQL(t)
	defer tdb.teardown(t)

	createUsersTable(t, tdb.DB)
	createTeamsTables(t, tdb.DB)
	createTasksTables(t, tdb.DB)

	userID := uuid.New()
	teamID := uuid.New()
	taskID := uuid.New()

	_, err := tdb.DB.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		userID.String(),
		"user@example.com",
		"hashed",
	)
	require.NoError(t, err)

	_, err = tdb.DB.Exec(
		`INSERT INTO teams (id, name, created_by) VALUES (?, ?, ?)`,
		teamID.String(),
		"Backend",
		userID.String(),
	)
	require.NoError(t, err)

	_, err = tdb.DB.Exec(
		`INSERT INTO team_members (user_id, team_id, role) VALUES (?, ?, ?)`,
		userID.String(),
		teamID.String(),
		"owner",
	)
	require.NoError(t, err)

	repo := taskmysql.NewTaskRepository(tdb.DB)

	task := &taskDomain.Task{
		ID:          taskID,
		Title:       "Test task",
		Description: "Description",
		Status:      taskDomain.StatusTodo,
		TeamID:      teamID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(context.Background(), task)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), taskID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, task.ID, got.ID)
	require.Equal(t, task.Title, got.Title)
	require.Equal(t, task.TeamID, got.TeamID)
}