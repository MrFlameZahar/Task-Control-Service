package handlers

import (
	"TaskControlService/internal/app/auth"
	"TaskControlService/internal/app/team"
	"TaskControlService/internal/app/task"
	"TaskControlService/internal/app/analytics"

	"go.uber.org/zap"
)

type Handler struct {
	authService *auth.AuthService
	teamService *team.TeamService
	taskService *task.TaskService
	analyticsService *analytics.AnalyticsService
	logger      *zap.Logger
}

func NewHandler(
	authService *auth.AuthService,
	teamService *team.TeamService,
	taskService *task.TaskService,
	analyticsService *analytics.AnalyticsService,
	logger *zap.Logger,

) *Handler {
	return &Handler{
		authService: authService,
		teamService: teamService,
		taskService: taskService,
		analyticsService: analyticsService,
		logger:      logger,
	}
}
