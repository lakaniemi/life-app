package server

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to remember the status code.
// Embedding the interface means every other method (Header, Write) is
// forwarded to the wrapped writer automatically; we only override
// WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the underlying writer, so
// features like flushing still work through this wrapper.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// logRequests logs one line per request with method, path, status, and
// duration.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Handlers that never call WriteHeader implicitly send 200.
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}
