package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pension-manager/core/domain"

	"github.com/go-chi/chi/v5"
)

// ClaimWithMember extends Claim with member details
type ClaimWithMember struct {
	domain.Claim
	MemberNo    string `json:"member_no"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MemberName  string `json:"member_name"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	BankName    string `json:"bank_name,omitempty"`
	BankBranch  string `json:"bank_branch,omitempty"`
	BankAccount string `json:"bank_account,omitempty"`
}

// registerClaimsRoutes registers claims management routes
func (s *Server) registerClaimsRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		// Claims CRUD
		r.Route("/api/claims", func(r chi.Router) {
			r.Post("/", s.handleCreateClaim)
			r.Get("/", s.handleListClaims)
			r.Get("/{id}", s.handleGetClaim)
			r.Put("/{id}/approve", s.handleApproveClaim)
			r.Put("/{id}/reject", s.handleRejectClaim)
			r.Put("/{id}/pay", s.handlePayClaim)
			r.Put("/{id}/partial-payment", s.handlePartialPayment)
			r.Get("/{id}/documents", s.handleGetClaimDocuments)
		})

		// Death benefits
		r.Route("/api/death-benefits", func(r chi.Router) {
			r.Get("/{claimId}", s.handleGetDeathBenefits)
			r.Put("/{claimId}/distribute", s.handleDistributeDeathBenefits)
		})
	})
}

// handleCreateClaim handles POST /claims
func (s *Server) handleCreateClaim(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		MemberID      string `json:"member_id"`
		ClaimType     string `json:"claim_type"`
		DateOfClaim   string `json:"date_of_claim"`
		DateOfLeaving string `json:"date_of_leaving,omitempty"`
		LeavingReason string `json:"leaving_reason,omitempty"`
		BankName      string `json:"bank_name,omitempty"`
		BankBranch    string `json:"bank_branch,omitempty"`
		BankAccount   string `json:"bank_account,omitempty"`
		Description   string `json:"description,omitempty"`
		Amount        int64  `json:"amount,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MemberID == "" || req.ClaimType == "" || req.DateOfClaim == "" {
		respondError(w, http.StatusBadRequest, "member_id, claim_type, and date_of_claim are required")
		return
	}

	dateOfClaim, err := time.Parse("2006-01-02", req.DateOfClaim)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date_of_claim format (use YYYY-MM-DD)")
		return
	}

	var dateOfLeaving time.Time
	if req.DateOfLeaving != "" {
		dateOfLeaving, err = time.Parse("2006-01-02", req.DateOfLeaving)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid date_of_leaving format (use YYYY-MM-DD)")
			return
		}
	}

	claim := &domain.Claim{
		MemberID:      req.MemberID,
		SchemeID:      schemeID,
		ClaimType:     domain.ClaimType(req.ClaimType),
		DateOfClaim:   dateOfClaim,
		DateOfLeaving: dateOfLeaving,
		LeavingReason: req.LeavingReason,
		Status:        domain.ClaimSubmitted,
		ExaminerID:    userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := claim.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		query := `
			INSERT INTO claims (id, member_id, scheme_id, claim_type, claim_form_no, date_of_claim,
			                    date_of_leaving, leaving_reason, status, examiner_id, bank_name,
			                    bank_branch, bank_account, description, amount, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`
		claimFormNo := fmt.Sprintf("CLM-%s-%d", time.Now().Format("20060102"), time.Now().UnixMilli())
		_, err := tx.ExecContext(r.Context(), query,
			claim.ID, claim.MemberID, claim.SchemeID, claim.ClaimType, claimFormNo,
			claim.DateOfClaim, claim.DateOfLeaving, claim.LeavingReason, claim.Status,
			claim.ExaminerID, req.BankName, req.BankBranch, req.BankAccount,
			req.Description, req.Amount, claim.CreatedAt, claim.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("create claim: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, created_at)
			VALUES (uuid_generate_v4(), $1, 'claim', $2, 'created', $3, NOW())
		`, schemeID, claim.ID, userID)
		return err
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create claim: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, claim)
}

// handleListClaims handles GET /claims
func (s *Server) handleListClaims(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")
	memberID := r.URL.Query().Get("member_id")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	if limit == "" {
		limit = "50"
	}
	if offset == "" {
		offset = "0"
	}

	query := `
		SELECT c.id, c.member_id, c.claim_type, c.claim_form_no, c.date_of_claim,
		       c.date_of_leaving, c.leaving_reason, c.status, c.examiner_id,
		       c.rejection_reason, c.settlement_date, c.cheque_ref, c.cheque_date,
		       c.amount, c.created_at, c.updated_at, c.reviewed_at, c.paid_at,
		       m.member_no, m.first_name, m.last_name
		FROM claims c
		JOIN members m ON m.id = c.member_id
		WHERE c.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND c.status = $%d", argCount)
		args = append(args, status)
	}
	if memberID != "" {
		argCount++
		query += fmt.Sprintf(" AND c.member_id = $%d", argCount)
		args = append(args, memberID)
	}

	query += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list claims: %v", err))
		return
	}
	defer rows.Close()

	type ClaimWithMember struct {
		domain.Claim
		MemberNo   string `json:"member_no"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MemberName string `json:"member_name"`
	}

	var claims []ClaimWithMember
	for rows.Next() {
		var c ClaimWithMember
		var examinerID, rejectionReason, chequeRef sql.NullString
		var dateOfLeaving, settlementDate, chequeDate, reviewedAt, paidAt sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.MemberID, &c.ClaimType, &c.ClaimFormNo, &c.DateOfClaim,
			&dateOfLeaving, &c.LeavingReason, &c.Status, &examinerID,
			&rejectionReason, &settlementDate, &chequeRef, &chequeDate,
			&c.Amount, &c.CreatedAt, &c.UpdatedAt, &reviewedAt, &c.PaidAt,
			&c.MemberNo, &c.FirstName, &c.LastName,
		); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to scan claim: %v", err))
			return
		}
		c.MemberName = c.FirstName + " " + c.LastName
		if dateOfLeaving.Valid {
			c.DateOfLeaving = dateOfLeaving.Time
		}
		if examinerID.Valid {
			c.ExaminerID = examinerID.String
		}
		if rejectionReason.Valid {
			c.RejectionReason = rejectionReason.String
		}
		if settlementDate.Valid {
			c.SettlementDate = settlementDate.Time
		}
		if chequeRef.Valid {
			c.ChequeRef = chequeRef.String
		}
		if chequeDate.Valid {
			c.ChequeDate = chequeDate.Time
		}
		if reviewedAt.Valid {
			c.ReviewedAt = reviewedAt.Time
		}
		if paidAt.Valid {
			c.PaidAt = paidAt.Time
		}
		claims = append(claims, c)
	}

	respondJSON(w, http.StatusOK, claims)
}

// handleGetClaim handles GET /claims/{id}
func (s *Server) handleGetClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	if claimID == "" {
		respondError(w, http.StatusBadRequest, "claim ID is required")
		return
	}

	query := `
		SELECT c.id, c.member_id, c.claim_type, c.claim_form_no, c.date_of_claim,
		       c.date_of_leaving, c.leaving_reason, c.status, c.examiner_id,
		       c.rejection_reason, c.settlement_date, c.cheque_ref, c.cheque_date,
		       c.amount, c.created_at, c.updated_at, c.reviewed_at, c.paid_at,
		       m.member_no, m.first_name, m.last_name, m.phone, m.email,
		       m.bank_name, m.bank_branch, m.bank_account
		FROM claims c
		JOIN members m ON m.id = c.member_id
		WHERE c.id = $1
	`
	var c ClaimWithMember
	var examinerID, rejectionReason, chequeRef sql.NullString
	var dateOfLeaving, settlementDate, chequeDate, reviewedAt, paidAt sql.NullTime
	err := s.db.QueryRowContext(r.Context(), query, claimID).Scan(
		&c.ID, &c.MemberID, &c.ClaimType, &c.ClaimFormNo, &c.DateOfClaim,
		&dateOfLeaving, &c.LeavingReason, &c.Status, &examinerID,
		&rejectionReason, &settlementDate, &chequeRef, &chequeDate,
		&c.Amount, &c.CreatedAt, &c.UpdatedAt, &reviewedAt, &c.PaidAt,
		&c.MemberNo, &c.FirstName, &c.LastName, &c.Phone, &c.Email,
		&c.BankName, &c.BankBranch, &c.BankAccount,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "claim not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get claim: %v", err))
		return
	}

	c.MemberName = c.FirstName + " " + c.LastName
	if dateOfLeaving.Valid {
		c.DateOfLeaving = dateOfLeaving.Time
	}
	if examinerID.Valid {
		c.ExaminerID = examinerID.String
	}
	if rejectionReason.Valid {
		c.RejectionReason = rejectionReason.String
	}
	if settlementDate.Valid {
		c.SettlementDate = settlementDate.Time
	}
	if chequeRef.Valid {
		c.ChequeRef = chequeRef.String
	}
	if chequeDate.Valid {
		c.ChequeDate = chequeDate.Time
	}
	if reviewedAt.Valid {
		c.ReviewedAt = reviewedAt.Time
	}
	if paidAt.Valid {
		c.PaidAt = paidAt.Time
	}

	respondJSON(w, http.StatusOK, c)
}

// handleApproveClaim handles PUT /claims/{id}/approve
func (s *Server) handleApproveClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Notes string `json:"notes,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		var currentStatus string
		err := tx.QueryRowContext(r.Context(), `SELECT status FROM claims WHERE id = $1`, claimID).Scan(&currentStatus)
		if err == sql.ErrNoRows {
			return fmt.Errorf("claim not found")
		}
		if err != nil {
			return fmt.Errorf("get claim: %w", err)
		}
		if currentStatus == "rejected" || currentStatus == "paid" {
			return fmt.Errorf("cannot approve claim with status: %s", currentStatus)
		}

		_, err = tx.ExecContext(r.Context(), `
			UPDATE claims SET status = 'accepted', examiner_id = $1, reviewed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`, userID, claimID)
		if err != nil {
			return fmt.Errorf("approve claim: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM claims WHERE id = $1), 'claim', $1, 'approved', $2, $3, NOW())
		`, claimID, userID, req.Notes)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved", "claim_id": claimID})
}

// handleRejectClaim handles PUT /claims/{id}/reject
func (s *Server) handleRejectClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "rejection reason is required")
		return
	}

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(r.Context(), `
			UPDATE claims SET status = 'rejected', examiner_id = $1, rejection_reason = $2,
			                  reviewed_at = NOW(), updated_at = NOW()
			WHERE id = $3
		`, userID, req.Reason, claimID)
		if err != nil {
			return fmt.Errorf("reject claim: %w", err)
		}

		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM claims WHERE id = $1), 'claim', $1, 'rejected', $2, $3, NOW())
		`, claimID, userID, req.Reason)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected", "claim_id": claimID})
}

// handlePayClaim handles PUT /claims/{id}/pay
func (s *Server) handlePayClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		ChequeRef  string `json:"cheque_ref,omitempty"`
		ChequeDate string `json:"cheque_date,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var chequeDate time.Time
	if req.ChequeDate != "" {
		var err error
		chequeDate, err = time.Parse("2006-01-02", req.ChequeDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid chequeque_date format (use YYYY-MM-DD)")
			return
		}
	}

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		var amount int64
		var memberID string
		err := tx.QueryRowContext(r.Context(), `SELECT amount, member_id FROM claims WHERE id = $1 AND status = 'accepted'`, claimID).Scan(&amount, &memberID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("claim not found or not in accepted status")
		}
		if err != nil {
			return fmt.Errorf("get claim: %w", err)
		}

		_, err = tx.ExecContext(r.Context(), `
			UPDATE claims SET status = 'paid', cheque_ref = $1, cheque_date = $2,
			                  settlement_date = NOW(), paid_at = NOW(), updated_at = NOW()
			WHERE id = $3
		`, req.ChequeRef, chequeDate, claimID)
		if err != nil {
			return fmt.Errorf("pay claim: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM claims WHERE id = $1), 'claim', $1, 'paid', $2, NOW())
		`, claimID, userID)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "paid", "claim_id": claimID})
}

// handlePartialPayment handles PUT /claims/{id}/partial-payment
func (s *Server) handlePartialPayment(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Amount int64  `json:"amount"`
		Ref    string `json:"ref"`
		Date   string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Amount <= 0 || req.Ref == "" || req.Date == "" {
		respondError(w, http.StatusBadRequest, "amount (>0), ref, and date are required")
		return
	}

	paymentDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid date format (use YYYY-MM-DD)")
		return
	}

	err = s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		// Get existing partial payments
		var existingPayments []byte
		err := tx.QueryRowContext(r.Context(), `SELECT partial_payments FROM claims WHERE id = $1`, claimID).Scan(&existingPayments)
		if err != nil {
			return fmt.Errorf("get claim: %w", err)
		}

		// Append new partial payment
		payment := domain.PartialPayment{
			Date:   paymentDate,
			Amount: req.Amount,
			Ref:    req.Ref,
		}

		_, err = tx.ExecContext(r.Context(), `
			UPDATE claims SET partial_payments = COALESCE(partial_payments, '[]'::jsonb) || $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, fmt.Sprintf(`[{"date":"%s","amount":%d,"ref":"%s"}]`, paymentDate.Format("2006-01-02"), payment.Amount, payment.Ref), claimID)
		if err != nil {
			return fmt.Errorf("record partial payment: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM claims WHERE id = $1), 'claim', $1, 'partial_payment', $2, $3, NOW())
		`, claimID, userID, fmt.Sprintf("Partial payment: %d ref: %s", req.Amount, req.Ref))
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "partial_payment_recorded",
		"claim_id": claimID,
		"amount":   req.Amount,
		"ref":      req.Ref,
	})
}

// handleGetClaimDocuments handles GET /claims/{id}/documents
func (s *Server) handleGetClaimDocuments(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "id")
	if claimID == "" {
		respondError(w, http.StatusBadRequest, "claim ID is required")
		return
	}

	query := `
		SELECT id, document_type, file_name, file_size, mime_type, uploaded_by, created_at
		FROM documents WHERE entity_type = 'claim' AND entity_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(r.Context(), query, claimID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get documents: %v", err))
		return
	}
	defer rows.Close()

	type Document struct {
		ID           string    `json:"id"`
		DocumentType string    `json:"document_type"`
		FileName     string    `json:"file_name"`
		FileSize     int64     `json:"file_size"`
		MimeType     string    `json:"mime_type"`
		UploadedBy   string    `json:"uploaded_by"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.DocumentType, &d.FileName, &d.FileSize, &d.MimeType, &d.UploadedBy, &d.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to scan document: %v", err))
			return
		}
		docs = append(docs, d)
	}

	respondJSON(w, http.StatusOK, docs)
}

// handleGetDeathBenefits handles GET /death-benefits/{claimId}
func (s *Server) handleGetDeathBenefits(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claimId")
	if claimID == "" {
		respondError(w, http.StatusBadRequest, "claim ID is required")
	}

	query := `
		SELECT c.id, c.member_id, c.amount, c.status,
		       m.first_name, m.last_name, m.date_of_death
		FROM claims c
		JOIN members m ON m.id = c.member_id
		WHERE c.id = $1 AND c.claim_type = 'death_in_service'
	`
	var deathBenefit struct {
		ClaimID     string    `json:"claim_id"`
		MemberID    string    `json:"member_id"`
		Amount      int64     `json:"amount"`
		Status      string    `json:"status"`
		FirstName   string    `json:"first_name"`
		LastName    string    `json:"last_name"`
		DateOfDeath time.Time `json:"date_of_death"`
	}
	err := s.db.QueryRowContext(r.Context(), query, claimID).Scan(
		&deathBenefit.ClaimID, &deathBenefit.MemberID, &deathBenefit.Amount,
		&deathBenefit.Status, &deathBenefit.FirstName, &deathBenefit.LastName,
		&deathBenefit.DateOfDeath,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "death benefit claim not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get death benefit: %v", err))
		return
	}

	// Get beneficiaries
	benefQuery := `
		SELECT b.id, b.name, b.relationship, b.date_of_birth, b.id_number, b.phone, b.physical_address, b.allocation_pct
		FROM beneficiaries b WHERE b.member_id = $1 ORDER BY b.allocation_pct DESC
	`
	bRows, err := s.db.QueryContext(r.Context(), benefQuery, deathBenefit.MemberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get beneficiaries: %v", err))
		return
	}
	defer bRows.Close()

	type Beneficiary struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		Relationship    string    `json:"relationship"`
		DateOfBirth     time.Time `json:"date_of_birth"`
		IDNumber        string    `json:"id_number"`
		Phone           string    `json:"phone"`
		PhysicalAddress string    `json:"physical_address"`
		AllocationPct   float64   `json:"allocation_pct"`
	}

	var beneficiaries []Beneficiary
	for bRows.Next() {
		var b Beneficiary
		if err := bRows.Scan(&b.ID, &b.Name, &b.Relationship, &b.DateOfBirth, &b.IDNumber, &b.Phone, &b.PhysicalAddress, &b.AllocationPct); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to scan beneficiary: %v", err))
			return
		}
		beneficiaries = append(beneficiaries, b)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"death_benefit": deathBenefit,
		"beneficiaries": beneficiaries,
	})
}

// handleDistributeDeathBenefits handles PUT /death-benefits/{claimId}/distribute
func (s *Server) handleDistributeDeathBenefits(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claimId")
	userID := GetUserID(r)

	var req struct {
		Distributions []struct {
			BeneficiaryID string `json:"beneficiary_id"`
			Amount        int64  `json:"amount"`
			EndUser       string `json:"end_user"` // e.g., school, guardian
		} `json:"distributions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		for _, dist := range req.Distributions {
			_, err := tx.ExecContext(r.Context(), `
				INSERT INTO beneficiary_distributions (id, claim_id, beneficiary_id, amount, end_user, distributed_by, created_at)
				VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())
			`, claimID, dist.BeneficiaryID, dist.Amount, dist.EndUser, userID)
			if err != nil {
				return fmt.Errorf("record distribution: %w", err)
			}
		}

		// Record audit event
		_, err := tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM claims WHERE id = $1), 'claim', $1, 'death_benefit_distributed', $2, NOW())
		`, claimID, userID)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "distributed", "claim_id": claimID})
}
