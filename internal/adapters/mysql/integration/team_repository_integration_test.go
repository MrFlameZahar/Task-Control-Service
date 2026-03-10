package integration

import (
	"context"
	"testing"
	"time"

	teammysql "TaskControlService/internal/adapters/mysql/team"
	teamDomain "TaskControlService/internal/domain/team"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTeamRepository_CreateAndGetByName(t *testing.T) {
	tdb := setupMySQL(t)
	defer tdb.teardown(t)

	createUsersTable(t, tdb.DB)
	createTeamsTables(t, tdb.DB)

	createdBy := uuid.New()
	_, err := tdb.DB.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		createdBy.String(),
		"createdby@example.com",
		"hashed",
	)
	require.NoError(t, err)

	repo := teammysql.NewTeamRepository(tdb.DB)

	team := &teamDomain.Team{
		ID:        uuid.New(),
		Name:      "Backend",
		CreatedBy:   createdBy,
		CreatedAt: time.Now(),
	}

	err = repo.Create(context.Background(), team)
	require.NoError(t, err)

	got, err := repo.GetByName(context.Background(), "Backend")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, team.Name, got.Name)
	require.Equal(t, team.ID, got.ID)
}