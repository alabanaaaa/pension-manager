package hospital

import (
	"testing"
	"time"

	"pension-manager/core/domain"
)

func TestHospitalValidate(t *testing.T) {
	// Valid hospital
	h := &domain.Hospital{
		Name:     "Nairobi Hospital",
		SchemeID: "scheme-001",
		Status:   "active",
	}
	if err := h.Validate(); err != nil {
		t.Errorf("Expected valid hospital, got error: %v", err)
	}

	// Missing name
	h.Name = ""
	if err := h.Validate(); err == nil {
		t.Error("Expected error for missing name")
	}
	h.Name = "Nairobi Hospital"

	// Missing scheme
	h.SchemeID = ""
	if err := h.Validate(); err == nil {
		t.Error("Expected error for missing scheme")
	}
	h.SchemeID = "scheme-001"

	// Negative balance
	h.AccountBalance = -1000
	if err := h.Validate(); err == nil {
		t.Error("Expected error for negative balance")
	}
}

func TestMedicalLimitValidate(t *testing.T) {
	now := time.Now()
	future := now.Add(365 * 24 * time.Hour)

	// Valid medical limit
	ml := &domain.MedicalLimit{
		MemberID:        "member-001",
		SchemeID:        "scheme-001",
		InpatientLimit:  500000,
		OutpatientLimit: 200000,
		Period:          "annual",
		EffectiveDate:   now,
		ExpiryDate:      future,
	}
	if err := ml.Validate(); err != nil {
		t.Errorf("Expected valid medical limit, got error: %v", err)
	}

	// Missing member
	ml.MemberID = ""
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for missing member")
	}
	ml.MemberID = "member-001"

	// Negative inpatient limit
	ml.InpatientLimit = -1000
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for negative inpatient limit")
	}
	ml.InpatientLimit = 500000

	// Negative outpatient limit
	ml.OutpatientLimit = -1000
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for negative outpatient limit")
	}
	ml.OutpatientLimit = 200000

	// Missing period
	ml.Period = ""
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for missing period")
	}
	ml.Period = "annual"

	// Missing effective date
	ml.EffectiveDate = time.Time{}
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for missing effective date")
	}
	ml.EffectiveDate = now

	// Expiry before effective date
	ml.ExpiryDate = now.Add(-24 * time.Hour)
	if err := ml.Validate(); err == nil {
		t.Error("Expected error for expiry before effective date")
	}
}

func TestMedicalExpenditureValidate(t *testing.T) {
	now := time.Now()

	// Valid expenditure
	me := &domain.MedicalExpenditure{
		MemberID:             "member-001",
		SchemeID:             "scheme-001",
		DateOfService:        now,
		ServiceType:          "outpatient",
		AmountCharged:        500000,
		AmountCovered:        400000,
		MemberResponsibility: 100000,
		Status:               "submitted",
	}
	if err := me.Validate(); err != nil {
		t.Errorf("Expected valid expenditure, got error: %v", err)
	}

	// Missing member
	me.MemberID = ""
	if err := me.Validate(); err == nil {
		t.Error("Expected error for missing member")
	}
	me.MemberID = "member-001"

	// Missing scheme
	me.SchemeID = ""
	if err := me.Validate(); err == nil {
		t.Error("Expected error for missing scheme")
	}
	me.SchemeID = "scheme-001"

	// Missing date of service
	me.DateOfService = time.Time{}
	if err := me.Validate(); err == nil {
		t.Error("Expected error for missing date of service")
	}
	me.DateOfService = now

	// Negative amount charged
	me.AmountCharged = -1000
	if err := me.Validate(); err == nil {
		t.Error("Expected error for negative amount charged")
	}
	me.AmountCharged = 500000

	// Negative amount covered
	me.AmountCovered = -1000
	if err := me.Validate(); err == nil {
		t.Error("Expected error for negative amount covered")
	}
	me.AmountCovered = 400000

	// Negative member responsibility
	me.MemberResponsibility = -1000
	if err := me.Validate(); err == nil {
		t.Error("Expected error for negative member responsibility")
	}
	me.MemberResponsibility = 100000

	// Amount charged less than covered + responsibility
	me.AmountCharged = 100000
	if err := me.Validate(); err == nil {
		t.Error("Expected error for amount charged < covered + responsibility")
	}
	me.AmountCharged = 500000
}

func TestExpenditureAlerts(t *testing.T) {
	alerts := &ExpenditureAlerts{
		PendingBills:       10,
		HighUrgencyBills:   3,
		MediumUrgencyBills: 4,
		LowUrgencyBills:    3,
		TotalPendingAmount: 5000000,
	}

	if alerts.PendingBills != 10 {
		t.Errorf("Expected 10 pending bills, got: %d", alerts.PendingBills)
	}
	if alerts.TotalPendingAmount != 5000000 {
		t.Errorf("Expected 5000000 total pending, got: %d", alerts.TotalPendingAmount)
	}
}
