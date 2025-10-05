-- path: backend/sql/plans.sql

-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1 AND is_active = true;

-- name: GetPlanBySlug :one
SELECT * FROM plans WHERE slug = $1 AND is_active = true;

-- name: ListActivePlans :many
SELECT * FROM plans 
WHERE is_active = true 
ORDER BY display_order ASC, price_cents ASC;