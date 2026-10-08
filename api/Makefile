IMAGE := life-app-api

# Local Postgres from compose.yaml. Override on the command line or in the
# environment to point at another database.
DATABASE_URL ?= postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable
export DATABASE_URL

.PHONY: run dev test lint fmt build docker-build docker-run db-up db-down

run:
	ENVIRONMENT=dev go run ./cmd/api

dev:
	ENVIRONMENT=dev go tool air

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 $(IMAGE)

db-up:
	docker compose up -d --wait

db-down:
	docker compose down
