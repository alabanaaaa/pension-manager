package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// registerBulkRoutes registers bulk processing routes
func (s *Server) registerBulkRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Route("/api/bulk", func(r chi.Router) {
			r.Post("/import/members", s.handleImportMembers)
			r.Post("/validate", s.handleValidateBulkUpdate)
			r.Post("/process/retirements", s.handleProcessRetirements)
			r.Post("/process/early-leavers", s.handleProcessEarlyLeavers)
			r.Post("/process/annual-posting", s.handleAnnualPosting)
			r.Get("/statements/batch", s.handleBatchStatements)
			r.Get("/statements/batch/export", s.handleBatchStatementsExport)
		})
	})
}

// registerReportRoutes registers contribution report routes
func (s *Server) registerReportRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer", "auditor"))

		r.Route("/api/reports/contributions", func(r chi.Router) {
			r.Get("/breakdown", s.handleContributionBreakdown)
			r.Get("/ytd", s.handleYTDContributions)
			r.Get("/cumulative", s.handleCumulativeContributions)
			r.Get("/registered-vs-unregistered", s.handleRegisteredVsUnregistered)
			r.Get("/trends", s.handleContributionTrends)
			r.Get("/avc-summary", s.handleAVCSummary)
			r.Get("/export", s.handleExportContributionReport)
		})
	})
}

// Bulk Processing Handlers

func (s *Server) handleImportMembers(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "CSV file is required")
		return
	}
	defer file.Close()

	result, err := s.bulkService.ImportMembersCSV(r.Context(), schemeID, file, userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleValidateBulkUpdate(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var req struct {
		MemberNos []string         `json:"member_nos"`
		Salaries  map[string]int64 `json:"salaries"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validation, err := s.bulkService.ValidateBulkUpdate(r.Context(), schemeID, req.MemberNos, req.Salaries)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, validation)
}

func (s *Server) handleProcessRetirements(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var req struct {
		RetirementType string `json:"retirement_type"`
	}
	decodeJSON(r, &req)
	if req.RetirementType == "" {
		req.RetirementType = "normal"
	}

	result, err := s.bulkService.ProcessRetirements(r.Context(), schemeID, req.RetirementType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleProcessEarlyLeavers(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	result, err := s.bulkService.ProcessEarlyLeavers(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleAnnualPosting(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid year")
		return
	}

	result, err := s.bulkService.AnnualPosting(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchStatements(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	department := r.URL.Query().Get("department")
	status := r.URL.Query().Get("status")

	data, err := s.bulkService.GetBatchStatementData(r.Context(), schemeID, department, status, time.Time{}, time.Time{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (s *Server) handleBatchStatementsExport(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	department := r.URL.Query().Get("department")
	status := r.URL.Query().Get("status")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	data, err := s.bulkService.GetBatchStatementData(r.Context(), schemeID, department, status, time.Time{}, time.Time{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=batch_statements.csv")

		writer := csv.NewWriter(w)
		writer.Write([]string{"Member No", "Full Name", "Department", "Balance", "Email"})
		for _, d := range data {
			writer.Write([]string{d.MemberNo, d.FullName, d.Department, fmt.Sprintf("%d", d.Balance), d.Email})
		}
		writer.Flush()
	} else {
		respondJSON(w, http.StatusOK, data)
	}
}

// Contribution Report Handlers

func (s *Server) handleContributionBreakdown(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	breakdown, err := s.reportsService.GetEmployeeEmployerBreakdown(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, breakdown)
}

func (s *Server) handleYTDContributions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	ytd, err := s.reportsService.GetYTDContributions(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ytd)
}

func (s *Server) handleCumulativeContributions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	cumulative, err := s.reportsService.GetCumulativeContributions(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, cumulative)
}

func (s *Server) handleRegisteredVsUnregistered(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	results, err := s.reportsService.GetRegisteredVsUnregistered(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, results)
}

func (s *Server) handleContributionTrends(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	trends, err := s.reportsService.GetContributionTrends(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, trends)
}

func (s *Server) handleAVCSummary(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	summary, err := s.reportsService.GetAVCSummary(r.Context(), schemeID, year)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (s *Server) handleExportContributionReport(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	reportType := r.URL.Query().Get("type")
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, _ := strconv.Atoi(yearStr)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=contribution_report_%s_%s.csv", reportType, yearStr))

	writer := csv.NewWriter(w)

	switch reportType {
	case "breakdown":
		data, _ := s.reportsService.GetEmployeeEmployerBreakdown(r.Context(), schemeID, year)
		writer.Write([]string{"Period", "Employees", "Employee Total", "Employer Total", "AVC Total", "Grand Total"})
		for _, d := range data {
			writer.Write([]string{d.Period, strconv.Itoa(d.EmployeeCount), fmt.Sprintf("%d", d.EmployeeTotal), fmt.Sprintf("%d", d.EmployerTotal), fmt.Sprintf("%d", d.AVCTotal), fmt.Sprintf("%d", d.GrandTotal)})
		}
	case "ytd":
		data, _ := s.reportsService.GetYTDContributions(r.Context(), schemeID, year)
		writer.Write([]string{"Member No", "Full Name", "Employee YTD", "Employer YTD", "AVC YTD", "Total YTD"})
		for _, d := range data {
			writer.Write([]string{d.MemberNo, d.FullName, fmt.Sprintf("%d", d.EmployeeYTD), fmt.Sprintf("%d", d.EmployerYTD), fmt.Sprintf("%d", d.AVCYTD), fmt.Sprintf("%d", d.TotalYTD)})
		}
	case "cumulative":
		data, _ := s.reportsService.GetCumulativeContributions(r.Context(), schemeID)
		writer.Write([]string{"Member No", "Full Name", "Employee Cumulative", "Employer Cumulative", "AVC Cumulative", "Total Cumulative"})
		for _, d := range data {
			writer.Write([]string{d.MemberNo, d.FullName, fmt.Sprintf("%d", d.EmployeeCumulative), fmt.Sprintf("%d", d.EmployerCumulative), fmt.Sprintf("%d", d.AVCCumulative), fmt.Sprintf("%d", d.TotalCumulative)})
		}
	case "trends":
		data, _ := s.reportsService.GetContributionTrends(r.Context(), schemeID, year)
		writer.Write([]string{"Month", "Employee Amount", "Employer Amount", "AVC Amount", "Total Amount", "Member Count"})
		for _, d := range data {
			writer.Write([]string{d.Month, fmt.Sprintf("%d", d.EmployeeAmount), fmt.Sprintf("%d", d.EmployerAmount), fmt.Sprintf("%d", d.AVCAmount), fmt.Sprintf("%d", d.TotalAmount), strconv.Itoa(d.MemberCount)})
		}
	default:
		respondError(w, http.StatusBadRequest, "invalid report type. Use: breakdown, ytd, cumulative, trends")
		return
	}

	writer.Flush()
}
