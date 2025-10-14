-- Canonical database schema for Launchpad
-- This file represents the desired state of the database schema

-- User accounts and authentication data for chain creators and participants
-- Stores basic user information and preferences for the launchpad system
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Authentication
    wallet_address VARCHAR(42) NOT NULL UNIQUE, -- Primary authentication via wallet
    email VARCHAR(320) UNIQUE, -- Optional email for notifications
    username VARCHAR(50) UNIQUE,

    -- Profile information
    display_name VARCHAR(100),
    bio TEXT,
    avatar_url TEXT,
    website_url TEXT,

    -- Social connections
    twitter_handle VARCHAR(50),
    github_username VARCHAR(100),
    telegram_handle VARCHAR(50),

    -- User preferences and status
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_tier VARCHAR(20) DEFAULT 'basic' CHECK (verification_tier IN ('basic', 'verified', 'premium')),

    -- JWT authentication
    email_verified_at TIMESTAMP WITH TIME ZONE, -- Track when email was verified
    jwt_version INTEGER NOT NULL DEFAULT 0, -- Increment to invalidate all JWTs for this user

    -- Activity tracking
    total_chains_created INTEGER NOT NULL DEFAULT 0,
    total_cnpy_invested DECIMAL(15,8) NOT NULL DEFAULT 0,
    reputation_score INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Pre-built blockchain templates that developers can use as starting points
-- Contains the technical specifications and default configurations for different chain types
CREATE TABLE chain_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Template identification
    template_name VARCHAR(100) NOT NULL UNIQUE,
    template_description TEXT NOT NULL,
    template_category VARCHAR(50) NOT NULL, -- 'defi', 'gaming', 'enterprise', 'social', etc.

    -- Technical specifications
    supported_language TEXT NOT NULL, -- 'go'

    -- Default economic parameters
    default_token_supply BIGINT NOT NULL DEFAULT 1000000000,

    -- Template metadata
    documentation_url TEXT,
    example_chains TEXT[], -- Array of successful chain examples

    -- Versioning and maintenance
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Core entity representing a blockchain chain in the Scanopy ecosystem
-- This is the primary table that holds all chain-specific configuration and metadata
-- Links to all other entities in the system through foreign key relationships
CREATE TABLE chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic chain identification
    chain_name VARCHAR(100) NOT NULL UNIQUE,
    token_name VARCHAR(100),
    token_symbol VARCHAR(20) NOT NULL,
    chain_description TEXT,

    -- Template and technical configuration
    template_id UUID REFERENCES chain_templates(id),
    consensus_mechanism VARCHAR(50) NOT NULL DEFAULT 'nestbft',

    -- Economic parameters
    token_total_supply BIGINT NOT NULL DEFAULT 1000000000,
    block_time_seconds INTEGER, -- Block time in seconds: 5, 10, 20, 30, 60, 120, 300, 600, 1800
    upgrade_block_height BIGINT, -- Block height for upgrades
    block_reward_amount DECIMAL(15,8), -- Block reward amount
    graduation_threshold DECIMAL(15,2) NOT NULL DEFAULT 50000.00, -- Amount in CNPY required for graduation
    creation_fee_cnpy DECIMAL(15,8) NOT NULL DEFAULT 100.00000000,

    -- Bonding curve parameters
    initial_cnpy_reserve DECIMAL(15,8) NOT NULL DEFAULT 10000.00000000,
    initial_token_supply BIGINT NOT NULL DEFAULT 800000000,
    bonding_curve_slope DECIMAL(15,8) NOT NULL DEFAULT 0.00000001,

    -- Launch configuration
    scheduled_launch_time TIMESTAMP WITH TIME ZONE,
    actual_launch_time TIMESTAMP WITH TIME ZONE,
    creator_initial_purchase_cnpy DECIMAL(15,8) DEFAULT 0,

    -- Chain status and lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_launch', 'virtual_active', 'graduated', 'failed')),
    is_graduated BOOLEAN NOT NULL DEFAULT FALSE,
    graduation_time TIMESTAMP WITH TIME ZONE,

    -- Network and deployment info
    chain_id VARCHAR(50) UNIQUE, -- Set when chain is deployed
    genesis_hash VARCHAR(64), -- Set when chain is deployed
    validator_min_stake DECIMAL(15,8) DEFAULT 1000.00000000,

    -- Audit trail
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- GitHub repository connections for chain development and auto-upgrade functionality
-- Links chains to their source code repositories for CI/CD and upgrade management
CREATE TABLE chain_repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chain_id UUID NOT NULL REFERENCES chains(id) ON DELETE CASCADE,

    -- Repository information
    github_url TEXT NOT NULL,
    repository_name VARCHAR(200) NOT NULL,
    repository_owner VARCHAR(100) NOT NULL,
    default_branch VARCHAR(100) NOT NULL DEFAULT 'main',

    -- Integration status
    is_connected BOOLEAN NOT NULL DEFAULT FALSE,
    oauth_token_hash VARCHAR(255), -- Encrypted GitHub access token
    webhook_secret VARCHAR(100), -- For GitHub webhook verification

    -- Auto-upgrade configuration
    auto_upgrade_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    upgrade_trigger VARCHAR(50) DEFAULT 'tag_release' CHECK (upgrade_trigger IN ('tag_release', 'main_push', 'manual')),
    last_sync_commit_hash VARCHAR(40),
    last_sync_time TIMESTAMP WITH TIME ZONE,

    -- Build and deployment tracking
    build_status VARCHAR(20) DEFAULT 'pending' CHECK (build_status IN ('pending', 'building', 'success', 'failed')),
    last_build_time TIMESTAMP WITH TIME ZONE,
    build_logs TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure one repo per chain
    UNIQUE(chain_id)
);

-- Social media and external links associated with chain projects
-- Stores all external references and social proof for launched chains
CREATE TABLE chain_social_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chain_id UUID NOT NULL REFERENCES chains(id) ON DELETE CASCADE,

    -- Link details
    platform VARCHAR(50) NOT NULL, -- 'twitter', 'telegram', 'discord', 'website', 'whitepaper', etc.
    url TEXT NOT NULL,
    display_name VARCHAR(200),

    -- Verification and metrics
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    follower_count INTEGER DEFAULT 0,
    last_metrics_update TIMESTAMP WITH TIME ZONE,

    -- Ordering and status
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Virtual liquidity pool state for chains in pre-graduation phase
-- Tracks the bonding curve mechanics and trading activity before mainnet deployment
CREATE TABLE virtual_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chain_id UUID NOT NULL REFERENCES chains(id) ON DELETE CASCADE,

    -- Current pool state
    cnpy_reserve DECIMAL(15,8) NOT NULL DEFAULT 0,
    token_reserve BIGINT NOT NULL DEFAULT 0,
    current_price_cnpy DECIMAL(15,8) NOT NULL DEFAULT 0,
    market_cap_usd DECIMAL(15,2) NOT NULL DEFAULT 0,

    -- Trading metrics
    total_volume_cnpy DECIMAL(15,8) NOT NULL DEFAULT 0,
    total_transactions INTEGER NOT NULL DEFAULT 0,
    unique_traders INTEGER NOT NULL DEFAULT 0,

    -- Pool status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Performance tracking
    price_24h_change_percent DECIMAL(8,4) DEFAULT 0,
    volume_24h_cnpy DECIMAL(15,8) DEFAULT 0,
    high_24h_cnpy DECIMAL(15,8) DEFAULT 0,
    low_24h_cnpy DECIMAL(15,8) DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure one pool per chain
    UNIQUE(chain_id)
);

-- Individual trading transactions within virtual pools
-- Records all buy/sell activity for bonding curve price discovery and user tracking
CREATE TABLE virtual_pool_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    virtual_pool_id UUID NOT NULL REFERENCES virtual_pools(id) ON DELETE CASCADE,
    chain_id UUID NOT NULL REFERENCES chains(id),
    user_id UUID NOT NULL REFERENCES users(id),

    -- Transaction details
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('buy', 'sell')),
    cnpy_amount DECIMAL(15,8) NOT NULL,
    token_amount BIGINT NOT NULL,
    price_per_token_cnpy DECIMAL(15,8) NOT NULL,

    -- Fees and slippage
    trading_fee_cnpy DECIMAL(15,8) NOT NULL DEFAULT 0,
    slippage_percent DECIMAL(8,4) DEFAULT 0,

    -- Blockchain transaction details
    transaction_hash VARCHAR(66), -- Set when transaction is executed
    block_height BIGINT,
    gas_used INTEGER,

    -- State after transaction
    pool_cnpy_reserve_after DECIMAL(15,8) NOT NULL,
    pool_token_reserve_after BIGINT NOT NULL,
    market_cap_after_usd DECIMAL(15,2) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User positions and holdings in virtual pools
-- Tracks individual user ownership and performance in each virtual chain
CREATE TABLE user_virtual_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id),
    chain_id UUID NOT NULL REFERENCES chains(id),
    virtual_pool_id UUID NOT NULL REFERENCES virtual_pools(id),

    -- Current position
    token_balance BIGINT NOT NULL DEFAULT 0,
    total_cnpy_invested DECIMAL(15,8) NOT NULL DEFAULT 0,
    total_cnpy_withdrawn DECIMAL(15,8) NOT NULL DEFAULT 0,
    average_entry_price_cnpy DECIMAL(15,8) NOT NULL DEFAULT 0,

    -- Performance metrics
    unrealized_pnl_cnpy DECIMAL(15,8) DEFAULT 0,
    realized_pnl_cnpy DECIMAL(15,8) DEFAULT 0,
    total_return_percent DECIMAL(8,4) DEFAULT 0,

    -- Position status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    first_purchase_at TIMESTAMP WITH TIME ZONE,
    last_activity_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure one position per user per chain
    UNIQUE(user_id, chain_id)
);

-- Media assets and files associated with chain projects
-- Stores logos, images, videos, and documents for chain marketing and documentation
CREATE TABLE chain_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chain_id UUID NOT NULL REFERENCES chains(id) ON DELETE CASCADE,

    -- File information
    asset_type VARCHAR(20) NOT NULL CHECK (asset_type IN ('logo', 'banner', 'screenshot', 'video', 'whitepaper', 'documentation')),
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),

    -- Asset metadata
    title VARCHAR(200),
    description TEXT,
    alt_text VARCHAR(500), -- For accessibility

    -- Display and organization
    display_order INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE, -- Primary logo/banner
    is_featured BOOLEAN NOT NULL DEFAULT FALSE, -- Featured in galleries

    -- Status and moderation
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    moderation_status VARCHAR(20) DEFAULT 'pending' CHECK (moderation_status IN ('pending', 'approved', 'rejected')),
    moderation_notes TEXT,

    -- Upload tracking
    uploaded_by UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Cryptographic keys for chain operations and governance
-- Stores encrypted private keys, public keys, and addresses for each chain
-- Uses Argon2 + AES-GCM encryption compatible with crypto.EncryptedPrivateKey
CREATE TABLE chain_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chain_id UUID NOT NULL REFERENCES chains(id) ON DELETE CASCADE,

    -- Cryptographic key data (based on crypto.EncryptedPrivateKey)
    address TEXT NOT NULL, -- Chain address as hex string
    public_key BYTEA NOT NULL, -- Raw public key bytes
    encrypted_private_key TEXT NOT NULL, -- Hex-encoded AES-GCM encrypted private key
    salt BYTEA NOT NULL CHECK (octet_length(salt) = 16), -- 16-byte salt for key derivation

    -- Key metadata
    key_nickname VARCHAR(100), -- Optional friendly name for the key
    key_purpose VARCHAR(50) DEFAULT 'chain_operation' CHECK (key_purpose IN ('chain_operation', 'governance', 'treasury', 'backup')),

    -- Security and lifecycle
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    rotation_count INTEGER NOT NULL DEFAULT 0, -- Track key rotations

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure one active operational key per chain
    UNIQUE(chain_id, key_purpose),
    UNIQUE(address) -- Address must be globally unique
);

-- Trigger to update the updated_at timestamp on record modification
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply the trigger to all tables with updated_at columns
CREATE TRIGGER update_chains_updated_at BEFORE UPDATE ON chains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_templates_updated_at BEFORE UPDATE ON chain_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_repos_updated_at BEFORE UPDATE ON chain_repositories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_social_updated_at BEFORE UPDATE ON chain_social_links FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pools_updated_at BEFORE UPDATE ON virtual_pools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON user_virtual_positions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON chain_assets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chain_keys_updated_at BEFORE UPDATE ON chain_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes separately
-- Indexes for chains table
CREATE INDEX idx_chains_status ON chains (status);
CREATE INDEX idx_chains_creator ON chains (created_by);
CREATE INDEX idx_chains_template ON chains (template_id);
CREATE INDEX idx_chains_launch_time ON chains (scheduled_launch_time);
CREATE INDEX idx_chains_graduation ON chains (is_graduated, graduation_time);

-- Indexes for chain_templates table
CREATE INDEX idx_templates_category ON chain_templates (template_category);
CREATE INDEX idx_templates_active ON chain_templates (is_active);

-- Indexes for users table
CREATE INDEX idx_users_wallet ON users (wallet_address);
CREATE INDEX idx_users_github ON users (github_username);
CREATE INDEX idx_users_reputation ON users (reputation_score DESC);

-- Indexes for chain_repositories table
CREATE INDEX idx_repos_chain ON chain_repositories (chain_id);
CREATE INDEX idx_repos_build_status ON chain_repositories (build_status);

-- Indexes for chain_social_links table
CREATE INDEX idx_social_chain ON chain_social_links (chain_id);
CREATE INDEX idx_social_platform ON chain_social_links (platform);
CREATE INDEX idx_social_verified ON chain_social_links (is_verified);

-- Indexes for virtual_pools table
CREATE INDEX idx_virtual_pools_active ON virtual_pools (is_active);
CREATE INDEX idx_virtual_pools_market_cap ON virtual_pools (market_cap_usd DESC);

-- Indexes for virtual_pool_transactions table
CREATE INDEX idx_vp_transactions_pool ON virtual_pool_transactions (virtual_pool_id);
CREATE INDEX idx_vp_transactions_user ON virtual_pool_transactions (user_id);
CREATE INDEX idx_vp_transactions_chain ON virtual_pool_transactions (chain_id);
CREATE INDEX idx_vp_transactions_time ON virtual_pool_transactions (created_at DESC);
CREATE INDEX idx_vp_transactions_type ON virtual_pool_transactions (transaction_type);

-- Indexes for user_virtual_positions table
CREATE INDEX idx_positions_user ON user_virtual_positions (user_id);
CREATE INDEX idx_positions_chain ON user_virtual_positions (chain_id);
CREATE INDEX idx_positions_active ON user_virtual_positions (is_active);
CREATE INDEX idx_positions_pnl ON user_virtual_positions (unrealized_pnl_cnpy DESC);

-- Indexes for chain_assets table
CREATE INDEX idx_assets_chain ON chain_assets (chain_id);
CREATE INDEX idx_assets_type ON chain_assets (asset_type);
CREATE INDEX idx_assets_primary ON chain_assets (is_primary);
CREATE INDEX idx_assets_moderation ON chain_assets (moderation_status);

-- Indexes for chain_keys table
CREATE INDEX idx_chain_keys_chain ON chain_keys (chain_id);
CREATE INDEX idx_chain_keys_address ON chain_keys (address);
CREATE INDEX idx_chain_keys_active ON chain_keys (is_active);
CREATE INDEX idx_chain_keys_purpose ON chain_keys (key_purpose);