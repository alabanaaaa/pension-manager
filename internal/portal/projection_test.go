package portal

import (
	"testing"
)

func TestProjectBenefits_DC(t *testing.T) {
	params := ProjectionParams{
		CurrentAge:       30,
		RetirementAge:    60,
		CurrentSalary:    50000,
		CurrentBalance:   100000,
		MemberRate:       5.0,
		SponsorRate:      10.0,
		SalaryGrowthRate: 5.0,
		InvestmentReturn: 8.0,
		SchemeType:       "dc",
		YearsOfService:   5,
	}

	result := ProjectBenefits(params)

	if result.SchemeType != "dc" {
		t.Errorf("Expected scheme type dc, got: %s", result.SchemeType)
	}
	if result.ProjectedBalance <= result.CurrentBalance {
		t.Errorf("Expected projected balance > current balance, got: %d vs %d", result.ProjectedBalance, result.CurrentBalance)
	}
	if result.TotalContributions <= 0 {
		t.Errorf("Expected positive total contributions, got: %d", result.TotalContributions)
	}
	if result.TotalInterest <= 0 {
		t.Errorf("Expected positive total interest, got: %d", result.TotalInterest)
	}
	if len(result.YearByYear) != 30 {
		t.Errorf("Expected 30 year projections, got: %d", len(result.YearByYear))
	}
	if result.IncomeReplacement <= 0 {
		t.Errorf("Expected positive IRR, got: %f", result.IncomeReplacement)
	}

	// Verify year-by-year data is increasing
	for i := 1; i < len(result.YearByYear); i++ {
		if result.YearByYear[i].EndBalance <= result.YearByYear[i-1].EndBalance {
			t.Errorf("Year %d balance (%d) should be > Year %d balance (%d)",
				i+1, result.YearByYear[i].EndBalance, i, result.YearByYear[i-1].EndBalance)
		}
	}
}

func TestProjectBenefits_DB(t *testing.T) {
	params := ProjectionParams{
		CurrentAge:       40,
		RetirementAge:    60,
		CurrentSalary:    80000,
		CurrentBalance:   500000,
		MemberRate:       5.0,
		SponsorRate:      10.0,
		SalaryGrowthRate: 5.0,
		InvestmentReturn: 8.0,
		SchemeType:       "db",
		YearsOfService:   15,
	}

	result := ProjectBenefits(params)

	if result.SchemeType != "db" {
		t.Errorf("Expected scheme type db, got: %s", result.SchemeType)
	}
	if result.EstimatedMonthly <= 0 {
		t.Errorf("Expected positive monthly pension, got: %d", result.EstimatedMonthly)
	}
	if result.EstimatedLumpSum <= 0 {
		t.Errorf("Expected positive lump sum, got: %d", result.EstimatedLumpSum)
	}
}

func TestProjectBenefits_AlreadyRetired(t *testing.T) {
	params := ProjectionParams{
		CurrentAge:     65,
		RetirementAge:  60,
		CurrentSalary:  50000,
		CurrentBalance: 100000,
		SchemeType:     "dc",
	}

	result := ProjectBenefits(params)

	if result.ProjectedBalance != result.CurrentBalance {
		t.Errorf("Expected projected balance = current balance for retired member")
	}
	if len(result.YearByYear) != 0 {
		t.Errorf("Expected no year projections for retired member")
	}
}

func TestCalculateIRR(t *testing.T) {
	irr := CalculateIRR(30000, 100000)
	// 30000*12 / 100000 * 100 = 360%
	if irr != 360.0 {
		t.Errorf("Expected IRR 360.0, got: %f", irr)
	}

	// Zero salary
	irr = CalculateIRR(30000, 0)
	if irr != 0 {
		t.Errorf("Expected IRR 0 for zero salary, got: %f", irr)
	}
}

func TestGenerateBenefitQuote(t *testing.T) {
	dcResult := &ProjectionResult{
		SchemeType:       "dc",
		CurrentBalance:   100000,
		ProjectedBalance: 500000,
	}

	quote := GenerateBenefitQuote("mem-001", nil, dcResult)

	if quote.MemberID != "mem-001" {
		t.Errorf("Expected member ID mem-001, got: %s", quote.MemberID)
	}
	if quote.DCProjection == nil {
		t.Error("Expected DC projection in quote")
	}
	if quote.Disclaimer == "" {
		t.Error("Expected disclaimer in quote")
	}
}

func TestDefaultProjectionParams(t *testing.T) {
	params := DefaultProjectionParams(35, 60, 60000, 200000, "dc")

	if params.CurrentAge != 35 {
		t.Errorf("Expected current age 35, got: %d", params.CurrentAge)
	}
	if params.RetirementAge != 60 {
		t.Errorf("Expected retirement age 60, got: %d", params.RetirementAge)
	}
	if params.MemberRate != 5.0 {
		t.Errorf("Expected member rate 5.0, got: %f", params.MemberRate)
	}
	if params.SponsorRate != 10.0 {
		t.Errorf("Expected sponsor rate 10.0, got: %f", params.SponsorRate)
	}
	if params.SchemeType != "dc" {
		t.Errorf("Expected scheme type dc, got: %s", params.SchemeType)
	}
}
