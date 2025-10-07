-- Chain repository fixtures for Launchpad API testing and development
-- GitHub integration settings for chains with source code repositories

INSERT INTO chain_repositories (id, chain_id, github_url, repository_name, repository_owner, default_branch, is_connected, auto_upgrade_enabled, upgrade_trigger, last_sync_commit_hash, last_sync_time, build_status, last_build_time) VALUES
    ('550e8400-e29b-41d4-a716-446655444001', '550e8400-e29b-41d4-a716-446655442001', 'https://github.com/alice-dev/defiswap-pro', 'defiswap-pro', 'alice-dev', 'main', true, true, 'tag_release', 'a1b2c3d4e5f6789012345678901234567890abcd', '2024-02-10 14:30:00+00', 'success', '2024-02-10 15:45:00+00'),
    ('550e8400-e29b-41d4-a716-446655444002', '550e8400-e29b-41d4-a716-446655442002', 'https://github.com/bob-games/metarealm-chain', 'metarealm-chain', 'bob-games', 'main', true, true, 'main_push', 'b2c3d4e5f6a78901234567890123456789abcdef', '2024-02-25 11:20:00+00', 'success', '2024-02-25 11:35:00+00');
