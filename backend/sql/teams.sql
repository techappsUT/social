-- path: backend/sql/teams.sql
-- Team Management SQLC Queries (Fixed query names)

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

-- name: AddTeamMember :one
INSERT INTO team_memberships (
    team_id,
    user_id,
    role_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: RemoveTeamMember :exec
UPDATE team_memberships
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE team_id = $1 
  AND user_id = $2 
  AND deleted_at IS NULL;

-- name: UpdateTeamMemberRole :one
UPDATE team_memberships
SET 
    role_id = $2,
    updated_at = NOW()
WHERE team_id = $1 
  AND user_id = $3 
  AND deleted_at IS NULL
RETURNING *;

-- name: GetTeamMembers :many
SELECT 
    tm.id,
    tm.team_id,
    tm.user_id,
    tm.role_id,
    tm.created_at,
    tm.updated_at,
    u.email,
    u.username,
    u.first_name,
    u.last_name,
    u.avatar_url,
    r.name as role_name,
    r.permissions
FROM team_memberships tm
INNER JOIN users u ON tm.user_id = u.id
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND tm.deleted_at IS NULL
  AND u.deleted_at IS NULL
ORDER BY tm.created_at ASC;

-- name: GetTeamMemberByUserID :one
SELECT 
    tm.id,
    tm.team_id,
    tm.user_id,
    tm.role_id,
    tm.created_at,
    tm.updated_at,
    r.name as role_name,
    r.permissions
FROM team_memberships tm
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND tm.user_id = $2 
  AND tm.deleted_at IS NULL;

-- name: CountTeamsByUserID :one
SELECT COUNT(*) FROM team_memberships tm
INNER JOIN teams t ON tm.team_id = t.id
WHERE tm.user_id = $1 
  AND tm.deleted_at IS NULL
  AND t.deleted_at IS NULL;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE name = $1;

-- name: GetTeamOwner :one
SELECT 
    tm.user_id,
    u.email,
    u.username
FROM team_memberships tm
INNER JOIN users u ON tm.user_id = u.id
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND r.name = 'owner'
  AND tm.deleted_at IS NULL
LIMIT 1;

-- name: CountAdminsInTeam :one
SELECT COUNT(*) FROM team_memberships tm
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND r.name IN ('owner', 'admin')
  AND tm.deleted_at IS NULL;

-- name: ExistsTeamMember :one
SELECT EXISTS(
    SELECT 1 FROM team_memberships
    WHERE team_id = $1 
      AND user_id = $2 
      AND deleted_at IS NULL
);