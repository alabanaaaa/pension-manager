package api

import (
	"encoding/json"
	"net/http"

	"pension-manager/internal/tax"

	"github.com/go-chi/chi/v5"
)

// registerTaxRoutes registers tax computation routes
func (s *Server) registerTaxRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Route("/api/tax", func(r chi.Router) {
			r.Post("/compute", s.handleComputeTax)
			r.Post("/compute/monthly", s.handleComputeMonthlyTax)
			r.Post("/compute/withdrawal", s.handleComputeWithdrawalTax)
			r.Post("/compute/multi-scheme", s.handleComputeMultiSchemeTax)
			r.Get("/brackets", s.handleGetTaxBrackets)
			r.Get("/reliefs", s.handleGetTaxReliefs)
			r.Get("/member/{memberID}", s.handleGetMemberTaxStatus)
		})
	})
}

// handleComputeTax handles POST /tax/compute
func (s *Server) handleComputeTax(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AnnualIncome    int64    `json:"annual_income"`
		ReliefTypes     []string `json:"relief_types,omitempty"`
		IsTaxExempt     bool     `json:"is_tax_exempt"`
		ExemptionReason string   `json:"exemption_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var reliefs []tax.Relief
	allReliefs := tax.GetAvailableReliefs()
	for _, rt := range req.ReliefTypes {
		for _, r := range allReliefs {
			if r.Name == rt {
				reliefs = append(reliefs, r)
			}
		}
	}

	result := tax.ComputeTax(req.AnnualIncome, reliefs, req.IsTaxExempt, req.ExemptionReason)
	respondJSON(w, http.StatusOK, result)
}

// handleComputeMonthlyTax handles POST /tax/compute/monthly
func (s *Server) handleComputeMonthlyTax(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MonthlyIncome   int64    `json:"monthly_income"`
		ReliefTypes     []string `json:"relief_types,omitempty"`
		IsTaxExempt     bool     `json:"is_tax_exempt"`
		ExemptionReason string   `json:"exemption_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var reliefs []tax.Relief
	allReliefs := tax.GetAvailableReliefs()
	for _, rt := range req.ReliefTypes {
		for _, r := range allReliefs {
			if r.Name == rt {
				reliefs = append(reliefs, r)
			}
		}
	}

	result := tax.ComputeTaxMonthly(req.MonthlyIncome, reliefs, req.IsTaxExempt, req.ExemptionReason)
	respondJSON(w, http.StatusOK, result)
}

// handleComputeWithdrawalTax handles POST /tax/compute/withdrawal
func (s *Server) handleComputeWithdrawalTax(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WithdrawalAmount int64  `json:"withdrawal_amount"`
		YearsOfService   int    `json:"years_of_service"`
		IsTaxExempt      bool   `json:"is_tax_exempt"`
		ExemptionReason  string `json:"exemption_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := tax.CalculateTaxOnWithdrawal(req.WithdrawalAmount, req.YearsOfService, req.IsTaxExempt, req.ExemptionReason)
	respondJSON(w, http.StatusOK, result)
}

// handleComputeMultiSchemeTax handles POST /tax/compute/multi-scheme
func (s *Server) handleComputeMultiSchemeTax(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MemberID        string          `json:"member_id"`
		Schemes         []tax.SchemeTax `json:"schemes"`
		IsTaxExempt     bool            `json:"is_tax_exempt"`
		ExemptionReason string          `json:"exemption_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := tax.ComputeMultiSchemeTax(req.MemberID, req.Schemes, req.IsTaxExempt, req.ExemptionReason)
	respondJSON(w, http.StatusOK, result)
}

// handleGetTaxBrackets handles GET /tax/brackets
func (s *Server) handleGetTaxBrackets(w http.ResponseWriter, r *http.Request) {
	brackets := tax.GetTaxBrackets()
	respondJSON(w, http.StatusOK, brackets)
}

// handleGetTaxReliefs handles GET /tax/reliefs
func (s *Server) handleGetTaxReliefs(w http.ResponseWriter, r *http.Request) {
	reliefs := tax.GetAvailableReliefs()
	respondJSON(w, http.StatusOK, reliefs)
}

// handleGetMemberTaxStatus handles GET /tax/member/{memberID}
func (s *Server) handleGetMemberTaxStatus(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		respondError(w, http.StatusBadRequest, "member ID is required")
		return
	}

	var isTaxExempt bool
	var taxExemptReason, taxExemptAttachment string
	var taxExemptCutoffDate interface{}
	var basicSalary int64
	var age int

	err := s.db.QueryRowContext(r.Context(), `
		SELECT basic_salary,
		       CASE WHEN tax_exempt_reason IS NOT NULL AND tax_exempt_cutoff_date > NOW() THEN true
		            WHEN DATE_PART('year', AGE(NOW(), date_of_birth)) >= 65 THEN true
		            ELSE false END as is_tax_exempt,
		       COALESCE(tax_exempt_reason, ''),
		       COALESCE(tax_exempt_attachment, ''),
		       tax_exempt_cutoff_date,
		       DATE_PART('year', AGE(NOW(), date_of_birth))::int as age
		FROM members WHERE id = $1
	`, memberID).Scan(&basicSalary, &isTaxExempt, &taxExemptReason, &taxExemptAttachment, &taxExemptCutoffDate, &age)
	if err != nil {
		respondError(w, http.StatusNotFound, "member not found")
		return
	}

	// Compute monthly tax
	monthlyTax := tax.ComputeTaxMonthly(basicSalary, []tax.Relief{tax.PersonalRelief}, isTaxExempt, taxExemptReason)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"member_id":              memberID,
		"age":                    age,
		"basic_salary":           basicSalary,
		"is_tax_exempt":          isTaxExempt,
		"tax_exempt_reason":      taxExemptReason,
		"tax_exempt_attachment":  taxExemptAttachment,
		"tax_exempt_cutoff_date": taxExemptCutoffDate,
		"monthly_tax":            monthlyTax,
		"annual_tax":             monthlyTax.NetTax * 12,
	})
}
