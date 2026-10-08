# syntax=docker/dockerfile:1

FROM golang:1.27 AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
# Static binary (no libc), required by distroless/static.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/api /api
# Run migrations from the same image by overriding the entrypoint:
# /migrate up (see compose.yaml).
COPY --from=build /out/migrate /migrate
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
