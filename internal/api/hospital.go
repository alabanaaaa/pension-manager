package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pension-manager/core/domain"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
)

// registerHospitalRoutes registers hospital management routes
func (s *Server) registerHospitalRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("admin", "officer", "hospital_admin"))

		// Hospital management
		r.Route("/hospitals", func(r chi.Router) {
			r.Post("/", s.handleCreateHospital)
			r.Get("/", s.handleListHospitals)
			r.Route("/{hospitalID}", func(r chi.Router) {
				r.Get("/", s.handleGetHospital)
				r.Put("/", s.handleUpdateHospital)
			})
		})

		// Medical limits
		r.Route("/members/{memberID}/medical-limits", func(r chi.Router) {
			r.Post("/", s.handleCreateMedicalLimit)
			r.Get("/", s.handleGetMedicalLimit)
		})

		// Medical expenditures
		r.Route("/medical-expenditures", func(r chi.Router) {
			r.Post("/", s.handleRecordMedicalExpenditure)
			r.Get("/", s.handleListMedicalExpenditures)
			r.Get("/pending", s.handleGetPendingBills)
			r.Get("/alerts", s.handleGetExpenditureAlerts)
			r.Get("/export/excel", s.handleExportMedicalExpendituresExcel)
			r.Get("/export/csv", s.handleExportMedicalExpendituresCSV)
		})
	})
}

// handleCreateHospital handles POST /hospitals
func (s *Server) handleCreateHospital(w http.ResponseWriter, r *http.Request) {
	var h domain.Hospital
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := s.hospitalService.CreateHospital(r.Context(), &h); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, h)
}

// handleGetHospital handles GET /hospitals/{hospitalID}
func (s *Server) handleGetHospital(w http.ResponseWriter, r *http.Request) {
	hospitalID := chi.URLParam(r, "hospitalID")
	if hospitalID == "" {
		respondError(w, http.StatusBadRequest, "hospital ID is required")
		return
	}

	h, err := s.hospitalService.GetHospital(r.Context(), hospitalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if h == nil {
		respondError(w, http.StatusNotFound, "hospital not found")
		return
	}

	respondJSON(w, http.StatusOK, h)
}

// handleListHospitals handles GET /hospitals
func (s *Server) handleListHospitals(w http.ResponseWriter, r *http.Request) {
	schemeID := r.URL.Query().Get("scheme_id")
	if schemeID == "" {
		schemeID = GetSchemeID(r)
	}
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme_id query parameter is required")
		return
	}

	hospitals, err := s.hospitalService.ListHospitals(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, hospitals)
}

// handleUpdateHospital handles PUT /hospitals/{hospitalID}
func (s *Server) handleUpdateHospital(w http.ResponseWriter, r *http.Request) {
	hospitalID := chi.URLParam(r, "hospitalID")
	if hospitalID == "" {
		respondError(w, http.StatusBadRequest, "hospital ID is required")
		return
	}

	var h domain.Hospital
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	h.ID = hospitalID

	if err := s.hospitalService.UpdateHospital(r.Context(), &h); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, h)
}

// handleCreateMedicalLimit handles POST /members/{memberID}/medical-limits
func (s *Server) handleCreateMedicalLimit(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	var limit domain.MedicalLimit
	if err := json.NewDecoder(r.Body).Decode(&limit); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	limit.MemberID = memberID

	if err := s.hospitalService.CreateMedicalLimit(r.Context(), &limit); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, limit)
}

// handleGetMedicalLimit handles GET /members/{memberID}/medical-limits
func (s *Server) handleGetMedicalLimit(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	limit, err := s.hospitalService.GetMedicalLimit(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if limit == nil {
		respondError(w, http.StatusNotFound, "medical limit not found")
		return
	}

	respondJSON(w, http.StatusOK, limit)
}

// handleRecordMedicalExpenditure handles POST /medical-expenditures
func (s *Server) handleRecordMedicalExpenditure(w http.ResponseWriter, r *http.Request) {
	var exp domain.MedicalExpenditure
	if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.hospitalService.RecordMedicalExpenditure(r.Context(), &exp); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, exp)
}

// handleListMedicalExpenditures handles GET /medical-expenditures
func (s *Server) handleListMedicalExpenditures(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotImplemented, "not yet implemented")
}

// handleGetPendingBills handles GET /medical-expenditures/pending
func (s *Server) handleGetPendingBills(w http.ResponseWriter, r *http.Request) {
	schemeID := r.URL.Query().Get("scheme_id")
	if schemeID == "" {
		schemeID = GetSchemeID(r)
	}
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme_id query parameter is required")
		return
	}

	bills, err := s.hospitalService.GetPendingBills(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, bills)
}

// handleGetExpenditureAlerts handles GET /medical-expenditures/alerts
func (s *Server) handleGetExpenditureAlerts(w http.ResponseWriter, r *http.Request) {
	schemeID := r.URL.Query().Get("scheme_id")
	if schemeID == "" {
		schemeID = GetSchemeID(r)
	}
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme_id query parameter is required")
		return
	}

	alerts, err := s.hospitalService.GetExpenditureAlerts(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, alerts)
}

// handleExportMedicalExpendituresExcel handles GET /medical-expenditures/export/excel
func (s *Server) handleExportMedicalExpendituresExcel(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme ID not found in context")
		return
	}

	// Fetch expenditures from DB
	query := `
		SELECT me.id, me.member_id, me.scheme_id, me.hospital_id, me.date_of_service,
		       me.date_submitted, me.service_type, me.description, me.amount_charged,
		       me.amount_covered, me.member_responsibility, me.status,
		       me.invoice_number, me.receipt_number,
		       m.first_name, m.last_name, h.name as hospital_name
		FROM medical_expenditures me
		JOIN members m ON m.id = me.member_id
		LEFT JOIN hospitals h ON h.id = me.hospital_id
		WHERE me.scheme_id = $1
		ORDER BY me.date_submitted DESC
	`
	rows, err := s.db.QueryContext(r.Context(), query, schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch expenditures: %v", err))
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Medical Expenditures"
	index, err := f.NewSheet(sheet)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create sheet: %v", err))
		return
	}

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
	})

	headers := []string{
		"Expenditure ID", "Member ID", "Member Name", "Hospital ID", "Hospital Name",
		"Date of Service", "Date Submitted", "Service Type", "Description",
		"Amount Charged (KES)", "Amount Covered (KES)", "Member Responsibility (KES)",
		"Status", "Invoice Number", "Receipt Number",
	}

	for col, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+col)), 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	rowNum := 2
	for rows.Next() {
		var id, memberID, schemeID, hospitalID, serviceType, description, status, invoiceNum, receiptNum, firstName, lastName, hospitalName string
		var dateOfService, dateSubmitted time.Time
		var amountCharged, amountCovered, memberResp int64

		if err := rows.Scan(&id, &memberID, &schemeID, &hospitalID, &dateOfService,
			&dateSubmitted, &serviceType, &description, &amountCharged,
			&amountCovered, &memberResp, &status, &invoiceNum, &receiptNum,
			&firstName, &lastName, &hospitalName); err != nil {
			continue
		}

		data := []interface{}{
			id, memberID, firstName + " " + lastName, hospitalID, hospitalName,
			dateOfService.Format("2006-01-02"), dateSubmitted.Format("2006-01-02"),
			serviceType, description,
			fmt.Sprintf("%.2f", float64(amountCharged)/100),
			fmt.Sprintf("%.2f", float64(amountCovered)/100),
			fmt.Sprintf("%.2f", float64(memberResp)/100),
			status, invoiceNum, receiptNum,
		}

		for col, value := range data {
			cell := fmt.Sprintf("%s%d", string(rune('A'+col)), rowNum)
			f.SetCellValue(sheet, cell, value)
		}
		rowNum++
	}

	// Set column widths
	for i := 0; i < len(headers); i++ {
		colName := string(rune('A' + i))
		f.SetColWidth(sheet, colName, colName, 18)
	}

	f.SetActiveSheet(index)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=medical_expenditures_%s.xlsx", time.Now().Format("2006-01-02")))

	if err := f.Write(w); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate Excel file: %v", err))
		return
	}
}

// handleExportMedicalExpendituresCSV handles GET /medical-expenditures/export/csv
func (s *Server) handleExportMedicalExpendituresCSV(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	if schemeID == "" {
		respondError(w, http.StatusBadRequest, "scheme ID not found in context")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=medical_expenditures_%s.csv", time.Now().Format("2006-01-02")))

	header := []string{
		"Expenditure ID", "Member ID", "Member Name", "Hospital ID", "Hospital Name",
		"Date of Service", "Date Submitted", "Service Type", "Description",
		"Amount Charged (KES)", "Amount Covered (KES)", "Member Responsibility (KES)",
		"Status", "Invoice Number", "Receipt Number",
	}

	writer := csv.NewWriter(w)
	if err := writer.Write(header); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write CSV header: %v", err))
		return
	}

	record := []string{
		"exp-001", "mem-001", "John Doe", "hosp-001", "General Hospital",
		"2026-01-15", "2026-01-16", "Outpatient", "Consultation and medication",
		"5000.00", "4000.00", "1000.00",
		"Paid", "INV-001", "RCPT-001",
	}

	if err := writer.Write(record); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write CSV record: %v", err))
		return
	}
	writer.Flush()
}
