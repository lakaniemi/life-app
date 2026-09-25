# syntax=docker/dockerfile:1

# --- Build stage: full Go toolchain, discarded after the build. ---
FROM golang:1.27 AS build
WORKDIR /src

# Download dependencies in their own layer, so they're only re-fetched when
# go.mod/go.sum change, not on every source edit.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 produces a fully static binary with no libc dependency, which
# is what lets it run on distroless/static. -trimpath drops local file paths;
# -s -w strip debug symbols to shrink the binary.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# --- Runtime stage: no shell, no package manager, just CA certs and tzdata. ---
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
