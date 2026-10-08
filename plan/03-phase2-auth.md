# Phase 2 — Backend auth (PR 2)

## Backend first, app later

The app does Google sign-in natively and gets a Google **ID token**: a JWT signed by Google that says who the user is. The API only *verifies* that token. It never handles OAuth redirects or a client secret. So the whole API contract is one endpoint, and it can be built and tested before the app exists:

- **Unit/handler tests** use a fake token verifier behind an interface.
- **Manual testing** uses real tokens: set the Google OAuth Playground to use *our own* Web client ID, and the ID tokens it issues have the right audience.
- The app's sign-in library (phase 5) issues ID tokens with the same Web client ID as `aud`, so the backend config doesn't change.

*Re-check the Playground and audience details against current Google docs when starting this phase.*

## Flow

```
App ── Google sign-in ──> ID token
App ── POST /auth/google {idToken} ──> API
       API: verify signature, iss, aud, exp (go-oidc, JWKS cached)
       API: find user by google_sub, or create them (name from the token)
       API: create session → { token, user }
App ── Authorization: Bearer <token> ──> every later request
```

## Scope

- **Wiring (carried over from phase 1):** `server.New` takes the `*db.Queries` and the pool (for transactions) from `cmd/api`.
- **Migration:** add the `sessions` table.
- **Token verification:** `github.com/coreos/go-oidc/v3`, behind a small interface so tests can fake it.
- **Sessions:**
  - The token is 32 random bytes from `crypto/rand`, base64url-encoded. Only its SHA-256 is stored.
  - Expiry is 90 days and sliding: `last_used_at` and `expires_at` are bumped at most once a day.
  - Expired rows are deleted at login.
- **`requireAuth` middleware:** reads the Bearer token, hashes it, looks it up and puts the user ID in the request context under a typed context key. A missing or expired session gets 401.
- **Endpoints:**
  - `POST /auth/google` `{idToken}` → `{token, user}`
  - `POST /auth/logout` deletes the current session.
  - `GET /me` → the user and their families (the families list stays empty until phase 3).
  - `PATCH /me` `{name}`
- **Config:** `GOOGLE_CLIENT_IDS`, a comma-separated list of allowed audiences, read in `cmd/api`.

## Teaching focus

- ID token vs access token: the API needs the *identity* (ID token), not access to Google APIs (access token).
- Opaque tokens vs JWTs: revocation, and the trade-off of one DB lookup per request.
- `context.WithValue` with an unexported key type, compared with Express's `req.user`.
- Middleware composition with plain `http.Handler`.

## Done when

- Handler tests cover valid, invalid and expired tokens, a missing Bearer header, and logout.
- A Playground-issued ID token logs in against the local API.
- The root `CLAUDE.md` "Auth" row is filled in.
