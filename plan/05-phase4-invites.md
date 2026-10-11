# Phase 4 — Invites (PR 4)

Joining a family uses an invite. An admin taps "Invite a member", and the app shows the invite as a QR code and a shareable link (phase 6). When that screen closes, the invite expires in the background, used or not. There's no interface for managing invites, so there's nothing to list or revoke.

## Scope

- **Migration:** add the `family_invites` table.
- **The token:**
  - 32 random bytes from `crypto/rand`, base64url-encoded, like session tokens. Nobody types it: QR and link cover every case, and the app's fallback is pasting the link.
  - Stored only as a SHA-256 hash. The API returns the token once, from the create call, so a read-only database leak can't be used to join families.
  - Multi-use until it expires after 24h: one QR on screen, the whole family scans it. 24h, not minutes, because a link sent over chat may be opened hours later.
- **QR content:** a deep link, e.g. `lifeapp://join?token=…`. The API only deals in tokens; the deep-link format is the app's concern.

## Endpoints

| Method & path                       | Who      | Notes                                                                               |
| ----------------------------------- | -------- | ----------------------------------------------------------------------------------- |
| `POST /families/{familyID}/invites` | admin    | 201 → `{invite: {token, expiresAt}}`                                                |
| `GET /invites/{token}`              | any user | Preview: `{family: {id, name}, alreadyMember}`, so the user confirms before joining |
| `POST /invites/{token}/accept`      | any user | Joins as `member` → `{family}` with the caller's role                               |

An unknown or expired token returns 404 for both cases, so there's no oracle for guessing tokens.

**Already a member.** Accepting succeeds without changing anything: the requested state already holds, a retry after a lost response must succeed, and an admin scanning their own QR must not be demoted. The preview's `alreadyMember` lets the app say "You're already in X" instead of offering to join.

**API vs UI.** The API enforces authentication, admin-only creation, 404 for unknown and expired tokens alike, and leaves an existing member's role untouched. The UI decides how to present those cases, and what happens when the user already belongs to a *different* family. The API allows that, since a user can be in several families (see phase 6).

## Done when

- Tests cover create (admin only, 403 for members), preview, accept, multi-use, expiry and already-a-member.
- CI is green.
