package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"pension-manager/internal/db"
	"pension-manager/internal/member"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerMemberRoutes(r chi.Router) {
	r.Get("/api/members/pending", s.listPendingMembers)
	r.Post("/api/members/pending", s.submitMember)
	r.Get("/api/members/pending/{id}", s.getPendingMember)
	r.Post("/api/members/pending/{id}/approve", s.approvePendingMember)
	r.Post("/api/members/pending/{id}/reject", s.rejectPendingMember)

	r.Get("/api/members/changes/pending", s.listPendingChanges)
	r.Post("/api/members/changes/pending/{id}/approve", s.approveMemberChange)
	r.Post("/api/members/changes/pending/{id}/reject", s.rejectMemberChange)
}

type memberService struct {
	db *db.DB
}

func (s *Server) listPendingMembers(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, member_no, first_name, last_name, id_number, sponsor_id, date_joined_scheme,
		       status, submitted_by, created_at, reviewed_at
		FROM pending_member_registrations
		WHERE scheme_id = $1 AND status = $2
		ORDER BY created_at DESC
	`, schemeID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, memberNo, firstName, lastName, idNo, sponsorID, submittedBy string
		var dateJoined time.Time
		var status string
		var createdAt, reviewedAt *time.Time
		if err := rows.Scan(&id, &memberNo, &firstName, &lastName, &idNo, &sponsorID, &dateJoined, &status, &submittedBy, &createdAt, &reviewedAt); err != nil {
			continue
		}
		m := map[string]interface{}{
			"id": id, "member_no": memberNo, "first_name": firstName, "last_name": lastName,
			"id_number": idNo, "sponsor_id": sponsorID, "date_joined_scheme": dateJoined,
			"status": status, "submitted_by": submittedBy, "created_at": createdAt,
		}
		if reviewedAt != nil {
			m["reviewed_at"] = *reviewedAt
		}
		results = append(results, m)
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, results)
}

func (s *Server) submitMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req member.PendingMemberInput
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		respondError(w, http.StatusBadRequest, "first_name, last_name, and date_of_birth are required")
		return
	}

	var regID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO pending_member_registrations (
			scheme_id, member_no, first_name, last_name, other_names, gender, date_of_birth,
			nationality, id_number, kra_pin, email, phone, postal_address, postal_code, town,
			marital_status, spouse_name, next_of_kin, next_of_kin_phone, bank_name, bank_branch,
			bank_account, payroll_no, designation, department, sponsor_id, date_first_appt,
			date_joined_scheme, expected_retirement, basic_salary, status, submitted_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,'pending',$31)
		RETURNING id
	`, schemeID, req.MemberNo, req.FirstName, req.LastName, req.OtherNames, req.Gender,
		req.DateOfBirth, req.Nationality, req.IDNumber, req.KRAPin, req.Email, req.Phone,
		req.PostalAddress, req.PostalCode, req.Town, req.MaritalStatus, req.SpouseName,
		req.NextOfKin, req.NextOfKinPhone, req.BankName, req.BankBranch, req.BankAccount,
		req.PayrollNo, req.Designation, req.Department, req.SponsorID, req.DateFirstAppt,
		req.DateJoinedScheme, req.ExpectedRetirement, req.BasicSalary, userID).Scan(&regID)

	if err != nil {
		slog.Error("submit member failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to submit member for approval")
		return
	}

	respondCreated(w, map[string]interface{}{"id": regID, "status": "pending"})
}

func (s *Server) getPendingMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	id := chi.URLParam(r, "id")

	var m member.PendingMemberInput
	var createdAt, reviewedAt *time.Time
	var status, submittedBy, reviewedBy, rejectionReason string

	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, member_no, first_name, last_name, other_names, gender, date_of_birth,
		       nationality, id_number, kra_pin, email, phone, postal_address, postal_code,
		       town, marital_status, spouse_name, next_of_kin, next_of_kin_phone, bank_name,
		       bank_branch, bank_account, payroll_no, designation, department, sponsor_id,
		       date_first_appt, date_joined_scheme, expected_retirement, basic_salary,
		       status, submitted_by, reviewed_by, rejection_reason, created_at, reviewed_at
		FROM pending_member_registrations
		WHERE id = $1 AND scheme_id = $2
	`, id, schemeID).Scan(&m.ID, &m.MemberNo, &m.FirstName, &m.LastName, &m.OtherNames,
		&m.Gender, &m.DateOfBirth, &m.Nationality, &m.IDNumber, &m.KRAPin, &m.Email,
		&m.Phone, &m.PostalAddress, &m.PostalCode, &m.Town, &m.MaritalStatus, &m.SpouseName,
		&m.NextOfKin, &m.NextOfKinPhone, &m.BankName, &m.BankBranch, &m.BankAccount,
		&m.PayrollNo, &m.Designation, &m.Department, &m.SponsorID, &m.DateFirstAppt,
		&m.DateJoinedScheme, &m.ExpectedRetirement, &m.BasicSalary, &status, &submittedBy,
		&reviewedBy, &rejectionReason, &createdAt, &reviewedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "pending member not found")
		return
	}

	result := map[string]interface{}{
		"id": m.ID, "member_no": m.MemberNo, "first_name": m.FirstName, "last_name": m.LastName,
		"other_names": m.OtherNames, "gender": m.Gender, "date_of_birth": m.DateOfBirth,
		"nationality": m.Nationality, "id_number": m.IDNumber, "kra_pin": m.KRAPin,
		"email": m.Email, "phone": m.Phone, "postal_address": m.PostalAddress,
		"postal_code": m.PostalCode, "town": m.Town, "marital_status": m.MaritalStatus,
		"spouse_name": m.SpouseName, "next_of_kin": m.NextOfKin, "next_of_kin_phone": m.NextOfKinPhone,
		"bank_name": m.BankName, "bank_branch": m.BankBranch, "bank_account": m.BankAccount,
		"payroll_no": m.PayrollNo, "designation": m.Designation, "department": m.Department,
		"sponsor_id": m.SponsorID, "date_first_appt": m.DateFirstAppt,
		"date_joined_scheme": m.DateJoinedScheme, "expected_retirement": m.ExpectedRetirement,
		"basic_salary": m.BasicSalary, "status": status, "submitted_by": submittedBy,
		"created_at": createdAt,
	}
	if rejectionReason != "" {
		result["rejection_reason"] = rejectionReason
	}
	if reviewedBy != "" {
		result["reviewed_by"] = reviewedBy
	}
	if reviewedAt != nil {
		result["reviewed_at"] = *reviewedAt
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) approvePendingMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	var pending member.PendingMemberInput
	err := s.db.QueryRowContext(r.Context(), `
		SELECT member_no, first_name, last_name, other_names, gender, date_of_birth,
		       nationality, id_number, kra_pin, email, phone, postal_address, postal_code,
		       town, marital_status, spouse_name, next_of_kin, next_of_kin_phone, bank_name,
		       bank_branch, bank_account, payroll_no, designation, department, sponsor_id,
		       date_first_appt, date_joined_scheme, expected_retirement, basic_salary
		FROM pending_member_registrations
		WHERE id = $1 AND scheme_id = $2 AND status = 'pending'
	`, id, schemeID).Scan(&pending.MemberNo, &pending.FirstName, &pending.LastName,
		&pending.OtherNames, &pending.Gender, &pending.DateOfBirth, &pending.Nationality,
		&pending.IDNumber, &pending.KRAPin, &pending.Email, &pending.Phone, &pending.PostalAddress,
		&pending.PostalCode, &pending.Town, &pending.MaritalStatus, &pending.SpouseName,
		&pending.NextOfKin, &pending.NextOfKinPhone, &pending.BankName, &pending.BankBranch,
		&pending.BankAccount, &pending.PayrollNo, &pending.Designation, &pending.Department,
		&pending.SponsorID, &pending.DateFirstAppt, &pending.DateJoinedScheme,
		&pending.ExpectedRetirement, &pending.BasicSalary)

	if err != nil {
		respondError(w, http.StatusNotFound, "pending member not found or already processed")
		return
	}

	var memberID string
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO members (scheme_id, member_no, first_name, last_name, other_names, gender,
			date_of_birth, nationality, id_number, kra_pin, email, phone, postal_address,
			postal_code, town, marital_status, spouse_name, next_of_kin, next_of_kin_phone,
			bank_name, bank_branch, bank_account, payroll_no, designation, department,
			sponsor_id, date_first_appt, date_joined_scheme, expected_retirement, basic_salary)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)
		RETURNING id
	`, schemeID, pending.MemberNo, pending.FirstName, pending.LastName, pending.OtherNames,
		pending.Gender, pending.DateOfBirth, pending.Nationality, pending.IDNumber, pending.KRAPin,
		pending.Email, pending.Phone, pending.PostalAddress, pending.PostalCode, pending.Town,
		pending.MaritalStatus, pending.SpouseName, pending.NextOfKin, pending.NextOfKinPhone,
		pending.BankName, pending.BankBranch, pending.BankAccount, pending.PayrollNo,
		pending.Designation, pending.Department, pending.SponsorID, pending.DateFirstAppt,
		pending.DateJoinedScheme, pending.ExpectedRetirement, pending.BasicSalary).Scan(&memberID)

	if err != nil {
		slog.Error("approve pending member failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create member")
		return
	}

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE pending_member_registrations SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2
	`, userID, id)
	if err != nil {
		slog.Error("update pending member status failed", "error", err)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "approved", "member_id": memberID})
}

func (s *Server) rejectPendingMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "rejection reason is required")
		return
	}

	result, err := s.db.ExecContext(r.Context(), `
		UPDATE pending_member_registrations SET status = 'rejected', rejection_reason = $1,
			reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3 AND scheme_id = $4 AND status = 'pending'
	`, req.Reason, userID, id, schemeID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to reject member")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "pending member not found or already processed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *Server) listPendingChanges(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT pmc.id, pmc.member_id, pmc.change_type, pmc.status, pmc.submitted_by,
		       pmc.created_at, pmc.reviewed_at, m.member_no, m.first_name, m.last_name
		FROM pending_member_changes pmc
		JOIN members m ON pmc.member_id = m.id
		WHERE m.scheme_id = $1 AND pmc.status = $2
		ORDER BY pmc.created_at DESC
	`, schemeID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, memberID, changeType, status, submittedBy string
		var createdAt, reviewedAt *time.Time
		var memberNo, firstName, lastName string
		if err := rows.Scan(&id, &memberID, &changeType, &status, &submittedBy, &createdAt, &reviewedAt, &memberNo, &firstName, &lastName); err != nil {
			continue
		}
		m := map[string]interface{}{
			"id": id, "member_id": memberID, "change_type": changeType, "status": status,
			"submitted_by": submittedBy, "created_at": createdAt,
			"member_no": memberNo, "member_name": firstName + " " + lastName,
		}
		if reviewedAt != nil {
			m["reviewed_at"] = *reviewedAt
		}
		results = append(results, m)
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, results)
}

func (s *Server) approveMemberChange(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	var change struct {
		MemberID   string
		ChangeType string
		BeforeJSON string
		AfterJSON  string
	}
	err := s.db.QueryRowContext(r.Context(), `
		SELECT member_id, change_type, before_values::text, after_values::text
		FROM pending_member_changes
		WHERE id = $1 AND status = 'pending'
	`, id).Scan(&change.MemberID, &change.ChangeType, &change.BeforeJSON, &change.AfterJSON)

	if err != nil {
		respondError(w, http.StatusNotFound, "pending change not found or already processed")
		return
	}

	var afterVals map[string]interface{}
	if err := json.Unmarshal([]byte(change.AfterJSON), &afterVals); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to parse after values")
		return
	}

	updates := []string{}
	args := []interface{}{}
	argCount := 0
	for field, value := range afterVals {
		argCount++
		updates = append(updates, fmt.Sprintf("%s = $%d", toSnakeCase(field), argCount))
		args = append(args, value)
	}

	if len(updates) > 0 {
		updates = append(updates, "updated_at = NOW()")
		args = append(args, change.MemberID)

		query := fmt.Sprintf("UPDATE members SET %s WHERE id = $%d", strings.Join(updates, ", "), argCount+1)
		_, err = s.db.ExecContext(r.Context(), query, args...)
		if err != nil {
			slog.Error("apply member change failed", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to apply change")
			return
		}
	}

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE pending_member_changes SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2
	`, userID, id)

	recordEvent(s.db, r.Context(), schemeID, "member", change.MemberID, "member_change_approved", afterVals, userID)

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (s *Server) rejectMemberChange(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "rejection reason is required")
		return
	}

	result, err := s.db.ExecContext(r.Context(), `
		UPDATE pending_member_changes SET status = 'rejected', rejection_reason = $1,
			reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`, req.Reason, userID, id)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to reject change")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "pending change not found or already processed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
