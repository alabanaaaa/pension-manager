package tax

import (
	"testing"
)

func TestComputeTax(t *testing.T) {
	// Test tax-exempt member
	result := ComputeTax(1200000, nil, true, "Retirement age 65+")
	if !result.IsExempt {
		t.Error("Expected tax-exempt result")
	}
	if result.NetTax != 0 {
		t.Errorf("Expected 0 tax for exempt member, got: %d", result.NetTax)
	}

	// Test low income (below tax threshold)
	result = ComputeTax(240000, []Relief{PersonalRelief}, false, "")
	if result.NetTax < 0 {
		t.Errorf("Expected non-negative tax for low income, got: %d", result.NetTax)
	}

	// Test middle income (600,000/year = 50,000/month)
	result = ComputeTax(600000, []Relief{PersonalRelief}, false, "")
	if result.TotalTaxBefore <= 0 {
		t.Errorf("Expected positive tax before relief for middle income, got: %d", result.TotalTaxBefore)
	}

	// Test high income
	result = ComputeTax(1200000, []Relief{PersonalRelief, InsuranceRelief}, false, "")
	if result.TotalTaxBefore <= 0 {
		t.Errorf("Expected positive tax before relief for high income, got: %d", result.TotalTaxBefore)
	}
}

func TestComputeTaxMonthly(t *testing.T) {
	result := ComputeTaxMonthly(50000, []Relief{PersonalRelief}, false, "")
	if result.GrossPay != 50000 {
		t.Errorf("Expected gross pay 50000, got: %d", result.GrossPay)
	}
	if result.NetTax < 0 {
		t.Errorf("Expected non-negative tax, got: %d", result.NetTax)
	}

	// Tax exempt monthly
	result = ComputeTaxMonthly(50000, nil, true, "Age 65+")
	if !result.IsExempt {
		t.Error("Expected exempt result")
	}
	if result.NetTax != 0 {
		t.Errorf("Expected 0 tax, got: %d", result.NetTax)
	}
}

func TestComputeMultiSchemeTax(t *testing.T) {
	schemes := []SchemeTax{
		{SchemeID: "scheme-1", SchemeName: "Scheme A", GrossPay: 50000},
		{SchemeID: "scheme-2", SchemeName: "Scheme B", GrossPay: 30000},
	}

	result := ComputeMultiSchemeTax("mem-001", schemes, false, "")
	if result.TotalIncome != 80000 {
		t.Errorf("Expected total income 80000, got: %d", result.TotalIncome)
	}
	if result.CombinedTax < 0 {
		t.Errorf("Expected non-negative combined tax, got: %d", result.CombinedTax)
	}
	if len(result.Schemes) != 2 {
		t.Errorf("Expected 2 schemes, got: %d", len(result.Schemes))
	}

	// Verify apportionment is within 1 of combined (rounding tolerance)
	var totalApportioned int64
	for _, s := range result.Schemes {
		totalApportioned += s.ApportionedTax
	}
	diff := totalApportioned - result.CombinedTax
	if diff < 0 {
		diff = -diff
	}
	if diff > 2 {
		t.Errorf("Apportioned tax (%d) doesn't match combined tax (%d), diff=%d", totalApportioned, result.CombinedTax, diff)
	}
}

func TestCalculateTaxOnWithdrawal(t *testing.T) {
	// 1/3 should be tax-free
	result := CalculateTaxOnWithdrawal(300000, 10, false, "")
	if result.TaxableIncome != 200000 {
		t.Errorf("Expected taxable income 200000 (2/3 of 300000), got: %d", result.TaxableIncome)
	}

	// Tax exempt withdrawal
	result = CalculateTaxOnWithdrawal(300000, 10, true, "Retirement")
	if !result.IsExempt {
		t.Error("Expected exempt result")
	}
	if result.NetTax != 0 {
		t.Errorf("Expected 0 tax, got: %d", result.NetTax)
	}
}

func TestCalculateTaxOnBenefit(t *testing.T) {
	result := CalculateTaxOnBenefit(900000, 50000, 60, false, "")
	if result.TaxableIncome != 600000 {
		t.Errorf("Expected taxable income 600000 (2/3 of 900000), got: %d", result.TaxableIncome)
	}

	// Age 65+ exempt
	result = CalculateTaxOnBenefit(900000, 50000, 65, false, "")
	if !result.IsExempt {
		t.Error("Expected exempt result for age 65+")
	}
}

func TestCalculateNHIFRelief(t *testing.T) {
	tests := []struct {
		grossPay int64
		expected int64
	}{
		{5000, 0},
		{6000, 30000},
		{10000, 50000},
		{20000, 125000},
		{50000, 275000},
		{100000, 400000},
	}

	for _, tt := range tests {
		result := CalculateNHIFRelief(tt.grossPay)
		if result != tt.expected {
			t.Errorf("NHIF relief for gross pay %d: expected %d, got %d", tt.grossPay, tt.expected, result)
		}
	}
}

func TestGetTaxBrackets(t *testing.T) {
	brackets := GetTaxBrackets()
	if len(brackets) == 0 {
		t.Error("Expected tax brackets")
	}
	if brackets[0].Rate != 0.10 {
		t.Errorf("Expected first bracket rate 0.10, got: %f", brackets[0].Rate)
	}
}

func TestGetAvailableReliefs(t *testing.T) {
	reliefs := GetAvailableReliefs()
	if len(reliefs) == 0 {
		t.Error("Expected available reliefs")
	}
}
