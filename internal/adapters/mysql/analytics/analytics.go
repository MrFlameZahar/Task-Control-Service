package analytics

import (
	"context"

	analyticsDomain "TaskControlService/internal/domain/analytics"

	"github.com/jmoiron/sqlx"
)

type teamStatsRow struct {
	TeamID             string `db:"team_id"`
	TeamName           string `db:"team_name"`
	MembersCount       int    `db:"members_count"`
	DoneTasksLast7Days int    `db:"done_tasks_last_7_days"`
}

type topCreatorRow struct {
	TeamID     string `db:"team_id"`
	UserID     string `db:"user_id"`
	TasksCount int    `db:"tasks_count"`
	Rank       int    `db:"rank_num"`
}

type integrityIssueRow struct {
	TaskID     string  `db:"task_id"`
	Title      string  `db:"title"`
	TeamID     string  `db:"team_id"`
	AssigneeID *string `db:"assignee_id"`
}

type AnalyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (a *AnalyticsRepository) GetTeamStats(ctx context.Context) ([]analyticsDomain.TeamStats, error) {
	query := `
	SELECT
		t.id AS team_id,
		t.name AS team_name,
		COUNT(DISTINCT tm.user_id) AS members_count,
		COUNT(DISTINCT CASE
			WHEN ta.status = 'done' AND ta.updated_at >= NOW() - INTERVAL 7 DAY
			THEN ta.id
		END) AS done_tasks_last_7_days
	FROM teams t
	LEFT JOIN team_members tm ON tm.team_id = t.id
	LEFT JOIN tasks ta ON ta.team_id = t.id
	GROUP BY t.id, t.name
	ORDER BY t.name
	`

	var row []teamStatsRow
	if err := a.db.SelectContext(ctx, &row, query); err != nil {
		return nil, err
	}

	var result []analyticsDomain.TeamStats
	for _, r := range row {
		result = append(result, analyticsDomain.TeamStats{
			TeamID:             r.TeamID,
			TeamName:           r.TeamName,
			MembersCount:       r.MembersCount,
			DoneTasksLast7Days: r.DoneTasksLast7Days,
		})
	}

	return result, nil
}

func (a *AnalyticsRepository) GetTopCreators(ctx context.Context) ([]analyticsDomain.TopCreator, error) {
	query := `
	SELECT team_id, user_id, tasks_count, rank_num
	FROM (
		SELECT
			t.team_id,
			t.created_by AS user_id,
			COUNT(*) AS tasks_count,
			ROW_NUMBER() OVER (
				PARTITION BY t.team_id
				ORDER BY COUNT(*) DESC, t.created_by
			) AS rank_num
		FROM tasks t
		WHERE t.created_at >= NOW() - INTERVAL 1 MONTH
		GROUP BY t.team_id, t.created_by
	) ranked
	WHERE rank_num <= 3
	ORDER BY team_id, rank_num
	`

	var row []topCreatorRow

	if err := a.db.SelectContext(ctx, &row, query); err != nil {
		return nil, err
	}

	var result []analyticsDomain.TopCreator

	for _, r := range row {
		result = append(result, analyticsDomain.TopCreator{
			TeamID:     r.TeamID,
			UserID:     r.UserID,
			TasksCount: r.TasksCount,
			Rank:       r.Rank,
		})
	}

	return result, nil
}

func (a *AnalyticsRepository) GetIntegrityIssues(ctx context.Context) ([]analyticsDomain.IntegrityIssue, error) {
	query := `
	SELECT
		t.id AS task_id,
		t.title,
		t.team_id,
		t.assignee_id
	FROM tasks t
	LEFT JOIN team_members tm
		ON tm.team_id = t.team_id
		AND tm.user_id = t.assignee_id
	WHERE t.assignee_id IS NOT NULL
	  AND tm.user_id IS NULL
	ORDER BY t.created_at DESC
	`

	var row []integrityIssueRow
	if err := a.db.SelectContext(ctx, &row, query); err != nil {
		return nil, err
	}

	var result []analyticsDomain.IntegrityIssue
	for _, r := range row {
		result = append(result, analyticsDomain.IntegrityIssue{
			TaskID:     r.TaskID,
			Title:      r.Title,
			TeamID:     r.TeamID,
			AssigneeID: r.AssigneeID,
		})
	}

	return result, nil
}
