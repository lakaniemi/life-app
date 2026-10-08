-- name: CreateNonce :exec
INSERT INTO auth_nonces (nonce, expires_at)
VALUES ($1, $2);

-- name: ConsumeNonce :one
-- Checks and consumes the nonce in one statement. Of two concurrent logins
-- with the same nonce, only one gets the row back.
DELETE FROM auth_nonces
WHERE nonce = $1 AND expires_at > now()
RETURNING nonce;

-- name: DeleteExpiredNonces :exec
DELETE FROM auth_nonces
WHERE expires_at <= now();
