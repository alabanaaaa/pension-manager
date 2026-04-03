-- ============================================================
-- 002: Hospital Management, M-Pesa tracking, Member enhancements
-- ============================================================

-- Hospital tables
CREATE TABLE IF NOT EXISTS hospitals (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    name            TEXT NOT NULL,
    address         TEXT,
    phone           TEXT,
    email           TEXT,
    license_number  TEXT,
    account_balance BIGINT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_hospitals_scheme ON hospitals(scheme_id);
CREATE INDEX idx_hospitals_status ON hospitals(status);

CREATE TABLE IF NOT EXISTS medical_limits (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id         UUID NOT NULL REFERENCES members(id),
    scheme_id         UUID NOT NULL REFERENCES schemes(id),
    inpatient_limit   BIGINT NOT NULL DEFAULT 0,
    outpatient_limit  BIGINT NOT NULL DEFAULT 0,
    period            TEXT NOT NULL DEFAULT 'annual',
    effective_date    DATE NOT NULL,
    expiry_date       DATE,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_medical_limits_member ON medical_limits(member_id);
CREATE INDEX idx_medical_limits_scheme ON medical_limits(scheme_id);

CREATE TABLE IF NOT EXISTS medical_expenditures (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id             UUID NOT NULL REFERENCES members(id),
    scheme_id             UUID NOT NULL REFERENCES schemes(id),
    hospital_id           UUID REFERENCES hospitals(id),
    date_of_service       DATE NOT NULL,
    date_submitted        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    service_type          TEXT NOT NULL CHECK (service_type IN ('inpatient', 'outpatient', 'pharmacy', 'dental', 'optical', 'maternity', 'surgical', 'other')),
    description           TEXT,
    amount_charged        BIGINT NOT NULL,
    amount_covered        BIGINT NOT NULL DEFAULT 0,
    member_responsibility BIGINT NOT NULL DEFAULT 0,
    status                TEXT NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'approved', 'rejected', 'paid')),
    invoice_number        TEXT,
    receipt_number        TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_medical_exp_member ON medical_expenditures(member_id);
CREATE INDEX idx_medical_exp_scheme ON medical_expenditures(scheme_id);
CREATE INDEX idx_medical_exp_hospital ON medical_expenditures(hospital_id);
CREATE INDEX idx_medical_exp_status ON medical_expenditures(status);
CREATE INDEX idx_medical_exp_date ON medical_expenditures(date_submitted);

-- M-Pesa tracking on contributions
ALTER TABLE contributions ADD COLUMN IF NOT EXISTS mpesa_checkout_id TEXT;
ALTER TABLE contributions ADD COLUMN IF NOT EXISTS mpesa_receipt TEXT;
ALTER TABLE contributions ADD COLUMN IF NOT EXISTS phone_number TEXT;
ALTER TABLE contributions ADD COLUMN IF NOT EXISTS transaction_id TEXT;
CREATE INDEX IF NOT EXISTS idx_contributions_mpesa ON contributions(mpesa_checkout_id);

-- OTP support on system_users
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS otp_code TEXT;
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS otp_expiry TIMESTAMPTZ;

-- Member enhancements
ALTER TABLE members ADD COLUMN IF NOT EXISTS date_of_death TIMESTAMPTZ;
ALTER TABLE members ADD COLUMN IF NOT EXISTS total_withdrawals BIGINT NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS last_withdrawal_date TIMESTAMPTZ;
ALTER TABLE members ADD COLUMN IF NOT EXISTS pin TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS photograph TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS fingerprint_data TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS children_under_21_count INT NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS membership_card_issue_date TIMESTAMPTZ;
ALTER TABLE members ADD COLUMN IF NOT EXISTS membership_card_status TEXT CHECK (membership_card_status IN ('issue', 'Not issued', 'returned', 'lost'));
ALTER TABLE members ADD COLUMN IF NOT EXISTS previous_sponsors TEXT[];
ALTER TABLE members ADD COLUMN IF NOT EXISTS cessation_date TIMESTAMPTZ;
ALTER TABLE members ADD COLUMN IF NOT EXISTS cessation_reason TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS tax_exempt_reason TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS tax_exempt_attachment TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS tax_exempt_cutoff_date TIMESTAMPTZ;
ALTER TABLE members ADD COLUMN IF NOT EXISTS member_contribution_rate NUMERIC(5,2) NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS sponsor_contribution_rate NUMERIC(5,2) NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS inpatient_limit BIGINT NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS outpatient_limit BIGINT NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS portal_enabled BOOLEAN NOT NULL DEFAULT true;

-- Maker-Checker: pending changes table
CREATE TABLE IF NOT EXISTS pending_changes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type     TEXT NOT NULL CHECK (entity_type IN ('member', 'beneficiary', 'claim')),
    entity_id       UUID NOT NULL,
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    requested_by    UUID NOT NULL,
    change_type     TEXT NOT NULL CHECK (change_type IN ('create', 'update', 'delete')),
    before_data     JSONB,
    after_data      JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by     UUID,
    reviewed_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_pending_changes_entity ON pending_changes(entity_type, entity_id);
CREATE INDEX idx_pending_changes_status ON pending_changes(status);
CREATE INDEX idx_pending_changes_scheme ON pending_changes(scheme_id);

-- Document management table
CREATE TABLE IF NOT EXISTS documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type     TEXT NOT NULL,
    entity_id       UUID NOT NULL,
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    document_type   TEXT NOT NULL,  -- death_certificate, id, marriage_cert, claim_form, etc.
    file_name       TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    mime_type       TEXT NOT NULL,
    storage_path    TEXT NOT NULL,  -- S3 key or local path
    uploaded_by     UUID NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_documents_entity ON documents(entity_type, entity_id);
CREATE INDEX idx_documents_scheme ON documents(scheme_id);
CREATE INDEX idx_documents_type ON documents(document_type);

-- Tax exemption reminders table
CREATE TABLE IF NOT EXISTS tax_exemption_reminders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id       UUID NOT NULL REFERENCES members(id),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    reminder_type   TEXT NOT NULL DEFAULT 'kra_renewal',
    due_date        DATE NOT NULL,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_tax_reminders_member ON tax_exemption_reminders(member_id);
CREATE INDEX idx_tax_reminders_due ON tax_exemption_reminders(due_date);

-- Member portal login tracking
CREATE TABLE IF NOT EXISTS member_login_log (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id   UUID NOT NULL REFERENCES members(id),
    login_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_login_log_member ON member_login_log(member_id);
CREATE INDEX idx_login_log_time ON member_login_log(login_at);

-- Member feedback
CREATE TABLE IF NOT EXISTS feedback (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id   UUID NOT NULL REFERENCES members(id),
    scheme_id   UUID NOT NULL REFERENCES schemes(id),
    subject     TEXT NOT NULL,
    message     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved')),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_feedback_member ON feedback(member_id);
CREATE INDEX idx_feedback_scheme ON feedback(scheme_id);

-- IP Blacklist for security
CREATE TABLE IF NOT EXISTS ip_blacklist (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_address  TEXT UNIQUE NOT NULL,
    reason      TEXT NOT NULL,
    added_by    UUID NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_ip_blacklist_active ON ip_blacklist(active);

-- Login attempt tracking
CREATE TABLE IF NOT EXISTS login_attempts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ip_address      TEXT NOT NULL,
    email           TEXT NOT NULL,
    success         BOOLEAN NOT NULL,
    attempted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address, attempted_at);
CREATE INDEX idx_login_attempts_email ON login_attempts(email, attempted_at);
