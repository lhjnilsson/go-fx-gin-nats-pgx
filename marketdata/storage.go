package marketdata

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type OHLC struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    uint64
	Timestamp time.Time
}

type OHLCStorage interface {
	Get(ctx context.Context, symbol string) ([]*OHLC, error)
	Add(ctx context.Context, symbol string, ohlc *OHLC) error
	Remove(ctx context.Context, symbol string) error
}

type PostgresStorage struct {
	conn *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, conn *pgxpool.Pool) (OHLCStorage, error) {
	_, err := conn.Exec(ctx, table)
	return &PostgresStorage{conn: conn}, err
}

func (s *PostgresStorage) Get(ctx context.Context, symbol string) ([]*OHLC, error) {
	rows, err := s.conn.Query(ctx, "SELECT open, high, low, close, volume, timestamp FROM ohlc WHERE symbol = $1", symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ohlc := []*OHLC{}
	for rows.Next() {
		var o OHLC
		err := rows.Scan(&o.Open, &o.High, &o.Low, &o.Close, &o.Volume, &o.Timestamp)
		if err != nil {
			return nil, err
		}
		ohlc = append(ohlc, &o)
	}
	return ohlc, nil
}

func (s *PostgresStorage) Add(ctx context.Context, symbol string, ohlc *OHLC) error {
	_, err := s.conn.Exec(ctx,
		"INSERT INTO ohlc (symbol, open, high, low, close, volume, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		symbol, ohlc.Open, ohlc.High, ohlc.Low, ohlc.Close, ohlc.Volume, ohlc.Timestamp)
	return err
}

func (s *PostgresStorage) Remove(ctx context.Context, symbol string) error {
	_, err := s.conn.Exec(ctx, "DELETE FROM ohlc WHERE symbol = $1", symbol)
	return err
}

const table = `
CREATE TABLE IF NOT EXISTS ohlc (
	id SERIAL PRIMARY KEY,
	symbol TEXT NOT NULL,
	open FLOAT NOT NULL,
	high FLOAT NOT NULL,
	low FLOAT NOT NULL,
	close FLOAT NOT NULL,
	volume INT NOT NULL,
	timestamp TIMESTAMP NOT NULL);`
