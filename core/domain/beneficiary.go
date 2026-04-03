package domain

import (
	"time"
)

// Beneficiary represents a member's beneficiary
type Beneficiary struct {
	ID              string    `json:"id"`
	MemberID        string    `json:"member_id"`
	Name            string    `json:"name"`
	Relationship    string    `json:"relationship"`
	DateOfBirth     time.Time `json:"date_of_birth,omitempty"`
	IDNumber        string    `json:"id_number,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	PhysicalAddress string    `json:"physical_address,omitempty"`
	AllocationPct   float64   `json:"allocation_pct"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Validate checks if a Beneficiary is valid
func (b *Beneficiary) Validate() error {
	if b.Name == "" {
		return NewValidationError("name", "name is required")
	}
	if b.Relationship == "" {
		return NewValidationError("relationship", "relationship is required")
	}
	if b.AllocationPct < 0 || b.AllocationPct > 100 {
		return NewValidationError("allocation_pct", "allocation must be between 0 and 100")
	}
	return nil
}
