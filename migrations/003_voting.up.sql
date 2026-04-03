-- ============================================================
-- 003: Online Voting System Enhancements
-- ============================================================

-- Elections table already exists in 001 with different columns
-- Add missing columns
ALTER TABLE elections ADD COLUMN IF NOT EXISTS type TEXT;
UPDATE elections SET type = election_type WHERE type IS NULL AND election_type IS NOT NULL;
ALTER TABLE elections ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE elections ADD COLUMN IF NOT EXISTS total_voters INT NOT NULL DEFAULT 0;
ALTER TABLE elections ADD COLUMN IF NOT EXISTS total_votes INT NOT NULL DEFAULT 0;
ALTER TABLE elections ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
-- Rename start_at/end_at to start_date/end_date if they exist
ALTER TABLE elections RENAME COLUMN start_at TO start_date;
ALTER TABLE elections RENAME COLUMN end_at TO end_date;
-- Add status 'archived' to check constraint if not exists
-- (PostgreSQL doesn't support adding to CHECK easily, so we'll handle in code)
CREATE INDEX IF NOT EXISTS idx_elections_dates ON elections(start_date, end_date);

-- Candidates table already exists in 001
-- Add missing columns
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS polling_station TEXT;
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS scheme_type TEXT CHECK (scheme_type IN ('db', 'dc', 'both'));
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS vote_count INT NOT NULL DEFAULT 0;
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS photo_url TEXT;
CREATE INDEX IF NOT EXISTS idx_candidates_election ON candidates(election_id);

-- Voter Register (new table)
CREATE TABLE IF NOT EXISTS voter_register (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    member_id       UUID NOT NULL REFERENCES members(id),
    check_no        TEXT,
    added_by        UUID NOT NULL,
    added_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(election_id, member_id)
);
CREATE INDEX IF NOT EXISTS idx_voter_register_election ON voter_register(election_id);
CREATE INDEX IF NOT EXISTS idx_voter_register_member ON voter_register(member_id);

-- Votes table already exists in 001 with 'channel' column
-- Add voting_method column if missing
ALTER TABLE votes ADD COLUMN IF NOT EXISTS voting_method TEXT;
UPDATE votes SET voting_method = channel WHERE voting_method IS NULL AND channel IS NOT NULL;
-- Make voting_method NOT NULL after migration
ALTER TABLE votes ALTER COLUMN voting_method SET NOT NULL;

-- Add mobile_number column if missing
ALTER TABLE votes ADD COLUMN IF NOT EXISTS mobile_number TEXT;

-- Add GPS latitude/longitude columns if missing
ALTER TABLE votes ADD COLUMN IF NOT EXISTS gps_latitude DOUBLE PRECISION;
ALTER TABLE votes ADD COLUMN IF NOT EXISTS gps_longitude DOUBLE PRECISION;

-- Add indexes if missing
CREATE INDEX IF NOT EXISTS idx_votes_election ON votes(election_id);
CREATE INDEX IF NOT EXISTS idx_votes_member ON votes(member_id, election_id);
CREATE INDEX IF NOT EXISTS idx_votes_candidate ON votes(candidate_id);
CREATE INDEX IF NOT EXISTS idx_votes_method ON votes(voting_method);

-- Prevent duplicate votes (member can only vote once per election per method)
DROP INDEX IF EXISTS idx_votes_unique_member_election_method;
CREATE UNIQUE INDEX IF NOT EXISTS idx_votes_unique_member_election_method ON votes(member_id, election_id, voting_method);
