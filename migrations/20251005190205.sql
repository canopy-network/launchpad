-- Modify "chain_keys" table
ALTER TABLE "chain_keys" DROP CONSTRAINT "chain_keys_address_check", ALTER COLUMN "address" TYPE text;
