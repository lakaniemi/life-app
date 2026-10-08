// Package dbtest provides fresh, migrated Postgres databases for integration
// tests.
package dbtest

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/lakaniemi/life-app/api/internal/migrations"
)

// New creates an empty database on the server at TEST_DATABASE_URL, applies
// all migrations to it and returns a pool connected to it. The database is
// dropped when the test ends. If TEST_DATABASE_URL is unset, the test is
// skipped, so that plain `go test` works without a database.
//
// Reading the env var here is a deliberate exception to "env vars only in
// cmd/": tests have no main to pass settings down from.
func New(t *testing.T) *pgxpool.Pool {
	t.Helper()

	serverURL := os.Getenv("TEST_DATABASE_URL")
	if serverURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := t.Context()

	admin, err := pgx.Connect(ctx, serverURL)
	if err != nil {
		t.Fatalf("connect to test server: %v", err)
	}
	name := "test_" + strings.ToLower(rand.Text())
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	// t.Context() is already canceled when cleanups run.
	t.Cleanup(func() {
		ctx := context.Background()
		if _, err := admin.Exec(ctx, "DROP DATABASE "+ident+" WITH (FORCE)"); err != nil {
			t.Errorf("drop test database: %v", err)
		}
		_ = admin.Close(ctx)
	})

	cfg, err := pgxpool.ParseConfig(serverURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	// Registered after the drop, so it runs before it (cleanups are LIFO).
	t.Cleanup(pool.Close)

	migrate(t, pool)
	return pool
}

func migrate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// goose works on database/sql, so wrap the pool rather than opening a
	// second set of connections.
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	provider, err := migrations.NewProvider(db)
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}
