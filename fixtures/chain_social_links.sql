-- Chain social link fixtures for Launchpad API testing and development
-- Social media and community links for chains

INSERT INTO chain_social_links (id, chain_id, platform, url, display_name, is_verified, follower_count, last_metrics_update, display_order, is_active) VALUES
    ('550e8400-e29b-41d4-a716-446655445001', '550e8400-e29b-41d4-a716-446655442001', 'twitter', 'https://twitter.com/defiswappro', 'DeFiSwap Pro Official', true, 15420, '2024-02-14 12:00:00+00', 1, true),
    ('550e8400-e29b-41d4-a716-446655445002', '550e8400-e29b-41d4-a716-446655442001', 'discord', 'https://discord.gg/defiswappro', 'DeFiSwap Community', true, 8930, '2024-02-14 12:00:00+00', 2, true),
    ('550e8400-e29b-41d4-a716-446655445003', '550e8400-e29b-41d4-a716-446655442001', 'website', 'https://defiswappro.io', 'Official Website', true, 0, null, 0, true),
    ('550e8400-e29b-41d4-a716-446655445004', '550e8400-e29b-41d4-a716-446655442002', 'twitter', 'https://twitter.com/metarealm', 'MetaRealm Gaming', true, 45200, '2024-02-28 10:00:00+00', 1, true),
    ('550e8400-e29b-41d4-a716-446655445005', '550e8400-e29b-41d4-a716-446655442002', 'telegram', 'https://t.me/metarealmgaming', 'MetaRealm Telegram', true, 12100, '2024-02-28 10:00:00+00', 2, true);
