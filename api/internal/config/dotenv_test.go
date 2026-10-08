package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWithDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("FROM_FILE=file\nBOTH=file\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	processEnv := map[string]string{"BOTH": "real"}

	getenv, err := WithDotEnv(func(key string) string { return processEnv[key] }, path)
	if err != nil {
		t.Fatalf("WithDotEnv() error = %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{key: "BOTH", want: "real"},
		{key: "FROM_FILE", want: "file"},
		{key: "UNSET", want: ""},
	}
	for _, tt := range tests {
		if got := getenv(tt.key); got != tt.want {
			t.Errorf("getenv(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestWithDotEnvMissingFile(t *testing.T) {
	processEnv := map[string]string{"KEY": "real"}

	getenv, err := WithDotEnv(func(key string) string { return processEnv[key] }, filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatalf("WithDotEnv() error = %v", err)
	}
	if got := getenv("KEY"); got != "real" {
		t.Errorf("getenv(KEY) = %q, want %q", got, "real")
	}
}
