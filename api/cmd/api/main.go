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

	// ENVIRONMENT is only read here, to pick concrete settings. Code deeper
	// in the app receives those settings rather than checking the
	// environment name itself.
	env := cmp.Or(getenv("ENVIRONMENT"), "prod")
	logger, err := newLogger(env, stdout)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort("", cmp.Or(getenv("PORT"), "8080"))

	srv := &http.Server{
		Handler:           server.New(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Bind the port before logging that we're listening, so a port conflict
	// fails here with a clear error instead of after a misleading log line.
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
		// Serve only returns before shutdown if the server fails.
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
