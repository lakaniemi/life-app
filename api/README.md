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
cp .env.example .env   # then fill in GOOGLE_CLIENT_IDS
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
make cover          # run tests and print coverage
make cover-html     # same, then open the HTML report
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

Variables can also go in `api/.env` (copy `.env.example`). Variables that are already set in the environment take precedence over the file. It's git-ignored and excluded from the container image, so deployments use their real environment.

| Variable            | `dev` default         | `prod` default | Notes |
| ------------------- | --------------------- | -------------- | ----- |
| `DATABASE_URL`      | Local Compose Postgres | None: required | Used by `cmd/api` and `cmd/migrate`. |
| `GOOGLE_CLIENT_IDS` | None: required        | None: required | Comma-separated OAuth client IDs whose Google ID tokens are accepted (the Web client ID). Only `cmd/api` needs it. See [docs/AUTH.md](../docs/AUTH.md). |
| `PORT`              | `8080`                | `8080`         | Example: `PORT=3000 make run`. |
| (log format)        | Coloured text         | JSON           | Not configurable separately. |

`TEST_DATABASE_URL` points the database tests at a Postgres server; without it they fail. Each test creates and drops its own database there. `make test` defaults it to local Postgres.

## Testing sign-in manually

The automated tests fake Google. To test with a real Google ID token before the app exists, use Google's authorization code flow from a terminal. It yields the same kind of token the app sends: one carrying our nonce, with the Web client ID as its audience. See [docs/AUTH.md](../docs/AUTH.md) for the flow itself.

**One-time setup** in the Cloud Console (Google Auth Platform):
1. Under **Branding**, fill in the consent screen.
2. Under **Audience**, choose External, keep it in Testing, and add your Google account as a test user.
3. Under **Clients**, create a **Web application** client with the authorized redirect URI `http://localhost:9999`. Nothing needs to listen there.
4. Put the client ID in `api/.env` as `GOOGLE_CLIENT_IDS`.
5. Keep the client secret in your shell only. The API never uses it; it's only needed for step 3 below.

With `make db-up migrate-up run` going, in another terminal:

```sh
CLIENT_ID='<client ID>'; CLIENT_SECRET='<client secret>'; REDIRECT='http://localhost:9999'

# 1. Get a nonce (valid 10 min; do the rest promptly)
NONCE=$(curl -s -X POST localhost:8080/auth/nonce | jq -r .nonce)

# 2. Open this URL, sign in, then copy code=… from the address bar (as-is, URL-encoded)
echo "https://accounts.google.com/o/oauth2/v2/auth?client_id=$CLIENT_ID&response_type=code&scope=openid%20profile&redirect_uri=$REDIRECT&state=manual-test&nonce=$NONCE"
CODE='<code>'

# 3. Exchange the code for an ID token (null = code expired/used or redirect mismatch)
ID_TOKEN=$(curl -s https://oauth2.googleapis.com/token -d "code=$CODE" -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" -d "redirect_uri=$REDIRECT" -d "grant_type=authorization_code" | jq -r .id_token)

# 4. Log in, then exercise the session
TOKEN=$(curl -s -X POST localhost:8080/auth/google -d "{\"idToken\":\"$ID_TOKEN\"}" | jq -r .token)
curl -s -X POST localhost:8080/auth/google -d "{\"idToken\":\"$ID_TOKEN\"}"   # replay: 401 invalid_nonce
curl -s localhost:8080/me -H "Authorization: Bearer $TOKEN"                   # 200
curl -s -X POST localhost:8080/auth/logout -H "Authorization: Bearer $TOKEN"  # 204
curl -s localhost:8080/me -H "Authorization: Bearer $TOKEN"                   # 401
```

## Database

- **Migrations** are SQL files in `internal/migrations/` (goose format, `-- +goose Up` / `-- +goose Down`). They're embedded in the binaries. `cmd/migrate` runs them locally and in deployment: the image includes `/migrate` next to `/api`.
- **Queries** are written as SQL in `internal/db/queries/`. sqlc turns them into typed Go functions in `internal/db/`. After changing a query or the schema, run `make generate` and commit the result. CI fails if the generated code is stale.
