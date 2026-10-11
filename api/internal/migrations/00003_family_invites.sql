-- +goose Up
-- Invites created by POST /families/{familyID}/invites. Multi-use until they
-- expire; there's no revocation. See the internal/family package doc.
CREATE TABLE family_invites (
    -- SHA-256 of the invite token; the token itself is never stored.
    token_hash  bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    family_id   uuid NOT NULL REFERENCES families ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);
CREATE INDEX family_invites_family_id_idx ON family_invites (family_id);

-- +goose Down
DROP TABLE family_invites;
