// Command devsession creates a session for a local development user and
// prints its token, for calling authenticated endpoints without going through
// Google sign-in:
//
//	TOKEN=$(make -s dev-session name=Alice)
//	curl localhost:8080/me -H "Authorization: Bearer $TOKEN"
//
// The user is found or created with the Google sub "dev:<name>". The command
// talks to the database directly and isn't part of the API or its container
// image, so it adds no way to sign in over HTTP.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/auth"
	"github.com/lakaniemi/life-app/api/internal/auth/google"
	"github.com/lakaniemi/life-app/api/internal/config"
	"github.com/lakaniemi/life-app/api/internal/db"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// err is a named result so the deferred conn.Close can add its error to it.
func run(ctx context.Context, args []string, stdout io.Writer) (err error) {
	flags := flag.NewFlagSet("devsession", flag.ContinueOnError)
	name := flags.String("name", "", "name of the dev user to sign in as (required)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return errors.New("usage: devsession -name <name>")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() {
		err = errors.Join(err, conn.Close(context.Background()))
	}()

	// stdout carries only the token, so it can be captured with $(...).
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := auth.NewService(logger, db.New(conn), nil)
	token, _, err := service.StartSession(ctx, google.Identity{Subject: "dev:" + *name, Name: *name})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, token)
	return err
}
