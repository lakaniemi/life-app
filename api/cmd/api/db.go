package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/lakaniemi/life-app/api/internal/migrations"
)

const pingTimeout = 5 * time.Second

// newPool connects eagerly so that a bad DATABASE_URL or unreachable
// database fails at startup instead of on the first request.
func newPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// migrate applies pending migrations before the server starts, so a new
// image's code never runs against an older schema.
func migrate(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) (err error) {
	// goose needs a database/sql handle. This one borrows connections from the
	// pool, and closing it leaves the pool open.
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	provider, err := migrations.NewProvider(db)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}
	results, err := provider.Up(ctx)
	// On failure, results still lists the migrations applied before it.
	for _, result := range results {
		logger.Info("migration applied", "migration", result.Source.Path, "duration", result.Duration.String())
	}
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
