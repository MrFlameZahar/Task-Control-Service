package main

import (
	"TaskControlService/internal/di"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func main() {
	container, err := di.New()
	if err != nil {
		panic(err)
	}

	server := http.Server{
		Addr:    ":" + container.Config.AppPort,
		Handler: container.Router,
	}
	go server.ListenAndServe()
	container.Logger.Info("Server started", zap.String("port", container.Config.AppPort))

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	container.Logger.Info("shutting down")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
    container.Logger.Error("graceful shutdown failed", zap.Error(err))

    if err := server.Close(); err != nil {
        container.Logger.Error("server close failed", zap.Error(err))
    	}
	}

	container.Logger.Info("server stopped")
}
