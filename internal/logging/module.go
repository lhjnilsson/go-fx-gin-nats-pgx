package logging

import (
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/internal/config"
	"go.uber.org/zap"
)

func New(config *config.ApplicationConfig) *zap.Logger {
	log, _ := zap.NewProduction()
	log.Info("Started logging")
	return log
}
