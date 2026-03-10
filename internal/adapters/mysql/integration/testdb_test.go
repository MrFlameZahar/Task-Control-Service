package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

type testDB struct {
	Container testcontainers.Container
	DB        *sqlx.DB
	DSN       string
}

func setupMySQL(t *testing.T) *testDB {
	t.Helper()

	ctx := context.Background()

	container, err := tcmysql.Run(
		ctx,
		"mysql:8.4",
		tcmysql.WithDatabase("task_control_test"),
		tcmysql.WithUsername("test"),
		tcmysql.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get mysql host: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("failed to get mysql port: %v", err)
	}

	dsn := fmt.Sprintf(
		"test:test@tcp(%s:%s)/task_control_test?parseTime=true",
		host,
		port.Port(),
	)

	var db *sqlx.DB
	for i := 0; i < 20; i++ {
		db, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("failed to connect mysql: %v", err)
	}

	return &testDB{
		Container: container,
		DB:        db,
		DSN:       dsn,
	}
}

func (tdb *testDB) teardown(t *testing.T) {
	t.Helper()

	if tdb.DB != nil {
		_ = tdb.DB.Close()
	}

	if tdb.Container != nil {
		_ = tdb.Container.Terminate(context.Background())
	}
}

func createUsersTable(t *testing.T, db *sqlx.DB) {
	t.Helper()

	schema := `
	CREATE TABLE users (
		id CHAR(36) PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
}

func createTeamsTables(t *testing.T, db *sqlx.DB) {
	t.Helper()

	teamsSchema := `
	CREATE TABLE teams (
		id CHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		created_by CHAR(36) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_teams_created_by (created_by),
		CONSTRAINT fk_teamscreated_by FOREIGN KEY (created_by) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(teamsSchema); err != nil {
		t.Fatalf("failed to create teams table: %v", err)
	}

	teamMembersSchema := `
	CREATE TABLE team_members (
		user_id CHAR(36) NOT NULL,
		team_id CHAR(36) NOT NULL,
		role VARCHAR(20) NOT NULL,
		PRIMARY KEY (user_id, team_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (team_id) REFERENCES teams(id)
	);
	`

	if _, err := db.Exec(teamMembersSchema); err != nil {
		t.Fatalf("failed to create team_members table: %v", err)
	}
}

func createTasksTables(t *testing.T, db *sqlx.DB) {
	t.Helper()

	tasksSchema := `
	CREATE TABLE tasks (
		id CHAR(36) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(32) NOT NULL,
		team_id CHAR(36) NOT NULL,
		assignee_id CHAR(36) NULL,
		created_by CHAR(36) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_tasks_team FOREIGN KEY (team_id) REFERENCES teams(id),
		CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES users(id),
		CONSTRAINT fk_tasks_created_by FOREIGN KEY (created_by) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(tasksSchema); err != nil {
		t.Fatalf("failed to create tasks table: %v", err)
	}

	taskHistorySchema := `
	CREATE TABLE task_history (
		id CHAR(36) PRIMARY KEY,
		task_id CHAR(36) NOT NULL,
		field VARCHAR(64) NOT NULL,
		old_value TEXT NULL,
		new_value TEXT NULL,
		changed_by CHAR(36) NOT NULL,
		changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_task_history_task FOREIGN KEY (task_id) REFERENCES tasks(id),
		CONSTRAINT fk_task_history_changed_by FOREIGN KEY (changed_by) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(taskHistorySchema); err != nil {
		t.Fatalf("failed to create task_history table: %v", err)
	}
}
