-- path: backend/sql/teams.sql
-- 🔄 REFACTORED - Removed duplicate team membership queries

-- name: CreateTeam :one
INSERT INTO teams (
    name,
    slug,
    avatar_url,
    settings,
    created_by
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTeamBySlug :one
SELECT * FROM teams
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListTeamsByUser :many
SELECT t.*
FROM teams t
INNER JOIN team_memberships tm ON t.id = tm.team_id
WHERE tm.user_id = $1 
  AND tm.deleted_at IS NULL
  AND t.deleted_at IS NULL
ORDER BY t.created_at DESC;

-- name: UpdateTeam :one
UPDATE teams
SET 
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTeam :exec
UPDATE teams
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CountTeamMembers :one
SELECT COUNT(*)
FROM team_memberships
WHERE team_id = $1 
  AND deleted_at IS NULL;