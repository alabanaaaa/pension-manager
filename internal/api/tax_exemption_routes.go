package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerTaxExemptionRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Post("/api/tax-exemptions", s.handleCreateTaxExemption)
		r.Get("/api/tax-exemptions", s.handleListTaxExemptions)
		r.Get("/api/tax-exemptions/{id}", s.handleGetTaxExemption)
		r.Put("/api/tax-exemptions/{id}", s.handleUpdateTaxExemption)
		r.Delete("/api/tax-exemptions/{id}", s.handleDeleteTaxExemption)
		r.Post("/api/tax-exemptions/{id}/approve", s.handleApproveTaxExemption)
		r.Post("/api/tax-exemptions/{id}/reject", s.handleRejectTaxExemption)
		r.Get("/api/members/{id}/tax-exemptions", s.handleGetMemberTaxExemptions)
	})
}

func (s *Server) handleCreateTaxExemption(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		MemberID      string `json:"member_id"`
		ExemptionType string `json:"exemption_type"`
		Reason        string `json:"reason"`
		CertificateNo string `json:"certificate_no"`
		ExpiryDate    string `json:"expiry_date"`
		MonthlyLimit  int64  `json:"monthly_limit"`
		ReliefAmount  int64  `json:"relief_amount"`
		KraReference  string `json:"kra_reference"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MemberID == "" || req.ExemptionType == "" {
		respondError(w, http.StatusBadRequest, "member_id and exemption_type are required")
		return
	}

	expiryDate, _ := time.Parse("2006-01-02", req.ExpiryDate)

	var exemptionID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO tax_exemptions (
			id, member_id, scheme_id, exemption_type, reason, certificate_no,
			expiry_date, monthly_limit, relief_amount, kra_reference, status,
			approved_by, approved_at, created_by, created_at, updated_at
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending', NULL, NULL, $10, NOW(), NOW()
		) RETURNING id
	`, req.MemberID, schemeID, req.ExemptionType, req.Reason, req.CertificateNo,
		expiryDate, req.MonthlyLimit, req.ReliefAmount, req.KraReference, userID).Scan(&exemptionID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create tax exemption")
		return
	}

	respondCreated(w, map[string]interface{}{"id": exemptionID, "status": "pending"})
}

func (s *Server) handleListTaxExemptions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	memberID := r.URL.Query().Get("member_id")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT te.id, te.member_id, te.scheme_id, te.exemption_type, te.reason,
		       te.certificate_no, te.expiry_date, te.monthly_limit, te.relief_amount,
		       te.kra_reference, te.status, te.approved_by, te.approved_at,
		       te.created_by, te.created_at, te.updated_at,
		       m.first_name, m.last_name, m.member_no
		FROM tax_exemptions te
		JOIN members m ON m.id = te.member_id
		WHERE te.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if memberID != "" {
		argCount++
		query += " AND te.member_id = $" + strconv.Itoa(argCount)
		args = append(args, memberID)
	}
	if status != "" {
		argCount++
		query += " AND te.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	argCount++
	query += " ORDER BY te.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var exemptions []map[string]interface{}
	for rows.Next() {
		var id, memberID, schemeID, exemptionType, reason, certNo, kraRef, status, createdBy string
		var expiryDate sql.NullTime
		var monthlyLimit, reliefAmount sql.NullInt64
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		var createdAt, updatedAt time.Time
		var firstName, lastName, memberNo string

		if err := rows.Scan(&id, &memberID, &schemeID, &exemptionType, &reason, &certNo,
			&expiryDate, &monthlyLimit, &reliefAmount, &kraRef, &status, &approvedBy, &approvedAt,
			&createdBy, &createdAt, &updatedAt, &firstName, &lastName, &memberNo); err != nil {
			continue
		}

		ex := map[string]interface{}{
			"id": id, "member_id": memberID, "scheme_id": schemeID,
			"exemption_type": exemptionType, "reason": reason,
			"certificate_no": certNo, "kra_reference": kraRef,
			"status": status, "created_by": createdBy,
			"created_at": createdAt, "updated_at": updatedAt,
			"member": map[string]string{
				"first_name": firstName, "last_name": lastName, "member_no": memberNo,
			},
		}
		if expiryDate.Valid {
			ex["expiry_date"] = expiryDate.Time
		}
		if monthlyLimit.Valid {
			ex["monthly_limit"] = monthlyLimit.Int64
		}
		if reliefAmount.Valid {
			ex["relief_amount"] = reliefAmount.Int64
		}
		if approvedBy.Valid {
			ex["approved_by"] = approvedBy.String
		}
		if approvedAt.Valid {
			ex["approved_at"] = approvedAt.Time
		}
		exemptions = append(exemptions, ex)
	}

	if exemptions == nil {
		exemptions = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, exemptions)
}

func (s *Server) handleGetTaxExemption(w http.ResponseWriter, r *http.Request) {
	exemptionID := chi.URLParam(r, "id")
	if exemptionID == "" {
		respondError(w, http.StatusBadRequest, "exemption ID is required")
		return
	}

	var exemption map[string]interface{}
	err := s.db.QueryRowContext(r.Context(), `
		SELECT te.*, m.first_name, m.last_name, m.member_no
		FROM tax_exemptions te
		JOIN members m ON m.id = te.member_id
		WHERE te.id = $1
	`, exemptionID).Scan(&exemption)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "tax exemption not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	respondJSON(w, http.StatusOK, exemption)
}

func (s *Server) handleUpdateTaxExemption(w http.ResponseWriter, r *http.Request) {
	exemptionID := chi.URLParam(r, "id")
	if exemptionID == "" {
		respondError(w, http.StatusBadRequest, "exemption ID is required")
		return
	}

	var req struct {
		Reason        string `json:"reason"`
		CertificateNo string `json:"certificate_no"`
		ExpiryDate    string `json:"expiry_date"`
		MonthlyLimit  int64  `json:"monthly_limit"`
		ReliefAmount  int64  `json:"relief_amount"`
		KraReference  string `json:"kra_reference"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argCount := 0

	if req.Reason != "" {
		argCount++
		updates = append(updates, "reason = $"+strconv.Itoa(argCount))
		args = append(args, req.Reason)
	}
	if req.CertificateNo != "" {
		argCount++
		updates = append(updates, "certificate_no = $"+strconv.Itoa(argCount))
		args = append(args, req.CertificateNo)
	}
	if req.ExpiryDate != "" {
		argCount++
		updates = append(updates, "expiry_date = $"+strconv.Itoa(argCount))
		args = append(args, req.ExpiryDate)
	}
	if req.MonthlyLimit > 0 {
		argCount++
		updates = append(updates, "monthly_limit = $"+strconv.Itoa(argCount))
		args = append(args, req.MonthlyLimit)
	}
	if req.ReliefAmount > 0 {
		argCount++
		updates = append(updates, "relief_amount = $"+strconv.Itoa(argCount))
		args = append(args, req.ReliefAmount)
	}
	if req.KraReference != "" {
		argCount++
		updates = append(updates, "kra_reference = $"+strconv.Itoa(argCount))
		args = append(args, req.KraReference)
	}

	argCount++
	args = append(args, exemptionID)
	query := "UPDATE tax_exemptions SET " + joinStrings(updates, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err := s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleDeleteTaxExemption(w http.ResponseWriter, r *http.Request) {
	exemptionID := chi.URLParam(r, "id")
	if exemptionID == "" {
		respondError(w, http.StatusBadRequest, "exemption ID is required")
		return
	}

	result, err := s.db.ExecContext(r.Context(), `DELETE FROM tax_exemptions WHERE id = $1`, exemptionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "tax exemption not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleApproveTaxExemption(w http.ResponseWriter, r *http.Request) {
	exemptionID := chi.URLParam(r, "id")
	if exemptionID == "" {
		respondError(w, http.StatusBadRequest, "exemption ID is required")
		return
	}

	userID := GetUserID(r)

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE tax_exemptions SET status = 'approved', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status = 'pending'
	`, userID, exemptionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "approval failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (s *Server) handleRejectTaxExemption(w http.ResponseWriter, r *http.Request) {
	exemptionID := chi.URLParam(r, "id")
	if exemptionID == "" {
		respondError(w, http.StatusBadRequest, "exemption ID is required")
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
		UPDATE tax_exemptions SET status = 'rejected', approved_by = $1, approved_at = NOW(), reason = COALESCE($2, reason), updated_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`, userID, req.Reason, exemptionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "rejection failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *Server) handleGetMemberTaxExemptions(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, exemption_type, reason, certificate_no, expiry_date, monthly_limit,
		       relief_amount, kra_reference, status, approved_by, approved_at, created_at, updated_at
		FROM tax_exemptions WHERE member_id = $1 ORDER BY created_at DESC
	`, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var exemptions []map[string]interface{}
	for rows.Next() {
		var id, exemptionType, reason, certNo, kraRef, status, approvedBy string
		var expiryDate sql.NullTime
		var monthlyLimit, reliefAmount sql.NullInt64
		var approvedAt sql.NullTime
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &exemptionType, &reason, &certNo, &expiryDate,
			&monthlyLimit, &reliefAmount, &kraRef, &status, &approvedBy, &approvedAt,
			&createdAt, &updatedAt); err != nil {
			continue
		}

		ex := map[string]interface{}{
			"id": id, "exemption_type": exemptionType, "reason": reason,
			"certificate_no": certNo, "kra_reference": kraRef,
			"status": status, "approved_by": approvedBy,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if expiryDate.Valid {
			ex["expiry_date"] = expiryDate.Time
		}
		if monthlyLimit.Valid {
			ex["monthly_limit"] = monthlyLimit.Int64
		}
		if reliefAmount.Valid {
			ex["relief_amount"] = reliefAmount.Int64
		}
		if approvedAt.Valid {
			ex["approved_at"] = approvedAt.Time
		}
		exemptions = append(exemptions, ex)
	}

	if exemptions == nil {
		exemptions = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, exemptions)
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
