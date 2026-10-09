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
}

// Always responds 200 while the process can serve requests, so a database
// outage doesn't make a liveness probe restart the server. Database problems
// are reported in the body instead.
func handleHealth(logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{Status: "ok", Database: "ok"}

		ctx, cancel := context.WithTimeout(r.Context(), healthPingTimeout)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			// The error stays in the logs: the endpoint is public.
			logger.WarnContext(r.Context(), "health check: database ping failed", "err", err)
			resp = healthResponse{Status: "degraded", Database: "unavailable"}
		}

		writeJSON(w, r, logger, http.StatusOK, resp)
	})
}
