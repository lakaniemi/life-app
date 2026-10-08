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

## Google Cloud setup

Sign-in needs two OAuth clients in the same Google Cloud project:

| Client      | What it identifies | Where it appears |
| ----------- | ------------------ | ---------------- |
| **Web**     | Our backend: the party the token is *for* | It's passed to Credential Manager as `serverClientId`, so it becomes the ID token's `aud`. It's the value in `GOOGLE_CLIENT_IDS` |
| **Android** | Our app, by package name + signing certificate SHA-1 | Never in tokens. Google Play services uses it to check that the app requesting sign-in really is ours |

The "Web" name is misleading for a mobile-only product. It's simply the client type Google uses to represent a backend server. No client secret or redirect URI is used.

## ID token verification

`api/internal/googleauth` verifies the token. It wraps [go-oidc](https://github.com/coreos/go-oidc) and implements Google's [backend verification checklist](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token):

| Check | What it stops |
| ----- | ------------- |
| **Signature** against Google's keys (`https://www.googleapis.com/oauth2/v3/certs`). Only RS256 is accepted | Forged tokens. Google rotates keys; go-oidc caches them and refetches on an unknown key ID. Accepting RS256 only rules out the classic `alg: none` and algorithm-confusion attacks |
| **`iss`** is `https://accounts.google.com` (or `accounts.google.com`, which Google also uses) | Tokens from another identity provider |
| **`exp`** has not passed | Old tokens. Google ID tokens live about an hour |
| **`aud`**: every audience is in `GOOGLE_CLIENT_IDS` | **Token substitution.** A Google ID token issued to *any other* app is just as validly signed. Without this check, a third-party app with Google sign-in could replay its users' tokens to log in as them here |

The audience check is ours, not go-oidc's. go-oidc's built-in check takes a single client ID and only requires the token to *contain* it. We take a list, and follow [OIDC Core §3.1.3.7](https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation), which says to reject a token with "additional audiences not trusted by the Client".

We use a library rather than hand-rolled JWT parsing, as Google "strongly recommend[s]". Signature verification is easy to get subtly wrong.

The nonce is checked separately, because it needs the database. See [Nonce](#nonce).

## Storage

Two tables (`api/internal/migrations/00002_auth.sql`):

- **`sessions`** stores the SHA-256 of each session token, never the token itself. A database leak then doesn't hand out working sessions. A fast hash is enough because the token is 32 random bytes, so there's nothing to brute-force. Slow hashes such as bcrypt exist for guessable passwords. `id` is separate from `token_hash`, so a session can be referred to (e.g. on logout) without the secret.
- **`auth_nonces`** stores nonces in plain text. A nonce isn't a credential: on its own it's useless without a Google-signed ID token that contains it, and it travels inside that token anyway. The `expires_at` index serves the cleanup of expired rows.
