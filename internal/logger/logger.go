package logger

import (
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return logger, nil
}