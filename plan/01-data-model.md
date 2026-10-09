# Data model

The target is PostgreSQL. The production host is undecided, so phase 1 pins a current major version that managed providers offer. Keep the SQL plain so it runs on any supported version.

```
users ──< sessions
  │
  └──< family_members >── families ──< family_invites
                             └──< (future family-owned tables)

auth_nonces   (standalone; see docs/AUTH.md)
```

## Tables

```sql
CREATE TABLE users (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub  text NOT NULL UNIQUE,
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE families (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE family_members (
    family_id   uuid NOT NULL REFERENCES families ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    role        text NOT NULL CHECK (role IN ('admin', 'member')),
    joined_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (family_id, user_id)
);
CREATE INDEX family_members_user_id_idx ON family_members (user_id);

CREATE TABLE sessions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    token_hash    bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at    timestamptz NOT NULL DEFAULT now(),
    expires_at    timestamptz NOT NULL,
    last_used_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE TABLE auth_nonces (
    nonce       text PRIMARY KEY,
    expires_at  timestamptz NOT NULL
);
CREATE INDEX auth_nonces_expires_at_idx ON auth_nonces (expires_at);

CREATE TABLE family_invites (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    family_id   uuid NOT NULL REFERENCES families ON DELETE CASCADE,
    code        text NOT NULL UNIQUE,
    created_by  uuid REFERENCES users ON DELETE SET NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz
);
```

Migrations add each table in the phase that first needs it: users/families/members in phase 1, sessions and nonces in phase 2, invites in phase 4.

## Why it looks like this

**Identity**
- Users are identified by Google's `sub` claim, which is stable and unique per Google account, never by email. There's no separate identities table. If Apple sign-in is ever needed, the migration makes `google_sub` nullable, adds `apple_sub`, and adds `CHECK (google_sub IS NOT NULL OR apple_sub IS NOT NULL)`.
- **No email is stored anywhere**, and the app doesn't request the `email` scope.
- `name` is copied from Google's `name` claim on first login only. After that it's user-editable and never overwritten.

**Sessions**
- Only a SHA-256 of the token is stored, so a database leak doesn't leak working tokens, which matters on a server someone else operates. A fast hash is fine because the token is 32 random bytes with nothing to brute-force. Slow hashes like bcrypt are for guessable passwords.
- `id` is separate from `token_hash` so sessions can be listed or revoked without exposing anything secret.

**Families and membership**
- `families` has no owner column. Control is expressed only through `family_members.role`, so there's one source of truth.
- The composite primary key prevents duplicate memberships and indexes "members of family X". The extra `user_id` index serves "my families", because a composite index only helps queries on its leading column.
- `role` is text + CHECK rather than a Postgres ENUM. Enum values can't easily be removed or renamed, while a CHECK constraint is simply dropped and recreated.

**Membership rules.** These are enforced in Go inside a transaction that locks the family's member rows (`SELECT … FOR UPDATE`), so concurrent requests can't both pass the check:
1. Every family has at least one admin.
2. The **sole admin can't leave or be demoted** while other members remain, so they must promote someone first.
3. The **sole member can't leave**, so they must delete the family. The UI shows "leave" as "delete" in that case.

**Invites**
- Codes are stored in plain text. An admin must be able to re-display the QR, the code is short-lived, and it only grants membership.
- Codes are multi-use until `expires_at` (default 24h) or `revoked_at`.

## Conventions

- `timestamptz` everywhere.
- UUID v4 from `gen_random_uuid()`. It's built in since PG13; `uuidv7()` needs PG18.
- `updated_at` is set explicitly in each UPDATE query. No triggers.
- Hard deletes, with no `deleted_at` columns.
- Authorization lives in Go, not Postgres row-level security.
- **Every family-owned table** gets `family_id uuid NOT NULL REFERENCES families ON DELETE CASCADE`, and every query on it filters by `family_id`. That one rule keeps families' data apart.
