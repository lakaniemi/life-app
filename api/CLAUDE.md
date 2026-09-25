# CLAUDE.md (api/)

- **Go 1.27, stdlib only.** Routing uses `net/http` ServeMux method and path patterns (`"GET /items/{id}"`, read with `r.PathValue("id")`). Don't add a router or framework without discussing it first.
- **Layout:** `cmd/api` is the entrypoint and only does wiring. All other code lives under `internal/`, which the compiler makes unimportable from outside this module. No `pkg/`.
- **Routes** are all registered in `internal/server/routes.go`. Handlers are `handleX(deps...) http.Handler` constructors that take their dependencies as arguments, with no globals.
- **Adding an endpoint:**
  1. Write the handler in its own file in `internal/server`.
  2. Register it in `routes.go`.
  3. Add a table-driven test next to it that goes through `New(...)`, so routing is tested too.
- **Before committing:** `make fmt lint test` must pass. CI runs the same checks.
- **Don't use URL paths ending in `z`** (e.g. `/healthz`). Cloud Run reserves some of them.
