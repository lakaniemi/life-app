# Life App API

Go HTTP API for Life App. Uses the standard library `net/http`; no framework. Data lives in PostgreSQL.

## Prerequisites

- [Go](https://go.dev/dl/) 1.27+
- [golangci-lint](https://golangci-lint.run/) v2
- [Docker](https://www.docker.com/), for local Postgres and the container image

On macOS:

```sh
brew install go golangci-lint
```

## Getting started

Run from `api/`:

```sh
make db-up        # start Postgres 18 in Docker
make migrate-up   # create the tables
make dev          # run with hot reload
curl -i localhost:8080/health
```

## Commands

```sh
make dev            # run locally with hot reload
make run            # run locally
make test           # run tests (needs make db-up)
make cover          # run tests, print coverage and open the HTML report
make lint           # lint
make fmt            # format
make generate       # regenerate Go code from SQL queries (sqlc)
make build          # build binary to bin/api
make docker-build   # build container image
make docker-run     # run the image like prod: Postgres, migrations, then the API on port 8080

make db-up          # start local Postgres
make db-down        # stop it; data is kept in a Docker volume

make migrate-up               # apply pending migrations
make migrate-down             # roll back the latest migration
make migrate-status           # list migrations and whether each is applied
make migrate-new name=<name>  # create a new SQL migration file
```

## Configuration

`ENVIRONMENT` (`prod` by default, or `dev`) picks the defaults for everything else. The `make` targets that run locally (`run`, `dev`, `migrate-*`) set `ENVIRONMENT=dev`. An env var always overrides its default.

| Variable            | `dev` default         | `prod` default | Notes |
| ------------------- | --------------------- | -------------- | ----- |
| `DATABASE_URL`      | Local Compose Postgres | None: required | Used by `cmd/api` and `cmd/migrate`. |
| `PORT`              | `8080`                | `8080`         | Example: `PORT=3000 make run`. |
| (log format)        | Coloured text         | JSON           | Not configurable separately. |

`TEST_DATABASE_URL` points the database tests at a Postgres server; without it they fail. Each test creates and drops its own database there. `make test` defaults it to local Postgres.

## Database

- **Migrations** are SQL files in `internal/migrations/` (goose format, `-- +goose Up` / `-- +goose Down`). They're embedded in the binaries. `cmd/migrate` runs them locally and in deployment: the image includes `/migrate` next to `/api`.
- **Queries** are written as SQL in `internal/db/queries/`. sqlc turns them into typed Go functions in `internal/db/`. After changing a query or the schema, run `make generate` and commit the result. CI fails if the generated code is stale.
