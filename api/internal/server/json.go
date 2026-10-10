package server

import (
	"encoding/json"
	"errors"
	"fmt"
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
	// Fields lists the failing fields of a request that failed validation.
	Fields []fieldError `json:"fields,omitempty"`
}

// writeJSON responds with status and v as the body. It encodes v before
// sending anything, so an error means nothing was written and the caller can
// still respond with an error instead.
func writeJSON(w http.ResponseWriter, status int, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Writing fails only if the client has gone away, and then there's no one
	// left to tell.
	_, _ = w.Write(append(body, '\n'))
	return nil
}

// decodeJSON decodes the request body into v, returning an *apiError if it
// can't. The decoder's own message isn't sent, because it names Go types.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	err := json.NewDecoder(r.Body).Decode(v)
	var tooLarge *http.MaxBytesError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &tooLarge):
		return newAPIError(http.StatusRequestEntityTooLarge, "body_too_large", "request body is too large")
	default:
		return newAPIError(http.StatusBadRequest, "invalid_body", "request body is not a valid JSON object")
	}
}
