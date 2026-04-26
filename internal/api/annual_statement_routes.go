package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerAnnualStatementRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Post("/api/annual-statements/generate", s.handleGenerateAnnualStatements)
		r.Get("/api/annual-statements", s.handleListAnnualStatements)
		r.Get("/api/annual-statements/{id}", s.handleGetAnnualStatement)
		r.Get("/api/annual-statements/{id}/pdf", s.handleDownloadAnnualStatementPDF)
		r.Post("/api/annual-statements/{id}/email", s.handleEmailAnnualStatement)
		r.Post("/api/annual-statements/bulk-email", s.handleBulkEmailStatements)
		r.Get("/api/members/{id}/statements", s.handleGetMemberStatements)
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Put("/api/annual-statements/{id}/hold", s.handleHoldAnnualStatement)
		r.Put("/api/annual-statements/{id}/release", s.handleReleaseAnnualStatement)
		r.Delete("/api/annual-statements/{id}", s.handleDeleteAnnualStatement)
	})
}

func (s *Server) handleGenerateAnnualStatements(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		Year       int    `json:"year"`
		SchemeID   string `json:"scheme_id"`
		Department string `json:"department"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Year == 0 {
		req.Year = time.Now().Year() - 1
	}
	if req.SchemeID != "" {
		schemeID = req.SchemeID
	}

	query := `
		SELECT m.id, m.member_no, m.first_name, m.last_name, m.email, m.phone,
		       m.account_balance, m.basic_salary, m.date_joined_scheme,
		       m.member_contribution_rate, m.sponsor_contribution_rate
		FROM members m
		WHERE m.scheme_id = $1 AND m.membership_status = 'active'
	`
	args := []interface{}{schemeID}

	if req.Department != "" {
		query += " AND m.department = $2"
		args = append(args, req.Department)
	}

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var generated int
	var failed int
	for rows.Next() {
		var memberID, memberNo, firstName, lastName, email, phone string
		var balance int64
		var basicSalary int
		var dateJoined sql.NullTime
		var memberRate, sponsorRate sql.NullFloat64

		if err := rows.Scan(&memberID, &memberNo, &firstName, &lastName, &email, &phone,
			&balance, &basicSalary, &dateJoined, &memberRate, &sponsorRate); err != nil {
			failed++
			continue
		}

		totalContrib, memberContrib, sponsorContrib := s.calculateAnnualContributions(memberID, req.Year)
		openingBalance := balance - totalContrib

		_, err := s.db.ExecContext(r.Context(), `
			INSERT INTO annual_statements (
				id, member_id, scheme_id, year, opening_balance, total_contributions,
				member_contributions, sponsor_contributions, closing_balance,
				basic_salary, member_rate, sponsor_rate, status,
				generated_by, generated_at, created_at, updated_at
			) VALUES (
				uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'generated', $12, NOW(), NOW(), NOW()
			) ON CONFLICT (member_id, year) DO UPDATE SET
				opening_balance = $4, total_contributions = $5,
				member_contributions = $6, sponsor_contributions = $7,
				closing_balance = $8, status = 'generated',
				generated_by = $12, generated_at = NOW(), updated_at = NOW()
		`, memberID, schemeID, req.Year, openingBalance, totalContrib,
			memberContrib, sponsorContrib, balance, basicSalary,
			memberRate.Float64, sponsorRate.Float64, userID)

		if err != nil {
			failed++
			continue
		}
		generated++
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":      req.Year,
		"generated": generated,
		"failed":    failed,
		"total":     generated + failed,
	})
}

func (s *Server) calculateAnnualContributions(memberID string, year int) (total, member, sponsor int64) {
	s.db.QueryRowContext(nil, `
		SELECT COALESCE(SUM(total_amount), 0),
		       COALESCE(SUM(employee_amount), 0),
		       COALESCE(SUM(employer_amount), 0)
		FROM contributions
		WHERE member_id = $1 AND EXTRACT(YEAR FROM period) = $2 AND status = 'confirmed'
	`, memberID, year).Scan(&total, &member, &sponsor)
	return
}

func (s *Server) handleListAnnualStatements(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	memberID := r.URL.Query().Get("member_id")
	year := r.URL.Query().Get("year")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT as2.id, as2.member_id, as2.scheme_id, as2.year, as2.opening_balance,
		       as2.total_contributions, as2.member_contributions, as2.sponsor_contributions,
		       as2.closing_balance, as2.basic_salary, as2.member_rate, as2.sponsor_rate,
		       as2.status, as2.hold_reason, as2.generated_by, as2.generated_at,
		       as2.created_at, as2.updated_at,
		       m.first_name, m.last_name, m.member_no, m.email
		FROM annual_statements as2
		JOIN members m ON m.id = as2.member_id
		WHERE as2.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if memberID != "" {
		argCount++
		query += " AND as2.member_id = $" + strconv.Itoa(argCount)
		args = append(args, memberID)
	}
	if year != "" {
		argCount++
		query += " AND as2.year = $" + strconv.Itoa(argCount)
		args = append(args, year)
	}
	if status != "" {
		argCount++
		query += " AND as2.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	argCount++
	query += " ORDER BY as2.year DESC, m.last_name LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var statements []map[string]interface{}
	for rows.Next() {
		var id, memberID, schemeID, year, status, generatedBy string
		var openingBal, totalContrib, memberContrib, sponsorContrib, closingBal, basicSalary int64
		var memberRate, sponsorRate sql.NullFloat64
		var holdReason sql.NullString
		var generatedAt, createdAt, updatedAt time.Time
		var firstName, lastName, memberNo, email string

		if err := rows.Scan(&id, &memberID, &schemeID, &year, &openingBal,
			&totalContrib, &memberContrib, &sponsorContrib, &closingBal, &basicSalary,
			&memberRate, &sponsorRate, &status, &holdReason, &generatedBy,
			&generatedAt, &createdAt, &updatedAt, &firstName, &lastName, &memberNo, &email); err != nil {
			continue
		}

		stmt := map[string]interface{}{
			"id": id, "member_id": memberID, "scheme_id": schemeID, "year": year,
			"opening_balance": openingBal, "total_contributions": totalContrib,
			"member_contributions": memberContrib, "sponsor_contributions": sponsorContrib,
			"closing_balance": closingBal, "basic_salary": basicSalary,
			"status": status, "generated_by": generatedBy,
			"generated_at": generatedAt, "created_at": createdAt, "updated_at": updatedAt,
			"member": map[string]string{
				"first_name": firstName, "last_name": lastName,
				"member_no": memberNo, "email": email,
			},
		}
		if memberRate.Valid {
			stmt["member_rate"] = memberRate.Float64
		}
		if sponsorRate.Valid {
			stmt["sponsor_rate"] = sponsorRate.Float64
		}
		if holdReason.Valid {
			stmt["hold_reason"] = holdReason.String
		}
		statements = append(statements, stmt)
	}

	if statements == nil {
		statements = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, statements)
}

func (s *Server) handleGetAnnualStatement(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	var stmt map[string]interface{}
	err := s.db.QueryRowContext(r.Context(), `
		SELECT as2.*, m.first_name, m.last_name, m.member_no, m.email, m.phone,
		       m.date_joined_scheme, m.basic_salary
		FROM annual_statements as2
		JOIN members m ON m.id = as2.member_id
		WHERE as2.id = $1
	`, statementID).Scan(&stmt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "annual statement not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	respondJSON(w, http.StatusOK, stmt)
}

func (s *Server) handleDownloadAnnualStatementPDF(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	var memberID, year string
	var closingBalance int64
	var firstName, lastName, memberNo, email string

	err := s.db.QueryRowContext(r.Context(), `
		SELECT as2.member_id, as2.year, as2.closing_balance,
		       m.first_name, m.last_name, m.member_no, m.email
		FROM annual_statements as2
		JOIN members m ON m.id = as2.member_id
		WHERE as2.id = $1
	`, statementID).Scan(&memberID, &year, &closingBalance, &firstName, &lastName, &memberNo, &email)

	if err != nil {
		respondError(w, http.StatusNotFound, "annual statement not found")
		return
	}

	pdf := generateAnnualStatementPDF(memberNo, firstName, lastName, email, year, closingBalance)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=annual_statement_%s_%s.pdf", memberNo, year))
	w.Write(pdf)
}

func (s *Server) handleEmailAnnualStatement(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	var memberID, year, email, firstName, lastName string
	err := s.db.QueryRowContext(r.Context(), `
		SELECT as2.member_id, as2.year, m.email, m.first_name, m.last_name
		FROM annual_statements as2
		JOIN members m ON m.id = as2.member_id
		WHERE as2.id = $1
	`, statementID).Scan(&memberID, &year, &email, &firstName, &lastName)

	if err != nil {
		respondError(w, http.StatusNotFound, "annual statement not found")
		return
	}

	if email == "" {
		respondError(w, http.StatusBadRequest, "member has no email address")
		return
	}

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE annual_statements SET email_sent = true, email_sent_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, statementID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update statement")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "email_queued",
		"email":   email,
		"message": fmt.Sprintf("Annual statement for %d will be sent to %s", year, email),
	})
}

func (s *Server) handleBulkEmailStatements(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Year       int      `json:"year"`
		SchemeID   string   `json:"scheme_id"`
		Department string   `json:"department"`
		MemberIDs  []string `json:"member_ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Year == 0 {
		req.Year = time.Now().Year() - 1
	}

	schemeID := GetSchemeID(r)
	if req.SchemeID != "" {
		schemeID = req.SchemeID
	}

	query := `
		SELECT as2.id, m.email, m.first_name, m.last_name, m.member_no
		FROM annual_statements as2
		JOIN members m ON m.id = as2.member_id
		WHERE as2.scheme_id = $1 AND as2.year = $2 AND as2.email_sent = false
	`
	args := []interface{}{schemeID, req.Year}
	argCount := 2

	if req.Department != "" {
		argCount++
		query += " AND m.department = $" + strconv.Itoa(argCount)
		args = append(args, req.Department)
	}
	if len(req.MemberIDs) > 0 {
		argCount++
		query += " AND as2.member_id = ANY($" + strconv.Itoa(argCount) + ")"
		args = append(args, req.MemberIDs)
	}

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var queued int
	var failedEmails []string
	for rows.Next() {
		var id, email, firstName, lastName, memberNo string
		if err := rows.Scan(&id, &email, &firstName, &lastName, &memberNo); err != nil {
			continue
		}

		if email == "" {
			failedEmails = append(failedEmails, memberNo+": no email")
			continue
		}

		_, err := s.db.ExecContext(r.Context(), `
			UPDATE annual_statements SET email_sent = true, email_sent_at = NOW(), updated_at = NOW()
			WHERE id = $1
		`, id)
		if err != nil {
			failedEmails = append(failedEmails, memberNo+": update failed")
			continue
		}
		queued++
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":          req.Year,
		"queued":        queued,
		"failed_count":  len(failedEmails),
		"failed_emails": failedEmails,
	})
}

func (s *Server) handleGetMemberStatements(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, year, opening_balance, total_contributions, member_contributions,
		       sponsor_contributions, closing_balance, basic_salary, member_rate, sponsor_rate,
		       status, hold_reason, generated_at, email_sent, email_sent_at
		FROM annual_statements
		WHERE member_id = $1
		ORDER BY year DESC
	`, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var statements []map[string]interface{}
	for rows.Next() {
		var id string
		var year int
		var openingBal, totalContrib, memberContrib, sponsorContrib, closingBal, basicSalary int64
		var memberRate, sponsorRate sql.NullFloat64
		var status string
		var holdReason sql.NullString
		var generatedAt sql.NullTime
		var emailSent bool
		var emailSentAt sql.NullTime

		if err := rows.Scan(&id, &year, &openingBal, &totalContrib, &memberContrib,
			&sponsorContrib, &closingBal, &basicSalary, &memberRate, &sponsorRate,
			&status, &holdReason, &generatedAt, &emailSent, &emailSentAt); err != nil {
			continue
		}

		stmt := map[string]interface{}{
			"id": id, "year": year, "opening_balance": openingBal,
			"total_contributions": totalContrib, "member_contributions": memberContrib,
			"sponsor_contributions": sponsorContrib, "closing_balance": closingBal,
			"basic_salary": basicSalary, "status": status,
			"email_sent": emailSent,
		}
		if memberRate.Valid {
			stmt["member_rate"] = memberRate.Float64
		}
		if sponsorRate.Valid {
			stmt["sponsor_rate"] = sponsorRate.Float64
		}
		if holdReason.Valid {
			stmt["hold_reason"] = holdReason.String
		}
		if generatedAt.Valid {
			stmt["generated_at"] = generatedAt.Time
		}
		if emailSentAt.Valid {
			stmt["email_sent_at"] = emailSentAt.Time
		}
		statements = append(statements, stmt)
	}

	if statements == nil {
		statements = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, statements)
}

func (s *Server) handleHoldAnnualStatement(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "reason is required")
		return
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE annual_statements SET status = 'on_hold', hold_reason = $1, updated_at = NOW()
		WHERE id = $2
	`, req.Reason, statementID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "on_hold"})
}

func (s *Server) handleReleaseAnnualStatement(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE annual_statements SET status = 'generated', hold_reason = NULL, updated_at = NOW()
		WHERE id = $1
	`, statementID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

func (s *Server) handleDeleteAnnualStatement(w http.ResponseWriter, r *http.Request) {
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		respondError(w, http.StatusBadRequest, "statement ID is required")
		return
	}

	result, err := s.db.ExecContext(r.Context(), `DELETE FROM annual_statements WHERE id = $1`, statementID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "annual statement not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func generateAnnualStatementPDF(memberNo, firstName, lastName, email, year string, closingBalance int64) []byte {
	content := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 400 >>
stream
BT
/F1 14 Tf
72 720 Td
(ANNUAL PENSION STATEMENT) Tj
0 -30 Td
/F1 12 Tf
(Member: %s %s) Tj
0 -20 Td
(Member No: %s) Tj
0 -20 Td
(Year: %s) Tj
0 -40 Td
(Closing Balance: %d KES) Tj
0 -20 Td
(Email: %s) Tj
0 -40 Td
(This is a computer-generated statement.) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000266 00000 n
0000000517 00000 n
trailer
<< /Size 6 /Root 1 0 R >>
startxref
589
%%%%EOF`, firstName, lastName, memberNo, year, closingBalance, email)
	return []byte(content)
}
