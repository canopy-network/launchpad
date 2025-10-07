-- Modify "chain_templates" table
ALTER TABLE "chain_templates" DROP CONSTRAINT "chain_templates_complexity_level_check", DROP COLUMN "default_consensus", DROP COLUMN "default_validator_count", DROP COLUMN "complexity_level", DROP COLUMN "estimated_deployment_time_minutes";
