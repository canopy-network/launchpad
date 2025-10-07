-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "email_verified_at" timestamptz NULL, ADD COLUMN "jwt_version" integer NOT NULL DEFAULT 0;
