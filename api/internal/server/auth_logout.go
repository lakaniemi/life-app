package server

import (
	"log/slog"
	"net/http"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// handleLogout deletes the caller's current session. Other sessions (e.g. on
// other devices) stay valid.
func handleLogout(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := queries.DeleteSession(r.Context(), sessionFrom(r.Context()).SessionID); err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
