package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRequireAppCompat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantCode   string
	}{
		{name: "no header", header: "", wantStatus: http.StatusNoContent},
		{name: "the minimum level", header: strconv.Itoa(minAppCompatLevel), wantStatus: http.StatusNoContent},
		{name: "above the minimum", header: strconv.Itoa(minAppCompatLevel + 1), wantStatus: http.StatusNoContent},
		{name: "below the minimum", header: strconv.Itoa(minAppCompatLevel - 1), wantStatus: http.StatusBadRequest, wantCode: "app_outdated"},
		{name: "negative", header: "-1", wantStatus: http.StatusBadRequest, wantCode: "app_outdated"},
		{name: "not an integer", header: "abc", wantStatus: http.StatusBadRequest, wantCode: "invalid_compat_level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := newRequireAppCompat(slog.New(slog.DiscardHandler))(next)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
			if tt.header != "" {
				req.Header.Set(appCompatLevelHeader, tt.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			checkStatus(t, rec, tt.wantStatus)
			if tt.wantCode != "" {
				checkDiff(t, "error code", tt.wantCode, decode[errorResponse](t, rec).Error.Code)
			}
		})
	}
}
