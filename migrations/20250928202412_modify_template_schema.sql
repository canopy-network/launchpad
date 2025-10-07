-- Modify "chain_templates" table
ALTER TABLE "chain_templates" DROP COLUMN "default_modules", DROP COLUMN "required_modules";
-- Modify "chains" table
ALTER TABLE "chains" ALTER COLUMN "template_id" DROP NOT NULL;
