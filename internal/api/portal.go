package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pension-manager/internal/db"
	"pension-manager/internal/documents"
	"pension-manager/internal/portal"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// registerPortalRoutes registers member portal routes
func (s *Server) registerPortalRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))
		r.Use(MemberPortalMiddleware(s.db))

		r.Route("/api/portal", func(r chi.Router) {
			r.Get("/profile", s.handleGetProfile)
			r.Get("/beneficiaries", s.handleGetBeneficiaries)
			r.Get("/contributions", s.handleGetContributions)
			r.Get("/contributions/annual", s.handleGetAnnualContributions)
			r.Get("/change-requests", s.handleGetChangeRequests)
			r.Post("/change-requests/contact", s.handleRequestContactChange)
			r.Post("/change-requests/beneficiary", s.handleRequestBeneficiaryChange)
			r.Post("/feedback", s.handleSubmitFeedback)
			r.Get("/feedback", s.handleGetFeedback)
			r.Get("/login-stats", s.handleGetLoginStats)
			r.Get("/statement", s.handleGetStatement)
			r.Get("/statement/pdf", s.handleDownloadStatementPDF)
			r.Post("/projection", s.handleProjectBenefits)
			r.Get("/projection/quote", s.handleGetBenefitQuote)
			r.Post("/photo-upload", s.handleUploadPassportPhoto)
			r.Get("/documents", s.handleGetSchemeDocuments)
			r.Get("/documents/{id}/download", s.handleDownloadSchemeDocument)
		})
		// Admin-only portal management routes
		r.Group(func(admin chi.Router) {
			admin.Use(RoleMiddleware("super_admin", "admin"))
			admin.Put("/admin/members/{id}/portal", s.handleToggleMemberPortal)
			admin.Put("/admin/schemes/{id}/lock", s.handleLockScheme)
			admin.Get("/admin/members/utilization", s.handleGetMemberUtilization)
		})
	})
}

// MemberPortalMiddleware extracts member_id from the authenticated user context
func MemberPortalMiddleware(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			if userID == "" {
				respondError(w, http.StatusUnauthorized, "not authenticated")
				return
			}

			// Look up member_id from system_users
			var memberID sql.NullString
			err := database.QueryRowContext(r.Context(), `SELECT member_id FROM system_users WHERE id = $1`, userID).Scan(&memberID)
			if err != nil || !memberID.Valid || memberID.String == "" {
				respondError(w, http.StatusForbidden, "member portal access not enabled for this user")
				return
			}

			// Check if member is locked out
			var portalEnabled bool
			err = database.QueryRowContext(r.Context(), `SELECT portal_enabled FROM members WHERE id = $1`, memberID.String).Scan(&portalEnabled)
			if err == nil && !portalEnabled {
				respondError(w, http.StatusForbidden, "member portal access has been disabled by administrator")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "member_id", memberID.String)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// handleGetProfile handles GET /portal/profile
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	profile, err := s.portalService.GetMemberProfile(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Track login
	_ = s.portalService.TrackLogin(r.Context(), memberID)

	respondJSON(w, http.StatusOK, profile)
}

// handleGetBeneficiaries handles GET /portal/beneficiaries
func (s *Server) handleGetBeneficiaries(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	beneficiaries, err := s.portalService.GetMemberBeneficiaries(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, beneficiaries)
}

// handleGetContributions handles GET /portal/contributions
func (s *Server) handleGetContributions(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	contributions, err := s.portalService.GetMemberContributions(r.Context(), memberID, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, contributions)
}

// handleGetAnnualContributions handles GET /portal/contributions/annual
func (s *Server) handleGetAnnualContributions(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	annual, err := s.portalService.GetAnnualContributions(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, annual)
}

// handleGetChangeRequests handles GET /portal/change-requests
func (s *Server) handleGetChangeRequests(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	requests, err := s.portalService.GetMemberChangeRequests(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, requests)
}

// handleRequestContactChange handles POST /portal/change-requests/contact
func (s *Server) handleRequestContactChange(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	schemeID := GetSchemeID(r)

	var req struct {
		Email         string `json:"email,omitempty"`
		MobileNumber  string `json:"mobile_number,omitempty"`
		PostalAddress string `json:"postal_address,omitempty"`
		PostalCode    string `json:"postal_code,omitempty"`
		Town          string `json:"town,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get current values for before_data
	var currentEmail, currentPhone, currentAddr, currentCode, currentTown sql.NullString
	err := s.db.QueryRowContext(r.Context(), `
		SELECT email, phone, postal_address, postal_code, town FROM members WHERE id = $1
	`, memberID).Scan(&currentEmail, &currentPhone, &currentAddr, &currentCode, &currentTown)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get current data")
		return
	}

	beforeData := map[string]interface{}{
		"email":          currentEmail.String,
		"mobile_number":  currentPhone.String,
		"postal_address": currentAddr.String,
		"postal_code":    currentCode.String,
		"town":           currentTown.String,
	}

	afterData := map[string]interface{}{
		"email":          req.Email,
		"mobile_number":  req.MobileNumber,
		"postal_address": req.PostalAddress,
		"postal_code":    req.PostalCode,
		"town":           req.Town,
	}

	if err := s.portalService.CreateChangeRequest(r.Context(), memberID, schemeID, "contact_change", beforeData, afterData); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "change_request_submitted"})
}

// handleRequestBeneficiaryChange handles POST /portal/change-requests/beneficiary
func (s *Server) handleRequestBeneficiaryChange(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	schemeID := GetSchemeID(r)

	var req struct {
		Action        string  `json:"action"` // add, remove, update_allocation
		BeneficiaryID string  `json:"beneficiary_id,omitempty"`
		Name          string  `json:"name,omitempty"`
		Relationship  string  `json:"relationship,omitempty"`
		IDNumber      string  `json:"id_number,omitempty"`
		Phone         string  `json:"phone,omitempty"`
		Address       string  `json:"address,omitempty"`
		AllocationPct float64 `json:"allocation_pct,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Action == "" {
		respondError(w, http.StatusBadRequest, "action is required")
		return
	}

	afterData := map[string]interface{}{
		"action":         req.Action,
		"beneficiary_id": req.BeneficiaryID,
		"name":           req.Name,
		"relationship":   req.Relationship,
		"id_number":      req.IDNumber,
		"phone":          req.Phone,
		"address":        req.Address,
		"allocation_pct": req.AllocationPct,
	}

	if err := s.portalService.CreateChangeRequest(r.Context(), memberID, schemeID, "beneficiary_change", nil, afterData); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "change_request_submitted"})
}

// handleSubmitFeedback handles POST /portal/feedback
func (s *Server) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	schemeID := GetSchemeID(r)

	var req struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Subject == "" || req.Message == "" {
		respondError(w, http.StatusBadRequest, "subject and message are required")
		return
	}

	if err := s.portalService.SubmitFeedback(r.Context(), memberID, schemeID, req.Subject, req.Message); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "feedback_submitted"})
}

// handleGetFeedback handles GET /portal/feedback
func (s *Server) handleGetFeedback(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	feedbacks, err := s.portalService.GetMemberFeedback(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, feedbacks)
}

// handleGetLoginStats handles GET /portal/login-stats
func (s *Server) handleGetLoginStats(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	stats, err := s.portalService.GetMemberLoginStats(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// handleGetStatement handles GET /portal/statement
func (s *Server) handleGetStatement(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	profile, err := s.portalService.GetMemberProfile(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	contributions, err := s.portalService.GetMemberContributions(r.Context(), memberID, time.Time{}, time.Time{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	statement := map[string]interface{}{
		"member":        profile,
		"contributions": contributions,
		"generated_at":  time.Now(),
	}

	respondJSON(w, http.StatusOK, statement)
}

// handleDownloadStatementPDF handles GET /portal/statement/pdf
func (s *Server) handleDownloadStatementPDF(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	var memberNo, firstName, lastName, email, phone string
	var balance int64
	err := s.db.QueryRowContext(r.Context(), `
		SELECT member_no, first_name, last_name, email, phone, account_balance
		FROM members WHERE id = $1
	`, memberID).Scan(&memberNo, &firstName, &lastName, &email, &phone, &balance)
	if err != nil {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	contributions, err := s.portalService.GetMemberContributions(r.Context(), memberID, time.Time{}, time.Time{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pdf := generateMemberStatementPDF(memberNo, firstName, lastName, email, phone, balance, contributions)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=statement_%s_%s.pdf", memberNo, time.Now().Format("2006-01-02")))
	w.Write(pdf)
}

// handleToggleMemberPortal handles PUT /admin/members/{id}/portal
func (s *Server) handleToggleMemberPortal(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "id")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE members SET portal_enabled = $1, updated_at = NOW() WHERE id = $2
	`, req.Enabled, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update portal access")
		return
	}

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": status})
}

// handleLockScheme handles PUT /admin/schemes/{id}/lock
func (s *Server) handleLockScheme(w http.ResponseWriter, r *http.Request) {
	schemeID := chi.URLParam(r, "id")
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme ID is required")
		return
	}

	var req struct {
		Locked bool `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	status := "unlocked"
	if req.Locked {
		status = "locked"
	}

	_, err := s.db.ExecContext(r.Context(), `
		UPDATE schemes SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update scheme status")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": status})
}

// handleGetMemberUtilization handles GET /admin/members/utilization
func (s *Server) handleGetMemberUtilization(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	query := `
		SELECT m.id, m.member_no, m.first_name, m.last_name, m.email,
		       COUNT(DISTINCT mll.id) as total_logins,
		       MAX(mll.login_at) as last_login,
		       COUNT(DISTINCT mll.id) FILTER (WHERE mll.login_at > NOW() - INTERVAL '30 days') as logins_last_30_days
		FROM members m
		LEFT JOIN member_login_log mll ON mll.member_id = m.id
		WHERE m.scheme_id = $1
		GROUP BY m.id, m.member_no, m.first_name, m.last_name, m.email
		ORDER BY total_logins DESC
	`
	rows, err := s.db.QueryContext(r.Context(), query, schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type MemberUtilization struct {
		MemberID     string     `json:"member_id"`
		MemberNo     string     `json:"member_no"`
		FullName     string     `json:"full_name"`
		Email        string     `json:"email"`
		TotalLogins  int        `json:"total_logins"`
		LastLogin    *time.Time `json:"last_login,omitempty"`
		LoginsLast30 int        `json:"logins_last_30_days"`
	}

	var utilizations []MemberUtilization
	for rows.Next() {
		var u MemberUtilization
		var lastLogin sql.NullTime
		if err := rows.Scan(&u.MemberID, &u.MemberNo, &u.FullName, &u.Email, &u.TotalLogins, &lastLogin, &u.LoginsLast30); err != nil {
			continue
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		utilizations = append(utilizations, u)
	}

	respondJSON(w, http.StatusOK, utilizations)
}

// handleProjectBenefits handles POST /portal/projection
func (s *Server) handleProjectBenefits(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	var req struct {
		RetirementAge    int     `json:"retirement_age"`
		SalaryGrowthRate float64 `json:"salary_growth_rate"`
		InvestmentReturn float64 `json:"investment_return"`
		SchemeType       string  `json:"scheme_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RetirementAge == 0 {
		req.RetirementAge = 60
	}
	if req.SalaryGrowthRate == 0 {
		req.SalaryGrowthRate = 5.0
	}
	if req.InvestmentReturn == 0 {
		req.InvestmentReturn = 8.0
	}
	if req.SchemeType == "" {
		req.SchemeType = "dc"
	}

	var currentAge, basicSalary int
	var balance int64
	var memberRate, sponsorRate float64
	err := s.db.QueryRowContext(r.Context(), `
		SELECT DATE_PART('year', AGE(NOW(), date_of_birth))::int,
		       basic_salary, account_balance,
		       member_contribution_rate, sponsor_contribution_rate
		FROM members WHERE id = $1
	`, memberID).Scan(&currentAge, &basicSalary, &balance, &memberRate, &sponsorRate)
	if err != nil {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	if memberRate == 0 {
		memberRate = 5.0
	}
	if sponsorRate == 0 {
		sponsorRate = 10.0
	}

	params := portal.ProjectionParams{
		CurrentAge:       currentAge,
		RetirementAge:    req.RetirementAge,
		CurrentSalary:    int64(basicSalary),
		CurrentBalance:   balance,
		MemberRate:       memberRate,
		SponsorRate:      sponsorRate,
		SalaryGrowthRate: req.SalaryGrowthRate,
		InvestmentReturn: req.InvestmentReturn,
		SchemeType:       req.SchemeType,
		YearsOfService:   currentAge - 25,
	}

	result := portal.ProjectBenefits(params)
	respondJSON(w, http.StatusOK, result)
}

// handleGetBenefitQuote handles GET /portal/projection/quote
func (s *Server) handleGetBenefitQuote(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)

	var currentAge, basicSalary int
	var balance int64
	var memberRate, sponsorRate float64
	var schemeType string
	err := s.db.QueryRowContext(r.Context(), `
		SELECT DATE_PART('year', AGE(NOW(), date_of_birth))::int,
		       basic_salary, account_balance,
		       member_contribution_rate, sponsor_contribution_rate,
		       COALESCE(
		         (SELECT scheme_type FROM schemes s JOIN members m ON m.scheme_id = s.id WHERE m.id = $1),
		         'dc'
		       )
		FROM members WHERE id = $1
	`, memberID).Scan(&currentAge, &basicSalary, &balance, &memberRate, &sponsorRate, &schemeType)
	if err != nil {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	if memberRate == 0 {
		memberRate = 5.0
	}
	if sponsorRate == 0 {
		sponsorRate = 10.0
	}

	retirementAge := 60
	params := portal.ProjectionParams{
		CurrentAge:       currentAge,
		RetirementAge:    retirementAge,
		CurrentSalary:    int64(basicSalary),
		CurrentBalance:   balance,
		MemberRate:       memberRate,
		SponsorRate:      sponsorRate,
		SalaryGrowthRate: 5.0,
		InvestmentReturn: 8.0,
		SchemeType:       schemeType,
		YearsOfService:   currentAge - 25,
	}

	result := portal.ProjectBenefits(params)
	quote := portal.GenerateBenefitQuote(memberID, nil, result)
	respondJSON(w, http.StatusOK, quote)
}

// handleUploadPassportPhoto handles POST /portal/photo-upload
func (s *Server) handleUploadPassportPhoto(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	schemeID := GetSchemeID(r)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		respondError(w, http.StatusBadRequest, "photo file is required")
		return
	}
	defer file.Close()

	// Validate image type
	if header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png" {
		respondError(w, http.StatusBadRequest, "only JPEG and PNG images are allowed")
		return
	}

	// Upload to document storage
	doc := &documents.Document{
		ID:           uuid.New().String(),
		EntityType:   "member",
		EntityID:     memberID,
		SchemeID:     schemeID,
		DocumentType: "passport_photo",
		UploadedBy:   memberID,
	}

	if err := s.docService.UploadDocument(r.Context(), doc, file, header); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to upload photo: %v", err))
		return
	}

	// Create maker-checker request for photo update
	afterData := map[string]interface{}{
		"document_id": doc.ID,
		"file_name":   doc.FileName,
	}
	if err := s.portalService.CreateChangeRequest(r.Context(), memberID, schemeID, "photo_update", nil, afterData); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create change request: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"status":  "photo_uploaded",
		"message": "Photo uploaded successfully. Pending approval by pensions officer.",
	})
}

// handleGetSchemeDocuments handles GET /portal/documents
func (s *Server) handleGetSchemeDocuments(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	docs, err := s.docService.ListDocuments(r.Context(), "scheme", schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, docs)
}

// handleDownloadSchemeDocument handles GET /portal/documents/{id}/download
func (s *Server) handleDownloadSchemeDocument(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		respondError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	reader, doc, err := s.docService.DownloadDocument(r.Context(), documentID)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("document not found: %v", err))
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", doc.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.FileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", doc.FileSize))

	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// generateMemberStatementPDF generates a PDF member statement
func generateMemberStatementPDF(memberNo, firstName, lastName, email, phone string, balance int64, contributions []portal.MemberContribution) []byte {
	pdf := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n4 0 obj\n<< /Length 200 >>\nstream\nBT\n/F1 12 Tf\n72 720 Td\n(Member Statement) Tj\n0 -20 Td\n(Member: " + firstName + " " + lastName + ") Tj\n0 -20 Td\n(Member No: " + memberNo + ") Tj\n0 -20 Td\n(Balance: " + fmt.Sprintf("%d", balance) + ") Tj\n0 -20 Td\n(Contributions: " + fmt.Sprintf("%d", len(contributions)) + " records) Tj\nET\nendstream\nendobj\n5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000266 00000 n \n0000000517 00000 n \ntrailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n589\n%%EOF")
	return pdf
}
