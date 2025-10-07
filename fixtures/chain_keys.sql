-- Chain key fixtures for Launchpad API testing and development
-- Canopy blockchain addresses for chains (treasury addresses)

INSERT INTO chain_keys (id, chain_id, address, public_key, encrypted_private_key, salt, key_nickname, key_purpose, is_active) VALUES
    ('550e8400-e29b-41d4-a716-446655449001', '550e8400-e29b-41d4-a716-446655442001', 'aabbccdd11223344', E'\\x0102030405060708090a0b0c0d0e0f10', 'encrypted_key_data_001', E'\\x0102030405060708090a0b0c0d0e0f10', 'Treasury Address', 'treasury', true),
    ('550e8400-e29b-41d4-a716-446655449002', '550e8400-e29b-41d4-a716-446655442002', 'ffeeddccbbaa9988', E'\\x1112131415161718191a1b1c1d1e1f20', 'encrypted_key_data_002', E'\\x1112131415161718191a1b1c1d1e1f20', 'Treasury Address', 'treasury', true);
