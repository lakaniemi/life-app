# Life App API

Go HTTP API for Life App. Uses the standard library `net/http`; no framework.

## Prerequisites

- [Go](https://go.dev/dl/) 1.27+
- [golangci-lint](https://golangci-lint.run/) v2
- [Docker](https://www.docker.com/) (only for building the container image)

On macOS:

```sh
brew install go golangci-lint
```

## Commands

Run from `api/`. The `make` targets are shortcuts. The equivalent raw command is shown alongside each one.

| Task             | Make                | Raw command                                         |
| ---------------- | ------------------- | --------------------------------------------------- |
| Run locally      | `make run`          | `go run ./cmd/api`                                  |
| Test             | `make test`         | `go test -race ./...`                               |
| Lint             | `make lint`         | `golangci-lint run`                                 |
| Format           | `make fmt`          | `golangci-lint fmt`                                 |
| Build binary     | `make build`        | `CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api` |
| Build image      | `make docker-build` | `docker build -t life-app-api .`                    |
| Run image        | `make docker-run`   | `docker run --rm -p 8080:8080 life-app-api`         |

The server listens on `$PORT` (default `8080`). Example: `PORT=3000 make run`.

```sh
curl -i localhost:8080/health
```

## Endpoints

| Method | Path      | Description                                       |
| ------ | --------- | ------------------------------------------------- |
| GET    | `/health` | Liveness check. Returns `{"status":"ok"}`.        |

## Layout

```
cmd/api/          entrypoint: config, logger, server lifecycle
internal/server/  routes, handlers, middleware, and their tests
```

CI (`.github/workflows/api.yml`) runs lint, tests, and a Docker build on every PR that touches `api/`.
