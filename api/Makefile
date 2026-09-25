# Thin shortcuts over go/docker commands. Each target is a one-liner you
# can also run directly.

IMAGE := life-app-api

.PHONY: run test lint fmt build docker-build docker-run

run:
	go run ./cmd/api

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
