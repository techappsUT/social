-- path: backend/sql/subscriptions.sql

-- name: CreateSubscription :one
INSERT INTO subscriptions (
    team_id,
    plan_id,
    status,
    stripe_subscription_id,
    stripe_customer_id,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetSubscriptionByTeamID :one
SELECT s.*, p.name as plan_name, p.features as plan_features
FROM subscriptions s
INNER JOIN plans p ON s.plan_id = p.id
WHERE s.team_id = $1;

-- name: GetSubscriptionByStripeID :one
SELECT * FROM subscriptions
WHERE stripe_subscription_id = $1;

-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions
SET 
    status = $2,
    current_period_start = COALESCE(sqlc.narg('current_period_start'), current_period_start),
    current_period_end = COALESCE(sqlc.narg('current_period_end'), current_period_end),
    updated_at = NOW()
WHERE id = $1;

-- name: CancelSubscription :exec
UPDATE subscriptions
SET 
    cancel_at_period_end = true,
    canceled_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListExpiringSubscriptions :many
SELECT s.*, t.name as team_name
FROM subscriptions s
INNER JOIN teams t ON s.team_id = t.id
WHERE s.status = 'active'
  AND s.current_period_end BETWEEN $1 AND $2
  AND s.cancel_at_period_end = false
ORDER BY s.current_period_end ASC;