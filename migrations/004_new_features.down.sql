-- ============================================================
-- Pension Manager — Rollback New Features Migration
-- ============================================================

-- Drop in reverse order of creation (due to FK constraints)

DROP TABLE IF EXISTS annual_statements;
DROP TABLE IF EXISTS tax_exemptions;
DROP TABLE IF EXISTS audit_log_chain;
DROP TABLE IF EXISTS merkle_roots;
DROP TABLE IF EXISTS signature_configs;
DROP TABLE IF EXISTS digital_signatures;
DROP TABLE IF EXISTS beneficiary_drawdowns;
DROP TABLE IF EXISTS death_in_service;
DROP TABLE IF EXISTS pending_claims;
DROP TABLE IF EXISTS unregistered_contributions;
DROP TABLE IF EXISTS edi_records;
DROP TABLE IF EXISTS edi_files;
DROP TABLE IF EXISTS pending_member_changes;
DROP TABLE IF EXISTS pending_member_registrations;
