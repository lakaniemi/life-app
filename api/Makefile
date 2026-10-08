IMAGE := life-app-api
# Each test creates and drops its own database on this server, so local dev
# data is untouched.
TEST_DATABASE_URL ?= postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable

.PHONY: run dev test lint fmt build docker-build docker-run db-up db-down \
	migrate-up migrate-down migrate-status migrate-new generate

run:
	ENVIRONMENT=dev go run ./cmd/api

dev:
	ENVIRONMENT=dev go tool air

test:
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -race ./...

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
