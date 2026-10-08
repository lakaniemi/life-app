-- name: CreateUser :one
INSERT INTO users (google_sub, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByGoogleSub :one
SELECT * FROM users
WHERE google_sub = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserName :one
UPDATE users
SET name = $2, updated_at = now()
WHERE id = $1
RETURNING *;
