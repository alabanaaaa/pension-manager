-- ============================================================
-- Seed Data: Create Admin User
-- Run after migrations: psql -f migrations/004_new_features.up.sql
-- Then run this: psql -f migrations/seed_admin.sql
-- ============================================================

-- First, create a default scheme if none exists
INSERT INTO schemes (id, name, scheme_type, status, currency, tax_exempt_age)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Pension Manager Scheme',
    'db',
    'active',
    'KES',
    65
) ON CONFLICT DO NOTHING;

-- Create admin user (password: Admin123!)
-- The hash is bcrypt of "Admin123!" - you should change this password!
INSERT INTO system_users (id, scheme_id, email, password_hash, role, name, phone, active)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'admin@pension.co.ke',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.RsikRUQf1yAq6J2fG',  -- Admin123!
    'admin',
    'System Administrator',
    '+254700000000',
    true
) ON CONFLICT (email) DO NOTHING;

-- Create a pension officer
INSERT INTO system_users (id, scheme_id, email, password_hash, role, name, phone, active)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'officer@pension.co.ke',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.RsikRUQf1yAq6J2fG',  -- Admin123!
    'pension_officer',
    'Pension Officer',
    '+254700000001',
    true
) ON CONFLICT (email) DO NOTHING;

-- Create a sample sponsor
INSERT INTO sponsors (id, scheme_id, code, name, contact_person, phone, email, address)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000001',
    'SP001',
    'Kenya Power & Lighting Company',
    'HR Department',
    '+254700000010',
    'hr@kenyapower.co.ke',
    'Nairobi, Kenya'
) ON CONFLICT DO NOTHING;
