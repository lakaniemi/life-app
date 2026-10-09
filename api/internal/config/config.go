// Package config turns environment variables into the settings the
// entrypoints in cmd/ pass down to the rest of the code.
package config

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
)

// LogFormat selects how log lines are written.
type LogFormat string

// Log formats. prod uses JSON so a log platform can parse the fields.
const (
	LogFormatPretty LogFormat = "pretty"
	LogFormatJSON   LogFormat = "json"
)

// localDatabaseURL points at the Postgres in compose.yaml.
const localDatabaseURL = "postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable" //nolint:gosec // local-only dev credentials, same as compose.yaml

// defaultEnvironment is used when ENVIRONMENT is unset.
const defaultEnvironment = "prod"

// defaults holds each environment's settings; environment variables override
// them. ENVIRONMENT picks the entry. An empty field has no default, and
// fromEnv says which of those are required.
var defaults = map[string]Config{
	"prod": {
		LogFormat: LogFormatJSON,
		Port:      "8080",
	},
	"dev": {
		LogFormat:   LogFormatPretty,
		Port:        "8080",
		DatabaseURL: localDatabaseURL,
	},
}

// Config holds concrete settings. It deliberately has no environment name, so
// code can't branch on "are we in dev?"; ENVIRONMENT only picks defaults.
type Config struct {
	LogFormat   LogFormat
	Port        string
	DatabaseURL string
	// GoogleClientIDs are the OAuth client IDs whose Google ID tokens are
	// accepted (the token's aud). See docs/AUTH.md.
	GoogleClientIDs []string
}

// Load reads configuration from the environment. Variables in a .env file in
// the working directory are added first, without overriding ones that are
// already set; the file is optional. Keep .env out of container images
// (.dockerignore), so deployments only use their real environment.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	return fromEnv(os.Getenv)
}

// fromEnv builds the Config from getenv, so tests can pass a map instead of
// the real environment. prod has no default DATABASE_URL, so it's required
// there. GOOGLE_CLIENT_IDS has no default anywhere, but only the API server
// requires it, so it's validated where it's used (google.New), not here.
func fromEnv(getenv func(string) string) (Config, error) {
	env := cmp.Or(getenv("ENVIRONMENT"), defaultEnvironment)
	d, ok := defaults[env]
	if !ok {
		return Config{}, fmt.Errorf("unknown ENVIRONMENT %q (want one of %q)", env, slices.Sorted(maps.Keys(defaults)))
	}

	cfg := Config{
		LogFormat:       d.LogFormat,
		Port:            cmp.Or(getenv("PORT"), d.Port),
		DatabaseURL:     cmp.Or(getenv("DATABASE_URL"), d.DatabaseURL),
		GoogleClientIDs: splitList(getenv("GOOGLE_CLIENT_IDS")),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	return cfg, nil
}

// splitList parses a comma-separated list. Whitespace also separates items,
// and empty items are dropped.
func splitList(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
}
