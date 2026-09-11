-- name: AddCard :one
INSERT INTO cards (owner_id, token, type, last4)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetCards :many
SELECT id, owner_id, token, type, linkedAt, last4
FROM cards
WHERE owner_id = $1;

-- name: GetCardByID :one
SELECT id, owner_id, token, type, linkedAt, last4
FROM cards
WHERE id = $1;

-- name: RemoveCard :exec
DELETE FROM cards
WHERE id = $1;

-- name: AddPayment :exec
INSERT INTO payments (owner_id, card_id, amount, currency, status)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET status = $2
WHERE id = $1;

-- name: GetPayments :many
SELECT id, card_id, amount, currency, status, created_at
FROM payments
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
