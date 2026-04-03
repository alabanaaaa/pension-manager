package hospital

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pension-manager/core/domain"
	"pension-manager/internal/db"
)

// HospitalService manages hospital accounts and medical expenditures
type HospitalService struct {
	db *db.DB
}

// ExpenditureAlerts represents alert counts for medical expenditures
type ExpenditureAlerts struct {
	PendingBills       int   `json:"pending_bills"`
	HighUrgencyBills   int   `json:"high_urgency_bills"`
	MediumUrgencyBills int   `json:"medium_urgency_bills"`
	LowUrgencyBills    int   `json:"low_urgency_bills"`
	TotalPendingAmount int64 `json:"total_pending_amount"`
}

// NewHospitalService creates a new hospital service
func NewHospitalService(db *db.DB) *HospitalService {
	return &HospitalService{db: db}
}

// CreateHospital creates a new hospital account
func (s *HospitalService) CreateHospital(ctx context.Context, hospital *domain.Hospital) error {
	if hospital == nil {
		return errors.New("hospital cannot be nil")
	}
	if err := hospital.Validate(); err != nil {
		return err
	}

	now := time.Now()
	hospital.CreatedAt = now
	hospital.UpdatedAt = now

	query := `
		INSERT INTO hospitals (id, scheme_id, name, address, phone, email, license_number, 
		                       account_balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := s.db.ExecContext(ctx, query,
		hospital.ID, hospital.SchemeID, hospital.Name, hospital.Address,
		hospital.Phone, hospital.Email, hospital.LicenseNumber,
		hospital.AccountBalance, hospital.Status, hospital.CreatedAt, hospital.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create hospital: %w", err)
	}
	return nil
}

// GetHospital retrieves a hospital by ID
func (s *HospitalService) GetHospital(ctx context.Context, hospitalID string) (*domain.Hospital, error) {
	if hospitalID == "" {
		return nil, errors.New("hospital ID cannot be empty")
	}

	query := `
		SELECT id, scheme_id, name, address, phone, email, license_number,
		       account_balance, status, created_at, updated_at
		FROM hospitals WHERE id = $1
	`
	h := &domain.Hospital{}
	err := s.db.QueryRowContext(ctx, query, hospitalID).Scan(
		&h.ID, &h.SchemeID, &h.Name, &h.Address, &h.Phone, &h.Email,
		&h.LicenseNumber, &h.AccountBalance, &h.Status, &h.CreatedAt, &h.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get hospital: %w", err)
	}
	return h, nil
}

// UpdateHospital updates an existing hospital
func (s *HospitalService) UpdateHospital(ctx context.Context, hospital *domain.Hospital) error {
	if hospital == nil {
		return errors.New("hospital cannot be nil")
	}
	if hospital.ID == "" {
		return errors.New("hospital ID is required")
	}
	if err := hospital.Validate(); err != nil {
		return err
	}

	hospital.UpdatedAt = time.Now()

	query := `
		UPDATE hospitals SET name = $1, address = $2, phone = $3, email = $4,
		                     license_number = $5, account_balance = $6, status = $7,
		                     updated_at = $8
		WHERE id = $9
	`
	result, err := s.db.ExecContext(ctx, query,
		hospital.Name, hospital.Address, hospital.Phone, hospital.Email,
		hospital.LicenseNumber, hospital.AccountBalance, hospital.Status,
		hospital.UpdatedAt, hospital.ID,
	)
	if err != nil {
		return fmt.Errorf("update hospital: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hospital not found")
	}
	return nil
}

// ListHospitals returns all hospitals for a scheme
func (s *HospitalService) ListHospitals(ctx context.Context, schemeID string) ([]*domain.Hospital, error) {
	if schemeID == "" {
		return nil, errors.New("scheme ID cannot be empty")
	}

	query := `
		SELECT id, scheme_id, name, address, phone, email, license_number,
		       account_balance, status, created_at, updated_at
		FROM hospitals WHERE scheme_id = $1 ORDER BY name
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("list hospitals: %w", err)
	}
	defer rows.Close()

	var hospitals []*domain.Hospital
	for rows.Next() {
		h := &domain.Hospital{}
		if err := rows.Scan(
			&h.ID, &h.SchemeID, &h.Name, &h.Address, &h.Phone, &h.Email,
			&h.LicenseNumber, &h.AccountBalance, &h.Status, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan hospital: %w", err)
		}
		hospitals = append(hospitals, h)
	}
	return hospitals, rows.Err()
}

// CreateMedicalLimit sets medical limits for a member
func (s *HospitalService) CreateMedicalLimit(ctx context.Context, limit *domain.MedicalLimit) error {
	if limit == nil {
		return errors.New("medical limit cannot be nil")
	}
	if err := limit.Validate(); err != nil {
		return err
	}

	now := time.Now()
	limit.CreatedAt = now
	limit.UpdatedAt = now

	query := `
		INSERT INTO medical_limits (id, member_id, scheme_id, inpatient_limit, outpatient_limit,
		                            period, effective_date, expiry_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.ExecContext(ctx, query,
		limit.ID, limit.MemberID, limit.SchemeID, limit.InpatientLimit, limit.OutpatientLimit,
		limit.Period, limit.EffectiveDate, limit.ExpiryDate, limit.CreatedAt, limit.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create medical limit: %w", err)
	}
	return nil
}

// GetMedicalLimit retrieves medical limits for a member
func (s *HospitalService) GetMedicalLimit(ctx context.Context, memberID string) (*domain.MedicalLimit, error) {
	if memberID == "" {
		return nil, errors.New("member ID cannot be empty")
	}

	query := `
		SELECT id, member_id, scheme_id, inpatient_limit, outpatient_limit,
		       period, effective_date, expiry_date, created_at, updated_at
		FROM medical_limits WHERE member_id = $1 ORDER BY effective_date DESC LIMIT 1
	`
	limit := &domain.MedicalLimit{}
	err := s.db.QueryRowContext(ctx, query, memberID).Scan(
		&limit.ID, &limit.MemberID, &limit.SchemeID, &limit.InpatientLimit, &limit.OutpatientLimit,
		&limit.Period, &limit.EffectiveDate, &limit.ExpiryDate, &limit.CreatedAt, &limit.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get medical limit: %w", err)
	}
	return limit, nil
}

// RecordMedicalExpenditure records a medical expense
func (s *HospitalService) RecordMedicalExpenditure(ctx context.Context, expenditure *domain.MedicalExpenditure) error {
	if expenditure == nil {
		return errors.New("medical expenditure cannot be nil")
	}
	if err := expenditure.Validate(); err != nil {
		return err
	}

	now := time.Now()
	expenditure.DateSubmitted = now
	expenditure.CreatedAt = now
	expenditure.UpdatedAt = now

	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		// Insert expenditure record
		query := `
			INSERT INTO medical_expenditures (id, member_id, scheme_id, hospital_id, date_of_service,
			                                  date_submitted, service_type, description, amount_charged,
			                                  amount_covered, member_responsibility, status,
			                                  invoice_number, receipt_number, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`
		_, err := tx.ExecContext(ctx, query,
			expenditure.ID, expenditure.MemberID, expenditure.SchemeID, expenditure.HospitalID,
			expenditure.DateOfService, expenditure.DateSubmitted, expenditure.ServiceType,
			expenditure.Description, expenditure.AmountCharged, expenditure.AmountCovered,
			expenditure.MemberResponsibility, expenditure.Status, expenditure.InvoiceNumber,
			expenditure.ReceiptNumber, expenditure.CreatedAt, expenditure.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("record medical expenditure: %w", err)
		}

		// Update hospital balance if payment is made
		if expenditure.Status == "paid" && expenditure.HospitalID != "" {
			updateQuery := `
				UPDATE hospitals SET account_balance = account_balance + $1, updated_at = $2
				WHERE id = $3
			`
			_, err = tx.ExecContext(ctx, updateQuery, expenditure.AmountCovered, now, expenditure.HospitalID)
			if err != nil {
				return fmt.Errorf("update hospital balance: %w", err)
			}
		}

		return nil
	})
}

// GetPendingBills returns medical expenditures that are pending payment for over 45 days
func (s *HospitalService) GetPendingBills(ctx context.Context, schemeID string) ([]*domain.MedicalExpenditure, error) {
	if schemeID == "" {
		return nil, errors.New("scheme ID cannot be empty")
	}

	query := `
		SELECT id, member_id, scheme_id, hospital_id, date_of_service, date_submitted,
		       service_type, description, amount_charged, amount_covered, member_responsibility,
		       status, invoice_number, receipt_number, created_at, updated_at
		FROM medical_expenditures
		WHERE scheme_id = $1
		  AND status IN ('submitted', 'approved')
		  AND date_submitted < NOW() - INTERVAL '45 days'
		ORDER BY date_submitted ASC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("get pending bills: %w", err)
	}
	defer rows.Close()

	var bills []*domain.MedicalExpenditure
	for rows.Next() {
		exp := &domain.MedicalExpenditure{}
		if err := rows.Scan(
			&exp.ID, &exp.MemberID, &exp.SchemeID, &exp.HospitalID, &exp.DateOfService,
			&exp.DateSubmitted, &exp.ServiceType, &exp.Description, &exp.AmountCharged,
			&exp.AmountCovered, &exp.MemberResponsibility, &exp.Status, &exp.InvoiceNumber,
			&exp.ReceiptNumber, &exp.CreatedAt, &exp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan expenditure: %w", err)
		}
		bills = append(bills, exp)
	}
	return bills, rows.Err()
}

// GetExpenditureAlerts returns various alerts related to medical expenditures
func (s *HospitalService) GetExpenditureAlerts(ctx context.Context, schemeID string) (*ExpenditureAlerts, error) {
	if schemeID == "" {
		return nil, errors.New("scheme ID cannot be empty")
	}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('submitted', 'approved')) as pending_bills,
			COUNT(*) FILTER (WHERE status IN ('submitted', 'approved') AND date_submitted < NOW() - INTERVAL '60 days') as high_urgency,
			COUNT(*) FILTER (WHERE status IN ('submitted', 'approved') AND date_submitted >= NOW() - INTERVAL '60 days' AND date_submitted < NOW() - INTERVAL '45 days') as medium_urgency,
			COUNT(*) FILTER (WHERE status IN ('submitted', 'approved') AND date_submitted >= NOW() - INTERVAL '45 days') as low_urgency,
			COALESCE(SUM(amount_charged) FILTER (WHERE status IN ('submitted', 'approved')), 0) as total_pending_amount
		FROM medical_expenditures
		WHERE scheme_id = $1
	`
	alerts := &ExpenditureAlerts{}
	err := s.db.QueryRowContext(ctx, query, schemeID).Scan(
		&alerts.PendingBills, &alerts.HighUrgencyBills, &alerts.MediumUrgencyBills,
		&alerts.LowUrgencyBills, &alerts.TotalPendingAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("get expenditure alerts: %w", err)
	}
	return alerts, nil
}

// UpdateHospitalBalance updates the account balance for a hospital
func (s *HospitalService) UpdateHospitalBalance(ctx context.Context, hospitalID string, amount int64) error {
	if hospitalID == "" {
		return errors.New("hospital ID cannot be empty")
	}

	query := `
		UPDATE hospitals SET account_balance = account_balance + $1, updated_at = $2
		WHERE id = $3
	`
	result, err := s.db.ExecContext(ctx, query, amount, time.Now(), hospitalID)
	if err != nil {
		return fmt.Errorf("update hospital balance: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hospital not found")
	}
	return nil
}
