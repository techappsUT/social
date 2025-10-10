-- path: backend/sql/plans.sql
-- 🔄 REFACTORED - Use price_monthly/yearly, stripe_price_id_monthly/yearly

-- name: CreatePlan :one
INSERT INTO plans (
    name,
    slug,
    description,
    price_monthly,
    price_yearly,
    features,
    limits,
    is_active,
    stripe_price_id_monthly,
    stripe_price_id_yearly
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetPlanByID :one
SELECT * FROM plans WHERE id = $1;

-- name: GetPlanBySlug :one
SELECT * FROM plans WHERE slug = $1;

-- name: GetPlanByStripeMonthlyID :one
SELECT * FROM plans WHERE stripe_price_id_monthly = $1;

-- name: GetPlanByStripeYearlyID :one
SELECT * FROM plans WHERE stripe_price_id_yearly = $1;

-- name: ListActivePlans :many
SELECT * FROM plans
WHERE is_active = TRUE
ORDER BY price_monthly ASC;

-- name: ListAllPlans :many
SELECT * FROM plans
ORDER BY price_monthly ASC;

-- name: UpdatePlan :one
UPDATE plans
SET 
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    price_monthly = COALESCE(sqlc.narg('price_monthly'), price_monthly),
    price_yearly = COALESCE(sqlc.narg('price_yearly'), price_yearly),
    features = COALESCE(sqlc.narg('features'), features),
    limits = COALESCE(sqlc.narg('limits'), limits),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeactivatePlan :exec
UPDATE plans
SET 
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1;