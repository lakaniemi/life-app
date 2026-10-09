-- +goose Up
CREATE TABLE sessions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users ON DELETE CASCADE,
    -- SHA-256 of the session token (32 bytes); the token itself is never stored.
    token_hash    bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at    timestamptz NOT NULL DEFAULT now(),
    expires_at    timestamptz NOT NULL,
    last_used_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- Login nonces issued by POST /auth/nonce. Each is single-use: login deletes
-- it. See docs/AUTH.md.
CREATE TABLE auth_nonces (
    nonce       text PRIMARY KEY,
    expires_at  timestamptz NOT NULL
);
CREATE INDEX auth_nonces_expires_at_idx ON auth_nonces (expires_at);

-- +goose Down
DROP TABLE auth_nonces;
DROP TABLE sessions;
