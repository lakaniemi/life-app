-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByTokenHash :one
-- An expired session is treated exactly like a missing one.
SELECT * FROM sessions
WHERE token_hash = $1 AND expires_at > now();

-- name: TouchSession :exec
UPDATE sessions
SET last_used_at = now(), expires_at = $2
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteExpiredSessionsForUser :exec
DELETE FROM sessions
WHERE user_id = $1 AND expires_at <= now();
