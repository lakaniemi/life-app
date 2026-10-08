package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// Long enough for the user to finish Google sign-in.
const nonceTTL = 10 * time.Minute

type nonceResponse struct {
	Nonce string `json:"nonce"`
}

// handleCreateNonce issues a single-use nonce for the app to pass to Google
// sign-in, which embeds it in the ID token. See docs/AUTH.md.
func handleCreateNonce(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cleaning up here keeps the table bounded without a background job.
		if err := queries.DeleteExpiredNonces(r.Context()); err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		nonce := newRandomToken()
		err := queries.CreateNonce(r.Context(), db.CreateNonceParams{
			Nonce:     nonce,
			ExpiresAt: time.Now().Add(nonceTTL),
		})
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, nonceResponse{Nonce: nonce})
	})
}
