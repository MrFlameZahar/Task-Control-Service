package task

import (
	taskDomain "TaskControlService/internal/domain/task"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var _ taskDomain.Repository = (*TaskRepository)(nil)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

type taskRow struct {
	ID          string         `db:"id"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	Status      string         `db:"status"`
	TeamID      string         `db:"team_id"`
	AssigneeID  sql.NullString `db:"assignee_id"`
	CreatedBy   string         `db:"created_by"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type historyRow struct {
	ID        string         `db:"id"`
	TaskID    string         `db:"task_id"`
	Field     string         `db:"field"`
	OldValue  sql.NullString `db:"old_value"`
	NewValue  sql.NullString `db:"new_value"`
	ChangedBy string         `db:"changed_by"`
	ChangedAt time.Time      `db:"changed_at"`
}

func (r *TaskRepository) Create(ctx context.Context, task *taskDomain.Task) error {
	query := `
	INSERT INTO tasks (
		id,
		title,
		description,
		status,
		team_id,
		assignee_id,
		created_by,
		created_at,
		updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var assigneeID any
	if task.AssigneeID != nil {
		assigneeID = task.AssigneeID.String()
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		task.ID.String(),
		task.Title,
		task.Description,
		string(task.Status),
		task.TeamID.String(),
		assigneeID,
		task.CreatedBy.String(),
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID uuid.UUID) (*taskDomain.Task, error) {
	query := `
	SELECT
		id,
		title,
		description,
		status,
		team_id,
		assignee_id,
		created_by,
		created_at,
		updated_at
	FROM tasks
	WHERE id = ?
	LIMIT 1
	`

	var row taskRow
	err := r.db.GetContext(ctx, &row, query, taskID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	task, err := mapTaskRow(row)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) List(ctx context.Context, filter taskDomain.ListFilter) ([]taskDomain.Task, error) {
	var (
		sb   strings.Builder
		args []any
	)

	sb.WriteString(`
	SELECT
		id,
		title,
		description,
		status,
		team_id,
		assignee_id,
		created_by,
		created_at,
		updated_at
	FROM tasks
	WHERE team_id = ?
	`)
	args = append(args, filter.TeamID.String())

	if filter.Status != nil {
		sb.WriteString(` AND status = ?`)
		args = append(args, string(*filter.Status))
	}

	if filter.AssigneeID != nil {
		sb.WriteString(` AND assignee_id = ?`)
		args = append(args, filter.AssigneeID.String())
	}

	sb.WriteString(` ORDER BY created_at DESC`)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	sb.WriteString(` LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)

	var rows []taskRow
	if err := r.db.SelectContext(ctx, &rows, sb.String(), args...); err != nil {
		return nil, err
	}

	result := make([]taskDomain.Task, 0, len(rows))
	for _, row := range rows {
		taskEntity, err := mapTaskRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *taskEntity)
	}

	return result, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *taskDomain.Task) error {
	query := `
	UPDATE tasks
	SET
		title = ?,
		description = ?,
		status = ?,
		team_id = ?,
		assignee_id = ?,
		updated_at = ?
	WHERE id = ?
	`

	var assigneeID any
	if task.AssigneeID != nil {
		assigneeID = task.AssigneeID.String()
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		string(task.Status),
		task.TeamID.String(),
		assigneeID,
		task.UpdatedAt,
		task.ID.String(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *TaskRepository) AddHistory(ctx context.Context, entry *taskDomain.History) error {
	query := `
	INSERT INTO task_history (
		id,
		task_id,
		field,
		old_value,
		new_value,
		changed_by,
		changed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var oldValue any
	if entry.OldValue != nil {
		oldValue = *entry.OldValue
	}

	var newValue any
	if entry.NewValue != nil {
		newValue = *entry.NewValue
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		entry.ID.String(),
		entry.TaskID.String(),
		entry.Field,
		oldValue,
		newValue,
		entry.ChangedBy.String(),
		entry.ChangedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) GetHistory(ctx context.Context, taskID uuid.UUID) ([]taskDomain.History, error) {
	query := `
	SELECT
		id,
		task_id,
		field,
		old_value,
		new_value,
		changed_by,
		changed_at
	FROM task_history
	WHERE task_id = ?
	ORDER BY changed_at DESC
	`

	var rows []historyRow
	if err := r.db.SelectContext(ctx, &rows, query, taskID.String()); err != nil {
		return nil, err
	}

	result := make([]taskDomain.History, 0, len(rows))
	for _, row := range rows {
		historyEntry, err := mapHistoryRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *historyEntry)
	}

	return result, nil
}

func mapTaskRow(row taskRow) (*taskDomain.Task, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, err
	}

	teamID, err := uuid.Parse(row.TeamID)
	if err != nil {
		return nil, err
	}

	createdBy, err := uuid.Parse(row.CreatedBy)
	if err != nil {
		return nil, err
	}

	var assigneeID *uuid.UUID
	if row.AssigneeID.Valid {
		parsedAssigneeID, err := uuid.Parse(row.AssigneeID.String)
		if err != nil {
			return nil, err
		}
		assigneeID = &parsedAssigneeID
	}

	return &taskDomain.Task{
		ID:          id,
		Title:       row.Title,
		Description: row.Description,
		Status:      taskDomain.Status(row.Status),
		TeamID:      teamID,
		AssigneeID:  assigneeID,
		CreatedBy:   createdBy,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func mapHistoryRow(row historyRow) (*taskDomain.History, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, err
	}

	taskID, err := uuid.Parse(row.TaskID)
	if err != nil {
		return nil, err
	}

	changedBy, err := uuid.Parse(row.ChangedBy)
	if err != nil {
		return nil, err
	}

	var oldValue *string
	if row.OldValue.Valid {
		value := row.OldValue.String
		oldValue = &value
	}

	var newValue *string
	if row.NewValue.Valid {
		value := row.NewValue.String
		newValue = &value
	}

	return &taskDomain.History{
		ID:        id,
		TaskID:    taskID,
		Field:     row.Field,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangedBy: changedBy,
		ChangedAt: row.ChangedAt,
	}, nil
}