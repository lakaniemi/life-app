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
make test           # run tests; database tests skip
make test-db        # run tests including database tests (needs make db-up)
make lint           # lint
make fmt            # format
make generate       # regenerate Go code from SQL queries (sqlc)
make build          # build binary to bin/api
make docker-build   # build container image
make docker-run     # run container image on port 8080, against local Postgres

make db-up          # start local Postgres
make db-down        # stop it; data is kept in a Docker volume

make migrate-up               # apply pending migrations
make migrate-down             # roll back the latest migration
make migrate-status           # list migrations and whether each is applied
make migrate-new name=<name>  # create a new SQL migration file
```

## Configuration

| Variable            | Default                                    | Notes |
| ------------------- | ------------------------------------------ | ----- |
| `DATABASE_URL`      | Required. `make` sets it to local Postgres: `postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable` | Read by `cmd/api` and `cmd/migrate`. |
| `PORT`              | `8080`                                     | Example: `PORT=3000 make run`. |
| `ENVIRONMENT`       | `prod` (JSON logs)                         | `dev` gives coloured logs. `make run` and `make dev` set it. |
| `TEST_DATABASE_URL` | Unset (database tests skip)                | Any database on the target server. Each test creates and drops its own database there. `make test-db` sets it. |

## Database

- **Migrations** are SQL files in `internal/migrations/` (goose format, `-- +goose Up` / `-- +goose Down`). They're embedded in the binaries. `cmd/migrate` runs them locally and in deployment: the image includes `/migrate` next to `/api`.
- **Queries** are written as SQL in `internal/db/queries/`. sqlc turns them into typed Go functions in `internal/db/`. After changing a query or the schema, run `make generate` and commit the result. CI fails if the generated code is stale.
