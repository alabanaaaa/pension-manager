package domain

import (
	"time"
)

// Sponsor represents an employer/organization that remits contributions
type Sponsor struct {
	ID            string    `json:"id"`
	SchemeID      string    `json:"scheme_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	ContactPerson string    `json:"contact_person,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Email         string    `json:"email,omitempty"`
	Address       string    `json:"address,omitempty"`
	PayMode       string    `json:"pay_mode,omitempty"` // cheque, bank_transfer, cash, mpesa, standing_order
	BankName      string    `json:"bank_name,omitempty"`
	BankBranch    string    `json:"bank_branch,omitempty"`
	BankAccount   string    `json:"bank_account,omitempty"`
	TotalMembers  int       `json:"total_members"`
	Status        string    `json:"status"` // active, inactive
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate checks if a Sponsor is valid
func (s *Sponsor) Validate() error {
	if s.Name == "" {
		return NewValidationError("name", "sponsor name is required")
	}
	if s.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if s.Code == "" {
		return NewValidationError("code", "sponsor code is required")
	}
	return nil
}

// ContributionSchedule represents a monthly remittance from a sponsor
type ContributionSchedule struct {
	ID                  string    `json:"id"`
	SponsorID           string    `json:"sponsor_id"`
	SchemeID            string    `json:"scheme_id"`
	Period              time.Time `json:"period"`
	TotalEmployees      int       `json:"total_employees"`
	TotalAmount         int64     `json:"total_amount"`
	PrevEmployees       int       `json:"prev_employees,omitempty"`
	PrevAmount          int64     `json:"prev_amount,omitempty"`
	EmployeeDiff        int       `json:"employee_diff,omitempty"`
	AmountDiff          int64     `json:"amount_diff,omitempty"`
	Status              string    `json:"status"` // pending, balanced, on_hold, posted
	ReconciliationNotes string    `json:"reconciliation_notes,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	PostedAt            time.Time `json:"posted_at,omitempty"`
}

// Validate checks if a ContributionSchedule is valid
func (cs *ContributionSchedule) Validate() error {
	if cs.SponsorID == "" {
		return NewValidationError("sponsor_id", "sponsor is required")
	}
	if cs.Period.IsZero() {
		return NewValidationError("period", "period is required")
	}
	if cs.TotalEmployees < 0 {
		return NewValidationError("total_employees", "total employees cannot be negative")
	}
	if cs.TotalAmount < 0 {
		return NewValidationError("total_amount", "total amount cannot be negative")
	}
	return nil
}
