-- path: backend/sql/roles.sql
-- 🆕 NEW - Role management queries

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name ASC;

-- name: ListSystemRoles :many
SELECT * FROM roles
WHERE is_system = TRUE
ORDER BY name ASC;

-- name: CreateRole :one
INSERT INTO roles (
    name,
    description,
    permissions
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET 
    description = COALESCE(sqlc.narg('description'), description),
    permissions = COALESCE(sqlc.narg('permissions'), permissions),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND is_system = FALSE
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1 AND is_system = FALSE;