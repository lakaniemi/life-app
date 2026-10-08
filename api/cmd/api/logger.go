package main

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

// newLogger returns a JSON logger for production (Cloud Logging parses it)
// and a coloured, human-readable one for local development.
func newLogger(env string, w io.Writer) (*slog.Logger, error) {
	switch env {
	case "prod":
		return slog.New(slog.NewJSONHandler(w, nil)), nil
	case "dev":
		return slog.New(tint.NewTextHandler(w, &tint.Options{TimeFormat: time.TimeOnly})), nil
	default:
		return nil, fmt.Errorf("unknown ENVIRONMENT %q (want \"prod\" or \"dev\")", env)
	}
}
