// Package migrations holds the database schema migrations, embedded into any
// binary that imports it.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Dir is the migrations directory relative to the module root, used when
// creating new migration files on disk.
const Dir = "internal/migrations"

//go:embed *.sql
var fsys embed.FS

// NewProvider returns a goose provider that applies the embedded migrations
// to db. It holds a Postgres advisory lock while migrating, so two processes
// (e.g. an API starting while cmd/migrate runs) never migrate at once.
func NewProvider(db *sql.DB) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("create migration lock: %w", err)
	}
	return goose.NewProvider(goose.DialectPostgres, db, fsys, goose.WithSessionLocker(locker))
}
