package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "dev defaults",
			env:  map[string]string{"ENVIRONMENT": "dev"},
			want: Config{LogFormat: LogFormatPretty, Port: "8080", DatabaseURL: localDatabaseURL},
		},
		{
			name: "prod is the default environment",
			env:  map[string]string{"DATABASE_URL": "postgres://prod"},
			want: Config{LogFormat: LogFormatJSON, Port: "8080", DatabaseURL: "postgres://prod"},
		},
		{
			name:    "prod has no default database",
			env:     map[string]string{"ENVIRONMENT": "prod"},
			wantErr: true,
		},
		{
			name: "env vars override dev defaults",
			env: map[string]string{
				"ENVIRONMENT":  "dev",
				"PORT":         "3000",
				"DATABASE_URL": "postgres://other",
			},
			want: Config{LogFormat: LogFormatPretty, Port: "3000", DatabaseURL: "postgres://other"},
		},
		{
			name:    "unknown environment",
			env:     map[string]string{"ENVIRONMENT": "staging", "DATABASE_URL": "postgres://x"},
			wantErr: true,
		},
		{
			name: "client ID list is trimmed and skips empty items",
			env:  map[string]string{"ENVIRONMENT": "dev", "GOOGLE_CLIENT_IDS": " web , ,ios,"},
			want: Config{LogFormat: LogFormatPretty, Port: "8080", DatabaseURL: localDatabaseURL, GoogleClientIDs: []string{"web", "ios"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }

			got, err := Load(getenv)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() = %+v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
