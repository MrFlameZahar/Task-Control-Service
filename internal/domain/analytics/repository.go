package analytics

import "context"

type Repository interface {
	GetTeamStats(ctx context.Context) ([]TeamStats, error)
	GetTopCreators(ctx context.Context) ([]TopCreator, error)
	GetIntegrityIssues(ctx context.Context) ([]IntegrityIssue, error)
}