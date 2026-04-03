-- ============================================================
-- Pension Manager — Initial Schema
-- Post Retirement Medical Fund & Pension Management System
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- SCHEMES (DB / DC / Medical Fund)
-- ============================================================
CREATE TABLE schemes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    scheme_type     TEXT NOT NULL CHECK (scheme_type IN ('db', 'dc', 'medical')),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
    currency        TEXT DEFAULT 'KES',
    tax_exempt_age  INT DEFAULT 65,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- MEMBERS
-- ============================================================
CREATE TABLE members (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    member_no           TEXT UNIQUE NOT NULL,
    first_name          TEXT NOT NULL,
    last_name           TEXT NOT NULL,
    other_names         TEXT,
    gender              TEXT CHECK (gender IN ('male', 'female', 'other')),
    date_of_birth       DATE NOT NULL,
    nationality         TEXT DEFAULT 'Kenyan',
    id_number           TEXT UNIQUE,
    kra_pin             TEXT,
    email               TEXT,
    phone               TEXT,
    postal_address      TEXT,
    postal_code         TEXT,
    town                TEXT,
    marital_status      TEXT CHECK (marital_status IN ('single', 'married', 'separated', 'divorced', 'widowed')),
    spouse_name         TEXT,
    next_of_kin         TEXT,
    next_of_kin_phone   TEXT,
    bank_name           TEXT,
    bank_branch         TEXT,
    bank_account        TEXT,
    payroll_no          TEXT,
    designation         TEXT,
    department          TEXT,
    sponsor_id          UUID,             -- FK deferred for circular ref
    date_first_appt     DATE,
    date_joined_scheme  DATE NOT NULL,
    expected_retirement DATE,
    membership_status   TEXT NOT NULL DEFAULT 'active' CHECK (membership_status IN ('active', 'inactive', 'suspended', 'deferred', 'retired', 'deceased')),
    basic_salary        BIGINT DEFAULT 0,
    account_balance     BIGINT DEFAULT 0,  -- cumulative to date
    last_contribution   TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_members_scheme ON members(scheme_id);
CREATE INDEX idx_members_status ON members(membership_status);
CREATE INDEX idx_members_sponsor ON members(sponsor_id);

-- ============================================================
-- SPONSORS (Employers)
-- ============================================================
CREATE TABLE sponsors (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    code            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    contact_person  TEXT,
    phone           TEXT,
    email           TEXT,
    address         TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_sponsors_scheme ON sponsors(scheme_id);

-- Add FK after sponsors table exists
ALTER TABLE members ADD CONSTRAINT fk_members_sponsor
    FOREIGN KEY (sponsor_id) REFERENCES sponsors(id);

-- ============================================================
-- BENEFICIARIES
-- ============================================================
CREATE TABLE beneficiaries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id       UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    relationship    TEXT NOT NULL,
    date_of_birth   DATE,
    id_number       TEXT,
    phone           TEXT,
    physical_address TEXT,
    allocation_pct  NUMERIC(5,2) NOT NULL CHECK (allocation_pct >= 0 AND allocation_pct <= 100),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'removed', 'deceased')),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_beneficiaries_member ON beneficiaries(member_id);

-- ============================================================
-- CONTRIBUTIONS
-- ============================================================
CREATE TABLE contributions (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id               UUID NOT NULL REFERENCES members(id),
    scheme_id               UUID NOT NULL REFERENCES schemes(id),
    sponsor_id              UUID REFERENCES sponsors(id),
    period                  DATE NOT NULL,  -- first day of contribution month
    employee_amount         BIGINT NOT NULL DEFAULT 0,
    employer_amount         BIGINT NOT NULL DEFAULT 0,
    avc_amount              BIGINT NOT NULL DEFAULT 0,  -- Additional Voluntary Contribution
    total_amount            BIGINT NOT NULL GENERATED ALWAYS AS (employee_amount + employer_amount + avc_amount) STORED,
    payment_method          TEXT CHECK (payment_method IN ('mpesa', 'bank_transfer', 'cheque', 'cash', 'standing_order')),
    payment_ref             TEXT,
    receipt_no              TEXT,
    status                  TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'reconciled', 'on_hold', 'rejected')),
    registered              BOOLEAN DEFAULT true,  -- registered vs unregistered contributions
    notes                   TEXT,
    created_by              UUID,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at            TIMESTAMPTZ
);
CREATE INDEX idx_contributions_member ON contributions(member_id);
CREATE INDEX idx_contributions_scheme ON contributions(scheme_id);
CREATE INDEX idx_contributions_period ON contributions(scheme_id, period);
CREATE INDEX idx_contributions_status ON contributions(status);

-- ============================================================
-- CONTRIBUTION SCHEDULES (monthly remittance from sponsors)
-- ============================================================
CREATE TABLE contribution_schedules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sponsor_id      UUID NOT NULL REFERENCES sponsors(id),
    period          DATE NOT NULL,
    total_employees INT NOT NULL,
    total_amount    BIGINT NOT NULL,
    prev_employees  INT,
    prev_amount     BIGINT,
    employee_diff   INT,
    amount_diff     BIGINT,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'balanced', 'on_hold', 'posted')),
    reconciliation_notes TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    posted_at       TIMESTAMPTZ
);
CREATE INDEX idx_schedules_sponsor ON contribution_schedules(sponsor_id, period);

-- ============================================================
-- CLAIMS / WITHDRAWALS
-- ============================================================
CREATE TABLE claims (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    claim_type          TEXT NOT NULL CHECK (claim_type IN (
        'normal_retirement', 'early_retirement', 'late_retirement',
        'ill_health_retirement', 'death_in_service', 'leaving_service',
        'deferred_retirement', 'medical_claim', 'ex_gratia'
    )),
    claim_form_no       TEXT,
    date_of_claim       DATE NOT NULL,
    date_of_leaving     DATE,
    leaving_reason      TEXT,
    status              TEXT NOT NULL DEFAULT 'submitted' CHECK (status IN (
        'submitted', 'resubmission', 'under_review', 'rejected', 'accepted', 'paid'
    )),
    rejection_reason    TEXT,
    examiner_id         UUID,  -- user who examined
    settlement_date     DATE,
    cheque_ref          TEXT,  -- bank transfer reference or cheque number
    cheque_date         DATE,
    amount              BIGINT,
    partial_payments    JSONB,  -- [{date, amount, ref}]
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ,
    paid_at             TIMESTAMPTZ
);
CREATE INDEX idx_claims_member ON claims(member_id);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_claims_type ON claims(claim_type);

-- ============================================================
-- CLAIM DOCUMENTS
-- ============================================================
CREATE TABLE claim_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    claim_id        UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    doc_type        TEXT NOT NULL CHECK (doc_type IN (
        'death_certificate', 'id_copy', 'kra_pin', 'sponsor_clearance',
        'letters_of_administration', 'relationship_affidavit',
        'marriage_certificate', 'claim_form', 'bank_statement', 'other'
    )),
    file_name       TEXT NOT NULL,
    file_path       TEXT NOT NULL,  -- S3 key
    file_size       BIGINT,
    uploaded_by     UUID,
    uploaded_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_claim_docs_claim ON claim_documents(claim_id);

-- ============================================================
-- ELECTIONS (Online Voting)
-- ============================================================
CREATE TABLE elections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    title           TEXT NOT NULL,
    description     TEXT,
    election_type   TEXT NOT NULL CHECK (election_type IN ('trustee', 'agenda', 'agm')),
    max_candidates  INT NOT NULL DEFAULT 3,
    start_at        TIMESTAMPTZ NOT NULL,
    end_at          TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'open', 'closed', 'cancelled')),
    allow_ussd      BOOLEAN DEFAULT false,
    allow_web       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_elections_scheme ON elections(scheme_id);
CREATE INDEX idx_elections_status ON elections(status);

-- ============================================================
-- CANDIDATES
-- ============================================================
CREATE TABLE candidates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    position        TEXT NOT NULL,
    manifesto       TEXT,
    photo_path      TEXT,  -- S3 key
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_candidates_election ON candidates(election_id);

-- ============================================================
-- VOTES
-- ============================================================
CREATE TABLE votes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id),
    member_id       UUID NOT NULL REFERENCES members(id),
    candidate_id    UUID NOT NULL REFERENCES candidates(id),
    channel         TEXT NOT NULL CHECK (channel IN ('web', 'ussd')),
    voter_phone     TEXT,
    gps_location    TEXT,  -- lat,lng
    ip_address      TEXT,
    user_agent      TEXT,
    voted_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(election_id, member_id, candidate_id)
);
CREATE INDEX idx_votes_election ON votes(election_id);
CREATE INDEX idx_votes_member ON votes(member_id);
CREATE INDEX idx_votes_candidate ON votes(candidate_id);

-- ============================================================
-- MEMBER PORTAL CHANGE REQUESTS (Maker-Checker)
-- ============================================================
CREATE TABLE change_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id       UUID NOT NULL REFERENCES members(id),
    request_type    TEXT NOT NULL CHECK (request_type IN (
        'update_contact', 'add_beneficiary', 'remove_beneficiary',
        'change_allocation', 'update_photo', 'update_bank_details'
    )),
    old_values      JSONB,
    new_values      JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason TEXT,
    requested_by    UUID,
    reviewed_by     UUID,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ
);
CREATE INDEX idx_change_requests_member ON change_requests(member_id);
CREATE INDEX idx_change_requests_status ON change_requests(status);

-- ============================================================
-- EVENTS (Hash-Chained Audit Trail)
-- ============================================================
CREATE TABLE events (
    id              BIGSERIAL PRIMARY KEY,
    scheme_id       UUID REFERENCES schemes(id),
    entity_type     TEXT NOT NULL,  -- member, contribution, claim, vote, beneficiary
    entity_id       UUID NOT NULL,
    event_seq       BIGINT NOT NULL,
    event_type      TEXT NOT NULL,
    event_data      JSONB NOT NULL,
    previous_hash   TEXT NOT NULL DEFAULT '',
    event_hash      TEXT NOT NULL,
    created_by      UUID,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_events_scheme_seq ON events(scheme_id, event_seq);
CREATE INDEX idx_events_entity ON events(entity_type, entity_id);
CREATE INDEX idx_events_type ON events(scheme_id, event_type);

-- ============================================================
-- SYSTEM USERS (Admin, Officer, Member Portal)
-- ============================================================
CREATE TABLE system_users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID REFERENCES schemes(id),
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'pension_officer', 'claims_examiner', 'auditor', 'member')),
    member_id       UUID REFERENCES members(id),  -- for member portal users
    name            TEXT NOT NULL,
    phone           TEXT,
    active          BOOLEAN DEFAULT true,
    locked          BOOLEAN DEFAULT false,
    failed_logins   INT DEFAULT 0,
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_system_users_scheme ON system_users(scheme_id);
CREATE INDEX idx_system_users_role ON system_users(role);

-- ============================================================
-- AUDIT LOG (Human-readable audit trail)
-- ============================================================
CREATE TABLE audit_log (
    id              BIGSERIAL PRIMARY KEY,
    scheme_id       UUID REFERENCES schemes(id),
    user_id         UUID REFERENCES system_users(id),
    action          TEXT NOT NULL,
    entity_type     TEXT,
    entity_id       UUID,
    old_values      JSONB,
    new_values      JSONB,
    ip_address      TEXT,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_scheme ON audit_log(scheme_id);
CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);

-- ============================================================
-- M-PESA TRANSACTIONS
-- ============================================================
CREATE TABLE mpesa_transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID REFERENCES schemes(id),
    member_id       UUID REFERENCES members(id),
    mpesa_receipt   TEXT UNIQUE NOT NULL,
    phone_number    TEXT NOT NULL,
    amount          BIGINT NOT NULL,
    contribution_id UUID REFERENCES contributions(id),
    checkout_id     TEXT,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'failed', 'cancelled')),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_mpesa_scheme ON mpesa_transactions(scheme_id);
CREATE INDEX idx_mpesa_member ON mpesa_transactions(member_id);

-- ============================================================
-- DOCUMENTS (General file storage metadata)
-- ============================================================
CREATE TABLE documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID REFERENCES schemes(id),
    entity_type     TEXT NOT NULL,
    entity_id       UUID NOT NULL,
    doc_type        TEXT NOT NULL,
    file_name       TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    file_size       BIGINT,
    uploaded_by     UUID REFERENCES system_users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_documents_entity ON documents(entity_type, entity_id);

-- ============================================================
-- SYSTEM METRICS (Settings, config)
-- ============================================================
CREATE TABLE system_metrics (
    id              BIGSERIAL PRIMARY KEY,
    scheme_id       UUID REFERENCES schemes(id),
    metric_name     TEXT NOT NULL,
    metric_value    TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_metrics_unique ON system_metrics(scheme_id, metric_name);

-- ============================================================
-- NOTIFICATIONS (SMS, Email queue)
-- ============================================================
CREATE TABLE notifications (
    id              BIGSERIAL PRIMARY KEY,
    scheme_id       UUID REFERENCES schemes(id),
    recipient       TEXT NOT NULL,
    channel         TEXT NOT NULL CHECK (channel IN ('sms', 'email')),
    subject         TEXT,
    body            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    sent_at         TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_notifications_status ON notifications(status);

-- ============================================================
-- REMINDERS (Tax exemption renewals, pending bills >45 days)
-- ============================================================
CREATE TABLE reminders (
    id              BIGSERIAL PRIMARY KEY,
    scheme_id       UUID REFERENCES schemes(id),
    reminder_type   TEXT NOT NULL CHECK (reminder_type IN ('tax_exemption', 'pending_bill', 'contribution_overdue')),
    entity_type     TEXT,
    entity_id       UUID,
    message         TEXT NOT NULL,
    due_date        DATE,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'dismissed')),
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_reminders_status ON reminders(status, due_date);
