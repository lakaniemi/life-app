package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const maxRequestBodyBytes = 64 << 10

type errorResponse struct {
	Error errorBody `json:"error"`
}

// errorBody is the error shape for every endpoint. Code is machine-readable
// and stable; Message is for humans and may change.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.ErrorContext(r.Context(), "encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, code, message string) {
	writeJSON(w, r, logger, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

// writeInternalError logs err and responds 500 without leaking its details.
func writeInternalError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	logger.ErrorContext(r.Context(), "internal error", "err", err)
	writeError(w, r, logger, http.StatusInternalServerError, "internal", "internal server error")
}

// decodeJSON decodes the request body into v, capping its size.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	return nil
}
