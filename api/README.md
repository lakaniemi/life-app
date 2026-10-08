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

Run from `api/`:

```sh
make run            # run locally
make test           # run tests
make lint           # lint
make fmt            # format
make build          # build binary to bin/api
make docker-build   # build container image
make docker-run     # run container image on port 8080
```

The server listens on `$PORT` (default `8080`). Example: `PORT=3000 make run`.

`$ENVIRONMENT` is `prod` (default, JSON logs) or `dev` (coloured logs). The `make` run targets set `dev`.

```sh
curl -i localhost:8080/health
```
