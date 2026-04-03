package portal

import (
	"math"
	"time"
)

// ProjectionParams holds parameters for benefit projection
type ProjectionParams struct {
	CurrentAge       int
	RetirementAge    int
	CurrentSalary    int64
	CurrentBalance   int64
	MemberRate       float64 // percentage
	SponsorRate      float64 // percentage
	SalaryGrowthRate float64 // annual percentage
	InvestmentReturn float64 // annual percentage
	SchemeType       string  // "db" or "dc"
	YearsOfService   int
}

// ProjectionResult holds the benefit projection result
type ProjectionResult struct {
	SchemeType         string           `json:"scheme_type"`
	CurrentBalance     int64            `json:"current_balance"`
	ProjectedBalance   int64            `json:"projected_balance"`
	TotalContributions int64            `json:"total_contributions"`
	TotalInterest      int64            `json:"total_interest"`
	EstimatedMonthly   int64            `json:"estimated_monthly_pension"`
	EstimatedLumpSum   int64            `json:"estimated_lump_sum"`
	IncomeReplacement  float64          `json:"income_replacement_ratio"`
	YearByYear         []YearProjection `json:"year_by_year"`
}

// YearProjection holds a single year's projection
type YearProjection struct {
	Year           int   `json:"year"`
	Age            int   `json:"age"`
	Salary         int64 `json:"salary"`
	MemberContrib  int64 `json:"member_contribution"`
	SponsorContrib int64 `json:"sponsor_contribution"`
	Interest       int64 `json:"interest"`
	EndBalance     int64 `json:"end_balance"`
}

// ProjectBenefits calculates benefit projections for DB and DC schemes
func ProjectBenefits(params ProjectionParams) *ProjectionResult {
	yearsToRetire := params.RetirementAge - params.CurrentAge
	if yearsToRetire <= 0 {
		return &ProjectionResult{
			SchemeType:       params.SchemeType,
			CurrentBalance:   params.CurrentBalance,
			ProjectedBalance: params.CurrentBalance,
		}
	}

	balance := params.CurrentBalance
	totalContributions := int64(0)
	totalInterest := int64(0)
	salary := params.CurrentSalary
	var yearByYear []YearProjection

	for year := 1; year <= yearsToRetire; year++ {
		// Project salary growth
		salary = int64(float64(salary) * (1 + params.SalaryGrowthRate/100))

		// Calculate contributions
		memberContrib := int64(float64(salary) * params.MemberRate / 100)
		sponsorContrib := int64(float64(salary) * params.SponsorRate / 100)
		annualContrib := memberContrib + sponsorContrib

		// Calculate interest on opening balance + half of contributions (mid-year assumption)
		openingBalance := balance
		interest := int64(float64(openingBalance+annualContrib/2) * params.InvestmentReturn / 100)

		balance = openingBalance + annualContrib + interest
		totalContributions += annualContrib
		totalInterest += interest

		yearByYear = append(yearByYear, YearProjection{
			Year:           time.Now().Year() + year,
			Age:            params.CurrentAge + year,
			Salary:         salary,
			MemberContrib:  memberContrib,
			SponsorContrib: sponsorContrib,
			Interest:       interest,
			EndBalance:     balance,
		})
	}

	// Calculate retirement benefits
	var estimatedMonthly, estimatedLumpSum int64
	var irr float64

	if params.SchemeType == "db" {
		// DB: Final salary * years of service * accrual rate (typically 1/60 or 1/80)
		accrualRate := 1.0 / 60.0
		finalSalary := salary
		totalYears := params.YearsOfService + yearsToRetire
		annualPension := int64(float64(finalSalary) * float64(totalYears) * accrualRate)
		estimatedMonthly = annualPension / 12
		// 1/3 commutation
		estimatedLumpSum = balance / 3
		irr = float64(annualPension) / float64(finalSalary) * 100
	} else {
		// DC: Use accumulated balance
		// 2/3 lump sum, 1/3 annuity
		estimatedLumpSum = balance * 2 / 3
		annuityAmount := balance / 3
		// Annuity factor (simplified - 20 year annuity at 5%)
		annuityFactor := 12.462 // PV of 1/month for 20 years at 5%
		estimatedMonthly = int64(float64(annuityAmount) / annuityFactor)
		irr = float64(estimatedMonthly*12) / float64(salary) * 100
	}

	return &ProjectionResult{
		SchemeType:         params.SchemeType,
		CurrentBalance:     params.CurrentBalance,
		ProjectedBalance:   balance,
		TotalContributions: totalContributions,
		TotalInterest:      totalInterest,
		EstimatedMonthly:   estimatedMonthly,
		EstimatedLumpSum:   estimatedLumpSum,
		IncomeReplacement:  math.Round(irr*100) / 100,
		YearByYear:         yearByYear,
	}
}

// CalculateIRR calculates the Income Replacement Ratio
func CalculateIRR(monthlyPension int64, finalSalary int64) float64 {
	if finalSalary == 0 {
		return 0
	}
	irr := float64(monthlyPension*12) / float64(finalSalary) * 100
	return math.Round(irr*100) / 100
}

// GenerateBenefitQuote generates a benefit quote for a member
func GenerateBenefitQuote(memberID string, dbResult, dcResult *ProjectionResult) *BenefitQuote {
	return &BenefitQuote{
		MemberID:     memberID,
		GeneratedAt:  time.Now(),
		DBProjection: dbResult,
		DCProjection: dcResult,
		Disclaimer:   "This is a projection only. Actual benefits may vary based on market conditions, salary changes, and scheme rules.",
	}
}

// BenefitQuote holds benefit projections for both DB and DC schemes
type BenefitQuote struct {
	MemberID     string            `json:"member_id"`
	GeneratedAt  time.Time         `json:"generated_at"`
	DBProjection *ProjectionResult `json:"db_projection,omitempty"`
	DCProjection *ProjectionResult `json:"dc_projection,omitempty"`
	Disclaimer   string            `json:"disclaimer"`
}

// DefaultProjectionParams returns default projection parameters
func DefaultProjectionParams(currentAge, retirementAge int, currentSalary, currentBalance int64, schemeType string) ProjectionParams {
	return ProjectionParams{
		CurrentAge:       currentAge,
		RetirementAge:    retirementAge,
		CurrentSalary:    currentSalary,
		CurrentBalance:   currentBalance,
		MemberRate:       5.0,  // 5% member contribution
		SponsorRate:      10.0, // 10% sponsor contribution
		SalaryGrowthRate: 5.0,  // 5% annual salary growth
		InvestmentReturn: 8.0,  // 8% annual investment return
		SchemeType:       schemeType,
		YearsOfService:   currentAge - 25, // rough estimate
	}
}
