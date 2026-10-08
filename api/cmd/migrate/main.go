// Command migrate applies or inspects the database schema migrations, and
// creates new migration files.
//
// Usage:
//
//	migrate up              apply all pending migrations
//	migrate down            roll back the most recent migration
//	migrate status          list migrations and whether each is applied
//	migrate create <name>   write a new empty SQL migration file
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Registers the "pgx" driver for database/sql, which goose requires.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/lakaniemi/life-app/api/internal/migrations"
)

const usage = "usage: migrate up | down | status | create <name>"

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args[1:], os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// err is a named result so the deferred db.Close can add its error to it.
func run(ctx context.Context, args []string, getenv func(string) string) (err error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(args) == 0 {
		return errors.New(usage)
	}
	command := args[0]

	// create only writes a file, so it runs without a database.
	if command == "create" {
		if len(args) != 2 {
			return errors.New("usage: migrate create <name>")
		}
		goose.SetSequential(true)
		return goose.Create(nil, migrations.Dir, args[1], "sql")
	}

	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	provider, err := migrations.NewProvider(db)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	switch command {
	case "up":
		return up(ctx, provider)
	case "down":
		return down(ctx, provider)
	case "status":
		return status(ctx, provider)
	default:
		return fmt.Errorf("unknown command %q\n%s", command, usage)
	}
}

func up(ctx context.Context, provider *goose.Provider) error {
	results, err := provider.Up(ctx)
	// On failure, results still lists the migrations applied before it.
	for _, r := range results {
		fmt.Println(r)
	}
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	if len(results) == 0 {
		fmt.Println("no pending migrations")
	}
	return nil
}

func down(ctx context.Context, provider *goose.Provider) error {
	result, err := provider.Down(ctx)
	if err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	fmt.Println(result)
	return nil
}

func status(ctx context.Context, provider *goose.Provider) error {
	statuses, err := provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}
	for _, s := range statuses {
		appliedAt := "-"
		if !s.AppliedAt.IsZero() {
			appliedAt = s.AppliedAt.Format(time.DateTime)
		}
		fmt.Printf("%-8s %-20s %s\n", s.State, appliedAt, s.Source.Path)
	}
	return nil
}
