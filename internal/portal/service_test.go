package portal

import (
	"testing"
	"time"
)

func TestMemberProfile_Struct(t *testing.T) {
	profile := &MemberProfile{
		PersonalInfo: PersonalInfo{
			FullName:      "John Doe",
			NationalID:    "12345678",
			Gender:        "Male",
			DateOfBirth:   time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
			PhoneNumber:   "+254712345678",
			MaritalStatus: "Married",
			SpouseName:    "Jane Doe",
			KRAPIN:        "A001234567B",
			Nationality:   "Kenyan",
			Age:           36,
		},
		ContactInfo: ContactInfo{
			Email:         "john@example.com",
			MobileNumber:  "+254712345678",
			WorkTelephone: "020-1234567",
			PostalCode:    "00100",
			PostalAddress: "P.O. Box 12345",
			Town:          "Nairobi",
		},
		EmploymentInfo: EmploymentInfo{
			SchemeName:       "Test Scheme",
			SponsorName:      "Test Sponsor",
			SponsorCode:      "SP001",
			DateJoinedScheme: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			PayrollNo:        "PAY001",
			BankName:         "KCB",
			BankBranch:       "Kenyatta Avenue",
			BankAccount:      "1234567890",
			Designation:      "Software Engineer",
			Department:       "IT",
			BasicSalary:      150000,
			MemberNo:         "M001",
		},
		MedicalLimits: MedicalLimits{
			InpatientLimit:  500000,
			OutpatientLimit: 200000,
		},
		AccountSummary: AccountSummary{
			AccountBalance:   2500000,
			TotalWithdrawals: 100000,
			MembershipStatus: "active",
		},
	}

	if profile.PersonalInfo.FullName != "John Doe" {
		t.Errorf("Expected full name John Doe, got: %s", profile.PersonalInfo.FullName)
	}
	if profile.EmploymentInfo.BasicSalary != 150000 {
		t.Errorf("Expected basic salary 150000, got: %d", profile.EmploymentInfo.BasicSalary)
	}
	if profile.AccountSummary.AccountBalance != 2500000 {
		t.Errorf("Expected account balance 2500000, got: %d", profile.AccountSummary.AccountBalance)
	}
}

func TestBeneficiary_Struct(t *testing.T) {
	b := Beneficiary{
		ID:              "ben-001",
		Name:            "Jane Doe",
		NationalID:      "87654321",
		MobileTelephone: "+254712345679",
		Relationship:    "Spouse",
		PhysicalAddress: "Nairobi",
		AllocationPct:   50.0,
		DateOfBirth:     time.Date(1992, 5, 20, 0, 0, 0, 0, time.UTC),
	}

	if b.Name != "Jane Doe" {
		t.Errorf("Expected name Jane Doe, got: %s", b.Name)
	}
	if b.AllocationPct != 50.0 {
		t.Errorf("Expected allocation 50.0, got: %f", b.AllocationPct)
	}
}

func TestMemberContribution_Struct(t *testing.T) {
	c := MemberContribution{
		ID:             "cont-001",
		Period:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EmployeeAmount: 7500,
		EmployerAmount: 15000,
		AVCAmount:      5000,
		TotalAmount:    27500,
		PaymentMethod:  "mpesa",
		Status:         "confirmed",
		ReceiptNo:      "RCP001",
	}

	if c.TotalAmount != 27500 {
		t.Errorf("Expected total amount 27500, got: %d", c.TotalAmount)
	}
	if c.PaymentMethod != "mpesa" {
		t.Errorf("Expected payment method mpesa, got: %s", c.PaymentMethod)
	}
}

func TestAnnualContribution_Struct(t *testing.T) {
	a := AnnualContribution{
		Year:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EmployeeTotal: 90000,
		EmployerTotal: 180000,
		AVCTotal:      60000,
		GrandTotal:    330000,
	}

	if a.GrandTotal != 330000 {
		t.Errorf("Expected grand total 330000, got: %d", a.GrandTotal)
	}
}

func TestChangeRequest_Struct(t *testing.T) {
	reason := "Invalid data"
	reviewedAt := time.Now()
	cr := ChangeRequest{
		ID:              "cr-001",
		RequestType:     "contact_change",
		Status:          "rejected",
		RejectionReason: &reason,
		CreatedAt:       time.Now().Add(-24 * time.Hour),
		ReviewedAt:      &reviewedAt,
	}

	if cr.RequestType != "contact_change" {
		t.Errorf("Expected request type contact_change, got: %s", cr.RequestType)
	}
	if cr.Status != "rejected" {
		t.Errorf("Expected status rejected, got: %s", cr.Status)
	}
	if *cr.RejectionReason != "Invalid data" {
		t.Errorf("Expected rejection reason 'Invalid data', got: %s", *cr.RejectionReason)
	}
}

func TestFeedback_Struct(t *testing.T) {
	f := Feedback{
		ID:        "fb-001",
		MemberID:  "mem-001",
		SchemeID:  "scheme-001",
		Subject:   "Question about benefits",
		Message:   "How do I project my benefits?",
		Status:    "open",
		CreatedAt: time.Now(),
	}

	if f.Subject != "Question about benefits" {
		t.Errorf("Expected subject, got: %s", f.Subject)
	}
	if f.Status != "open" {
		t.Errorf("Expected status open, got: %s", f.Status)
	}
}

func TestLoginStats_Struct(t *testing.T) {
	lastLogin := time.Now()
	ls := LoginStats{
		TotalLogins:      25,
		LastLogin:        &lastLogin,
		LoginsLast30Days: 5,
	}

	if ls.TotalLogins != 25 {
		t.Errorf("Expected total logins 25, got: %d", ls.TotalLogins)
	}
	if ls.LoginsLast30Days != 5 {
		t.Errorf("Expected logins last 30 days 5, got: %d", ls.LoginsLast30Days)
	}
}
