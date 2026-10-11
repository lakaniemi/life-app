-- name: CreateInvite :one
INSERT INTO family_invites (token_hash, family_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteExpiredFamilyInvites :exec
DELETE FROM family_invites
WHERE family_id = $1 AND expires_at <= now();

-- name: PreviewInvite :one
-- Returns the family of an unexpired invite, and whether the user is already
-- a member of it.
SELECT f.id, f.name,
       EXISTS (
           SELECT 1 FROM family_members fm
           WHERE fm.family_id = f.id AND fm.user_id = $2
       ) AS already_member
FROM families f
JOIN family_invites i ON i.family_id = f.id
WHERE i.token_hash = $1 AND i.expires_at > now();

-- name: GetFamilyByInviteToken :one
-- Returns the family an unexpired invite belongs to.
SELECT f.* FROM families f
JOIN family_invites i ON i.family_id = f.id
WHERE i.token_hash = $1 AND i.expires_at > now();
