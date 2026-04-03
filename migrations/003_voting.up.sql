-- ============================================================
-- 003: Online Voting System
-- ============================================================

-- Elections
CREATE TABLE IF NOT EXISTS elections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    title           TEXT NOT NULL,
    description     TEXT,
    type            TEXT NOT NULL CHECK (type IN ('trustee', 'agm', 'scheme_agenda', 'other')),
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'open', 'closed', 'archived')),
    max_candidates  INT NOT NULL DEFAULT 3,
    start_date      TIMESTAMPTZ NOT NULL,
    end_date        TIMESTAMPTZ NOT NULL,
    created_by      UUID NOT NULL,
    total_voters    INT NOT NULL DEFAULT 0,
    total_votes     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_elections_scheme ON elections(scheme_id);
CREATE INDEX idx_elections_status ON elections(status);
CREATE INDEX idx_elections_dates ON elections(start_date, end_date);

-- Candidates
CREATE TABLE IF NOT EXISTS candidates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    position        TEXT,
    manifesto       TEXT,
    photo_url       TEXT,
    polling_station TEXT,
    scheme_type     TEXT CHECK (scheme_type IN ('db', 'dc', 'both')),
    vote_count      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_candidates_election ON candidates(election_id);

-- Voter Register (eligible voters per election)
CREATE TABLE IF NOT EXISTS voter_register (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    member_id       UUID NOT NULL REFERENCES members(id),
    check_no        TEXT,
    added_by        UUID NOT NULL,
    added_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(election_id, member_id)
);
CREATE INDEX idx_voter_register_election ON voter_register(election_id);
CREATE INDEX idx_voter_register_member ON voter_register(member_id);

-- Votes
CREATE TABLE IF NOT EXISTS votes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id),
    member_id       UUID NOT NULL REFERENCES members(id),
    candidate_id    UUID NOT NULL REFERENCES candidates(id),
    voting_method   TEXT NOT NULL CHECK (voting_method IN ('web', 'ussd', 'url')),
    voted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mobile_number   TEXT,
    gps_latitude    DOUBLE PRECISION,
    gps_longitude   DOUBLE PRECISION,
    ip_address      TEXT,
    user_agent      TEXT
);
CREATE INDEX idx_votes_election ON votes(election_id);
CREATE INDEX idx_votes_member ON votes(member_id, election_id);
CREATE INDEX idx_votes_candidate ON votes(candidate_id);
CREATE INDEX idx_votes_method ON votes(voting_method);

-- Prevent duplicate votes (member can only vote once per election per method)
CREATE UNIQUE INDEX idx_votes_unique_member_election_method ON votes(member_id, election_id, voting_method);
