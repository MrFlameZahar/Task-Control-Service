package analytics

import (
	"context"

	analyticsDomain "TaskControlService/internal/domain/analytics"
)

type AnalyticsService struct {
	repo analyticsDomain.Repository
}

func NewService(repo analyticsDomain.Repository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (a *AnalyticsService) GetTeamStats(ctx context.Context) ([]analyticsDomain.TeamStats, error) {
	teamStats, err := a.repo.GetTeamStats(ctx)
	if err != nil {
		return nil, err
	}
	
	return teamStats, nil
}

func (a *AnalyticsService) GetTopCreators(ctx context.Context) ([]analyticsDomain.TopCreator, error) {
	topCreators, err := a.repo.GetTopCreators(ctx)
	if err != nil {
		return nil, err
	}

	return topCreators, nil
}

func (a *AnalyticsService) GetIntegrityIssues(ctx context.Context) ([]analyticsDomain.IntegrityIssue, error) {
	integrityIssue, err := a.repo.GetIntegrityIssues(ctx)
	if err != nil {
		return nil, err
	}

	return integrityIssue, nil
}