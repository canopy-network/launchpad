-- User fixtures for Launchpad API testing and development
-- Creates sample user accounts with different verification levels

INSERT INTO users (id, wallet_address, email, username, display_name, bio, github_username, is_verified, verification_tier, total_chains_created, reputation_score) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', '0x1234567890abcdef1234567890abcdef12345678', 'alice@example.com', 'alice_dev', 'Alice Johnson', 'Blockchain developer and DeFi enthusiast', 'alice-dev', true, 'verified', 3, 850),
    ('550e8400-e29b-41d4-a716-446655440001', '0xabcdef1234567890abcdef1234567890abcdef12', 'bob@example.com', 'bob_gaming', 'Bob Smith', 'Game developer building the future of Web3 gaming', 'bob-games', true, 'premium', 5, 1200),
    ('550e8400-e29b-41d4-a716-446655440002', '0x9876543210fedcba9876543210fedcba98765432', 'carol@example.com', 'carol_enterprise', 'Carol Williams', 'Enterprise blockchain solutions architect', 'carol-enterprise', true, 'verified', 2, 750),
    ('550e8400-e29b-41d4-a716-446655440003', '0xfedcba9876543210fedcba9876543210fedcba98', 'dave@example.com', 'dave_social', 'Dave Brown', 'Building social networks on blockchain', 'dave-social', false, 'basic', 1, 300);
