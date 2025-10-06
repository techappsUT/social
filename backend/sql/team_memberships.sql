-- path: backend/sql/team_memberships.sql

-- name: AddMemberToTeam :one
INSERT INTO team_memberships (
    team_id,
    user_id,
    role_id,
    invited_by,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetTeamMembership :one
SELECT tm.*, r.name as role_name, r.permissions as role_permissions
FROM team_memberships tm
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND tm.user_id = $2 
  AND tm.deleted_at IS NULL;

-- name: ListTeamMembers :many
SELECT 
    tm.*,
    u.email,
    u.full_name,
    u.avatar_url,
    r.name as role_name
FROM team_memberships tm
INNER JOIN users u ON tm.user_id = u.id
INNER JOIN roles r ON tm.role_id = r.id
WHERE tm.team_id = $1 
  AND tm.deleted_at IS NULL
  AND u.deleted_at IS NULL
ORDER BY tm.created_at ASC;

-- name: UpdateMemberRole :one
UPDATE team_memberships
SET 
    role_id = $2,
    updated_at = NOW()
WHERE team_id = $1 AND user_id = $3 AND deleted_at IS NULL
RETURNING *;

-- name: CreateInvitation :one
INSERT INTO team_memberships (
    team_id,
    user_id,
    role_id,
    invited_by,
    invitation_token,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, false
)
RETURNING *;

-- name: AcceptInvitation :exec
UPDATE team_memberships
SET 
    invitation_accepted_at = NOW(),
    is_active = true,
    updated_at = NOW()
WHERE invitation_token = $1 AND invitation_accepted_at IS NULL;

-- name: RemoveMemberFromTeam :exec
UPDATE team_memberships
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE team_id = $1 AND user_id = $2;