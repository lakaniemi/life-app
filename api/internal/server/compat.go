package server

import (
	"log/slog"
	"net/http"
	"strconv"
)

// minAppCompatLevel is the lowest API compatibility level an app may declare.
// It is a constant so that raising it ships in the same deploy as the breaking
// change it protects: apps below it get app_outdated and the "Update required"
// screen. See app/docs/ARCHITECTURE.md, "Version compatibility".
const minAppCompatLevel = 1

const appCompatLevelHeader = "App-Compat-Level"

// newRequireAppCompat returns requireAppCompat, which turns away apps that
// declare a compatibility level below minAppCompatLevel with 400 app_outdated.
// Requests without the header (curl, tests, health checks) pass, because only
// the app sends it and every build does.
func newRequireAppCompat(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get(appCompatLevelHeader)
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			level, err := strconv.Atoi(header)
			if err != nil {
				writeErrorResponse(w, r, logger, newAPIError(http.StatusBadRequest, "invalid_compat_level", appCompatLevelHeader+" is not an integer"))
				return
			}
			if level < minAppCompatLevel {
				writeErrorResponse(w, r, logger, newAPIError(http.StatusBadRequest, "app_outdated", "this app version is too old for the API"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
