package instrument

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Instrument struct {
	ID       string
	Symbol   string
	Name     string
	Class    string
	Exchange string
}

type InstrumentStorage interface {
	Get(ctx context.Context, symbol string) (*Instrument, error)
	Add(ctx context.Context, instrument *Instrument) error
	Remove(ctx context.Context, symbol string) error
}

type PostgresStorage struct {
	conn *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, conn *pgxpool.Pool) (InstrumentStorage, error) {
	_, err := conn.Exec(ctx, table)
	return &PostgresStorage{conn: conn}, err
}

func (s *PostgresStorage) Get(ctx context.Context, symbol string) (*Instrument, error) {
	instrument := Instrument{}
	err := s.conn.QueryRow(ctx, "SELECT id, name, class, exchange FROM instruments WHERE symbol = $1", symbol).Scan(
		&instrument.ID, &instrument.Name, &instrument.Class, &instrument.Exchange)
	return &instrument, err
}

func (s *PostgresStorage) Add(ctx context.Context, instrument *Instrument) error {
	return s.conn.QueryRow(ctx,
		"INSERT INTO instruments (id, symbol, name, class, exchange) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		instrument.ID, instrument.Symbol, instrument.Name, instrument.Class, instrument.Exchange).Scan(&instrument.ID)
}

func (s *PostgresStorage) Remove(ctx context.Context, symbol string) error {
	_, err := s.conn.Exec(ctx, "DELETE FROM instruments WHERE symbol = $1", symbol)
	return err
}

const table = `
CREATE TABLE IF NOT EXISTS instruments (
	id TEXT PRIMARY KEY,
	symbol TEXT NOT NULL,
	name TEXT NOT NULL,
	class TEXT NOT NULL,
	exchange TEXT NOT NULL
);
`
