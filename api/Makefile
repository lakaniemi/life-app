IMAGE := life-app-api
# Each test creates and drops its own database on this server, so local dev
# data is untouched.
TEST_DATABASE_URL ?= postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable

.PHONY: run dev test cover cover-html lint fmt build docker-build docker-run db-up db-down psql \
	migrate-up migrate-down migrate-status migrate-new generate dev-session

run:
	ENVIRONMENT=dev go run ./cmd/api

dev:
	ENVIRONMENT=dev go tool air

test:
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -race ./...

# -coverpkg counts coverage across packages, so code exercised by endpoint
# tests (e.g. sqlc queries) shows up as covered.
cover:
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -race -coverprofile=coverage.out -coverpkg=./internal/... ./...
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker compose up --build

db-up:
	docker compose up -d --wait postgres

db-down:
	docker compose down postgres

# Runs psql inside the container, so no local install is needed.
psql:
	docker compose exec postgres psql -U lifeapp -d lifeapp

migrate-up:
	ENVIRONMENT=dev go run ./cmd/migrate up

migrate-down:
	ENVIRONMENT=dev go run ./cmd/migrate down

migrate-status:
	ENVIRONMENT=dev go run ./cmd/migrate status

migrate-new:
	@test -n "$(name)" || (echo "usage: make migrate-new name=<name>" && exit 1)
	go run ./cmd/migrate create $(name)

generate:
	go tool sqlc generate

# Prints a session token for dev user $(name), creating the user if needed.
# Use -s to capture it: TOKEN=$$(make -s dev-session name=Alice)
dev-session:
	@test -n "$(name)" || (echo "usage: make dev-session name=<name>" >&2 && exit 1)
	@ENVIRONMENT=dev go run ./cmd/devsession -name '$(name)'
