-- ============================================================
-- Pension Manager — New Features Migration
-- Member Management, Reconciliation, Claims, Benefits, 
-- Signatures, and Enhanced Audit
-- ============================================================

-- ============================================================
-- PENDING MEMBER REGISTRATIONS (Maker-Checker Workflow)
-- ============================================================
CREATE TABLE pending_member_registrations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    member_no           TEXT NOT NULL,
    first_name          TEXT NOT NULL,
    last_name           TEXT NOT NULL,
    other_names         TEXT,
    gender              TEXT,
    date_of_birth       DATE NOT NULL,
    nationality         TEXT DEFAULT 'Kenyan',
    id_number           TEXT,
    kra_pin             TEXT,
    email               TEXT,
    phone               TEXT,
    postal_address      TEXT,
    postal_code         TEXT,
    town                TEXT,
    marital_status      TEXT,
    spouse_name         TEXT,
    next_of_kin         TEXT,
    next_of_kin_phone   TEXT,
    bank_name           TEXT,
    bank_branch         TEXT,
    bank_account        TEXT,
    payroll_no          TEXT,
    designation         TEXT,
    department          TEXT,
    sponsor_id          UUID REFERENCES sponsors(id),
    date_first_appt     DATE,
    date_joined_scheme  DATE NOT NULL,
    expected_retirement DATE,
    basic_salary        BIGINT DEFAULT 0,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason    TEXT,
    submitted_by        UUID,
    reviewed_by         UUID,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ
);
CREATE INDEX idx_pending_reg_status ON pending_member_registrations(status);
CREATE INDEX idx_pending_reg_scheme ON pending_member_registrations(scheme_id);

-- ============================================================
-- PENDING MEMBER CHANGES (Maker-Checker for Updates)
-- ============================================================
CREATE TABLE pending_member_changes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    change_type         TEXT NOT NULL CHECK (change_type IN (
        'update_personal', 'update_contact', 'update_bank',
        'update_beneficiary', 'update_nok', 'update_employment'
    )),
    before_values       JSONB NOT NULL,
    after_values        JSONB NOT NULL,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason    TEXT,
    submitted_by        UUID,
    reviewed_by         UUID,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ
);
CREATE INDEX idx_pending_changes_member ON pending_member_changes(member_id);
CREATE INDEX idx_pending_changes_status ON pending_member_changes(status);

-- ============================================================
-- EDI PROCESSING (Employer Data Interchange)
-- ============================================================
CREATE TABLE edi_files (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sponsor_id          UUID NOT NULL REFERENCES sponsors(id),
    file_name           TEXT NOT NULL,
    file_hash           TEXT NOT NULL,
    record_count        INT DEFAULT 0,
    total_employee_contributions BIGINT DEFAULT 0,
    total_employer_contributions BIGINT DEFAULT 0,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'matched', 'partial_mismatch', 'mismatch', 'reconciled')),
    processed_at        TIMESTAMPTZ,
    error_message       TEXT,
    uploaded_by         UUID,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_edi_sponsor ON edi_files(sponsor_id);
CREATE INDEX idx_edi_status ON edi_files(status);

CREATE TABLE edi_records (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    edi_file_id         UUID NOT NULL REFERENCES edi_files(id),
    record_type         TEXT NOT NULL,
    record_data         JSONB NOT NULL,
    member_id           UUID REFERENCES members(id),
    matched             BOOLEAN DEFAULT false,
    match_confidence    NUMERIC(5,2) DEFAULT 0,
    discrepancy_notes   TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_edi_records_file ON edi_records(edi_file_id);
CREATE INDEX idx_edi_records_member ON edi_records(member_id);

-- ============================================================
-- UNREGISTERED CONTRIBUTIONS (Tracked but not in system)
-- ============================================================
CREATE TABLE unregistered_contributions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    sponsor_id          UUID REFERENCES sponsors(id),
    sponsor_name        TEXT,
    member_name         TEXT NOT NULL,
    member_id_no        TEXT,
    period              DATE NOT NULL,
    employee_amount     BIGINT NOT NULL,
    employer_amount     BIGINT NOT NULL,
    total_amount        BIGINT NOT NULL,
    payment_ref         TEXT,
    payment_date        DATE,
    tracking_status     TEXT NOT NULL DEFAULT 'pending' CHECK (tracking_status IN ('pending', 'contacted', 'registered', 'untraceable')),
    first_contact_date  DATE,
    last_contact_date   DATE,
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_unreg_scheme ON unregistered_contributions(scheme_id);
CREATE INDEX idx_unreg_sponsor ON unregistered_contributions(sponsor_id);
CREATE INDEX idx_unreg_period ON unregistered_contributions(period);
CREATE INDEX idx_unreg_status ON unregistered_contributions(tracking_status);

-- ============================================================
-- PENDING CLAIMS (Maker-Checker for Claims)
-- ============================================================
CREATE TABLE pending_claims (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    claim_type          TEXT NOT NULL CHECK (claim_type IN (
        'normal_retirement', 'early_retirement', 'late_retirement',
        'ill_health_retirement', 'death_in_service', 'leaving_service',
        'refund', 'medical_claim', 'ex_gratia'
    )),
    claim_form_no       TEXT,
    date_of_claim       DATE NOT NULL,
    date_of_leaving     DATE,
    leaving_reason      TEXT,
    estimated_amount    BIGINT,
    supporting_docs     JSONB,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'requires_info')),
    rejection_reason    TEXT,
    submitted_by        UUID,
    reviewed_by         UUID,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ
);
CREATE INDEX idx_pending_claims_member ON pending_claims(member_id);
CREATE INDEX idx_pending_claims_scheme ON pending_claims(scheme_id);
CREATE INDEX idx_pending_claims_type ON pending_claims(claim_type);
CREATE INDEX idx_pending_claims_status ON pending_claims(status);

-- ============================================================
-- DEATH IN SERVICE
-- ============================================================
CREATE TABLE death_in_service (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    date_of_death       DATE NOT NULL,
    cause_of_death      TEXT,
    death_certificate_no TEXT,
    employer_notified    BOOLEAN DEFAULT false,
    employer_notify_date DATE,
    claim_id            UUID REFERENCES claims(id),
    claim_status        TEXT CHECK (claim_status IN ('pending', 'submitted', 'approved', 'paid')),
    total_benefit       BIGINT,
    breakdown           JSONB,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_dis_member ON death_in_service(member_id);
CREATE INDEX idx_dis_scheme ON death_in_service(scheme_id);

-- ============================================================
-- BENEFICIARY DRAWDOWNS
-- ============================================================
CREATE TABLE beneficiary_drawdowns (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    death_in_service_id UUID NOT NULL REFERENCES death_in_service(id),
    beneficiary_id      UUID NOT NULL REFERENCES beneficiaries(id),
    member_id           UUID NOT NULL REFERENCES members(id),
    allocation_amount   BIGINT NOT NULL,
    drawdown_type      TEXT CHECK (drawdown_type IN ('lump_sum', 'installment', 'annuity')),
    installment_amount  BIGINT,
    installment_periods  INT,
    start_date          DATE,
    next_payment_date   DATE,
    total_paid          BIGINT DEFAULT 0,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'completed', 'suspended')),
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_drawdown_dis ON beneficiary_drawdowns(death_in_service_id);
CREATE INDEX idx_drawdown_beneficiary ON beneficiary_drawdowns(beneficiary_id);
CREATE INDEX idx_drawdown_status ON beneficiary_drawdowns(status);

-- ============================================================
-- DIGITAL SIGNATURES
-- ============================================================
CREATE TABLE digital_signatures (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    signer_id           UUID NOT NULL,
    signer_type         TEXT NOT NULL CHECK (signer_type IN ('user', 'member', 'trustee')),
    document_type       TEXT NOT NULL,
    document_hash       TEXT NOT NULL,
    signature           TEXT NOT NULL,
    public_key          TEXT,
    merkle_root         TEXT,
    merkle_path         JSONB,
    ipfs_cid            TEXT,
    blockchain_tx       TEXT,
    signed_at           TIMESTAMPTZ DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    revoked             BOOLEAN DEFAULT false,
    revoked_at          TIMESTAMPTZ,
    revoked_by         UUID
);
CREATE INDEX idx_signatures_scheme ON digital_signatures(scheme_id);
CREATE INDEX idx_signatures_hash ON digital_signatures(document_hash);
CREATE INDEX idx_signatures_signer ON digital_signatures(signer_id);
CREATE INDEX idx_signatures_merkle ON digital_signatures(merkle_root);

CREATE TABLE signature_configs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    config_type         TEXT NOT NULL CHECK (config_type IN ('single', 'multi_sig', 'threshold')),
    required_signers    INT DEFAULT 1,
    signer_roles        JSONB,
    auto_expire_hours   INT DEFAULT 168,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_sig_config_scheme ON signature_configs(scheme_id);

CREATE TABLE merkle_roots (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    root_hash           TEXT NOT NULL,
    tree_type           TEXT NOT NULL CHECK (tree_type IN ('daily_transactions', 'monthly_contributions', 'claim_payouts', 'audit_snapshot')),
    period_start        DATE,
    period_end          DATE,
    leaf_count          INT,
    tree_depth          INT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_merkle_scheme ON merkle_roots(scheme_id);
CREATE INDEX idx_merkle_root ON merkle_roots(root_hash);

-- ============================================================
-- ENHANCED AUDIT LOG (with chain hashing)
-- ============================================================
CREATE TABLE audit_log_chain (
    id                  BIGSERIAL PRIMARY KEY,
    scheme_id           UUID REFERENCES schemes(id),
    user_id             UUID REFERENCES system_users(id),
    actor_type          TEXT CHECK (actor_type IN ('user', 'system', 'api', 'scheduled')),
    actor_ip            TEXT,
    actor_device        TEXT,
    actor_location      TEXT,
    action              TEXT NOT NULL,
    entity_type         TEXT,
    entity_id           UUID,
    old_values          JSONB,
    new_values          JSONB,
    previous_hash       TEXT NOT NULL DEFAULT '',
    record_hash         TEXT NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_chain_scheme ON audit_log_chain(scheme_id);
CREATE INDEX idx_audit_chain_user ON audit_log_chain(user_id);
CREATE INDEX idx_audit_chain_entity ON audit_log_chain(entity_type, entity_id);
CREATE INDEX idx_audit_chain_time ON audit_log_chain(created_at);

-- ============================================================
-- TAX EXEMPTIONS (KRA Integration)
-- ============================================================
CREATE TABLE tax_exemptions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    kra_certificate_no  TEXT NOT NULL,
    exemption_type      TEXT CHECK (exemption_type IN ('temporary', 'permanent', 'medical')),
    granted_date        DATE NOT NULL,
    expiry_date         DATE,
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'revoked', 'pending')),
    kra_approval_ref    TEXT,
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_tax_exemp_member ON tax_exemptions(member_id);
CREATE INDEX idx_tax_exemp_status ON tax_exemptions(status);
CREATE INDEX idx_tax_exemp_expiry ON tax_exemptions(expiry_date);

-- ============================================================
-- ANNUAL BENEFIT STATEMENTS
-- ============================================================
CREATE TABLE annual_statements (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    member_id           UUID NOT NULL REFERENCES members(id),
    scheme_id           UUID NOT NULL REFERENCES schemes(id),
    year                INT NOT NULL,
    opening_balance     BIGINT DEFAULT 0,
    employee_contribs   BIGINT DEFAULT 0,
    employer_contribs   BIGINT DEFAULT 0,
    interest_earned     BIGINT DEFAULT 0,
    withdrawals         BIGINT DEFAULT 0,
    closing_balance     BIGINT DEFAULT 0,
    fund_value          BIGINT DEFAULT 0,
    declared_rate       NUMERIC(5,4),
    statement_hash      TEXT,
    issued_at           TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_stmt_member ON annual_statements(member_id);
CREATE INDEX idx_stmt_year ON annual_statements(year);
CREATE UNIQUE INDEX idx_stmt_member_year ON annual_statements(member_id, year);
