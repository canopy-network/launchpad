-- Modify "chains" table
ALTER TABLE "chains" ADD COLUMN "token_name" character varying(100) NULL, ADD COLUMN "block_time_seconds" integer NULL, ADD COLUMN "upgrade_block_height" bigint NULL, ADD COLUMN "block_reward_amount" numeric(15,8) NULL;
-- Create "wallets" table
CREATE TABLE "wallets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NULL,
  "chain_id" uuid NULL,
  "address" text NOT NULL,
  "public_key" text NOT NULL,
  "encrypted_private_key" text NOT NULL,
  "salt" bytea NOT NULL,
  "wallet_name" character varying(100) NULL,
  "wallet_description" text NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "is_locked" boolean NOT NULL DEFAULT false,
  "last_used_at" timestamptz NULL,
  "password_changed_at" timestamptz NULL,
  "failed_decrypt_attempts" integer NOT NULL DEFAULT 0,
  "locked_until" timestamptz NULL,
  "created_by" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "wallets_address_key" UNIQUE ("address"),
  CONSTRAINT "wallets_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "wallets_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "wallets_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "wallets_check" CHECK ((user_id IS NOT NULL) OR (chain_id IS NOT NULL) OR (created_by IS NOT NULL)),
  CONSTRAINT "wallets_salt_check" CHECK (octet_length(salt) = 16)
);
-- Create index "idx_wallets_active" to table: "wallets"
CREATE INDEX "idx_wallets_active" ON "wallets" ("is_active");
-- Create index "idx_wallets_address" to table: "wallets"
CREATE INDEX "idx_wallets_address" ON "wallets" ("address");
-- Create index "idx_wallets_chain" to table: "wallets"
CREATE INDEX "idx_wallets_chain" ON "wallets" ("chain_id") WHERE (chain_id IS NOT NULL);
-- Create index "idx_wallets_created_by" to table: "wallets"
CREATE INDEX "idx_wallets_created_by" ON "wallets" ("created_by") WHERE (created_by IS NOT NULL);
-- Create index "idx_wallets_user" to table: "wallets"
CREATE INDEX "idx_wallets_user" ON "wallets" ("user_id") WHERE (user_id IS NOT NULL);
