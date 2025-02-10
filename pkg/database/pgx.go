package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPgx(conn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
