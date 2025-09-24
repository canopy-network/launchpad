-- Core entity representing a blockchain chain in the Scanopy ecosystem
-- This is the primary table that holds all chain-specific configuration and metadata
-- Links to all other entities in the system through foreign key relationships
CREATE TABLE chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Basic chain identification
    chain_name VARCHAR(100) NOT NULL UNIQUE,
    token_symbol VARCHAR(20) NOT NULL,
    chain_description TEXT,
    
    -- Template and technical configuration
    template_id UUID NOT NULL REFERENCES chain_templates(id),
    consensus_mechanism VARCHAR(50) NOT NULL DEFAULT 'nestbft',
    
    -- Economic parameters
    token_total_supply BIGINT NOT NULL DEFAULT 1000000000,
    graduation_threshold_usd DECIMAL(15,2) NOT NULL DEFAULT 50000.00,
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
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Indexes for common queries
    INDEX idx_chains_status (status),
    INDEX idx_chains_creator (created_by),
    INDEX idx_chains_template (template_id),
    INDEX idx_chains_launch_time (scheduled_launch_time),
    INDEX idx_chains_graduation (is_graduated, graduation_time)
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
    supported_languages TEXT[] NOT NULL, -- ['go', 'rust', 'javascript']
    default_modules TEXT[] NOT NULL, -- ['staking', 'governance', 'ibc']
    required_modules TEXT[] NOT NULL,
    
    -- Default economic parameters
    default_consensus VARCHAR(50) NOT NULL DEFAULT 'tendermint',
    default_token_supply BIGINT NOT NULL DEFAULT 1000000000,
    default_validator_count INTEGER NOT NULL DEFAULT 10,
    
    -- Template metadata
    complexity_level VARCHAR(20) NOT NULL CHECK (complexity_level IN ('beginner', 'intermediate', 'advanced')),
    estimated_deployment_time_minutes INTEGER NOT NULL DEFAULT 30,
    documentation_url TEXT,
    example_chains TEXT[], -- Array of successful chain examples
    
    -- Versioning and maintenance
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    INDEX idx_templates_category (template_category),
    INDEX idx_templates_complexity (complexity_level),
    INDEX idx_templates_active (is_active)
);

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
    notification_preferences JSONB DEFAULT '{"email": false, "browser": true}',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_tier VARCHAR(20) DEFAULT 'basic' CHECK (verification_tier IN ('basic', 'verified', 'premium')),
    
    -- Activity tracking
    total_chains_created INTEGER NOT NULL DEFAULT 0,
    total_cnpy_invested DECIMAL(15,8) NOT NULL DEFAULT 0,
    reputation_score INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_users_wallet (wallet_address),
    INDEX idx_users_github (github_username),
    INDEX idx_users_reputation (reputation_score DESC)
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
    UNIQUE(chain_id),
    INDEX idx_repos_chain (chain_id),
    INDEX idx_repos_build_status (build_status)
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
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    INDEX idx_social_chain (chain_id),
    INDEX idx_social_platform (platform),
    INDEX idx_social_verified (is_verified)
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
    UNIQUE(chain_id),
    INDEX idx_virtual_pools_active (is_active),
    INDEX idx_virtual_pools_market_cap (market_cap_usd DESC)
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
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    INDEX idx_vp_transactions_pool (virtual_pool_id),
    INDEX idx_vp_transactions_user (user_id),
    INDEX idx_vp_transactions_chain (chain_id),
    INDEX idx_vp_transactions_time (created_at DESC),
    INDEX idx_vp_transactions_type (transaction_type)
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
    UNIQUE(user_id, chain_id),
    INDEX idx_positions_user (user_id),
    INDEX idx_positions_chain (chain_id),
    INDEX idx_positions_active (is_active),
    INDEX idx_positions_pnl (unrealized_pnl_cnpy DESC)
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
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    INDEX idx_assets_chain (chain_id),
    INDEX idx_assets_type (asset_type),
    INDEX idx_assets_primary (is_primary),
    INDEX idx_assets_moderation (moderation_status)
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
