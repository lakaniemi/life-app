# Phase 1 — DB foundation (PR 1)

Get the API talking to Postgres with versioned migrations and generated queries. No new endpoints.

## Scope

- **Local Postgres:** `api/compose.yaml` with a pinned current Postgres major (chosen at phase start) and a named volume. `make db-up` / `make db-down`.
- **Config:** `DATABASE_URL` is read only in `cmd/api` (per `api/CLAUDE.md`). Document a local default in `api/README.md`.
- **Migrations with goose** (`github.com/pressly/goose/v3`):
  - Plain SQL files in `api/migrations/`, each with `-- +goose Up` / `-- +goose Down` sections, embedded into the binary with `embed.FS`.
  - The goose CLI is added as a `tool` in `go.mod`, the same way `air` is.
  - Make targets: `migrate-up`, `migrate-down`, `migrate-status`, `migrate-new name=…`.
  - The first migration creates `users`, `families` and `family_members` (see [01-data-model.md](01-data-model.md)).
  - A `cmd/migrate` entrypoint runs the embedded migrations from the same image, so a deployment can run them without the CLI. *When* it runs in deployment is decided later.
- **DB access with pgx + sqlc:**
  - `pgx/v5` with `pgxpool`. `cmd/api` creates the pool, pings it at startup and closes it on shutdown.
  - `api/sqlc.yaml` uses `migrations/` as the schema. sqlc understands goose annotations, so there's one source of truth for the schema.
  - Queries go in `api/internal/db/queries/*.sql`, and the generated Go goes in `api/internal/db/` (committed).
  - `make generate`. CI checks that the generated code is up to date with `sqlc diff`.
  - `server.New` takes the `*db.Queries` (and the pool, for transactions) as dependencies. That's still constructor injection, with no globals.
- **Tests:**
  - Integration tests read `TEST_DATABASE_URL` and call `t.Skip` when it's unset, so `make test` still works without a database.
  - A test helper migrates a fresh database.
  - CI gets a `services: postgres` container (same version as compose) and sets `TEST_DATABASE_URL`.

## Teaching focus

- `database/sql` vs pgx, and why to use pgx directly for a Postgres-only service.
- What sqlc generates, read alongside the SQL.
- `context.Context` flowing into every query: cancellation when the client disconnects.
- `errors.Is(err, pgx.ErrNoRows)`.
- Why the pool lives in `main` and is passed down.

## Done when

- `make db-up migrate-up` creates the tables locally.
- `make fmt lint test` passes with and without `TEST_DATABASE_URL`.
- CI is green, including the integration tests and the `sqlc diff` check.
- The root `CLAUDE.md` decisions table records goose and sqlc + pgx.
