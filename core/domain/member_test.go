package domain

import (
	"testing"
	"time"
)

func TestMemberValidate(t *testing.T) {
	// Valid member
	member := &Member{
		ID:               "mem-001",
		SchemeID:         "scheme-001",
		MemberNo:         "M001",
		FirstName:        "John",
		LastName:         "Doe",
		DateOfBirth:      time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC),
		DateJoinedScheme: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		MembershipStatus: StatusDeceased,
		BasicSalary:      50000,
	}

	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member, got: %v", err)
	}

	// Invalid member - missing first name
	member.FirstName = ""
	if err := member.Validate(); err == nil {
		t.Error("Expected error for missing first name")
	}

	// Invalid member - negative inpatient limit
	member.FirstName = "John"
	member.InpatientLimit = -1000
	if err := member.Validate(); err == nil {
		t.Error("Expected error for negative inpatient limit")
	}

	// Invalid member - invalid PIN length
	member.InpatientLimit = 0
	member.PIN = "123"
	if err := member.Validate(); err == nil {
		t.Error("Expected error for PIN too short")
	}

	// Invalid member - PIN with non-digits
	member.PIN = "1234"
	member.PIN = "12a4"
	if err := member.Validate(); err == nil {
		t.Error("Expected error for PIN with non-digits")
	}

	// Invalid member - date of death before date of birth
	member.PIN = "1234"
	member.DateOfDeath = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := member.Validate(); err == nil {
		t.Error("Expected error for date of death before date of birth")
	}

	// Valid member with PIN
	member.DateOfDeath = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	member.PIN = "1234"
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with PIN, got: %v", err)
	}

	// Valid member with biometrics
	member.Photograph = "photo.jpg"
	member.FingerprintData = "fingerprint_data"
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with biometrics, got: %v", err)
	}

	// Invalid member - photograph without fingerprint
	member.FingerprintData = ""
	if err := member.Validate(); err == nil {
		t.Error("Expected error for photograph without fingerprint data")
	}

	// Invalid member - fingerprint without photograph
	member.FingerprintData = "fingerprint_data"
	member.Photograph = ""
	if err := member.Validate(); err == nil {
		t.Error("Expected error for fingerprint data without photograph")
	}

	// Valid member with both biometrics
	member.Photograph = "photo.jpg"
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with biometrics, got: %v", err)
	}

	// Invalid member - negative children under 21
	member.ChildrenUnder21Count = -1
	if err := member.Validate(); err == nil {
		t.Error("Expected error for negative children under 21 count")
	}

	// Valid member
	member.ChildrenUnder21Count = 2
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with children count, got: %v", err)
	}

	// Invalid member - invalid membership card status
	member.MembershipCardStatus = "invalid"
	if err := member.Validate(); err == nil {
		t.Error("Expected error for invalid membership card status")
	}

	// Valid member
	member.MembershipCardStatus = "issue"
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with membership card status, got: %v", err)
	}

	// Invalid member - cessation date before date of birth
	member.CessationDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := member.Validate(); err == nil {
		t.Error("Expected error for cessation date before date of birth")
	}

	// Valid member
	member.CessationDate = time.Time{}
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member without cessation date, got: %v", err)
	}

	// Invalid member - tax exemption cutoff date without reason
	member.TaxExemptCutoffDate = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := member.Validate(); err == nil {
		t.Error("Expected error for tax exemption cutoff date without reason")
	}

	// Invalid member - tax exemption reason without cutoff date
	member.TaxExemptReason = "Old age"
	member.TaxExemptCutoffDate = time.Time{}
	if err := member.Validate(); err == nil {
		t.Error("Expected error for tax exemption reason without cutoff date")
	}

	// Valid member with tax exemption
	member.TaxExemptCutoffDate = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with tax exemption, got: %v", err)
	}

	// Invalid member - negative member contribution rate
	member.MemberContributionRate = -5
	if err := member.Validate(); err == nil {
		t.Error("Expected error for negative member contribution rate")
	}

	// Invalid member - member contribution rate > 100
	member.MemberContributionRate = 0
	member.MemberContributionRate = 105
	if err := member.Validate(); err == nil {
		t.Error("Expected error for member contribution rate > 100")
	}

	// Valid member
	member.MemberContributionRate = 5
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with contribution rate, got: %v", err)
	}

	// Invalid member - negative sponsor contribution rate
	member.SponsorContributionRate = -5
	if err := member.Validate(); err == nil {
		t.Error("Expected error for negative sponsor contribution rate")
	}

	// Invalid member - sponsor contribution rate > 100
	member.SponsorContributionRate = 0
	member.SponsorContributionRate = 105
	if err := member.Validate(); err == nil {
		t.Error("Expected error for sponsor contribution rate > 100")
	}

	// Valid member
	member.SponsorContributionRate = 5
	if err := member.Validate(); err != nil {
		t.Errorf("Expected no error for valid member with sponsor contribution rate, got: %v", err)
	}
}

func TestMemberMethods(t *testing.T) {
	// Use a member who is deceased (date of death in the past)
	member := &Member{
		ID:                      "mem-001",
		SchemeID:                "scheme-001",
		MemberNo:                "M001",
		FirstName:               "John",
		LastName:                "Doe",
		DateOfBirth:             time.Date(1940, 1, 1, 0, 0, 0, 0, time.UTC), // Born 1940
		DateJoinedScheme:        time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), // Joined 1970
		ExpectedRetirement:      time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC), // Retired 2005
		MembershipStatus:        StatusDeceased,
		BasicSalary:             50000,
		DateOfDeath:             time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Died 2020
		TotalWithdrawals:        10000,
		LastWithdrawalDate:      time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		PIN:                     "1234",
		Photograph:              "photo.jpg",
		FingerprintData:         "fingerprint_data",
		ChildrenUnder21Count:    2,
		MembershipCardIssueDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		MembershipCardStatus:    "issue",
		PreviousSponsors:        []string{"sponsor-1", "sponsor-2"},
		CessationDate:           time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		CessationReason:         "Transfer to another scheme",
		TaxExemptReason:         "Retirement age",
		TaxExemptCutoffDate:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		MemberContributionRate:  5.0,
		SponsorContributionRate: 10.0,
		InpatientLimit:          100000,
		OutpatientLimit:         50000,
	}

	// Test IsDeceased (member has StatusDeceased set)
	if !member.IsDeceased() {
		t.Error("Expected member to be deceased")
	}

	// Test IsActive (deceased member is not active)
	if member.IsActive() {
		t.Error("Expected deceased member to not be active")
	}

	// Test HasWithdrawals
	if !member.HasWithdrawals() {
		t.Error("Expected member to have withdrawals")
	}

	// Test HasBiometrics
	if !member.HasBiometrics() {
		t.Error("Expected member to have biometrics")
	}

	// Test GetDependentCount
	depCount := member.GetDependentCount()
	if depCount != 2 {
		t.Errorf("Expected dependent count 2, got: %d", depCount)
	}

	// Test IsMembershipCardValid
	if !member.IsMembershipCardValid() {
		t.Error("Expected membership card to be valid")
	}

	// Test GetPreviousSponsorsCount
	sponsorCount := member.GetPreviousSponsorsCount()
	if sponsorCount != 2 {
		t.Errorf("Expected previous sponsors count 2, got: %d", sponsorCount)
	}

	// Test HasCessationDetails
	if !member.HasCessationDetails() {
		t.Error("Expected member to have cessation details")
	}
}
