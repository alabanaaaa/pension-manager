package domain

import (
	"time"
)

// MembershipStatus represents a member's current status in the scheme
type MembershipStatus string

const (
	StatusActive    MembershipStatus = "active"
	StatusInactive  MembershipStatus = "inactive"
	StatusSuspended MembershipStatus = "suspended"
	StatusDeferred  MembershipStatus = "deferred"
	StatusRetired   MembershipStatus = "retired"
	StatusDeceased  MembershipStatus = "deceased"
)

// MaritalStatus represents a member's marital status
type MaritalStatus string

const (
	Single    MaritalStatus = "single"
	Married   MaritalStatus = "married"
	Separated MaritalStatus = "separated"
	Divorced  MaritalStatus = "divorced"
	Widowed   MaritalStatus = "widowed"
)

// Gender represents a member's gender
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Other  Gender = "other"
)

// Member represents a pension scheme member
type Member struct {
	ID                 string           `json:"id"`
	SchemeID           string           `json:"scheme_id"`
	MemberNo           string           `json:"member_no"`
	FirstName          string           `json:"first_name"`
	LastName           string           `json:"last_name"`
	OtherNames         string           `json:"other_names,omitempty"`
	Gender             Gender           `json:"gender,omitempty"`
	DateOfBirth        time.Time        `json:"date_of_birth"`
	DateOfDeath        time.Time        `json:"date_of_death,omitempty"`
	Nationality        string           `json:"nationality"`
	IDNumber           string           `json:"id_number,omitempty"`
	KRAPIN             string           `json:"kra_pin,omitempty"`
	Email              string           `json:"email,omitempty"`
	Phone              string           `json:"phone,omitempty"`
	PostalAddress      string           `json:"postal_address,omitempty"`
	PostalCode         string           `json:"postal_code,omitempty"`
	Town               string           `json:"town,omitempty"`
	MaritalStatus      MaritalStatus    `json:"marital_status,omitempty"`
	SpouseName         string           `json:"spouse_name,omitempty"`
	NextOfKin          string           `json:"next_of_kin,omitempty"`
	NextOfKinPhone     string           `json:"next_of_kin_phone,omitempty"`
	BankName           string           `json:"bank_name,omitempty"`
	BankBranch         string           `json:"bank_branch,omitempty"`
	BankAccount        string           `json:"bank_account,omitempty"`
	PayrollNo          string           `json:"payroll_no,omitempty"`
	Designation        string           `json:"designation,omitempty"`
	Department         string           `json:"department,omitempty"`
	SponsorID          string           `json:"sponsor_id,omitempty"`
	DateFirstAppt      time.Time        `json:"date_first_appt,omitempty"`
	DateJoinedScheme   time.Time        `json:"date_joined_scheme"`
	ExpectedRetirement time.Time        `json:"expected_retirement,omitempty"`
	MembershipStatus   MembershipStatus `json:"membership_status"`
	BasicSalary        int64            `json:"basic_salary"`
	AccountBalance     int64            `json:"account_balance"`
	// Withdrawal tracking
	TotalWithdrawals   int64     `json:"total_withdrawals,omitempty"` // Total amount withdrawn
	LastWithdrawalDate time.Time `json:"last_withdrawal_date,omitempty"`
	// Member authentication
	PIN string `json:"pin,omitempty"` // Member PIN for authentication
	// Biometrics support
	Photograph      string `json:"photograph,omitempty"`       // Scanned photograph/fingerprints
	FingerprintData string `json:"fingerprint_data,omitempty"` // Fingerprint template data
	// Family and dependents
	ChildrenUnder21Count int `json:"children_under_21_count,omitempty"` // Number of children under 21
	// Membership card details
	MembershipCardIssueDate time.Time `json:"membership_card_issue_date,omitempty"`
	MembershipCardStatus    string    `json:"membership_card_status,omitempty"` // issue, Not issued, returned, lost
	// Sponsor history
	PreviousSponsors []string `json:"previous_sponsors,omitempty"` // History of previous sponsors
	// Cessation/transfer details
	CessationDate   time.Time `json:"cessation_date,omitempty"`
	CessationReason string    `json:"cessation_reason,omitempty"` // User defined coded reasons
	// Tax exemption
	TaxExemptReason     string    `json:"tax_exempt_reason,omitempty"`      // Reason for tax exemption
	TaxExemptAttachment string    `json:"tax_exempt_attachment,omitempty"`  // Attachment for tax exemption certificate
	TaxExemptCutoffDate time.Time `json:"tax_exempt_cutoff_date,omitempty"` // Specified cutoff date
	// Contribution rates
	MemberContributionRate  float64 `json:"member_contribution_rate,omitempty"`  // Member contribution rate (%)
	SponsorContributionRate float64 `json:"sponsor_contribution_rate,omitempty"` // Sponsor contribution rate (%)
	// Medical limits
	InpatientLimit   int64     `json:"inpatient_limit,omitempty"`  // Maximum inpatient coverage
	OutpatientLimit  int64     `json:"outpatient_limit,omitempty"` // Maximum outpatient coverage
	LastContribution time.Time `json:"last_contribution,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Validate checks if a Member is valid
func (m *Member) Validate() error {
	if m.FirstName == "" {
		return NewValidationError("first_name", "first name is required")
	}
	if m.LastName == "" {
		return NewValidationError("last_name", "last name is required")
	}
	if m.MemberNo == "" {
		return NewValidationError("member_no", "member number is required")
	}
	if m.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if m.DateOfBirth.IsZero() {
		return NewValidationError("date_of_birth", "date of birth is required")
	}
	if m.DateJoinedScheme.IsZero() {
		return NewValidationError("date_joined_scheme", "date joined scheme is required")
	}
	// Validate date of death if set (must not be before date of birth)
	if !m.DateOfDeath.IsZero() && m.DateOfDeath.Before(m.DateOfBirth) {
		return NewValidationError("date_of_death", "date of death cannot be before date of birth")
	}
	// Validate withdrawal tracking
	if m.TotalWithdrawals < 0 {
		return NewValidationError("total_withdrawals", "total withdrawals cannot be negative")
	}
	if !m.LastWithdrawalDate.IsZero() && m.LastWithdrawalDate.Before(m.DateOfBirth) {
		return NewValidationError("last_withdrawal_date", "last withdrawal date cannot be before date of birth")
	}
	// Validate PIN if set (should be 4-6 digits)
	if m.PIN != "" {
		if len(m.PIN) < 4 || len(m.PIN) > 6 {
			return NewValidationError("pin", "PIN must be between 4 and 6 digits")
		}
		// Check if PIN contains only digits
		for _, char := range m.PIN {
			if char < '0' || char > '9' {
				return NewValidationError("pin", "PIN must contain only digits")
			}
		}
	}
	// Validate biometrics fields
	if m.Photograph != "" && m.FingerprintData == "" {
		return NewValidationError("fingerprint_data", "fingerprint data is required when photograph is provided")
	}
	if m.FingerprintData != "" && m.Photograph == "" {
		return NewValidationError("photograph", "photograph is required when fingerprint data is provided")
	}
	// Validate children under 21 count
	if m.ChildrenUnder21Count < 0 {
		return NewValidationError("children_under_21_count", "children under 21 count cannot be negative")
	}
	// Validate membership card status
	if m.MembershipCardStatus != "" {
		validStatuses := map[string]bool{
			"issue": true, "Not issued": true, "returned": true, "lost": true,
		}
		if !validStatuses[m.MembershipCardStatus] {
			return NewValidationError("membership_card_status", "invalid membership card status")
		}
	}
	// Validate previous sponsors (no validation needed for slice of strings)
	// Validate cessation date if set
	if !m.CessationDate.IsZero() && m.CessationDate.Before(m.DateOfBirth) {
		return NewValidationError("cessation_date", "cessation date cannot be before date of birth")
	}
	// Validate tax exemption fields
	if m.TaxExemptReason != "" && m.TaxExemptCutoffDate.IsZero() {
		return NewValidationError("tax_exempt_cutoff_date", "cutoff date is required when tax exemption reason is provided")
	}
	if m.TaxExemptCutoffDate != (time.Time{}) && m.TaxExemptReason == "" {
		return NewValidationError("tax_exempt_reason", "reason is required when tax exemption cutoff date is provided")
	}
	if !m.TaxExemptCutoffDate.IsZero() && m.TaxExemptCutoffDate.Before(m.DateOfBirth) {
		return NewValidationError("tax_exempt_cutoff_date", "tax exemption cutoff date cannot be before date of birth")
	}
	// Validate contribution rates
	if m.MemberContributionRate < 0 || m.MemberContributionRate > 100 {
		return NewValidationError("member_contribution_rate", "member contribution rate must be between 0 and 100")
	}
	if m.SponsorContributionRate < 0 || m.SponsorContributionRate > 100 {
		return NewValidationError("sponsor_contribution_rate", "sponsor contribution rate must be between 0 and 100")
	}
	// Validate medical limits if set
	if m.InpatientLimit < 0 {
		return NewValidationError("inpatient_limit", "inpatient limit cannot be negative")
	}
	if m.OutpatientLimit < 0 {
		return NewValidationError("outpatient_limit", "outpatient limit cannot be negative")
	}
	return nil
}

// CurrentAge calculates the member's current age
func (m *Member) CurrentAge() int {
	now := time.Now()
	age := now.Year() - m.DateOfBirth.Year()
	if now.YearDay() < m.DateOfBirth.YearDay() {
		age--
	}
	return age
}

// YearsToRetirement calculates years until expected retirement
func (m *Member) YearsToRetirement() int {
	if m.ExpectedRetirement.IsZero() {
		return 0
	}
	return m.ExpectedRetirement.Year() - time.Now().Year()
}

// IsTaxExempt checks if member qualifies for tax exemption based on age and explicit exemption
func (m *Member) IsTaxExempt() bool {
	// Check age-based exemption (65+)
	if m.CurrentAge() >= 65 {
		return true
	}
	// Check explicit tax exemption with valid cutoff date
	if m.TaxExemptReason != "" && !m.TaxExemptCutoffDate.IsZero() {
		return time.Now().Before(m.TaxExemptCutoffDate)
	}
	return false
}

// IsTaxExemptByAge checks if member qualifies for tax exemption based on age only (65+)
func (m *Member) IsTaxExemptByAge() bool {
	return m.CurrentAge() >= 65
}

// HasExplicitTaxExempt checks if member has an explicit tax exemption with reason and cutoff date
func (m *Member) HasExplicitTaxExempt() bool {
	return m.TaxExemptReason != "" && !m.TaxExemptCutoffDate.IsZero() &&
		time.Now().Before(m.TaxExemptCutoffDate)
}

// GetYearsToRetirement calculates years until expected retirement
func (m *Member) GetYearsToRetirement() int {
	if m.ExpectedRetirement.IsZero() {
		return 0
	}
	return m.ExpectedRetirement.Year() - time.Now().Year()
}

// GetAgeAtDeath calculates age at death if date of death is set
func (m *Member) GetAgeAtDeath() int {
	if m.DateOfDeath.IsZero() {
		return 0
	}
	year := m.DateOfDeath.Year() - m.DateOfBirth.Year()
	if m.DateOfDeath.YearDay() < m.DateOfBirth.YearDay() {
		year--
	}
	return year
}

// GetMembershipDuration returns how long the member has been in the scheme
func (m *Member) GetMembershipDuration() int {
	if m.DateJoinedScheme.IsZero() {
		return 0
	}
	years := time.Now().Year() - m.DateJoinedScheme.Year()
	if time.Now().YearDay() < m.DateJoinedScheme.YearDay() {
		years--
	}
	return years
}

// IsActive checks if member status is active
func (m *Member) IsActive() bool {
	return m.MembershipStatus == StatusActive
}

// IsDeceased checks if member status is deceased
func (m *Member) IsDeceased() bool {
	return m.MembershipStatus == StatusDeceased
}

// HasWithdrawals checks if member has made any withdrawals
func (m *Member) HasWithdrawals() bool {
	return m.TotalWithdrawals > 0
}

// GetLastWithdrawalDaysAgo returns days since last withdrawal
func (m *Member) GetLastWithdrawalDaysAgo() int {
	if m.LastWithdrawalDate.IsZero() {
		return -1 // Indicates no withdrawal
	}
	return int(time.Since(m.LastWithdrawalDate).Hours() / 24)
}

// GetDaysToTaxExemptCutoff returns days until tax exemption cutoff date
func (m *Member) GetDaysToTaxExemptCutoff() int {
	if m.TaxExemptCutoffDate.IsZero() {
		return -1 // Indicates no cutoff date
	}
	if time.Now().After(m.TaxExemptCutoffDate) {
		return 0 // Already expired
	}
	return int(m.TaxExemptCutoffDate.Sub(time.Now()).Hours() / 24)
}

// HasBiometrics checks if member has biometric data (photograph and fingerprint)
func (m *Member) HasBiometrics() bool {
	return m.Photograph != "" && m.FingerprintData != ""
}

// GetDependentCount returns number of dependents (children under 21)
func (m *Member) GetDependentCount() int {
	return m.ChildrenUnder21Count
}

// IsMembershipCardValid checks if membership card status is valid
func (m *Member) IsMembershipCardValid() bool {
	return m.MembershipCardStatus == "issue"
}

// GetPreviousSponsorsCount returns number of previous sponsors
func (m *Member) GetPreviousSponsorsCount() int {
	return len(m.PreviousSponsors)
}

// HasCessationDetails checks if member has cessation/transfer details
func (m *Member) HasCessationDetails() bool {
	return !m.CessationDate.IsZero() || m.CessationReason != ""
}
