package tax

// KRA Tax Brackets (2024 rates - PAYE)
var taxBrackets = []TaxBracket{
	{Min: 0, Max: 240000, Rate: 0.10},
	{Min: 240001, Max: 323333, Rate: 0.25},
	{Min: 323334, Max: 416667, Rate: 0.30},
	{Min: 416668, Max: 583333, Rate: 0.325},
	{Min: 583334, Max: 800000, Rate: 0.35},
	{Min: 800001, Max: 0, Rate: 0.35},
}

// TaxBracket represents a KRA PAYE tax bracket
type TaxBracket struct {
	Min  int64
	Max  int64
	Rate float64
}

// Relief represents a tax relief amount
type Relief struct {
	Name        string
	Amount      int64
	Description string
}

var (
	InsuranceRelief = Relief{Name: "insurance", Amount: 500000, Description: "Insurance relief (5,000 KES/month)"}
	PersonalRelief  = Relief{Name: "personal", Amount: 240000, Description: "Personal relief (2,400 KES/month)"}
	ShelterRelief   = Relief{Name: "shelter", Amount: 125000, Description: "Shelter relief (12,500 KES/month)"}
)

// TaxResult holds the computed tax
type TaxResult struct {
	GrossPay        int64           `json:"gross_pay"`
	TaxableIncome   int64           `json:"taxable_income"`
	TotalTaxBefore  int64           `json:"total_tax_before_relief"`
	Reliefs         []ReliefApplied `json:"reliefs"`
	TotalRelief     int64           `json:"total_relief"`
	NetTax          int64           `json:"net_tax"`
	EffectiveRate   float64         `json:"effective_rate"`
	IsExempt        bool            `json:"is_exempt"`
	ExemptionReason string          `json:"exemption_reason,omitempty"`
}

// ReliefApplied represents an applied tax relief
type ReliefApplied struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

// MultiSchemeTaxResult holds tax results for multiple schemes
type MultiSchemeTaxResult struct {
	MemberID      string      `json:"member_id"`
	Schemes       []SchemeTax `json:"schemes"`
	CombinedTax   int64       `json:"combined_tax"`
	TotalIncome   int64       `json:"total_income"`
	EffectiveRate float64     `json:"effective_rate"`
}

// SchemeTax holds tax info for a single scheme
type SchemeTax struct {
	SchemeID       string `json:"scheme_id"`
	SchemeName     string `json:"scheme_name"`
	GrossPay       int64  `json:"gross_pay"`
	TaxableIncome  int64  `json:"taxable_income"`
	TaxAmount      int64  `json:"tax_amount"`
	ApportionedTax int64  `json:"apportioned_tax"`
}

// ComputeTax calculates PAYE tax for a given annual income
func ComputeTax(annualIncome int64, reliefs []Relief, isExempt bool, exemptionReason string) *TaxResult {
	if isExempt {
		return &TaxResult{
			GrossPay:        annualIncome,
			TaxableIncome:   annualIncome,
			TotalTaxBefore:  0,
			NetTax:          0,
			EffectiveRate:   0,
			IsExempt:        true,
			ExemptionReason: exemptionReason,
		}
	}

	totalTax := int64(0)
	for _, bracket := range taxBrackets {
		if annualIncome <= bracket.Min {
			break
		}
		var taxableInBracket int64
		if bracket.Max == 0 || annualIncome > bracket.Max {
			taxableInBracket = bracket.Max - bracket.Min + 1
			if bracket.Max == 0 {
				taxableInBracket = annualIncome - bracket.Min + 1
			}
		} else {
			taxableInBracket = annualIncome - bracket.Min + 1
		}
		totalTax += int64(float64(taxableInBracket) * bracket.Rate)
	}

	var reliefsApplied []ReliefApplied
	totalRelief := int64(0)
	for _, relief := range reliefs {
		reliefsApplied = append(reliefsApplied, ReliefApplied{Name: relief.Name, Amount: relief.Amount})
		totalRelief += relief.Amount
	}

	netTax := totalTax - totalRelief
	if netTax < 0 {
		netTax = 0
	}

	effectiveRate := float64(0)
	if annualIncome > 0 {
		effectiveRate = float64(netTax) / float64(annualIncome) * 100
	}

	return &TaxResult{
		GrossPay:       annualIncome,
		TaxableIncome:  annualIncome,
		TotalTaxBefore: totalTax,
		Reliefs:        reliefsApplied,
		TotalRelief:    totalRelief,
		NetTax:         netTax,
		EffectiveRate:  effectiveRate,
	}
}

// ComputeTaxMonthly computes monthly PAYE from annual income
func ComputeTaxMonthly(monthlyIncome int64, reliefs []Relief, isExempt bool, exemptionReason string) *TaxResult {
	annualIncome := monthlyIncome * 12
	result := ComputeTax(annualIncome, reliefs, isExempt, exemptionReason)
	annualRelief := int64(0)
	for _, r := range result.Reliefs {
		annualRelief += r.Amount
	}
	monthlyRelief := annualRelief / 12
	result.GrossPay = monthlyIncome
	result.TaxableIncome = monthlyIncome
	result.TotalTaxBefore = result.TotalTaxBefore / 12
	result.TotalRelief = monthlyRelief
	result.NetTax = result.NetTax / 12
	return result
}

// ComputeMultiSchemeTax computes tax across multiple schemes for a member
func ComputeMultiSchemeTax(memberID string, schemes []SchemeTax, isExempt bool, exemptionReason string) *MultiSchemeTaxResult {
	var totalIncome int64
	for i := range schemes {
		schemeTax := ComputeTax(schemes[i].GrossPay*12, nil, false, "")
		schemes[i].TaxAmount = schemeTax.NetTax
		totalIncome += schemes[i].GrossPay
	}

	combinedResult := ComputeTax(totalIncome*12, nil, isExempt, exemptionReason)
	combinedTax := combinedResult.NetTax

	for i := range schemes {
		if totalIncome > 0 {
			ratio := float64(schemes[i].GrossPay) / float64(totalIncome)
			schemes[i].ApportionedTax = int64(float64(combinedTax) * ratio)
		}
	}

	effectiveRate := float64(0)
	if totalIncome > 0 {
		effectiveRate = float64(combinedTax) / float64(totalIncome*12) * 100
	}

	return &MultiSchemeTaxResult{
		MemberID:      memberID,
		Schemes:       schemes,
		CombinedTax:   combinedTax,
		TotalIncome:   totalIncome,
		EffectiveRate: effectiveRate,
	}
}

// GetTaxBrackets returns the current KRA tax brackets
func GetTaxBrackets() []TaxBracket {
	return taxBrackets
}

// GetAvailableReliefs returns all available tax reliefs
func GetAvailableReliefs() []Relief {
	return []Relief{InsuranceRelief, PersonalRelief, ShelterRelief}
}

// CalculateNHIFRelief calculates NHIF relief based on gross pay
func CalculateNHIFRelief(grossPay int64) int64 {
	if grossPay <= 5999 {
		return 0
	} else if grossPay <= 7999 {
		return 30000
	} else if grossPay <= 11999 {
		return 50000
	} else if grossPay <= 14999 {
		return 75000
	} else if grossPay <= 19999 {
		return 100000
	} else if grossPay <= 24999 {
		return 125000
	} else if grossPay <= 29999 {
		return 150000
	} else if grossPay <= 34999 {
		return 175000
	} else if grossPay <= 39999 {
		return 200000
	} else if grossPay <= 44999 {
		return 225000
	} else if grossPay <= 49999 {
		return 250000
	} else if grossPay <= 59999 {
		return 275000
	} else if grossPay <= 69999 {
		return 300000
	} else if grossPay <= 79999 {
		return 325000
	} else if grossPay <= 89999 {
		return 350000
	} else if grossPay <= 99999 {
		return 375000
	}
	return 400000
}

// CalculateTaxOnWithdrawal calculates tax on pension withdrawal
func CalculateTaxOnWithdrawal(withdrawalAmount int64, yearsOfService int, isTaxExempt bool, exemptionReason string) *TaxResult {
	if isTaxExempt {
		return &TaxResult{
			GrossPay:        withdrawalAmount,
			TaxableIncome:   withdrawalAmount,
			TotalTaxBefore:  0,
			NetTax:          0,
			EffectiveRate:   0,
			IsExempt:        true,
			ExemptionReason: exemptionReason,
		}
	}
	taxFreeAmount := withdrawalAmount / 3
	taxableAmount := withdrawalAmount - taxFreeAmount
	result := ComputeTax(taxableAmount, []Relief{PersonalRelief}, false, "")
	result.GrossPay = withdrawalAmount
	result.TaxableIncome = taxableAmount
	return result
}

// CalculateTaxOnBenefit calculates tax on retirement benefit
func CalculateTaxOnBenefit(benefitAmount int64, monthlyPension int64, age int, isTaxExempt bool, exemptionReason string) *TaxResult {
	if isTaxExempt || age >= 65 {
		return &TaxResult{
			GrossPay:        benefitAmount,
			TaxableIncome:   benefitAmount,
			TotalTaxBefore:  0,
			NetTax:          0,
			EffectiveRate:   0,
			IsExempt:        true,
			ExemptionReason: exemptionReason,
		}
	}
	taxFreeAmount := benefitAmount / 3
	taxableAmount := benefitAmount - taxFreeAmount
	result := ComputeTax(taxableAmount, []Relief{PersonalRelief}, false, "")
	result.GrossPay = benefitAmount
	result.TaxableIncome = taxableAmount
	return result
}

// CalculateMonthlyPensionTax calculates tax on monthly pension
func CalculateMonthlyPensionTax(monthlyPension int64, age int, isTaxExempt bool, exemptionReason string) *TaxResult {
	if isTaxExempt || age >= 65 {
		return &TaxResult{
			GrossPay:        monthlyPension,
			TaxableIncome:   monthlyPension,
			TotalTaxBefore:  0,
			NetTax:          0,
			EffectiveRate:   0,
			IsExempt:        true,
			ExemptionReason: exemptionReason,
		}
	}
	return ComputeTaxMonthly(monthlyPension, []Relief{PersonalRelief, InsuranceRelief}, false, "")
}
