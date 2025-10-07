-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "wallet_address" character varying(42) NOT NULL,
  "email" character varying(320) NULL,
  "username" character varying(50) NULL,
  "display_name" character varying(100) NULL,
  "bio" text NULL,
  "avatar_url" text NULL,
  "website_url" text NULL,
  "twitter_handle" character varying(50) NULL,
  "github_username" character varying(100) NULL,
  "telegram_handle" character varying(50) NULL,
  "notification_preferences" jsonb NULL DEFAULT '{"email": false, "browser": true}',
  "is_verified" boolean NOT NULL DEFAULT false,
  "verification_tier" character varying(20) NULL DEFAULT 'basic',
  "total_chains_created" integer NOT NULL DEFAULT 0,
  "total_cnpy_invested" numeric(15,8) NOT NULL DEFAULT 0,
  "reputation_score" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "last_active_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_username_key" UNIQUE ("username"),
  CONSTRAINT "users_wallet_address_key" UNIQUE ("wallet_address"),
  CONSTRAINT "users_verification_tier_check" CHECK ((verification_tier)::text = ANY ((ARRAY['basic'::character varying, 'verified'::character varying, 'premium'::character varying])::text[]))
);
-- Create index "idx_users_github" to table: "users"
CREATE INDEX "idx_users_github" ON "users" ("github_username");
-- Create index "idx_users_reputation" to table: "users"
CREATE INDEX "idx_users_reputation" ON "users" ("reputation_score" DESC);
-- Create index "idx_users_wallet" to table: "users"
CREATE INDEX "idx_users_wallet" ON "users" ("wallet_address");
-- Create "chain_templates" table
CREATE TABLE "chain_templates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "template_name" character varying(100) NOT NULL,
  "template_description" text NOT NULL,
  "template_category" character varying(50) NOT NULL,
  "supported_language" text NOT NULL,
  "default_modules" text[] NOT NULL,
  "required_modules" text[] NOT NULL,
  "default_consensus" character varying(50) NOT NULL DEFAULT 'tendermint',
  "default_token_supply" bigint NOT NULL DEFAULT 1000000000,
  "default_validator_count" integer NOT NULL DEFAULT 10,
  "complexity_level" character varying(20) NOT NULL,
  "estimated_deployment_time_minutes" integer NOT NULL DEFAULT 30,
  "documentation_url" text NULL,
  "example_chains" text[] NULL,
  "version" character varying(20) NOT NULL DEFAULT '1.0.0',
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chain_templates_template_name_key" UNIQUE ("template_name"),
  CONSTRAINT "chain_templates_complexity_level_check" CHECK ((complexity_level)::text = ANY ((ARRAY['beginner'::character varying, 'intermediate'::character varying, 'advanced'::character varying])::text[]))
);
-- Create index "idx_templates_active" to table: "chain_templates"
CREATE INDEX "idx_templates_active" ON "chain_templates" ("is_active");
-- Create index "idx_templates_category" to table: "chain_templates"
CREATE INDEX "idx_templates_category" ON "chain_templates" ("template_category");
-- Create index "idx_templates_complexity" to table: "chain_templates"
CREATE INDEX "idx_templates_complexity" ON "chain_templates" ("complexity_level");
-- Create "chains" table
CREATE TABLE "chains" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_name" character varying(100) NOT NULL,
  "token_symbol" character varying(20) NOT NULL,
  "chain_description" text NULL,
  "template_id" uuid NOT NULL,
  "consensus_mechanism" character varying(50) NOT NULL DEFAULT 'nestbft',
  "token_total_supply" bigint NOT NULL DEFAULT 1000000000,
  "graduation_threshold" numeric(15,2) NOT NULL DEFAULT 50000.00,
  "creation_fee_cnpy" numeric(15,8) NOT NULL DEFAULT 100.00000000,
  "initial_cnpy_reserve" numeric(15,8) NOT NULL DEFAULT 10000.00000000,
  "initial_token_supply" bigint NOT NULL DEFAULT 800000000,
  "bonding_curve_slope" numeric(15,8) NOT NULL DEFAULT 0.00000001,
  "scheduled_launch_time" timestamptz NULL,
  "actual_launch_time" timestamptz NULL,
  "creator_initial_purchase_cnpy" numeric(15,8) NULL DEFAULT 0,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "is_graduated" boolean NOT NULL DEFAULT false,
  "graduation_time" timestamptz NULL,
  "chain_id" character varying(50) NULL,
  "genesis_hash" character varying(64) NULL,
  "validator_min_stake" numeric(15,8) NULL DEFAULT 1000.00000000,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chains_chain_id_key" UNIQUE ("chain_id"),
  CONSTRAINT "chains_chain_name_key" UNIQUE ("chain_name"),
  CONSTRAINT "chains_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chains_template_id_fkey" FOREIGN KEY ("template_id") REFERENCES "chain_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chains_status_check" CHECK ((status)::text = ANY ((ARRAY['draft'::character varying, 'pending_launch'::character varying, 'virtual_active'::character varying, 'graduated'::character varying, 'failed'::character varying])::text[]))
);
-- Create index "idx_chains_creator" to table: "chains"
CREATE INDEX "idx_chains_creator" ON "chains" ("created_by");
-- Create index "idx_chains_graduation" to table: "chains"
CREATE INDEX "idx_chains_graduation" ON "chains" ("is_graduated", "graduation_time");
-- Create index "idx_chains_launch_time" to table: "chains"
CREATE INDEX "idx_chains_launch_time" ON "chains" ("scheduled_launch_time");
-- Create index "idx_chains_status" to table: "chains"
CREATE INDEX "idx_chains_status" ON "chains" ("status");
-- Create index "idx_chains_template" to table: "chains"
CREATE INDEX "idx_chains_template" ON "chains" ("template_id");
-- Create "chain_assets" table
CREATE TABLE "chain_assets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_id" uuid NOT NULL,
  "asset_type" character varying(20) NOT NULL,
  "file_name" character varying(255) NOT NULL,
  "file_url" text NOT NULL,
  "file_size_bytes" bigint NULL,
  "mime_type" character varying(100) NULL,
  "title" character varying(200) NULL,
  "description" text NULL,
  "alt_text" character varying(500) NULL,
  "display_order" integer NOT NULL DEFAULT 0,
  "is_primary" boolean NOT NULL DEFAULT false,
  "is_featured" boolean NOT NULL DEFAULT false,
  "is_active" boolean NOT NULL DEFAULT true,
  "moderation_status" character varying(20) NULL DEFAULT 'pending',
  "moderation_notes" text NULL,
  "uploaded_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chain_assets_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chain_assets_uploaded_by_fkey" FOREIGN KEY ("uploaded_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chain_assets_asset_type_check" CHECK ((asset_type)::text = ANY ((ARRAY['logo'::character varying, 'banner'::character varying, 'screenshot'::character varying, 'video'::character varying, 'whitepaper'::character varying, 'documentation'::character varying])::text[])),
  CONSTRAINT "chain_assets_moderation_status_check" CHECK ((moderation_status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[]))
);
-- Create index "idx_assets_chain" to table: "chain_assets"
CREATE INDEX "idx_assets_chain" ON "chain_assets" ("chain_id");
-- Create index "idx_assets_moderation" to table: "chain_assets"
CREATE INDEX "idx_assets_moderation" ON "chain_assets" ("moderation_status");
-- Create index "idx_assets_primary" to table: "chain_assets"
CREATE INDEX "idx_assets_primary" ON "chain_assets" ("is_primary");
-- Create index "idx_assets_type" to table: "chain_assets"
CREATE INDEX "idx_assets_type" ON "chain_assets" ("asset_type");
-- Create "chain_repositories" table
CREATE TABLE "chain_repositories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_id" uuid NOT NULL,
  "github_url" text NOT NULL,
  "repository_name" character varying(200) NOT NULL,
  "repository_owner" character varying(100) NOT NULL,
  "default_branch" character varying(100) NOT NULL DEFAULT 'main',
  "is_connected" boolean NOT NULL DEFAULT false,
  "oauth_token_hash" character varying(255) NULL,
  "webhook_secret" character varying(100) NULL,
  "auto_upgrade_enabled" boolean NOT NULL DEFAULT true,
  "upgrade_trigger" character varying(50) NULL DEFAULT 'tag_release',
  "last_sync_commit_hash" character varying(40) NULL,
  "last_sync_time" timestamptz NULL,
  "build_status" character varying(20) NULL DEFAULT 'pending',
  "last_build_time" timestamptz NULL,
  "build_logs" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chain_repositories_chain_id_key" UNIQUE ("chain_id"),
  CONSTRAINT "chain_repositories_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chain_repositories_build_status_check" CHECK ((build_status)::text = ANY ((ARRAY['pending'::character varying, 'building'::character varying, 'success'::character varying, 'failed'::character varying])::text[])),
  CONSTRAINT "chain_repositories_upgrade_trigger_check" CHECK ((upgrade_trigger)::text = ANY ((ARRAY['tag_release'::character varying, 'main_push'::character varying, 'manual'::character varying])::text[]))
);
-- Create index "idx_repos_build_status" to table: "chain_repositories"
CREATE INDEX "idx_repos_build_status" ON "chain_repositories" ("build_status");
-- Create index "idx_repos_chain" to table: "chain_repositories"
CREATE INDEX "idx_repos_chain" ON "chain_repositories" ("chain_id");
-- Create "chain_social_links" table
CREATE TABLE "chain_social_links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_id" uuid NOT NULL,
  "platform" character varying(50) NOT NULL,
  "url" text NOT NULL,
  "display_name" character varying(200) NULL,
  "is_verified" boolean NOT NULL DEFAULT false,
  "follower_count" integer NULL DEFAULT 0,
  "last_metrics_update" timestamptz NULL,
  "display_order" integer NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "chain_social_links_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_social_chain" to table: "chain_social_links"
CREATE INDEX "idx_social_chain" ON "chain_social_links" ("chain_id");
-- Create index "idx_social_platform" to table: "chain_social_links"
CREATE INDEX "idx_social_platform" ON "chain_social_links" ("platform");
-- Create index "idx_social_verified" to table: "chain_social_links"
CREATE INDEX "idx_social_verified" ON "chain_social_links" ("is_verified");
-- Create "virtual_pools" table
CREATE TABLE "virtual_pools" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chain_id" uuid NOT NULL,
  "cnpy_reserve" numeric(15,8) NOT NULL DEFAULT 0,
  "token_reserve" bigint NOT NULL DEFAULT 0,
  "current_price_cnpy" numeric(15,8) NOT NULL DEFAULT 0,
  "market_cap_usd" numeric(15,2) NOT NULL DEFAULT 0,
  "total_volume_cnpy" numeric(15,8) NOT NULL DEFAULT 0,
  "total_transactions" integer NOT NULL DEFAULT 0,
  "unique_traders" integer NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "price_24h_change_percent" numeric(8,4) NULL DEFAULT 0,
  "volume_24h_cnpy" numeric(15,8) NULL DEFAULT 0,
  "high_24h_cnpy" numeric(15,8) NULL DEFAULT 0,
  "low_24h_cnpy" numeric(15,8) NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "virtual_pools_chain_id_key" UNIQUE ("chain_id"),
  CONSTRAINT "virtual_pools_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_virtual_pools_active" to table: "virtual_pools"
CREATE INDEX "idx_virtual_pools_active" ON "virtual_pools" ("is_active");
-- Create index "idx_virtual_pools_market_cap" to table: "virtual_pools"
CREATE INDEX "idx_virtual_pools_market_cap" ON "virtual_pools" ("market_cap_usd" DESC);
-- Create "user_virtual_positions" table
CREATE TABLE "user_virtual_positions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "chain_id" uuid NOT NULL,
  "virtual_pool_id" uuid NOT NULL,
  "token_balance" bigint NOT NULL DEFAULT 0,
  "total_cnpy_invested" numeric(15,8) NOT NULL DEFAULT 0,
  "total_cnpy_withdrawn" numeric(15,8) NOT NULL DEFAULT 0,
  "average_entry_price_cnpy" numeric(15,8) NOT NULL DEFAULT 0,
  "unrealized_pnl_cnpy" numeric(15,8) NULL DEFAULT 0,
  "realized_pnl_cnpy" numeric(15,8) NULL DEFAULT 0,
  "total_return_percent" numeric(8,4) NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "first_purchase_at" timestamptz NULL,
  "last_activity_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_virtual_positions_user_id_chain_id_key" UNIQUE ("user_id", "chain_id"),
  CONSTRAINT "user_virtual_positions_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "user_virtual_positions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "user_virtual_positions_virtual_pool_id_fkey" FOREIGN KEY ("virtual_pool_id") REFERENCES "virtual_pools" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_positions_active" to table: "user_virtual_positions"
CREATE INDEX "idx_positions_active" ON "user_virtual_positions" ("is_active");
-- Create index "idx_positions_chain" to table: "user_virtual_positions"
CREATE INDEX "idx_positions_chain" ON "user_virtual_positions" ("chain_id");
-- Create index "idx_positions_pnl" to table: "user_virtual_positions"
CREATE INDEX "idx_positions_pnl" ON "user_virtual_positions" ("unrealized_pnl_cnpy" DESC);
-- Create index "idx_positions_user" to table: "user_virtual_positions"
CREATE INDEX "idx_positions_user" ON "user_virtual_positions" ("user_id");
-- Create "virtual_pool_transactions" table
CREATE TABLE "virtual_pool_transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "virtual_pool_id" uuid NOT NULL,
  "chain_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "transaction_type" character varying(10) NOT NULL,
  "cnpy_amount" numeric(15,8) NOT NULL,
  "token_amount" bigint NOT NULL,
  "price_per_token_cnpy" numeric(15,8) NOT NULL,
  "trading_fee_cnpy" numeric(15,8) NOT NULL DEFAULT 0,
  "slippage_percent" numeric(8,4) NULL DEFAULT 0,
  "transaction_hash" character varying(66) NULL,
  "block_height" bigint NULL,
  "gas_used" integer NULL,
  "pool_cnpy_reserve_after" numeric(15,8) NOT NULL,
  "pool_token_reserve_after" bigint NOT NULL,
  "market_cap_after_usd" numeric(15,2) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "virtual_pool_transactions_chain_id_fkey" FOREIGN KEY ("chain_id") REFERENCES "chains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "virtual_pool_transactions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "virtual_pool_transactions_virtual_pool_id_fkey" FOREIGN KEY ("virtual_pool_id") REFERENCES "virtual_pools" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "virtual_pool_transactions_transaction_type_check" CHECK ((transaction_type)::text = ANY ((ARRAY['buy'::character varying, 'sell'::character varying])::text[]))
);
-- Create index "idx_vp_transactions_chain" to table: "virtual_pool_transactions"
CREATE INDEX "idx_vp_transactions_chain" ON "virtual_pool_transactions" ("chain_id");
-- Create index "idx_vp_transactions_pool" to table: "virtual_pool_transactions"
CREATE INDEX "idx_vp_transactions_pool" ON "virtual_pool_transactions" ("virtual_pool_id");
-- Create index "idx_vp_transactions_time" to table: "virtual_pool_transactions"
CREATE INDEX "idx_vp_transactions_time" ON "virtual_pool_transactions" ("created_at" DESC);
-- Create index "idx_vp_transactions_type" to table: "virtual_pool_transactions"
CREATE INDEX "idx_vp_transactions_type" ON "virtual_pool_transactions" ("transaction_type");
-- Create index "idx_vp_transactions_user" to table: "virtual_pool_transactions"
CREATE INDEX "idx_vp_transactions_user" ON "virtual_pool_transactions" ("user_id");
