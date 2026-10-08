package main

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"

	"github.com/lakaniemi/life-app/api/internal/config"
)

func newLogger(format config.LogFormat, w io.Writer) (*slog.Logger, error) {
	switch format {
	case config.LogFormatJSON:
		return slog.New(slog.NewJSONHandler(w, nil)), nil
	case config.LogFormatPretty:
		return slog.New(tint.NewTextHandler(w, &tint.Options{TimeFormat: time.TimeOnly})), nil
	default:
		return nil, fmt.Errorf("unknown log format %q", format)
	}
}
