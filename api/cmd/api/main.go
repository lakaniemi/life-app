// Command api is the Life App HTTP API server.
package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// distroless has no /usr/share/zoneinfo; without this, time.LoadLocation
	// fails in the container.
	_ "time/tzdata"

	"github.com/lakaniemi/life-app/api/internal/server"
)

// Cloud Run allows 10s between SIGTERM and SIGKILL.
const shutdownTimeout = 8 * time.Second

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	env := cmp.Or(getenv("ENVIRONMENT"), "prod")
	logger, err := newLogger(env, stdout)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort("", cmp.Or(getenv("PORT"), "8080"))

	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}
	pool, err := newPool(ctx, databaseURL)
	if err != nil {
		return err
	}
	// Deferred, so it runs after srv.Shutdown below has drained in-flight
	// requests that may still be using connections.
	defer pool.Close()
	logger.Info("database connected")

	srv := &http.Server{
		Handler:           server.New(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	logger.Info("server listening", "addr", ln.Addr().String())

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	logger.Info("server stopped")
	return nil
}
