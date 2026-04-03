package domain

import (
	"time"
)

// ClaimType represents the type of claim/withdrawal
type ClaimType string

const (
	NormalRetirement    ClaimType = "normal_retirement"
	EarlyRetirement     ClaimType = "early_retirement"
	LateRetirement      ClaimType = "late_retirement"
	IllHealthRetirement ClaimType = "ill_health_retirement"
	DeathInService      ClaimType = "death_in_service"
	LeavingService      ClaimType = "leaving_service"
	DeferredRetirement  ClaimType = "deferred_retirement"
	MedicalClaim        ClaimType = "medical_claim"
	ExGratia            ClaimType = "ex_gratia"
)

// ClaimStatus represents the status of a claim
type ClaimStatus string

const (
	ClaimSubmitted    ClaimStatus = "submitted"
	ClaimResubmission ClaimStatus = "resubmission"
	ClaimUnderReview  ClaimStatus = "under_review"
	ClaimRejected     ClaimStatus = "rejected"
	ClaimAccepted     ClaimStatus = "accepted"
	ClaimPaid         ClaimStatus = "paid"
)

// Claim represents a benefit claim or withdrawal request
type Claim struct {
	ID              string           `json:"id"`
	MemberID        string           `json:"member_id"`
	SchemeID        string           `json:"scheme_id"`
	ClaimType       ClaimType        `json:"claim_type"`
	ClaimFormNo     string           `json:"claim_form_no,omitempty"`
	DateOfClaim     time.Time        `json:"date_of_claim"`
	DateOfLeaving   time.Time        `json:"date_of_leaving,omitempty"`
	LeavingReason   string           `json:"leaving_reason,omitempty"`
	Status          ClaimStatus      `json:"status"`
	RejectionReason string           `json:"rejection_reason,omitempty"`
	ExaminerID      string           `json:"examiner_id,omitempty"`
	SettlementDate  time.Time        `json:"settlement_date,omitempty"`
	ChequeRef       string           `json:"cheque_ref,omitempty"`
	ChequeDate      time.Time        `json:"cheque_date,omitempty"`
	Amount          int64            `json:"amount,omitempty"`
	PartialPayments []PartialPayment `json:"partial_payments,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	ReviewedAt      time.Time        `json:"reviewed_at,omitempty"`
	PaidAt          time.Time        `json:"paid_at,omitempty"`
}

// PartialPayment represents a partial payment against a claim
type PartialPayment struct {
	Date   time.Time `json:"date"`
	Amount int64     `json:"amount"`
	Ref    string    `json:"ref"`
}

// Validate checks if a Claim is valid
func (c *Claim) Validate() error {
	if c.MemberID == "" {
		return NewValidationError("member_id", "member is required")
	}
	if c.SchemeID == "" {
		return NewValidationError("scheme_id", "scheme is required")
	}
	if c.ClaimType == "" {
		return NewValidationError("claim_type", "claim type is required")
	}
	if c.DateOfClaim.IsZero() {
		return NewValidationError("date_of_claim", "date of claim is required")
	}
	return nil
}
