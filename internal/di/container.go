package di

import (
	"TaskControlService/internal/adapters/email"
	"TaskControlService/internal/adapters/mysql"
	analyticsRepo "TaskControlService/internal/adapters/mysql/analytics"
	taskRepo "TaskControlService/internal/adapters/mysql/task"
	teamRepo "TaskControlService/internal/adapters/mysql/team"
	"TaskControlService/internal/adapters/mysql/user"
	redisAdapter "TaskControlService/internal/adapters/redis"
	"TaskControlService/internal/app/analytics"
	"TaskControlService/internal/app/auth"
	"TaskControlService/internal/app/task"
	"TaskControlService/internal/app/team"
	"TaskControlService/internal/config"
	"TaskControlService/internal/logger"
	"TaskControlService/internal/metrics"
	"TaskControlService/internal/ports/http"
	"TaskControlService/internal/ports/http/handlers"
	"TaskControlService/internal/ports/http/middleware"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Container struct {
	Config *config.Config
	Logger *zap.Logger
	Router *chi.Mux

	DB    *sqlx.DB
	Redis *redis.Client
}

func New() (*Container, error) {

	cfg, err := config.MustLoad()
	if err != nil {
		return nil, err
	}

	log, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	db, err := mysql.NewDB(*cfg)
	if err != nil {
		return nil, err
	}
	
	httpMetrics := metrics.NewHTTPMetrics()

	redisClient := redisAdapter.NewClient(cfg)
	taskCache := redisAdapter.NewTaskCache(redisClient)

	rateLimiter := middleware.NewRateLimiter(100, 100, 10*time.Minute)
	go rateLimiter.CleanupLoop()

	mockSender := email.NewMockSender(0.3)
	emailSender := email.NewCircuitBreakerSender(mockSender)
	
	userRepository := user.NewUserRepository(db)
	teamRepository := teamRepo.NewTeamRepository(db)
	taskRepository := taskRepo.NewTaskRepository(db)
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db)

	authService := auth.NewService(userRepository, cfg.JWTSecret)
	teamService := team.NewService(teamRepository, userRepository, emailSender, log)
	taskService := task.NewService(taskRepository, teamRepository, taskCache)
	analyticsService := analytics.NewService(analyticsRepository)

	handlers := handlers.NewHandler(authService, teamService, taskService, analyticsService, log)

	router := http.NewRouter(
		log,
		handlers,
		cfg.JWTSecret,
		httpMetrics,
		rateLimiter,
	)

	return &Container{
		Config: cfg,
		Logger: log,
		Router: router,
		DB:     db,
		Redis:  redisClient,
	}, nil
}
