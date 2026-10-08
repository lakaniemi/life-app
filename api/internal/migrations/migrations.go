// Package migrations holds the database schema migrations, embedded into any
// binary that imports it.
package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

// Dir is the migrations directory relative to the module root, used when
// creating new migration files on disk.
const Dir = "internal/migrations"

//go:embed *.sql
var fsys embed.FS

// NewProvider returns a goose provider that applies the embedded migrations
// to db.
func NewProvider(db *sql.DB) (*goose.Provider, error) {
	return goose.NewProvider(goose.DialectPostgres, db, fsys)
}
