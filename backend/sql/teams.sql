-- path: backend/sql/teams.sql

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

-- name: UpdateTeam :one
UPDATE teams
SET 
    name = COALESCE(sqlc.narg('name'), name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: ListTeamsByUser :many
SELECT t.* FROM teams t
INNER JOIN team_memberships tm ON t.id = tm.team_id
WHERE tm.user_id = $1 
  AND t.deleted_at IS NULL 
  AND tm.deleted_at IS NULL
  AND tm.is_active = true
ORDER BY t.created_at DESC;

-- name: SoftDeleteTeam :exec
UPDATE teams
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;