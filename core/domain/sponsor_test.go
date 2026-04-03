package domain

import (
	"testing"
	"time"
)

func TestSponsorValidate(t *testing.T) {
	s := &Sponsor{
		Name:     "Test Sponsor",
		SchemeID: "scheme-001",
		Code:     "SP001",
	}
	if err := s.Validate(); err != nil {
		t.Errorf("Expected valid sponsor, got error: %v", err)
	}

	s.Name = ""
	if err := s.Validate(); err == nil {
		t.Error("Expected error for missing name")
	}
	s.Name = "Test Sponsor"

	s.SchemeID = ""
	if err := s.Validate(); err == nil {
		t.Error("Expected error for missing scheme")
	}
	s.SchemeID = "scheme-001"

	s.Code = ""
	if err := s.Validate(); err == nil {
		t.Error("Expected error for missing code")
	}
}

func TestContributionScheduleValidate(t *testing.T) {
	cs := &ContributionSchedule{
		SponsorID:      "sponsor-001",
		Period:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TotalEmployees: 100,
		TotalAmount:    50000000,
	}
	if err := cs.Validate(); err != nil {
		t.Errorf("Expected valid schedule, got error: %v", err)
	}

	cs.SponsorID = ""
	if err := cs.Validate(); err == nil {
		t.Error("Expected error for missing sponsor")
	}
	cs.SponsorID = "sponsor-001"

	cs.Period = time.Time{}
	if err := cs.Validate(); err == nil {
		t.Error("Expected error for missing period")
	}
	cs.Period = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cs.TotalEmployees = -1
	if err := cs.Validate(); err == nil {
		t.Error("Expected error for negative employees")
	}
	cs.TotalEmployees = 100

	cs.TotalAmount = -1000
	if err := cs.Validate(); err == nil {
		t.Error("Expected error for negative amount")
	}
}
