
package integration

import (
	"context"
	"testing"

	usermysql "TaskControlService/internal/adapters/mysql/user"
	userDomain "TaskControlService/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGetByEmail(t *testing.T) {
	tdb := setupMySQL(t)
	defer tdb.teardown(t)

	createUsersTable(t, tdb.DB)

	repo := usermysql.NewUserRepository(tdb.DB)

	user := &userDomain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	got, err := repo.GetByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.ID, got.ID)
}