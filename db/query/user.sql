-- name: CreateUser :one
INSERT INTO users (
  username,
  full_name,
  email,
  hashed_password
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET 
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  full_name = COALESCE(sqlc.narg(full_name), full_name),
  email = COALESCE(sqlc.narg(email), email),
  password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
  is_email_verify = COALESCE(sqlc.narg(is_email_verify), is_email_verify)
WHERE username = sqlc.arg(username)
RETURNING *;

