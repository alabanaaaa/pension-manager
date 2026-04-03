package sponsor

import (
	"testing"
	"time"

	"pension-manager/core/domain"
)

func TestSponsorStats_Struct(t *testing.T) {
	stats := &SponsorStats{
		TotalSchedules:     24,
		PostedSchedules:    20,
		PendingSchedules:   4,
		TotalContributions: 120000000,
		TotalEmployees:     2400,
	}

	if stats.TotalSchedules != 24 {
		t.Errorf("Expected total schedules 24, got: %d", stats.TotalSchedules)
	}
	if stats.PostedSchedules != 20 {
		t.Errorf("Expected posted schedules 20, got: %d", stats.PostedSchedules)
	}
	if stats.PendingSchedules != 4 {
		t.Errorf("Expected pending schedules 4, got: %d", stats.PendingSchedules)
	}
	if stats.TotalContributions != 120000000 {
		t.Errorf("Expected total contributions 120000000, got: %d", stats.TotalContributions)
	}
	if stats.TotalEmployees != 2400 {
		t.Errorf("Expected total employees 2400, got: %d", stats.TotalEmployees)
	}
}

func TestSponsorDomain_Validate(t *testing.T) {
	s := &domain.Sponsor{
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
}

func TestContributionScheduleDomain_Validate(t *testing.T) {
	cs := &domain.ContributionSchedule{
		SponsorID:      "sponsor-001",
		Period:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TotalEmployees: 100,
		TotalAmount:    50000000,
	}
	if err := cs.Validate(); err != nil {
		t.Errorf("Expected valid schedule, got error: %v", err)
	}

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
