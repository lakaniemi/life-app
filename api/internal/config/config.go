// Package config turns environment variables into the settings the
// entrypoints in cmd/ pass down to the rest of the code.
package config

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
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

// Config holds concrete settings. It deliberately has no environment name, so
// code can't branch on "are we in dev?"; ENVIRONMENT only picks defaults here.
type Config struct {
	LogFormat   LogFormat
	Port        string
	DatabaseURL string
	// GoogleClientIDs are the OAuth client IDs whose Google ID tokens are
	// accepted (the token's aud). See docs/AUTH.md.
	GoogleClientIDs []string
}

// Load reads configuration through getenv (os.Getenv outside tests). prod has
// no default DATABASE_URL, so it's required there. GOOGLE_CLIENT_IDS has no
// default anywhere, but only the API server requires it, so it's validated
// where it's used (googleauth.New), not here.
func Load(getenv func(string) string) (Config, error) {
	var defaults Config
	switch env := cmp.Or(getenv("ENVIRONMENT"), "prod"); env {
	case "prod":
		defaults = Config{LogFormat: LogFormatJSON, Port: "8080"}
	case "dev":
		defaults = Config{LogFormat: LogFormatPretty, Port: "8080", DatabaseURL: localDatabaseURL}
	default:
		return Config{}, fmt.Errorf("unknown ENVIRONMENT %q (want \"prod\" or \"dev\")", env)
	}

	cfg := Config{
		LogFormat:       defaults.LogFormat,
		Port:            cmp.Or(getenv("PORT"), defaults.Port),
		DatabaseURL:     cmp.Or(getenv("DATABASE_URL"), defaults.DatabaseURL),
		GoogleClientIDs: splitList(getenv("GOOGLE_CLIENT_IDS")),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	return cfg, nil
}

// splitList parses a comma-separated list, ignoring whitespace and empty items.
func splitList(s string) []string {
	var items []string
	for item := range strings.SplitSeq(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
