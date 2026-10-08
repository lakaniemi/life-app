package config

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"

	"github.com/joho/godotenv"
)

// WithDotEnv returns a getenv that falls back to the variables in the file at
// path when the real environment doesn't set them. A missing file is not an
// error. Nothing is written to the process environment, so the values only
// reach code that is handed this getenv.
func WithDotEnv(getenv func(string) string, path string) (func(string) string, error) {
	values, err := godotenv.Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return getenv, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return func(key string) string {
		return cmp.Or(getenv(key), values[key])
	}, nil
}
