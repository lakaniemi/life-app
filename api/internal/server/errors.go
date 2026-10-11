package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/lakaniemi/life-app/api/internal/auth"
	"github.com/lakaniemi/life-app/api/internal/family"
)

// apiHandler is a handler that returns its errors instead of writing error
// responses itself. Its ServeHTTP method writes them, so it is an
// http.Handler and routes register it directly. A handler returns an error
// only before it has written anything; writeJSON keeps to that by encoding
// before it sends.
type apiHandler func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP runs the handler and writes any error it returns with
// writeErrorResponse. The logger comes from the request context because a
// method on a function type has nowhere to hold one; server.New puts it there.
func (handler apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := handler(w, r); err != nil {
		writeErrorResponse(w, r, loggerFrom(r.Context()), err)
	}
}

// apiError is an error whose response the HTTP layer has already decided,
// such as a malformed request body.
type apiError struct {
	status  int
	code    string
	message string
	fields  []fieldError
}

func (e *apiError) Error() string {
	return e.message
}

func newAPIError(status int, code, message string) *apiError {
	return &apiError{status: status, code: code, message: message}
}

// domainErrors maps the sentinel errors of domain packages to responses.
// The message is the sentinel's own text, never the wrapping chain, which may
// hold details for the logs only (e.g. why a Google ID token was rejected).
// Add new domain sentinels here.
var domainErrors = []struct {
	err    error
	status int
	code   string
}{
	{family.ErrFamilyNotFound, http.StatusNotFound, "not_found"},
	{family.ErrFamilyMemberNotFound, http.StatusNotFound, "not_found"},
	{family.ErrInviteNotFound, http.StatusNotFound, "not_found"},
	{family.ErrForbidden, http.StatusForbidden, "forbidden"},
	// Rule violations are 409: the request is valid, but the family's current
	// state doesn't allow it.
	{family.ErrSoleAdmin, http.StatusConflict, "sole_admin"},
	{family.ErrSoleMember, http.StatusConflict, "sole_member"},
	{auth.ErrInvalidIDToken, http.StatusUnauthorized, "invalid_id_token"},
	{auth.ErrInvalidNonce, http.StatusUnauthorized, "invalid_nonce"},
}

// writeErrorResponse responds to err: an *apiError as it says, a domain
// sentinel as domainErrors maps it, and anything else as a 500 that is logged
// but doesn't leak its details.
func writeErrorResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	body, status := errorBodyFor(err)
	if status == http.StatusInternalServerError {
		logger.ErrorContext(r.Context(), "internal error", "err", err)
	}
	if err := writeJSON(w, status, errorResponse{Error: body}); err != nil {
		logger.ErrorContext(r.Context(), "encode error response", "err", err)
	}
}

func errorBodyFor(err error) (errorBody, int) {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return errorBody{Code: apiErr.code, Message: apiErr.message, Fields: apiErr.fields}, apiErr.status
	}
	for _, mapping := range domainErrors {
		if errors.Is(err, mapping.err) {
			return errorBody{Code: mapping.code, Message: mapping.err.Error()}, mapping.status
		}
	}
	return errorBody{Code: "internal", Message: "internal server error"}, http.StatusInternalServerError
}
