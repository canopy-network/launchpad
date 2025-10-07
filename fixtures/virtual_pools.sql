-- Virtual pool fixtures for Launchpad API testing and development
-- Trading pools for active chains with bonding curve mechanics

INSERT INTO virtual_pools (id, chain_id, cnpy_reserve, token_reserve, current_price_cnpy, market_cap_usd, total_volume_cnpy, total_transactions, unique_traders, is_active, price_24h_change_percent, volume_24h_cnpy, high_24h_cnpy, low_24h_cnpy) VALUES
    ('550e8400-e29b-41d4-a716-446655443001', '550e8400-e29b-41d4-a716-446655442001', 25000.50000000, 750000000, 0.00003333, 45000.00, 8500.75000000, 245, 67, true, 5.23, 1200.50000000, 0.00003500, 0.00003100),
    ('550e8400-e29b-41d4-a716-446655443002', '550e8400-e29b-41d4-a716-446655442002', 35000.25000000, 350000000, 0.00010000, 85000.00, 15200.25000000, 412, 123, false, 12.45, 2100.75000000, 0.00011200, 0.00009800);
