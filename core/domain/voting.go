package domain

import (
	"time"
)

// ElectionStatus represents the current state of an election
type ElectionStatus string

const (
	ElectionDraft    ElectionStatus = "draft"
	ElectionOpen     ElectionStatus = "open"
	ElectionClosed   ElectionStatus = "closed"
	ElectionArchived ElectionStatus = "archived"
)

// ElectionType represents the type of election
type ElectionType string

const (
	ElectionTrustee      ElectionType = "trustee"
	ElectionAGM          ElectionType = "agm"
	ElectionSchemeAgenda ElectionType = "scheme_agenda"
	ElectionOther        ElectionType = "other"
)

// Election represents a voting election
type Election struct {
	ID            string         `json:"id"`
	SchemeID      string         `json:"scheme_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description,omitempty"`
	Type          ElectionType   `json:"type"`
	Status        ElectionStatus `json:"status"`
	MaxCandidates int            `json:"max_candidates"` // Max votes per member (default 3)
	StartDate     time.Time      `json:"start_date"`
	EndDate       time.Time      `json:"end_date"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	TotalVoters   int            `json:"total_voters"`
	TotalVotes    int            `json:"total_votes"`
}

// Validate checks if an Election is valid
func (e *Election) Validate() error {
	if e.Title == "" {
		return NewValidationError("title", "election title is required")
	}
	if e.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if e.StartDate.IsZero() {
		return NewValidationError("start_date", "start date is required")
	}
	if e.EndDate.IsZero() {
		return NewValidationError("end_date", "end date is required")
	}
	if e.EndDate.Before(e.StartDate) {
		return NewValidationError("end_date", "end date must be after start date")
	}
	if e.MaxCandidates <= 0 {
		e.MaxCandidates = 3 // Default to 3
	}
	return nil
}

// Candidate represents a candidate in an election
type Candidate struct {
	ID             string    `json:"id"`
	ElectionID     string    `json:"election_id"`
	Name           string    `json:"name"`
	Position       string    `json:"position,omitempty"`
	Manifesto      string    `json:"manifesto,omitempty"`
	PhotoURL       string    `json:"photo_url,omitempty"`
	PollingStation string    `json:"polling_station,omitempty"`
	SchemeType     string    `json:"scheme_type,omitempty"` // db, dc, or both
	VoteCount      int       `json:"vote_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validate checks if a Candidate is valid
func (c *Candidate) Validate() error {
	if c.Name == "" {
		return NewValidationError("name", "candidate name is required")
	}
	if c.ElectionID == "" {
		return NewValidationError("election_id", "election is required")
	}
	return nil
}

// VoterRegister represents the list of eligible voters for an election
type VoterRegister struct {
	ID         string    `json:"id"`
	ElectionID string    `json:"election_id"`
	MemberID   string    `json:"member_id"`
	CheckNo    string    `json:"check_no,omitempty"` // For web portal sign-in
	AddedBy    string    `json:"added_by"`
	AddedAt    time.Time `json:"added_at"`
}

// Vote represents a single vote cast by a member
type Vote struct {
	ID           string    `json:"id"`
	ElectionID   string    `json:"election_id"`
	MemberID     string    `json:"member_id"`
	CandidateID  string    `json:"candidate_id"`
	VotingMethod string    `json:"voting_method"` // web, ussd, url
	VotedAt      time.Time `json:"voted_at"`
	MobileNumber string    `json:"mobile_number,omitempty"`
	GPSLatitude  float64   `json:"gps_latitude,omitempty"`
	GPSLongitude float64   `json:"gps_longitude,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
}

// Validate checks if a Vote is valid
func (v *Vote) Validate() error {
	if v.ElectionID == "" {
		return NewValidationError("election_id", "election is required")
	}
	if v.MemberID == "" {
		return NewValidationError("member_id", "member is required")
	}
	if v.CandidateID == "" {
		return NewValidationError("candidate_id", "candidate is required")
	}
	if v.VotingMethod == "" {
		return NewValidationError("voting_method", "voting method is required")
	}
	return nil
}

// ElectionResult holds the results for a single candidate
type ElectionResult struct {
	CandidateID    string  `json:"candidate_id"`
	CandidateName  string  `json:"candidate_name"`
	Position       string  `json:"position,omitempty"`
	VoteCount      int     `json:"vote_count"`
	VotePercentage float64 `json:"vote_percentage"`
	PollingStation string  `json:"polling_station,omitempty"`
	SchemeType     string  `json:"scheme_type,omitempty"`
}

// ElectionSummary holds the overall election results
type ElectionSummary struct {
	ElectionID        string           `json:"election_id"`
	ElectionTitle     string           `json:"election_title"`
	ElectionType      ElectionType     `json:"election_type"`
	Status            ElectionStatus   `json:"status"`
	TotalVoters       int              `json:"total_voters"`
	TotalVotesCast    int              `json:"total_votes_cast"`
	TurnoutPercentage float64          `json:"turnout_percentage"`
	StartDate         time.Time        `json:"start_date"`
	EndDate           time.Time        `json:"end_date"`
	Results           []ElectionResult `json:"results"`
}

// VotedMemberReport holds info about a member who voted
type VotedMemberReport struct {
	MemberID     string    `json:"member_id"`
	MemberNo     string    `json:"member_no"`
	FullName     string    `json:"full_name"`
	VotingMethod string    `json:"voting_method"`
	VotedAt      time.Time `json:"voted_at"`
	MobileNumber string    `json:"mobile_number,omitempty"`
	GPSLocation  string    `json:"gps_location,omitempty"`
}
