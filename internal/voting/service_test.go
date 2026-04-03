package voting

import (
	"testing"
	"time"

	"pension-manager/core/domain"
)

func TestElectionValidate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	// Valid election
	e := &domain.Election{
		Title:         "Trustee Election 2026",
		SchemeID:      "scheme-001",
		Type:          domain.ElectionTrustee,
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
	c := &domain.Candidate{
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
	v := &domain.Vote{
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

func TestElectionStatusConstants(t *testing.T) {
	if domain.ElectionDraft != "draft" {
		t.Errorf("Expected ElectionDraft = 'draft', got: %s", domain.ElectionDraft)
	}
	if domain.ElectionOpen != "open" {
		t.Errorf("Expected ElectionOpen = 'open', got: %s", domain.ElectionOpen)
	}
	if domain.ElectionClosed != "closed" {
		t.Errorf("Expected ElectionClosed = 'closed', got: %s", domain.ElectionClosed)
	}
	if domain.ElectionArchived != "archived" {
		t.Errorf("Expected ElectionArchived = 'archived', got: %s", domain.ElectionArchived)
	}
}

func TestElectionTypeConstants(t *testing.T) {
	if domain.ElectionTrustee != "trustee" {
		t.Errorf("Expected ElectionTrustee = 'trustee', got: %s", domain.ElectionTrustee)
	}
	if domain.ElectionAGM != "agm" {
		t.Errorf("Expected ElectionAGM = 'agm', got: %s", domain.ElectionAGM)
	}
	if domain.ElectionSchemeAgenda != "scheme_agenda" {
		t.Errorf("Expected ElectionSchemeAgenda = 'scheme_agenda', got: %s", domain.ElectionSchemeAgenda)
	}
	if domain.ElectionOther != "other" {
		t.Errorf("Expected ElectionOther = 'other', got: %s", domain.ElectionOther)
	}
}
