-- Rollback 002: Hospital Management, M-Pesa tracking, Member enhancements

DROP TABLE IF EXISTS tax_exemption_reminders CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS pending_changes CASCADE;

-- Drop hospital tables
DROP TABLE IF EXISTS medical_expenditures CASCADE;
DROP TABLE IF EXISTS medical_limits CASCADE;
DROP TABLE IF EXISTS hospitals CASCADE;

-- Remove M-Pesa columns from contributions
ALTER TABLE contributions DROP COLUMN IF EXISTS mpesa_checkout_id;
ALTER TABLE contributions DROP COLUMN IF EXISTS mpesa_receipt;
ALTER TABLE contributions DROP COLUMN IF EXISTS phone_number;
ALTER TABLE contributions DROP COLUMN IF EXISTS transaction_id;

-- Remove member enhancement columns
ALTER TABLE members DROP COLUMN IF EXISTS date_of_death;
ALTER TABLE members DROP COLUMN IF EXISTS total_withdrawals;
ALTER TABLE members DROP COLUMN IF EXISTS last_withdrawal_date;
ALTER TABLE members DROP COLUMN IF EXISTS pin;
ALTER TABLE members DROP COLUMN IF EXISTS photograph;
ALTER TABLE members DROP COLUMN IF EXISTS fingerprint_data;
ALTER TABLE members DROP COLUMN IF EXISTS children_under_21_count;
ALTER TABLE members DROP COLUMN IF EXISTS membership_card_issue_date;
ALTER TABLE members DROP COLUMN IF EXISTS membership_card_status;
ALTER TABLE members DROP COLUMN IF EXISTS previous_sponsors;
ALTER TABLE members DROP COLUMN IF EXISTS cessation_date;
ALTER TABLE members DROP COLUMN IF EXISTS cessation_reason;
ALTER TABLE members DROP COLUMN IF EXISTS tax_exempt_reason;
ALTER TABLE members DROP COLUMN IF EXISTS tax_exempt_attachment;
ALTER TABLE members DROP COLUMN IF EXISTS tax_exempt_cutoff_date;
ALTER TABLE members DROP COLUMN IF EXISTS member_contribution_rate;
ALTER TABLE members DROP COLUMN IF EXISTS sponsor_contribution_rate;
ALTER TABLE members DROP COLUMN IF EXISTS inpatient_limit;
ALTER TABLE members DROP COLUMN IF EXISTS outpatient_limit;
