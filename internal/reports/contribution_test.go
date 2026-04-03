package reports

import (
	"testing"
)

func TestContributionBreakdown_Struct(t *testing.T) {
	cb := ContributionBreakdown{
		SchemeID:      "scheme-001",
		SchemeName:    "Test Scheme",
		Period:        "2026-01",
		EmployeeCount: 100,
		EmployeeTotal: 5000000,
		EmployerTotal: 10000000,
		AVCTotal:      1000000,
		GrandTotal:    16000000,
	}

	if cb.SchemeID != "scheme-001" {
		t.Errorf("Expected scheme-001, got: %s", cb.SchemeID)
	}
	if cb.EmployeeCount != 100 {
		t.Errorf("Expected 100 employees, got: %d", cb.EmployeeCount)
	}
	if cb.GrandTotal != 16000000 {
		t.Errorf("Expected grand total 16000000, got: %d", cb.GrandTotal)
	}
}

func TestYTDContribution_Struct(t *testing.T) {
	ytd := YTDContribution{
		MemberID:    "mem-001",
		MemberNo:    "M001",
		FullName:    "John Doe",
		EmployeeYTD: 300000,
		EmployerYTD: 600000,
		AVCYTD:      50000,
		TotalYTD:    950000,
	}

	if ytd.MemberNo != "M001" {
		t.Errorf("Expected M001, got: %s", ytd.MemberNo)
	}
	if ytd.TotalYTD != 950000 {
		t.Errorf("Expected total YTD 950000, got: %d", ytd.TotalYTD)
	}
}

func TestCumulativeContribution_Struct(t *testing.T) {
	cc := CumulativeContribution{
		MemberID:           "mem-001",
		MemberNo:           "M001",
		FullName:           "Jane Smith",
		EmployeeCumulative: 3600000,
		EmployerCumulative: 7200000,
		AVCCumulative:      600000,
		TotalCumulative:    11400000,
	}

	if cc.MemberNo != "M001" {
		t.Errorf("Expected M001, got: %s", cc.MemberNo)
	}
	if cc.TotalCumulative != 11400000 {
		t.Errorf("Expected total cumulative 11400000, got: %d", cc.TotalCumulative)
	}
}

func TestRegisteredVsUnregistered_Struct(t *testing.T) {
	rvu := RegisteredVsUnregistered{
		MemberID:          "mem-001",
		MemberNo:          "M001",
		FullName:          "Test Member",
		RegisteredTotal:   5000000,
		UnregisteredTotal: 500000,
		Year:              2026,
	}

	if rvu.Year != 2026 {
		t.Errorf("Expected year 2026, got: %d", rvu.Year)
	}
	if rvu.RegisteredTotal != 5000000 {
		t.Errorf("Expected registered total 5000000, got: %d", rvu.RegisteredTotal)
	}
}

func TestContributionTrend_Struct(t *testing.T) {
	ct := ContributionTrend{
		Month:          "2026-01",
		EmployeeAmount: 500000,
		EmployerAmount: 1000000,
		AVCAmount:      100000,
		TotalAmount:    1600000,
		MemberCount:    100,
	}

	if ct.Month != "2026-01" {
		t.Errorf("Expected month 2026-01, got: %s", ct.Month)
	}
	if ct.TotalAmount != 1600000 {
		t.Errorf("Expected total amount 1600000, got: %d", ct.TotalAmount)
	}
	if ct.MemberCount != 100 {
		t.Errorf("Expected member count 100, got: %d", ct.MemberCount)
	}
}

func TestAVCSummary_Struct(t *testing.T) {
	as := AVCSummary{
		MemberID:      "mem-001",
		MemberNo:      "M001",
		FullName:      "AVC Member",
		AVCYearToDate: 120000,
		AVCCumulative: 720000,
	}

	if as.MemberNo != "M001" {
		t.Errorf("Expected M001, got: %s", as.MemberNo)
	}
	if as.AVCCumulative != 720000 {
		t.Errorf("Expected AVC cumulative 720000, got: %d", as.AVCCumulative)
	}
}
