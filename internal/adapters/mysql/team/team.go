package team

import (
	"TaskControlService/internal/domain/team"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type teamRow struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedBy uuid.UUID `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

type memberRow struct {
	UserID uuid.UUID `db:"user_id"`
	TeamID uuid.UUID `db:"team_id"`
	Role   team.Role `db:"role"`
}

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *team.Team) error {

	query := `
	INSERT INTO teams (id, name, created_by, created_at)
	VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		team.ID.String(),
		team.Name,
		team.CreatedBy.String(),
		team.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *TeamRepository) AddMember(ctx context.Context, member *team.Member) error {

	query := `
	INSERT INTO team_members (user_id, team_id, role)
	VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		member.UserID.String(),
		member.TeamID.String(),
		member.Role,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *TeamRepository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]team.Team, error) {
	query := `
	SELECT t.id, t.name, t.created_by, t.created_at
	FROM teams t
	JOIN team_members tm ON tm.team_id = t.id
	WHERE tm.user_id = ?
	`

	var rows []teamRow

	err := r.db.SelectContext(ctx, &rows, query, userID.String())
	if err != nil {
		return nil, err
	}

	var teams []team.Team
	for _, t := range rows {
		teams = append(teams, team.Team{
			ID:        t.ID,
			Name:      t.Name,
			CreatedBy: t.CreatedBy,
			CreatedAt: t.CreatedAt,
		})
	}

	return teams, nil
}

func (r *TeamRepository) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]team.Member, error) {
	query := `
	SELECT user_id, team_id, role
	FROM team_members
	WHERE team_id = ?
	`

	var rows []memberRow

	err := r.db.SelectContext(ctx, &rows, query, teamID.String())
	if err != nil {
		return nil, err
	}

	var members []team.Member
	for _, m := range rows {
		members = append(members, team.Member{
			UserID: m.UserID,
			TeamID: m.TeamID,
			Role:   m.Role,
		})
	}

	return members, nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name string) (*team.Team, error) {
	query := `
	SELECT id, name, created_by, created_at
	FROM teams
	WHERE name = ?
	LIMIT 1
	`

	var row teamRow
	err := r.db.GetContext(ctx, &row, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var result team.Team
	result.ID = row.ID
	result.Name = row.Name
	result.CreatedBy = row.CreatedBy
	result.CreatedAt = row.CreatedAt

	return &result, nil
}

func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	query := `
	SELECT id, name, created_by, created_at
	FROM teams
	WHERE id = ?
	LIMIT 1
	`

	var row teamRow
	err := r.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var result team.Team
	result.ID = row.ID
	result.Name = row.Name
	result.CreatedBy = row.CreatedBy
	result.CreatedAt = row.CreatedAt

	return &result, nil
}

func (r *TeamRepository) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*team.Member, error) {
	query := `
	SELECT user_id, team_id, role
	FROM team_members
	WHERE team_id = ? AND user_id = ?
	LIMIT 1
	`

	var row memberRow
	err := r.db.GetContext(ctx, &row, query, teamID.String(), userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var result team.Member
	result.UserID = row.UserID
	result.TeamID = row.TeamID
	result.Role = row.Role

	return &result, nil
}

func (r *TeamRepository) CreateWithCreator(ctx context.Context, team *team.Team, member *team.Member) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	createTeamQuery := `
	INSERT INTO teams (id, name, created_by, created_at)
	VALUES (?, ?, ?, ?)
	`

	_, err = tx.ExecContext(
		ctx,
		createTeamQuery,
		team.ID.String(),
		team.Name,
		team.CreatedBy.String(),
		team.CreatedAt,
	)
	if err != nil {
		return err
	}

	addMemberQuery := `
	INSERT INTO team_members (user_id, team_id, role)
	VALUES (?, ?, ?)
	`

	_, err = tx.ExecContext(
		ctx,
		addMemberQuery,
		member.UserID.String(),
		member.TeamID.String(),
		member.Role,
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}