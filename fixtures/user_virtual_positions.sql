-- User virtual position fixtures for Launchpad API testing and development
-- User holdings and P&L tracking in virtual pools

INSERT INTO user_virtual_positions (id, user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested, total_cnpy_withdrawn, average_entry_price_cnpy, unrealized_pnl_cnpy, realized_pnl_cnpy, total_return_percent, is_active, first_purchase_at, last_activity_at) VALUES
    ('550e8400-e29b-41d4-a716-446655447001', '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655442001', '550e8400-e29b-41d4-a716-446655443001', 30000000, 1000.00000000, 0, 0.00003333, 0, 0, 0, true, '2024-02-12 10:30:00+00', '2024-02-12 10:30:00+00'),
    ('550e8400-e29b-41d4-a716-446655447002', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655442001', '550e8400-e29b-41d4-a716-446655443001', 14850000, 500.00000000, 0, 0.00003367, -15.00000000, 0, -3.0000, true, '2024-02-13 14:20:00+00', '2024-02-13 14:20:00+00'),
    ('550e8400-e29b-41d4-a716-446655447003', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655442002', '550e8400-e29b-41d4-a716-446655443002', 0, 2000.00000000, 2010.00000000, 0.00010256, 0, 10.00000000, 0.5000, false, '2024-02-20 09:15:00+00', '2024-02-22 16:45:00+00');
