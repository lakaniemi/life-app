# API testing strategy

The goal is high confidence from few tests. Most tests exercise endpoints end to end against a real Postgres. Everything else is the exception and needs a reason.

## Principles

- **Test the HTTP contract, not the implementation.** The app depends on routes, status codes and JSON shapes. Tests written against those survive refactoring. Tests of internals break when the code changes, even though nothing a client would notice changed.
- **Use real dependencies over test doubles.** A real database catches constraint, NULL and ordering behavior that a fake gets wrong.
- **Every test must be able to catch a plausible bug.** If you can't name the bug a test would catch, don't write it.

## Layers

In priority order:

1. **Endpoint tests (the default).** These live in `internal/server` and send requests through `server.New(...)` with `httptest`, so routing, middleware, handler, SQL and migrations all run together. Each test gets its own fresh database from `dbtest.New(t)`, so tests can run in parallel with `t.Parallel()`.
   - **Arrange:** seed data directly with sqlc queries. It's fast and makes the setup explicit.
   - **Act and assert:** through HTTP.
     - Check the status code.
     - Decode the JSON body and compare it with `cmp.Diff`. Use `cmpopts` to ignore generated IDs and timestamps.
   - **Side effects:** check a write through a follow-up request when the API exposes the result. Query the database only when it doesn't.
2. **Unit tests.** For pure functions with real branching, such as config loading, validation and domain calculations. No HTTP and no database.
3. **Direct query tests.** Only for non-trivial SQL that is awkward to reach through HTTP, such as aggregations or tricky NULL and ordering semantics. A plain CRUD query is covered by the endpoint that uses it.
4. **Migrations** need no tests of their own. Every database test applies all of them.

## Per-endpoint checklist

Write one case per behavior, not one per input combination:

- The happy path.
- One case per *kind* of validation failure (400), e.g. one missing field and one malformed field, not one case for every field.
- Not found (404), for routes with an ID.
- Conflicts (409), where uniqueness or state rules apply.
- **Authorization**, which is the most critical category. This applies to every endpoint that requires auth or acts on owned data, not just one feature:
  - An unauthenticated request is rejected (401).
  - A caller without access is rejected (403 or 404, whichever the API's convention is). This covers another user's or another family's data, and actions that need a role the caller doesn't have.

  A missing authorization check is a bug class to test for in every new endpoint.

## Don't test

- **sqlc-generated code.** The endpoint tests cover it.
- **Stdlib behavior**, such as ServeMux's 404 and 405 responses. The happy-path case already proves a route is registered.
- **Log output**, and **exact error wording** unless clients depend on it.
- **Private helpers** that endpoint tests already exercise.
- **Wiring in `cmd/*`.**
- **Whole responses via snapshot or golden files.** Assert the fields that matter.

## Test doubles

Don't use test doubles for the database. Use them only for external services and non-determinism, such as Google token verification or the clock. Put each one behind a small interface defined where it's used, and pass it into `server.New` like any other dependency.

## Conventions

- **Structure:** use table-driven tests when cases share a shape. Otherwise, write separate test functions.
- **Naming:** subtest names describe behavior, e.g. `"other family's list is not found"`.
- **Context and failures:** use `t.Context()`. Use `t.Fatalf` only when the rest of the test can't meaningfully run. Otherwise use `t.Errorf`.
- **Assertions:** use the stdlib, plus `github.com/google/go-cmp` for comparing structs and decoded bodies. Don't add assertion frameworks.
- **Shared helpers:** keep them in a `_test.go` file in `internal/server`, e.g. `newTestServer(t)` (returns the handler and a `*db.Queries` for seeding) and `doJSON(...)` (builds and runs a request).

## Coverage

`make cover` shows coverage locally, and `make cover-html` opens the HTML report. CI posts a summary as a PR comment, updated on each push. Coverage is measured across all of `internal/` (`-coverpkg`), so endpoint tests count toward the queries they run, not just the handlers.

There's no threshold. Use the report to find untested branches in handlers, not as a number to raise.

## Running

`make db-up`, then `make test`. Tests fail without `TEST_DATABASE_URL`, which `make test` sets to local Postgres. CI runs the same suite against its own Postgres service.
