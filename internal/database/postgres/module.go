package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/internal/config"
)

func New(config *config.ApplicationConfig) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), config.DatabaseURL)
}
