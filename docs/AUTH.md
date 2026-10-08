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

`api/internal/auth/google` verifies the token. It wraps [go-oidc](https://github.com/coreos/go-oidc) and implements Google's [backend verification checklist](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token):

| Check | What it stops |
| ----- | ------------- |
| **Signature** against Google's keys (`https://www.googleapis.com/oauth2/v3/certs`). Only RS256 is accepted | Forged tokens. Google rotates keys; go-oidc caches them and refetches on an unknown key ID. Accepting RS256 only rules out the classic `alg: none` and algorithm-confusion attacks |
| **`iss`** is `https://accounts.google.com` (or `accounts.google.com`, which Google also uses) | Tokens from another identity provider |
| **`exp`** has not passed | Old tokens. Google ID tokens live about an hour |
| **`aud`**: every audience is in `GOOGLE_CLIENT_IDS` | **Token substitution.** A Google ID token issued to *any other* app is just as validly signed. Without this check, a third-party app with Google sign-in could replay its users' tokens to log in as them here |

The audience check is ours, not go-oidc's. go-oidc's built-in check takes a single client ID and only requires the token to *contain* it. We take a list, and follow [OIDC Core §3.1.3.7](https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation), which says to reject a token with "additional audiences not trusted by the Client".

We use a library rather than hand-rolled JWT parsing, as Google "strongly recommend[s]". Signature verification is easy to get subtly wrong.

The nonce is checked separately, because it needs the database. See [Nonce](#nonce).

## Nonce

**The threat: replay.** Suppose someone gets hold of a valid Google ID token for our app, for example from a log, a proxy or a compromised device. Signature, `iss`, `aud` and `exp` all check out for the token's roughly one-hour lifetime. Without a nonce, they could exchange it at `POST /auth/google` for a 90-day session.

**The fix.** A nonce ("number used once") ties each ID token to one login attempt that *we* started:

1. The app calls `POST /auth/nonce`. The API stores a random value (256 bits) with a 10-minute expiry.
2. The app passes the nonce to Credential Manager (`setNonce`), and Google embeds it in the ID token's `nonce` claim. Because the token is signed, the nonce can't be swapped afterwards.
3. At `POST /auth/google`, after the token itself verifies, the API **consumes** the nonce. The login is rejected if the nonce is missing, unknown, expired or already used.

Consuming the nonce is one `DELETE … WHERE nonce = $1 AND expires_at > now() RETURNING nonce` statement. Check and use happen atomically: if two requests race with the same token, Postgres's row lock lets only one of them get the row back. A separate `SELECT` followed by a `DELETE` would let both pass.

The token is verified *before* the nonce is consumed, so a forged token can't burn a legitimate user's nonce.

**Sources:**
- [OIDC Core §3.1.3.7](https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation): if a nonce was sent, the claim "MUST be present and its value checked", and the client "SHOULD check the nonce value for replay attacks".
- Android's [Sign in with Google guide](https://developer.android.com/identity/sign-in/credential-manager-siwg-implementation) recommends `setNonce()` "to prevent replay attacks", with server-side code validating that the request and response nonces are identical.

**Alternative considered: a stateless, client-generated nonce.** In this variant, from [OIDC Core §15.5.2](https://openid.net/specs/openid-connect-core-1_0.html#NonceNotes), the client keeps a random secret, sends its hash as the nonce, and later proves possession of the secret.
- It needs no table.
- But it doesn't make a nonce single-use, and a token leaked together with the request body (e.g. through server logs) would still replay.
- Server-issued, single-use nonces close both gaps, for the cost of one small table and one extra request at login.

**Cleanup.** Each `POST /auth/nonce` first deletes expired nonces, using the `expires_at` index. That keeps the table bounded without a background job.

## Sessions

After login, the API issues its own session token. Google isn't involved again until the next login.

- **Opaque, not a JWT.** The token is 32 random bytes, base64url-encoded, and means nothing on its own. Every request looks it up in `sessions`.
  - That costs one indexed query per request.
  - In return, revocation is just deleting a row: logout works immediately, and a stolen phone's session can be killed.
  - A self-contained JWT stays valid until it expires, unless you add a denylist, which brings the database lookup back anyway.
- **Why not reuse Google's ID token as the session?** It expires in about an hour, can't be revoked by us, and would tie every request to Google.
- **Entropy.** 256 bits from `crypto/rand`, well above the [OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) minimum of 64.
- **Expiry: 90 days, sliding.** Each use pushes expiry 90 days out again. To avoid a write on every request, it's extended at most once a day (`last_used_at`). An active user stays signed in; a phone left unused for 90 days signs out.
  - **Deliberate deviation:** OWASP also recommends an *absolute* timeout (a hard cap regardless of activity). Its suggested values are hours, and they're aimed at browser sessions. Mobile apps conventionally stay signed in, so there's no absolute cap for now. See [Known gaps](#known-gaps).
- **Expired sessions** are rejected exactly like unknown ones. A user's expired rows are deleted when they log in.
- **Sent as** `Authorization: Bearer <token>` ([RFC 6750](https://www.rfc-editor.org/rfc/rfc6750)), never in a URL, where it would end up in logs (RFC 6750 §5.3). A 401 carries `WWW-Authenticate: Bearer`, plus `error="invalid_token"` when a token was sent but isn't valid (§3).

In the API, `internal/auth` holds this logic, with no HTTP code: nonces, `LoginWithGoogle`, `Authenticate` and `Logout`. The `requireAuth` middleware (`api/internal/server/middleware.go`) reads the Bearer token and calls `Authenticate`. It puts the user and session IDs in the request context, and wraps each protected route in `routes.go`.

## Storage

Two tables (`api/internal/migrations/00002_auth.sql`):

- **`sessions`** stores the SHA-256 of each session token, never the token itself. A database leak then doesn't hand out working sessions. A fast hash is enough because the token is 32 random bytes, so there's nothing to brute-force. Slow hashes such as bcrypt exist for guessable passwords. `id` is separate from `token_hash`, so a session can be referred to (e.g. on logout) without the secret.
- **`auth_nonces`** stores nonces in plain text. A nonce isn't a credential: on its own it's useless without a Google-signed ID token that contains it, and it travels inside that token anyway. The `expires_at` index serves the cleanup of expired rows.

## Identity and privacy

- **Users are identified by `sub`, never by email.** Google: "Only use Google ID token sub field as identifier for the user as it is unique among all Google Accounts and never reused", while "a Google Account can have multiple email addresses at different points in time" ([source](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token)).
- **No email is requested or stored.** Sign-in asks for `openid profile` only.
- **`name`** is copied from the token on first login only. After that it belongs to the user (`PATCH /me`), and later logins never overwrite it. A token without a name (no `profile` scope) creates the user with an empty name.
- Responses use dedicated structs (`userResponse`), so internal columns such as `google_sub` can't leak into JSON.

## App side (phase 5)

*To be completed when the app's sign-in is built.*

- **Native sign-in through Credential Manager**, not a browser redirect. [RFC 8252](https://www.rfc-editor.org/rfc/rfc8252) describes the general rules for native apps: an external user-agent, never an embedded WebView (§8.12), and PKCE (§6), because apps are public clients that can't keep a secret (§8.4). Credential Manager meets the same goals another way: Google Play services authenticates the user, and verifies the calling app by its package name and signing certificate (the Android OAuth client).
- **Library choice is open.** In `@react-native-google-signin/google-signin`, a nonce on Android is supported only by the "Universal sign-in" API, and that's in the paid, licensed version. The free "Original" API takes a nonce on iOS only, and on Android it uses Google's deprecated legacy SDK ([API reference](https://react-native-google-signin.github.io/docs/api), checked 2026-10). The alternatives are another Credential Manager library, or a small Expo native module of our own.
- **The session token is stored with `expo-secure-store`** (Android Keystore), never AsyncStorage, which is unencrypted.
- On a 401, the app treats itself as signed out.

## Decision log

| Decision | Alternatives | Why | Source |
| -------- | ------------ | --- | ------ |
| Native sign-in returns a Google ID token, which the API verifies | API-driven OAuth redirect flow with a client secret | The API never handles redirects or secrets, and the whole contract is one endpoint | [Android SIWG](https://developer.android.com/identity/sign-in/credential-manager-siwg-implementation) |
| ID token, not access token | Access token + userinfo call | We need identity, not access to Google APIs | [OIDC Core](https://openid.net/specs/openid-connect-core-1_0.html) |
| go-oidc for verification | Hand-rolled JWT parsing; Google's `google.golang.org/api/idtoken` | A maintained, spec-focused library; small dependency tree | [Google](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token) "strongly recommend[s]" a library |
| Our own audience check: all `aud` values ∈ `GOOGLE_CLIENT_IDS` | go-oidc's single-`ClientID` "contains" check | Allows several clients, and rejects untrusted extra audiences | [OIDC Core §3.1.3.7](https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation) |
| Server-issued, single-use nonce | No nonce; stateless client-generated nonce | Blocks replay of a leaked ID token, even one leaked with its request | [OIDC Core §3.1.3.7, §15.5.2](https://openid.net/specs/openid-connect-core-1_0.html#NonceNotes) |
| Opaque session token, hashed at rest | JWT session; Google's ID token as session | Instant revocation; a DB leak doesn't leak sessions | [OWASP Session Mgmt](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) |
| SHA-256 for token hashes | bcrypt/argon2 | 256-bit random tokens can't be brute-forced; slow hashes are for passwords | — |
| 90-day sliding expiry, touched at most daily | Short sessions + refresh tokens; absolute cap | Mobile users expect to stay signed in; at most one write per day | [OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) (deviation noted under Sessions) |
| `Authorization: Bearer` header | Cookie; query parameter | Standard for native clients; never in URLs | [RFC 6750](https://www.rfc-editor.org/rfc/rfc6750) §2.1, §5.3 |
| Identify by `sub`; no email | Email as key | `sub` is stable and never reused; less personal data held | [Google](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token) |

## Known gaps

- **No rate limiting.** In particular, `POST /auth/nonce` is unauthenticated, so a flood could grow `auth_nonces` for up to 10 minutes. This should be handled as a general API concern, at the edge or in middleware.
- **No absolute session lifetime.** A session in continuous use never expires. Adding a cap later is a one-line change to the session lookup (`created_at > now() - <cap>`).
- **No "sign out everywhere"** and no session listing. `sessions.id` exists so this can be added without exposing tokens.
- **Google key-fetch failures return 401**, the same as a bad token. They're logged as warnings.
- **Sign in with Apple** would be needed for a public App Store release. See `plan/01-data-model.md`.
- **Account deletion** isn't implemented.

## Sources

- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html): ID token validation (§3.1.3.7), nonce implementation notes (§15.5.2)
- [RFC 8252](https://www.rfc-editor.org/rfc/rfc8252): OAuth 2.0 for Native Apps
- [RFC 6750](https://www.rfc-editor.org/rfc/rfc6750): Bearer token usage
- Google: [Verify the Google ID token on your server side](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token)
- Android: [Sign in with Google via Credential Manager](https://developer.android.com/identity/sign-in/credential-manager-siwg-implementation)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
