# Phase 4 — Invites (PR 4)

Joining a family uses an invite code. The app shows it as a QR code, and the joining user scans it (phase 6).

## Scope

- **Migration:** add the `family_invites` table.
- **The code:**
  - About 10 characters from an unambiguous alphabet (no `0/O`, `1/I/l`), generated with `crypto/rand`. That's random enough, and still typeable by hand as a fallback.
  - Defaults to a 24h expiry.
  - Multi-use until it expires or is revoked: one QR on screen, the whole family scans it.
- **QR content:** a deep link, e.g. `lifeapp://join?code=ABCD234XYZ`. The API only deals in codes; the deep-link format is the app's concern.

## Endpoints

| Method & path                                   | Who        | Notes                                                       |
| ----------------------------------------------- | ---------- | ----------------------------------------------------------- |
| `POST /families/{familyID}/invites`             | admin      | → `{id, code, expiresAt}`                                   |
| `GET /families/{familyID}/invites`              | admin      | Active invites, so a QR can be shown again                  |
| `DELETE /families/{familyID}/invites/{inviteID}`| admin      | Sets `revoked_at`                                           |
| `GET /invites/{code}`                           | any user   | Preview: the family name, so the user confirms before joining |
| `POST /invites/{code}/accept`                   | any user   | Joins as `member`. Idempotent if already a member.          |

An unknown, expired or revoked code returns 404 for all three cases, so there's no oracle for guessing codes.

## Done when

- Tests cover accept, expiry, revocation, already-a-member and non-admin creating an invite (403).
- CI is green.
