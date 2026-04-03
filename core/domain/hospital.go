package domain

import (
	"time"
)

// Hospital represents a medical facility associated with the pension scheme
type Hospital struct {
	ID             string    `json:"id"`
	SchemeID       string    `json:"scheme_id"`
	Name           string    `json:"name"`
	Address        string    `json:"address,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	LicenseNumber  string    `json:"license_number,omitempty"`
	AccountBalance int64     `json:"account_balance"` // Balance in smallest currency unit (cents)
	Status         string    `json:"status"`          // active, inactive, suspended
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validate checks if a Hospital is valid
func (h *Hospital) Validate() error {
	if h.Name == "" {
		return NewValidationError("name", "hospital name is required")
	}
	if h.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if h.AccountBalance < 0 {
		return NewValidationError("account_balance", "account balance cannot be negative")
	}
	return nil
}

// MedicalLimit represents medical coverage limits for a member
type MedicalLimit struct {
	ID              string    `json:"id"`
	MemberID        string    `json:"member_id"`
	SchemeID        string    `json:"scheme_id"`
	InpatientLimit  int64     `json:"inpatient_limit"`  // Maximum inpatient coverage per period
	OutpatientLimit int64     `json:"outpatient_limit"` // Maximum outpatient coverage per period
	Period          string    `json:"period"`           // annual, monthly
	EffectiveDate   time.Time `json:"effective_date"`
	ExpiryDate      time.Time `json:"expiry_date,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Validate checks if a MedicalLimit is valid
func (m *MedicalLimit) Validate() error {
	if m.MemberID == "" {
		return NewValidationError("member_id", "member is required")
	}
	if m.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if m.InpatientLimit < 0 {
		return NewValidationError("inpatient_limit", "inpatient limit cannot be negative")
	}
	if m.OutpatientLimit < 0 {
		return NewValidationError("outpatient_limit", "outpatient limit cannot be negative")
	}
	if m.Period == "" {
		return NewValidationError("period", "period is required")
	}
	if m.EffectiveDate.IsZero() {
		return NewValidationError("effective_date", "effective date is required")
	}
	if m.ExpiryDate != (time.Time{}) && m.ExpiryDate.Before(m.EffectiveDate) {
		return NewValidationError("expiry_date", "expiry date must be after effective date")
	}
	return nil
}

// MedicalExpenditure represents a medical expense claim or payment
type MedicalExpenditure struct {
	ID                   string    `json:"id"`
	MemberID             string    `json:"member_id"`
	SchemeID             string    `json:"scheme_id"`
	HospitalID           string    `json:"hospital_id,omitempty"`
	DateOfService        time.Time `json:"date_of_service"`
	DateSubmitted        time.Time `json:"date_submitted"`
	ServiceType          string    `json:"service_type"` // inpatient, outpatient, pharmacy, etc.
	Description          string    `json:"description,omitempty"`
	AmountCharged        int64     `json:"amount_charged"`        // Amount charged by provider
	AmountCovered        int64     `json:"amount_covered"`        // Amount covered by scheme
	MemberResponsibility int64     `json:"member_responsibility"` // Amount member needs to pay
	Status               string    `json:"status"`                // submitted, approved, rejected, paid
	InvoiceNumber        string    `json:"invoice_number,omitempty"`
	ReceiptNumber        string    `json:"receipt_number,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Validate checks if a MedicalExpenditure is valid
func (m *MedicalExpenditure) Validate() error {
	if m.MemberID == "" {
		return NewValidationError("member_id", "member is required")
	}
	if m.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if m.DateOfService.IsZero() {
		return NewValidationError("date_of_service", "date of service is required")
	}
	if m.AmountCharged < 0 {
		return NewValidationError("amount_charged", "amount charged cannot be negative")
	}
	if m.AmountCovered < 0 {
		return NewValidationError("amount_covered", "amount covered cannot be negative")
	}
	if m.MemberResponsibility < 0 {
		return NewValidationError("member_responsibility", "member responsibility cannot be negative")
	}
	if m.AmountCharged < m.AmountCovered+m.MemberResponsibility {
		return NewValidationError("amount_charged", "amount charged must be greater than or equal to amount covered plus member responsibility")
	}
	return nil
}
