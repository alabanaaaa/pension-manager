package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerBeneficiaryDrawdownRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer", "claims_examiner"))

		r.Post("/api/death-benefits/{deathId}/drawdowns", s.handleCreateBeneficiaryDrawdown)
		r.Get("/api/death-benefits/{deathId}/drawdowns", s.handleListBeneficiaryDrawdowns)
		r.Get("/api/drawdowns", s.handleListAllDrawdowns)
		r.Get("/api/drawdowns/{id}", s.handleGetDrawdown)
		r.Put("/api/drawdowns/{id}", s.handleUpdateDrawdown)
		r.Post("/api/drawdowns/{id}/approve", s.handleApproveDrawdown)
		r.Post("/api/drawdowns/{id}/reject", s.handleRejectDrawdown)
		r.Post("/api/drawdowns/{id}/process", s.handleProcessDrawdownPayment)
	})
}

func (s *Server) handleCreateBeneficiaryDrawdown(w http.ResponseWriter, r *http.Request) {
	deathID := chi.URLParam(r, "deathId")
	if deathID == "" {
		respondError(w, http.StatusBadRequest, "death ID is required")
		return
	}

	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		BeneficiaryID string `json:"beneficiary_id"`
		Amount        int64  `json:"amount"`
		DrawdownType  string `json:"drawdown_type"`
		PaymentMethod string `json:"payment_method"`
		BankName      string `json:"bank_name"`
		BankBranch    string `json:"bank_branch"`
		AccountNumber string `json:"account_number"`
		Notes         string `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BeneficiaryID == "" || req.Amount <= 0 || req.DrawdownType == "" {
		respondError(w, http.StatusBadRequest, "beneficiary_id, amount (>0), and drawdown_type are required")
		return
	}

	var beneficiaryBalance int64
	err := s.db.QueryRowContext(r.Context(), `
		SELECT COALESCE(balance, 0) FROM death_beneficiaries WHERE id = $1 AND death_in_service_id = $2
	`, req.BeneficiaryID, deathID).Scan(&beneficiaryBalance)
	if err != nil {
		respondError(w, http.StatusNotFound, "beneficiary not found")
		return
	}

	if req.Amount > beneficiaryBalance {
		respondError(w, http.StatusBadRequest, "drawdown amount exceeds available beneficiary balance")
		return
	}

	var drawdownID string
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO beneficiary_drawdowns (
			id, death_in_service_id, beneficiary_id, scheme_id, amount, drawdown_type,
			payment_method, bank_name, bank_branch, account_number, notes,
			status, requested_by, created_at, updated_at
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending', $11, NOW(), NOW()
		) RETURNING id
	`, deathID, req.BeneficiaryID, schemeID, req.Amount, req.DrawdownType,
		req.PaymentMethod, req.BankName, req.BankBranch, req.AccountNumber,
		req.Notes, userID).Scan(&drawdownID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create drawdown")
		return
	}

	respondCreated(w, map[string]interface{}{
		"id":                drawdownID,
		"status":            "pending",
		"beneficiary_id":    req.BeneficiaryID,
		"amount":            req.Amount,
		"remaining_balance": beneficiaryBalance - req.Amount,
	})
}

func (s *Server) handleListBeneficiaryDrawdowns(w http.ResponseWriter, r *http.Request) {
	deathID := chi.URLParam(r, "deathId")
	if deathID == "" {
		respondError(w, http.StatusBadRequest, "death ID is required")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT bd.id, bd.death_in_service_id, bd.beneficiary_id, bd.scheme_id,
		       bd.amount, bd.drawdown_type, bd.payment_method, bd.bank_name,
		       bd.bank_branch, bd.account_number, bd.notes, bd.status,
		       bd.approved_by, bd.approved_at, bd.processed_by, bd.processed_at,
		       bd.payment_reference, bd.requested_by, bd.created_at, bd.updated_at,
		       db.name as beneficiary_name, db.relationship,
		       m.first_name, m.last_name, m.member_no
		FROM beneficiary_drawdowns bd
		JOIN death_beneficiaries db ON db.id = bd.beneficiary_id
		JOIN death_in_service dis ON dis.id = bd.death_in_service_id
		JOIN members m ON m.id = dis.member_id
		WHERE bd.death_in_service_id = $1
	`
	args := []interface{}{deathID}
	argCount := 1

	if status != "" {
		argCount++
		query += " AND bd.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	argCount++
	query += " ORDER BY bd.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var drawdowns []map[string]interface{}
	for rows.Next() {
		var id, deathID, beneficiaryID, schemeID, drawdownType, paymentMethod string
		var bankName, bankBranch, accountNumber, notes, status string
		var approvedBy, processedBy, requestedBy sql.NullString
		var approvedAt, processedAt sql.NullTime
		var paymentRef sql.NullString
		var amount int64
		var createdAt, updatedAt time.Time
		var beneficiaryName, relationship string
		var firstName, lastName, memberNo string

		if err := rows.Scan(&id, &deathID, &beneficiaryID, &schemeID,
			&amount, &drawdownType, &paymentMethod, &bankName,
			&bankBranch, &accountNumber, &notes, &status,
			&approvedBy, &approvedAt, &processedBy, &processedAt,
			&paymentRef, &requestedBy, &createdAt, &updatedAt,
			&beneficiaryName, &relationship,
			&firstName, &lastName, &memberNo); err != nil {
			continue
		}

		dd := map[string]interface{}{
			"id": id, "death_in_service_id": deathID, "beneficiary_id": beneficiaryID,
			"scheme_id": schemeID, "amount": amount, "drawdown_type": drawdownType,
			"payment_method": paymentMethod, "bank_name": bankName,
			"bank_branch": bankBranch, "account_number": accountNumber,
			"notes": notes, "status": status,
			"requested_by": requestedBy.String, "created_at": createdAt, "updated_at": updatedAt,
			"beneficiary": map[string]string{
				"name": beneficiaryName, "relationship": relationship,
			},
			"deceased_member": map[string]string{
				"first_name": firstName, "last_name": lastName, "member_no": memberNo,
			},
		}
		if approvedBy.Valid {
			dd["approved_by"] = approvedBy.String
		}
		if approvedAt.Valid {
			dd["approved_at"] = approvedAt.Time
		}
		if processedBy.Valid {
			dd["processed_by"] = processedBy.String
		}
		if processedAt.Valid {
			dd["processed_at"] = processedAt.Time
		}
		if paymentRef.Valid {
			dd["payment_reference"] = paymentRef.String
		}
		drawdowns = append(drawdowns, dd)
	}

	if drawdowns == nil {
		drawdowns = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, drawdowns)
}

func (s *Server) handleListAllDrawdowns(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT bd.id, bd.death_in_service_id, bd.beneficiary_id, bd.scheme_id,
		       bd.amount, bd.drawdown_type, bd.payment_method, bd.status,
		       bd.approved_by, bd.approved_at, bd.processed_at,
		       bd.payment_reference, bd.created_at,
		       db.name as beneficiary_name,
		       m.first_name, m.last_name, m.member_no
		FROM beneficiary_drawdowns bd
		JOIN death_beneficiaries db ON db.id = bd.beneficiary_id
		JOIN death_in_service dis ON dis.id = bd.death_in_service_id
		JOIN members m ON m.id = dis.member_id
		WHERE bd.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if status != "" {
		argCount++
		query += " AND bd.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	argCount++
	query += " ORDER BY bd.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var drawdowns []map[string]interface{}
	for rows.Next() {
		var id, deathSvcID, beneficiaryID, schemeID, drawdownType, paymentMethod, status string
		var approvedBy sql.NullString
		var approvedAt, processedAt sql.NullTime
		var paymentRef sql.NullString
		var amount int64
		var createdAt time.Time
		var beneficiaryName string
		var firstName, lastName, memberNo string

		if err := rows.Scan(&id, &deathSvcID, &beneficiaryID, &schemeID,
			&amount, &drawdownType, &paymentMethod, &status,
			&approvedBy, &approvedAt, &processedAt,
			&paymentRef, &createdAt,
			&beneficiaryName,
			&firstName, &lastName, &memberNo); err != nil {
			continue
		}

		dd := map[string]interface{}{
			"id": id, "beneficiary_id": beneficiaryID, "scheme_id": schemeID,
			"amount": amount, "drawdown_type": drawdownType,
			"payment_method": paymentMethod, "status": status,
			"created_at":       createdAt,
			"beneficiary_name": beneficiaryName,
			"member": map[string]string{
				"first_name": firstName, "last_name": lastName, "member_no": memberNo,
			},
		}
		if approvedBy.Valid {
			dd["approved_by"] = approvedBy.String
		}
		if approvedAt.Valid {
			dd["approved_at"] = approvedAt.Time
		}
		if processedAt.Valid {
			dd["processed_at"] = processedAt.Time
		}
		if paymentRef.Valid {
			dd["payment_reference"] = paymentRef.String
		}
		drawdowns = append(drawdowns, dd)
	}

	if drawdowns == nil {
		drawdowns = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, drawdowns)
}

func (s *Server) handleGetDrawdown(w http.ResponseWriter, r *http.Request) {
	drawdownID := chi.URLParam(r, "id")
	if drawdownID == "" {
		respondError(w, http.StatusBadRequest, "drawdown ID is required")
		return
	}

	var dd map[string]interface{}
	err := s.db.QueryRowContext(r.Context(), `
		SELECT bd.*, db.name as beneficiary_name, db.relationship,
		       m.first_name, m.last_name, m.member_no
		FROM beneficiary_drawdowns bd
		JOIN death_beneficiaries db ON db.id = bd.beneficiary_id
		JOIN death_in_service dis ON dis.id = bd.death_in_service_id
		JOIN members m ON m.id = dis.member_id
		WHERE bd.id = $1
	`, drawdownID).Scan(&dd)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "drawdown not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	respondJSON(w, http.StatusOK, dd)
}

func (s *Server) handleUpdateDrawdown(w http.ResponseWriter, r *http.Request) {
	drawdownID := chi.URLParam(r, "id")
	if drawdownID == "" {
		respondError(w, http.StatusBadRequest, "drawdown ID is required")
		return
	}

	var req struct {
		Amount        int64  `json:"amount"`
		PaymentMethod string `json:"payment_method"`
		BankName      string `json:"bank_name"`
		BankBranch    string `json:"bank_branch"`
		AccountNumber string `json:"account_number"`
		Notes         string `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argCount := 0

	if req.Amount > 0 {
		argCount++
		updates = append(updates, "amount = $"+strconv.Itoa(argCount))
		args = append(args, req.Amount)
	}
	if req.PaymentMethod != "" {
		argCount++
		updates = append(updates, "payment_method = $"+strconv.Itoa(argCount))
		args = append(args, req.PaymentMethod)
	}
	if req.BankName != "" {
		argCount++
		updates = append(updates, "bank_name = $"+strconv.Itoa(argCount))
		args = append(args, req.BankName)
	}
	if req.BankBranch != "" {
		argCount++
		updates = append(updates, "bank_branch = $"+strconv.Itoa(argCount))
		args = append(args, req.BankBranch)
	}
	if req.AccountNumber != "" {
		argCount++
		updates = append(updates, "account_number = $"+strconv.Itoa(argCount))
		args = append(args, req.AccountNumber)
	}
	if req.Notes != "" {
		argCount++
		updates = append(updates, "notes = $"+strconv.Itoa(argCount))
		args = append(args, req.Notes)
	}

	argCount++
	args = append(args, drawdownID)
	query := "UPDATE beneficiary_drawdowns SET " + joinStrings(updates, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err := s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleApproveDrawdown(w http.ResponseWriter, r *http.Request) {
	drawdownID := chi.URLParam(r, "id")
	if drawdownID == "" {
		respondError(w, http.StatusBadRequest, "drawdown ID is required")
		return
	}

	userID := GetUserID(r)

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE beneficiary_drawdowns SET status = 'approved', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status = 'pending'
	`, userID, drawdownID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "approval failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (s *Server) handleRejectDrawdown(w http.ResponseWriter, r *http.Request) {
	drawdownID := chi.URLParam(r, "id")
	if drawdownID == "" {
		respondError(w, http.StatusBadRequest, "drawdown ID is required")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r)

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE beneficiary_drawdowns SET status = 'rejected', approved_by = $1, approved_at = NOW(), notes = COALESCE($2, notes), updated_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`, userID, req.Reason, drawdownID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "rejection failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *Server) handleProcessDrawdownPayment(w http.ResponseWriter, r *http.Request) {
	drawdownID := chi.URLParam(r, "id")
	if drawdownID == "" {
		respondError(w, http.StatusBadRequest, "drawdown ID is required")
		return
	}

	userID := GetUserID(r)

	var req struct {
		PaymentReference string `json:"payment_reference"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PaymentReference == "" {
		respondError(w, http.StatusBadRequest, "payment_reference is required")
		return
	}

	var beneficiaryID string
	var amount int64
	err := s.db.QueryRowContext(r.Context(), `
		SELECT beneficiary_id, amount FROM beneficiary_drawdowns
		WHERE id = $1 AND status = 'approved'
	`, drawdownID).Scan(&beneficiaryID, &amount)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusBadRequest, "drawdown not found or not approved")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "transaction failed")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		UPDATE beneficiary_drawdowns SET status = 'paid', processed_by = $1, processed_at = NOW(), payment_reference = $2, updated_at = NOW()
		WHERE id = $3
	`, userID, req.PaymentReference, drawdownID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	_, err = tx.ExecContext(r.Context(), `
		UPDATE death_beneficiaries SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2
	`, amount, beneficiaryID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "balance update failed")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "commit failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":            "paid",
		"payment_reference": req.PaymentReference,
		"amount":            amount,
	})
}
