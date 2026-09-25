package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

// handleHealth reports that the process is up and serving requests. It is a
// liveness check only: it doesn't check downstream dependencies.
func handleHealth(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(healthResponse{Status: "ok"}); err != nil {
			logger.ErrorContext(r.Context(), "encode health response", "err", err)
		}
	})
}
