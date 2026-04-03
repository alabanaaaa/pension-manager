package domain

import (
	"testing"
	"time"
)

func TestElectionValidate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	// Valid election
	e := &Election{
		Title:         "Trustee Election 2026",
		SchemeID:      "scheme-001",
		Type:          ElectionTrustee,
		MaxCandidates: 3,
		StartDate:     now,
		EndDate:       future,
	}
	if err := e.Validate(); err != nil {
		t.Errorf("Expected valid election, got error: %v", err)
	}

	// Missing title
	e.Title = ""
	if err := e.Validate(); err == nil {
		t.Error("Expected error for missing title")
	}
	e.Title = "Trustee Election 2026"

	// Missing scheme
	e.SchemeID = ""
	if err := e.Validate(); err == nil {
		t.Error("Expected error for missing scheme")
	}
	e.SchemeID = "scheme-001"

	// End date before start date
	e.EndDate = now.Add(-24 * time.Hour)
	if err := e.Validate(); err == nil {
		t.Error("Expected error for end date before start date")
	}

	// Default max candidates
	e.MaxCandidates = 0
	e.EndDate = future
	if err := e.Validate(); err != nil {
		t.Errorf("Expected valid election with default max candidates, got: %v", err)
	}
	if e.MaxCandidates != 3 {
		t.Errorf("Expected default max candidates 3, got: %d", e.MaxCandidates)
	}
}

func TestCandidateValidate(t *testing.T) {
	// Valid candidate
	c := &Candidate{
		Name:       "John Doe",
		ElectionID: "election-001",
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Expected valid candidate, got error: %v", err)
	}

	// Missing name
	c.Name = ""
	if err := c.Validate(); err == nil {
		t.Error("Expected error for missing name")
	}
	c.Name = "John Doe"

	// Missing election
	c.ElectionID = ""
	if err := c.Validate(); err == nil {
		t.Error("Expected error for missing election")
	}
}

func TestVoteValidate(t *testing.T) {
	// Valid vote
	v := &Vote{
		ElectionID:   "election-001",
		MemberID:     "member-001",
		CandidateID:  "candidate-001",
		VotingMethod: "web",
	}
	if err := v.Validate(); err != nil {
		t.Errorf("Expected valid vote, got error: %v", err)
	}

	// Missing election
	v.ElectionID = ""
	if err := v.Validate(); err == nil {
		t.Error("Expected error for missing election")
	}
	v.ElectionID = "election-001"

	// Missing member
	v.MemberID = ""
	if err := v.Validate(); err == nil {
		t.Error("Expected error for missing member")
	}
	v.MemberID = "member-001"

	// Missing candidate
	v.CandidateID = ""
	if err := v.Validate(); err == nil {
		t.Error("Expected error for missing candidate")
	}
	v.CandidateID = "candidate-001"

	// Missing method
	v.VotingMethod = ""
	if err := v.Validate(); err == nil {
		t.Error("Expected error for missing voting method")
	}
}
