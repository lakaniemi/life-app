# CLAUDE.md (api/)

- **No web framework.** Routing uses `net/http` ServeMux method and path patterns (`"GET /items/{id}"`, read with `r.PathValue("id")`). Don't add a router or framework without discussing it first.
- **Layout:** `cmd/api` (server), `cmd/migrate` (migrations) and `cmd/devsession` (local dev sessions, not shipped in the image) are entrypoints and only do wiring. Never add a dev-only sign-in route to the API; local auth goes through `cmd/devsession`. All other code lives under `internal/`, which the compiler makes unimportable from outside this module. No `pkg/`.
- **Packages by responsibility, not by layer.** No `handlers/`, `services/` or `utils/` packages.
  - `internal/server` is the whole HTTP layer: routes, handlers, middleware and JSON helpers, all unexported.
  - Domain logic goes in packages named for what they provide (e.g. `internal/auth`). They don't import `net/http`, and they report caller mistakes as sentinel errors, which `internal/server` maps to status codes.
  - Every resource's operations go through its domain package, even plain CRUD, so handlers never run queries or import `pgx`. Domain packages own data access, map `pgx.ErrNoRows` to their own sentinels, and decide who may act. They return sqlc row types; there are no separate domain models.
  - Adapters for external services go in a subpackage of the domain that uses them (e.g. `internal/auth/google`). The domain depends on them through a small interface it defines itself (`auth.TokenVerifier`).
  - Inside a package, organize with file names: one file per route group or concern.
- **Naming:** the further a name is used from where it's declared, the more descriptive it is. One-letter names only for receivers, `w`/`r` in handlers, `t`/`tt` in tests, loop indices, and variables that live a few lines. Don't name variables after domain packages (`family`, `auth`), since that shadows the package; say what they hold instead (`familyRow`, `familyService`).
- **Routes** are all registered in `internal/server/routes.go`. Handlers are `handleX(deps...) apiHandler` constructors that take their dependencies as arguments, with no globals. A handler starts by reading its inputs into named locals: `ctx := r.Context()` when used more than once, then the caller (`actor := familyMembershipFrom(ctx)` or `userID := sessionFrom(ctx).UserID`), path values and the body. Then it calls the service with them, rather than nesting lookups in the call's arguments.
- **Errors:** an `apiHandler` returns its errors instead of writing error responses, and its `ServeHTTP` maps them in `writeErrorResponse` (`internal/server/errors.go`):
  - an `*apiError` (`newAPIError(status, code, message)`) for failures the HTTP layer decides, such as a bad body;
  - a domain sentinel through the `domainErrors` table, where new sentinels are added;
  - anything else becomes a logged 500 that doesn't leak details.

  Return an error only before writing anything. `return writeJSON(...)` is safe, since `writeJSON` encodes before it sends. Middlewares stay plain `http.Handler`s and call `writeErrorResponse` themselves.
- **Adding an endpoint:**
  1. Write the handler in the route group's file in `internal/server` (e.g. `auth.go`, `me.go`). It calls the resource's domain package, where the data access and rules go.
  2. Register it in `routes.go`.
  3. For a request body, declare normalization in go-playground/mold `mod` struct tags (e.g. `mod:"trim"` on user-entered text) and rules in go-playground/validator `validate` tags on the request struct. Then return the error from `decodeAndValidateRequest(w, r, &req)`, which decodes, normalizes and validates in that order. Validation failures respond 400 `invalid_body` with a `fields` list (`{"field","rule"}`, using the JSON field names).
  4. For a route under `/families/{familyID}`, register it with `withMiddlewares(requireAuth, requireFamilyMembership)(handler)` and pass the caller's membership (`familyMembershipFrom`) to the domain service as the actor. The service acts only on the actor's family and checks the role itself, under the family lock (see the `internal/family` package doc). Every query on family-owned data filters by that family ID.
  5. Add endpoint tests next to it. Add a family-scoped route to `TestFamilyRoutesAuthorization` too.
- **Compatibility:** installed apps lag behind the API, so changes stay backward compatible, following Google's AIP-180. The reasoning is in `app/docs/ARCHITECTURE.md`, "Version compatibility".
  - Endpoints, response fields and enum values may be added. Nothing is removed, renamed, retyped or reformatted, and no default changes.
  - No new required request fields. A new request field is allowed only if leaving it out means "behave as before". Never invent a default to make a field optional.
  - Updates are `PATCH` with partial bodies, so an old app can't wipe fields it doesn't know.
  - Anything else is a breaking change. Either add a new endpoint next to the old one and remove the old one once no supported app uses it, or raise the API's minimum app compatibility level (a constant, so it deploys with the change) in the same change, so old apps get `app_outdated` and the "Update required" screen. The app update that adopts the change bumps the app's level to match. No URL versioning.
- **Tests:** follow `docs/TESTING_STRATEGY.md`, including its per-endpoint checklist. Endpoint tests against a real database are the default; tests that need one call `dbtest.New(t)`. Don't test generated code.
- **Config:** `cmd/*` calls `config.Load()` once and passes the resulting `config.Config` down. Nothing else reads env vars, except `internal/dbtest` (`TEST_DATABASE_URL`), because tests have no `main`. `ENVIRONMENT` only selects an entry in the `defaults` map in `config.go` and isn't exposed, so no code branches on dev vs prod. Add a new setting as a `Config` field, with defaults in that map where they make sense.
- **Database:**
  - pgx/v5 with pgxpool directly, not `database/sql`, since the API is Postgres-only.
  - Schema changes are new goose migrations in `internal/migrations/` (`make migrate-new name=…`). Never edit a migration that has been merged. Migrations run through our own `cmd/migrate`, not the goose CLI, so local runs, tests and deployment share one code path.
  - Queries are SQL in `internal/db/queries/`. `internal/db/*.go` is generated by sqlc (`make generate`), so don't edit it by hand. sqlc is pinned as a `go tool`, not installed separately, because generated files embed its version and CI runs `sqlc diff`.
  - **Schema conventions:**
    - `timestamptz` for every timestamp.
    - Primary keys are UUIDs from `gen_random_uuid()`. `uuidv7()` would need Postgres 18, and the SQL stays portable to 17.
    - `updated_at` is set explicitly in each UPDATE query, with no triggers.
    - Hard deletes, with no `deleted_at` columns.
    - Authorization lives in Go, not Postgres row-level security.
    - A fixed set of values is `text` with a CHECK constraint, not a Postgres ENUM, whose values can't easily be removed or renamed.
    - No owner columns. Control over a family is expressed only through `family_members.role`, so there's one source of truth.
    - Every family-owned table has `family_id uuid NOT NULL REFERENCES families ON DELETE CASCADE`, and every query on it filters by `family_id`.
  - **Migrations are backward compatible with the previous API version**, which may still be running while the new one deploys. Breaking schema changes go in steps: add the new column or table, deploy code that uses it, and drop the old one in a later migration.
- **Don't use URL paths ending in `z`** (e.g. `/healthz`). Some hosting platforms (e.g. Cloud Run) reserve them.
