package claims

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"pension-manager/internal/db"
)

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

type ExitType string

const (
	ExitNormalRetirement      ExitType = "normal_retirement"
	ExitEarlyRetirement       ExitType = "early_retirement"
	ExitEarlyDeferred         ExitType = "early_deferred"
	ExitLateRetirement        ExitType = "late_retirement"
	ExitDeathInService        ExitType = "death_in_service"
	ExitLeavingService        ExitType = "leaving_service"
	ExitIllHealthRetirement   ExitType = "ill_health_retirement"
	ExitRefundOfContributions ExitType = "refund_of_contributions"
)

type ClaimStatus string

const (
	ClaimPending      ClaimStatus = "pending"
	ClaimUnderReview  ClaimStatus = "under_review"
	ClaimResubmission ClaimStatus = "resubmission"
	ClaimAccepted     ClaimStatus = "accepted"
	ClaimRejected     ClaimStatus = "rejected"
	ClaimPaid         ClaimStatus = "paid"
	ClaimCancelled    ClaimStatus = "cancelled"
)

type Claim struct {
	ID              string             `json:"id"`
	SchemeID        string             `json:"scheme_id"`
	MemberID        string             `json:"member_id"`
	ClaimType       ExitType           `json:"claim_type"`
	ClaimFormNo     string             `json:"claim_form_no"`
	DateOfClaim     time.Time          `json:"date_of_claim"`
	DateOfLeaving   *time.Time         `json:"date_of_leaving,omitempty"`
	LeavingReason   string             `json:"leaving_reason,omitempty"`
	Status          ClaimStatus        `json:"status"`
	RejectionReason *string            `json:"rejection_reason,omitempty"`
	ExaminerID      *string            `json:"examiner_id,omitempty"`
	ExaminerNotes   *string            `json:"examiner_notes,omitempty"`
	SettlementDate  *time.Time         `json:"settlement_date,omitempty"`
	ChequeRef       *string            `json:"cheque_ref,omitempty"`
	ChequeDate      *time.Time         `json:"cheque_date,omitempty"`
	Amount          int64              `json:"amount"`
	ApprovedAmount  *int64             `json:"approved_amount,omitempty"`
	PartialPayments []PartialPayment   `json:"partial_payments,omitempty"`
	Documents       []RequiredDocument `json:"documents,omitempty"`
	CreatedBy       string             `json:"created_by"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	ReviewedAt      *time.Time         `json:"reviewed_at,omitempty"`
	PaidAt          *time.Time         `json:"paid_at,omitempty"`
}

type PartialPayment struct {
	ID         string    `json:"id"`
	ClaimID    string    `json:"claim_id"`
	Amount     int64     `json:"amount"`
	PaymentRef string    `json:"payment_ref"`
	PaidAt     time.Time `json:"paid_at"`
	Notes      string    `json:"notes,omitempty"`
}

type RequiredDocument struct {
	ID         string     `json:"id"`
	ClaimID    string     `json:"claim_id"`
	DocType    string     `json:"doc_type"`
	FileName   string     `json:"file_name"`
	FilePath   string     `json:"file_path"`
	Uploaded   bool       `json:"uploaded"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
}

type ClaimSubmission struct {
	SchemeID      string     `json:"scheme_id"`
	MemberID      string     `json:"member_id"`
	ClaimType     ExitType   `json:"claim_type"`
	ClaimFormNo   string     `json:"claim_form_no"`
	DateOfClaim   time.Time  `json:"date_of_claim"`
	DateOfLeaving *time.Time `json:"date_of_leaving,omitempty"`
	LeavingReason string     `json:"leaving_reason,omitempty"`
	Amount        int64      `json:"amount"`
	Documents     []string   `json:"documents,omitempty"`
}

func (s *Service) SubmitClaim(ctx context.Context, submission *ClaimSubmission, createdBy string) (string, error) {
	var claimID string

	err := s.db.Transactional(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO claims (
				id, scheme_id, member_id, claim_type, claim_form_no, date_of_claim,
				date_of_leaving, leaving_reason, amount, status, created_by, created_at, updated_at
			) VALUES (
				uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, 'pending', $9, NOW(), NOW()
			) RETURNING id
		`
		err := tx.QueryRowContext(ctx, query,
			submission.SchemeID, submission.MemberID, submission.ClaimType,
			submission.ClaimFormNo, submission.DateOfClaim, submission.DateOfLeaving,
			submission.LeavingReason, submission.Amount, createdBy,
		).Scan(&claimID)
		if err != nil {
			return fmt.Errorf("insert claim: %w", err)
		}

		requiredDocs := s.getRequiredDocuments(submission.ClaimType)
		for _, docType := range requiredDocs {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO claim_documents (id, claim_id, doc_type, uploaded)
				VALUES (uuid_generate_v4(), $1, $2, false)
			`, claimID, docType)
			if err != nil {
				return fmt.Errorf("insert required doc: %w", err)
			}
		}

		s.logAuditEvent(ctx, tx, submission.SchemeID, "claim_submitted", claimID, createdBy,
			map[string]interface{}{
				"action":     "claim_submitted",
				"claim_type": submission.ClaimType,
				"amount":     submission.Amount,
			})

		return nil
	})

	if err != nil {
		return "", err
	}
	return claimID, nil
}

func (s *Service) getRequiredDocuments(claimType ExitType) []string {
	switch claimType {
	case ExitDeathInService:
		return []string{
			"death_certificate",
			"sponsor_clearance",
			"letters_of_administration",
			"relationship_affidavit",
			"marriage_certificate",
			"beneficiary_ids",
		}
	case ExitNormalRetirement, ExitLateRetirement:
		return []string{
			"retirement_letter",
			"clearance_certificate",
			"id_copy",
		}
	case ExitEarlyRetirement, ExitEarlyDeferred:
		return []string{
			"retirement_letter",
			"sponsor_clearance",
			"clearance_certificate",
			"id_copy",
		}
	case ExitIllHealthRetirement:
		return []string{
			"medical_certificate",
			"ill_health_evaluation",
			"retirement_letter",
			"sponsor_clearance",
			"id_copy",
		}
	case ExitLeavingService:
		return []string{
			"termination_letter",
			"clearance_certificate",
			"id_copy",
		}
	case ExitRefundOfContributions:
		return []string{
			"exit_letter",
			"clearance_certificate",
			"id_copy",
			"bank_details",
		}
	default:
		return []string{"id_copy"}
	}
}

func (s *Service) UpdateClaimStatus(ctx context.Context, claimID string, status ClaimStatus, examinerID, notes string) error {
	var schemeID string
	s.db.QueryRowContext(ctx, `SELECT scheme_id FROM claims WHERE id = $1`, claimID).Scan(&schemeID)

	updates := []string{"status = $1", "updated_at = NOW()"}
	args := []interface{}{status}

	if status == ClaimUnderReview {
		updates = append(updates, "examiner_id = $2", "examiner_notes = $3")
		args = append(args, examinerID, notes)
	}

	if status == ClaimAccepted || status == ClaimRejected {
		updates = append(updates, "reviewed_at = NOW()")
	}

	args = append(args, claimID)
	query := fmt.Sprintf("UPDATE claims SET %s WHERE id = $%d", joinStrings(updates, ", "), len(args))

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("claim not found")
	}

	s.logAuditEvent(ctx, nil, schemeID, "claim_status_updated", claimID, examinerID,
		map[string]interface{}{
			"action":     "claim_status_changed",
			"new_status": status,
			"notes":      notes,
		})

	return nil
}

func (s *Service) AcceptClaim(ctx context.Context, claimID, examinerID string, approvedAmount int64, notes string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE claims SET 
				status = 'accepted', 
				approved_amount = $1,
				examiner_id = $2,
				examiner_notes = $3,
				reviewed_at = NOW(),
				updated_at = NOW()
			WHERE id = $4
		`, approvedAmount, examinerID, notes, claimID)
		if err != nil {
			return err
		}

		var schemeID, memberID string
		tx.QueryRowContext(ctx, `SELECT scheme_id, member_id FROM claims WHERE id = $1`, claimID).Scan(&schemeID, &memberID)

		s.logAuditEvent(ctx, tx, schemeID, "claim_accepted", claimID, examinerID,
			map[string]interface{}{
				"action":          "claim_accepted",
				"approved_amount": approvedAmount,
			})

		return nil
	})
}

func (s *Service) RejectClaim(ctx context.Context, claimID, examinerID, reason string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE claims SET 
				status = 'rejected', 
				rejection_reason = $1,
				examiner_id = $2,
				reviewed_at = NOW(),
				updated_at = NOW()
			WHERE id = $3
		`, reason, examinerID, claimID)
		if err != nil {
			return err
		}

		var schemeID string
		tx.QueryRowContext(ctx, `SELECT scheme_id FROM claims WHERE id = $1`, claimID).Scan(&schemeID)

		s.logAuditEvent(ctx, tx, schemeID, "claim_rejected", claimID, examinerID,
			map[string]interface{}{
				"action": "claim_rejected",
				"reason": reason,
			})

		return nil
	})
}

func (s *Service) ProcessPartialPayment(ctx context.Context, claimID string, amount int64, paymentRef, notes string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var approvedAmount int64
		err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(approved_amount, amount) FROM claims WHERE id = $1
		`, claimID).Scan(&approvedAmount)
		if err != nil {
			return err
		}

		var totalPaid int64
		err = tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(amount), 0) FROM claim_partial_payments WHERE claim_id = $1
		`, claimID).Scan(&totalPaid)
		if err != nil {
			return err
		}

		remaining := approvedAmount - totalPaid
		if amount > remaining {
			return fmt.Errorf("payment amount (%d) exceeds remaining claim balance (%d)", amount, remaining)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO claim_partial_payments (id, claim_id, amount, payment_ref, paid_at, notes)
			VALUES (uuid_generate_v4(), $1, $2, $3, NOW(), $4)
		`, claimID, amount, paymentRef, notes)
		if err != nil {
			return err
		}

		newTotal := totalPaid + amount
		if newTotal >= approvedAmount {
			_, err = tx.ExecContext(ctx, `
				UPDATE claims SET status = 'paid', paid_at = NOW() WHERE id = $1
			`, claimID)
		}

		return err
	})
}

func (s *Service) MarkAsPaid(ctx context.Context, claimID string, chequeRef string, chequeDate time.Time) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		_, err := tx.ExecContext(ctx, `
			UPDATE claims SET 
				status = 'paid', 
				settlement_date = $1,
				cheque_ref = $2,
				cheque_date = $3,
				paid_at = $4,
				updated_at = $4
			WHERE id = $5 AND status = 'accepted'
		`, now, chequeRef, chequeDate, now, claimID)
		if err != nil {
			return err
		}

		var schemeID, memberID string
		tx.QueryRowContext(ctx, `SELECT scheme_id, member_id FROM claims WHERE id = $1`, claimID).Scan(&schemeID, &memberID)

		_, _ = tx.ExecContext(ctx, `
			UPDATE members SET membership_status = 'deceased' WHERE id = $1
		`, memberID)

		s.logAuditEvent(ctx, tx, schemeID, "claim_paid", claimID, "system",
			map[string]interface{}{
				"action":      "claim_paid",
				"cheque_ref":  chequeRef,
				"cheque_date": chequeDate,
			})

		return nil
	})
}

func (s *Service) GetClaim(ctx context.Context, claimID string) (*Claim, error) {
	var claim Claim
	var rejectionReason, examinerID, examinerNotes, chequeRef sql.NullString
	var settlementDate, chequeDate, reviewedAt, paidAt sql.NullTime
	var approvedAmount sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, scheme_id, member_id, claim_type, claim_form_no, date_of_claim,
		       date_of_leaving, leaving_reason, status, rejection_reason, examiner_id,
		       examiner_notes, settlement_date, cheque_ref, cheque_date, amount,
		       approved_amount, created_by, created_at, updated_at, reviewed_at, paid_at
		FROM claims WHERE id = $1
	`, claimID).Scan(
		&claim.ID, &claim.SchemeID, &claim.MemberID, &claim.ClaimType, &claim.ClaimFormNo,
		&claim.DateOfClaim, &claim.DateOfLeaving, &claim.LeavingReason, &claim.Status,
		&rejectionReason, &examinerID, &examinerNotes, &settlementDate, &chequeRef,
		&chequeDate, &claim.Amount, &approvedAmount, &claim.CreatedBy, &claim.CreatedAt,
		&claim.UpdatedAt, &reviewedAt, &paidAt,
	)
	if err != nil {
		return nil, err
	}

	if rejectionReason.Valid {
		claim.RejectionReason = &rejectionReason.String
	}
	if examinerID.Valid {
		claim.ExaminerID = &examinerID.String
	}
	if examinerNotes.Valid {
		claim.ExaminerNotes = &examinerNotes.String
	}
	if settlementDate.Valid {
		claim.SettlementDate = &settlementDate.Time
	}
	if chequeRef.Valid {
		claim.ChequeRef = &chequeRef.String
	}
	if chequeDate.Valid {
		claim.ChequeDate = &chequeDate.Time
	}
	if approvedAmount.Valid {
		claim.ApprovedAmount = &approvedAmount.Int64
	}
	if reviewedAt.Valid {
		claim.ReviewedAt = &reviewedAt.Time
	}
	if paidAt.Valid {
		claim.PaidAt = &paidAt.Time
	}

	return &claim, nil
}

func (s *Service) ListClaims(ctx context.Context, schemeID string, filters ClaimFilters, limit, offset int) ([]Claim, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM claims WHERE scheme_id = $1`
	args := []interface{}{schemeID}
	argNum := 2

	if filters.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filters.Status)
		argNum++
	}
	if filters.ClaimType != "" {
		countQuery += fmt.Sprintf(" AND claim_type = $%d", argNum)
		args = append(args, filters.ClaimType)
		argNum++
	}
	if !filters.StartDate.IsZero() {
		countQuery += fmt.Sprintf(" AND date_of_claim >= $%d", argNum)
		args = append(args, filters.StartDate)
		argNum++
	}
	if !filters.EndDate.IsZero() {
		countQuery += fmt.Sprintf(" AND date_of_claim <= $%d", argNum)
		args = append(args, filters.EndDate)
		argNum++
	}

	s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := `
		SELECT id, scheme_id, member_id, claim_type, claim_form_no, date_of_claim,
		       date_of_leaving, leaving_reason, status, rejection_reason, amount,
		       approved_amount, created_by, created_at, updated_at
		FROM claims WHERE scheme_id = $1
	`
	args = []interface{}{schemeID}
	argNum = 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filters.Status)
		argNum++
	}
	if filters.ClaimType != "" {
		query += fmt.Sprintf(" AND claim_type = $%d", argNum)
		args = append(args, filters.ClaimType)
		argNum++
	}
	if !filters.StartDate.IsZero() {
		query += fmt.Sprintf(" AND date_of_claim >= $%d", argNum)
		args = append(args, filters.StartDate)
		argNum++
	}
	if !filters.EndDate.IsZero() {
		query += fmt.Sprintf(" AND date_of_claim <= $%d", argNum)
		args = append(args, filters.EndDate)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var claims []Claim
	for rows.Next() {
		var c Claim
		var rejectionReason sql.NullString
		var approvedAmount sql.NullInt64
		if err := rows.Scan(&c.ID, &c.SchemeID, &c.MemberID, &c.ClaimType, &c.ClaimFormNo,
			&c.DateOfClaim, &c.DateOfLeaving, &c.LeavingReason, &c.Status,
			&rejectionReason, &c.Amount, &approvedAmount, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		if rejectionReason.Valid {
			c.RejectionReason = &rejectionReason.String
		}
		if approvedAmount.Valid {
			c.ApprovedAmount = &approvedAmount.Int64
		}
		claims = append(claims, c)
	}

	return claims, total, rows.Err()
}

type ClaimFilters struct {
	Status    ClaimStatus
	ClaimType ExitType
	StartDate time.Time
	EndDate   time.Time
	MemberID  string
}

func (s *Service) UploadClaimDocument(ctx context.Context, claimID, docType, fileName, filePath string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE claim_documents SET 
			file_name = $1, 
			file_path = $2, 
			uploaded = true, 
			uploaded_at = NOW()
		WHERE claim_id = $3 AND doc_type = $4
	`, fileName, filePath, claimID, docType)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document type %s not required for this claim", docType)
	}

	return nil
}

func (s *Service) GetClaimDocuments(ctx context.Context, claimID string) ([]RequiredDocument, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, claim_id, doc_type, file_name, file_path, uploaded, uploaded_at
		FROM claim_documents WHERE claim_id = $1
	`, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []RequiredDocument
	for rows.Next() {
		var d RequiredDocument
		var fileName, filePath sql.NullString
		var uploadedAt sql.NullTime
		if err := rows.Scan(&d.ID, &d.ClaimID, &d.DocType, &fileName, &filePath, &d.Uploaded, &uploadedAt); err != nil {
			continue
		}
		if fileName.Valid {
			d.FileName = fileName.String
		}
		if filePath.Valid {
			d.FilePath = filePath.String
		}
		if uploadedAt.Valid {
			d.UploadedAt = &uploadedAt.Time
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *Service) logAuditEvent(ctx context.Context, tx *sql.Tx, schemeID, entityType, entityID, actorID string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	if tx != nil {
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	} else {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
