package domain

import "time"

// ContributionType represents the type of contribution
type ContributionType string

const (
	EmployeeContribution ContributionType = "employee"
	EmployerContribution ContributionType = "employer"
	AVCContribution      ContributionType = "avc" // Additional Voluntary Contribution
)

// PaymentMethod represents how a contribution was paid
type PaymentMethod string

const (
	PaymentMpesa         PaymentMethod = "mpesa"
	PaymentBankTransfer  PaymentMethod = "bank_transfer"
	PaymentCheque        PaymentMethod = "cheque"
	PaymentCash          PaymentMethod = "cash"
	PaymentStandingOrder PaymentMethod = "standing_order"
)

// ContributionStatus represents the status of a contribution
type ContributionStatus string

const (
	StatusPending    ContributionStatus = "pending"
	StatusConfirmed  ContributionStatus = "confirmed"
	StatusReconciled ContributionStatus = "reconciled"
	StatusOnHold     ContributionStatus = "on_hold"
	StatusRejected   ContributionStatus = "rejected"
)

// Contribution represents a pension contribution
type Contribution struct {
	ID             string             `json:"id"`
	MemberID       string             `json:"member_id"`
	SchemeID       string             `json:"scheme_id"`
	SponsorID      string             `json:"sponsor_id,omitempty"`
	Period         time.Time          `json:"period"`
	EmployeeAmount int64              `json:"employee_amount"`
	EmployerAmount int64              `json:"employer_amount"`
	AVCAmount      int64              `json:"avc_amount"`
	TotalAmount    int64              `json:"total_amount"`
	PaymentMethod  PaymentMethod      `json:"payment_method,omitempty"`
	PaymentRef     string             `json:"payment_ref,omitempty"`
	ReceiptNo      string             `json:"receipt_no,omitempty"`
	Status         ContributionStatus `json:"status"`
	Registered     bool               `json:"registered"`
	Notes          string             `json:"notes,omitempty"`
	CreatedBy      string             `json:"created_by,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	ConfirmedAt    time.Time          `json:"confirmed_at,omitempty"`
}

// Total calculates the total contribution amount
func (c *Contribution) Total() int64 {
	return c.EmployeeAmount + c.EmployerAmount + c.AVCAmount
}

// Validate checks if a Contribution is valid
func (c *Contribution) Validate() error {
	if c.MemberID == "" {
		return NewValidationError("member_id", "member is required")
	}
	if c.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if c.Period.IsZero() {
		return NewValidationError("period", "period is required")
	}
	if c.EmployeeAmount < 0 {
		return NewValidationError("employee_amount", "employee amount cannot be negative")
	}
	if c.EmployerAmount < 0 {
		return NewValidationError("employer_amount", "employer amount cannot be negative")
	}
	if c.AVCAmount < 0 {
		return NewValidationError("avc_amount", "AVC amount cannot be negative")
	}
	if c.Total() == 0 {
		return NewValidationError("total", "total contribution must be greater than zero")
	}
	return nil
}
