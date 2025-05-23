-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1;
-- name: GetAccountForUpdate :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1 FOR NO KEY UPDATE;

-- name: CreateAccount :one
INSERT INTO accounts (
  owner,
  balance,
  currency
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListAccounts :many
SELECT * FROM accounts ORDER BY id LIMIT $1 OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts set balance = $1 WHERE id = $2 RETURNING *;

-- name: UpdateAccountBalancePlus :one
UPDATE accounts set balance = balance + sqlc.arg(amount) where id = $1 RETURNING *;

-- name: UpdateAccountBalanceMinus :one
UPDATE accounts set balance = balance - sqlc.arg(amount) where id = $1 RETURNING *;

-- name: DeleteAccount :one
DELETE FROM accounts WHERE id = $1 RETURNING *;

