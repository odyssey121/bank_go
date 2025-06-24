CREATE TABLE "email_verify" (
    "id" bigserial PRIMARY KEY,
    "username" varchar NOT NULL,
    "email" varchar NOT NULL,
    "code" varchar NOT NULL,
    "is_verified" boolean NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '1 hour')
);

ALTER TABLE "users"
ADD COLUMN IF NOT EXISTS is_email_verify boolean DEFAULT false;

ALTER TABLE "email_verify" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");