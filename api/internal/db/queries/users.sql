-- name: CreateUser :one
INSERT INTO users (google_sub, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByGoogleSub :one
SELECT * FROM users
WHERE google_sub = $1;
