IMAGE := life-app-api

.PHONY: run dev test test-db lint fmt build docker-build docker-run db-up db-down \
	migrate-up migrate-down migrate-status migrate-new generate

run:
	ENVIRONMENT=dev go run ./cmd/api

dev:
	ENVIRONMENT=dev go tool air

test:
	go test -race ./...

# Also runs the integration tests against local Postgres. Each test creates
# and drops its own database, so the local dev data is untouched.
test-db:
	TEST_DATABASE_URL='postgres://lifeapp:lifeapp@localhost:5432/lifeapp?sslmode=disable' go test -race ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api

docker-build:
	docker build -t $(IMAGE) .

# Inside the container, localhost is the container itself; Docker Desktop
# exposes the host as host.docker.internal.
docker-run:
	docker run --rm -p 8080:8080 \
		-e DATABASE_URL=$(subst localhost,host.docker.internal,$(DATABASE_URL)) \
		$(IMAGE)

db-up:
	docker compose up -d --wait

db-down:
	docker compose down

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
