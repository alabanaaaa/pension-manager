package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pension-manager/core/domain"
	"pension-manager/internal/auth"
	"pension-manager/internal/db"
	"pension-manager/internal/mpesa"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		OTP      string `json:"otp,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var userID, role, name, passwordHash string
	var schemeID sql.NullString
	var active, locked bool
	var failedLogins int
	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, scheme_id, role, name, password_hash, active, locked, failed_logins
		FROM system_users WHERE email = $1
	`, req.Email).Scan(&userID, &schemeID, &role, &name, &passwordHash, &active, &locked, &failedLogins)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if !active {
		respondError(w, http.StatusUnauthorized, "account is disabled")
		return
	}
	if locked {
		respondError(w, http.StatusLocked, "account is locked due to too many failed attempts. Contact admin to unlock.")
		return
	}
	if err := auth.CheckPassword(passwordHash, req.Password); err != nil {
		// Increment failed logins
		newFailed := failedLogins + 1
		if newFailed >= 5 {
			// Lock the account
			_, _ = s.db.ExecContext(r.Context(), `
				UPDATE system_users SET failed_logins = $1, locked = true, updated_at = NOW() WHERE id = $2
			`, newFailed, userID)
			respondError(w, http.StatusLocked, "account locked after too many failed attempts")
		} else {
			_, _ = s.db.ExecContext(r.Context(), `
				UPDATE system_users SET failed_logins = $1, updated_at = NOW() WHERE id = $2
			`, newFailed, userID)
			respondError(w, http.StatusUnauthorized, fmt.Sprintf("invalid credentials (%d attempts remaining)", 5-newFailed))
		}
		return
	}

	// Reset failed logins on successful login
	_, _ = s.db.ExecContext(r.Context(), `UPDATE system_users SET last_login = NOW(), failed_logins = 0 WHERE id = $1`, userID)

	// Verify OTP if provided
	if req.OTP != "" {
		var storedOTP string
		var otpExpiry time.Time
		err = s.db.QueryRowContext(r.Context(), `SELECT otp_code, otp_expiry FROM system_users WHERE id = $1`, userID).Scan(&storedOTP, &otpExpiry)
		if err != nil || storedOTP == "" || time.Now().After(otpExpiry) {
			respondError(w, http.StatusUnauthorized, "invalid or expired OTP")
			return
		}
		if storedOTP != req.OTP {
			respondError(w, http.StatusUnauthorized, "invalid OTP")
			return
		}
		// Clear OTP after successful use
		_, _ = s.db.ExecContext(r.Context(), `UPDATE system_users SET otp_code = NULL, otp_expiry = NULL WHERE id = $1`, userID)
	}

	accessToken, refreshToken, err := s.auth.GenerateToken(userID, schemeID.String, req.Email, role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	slog.Info("user logged in", "user_id", userID, "email", req.Email, "role", role)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       userID,
		"name":          name,
		"role":          role,
		"scheme_id":     schemeID.String,
	})
}

func (s *Server) memberLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MemberNo string `json:"member_no"`
		Pin      string `json:"pin"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MemberNo == "" || req.Pin == "" {
		respondError(w, http.StatusBadRequest, "member_no and pin are required")
		return
	}

	var memberID, schemeID, firstName, lastName, pinHash string
	var portalEnabled bool
	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, scheme_id, first_name, last_name, pin, COALESCE(portal_enabled, true)
		FROM members WHERE member_no = $1 AND membership_status = 'active'
	`, req.MemberNo).Scan(&memberID, &schemeID, &firstName, &lastName, &pinHash, &portalEnabled)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "invalid member number or inactive account")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if !portalEnabled {
		respondError(w, http.StatusForbidden, "member portal access has been disabled")
		return
	}
	if pinHash == "" {
		respondError(w, http.StatusUnauthorized, "PIN not set. Please contact administrator.")
		return
	}
	if err := auth.CheckPassword(pinHash, req.Pin); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	name := strings.TrimSpace(firstName + " " + lastName)
	accessToken, refreshToken, err := s.auth.GenerateToken(memberID, schemeID, "", "member")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	slog.Info("member logged in", "member_id", memberID, "member_no", req.MemberNo)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"member_id":     memberID,
		"name":          name,
		"role":          "member",
		"scheme_id":     schemeID,
	})
}

func (s *Server) requestOTP(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		respondError(w, http.StatusBadRequest, "email query parameter is required")
		return
	}

	var userID, phone string
	err := s.db.QueryRowContext(r.Context(), `SELECT id, phone FROM system_users WHERE email = $1`, email).Scan(&userID, &phone)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if phone == "" {
		respondError(w, http.StatusBadRequest, "user has no registered phone number")
		return
	}

	// Generate 6-digit OTP
	otp := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	expiry := time.Now().Add(5 * time.Minute)

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE system_users SET otp_code = $1, otp_expiry = $2, updated_at = NOW() WHERE id = $3
	`, otp, expiry, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate OTP")
		return
	}

	// In production, send via SMS gateway
	slog.Info("OTP generated", "phone", phone, "otp", otp, "expires_in", "5 minutes")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "OTP sent to your registered phone",
		"expires_in": 300,
		"phone":      maskPhone(phone),
	})
}

func (s *Server) unlockUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE system_users SET locked = false, failed_logins = 0, updated_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to unlock user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	return phone[:len(phone)-4] + "****"
}

func (s *Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	claims, err := s.auth.VerifyToken(req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	accessToken, _, err := s.auth.GenerateToken(claims.UserID, claims.SchemeID, claims.Email, claims.Role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	var totalMembers, activeMembers, totalContributions, pendingClaims int64
	var totalContribAmount int64
	_ = s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM members WHERE scheme_id = $1`, schemeID).Scan(&totalMembers)
	_ = s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM members WHERE scheme_id = $1 AND membership_status = 'active'`, schemeID).Scan(&activeMembers)
	_ = s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM contributions WHERE scheme_id = $1`, schemeID).Scan(&totalContributions)
	_ = s.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(total_amount), 0) FROM contributions WHERE scheme_id = $1 AND status = 'confirmed'`, schemeID).Scan(&totalContribAmount)
	_ = s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM claims WHERE scheme_id = $1 AND status IN ('submitted', 'under_review')`, schemeID).Scan(&pendingClaims)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_members":        totalMembers,
		"active_members":       activeMembers,
		"total_contributions":  totalContributions,
		"total_contrib_amount": totalContribAmount,
		"pending_claims":       pendingClaims,
	})
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	query := `SELECT id, member_no, first_name, last_name, id_number, phone, email, membership_status, department, account_balance, date_joined_scheme, created_at FROM members WHERE scheme_id = $1`
	args := []interface{}{schemeID}
	argCount := 1

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND membership_status = $%d", argCount)
		args = append(args, status)
	}
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR member_no ILIKE $%d OR id_number ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
	}
	query += fmt.Sprintf(" ORDER BY last_name, first_name LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var id, memberNo, firstName, lastName, idNumber, phone, email, st, dept string
		var balance int64
		var dateJoined, createdAt time.Time
		if err := rows.Scan(&id, &memberNo, &firstName, &lastName, &idNumber, &phone, &email, &st, &dept, &balance, &dateJoined, &createdAt); err != nil {
			continue
		}
		members = append(members, map[string]interface{}{
			"id": id, "member_no": memberNo, "first_name": firstName, "last_name": lastName,
			"id_number": idNumber, "phone": phone, "email": email, "membership_status": st,
			"department": dept, "account_balance": balance, "date_joined_scheme": dateJoined, "created_at": createdAt,
		})
	}
	if members == nil {
		members = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, members)
}

func (s *Server) createMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		MemberNo, FirstName, LastName, OtherNames, Gender, DateOfBirth, Nationality, IDNumber, KRAPIN    string
		Email, Phone, PostalAddress, PostalCode, Town, MaritalStatus, SpouseName                         string
		NextOfKin, NextOfKinPhone, BankName, BankBranch, BankAccount, PayrollNo, Designation, Department string
		SponsorID, DateFirstAppt, DateJoinedScheme, ExpectedRetire                                       string
		BasicSalary                                                                                      int64
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FirstName == "" || req.LastName == "" || req.MemberNo == "" || req.DateOfBirth == "" || req.DateJoinedScheme == "" {
		respondError(w, http.StatusBadRequest, "first_name, last_name, member_no, date_of_birth, and date_joined_scheme are required")
		return
	}

	dob, _ := time.Parse("2006-01-02", req.DateOfBirth)
	djs, _ := time.Parse("2006-01-02", req.DateJoinedScheme)
	expectedRetire, _ := time.Parse("2006-01-02", req.ExpectedRetire)
	dateFirstAppt, _ := time.Parse("2006-01-02", req.DateFirstAppt)

	var memberID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO members (scheme_id, member_no, first_name, last_name, other_names, gender,
			date_of_birth, nationality, id_number, kra_pin, email, phone, postal_address,
			postal_code, town, marital_status, spouse_name, next_of_kin, next_of_kin_phone,
			bank_name, bank_branch, bank_account, payroll_no, designation, department,
			sponsor_id, date_first_appt, date_joined_scheme, expected_retirement, basic_salary)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)
		RETURNING id
	`, schemeID, req.MemberNo, req.FirstName, req.LastName, req.OtherNames, req.Gender,
		dob, req.Nationality, req.IDNumber, req.KRAPIN, req.Email, req.Phone,
		req.PostalAddress, req.PostalCode, req.Town, req.MaritalStatus, req.SpouseName,
		req.NextOfKin, req.NextOfKinPhone, req.BankName, req.BankBranch, req.BankAccount,
		req.PayrollNo, req.Designation, req.Department, req.SponsorID, dateFirstAppt,
		djs, expectedRetire, req.BasicSalary).Scan(&memberID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondError(w, http.StatusConflict, "member number or ID number already exists")
			return
		}
		slog.Error("create member failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create member")
		return
	}

	recordEvent(s.db, r.Context(), schemeID, "member", memberID, "member_created", map[string]interface{}{
		"first_name": req.FirstName, "last_name": req.LastName, "member_no": req.MemberNo,
	}, userID)

	slog.Info("member created", "id", memberID, "name", req.FirstName+" "+req.LastName)
	respondCreated(w, map[string]interface{}{"id": memberID, "member_no": req.MemberNo, "first_name": req.FirstName, "last_name": req.LastName})
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	memberID := chi.URLParam(r, "id")

	var m domain.Member
	var dob, djs, er, dfa sql.NullTime
	var idNum, kraPin, email, phone, spouse, bankAcct sql.NullString
	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, scheme_id, member_no, first_name, last_name, other_names, gender,
			date_of_birth, nationality, id_number, kra_pin, email, phone, marital_status,
			spouse_name, bank_name, bank_branch, bank_account, payroll_no, designation,
			department, date_first_appt, date_joined_scheme, expected_retirement,
			membership_status, basic_salary, account_balance, last_contribution, created_at, updated_at
		FROM members WHERE id = $1 AND scheme_id = $2
	`, memberID, schemeID).Scan(
		&m.ID, &m.SchemeID, &m.MemberNo, &m.FirstName, &m.LastName, &m.OtherNames,
		&m.Gender, &dob, &m.Nationality, &idNum, &kraPin, &email, &phone,
		&m.MaritalStatus, &spouse, &m.BankName, &m.BankBranch, &bankAcct,
		&m.PayrollNo, &m.Designation, &m.Department, &dfa, &djs, &er,
		&m.MembershipStatus, &m.BasicSalary, &m.AccountBalance, &m.LastContribution,
		&m.CreatedAt, &m.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	if dob.Valid {
		m.DateOfBirth = dob.Time
	}
	if idNum.Valid {
		m.IDNumber = idNum.String
	}
	if kraPin.Valid {
		m.KRAPIN = kraPin.String
	}
	if email.Valid {
		m.Email = email.String
	}
	if phone.Valid {
		m.Phone = phone.String
	}
	if spouse.Valid {
		m.SpouseName = spouse.String
	}
	if bankAcct.Valid {
		m.BankAccount = bankAcct.String
	}
	respondJSON(w, http.StatusOK, m)
}

func (s *Server) updateMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	memberID := chi.URLParam(r, "id")

	var oldFirst, oldLast, oldPhone, oldEmail string
	err := s.db.QueryRowContext(r.Context(), `SELECT first_name, last_name, phone, email FROM members WHERE id = $1 AND scheme_id = $2`, memberID, schemeID).Scan(&oldFirst, &oldLast, &oldPhone, &oldEmail)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	var req struct {
		FirstName, LastName, Phone, Email, BankName, BankBranch, BankAccount, Department, Designation string
		BasicSalary                                                                                   *int64
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := []string{}
	args := []interface{}{}
	argCount := 0

	if req.FirstName != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("first_name = $%d", argCount))
		args = append(args, req.FirstName)
	}
	if req.LastName != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("last_name = $%d", argCount))
		args = append(args, req.LastName)
	}
	if req.Phone != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("phone = $%d", argCount))
		args = append(args, req.Phone)
	}
	if req.Email != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("email = $%d", argCount))
		args = append(args, req.Email)
	}
	if req.BankName != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("bank_name = $%d", argCount))
		args = append(args, req.BankName)
	}
	if req.BankBranch != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("bank_branch = $%d", argCount))
		args = append(args, req.BankBranch)
	}
	if req.BankAccount != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("bank_account = $%d", argCount))
		args = append(args, req.BankAccount)
	}
	if req.Department != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("department = $%d", argCount))
		args = append(args, req.Department)
	}
	if req.Designation != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("designation = $%d", argCount))
		args = append(args, req.Designation)
	}
	if req.BasicSalary != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("basic_salary = $%d", argCount))
		args = append(args, *req.BasicSalary)
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	updates = append(updates, fmt.Sprintf("updated_at = NOW()"))
	argCount++
	args = append(args, memberID)
	argCount++
	args = append(args, schemeID)

	query := fmt.Sprintf("UPDATE members SET %s WHERE id = $%d AND scheme_id = $%d", strings.Join(updates, ", "), argCount-1, argCount)
	result, err := s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		slog.Error("update member failed", "error", err)
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	recordEvent(s.db, r.Context(), schemeID, "member", memberID, "member_updated", map[string]interface{}{
		"old": map[string]string{"first_name": oldFirst, "last_name": oldLast, "phone": oldPhone, "email": oldEmail},
		"new": map[string]string{"first_name": req.FirstName, "last_name": req.LastName, "phone": req.Phone, "email": req.Email},
	}, userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "updated", "affected": rows})
}

func (s *Server) deactivateMember(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	memberID := chi.URLParam(r, "id")

	result, err := s.db.ExecContext(r.Context(), `UPDATE members SET membership_status = 'inactive', updated_at = NOW() WHERE id = $1 AND scheme_id = $2`, memberID, schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}
	recordEvent(s.db, r.Context(), schemeID, "member", memberID, "member_deactivated", nil, userID)
	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "deactivated", "affected": rows})
}

func (s *Server) listBeneficiaries(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, name, relationship, date_of_birth, id_number, phone, physical_address, allocation_pct, status, created_at
		FROM beneficiaries WHERE member_id = $1 ORDER BY allocation_pct DESC
	`, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var beneficiaries []map[string]interface{}
	for rows.Next() {
		var id, name, rel, phone, addr, status string
		var dob, idNum sql.NullString
		var alloc float64
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &rel, &dob, &idNum, &phone, &addr, &alloc, &status, &createdAt); err != nil {
			continue
		}
		b := map[string]interface{}{"id": id, "name": name, "relationship": rel, "phone": phone, "physical_address": addr, "allocation_pct": alloc, "status": status, "created_at": createdAt}
		if dob.Valid {
			b["date_of_birth"] = dob.String
		}
		if idNum.Valid {
			b["id_number"] = idNum.String
		}
		beneficiaries = append(beneficiaries, b)
	}
	if beneficiaries == nil {
		beneficiaries = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, beneficiaries)
}

func (s *Server) addBeneficiary(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Name, Relationship, DateOfBirth, IDNumber, Phone, PhysicalAddress string
		AllocationPct                                                     float64
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Relationship == "" {
		respondError(w, http.StatusBadRequest, "name and relationship are required")
		return
	}
	if req.AllocationPct < 0 || req.AllocationPct > 100 {
		respondError(w, http.StatusBadRequest, "allocation_pct must be between 0 and 100")
		return
	}

	var benID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO beneficiaries (member_id, name, relationship, date_of_birth, id_number, phone, physical_address, allocation_pct)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id
	`, memberID, req.Name, req.Relationship, req.DateOfBirth, req.IDNumber, req.Phone, req.PhysicalAddress, req.AllocationPct).Scan(&benID)

	if err != nil {
		slog.Error("add beneficiary failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to add beneficiary")
		return
	}

	recordEvent(s.db, r.Context(), "", "beneficiary", benID, "beneficiary_added", map[string]interface{}{
		"member_id": memberID, "name": req.Name, "relationship": req.Relationship, "allocation_pct": req.AllocationPct,
	}, userID)

	respondCreated(w, map[string]interface{}{"id": benID, "name": req.Name, "allocation_pct": req.AllocationPct})
}

func (s *Server) recordContribution(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		MemberID       string `json:"member_id"`
		SponsorID      string `json:"sponsor_id,omitempty"`
		Period         string `json:"period"`
		EmployeeAmount int64  `json:"employee_amount"`
		EmployerAmount int64  `json:"employer_amount"`
		AVCAmount      int64  `json:"avc_amount"`
		PaymentMethod  string `json:"payment_method,omitempty"`
		PaymentRef     string `json:"payment_ref,omitempty"`
		ReceiptNo      string `json:"receipt_no,omitempty"`
		Registered     *bool  `json:"registered"`
		Notes          string `json:"notes,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MemberID == "" || req.Period == "" {
		respondError(w, http.StatusBadRequest, "member_id and period are required")
		return
	}
	if req.EmployeeAmount < 0 || req.EmployerAmount < 0 || req.AVCAmount < 0 {
		respondError(w, http.StatusBadRequest, "amounts cannot be negative")
		return
	}
	total := req.EmployeeAmount + req.EmployerAmount + req.AVCAmount
	if total == 0 {
		respondError(w, http.StatusBadRequest, "total contribution must be greater than zero")
		return
	}

	period, _ := time.Parse("2006-01-02", req.Period)
	registered := true
	if req.Registered != nil {
		registered = *req.Registered
	}

	var contribID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO contributions (member_id, scheme_id, sponsor_id, period, employee_amount, employer_amount, avc_amount, total_amount, payment_method, payment_ref, receipt_no, status, registered, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id
	`, req.MemberID, schemeID, req.SponsorID, period, req.EmployeeAmount, req.EmployerAmount, req.AVCAmount, total, req.PaymentMethod, req.PaymentRef, req.ReceiptNo, "confirmed", registered, req.Notes, userID).Scan(&contribID)

	if err != nil {
		slog.Error("record contribution failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to record contribution")
		return
	}

	// Update member balance
	_, _ = s.db.ExecContext(r.Context(), `
		UPDATE members SET account_balance = account_balance + $1, last_contribution = NOW() WHERE id = $2
	`, total, req.MemberID)

	recordEvent(s.db, r.Context(), schemeID, "contribution", contribID, "contribution_recorded", map[string]interface{}{
		"member_id": req.MemberID, "employee_amount": req.EmployeeAmount, "employer_amount": req.EmployerAmount,
		"avc_amount": req.AVCAmount, "total": total, "period": req.Period,
	}, userID)

	slog.Info("contribution recorded", "id", contribID, "member", req.MemberID, "total", total)
	respondCreated(w, map[string]interface{}{"id": contribID, "total": total, "status": "confirmed"})
}

func (s *Server) listContributions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	status := r.URL.Query().Get("status")

	query := `SELECT id, member_id, period, employee_amount, employer_amount, avc_amount, total_amount, payment_method, status, registered, created_at FROM contributions WHERE scheme_id = $1`
	args := []interface{}{schemeID}
	argCount := 1

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	query += fmt.Sprintf(" ORDER BY period DESC LIMIT $%d", argCount+1)
	args = append(args, limit)

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var contributions []map[string]interface{}
	for rows.Next() {
		var id, memberID, paymentMethod, status string
		var empAmt, erAmt, avcAmt, total int64
		var period, createdAt time.Time
		var registered bool
		if err := rows.Scan(&id, &memberID, &period, &empAmt, &erAmt, &avcAmt, &total, &paymentMethod, &status, &registered, &createdAt); err != nil {
			continue
		}
		contributions = append(contributions, map[string]interface{}{
			"id": id, "member_id": memberID, "period": period, "employee_amount": empAmt,
			"employer_amount": erAmt, "avc_amount": avcAmt, "total_amount": total,
			"payment_method": paymentMethod, "status": status, "registered": registered, "created_at": createdAt,
		})
	}
	if contributions == nil {
		contributions = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, contributions)
}

func (s *Server) memberContributions(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, period, employee_amount, employer_amount, avc_amount, total_amount, payment_method, status, created_at
		FROM contributions WHERE member_id = $1 ORDER BY period DESC LIMIT 100
	`, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var contributions []map[string]interface{}
	for rows.Next() {
		var id, paymentMethod, status string
		var empAmt, erAmt, avcAmt, total int64
		var period, createdAt time.Time
		if err := rows.Scan(&id, &period, &empAmt, &erAmt, &avcAmt, &total, &paymentMethod, &status, &createdAt); err != nil {
			continue
		}
		contributions = append(contributions, map[string]interface{}{
			"id": id, "period": period, "employee_amount": empAmt, "employer_amount": erAmt,
			"avc_amount": avcAmt, "total_amount": total, "payment_method": paymentMethod,
			"status": status, "created_at": createdAt,
		})
	}
	if contributions == nil {
		contributions = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, contributions)
}

func (s *Server) mpesaContribution(w http.ResponseWriter, r *http.Request) {
	_ = GetSchemeID(r)

	var req struct {
		PhoneNumber string `json:"phone_number"`
		MemberID    string `json:"member_id"`
		Amount      int64  `json:"amount"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PhoneNumber == "" || req.MemberID == "" || req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "phone_number, member_id, and amount (>0) are required")
		return
	}
	if s.mpesaClient == nil {
		respondError(w, http.StatusServiceUnavailable, "M-Pesa not configured")
		return
	}

	accountRef := "PENSION-" + req.MemberID[:8]
	resp, err := s.mpesaClient.STKPush(mpesa.STKPushRequest{
		PhoneNumber: req.PhoneNumber,
		Amount:      req.Amount,
		AccountRef:  accountRef,
		Description: fmt.Sprintf("Pension Contribution - %s", req.MemberID),
	})
	if err != nil {
		slog.Error("STK push failed", "error", err)
		respondError(w, http.StatusPaymentRequired, "Payment initiation failed")
		return
	}
	if resp.ResponseCode != "0" {
		respondError(w, http.StatusPaymentRequired, resp.ResponseDesc)
		return
	}

	slog.Info("M-Pesa STK push initiated", "checkout_id", resp.CheckoutRequestID, "amount", req.Amount)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "checkout_id": resp.CheckoutRequestID,
		"message": "Payment initiated. Check your phone for STK push.", "amount": req.Amount,
	})
}

func (s *Server) mpesaCallback(w http.ResponseWriter, r *http.Request) {
	var callback struct {
		Body struct {
			STKCallback struct {
				MerchantRequestID string `json:"MerchantRequestID"`
				CheckoutRequestID string `json:"CheckoutRequestID"`
				ResultCode        int    `json:"ResultCode"`
				ResultDesc        string `json:"ResultDesc"`
				CallbackMetadata  struct {
					Item []struct {
						Name  string      `json:"Name"`
						Value interface{} `json:"Value"`
					} `json:"Item"`
				} `json:"CallbackMetadata"`
			} `json:"stkCallback"`
		} `json:"Body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		slog.Error("parse M-Pesa callback failed", "error", err)
		http.Error(w, "invalid callback", http.StatusBadRequest)
		return
	}

	resultCode := callback.Body.STKCallback.ResultCode
	checkoutID := callback.Body.STKCallback.CheckoutRequestID

	if resultCode != 0 {
		slog.Info("M-Pesa payment cancelled/failed",
			"checkout_id", checkoutID,
			"result_code", resultCode,
			"result_desc", callback.Body.STKCallback.ResultDesc,
		)
		// Update contribution status to rejected
		_, err := s.db.ExecContext(r.Context(), `
			UPDATE contributions SET status = 'rejected', updated_at = NOW()
			WHERE mpesa_checkout_id = $1
		`, checkoutID)
		if err != nil {
			slog.Error("failed to update contribution status", "error", err)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
		return
	}

	// Extract callback metadata
	var amount float64
	var mpesaReceipt string
	var phoneNumber string
	var transactionID string

	for _, item := range callback.Body.STKCallback.CallbackMetadata.Item {
		switch item.Name {
		case "MpesaReceiptNumber":
			if v, ok := item.Value.(string); ok {
				mpesaReceipt = v
			}
		case "PhoneNumber":
			if v, ok := item.Value.(float64); ok {
				phoneNumber = fmt.Sprintf("%.0f", v)
			}
		case "Amount":
			if v, ok := item.Value.(float64); ok {
				amount = v
			}
		case "TransactionId":
			if v, ok := item.Value.(string); ok {
				transactionID = v
			}
		}
	}

	slog.Info("M-Pesa payment successful",
		"receipt", mpesaReceipt,
		"amount", amount,
		"phone", phoneNumber,
		"checkout_id", checkoutID,
	)

	// Update contribution status to confirmed
	_, err := s.db.ExecContext(r.Context(), `
		UPDATE contributions
		SET status = 'confirmed', mpesa_receipt = $1, phone_number = $2,
		    transaction_id = $3, confirmed_at = NOW(), updated_at = NOW()
		WHERE mpesa_checkout_id = $4
	`, mpesaReceipt, phoneNumber, transactionID, checkoutID)
	if err != nil {
		slog.Error("failed to update contribution after M-Pesa callback", "error", err)
	}

	// Update member balance
	_, err = s.db.ExecContext(r.Context(), `
		UPDATE members SET account_balance = account_balance + $1, updated_at = NOW()
		WHERE id = (SELECT member_id FROM contributions WHERE mpesa_checkout_id = $2)
	`, int64(amount*100), checkoutID) // Convert to cents
	if err != nil {
		slog.Error("failed to update member balance", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "success",
		"receipt":        mpesaReceipt,
		"transaction_id": transactionID,
	})
}

func (s *Server) reconcileContributions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var req struct {
		SponsorID string `json:"sponsor_id"`
		Period    string `json:"period"`
		TotalAmt  int64  `json:"total_amount"`
		TotalEmp  int    `json:"total_employees"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	period, _ := time.Parse("2006-01-02", req.Period)
	var prevAmt int64
	var prevEmp int
	_ = s.db.QueryRowContext(r.Context(), `
		SELECT COALESCE(SUM(total_amount), 0), COUNT(DISTINCT member_id)
		FROM contributions WHERE scheme_id = $1 AND sponsor_id = $2 AND period < $3 AND status = 'confirmed'
		ORDER BY period DESC LIMIT 1
	`, schemeID, req.SponsorID, period).Scan(&prevAmt, &prevEmp)

	empDiff := req.TotalEmp - prevEmp
	amtDiff := req.TotalAmt - prevAmt

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sponsor_id":           req.SponsorID,
		"period":               req.Period,
		"expected_amount":      req.TotalAmt,
		"expected_employees":   req.TotalEmp,
		"previous_amount":      prevAmt,
		"previous_employees":   prevEmp,
		"employee_difference":  empDiff,
		"amount_difference":    amtDiff,
		"discrepancy_detected": amtDiff != 0 || empDiff != 0,
	})
}

func (s *Server) quarterlyReport(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	year := r.URL.Query().Get("year")
	quarter := r.URL.Query().Get("quarter")

	if year == "" || quarter == "" {
		respondError(w, http.StatusBadRequest, "year and quarter are required")
		return
	}

	var totalContrib, totalEmployee, totalEmployer, totalAVC int64
	var memberCount int
	err := s.db.QueryRowContext(r.Context(), `
		SELECT COUNT(DISTINCT member_id), COALESCE(SUM(total_amount), 0),
			COALESCE(SUM(employee_amount), 0), COALESCE(SUM(employer_amount), 0), COALESCE(SUM(avc_amount), 0)
		FROM contributions WHERE scheme_id = $1 AND EXTRACT(YEAR FROM period) = $2
		AND EXTRACT(QUARTER FROM period) = $3 AND status = 'confirmed'
	`, schemeID, year, quarter).Scan(&memberCount, &totalContrib, &totalEmployee, &totalEmployer, &totalAVC)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":               year,
		"quarter":            quarter,
		"total_members":      memberCount,
		"total_contribution": totalContrib,
		"total_employee":     totalEmployee,
		"total_employer":     totalEmployer,
		"total_avc":          totalAVC,
	})
}

func (s *Server) contributionReport(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	query := `SELECT EXTRACT(YEAR FROM period)::int as year, EXTRACT(MONTH FROM period)::int as month,
		COUNT(*) as count, COALESCE(SUM(total_amount), 0) as total,
		COALESCE(SUM(employee_amount), 0) as employee, COALESCE(SUM(employer_amount), 0) as employer
		FROM contributions WHERE scheme_id = $1 AND status = 'confirmed'`
	args := []interface{}{schemeID}
	argCount := 1

	if from != "" {
		argCount++
		query += fmt.Sprintf(" AND period >= $%d", argCount)
		args = append(args, from)
	}
	if to != "" {
		argCount++
		query += fmt.Sprintf(" AND period <= $%d", argCount)
		args = append(args, to)
	}
	query += " GROUP BY year, month ORDER BY year, month"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var report []map[string]interface{}
	for rows.Next() {
		var year, month, count int
		var total, emp, er int64
		if err := rows.Scan(&year, &month, &count, &total, &emp, &er); err != nil {
			continue
		}
		report = append(report, map[string]interface{}{
			"year": year, "month": month, "count": count,
			"total": total, "employee": emp, "employer": er,
		})
	}
	if report == nil {
		report = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	exportType := r.URL.Query().Get("type")
	if exportType == "" {
		exportType = "members"
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	switch exportType {
	case "members":
		writer.Write([]string{"member_no", "first_name", "last_name", "id_number", "phone", "email", "status", "department", "balance", "date_joined"})
		rows, err := s.db.QueryContext(r.Context(), `
			SELECT member_no, first_name, last_name, id_number, phone, email, membership_status, department, account_balance, date_joined_scheme
			FROM members WHERE scheme_id = $1 ORDER BY last_name
		`, schemeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var mn, fn, ln, idn, ph, em, st, dept string
			var bal int64
			var dj time.Time
			if err := rows.Scan(&mn, &fn, &ln, &idn, &ph, &em, &st, &dept, &bal, &dj); err != nil {
				continue
			}
			writer.Write([]string{mn, fn, ln, idn, ph, em, st, dept, fmt.Sprintf("%d", bal), dj.Format("2006-01-02")})
		}
	case "contributions":
		writer.Write([]string{"member_id", "period", "employee", "employer", "avc", "total", "status", "payment_method"})
		rows, err := s.db.QueryContext(r.Context(), `
			SELECT member_id, period, employee_amount, employer_amount, avc_amount, total_amount, status, payment_method
			FROM contributions WHERE scheme_id = $1 ORDER BY period DESC
		`, schemeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "query failed")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var mid, pm, st string
			var emp, er, avc, total int64
			var period time.Time
			if err := rows.Scan(&mid, &period, &emp, &er, &avc, &total, &st, &pm); err != nil {
				continue
			}
			writer.Write([]string{mid, period.Format("2006-01"), fmt.Sprintf("%d", emp), fmt.Sprintf("%d", er), fmt.Sprintf("%d", avc), fmt.Sprintf("%d", total), st, pm})
		}
	default:
		respondError(w, http.StatusBadRequest, "unknown export type: "+exportType)
		return
	}

	writer.Flush()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.csv", exportType, time.Now().Format("2006-01-02")))
	w.Write([]byte(buf.String()))
}

func (s *Server) ghostReport(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var flaggedMembers int
	var duplicateClaims int
	var unusualContributions int

	_ = s.db.QueryRowContext(r.Context(), `
		SELECT COUNT(DISTINCT member_id) FROM contributions WHERE scheme_id = $1 AND status = 'rejected'
	`, schemeID).Scan(&flaggedMembers)
	_ = s.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM claims WHERE scheme_id = $1 AND status = 'rejected'
	`, schemeID).Scan(&duplicateClaims)
	_ = s.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM contributions WHERE scheme_id = $1 AND total_amount > 1000000
	`, schemeID).Scan(&unusualContributions)

	riskScore := 0
	if flaggedMembers > 0 {
		riskScore += 20
	}
	if duplicateClaims > 5 {
		riskScore += 30
	}
	if unusualContributions > 0 {
		riskScore += 25
	}
	if riskScore > 100 {
		riskScore = 100
	}

	var anomalies []map[string]interface{}
	if flaggedMembers > 0 {
		anomalies = append(anomalies, map[string]interface{}{"type": "rejected_contributions", "count": flaggedMembers, "severity": "medium"})
	}
	if duplicateClaims > 5 {
		anomalies = append(anomalies, map[string]interface{}{"type": "high_claim_rejection_rate", "count": duplicateClaims, "severity": "high"})
	}
	if unusualContributions > 0 {
		anomalies = append(anomalies, map[string]interface{}{"type": "unusually_large_contributions", "count": unusualContributions, "severity": "low"})
	}
	if anomalies == nil {
		anomalies = []map[string]interface{}{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"risk_score":            riskScore,
		"flagged_members":       flaggedMembers,
		"duplicate_claims":      duplicateClaims,
		"unusual_contributions": unusualContributions,
		"anomalies":             anomalies,
		"generated_at":          time.Now(),
	})
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, email, role, name, phone, active, locked, last_login, created_at
		FROM system_users ORDER BY name
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, email, role, name, phone string
		var active, locked bool
		var lastLogin sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &role, &name, &phone, &active, &locked, &lastLogin, &createdAt); err != nil {
			continue
		}
		u := map[string]interface{}{"id": id, "email": email, "role": role, "name": name, "phone": phone, "active": active, "locked": locked, "created_at": createdAt}
		if lastLogin.Valid {
			u["last_login"] = lastLogin.Time
		}
		users = append(users, u)
	}
	if users == nil {
		users = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, users)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email, Password, Role, Name, Phone string
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" || req.Role == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "email, password, role, and name are required")
		return
	}

	validRoles := map[string]bool{"super_admin": true, "admin": true, "pension_officer": true, "claims_examiner": true, "auditor": true, "member": true}
	if !validRoles[req.Role] {
		respondError(w, http.StatusBadRequest, "invalid role")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var userID string
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO system_users (email, password_hash, role, name, phone) VALUES ($1,$2,$3,$4,$5) RETURNING id
	`, req.Email, hash, req.Role, req.Name, req.Phone).Scan(&userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	respondCreated(w, map[string]interface{}{"id": userID, "email": req.Email, "role": req.Role, "name": req.Name})
}

func (s *Server) updateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req struct {
		Role string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validRoles := map[string]bool{"super_admin": true, "admin": true, "pension_officer": true, "claims_examiner": true, "auditor": true, "member": true}
	if !validRoles[req.Role] {
		respondError(w, http.StatusBadRequest, "invalid role")
		return
	}

	_, err := s.db.ExecContext(r.Context(), `UPDATE system_users SET role = $1, updated_at = NOW() WHERE id = $2`, req.Role, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated", "role": req.Role})
}

func (s *Server) disableUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	result, err := s.db.ExecContext(r.Context(), `UPDATE system_users SET active = false, locked = true, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	rows, _ := result.RowsAffected()
	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "disabled", "affected": rows})
}

func recordEvent(db *db.DB, ctx context.Context, schemeID, entityType, entityID, eventType string, data interface{}, createdBy string) {
	dataJSON, _ := json.Marshal(data)
	var lastHash string
	_ = db.QueryRowContext(ctx, `SELECT COALESCE(event_hash, '') FROM events WHERE scheme_id = $1 ORDER BY event_seq DESC LIMIT 1`, schemeID).Scan(&lastHash)
	var seq int64
	_ = db.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_seq), 0) FROM events WHERE scheme_id = $1`, schemeID).Scan(&seq)
	seq++
	content := string(dataJSON) + lastHash
	sum := sha256.Sum256([]byte(content))
	eventHash := hex.EncodeToString(sum[:])
	_, _ = db.ExecContext(ctx, `
		INSERT INTO events (scheme_id, entity_type, entity_id, event_seq, event_type, event_data, previous_hash, event_hash, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, schemeID, entityType, entityID, seq, eventType, string(dataJSON), lastHash, eventHash, createdBy)
}
