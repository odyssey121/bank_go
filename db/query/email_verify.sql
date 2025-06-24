-- name: CreateEmailVerify :one
INSERT INTO email_verify (
  username,
  email,
  code
) VALUES (
  $1, $2, $3
) RETURNING *;


-- name: UpdateEmailVerify :one
UPDATE email_verify SET 
  is_verified = COALESCE(sqlc.narg(is_verified), is_verified)
WHERE id = $1
RETURNING *;

-- name: GetEmailVerify :one
SELECT * FROM email_verify WHERE id = $1 LIMIT 1;