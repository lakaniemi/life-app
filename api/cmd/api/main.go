// Command api is the Life App HTTP API server.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Embeds the IANA timezone database into the binary. The distroless
	// runtime image has no /usr/share/zoneinfo, so without this
	// time.LoadLocation would fail in the container.
	_ "time/tzdata"

	"github.com/lakaniemi/life-app/api/internal/server"
)

// shutdownTimeout is how long in-flight requests get to finish after a
// shutdown signal. Cloud Run allows 10s between SIGTERM and SIGKILL.
const shutdownTimeout = 8 * time.Second

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// run holds the real program. Taking the environment and output as
// arguments, rather than reading globals, keeps it testable.
func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              net.JoinHostPort("", port),
		Handler:           server.New(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		// ListenAndServe only returns early on failure, e.g. port in use.
		return fmt.Errorf("listen: %w", err)
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
