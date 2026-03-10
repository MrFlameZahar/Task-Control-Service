package http

import (
	"TaskControlService/internal/metrics"
	"TaskControlService/internal/ports/http/handlers"
	"TaskControlService/internal/ports/http/middleware"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger,
	handler *handlers.Handler,
	jwtSecret string,
	httpMetrics *metrics.HTTPMetrics,
	rateLimiter *middleware.RateLimiter,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Metrics(httpMetrics))
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recoverer)

	router.Handle("/metrics", promhttp.Handler())

	router.Route("/api/v1", func(r chi.Router) {

		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))

			if rateLimiter != nil {
				r.Use(rateLimiter.Middleware)
			}

			r.Get("/analytics/team-stats", handler.GetTeamStats)
			r.Get("/analytics/top-creators", handler.GetTopCreators)
			r.Get("/analytics/integrity", handler.GetIntegrityIssues)

			r.Get("/me", handler.Me)

			r.Post("/teams", handler.CreateTeam)
			r.Get("/teams", handler.GetTeams)
			r.Post("/teams/{id}/invite", handler.InviteMember)
			r.Get("/teams/{id}/members", handler.GetTeamMembers)

			r.Post("/tasks", handler.CreateTask)
			r.Get("/tasks", handler.GetTasks)
			r.Get("/tasks/{id}/history", handler.GetTaskHistory)
			r.Put("/tasks/{id}", handler.UpdateTask)

		})
	})

	return router
}
