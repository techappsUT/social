-- path: backend/sql/invoices.sql

-- name: CreateInvoice :one
INSERT INTO invoices (
    subscription_id,
    team_id,
    stripe_invoice_id,
    invoice_number,
    status,
    subtotal,
    tax,
    total,
    amount_due,
    currency,
    invoice_date,
    due_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetInvoiceByID :one
SELECT * FROM invoices WHERE id = $1;

-- name: GetInvoiceByStripeID :one
SELECT * FROM invoices WHERE stripe_invoice_id = $1;

-- name: ListInvoicesByTeam :many
SELECT * FROM invoices
WHERE team_id = $1
ORDER BY invoice_date DESC
LIMIT $2 OFFSET $3;

-- name: UpdateInvoiceStatus :exec
UPDATE invoices
SET 
    status = $2,
    amount_paid = COALESCE(sqlc.narg('amount_paid'), amount_paid),
    paid_at = COALESCE(sqlc.narg('paid_at'), paid_at),
    invoice_pdf_url = COALESCE(sqlc.narg('invoice_pdf_url'), invoice_pdf_url),
    hosted_invoice_url = COALESCE(sqlc.narg('hosted_invoice_url'), hosted_invoice_url),
    updated_at = NOW()
WHERE id = $1;