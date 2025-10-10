-- path: backend/sql/invoices.sql
-- 🔄 REFACTORED - Use due_date, paid_at (no invoice_date, no invoice_pdf_url)

-- name: CreateInvoice :one
INSERT INTO invoices (
    subscription_id,
    stripe_invoice_id,
    amount_due,
    amount_paid,
    currency,
    status,
    due_date,
    paid_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetInvoiceByID :one
SELECT * FROM invoices WHERE id = $1;

-- name: GetInvoiceByStripeID :one
SELECT * FROM invoices WHERE stripe_invoice_id = $1;

-- name: ListInvoicesBySubscription :many
SELECT * FROM invoices
WHERE subscription_id = $1
ORDER BY created_at DESC;

-- name: ListInvoicesByTeam :many
SELECT i.*
FROM invoices i
INNER JOIN subscriptions s ON i.subscription_id = s.id
WHERE s.team_id = $1
ORDER BY i.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateInvoiceStatus :exec
UPDATE invoices
SET 
    status = $2,
    amount_paid = COALESCE(sqlc.narg('amount_paid'), amount_paid),
    paid_at = COALESCE(sqlc.narg('paid_at'), paid_at),
    updated_at = NOW()
WHERE id = $1;

-- name: CountUnpaidInvoices :one
SELECT COUNT(*)
FROM invoices i
INNER JOIN subscriptions s ON i.subscription_id = s.id
WHERE s.team_id = $1
  AND i.status IN ('open', 'past_due');