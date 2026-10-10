package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lakaniemi/life-app/api/internal/auth"
)

// These cases can't be reached through the endpoints: no endpoint fails with
// an arbitrary internal error on demand, and the fake Google verifier's
// rejection reasons aren't secrets.
func TestAPIHandlerErrorsDontLeakDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   errorBody
	}{
		{
			name:       "unexpected error is a generic 500",
			err:        errors.New("connect to 10.0.0.5: password authentication failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   errorBody{Code: "internal", Message: "internal server error"},
		},
		{
			name:       "wrapped sentinel responds with only its own message",
			err:        fmt.Errorf("%w: audience %q is not ours", auth.ErrInvalidIDToken, "other-client"),
			wantStatus: http.StatusUnauthorized,
			wantBody:   errorBody{Code: "invalid_id_token", Message: auth.ErrInvalidIDToken.Error()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := apiHandler(func(http.ResponseWriter, *http.Request) error { return tt.err })
			ctx := context.WithValue(t.Context(), loggerKey{}, slog.New(slog.DiscardHandler))
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			checkStatus(t, rec, tt.wantStatus)
			checkDiff(t, "error", tt.wantBody, decode[errorResponse](t, rec).Error)
		})
	}
}
