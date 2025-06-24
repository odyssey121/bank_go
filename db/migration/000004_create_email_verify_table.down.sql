DROP TABLE IF EXISTS "email_verify";

ALTER TABLE "users"
DROP COLUMN IF EXISTS is_email_verify;