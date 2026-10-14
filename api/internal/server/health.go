package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Kept short so a hanging database can't make the health check itself time
// out at the caller.
const healthPingTimeout = 2 * time.Second

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	// The build's git commit, so a deploy can confirm the new build is live.
	Version string `json:"version"`
}

// Always responds 200 while the process can serve requests, so a database
// outage doesn't make a liveness probe restart the server. Database problems
// are reported in the body instead.
func handleHealth(logger *slog.Logger, pool *pgxpool.Pool, version string) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		resp := healthResponse{Status: "ok", Database: "ok", Version: version}

		ctx, cancel := context.WithTimeout(r.Context(), healthPingTimeout)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			// The error stays in the logs: the endpoint is public.
			logger.WarnContext(r.Context(), "health check: database ping failed", "err", err)
			resp.Status = "degraded"
			resp.Database = "unavailable"
		}

		return writeJSON(w, http.StatusOK, resp)
	}
}
