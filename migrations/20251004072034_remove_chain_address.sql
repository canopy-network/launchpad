-- Modify "chains" table
ALTER TABLE "chains" DROP CONSTRAINT "chains_address_check", DROP COLUMN "address";
-- Create "chain_keys" table
CREATE TABLE "chain_keys" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_id" uuid NOT NULL,
  "address" bytea NOT NULL,
  "public_key" bytea NOT NULL,
  "encrypted_private_key" text NOT NULL,
  "salt" bytea NOT NULL,
  "key_nickname" character varying(100) NULL,
  "key_purpose" character varying(50) NULL DEFAULT 'chain_operation',
  "is_active" boolean NOT NULL DEFAULT true,
  "last_used_at" timestamptz NULL,
  "rotation_count" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chain_keys_address_key" UNIQUE ("address"),
  CONSTRAINT "chain_keys_chain_id_key_purpose_key" UNIQUE ("chain_id", "key_purpose"),
  CONSTRAINT "chain_keys_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chain_keys_address_check" CHECK (octet_length(address) = 20),
  CONSTRAINT "chain_keys_key_purpose_check" CHECK ((key_purpose)::text = ANY ((ARRAY['chain_operation'::character varying, 'governance'::character varying, 'treasury'::character varying, 'backup'::character varying])::text[])),
  CONSTRAINT "chain_keys_salt_check" CHECK (octet_length(salt) = 16)
);
-- Create index "idx_chain_keys_active" to table: "chain_keys"
CREATE INDEX "idx_chain_keys_active" ON "chain_keys" ("is_active");
-- Create index "idx_chain_keys_address" to table: "chain_keys"
CREATE INDEX "idx_chain_keys_address" ON "chain_keys" ("address");
-- Create index "idx_chain_keys_chain" to table: "chain_keys"
CREATE INDEX "idx_chain_keys_chain" ON "chain_keys" ("chain_id");
-- Create index "idx_chain_keys_purpose" to table: "chain_keys"
CREATE INDEX "idx_chain_keys_purpose" ON "chain_keys" ("key_purpose");
