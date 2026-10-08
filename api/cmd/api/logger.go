package main

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

// prod logs JSON because Cloud Logging parses it.
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
