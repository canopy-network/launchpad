-- Modify "chains" table
ALTER TABLE "chains" ADD CONSTRAINT "chains_address_check" CHECK (octet_length(address) = 20), ADD COLUMN "address" bytea NULL;
