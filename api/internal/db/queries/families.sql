-- name: CreateFamily :one
INSERT INTO families (name)
VALUES ($1)
RETURNING *;

-- name: GetFamily :one
SELECT * FROM families
WHERE id = $1;

-- name: UpdateFamilyName :one
UPDATE families
SET name = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteFamily :exec
DELETE FROM families
WHERE id = $1;

-- name: LockFamily :one
-- Locks the family row until the transaction ends, serializing changes to the
-- family's membership. See the internal/family package doc for why.
SELECT id FROM families
WHERE id = $1
FOR UPDATE;

-- name: ListFamiliesForUser :many
SELECT f.id, f.name, fm.role
FROM families f
JOIN family_members fm ON fm.family_id = f.id
WHERE fm.user_id = $1
ORDER BY f.name, f.id;

-- name: AddFamilyMember :exec
INSERT INTO family_members (family_id, user_id, role)
VALUES ($1, $2, $3);

-- name: AddFamilyMemberIfAbsent :exec
-- Leaves an existing membership, and its role, as it is.
INSERT INTO family_members (family_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (family_id, user_id) DO NOTHING;

-- name: GetFamilyMember :one
SELECT * FROM family_members
WHERE family_id = $1 AND user_id = $2;

-- name: ListFamilyMembers :many
SELECT u.id AS user_id, u.name, fm.role, fm.joined_at
FROM family_members fm
JOIN users u ON u.id = fm.user_id
WHERE fm.family_id = $1
ORDER BY fm.joined_at, u.id;

-- name: UpdateFamilyMemberRole :exec
UPDATE family_members
SET role = $3
WHERE family_id = $1 AND user_id = $2;

-- name: DeleteFamilyMember :exec
DELETE FROM family_members
WHERE family_id = $1 AND user_id = $2;
