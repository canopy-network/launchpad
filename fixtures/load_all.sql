-- Master fixture loader for Launchpad API testing and development
-- This file loads all fixture data in the correct dependency order

-- Load fixtures in dependency order
\i fixtures/users.sql
\i fixtures/chain_templates.sql
\i fixtures/chains.sql
\i fixtures/chain_keys.sql
\i fixtures/virtual_pools.sql
\i fixtures/virtual_pool_transactions.sql
\i fixtures/user_virtual_positions.sql
\i fixtures/chain_repositories.sql
\i fixtures/chain_social_links.sql
\i fixtures/chain_assets.sql

-- Update user statistics based on fixture data
UPDATE users SET
    total_chains_created = (SELECT COUNT(*) FROM chains WHERE created_by = users.id),
    total_cnpy_invested = COALESCE((SELECT SUM(total_cnpy_invested) FROM user_virtual_positions WHERE user_id = users.id), 0);

-- Analyze tables for query optimizer
ANALYZE chains;
ANALYZE users;
ANALYZE chain_templates;
ANALYZE virtual_pools;
ANALYZE virtual_pool_transactions;
ANALYZE user_virtual_positions;
