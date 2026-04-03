package voting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pension-manager/core/domain"
	"pension-manager/internal/db"
)

// Service manages online voting operations
type Service struct {
	db *db.DB
}

// NewService creates a new voting service
func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// CreateElection creates a new election
func (s *Service) CreateElection(ctx context.Context, election *domain.Election) error {
	if err := election.Validate(); err != nil {
		return err
	}

	now := time.Now()
	election.CreatedAt = now
	election.UpdatedAt = now
	if election.Status == "" {
		election.Status = domain.ElectionDraft
	}

	query := `
		INSERT INTO elections (id, scheme_id, title, description, type, status, max_candidates,
		                       start_date, end_date, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.db.ExecContext(ctx, query,
		election.ID, election.SchemeID, election.Title, election.Description,
		election.Type, election.Status, election.MaxCandidates,
		election.StartDate, election.EndDate, election.CreatedBy,
		election.CreatedAt, election.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create election: %w", err)
	}
	return nil
}

// GetElection retrieves an election by ID
func (s *Service) GetElection(ctx context.Context, electionID string) (*domain.Election, error) {
	query := `
		SELECT id, scheme_id, title, description, type, status, max_candidates,
		       start_date, end_date, created_by, total_voters, total_votes, created_at, updated_at
		FROM elections WHERE id = $1
	`
	e := &domain.Election{}
	err := s.db.QueryRowContext(ctx, query, electionID).Scan(
		&e.ID, &e.SchemeID, &e.Title, &e.Description, &e.Type, &e.Status, &e.MaxCandidates,
		&e.StartDate, &e.EndDate, &e.CreatedBy, &e.TotalVoters, &e.TotalVotes,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get election: %w", err)
	}
	return e, nil
}

// ListElections returns all elections for a scheme
func (s *Service) ListElections(ctx context.Context, schemeID string) ([]*domain.Election, error) {
	query := `
		SELECT id, scheme_id, title, description, type, status, max_candidates,
		       start_date, end_date, created_by, total_voters, total_votes, created_at, updated_at
		FROM elections WHERE scheme_id = $1 ORDER BY start_date DESC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("list elections: %w", err)
	}
	defer rows.Close()

	var elections []*domain.Election
	for rows.Next() {
		e := &domain.Election{}
		if err := rows.Scan(
			&e.ID, &e.SchemeID, &e.Title, &e.Description, &e.Type, &e.Status, &e.MaxCandidates,
			&e.StartDate, &e.EndDate, &e.CreatedBy, &e.TotalVoters, &e.TotalVotes,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan election: %w", err)
		}
		elections = append(elections, e)
	}
	return elections, rows.Err()
}

// UpdateElectionStatus updates an election's status
func (s *Service) UpdateElectionStatus(ctx context.Context, electionID string, status domain.ElectionStatus) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE elections SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, electionID)
	if err != nil {
		return fmt.Errorf("update election status: %w", err)
	}
	return nil
}

// AddCandidate adds a candidate to an election
func (s *Service) AddCandidate(ctx context.Context, candidate *domain.Candidate) error {
	if err := candidate.Validate(); err != nil {
		return err
	}

	now := time.Now()
	candidate.CreatedAt = now
	candidate.UpdatedAt = now

	query := `
		INSERT INTO candidates (id, election_id, name, position, manifesto, photo_url,
		                        polling_station, scheme_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.ExecContext(ctx, query,
		candidate.ID, candidate.ElectionID, candidate.Name, candidate.Position,
		candidate.Manifesto, candidate.PhotoURL, candidate.PollingStation,
		candidate.SchemeType, candidate.CreatedAt, candidate.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("add candidate: %w", err)
	}
	return nil
}

// ListCandidates returns all candidates for an election
func (s *Service) ListCandidates(ctx context.Context, electionID string) ([]*domain.Candidate, error) {
	query := `
		SELECT id, election_id, name, position, manifesto, photo_url, polling_station,
		       scheme_type, vote_count, created_at, updated_at
		FROM candidates WHERE election_id = $1 ORDER BY vote_count DESC, name ASC
	`
	rows, err := s.db.QueryContext(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()

	var candidates []*domain.Candidate
	for rows.Next() {
		c := &domain.Candidate{}
		if err := rows.Scan(
			&c.ID, &c.ElectionID, &c.Name, &c.Position, &c.Manifesto, &c.PhotoURL,
			&c.PollingStation, &c.SchemeType, &c.VoteCount, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// AddVoter adds a member to the voter register
func (s *Service) AddVoter(ctx context.Context, electionID, memberID, checkNo, addedBy string) error {
	query := `
		INSERT INTO voter_register (id, election_id, member_id, check_no, added_by, added_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW())
		ON CONFLICT (election_id, member_id) DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, electionID, memberID, checkNo, addedBy)
	if err != nil {
		return fmt.Errorf("add voter: %w", err)
	}

	// Update total voters count
	_, err = s.db.ExecContext(ctx, `
		UPDATE elections SET total_voters = (SELECT COUNT(*) FROM voter_register WHERE election_id = $1)
		WHERE id = $1
	`, electionID)
	return err
}

// BulkAddVoters adds multiple members to the voter register
func (s *Service) BulkAddVoters(ctx context.Context, electionID string, memberIDs []string, addedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		for _, memberID := range memberIDs {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO voter_register (id, election_id, member_id, added_by, added_at)
				VALUES (uuid_generate_v4(), $1, $2, $3, NOW())
				ON CONFLICT (election_id, member_id) DO NOTHING
			`, electionID, memberID, addedBy)
			if err != nil {
				return fmt.Errorf("add voter %s: %w", memberID, err)
			}
		}

		_, err := tx.ExecContext(ctx, `
			UPDATE elections SET total_voters = (SELECT COUNT(*) FROM voter_register WHERE election_id = $1)
			WHERE id = $1
		`, electionID)
		return err
	})
}

// IsEligibleToVote checks if a member is eligible to vote in an election
func (s *Service) IsEligibleToVote(ctx context.Context, electionID, memberID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM voter_register WHERE election_id = $1 AND member_id = $2)
	`, electionID, memberID).Scan(&exists)
	return exists, err
}

// HasVoted checks if a member has already voted in an election
func (s *Service) HasVoted(ctx context.Context, electionID, memberID, method string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM votes WHERE election_id = $1 AND member_id = $2 AND voting_method = $3)
	`, electionID, memberID, method).Scan(&exists)
	return exists, err
}

// HasVotedAny checks if a member has voted by ANY method (USSD or web/URL are mutually exclusive)
func (s *Service) HasVotedAny(ctx context.Context, electionID, memberID string) (bool, string, error) {
	var method string
	err := s.db.QueryRowContext(ctx, `
		SELECT voting_method FROM votes WHERE election_id = $1 AND member_id = $2 LIMIT 1
	`, electionID, memberID).Scan(&method)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, method, nil
}

// GetVoteCountForMember returns how many candidates a member has voted for
func (s *Service) GetVoteCountForMember(ctx context.Context, electionID, memberID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM votes WHERE election_id = $1 AND member_id = $2
	`, electionID, memberID).Scan(&count)
	return count, err
}

// CastVote records a vote
func (s *Service) CastVote(ctx context.Context, vote *domain.Vote) error {
	if err := vote.Validate(); err != nil {
		return err
	}

	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		// Check election is open
		var status domain.ElectionStatus
		var maxCandidates int
		err := tx.QueryRowContext(ctx, `
			SELECT status, max_candidates FROM elections WHERE id = $1
		`, vote.ElectionID).Scan(&status, &maxCandidates)
		if err != nil {
			return fmt.Errorf("get election: %w", err)
		}
		if status != domain.ElectionOpen {
			return fmt.Errorf("election is not open (status: %s)", status)
		}

		// Check voter is registered
		var registered bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM voter_register WHERE election_id = $1 AND member_id = $2)
		`, vote.ElectionID, vote.MemberID).Scan(&registered)
		if err != nil {
			return fmt.Errorf("check voter registration: %w", err)
		}
		if !registered {
			return errors.New("member is not registered to vote in this election")
		}

		// Check hasn't voted by any method (USSD and web/URL are mutually exclusive)
		hasVoted, votedMethod, err := s.HasVotedAny(ctx, vote.ElectionID, vote.MemberID)
		if err != nil {
			return fmt.Errorf("check if voted: %w", err)
		}
		if hasVoted {
			return fmt.Errorf("member has already voted via %s", votedMethod)
		}

		// Check max candidates limit
		voteCount, err := s.GetVoteCountForMember(ctx, vote.ElectionID, vote.MemberID)
		if err != nil {
			return fmt.Errorf("get vote count: %w", err)
		}
		if voteCount >= maxCandidates {
			return fmt.Errorf("maximum %d votes allowed per member", maxCandidates)
		}

		// Record vote
		vote.VotedAt = time.Now()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO votes (id, election_id, member_id, candidate_id, voting_method, voted_at,
			                   mobile_number, gps_latitude, gps_longitude, ip_address, user_agent)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, vote.ElectionID, vote.MemberID, vote.CandidateID, vote.VotingMethod,
			vote.VotedAt, vote.MobileNumber, vote.GPSLatitude, vote.GPSLongitude,
			vote.IPAddress, vote.UserAgent)
		if err != nil {
			return fmt.Errorf("record vote: %w", err)
		}

		// Update candidate vote count
		_, err = tx.ExecContext(ctx, `
			UPDATE candidates SET vote_count = vote_count + 1, updated_at = NOW() WHERE id = $1
		`, vote.CandidateID)
		if err != nil {
			return fmt.Errorf("update candidate vote count: %w", err)
		}

		// Update election total votes
		_, err = tx.ExecContext(ctx, `
			UPDATE elections SET total_votes = total_votes + 1, updated_at = NOW() WHERE id = $1
		`, vote.ElectionID)
		return err
	})
}

// GetElectionResults returns the full election results
func (s *Service) GetElectionResults(ctx context.Context, electionID string) (*domain.ElectionSummary, error) {
	var summary domain.ElectionSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT e.id, e.title, e.type, e.status, e.total_voters, e.total_votes, e.start_date, e.end_date
		FROM elections e WHERE e.id = $1
	`, electionID).Scan(
		&summary.ElectionID, &summary.ElectionTitle, &summary.ElectionType, &summary.Status,
		&summary.TotalVoters, &summary.TotalVotesCast, &summary.StartDate, &summary.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("get election: %w", err)
	}

	if summary.TotalVoters > 0 {
		summary.TurnoutPercentage = float64(summary.TotalVotesCast) / float64(summary.TotalVoters) * 100
	}

	// Get candidate results
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, position, vote_count, polling_station, scheme_type
		FROM candidates WHERE election_id = $1 ORDER BY vote_count DESC
	`, electionID)
	if err != nil {
		return nil, fmt.Errorf("query candidates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r domain.ElectionResult
		if err := rows.Scan(&r.CandidateID, &r.CandidateName, &r.Position, &r.VoteCount,
			&r.PollingStation, &r.SchemeType); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		if summary.TotalVotesCast > 0 {
			r.VotePercentage = float64(r.VoteCount) / float64(summary.TotalVotesCast) * 100
		}
		summary.Results = append(summary.Results, r)
	}

	return &summary, nil
}

// GetResultsByPollingStation returns results filtered by polling station
func (s *Service) GetResultsByPollingStation(ctx context.Context, electionID, pollingStation string) ([]domain.ElectionResult, error) {
	query := `
		SELECT c.id, c.name, c.position, c.vote_count, c.polling_station, c.scheme_type
		FROM candidates c
		WHERE c.election_id = $1 AND c.polling_station = $2
		ORDER BY c.vote_count DESC
	`
	rows, err := s.db.QueryContext(ctx, query, electionID, pollingStation)
	if err != nil {
		return nil, fmt.Errorf("query results: %w", err)
	}
	defer rows.Close()

	var results []domain.ElectionResult
	for rows.Next() {
		var r domain.ElectionResult
		if err := rows.Scan(&r.CandidateID, &r.CandidateName, &r.Position, &r.VoteCount,
			&r.PollingStation, &r.SchemeType); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetResultsBySchemeType returns results filtered by scheme type (DB or DC)
func (s *Service) GetResultsBySchemeType(ctx context.Context, electionID, schemeType string) ([]domain.ElectionResult, error) {
	query := `
		SELECT c.id, c.name, c.position, c.vote_count, c.polling_station, c.scheme_type
		FROM candidates c
		WHERE c.election_id = $1 AND (c.scheme_type = $2 OR c.scheme_type = 'both')
		ORDER BY c.vote_count DESC
	`
	rows, err := s.db.QueryContext(ctx, query, electionID, schemeType)
	if err != nil {
		return nil, fmt.Errorf("query results: %w", err)
	}
	defer rows.Close()

	var results []domain.ElectionResult
	for rows.Next() {
		var r domain.ElectionResult
		if err := rows.Scan(&r.CandidateID, &r.CandidateName, &r.Position, &r.VoteCount,
			&r.PollingStation, &r.SchemeType); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetVotedMembersReport returns a list of members who have voted
func (s *Service) GetVotedMembersReport(ctx context.Context, electionID string) ([]domain.VotedMemberReport, error) {
	query := `
		SELECT v.member_id, m.member_no, m.first_name || ' ' || m.last_name as full_name,
		       v.voting_method, v.voted_at, v.mobile_number,
		       CASE WHEN v.gps_latitude IS NOT NULL THEN
		            format('%.4f, %.4f', v.gps_latitude, v.gps_longitude)
		            ELSE '' END as gps_location
		FROM votes v
		JOIN members m ON m.id = v.member_id
		WHERE v.election_id = $1
		ORDER BY v.voted_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("query voted members: %w", err)
	}
	defer rows.Close()

	var reports []domain.VotedMemberReport
	for rows.Next() {
		var r domain.VotedMemberReport
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &r.VotingMethod,
			&r.VotedAt, &r.MobileNumber, &r.GPSLocation); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// GetNotVotedMembers returns members who haven't voted yet
func (s *Service) GetNotVotedMembers(ctx context.Context, electionID string) ([]domain.VotedMemberReport, error) {
	query := `
		SELECT vr.member_id, m.member_no, m.first_name || ' ' || m.last_name as full_name,
		       '', NULL, m.phone, ''
		FROM voter_register vr
		JOIN members m ON m.id = vr.member_id
		WHERE vr.election_id = $1
		  AND NOT EXISTS (SELECT 1 FROM votes v WHERE v.election_id = $1 AND v.member_id = vr.member_id)
		ORDER BY m.last_name, m.first_name
	`
	rows, err := s.db.QueryContext(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("query not voted members: %w", err)
	}
	defer rows.Close()

	var reports []domain.VotedMemberReport
	for rows.Next() {
		var r domain.VotedMemberReport
		var votedAt sql.NullTime
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &r.VotingMethod,
			&votedAt, &r.MobileNumber, &r.GPSLocation); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// GetVotingStats returns real-time voting statistics
func (s *Service) GetVotingStats(ctx context.Context, electionID string) (*VotingStats, error) {
	stats := &VotingStats{}
	query := `
		SELECT
			e.total_voters,
			e.total_votes,
			COUNT(*) FILTER (WHERE v.voting_method = 'web') as web_votes,
			COUNT(*) FILTER (WHERE v.voting_method = 'ussd') as ussd_votes,
			COUNT(*) FILTER (WHERE v.voting_method = 'url') as url_votes,
			MAX(v.voted_at) as last_vote_at
		FROM elections e
		LEFT JOIN votes v ON v.election_id = e.id
		WHERE e.id = $1
		GROUP BY e.id, e.total_voters, e.total_votes
	`
	err := s.db.QueryRowContext(ctx, query, electionID).Scan(
		&stats.TotalVoters, &stats.TotalVotes, &stats.WebVotes, &stats.USSDVotes,
		&stats.URLVotes, &stats.LastVoteAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get voting stats: %w", err)
	}

	if stats.TotalVoters > 0 {
		stats.TurnoutPercentage = float64(stats.TotalVotes) / float64(stats.TotalVoters) * 100
	}

	return stats, nil
}

// VotingStats holds real-time voting statistics
type VotingStats struct {
	TotalVoters       int        `json:"total_voters"`
	TotalVotes        int        `json:"total_votes"`
	TurnoutPercentage float64    `json:"turnout_percentage"`
	WebVotes          int        `json:"web_votes"`
	USSDVotes         int        `json:"ussd_votes"`
	URLVotes          int        `json:"url_votes"`
	LastVoteAt        *time.Time `json:"last_vote_at,omitempty"`
}
