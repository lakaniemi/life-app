# Authentication

How users sign in, and why it works this way. Each decision links to its source, so it can be re-checked later. The [decision log](#decision-log) summarizes them.

## Overview

Google proves who the user is, once, at sign-in. After that, the API uses its own session token and Google isn't involved again.

```
App ── POST /auth/nonce ─────────────────────────▶ API  stores nonce (10 min TTL)
App ◀─ {nonce} ──────────────────────────────────
App ── Credential Manager (serverClientId = Web client ID, nonce) ──▶ Google
App ◀─ ID token {iss, aud=Web client ID, sub, name, exp, nonce} ────
App ── POST /auth/google {idToken} ──────────────▶ API
       1. verify signature (JWKS), iss, exp
       2. check aud ⊆ GOOGLE_CLIENT_IDS
       3. consume nonce: single-use, unexpired
       4. find/create user by sub; create session
App ◀─ {token, user} ────────────────────────────
App ── Authorization: Bearer <token> ─────────────▶ every later request
```

**Terms:**
- **OAuth 2.0** is for *authorization*. It issues **access tokens** that let an app call APIs, such as Google Calendar, on the user's behalf.
- **OpenID Connect (OIDC)** builds on it for *authentication*. It issues an **ID token**: a JWT, signed by the provider, saying who the user is.

We only need identity, so the API only ever sees the ID token. Google access tokens and refresh tokens never reach it.

## Storage

Two tables (`api/internal/migrations/00002_auth.sql`):

- **`sessions`** stores the SHA-256 of each session token, never the token itself. A database leak then doesn't hand out working sessions. A fast hash is enough because the token is 32 random bytes, so there's nothing to brute-force. Slow hashes such as bcrypt exist for guessable passwords. `id` is separate from `token_hash`, so a session can be referred to (e.g. on logout) without the secret.
- **`auth_nonces`** stores nonces in plain text. A nonce isn't a credential: on its own it's useless without a Google-signed ID token that contains it, and it travels inside that token anyway. The `expires_at` index serves the cleanup of expired rows.
