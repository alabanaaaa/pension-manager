package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerClaimsRoutesV2(r chi.Router) {
	r.Get("/api/claims/pending", s.listPendingClaims)
	r.Post("/api/claims/pending", s.submitPendingClaim)
	r.Get("/api/claims/pending/{id}", s.getPendingClaim)
	r.Post("/api/claims/pending/{id}/approve", s.approvePendingClaim)
	r.Post("/api/claims/pending/{id}/reject", s.rejectPendingClaim)
}

func (s *Server) listPendingClaims(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	query := `
		SELECT pc.id, pc.member_id, pc.claim_type, pc.claim_form_no, pc.date_of_claim,
		       pc.date_of_leaving, pc.leaving_reason, pc.estimated_amount, pc.status,
		       pc.submitted_by, pc.created_at, pc.reviewed_at,
		       m.member_no, m.first_name, m.last_name, m.account_balance
		FROM pending_claims pc
		JOIN members m ON pc.member_id = m.id
		WHERE pc.scheme_id = $1 AND pc.status = $2
		ORDER BY pc.created_at DESC
	`

	rows, err := s.db.QueryContext(r.Context(), query, schemeID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var claims []map[string]interface{}
	for rows.Next() {
		var id, memberID, claimType, claimFormNo, dateOfClaim, leavingReason, status, submittedBy string
		var dateOfLeaving, reviewedAt *time.Time
		var estimatedAmt int64
		var memberNo, firstName, lastName string
		var accountBalance int64
		var createdAt time.Time
		if err := rows.Scan(&id, &memberID, &claimType, &claimFormNo, &dateOfClaim, &dateOfLeaving,
			&leavingReason, &estimatedAmt, &status, &submittedBy, &createdAt, &reviewedAt,
			&memberNo, &firstName, &lastName, &accountBalance); err != nil {
			continue
		}
		c := map[string]interface{}{
			"id": id, "member_id": memberID, "claim_type": claimType,
			"claim_form_no": claimFormNo, "date_of_claim": dateOfClaim,
			"leaving_reason": leavingReason, "estimated_amount": estimatedAmt,
			"status": status, "submitted_by": submittedBy,
			"member_no": memberNo, "member_name": firstName + " " + lastName,
			"account_balance": accountBalance,
		}
		if dateOfLeaving != nil {
			c["date_of_leaving"] = *dateOfLeaving
		}
		if reviewedAt != nil {
			c["reviewed_at"] = *reviewedAt
		}
		claims = append(claims, c)
	}
	if claims == nil {
		claims = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, claims)
}

func (s *Server) submitPendingClaim(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		MemberID       string          `json:"member_id"`
		ClaimType      string          `json:"claim_type"`
		ClaimFormNo    string          `json:"claim_form_no"`
		DateOfClaim    string          `json:"date_of_claim"`
		DateOfLeaving  string          `json:"date_of_leaving"`
		LeavingReason  string          `json:"leaving_reason"`
		EstimatedAmt   int64           `json:"estimated_amount"`
		SupportingDocs json.RawMessage `json:"supporting_docs"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MemberID == "" || req.ClaimType == "" {
		respondError(w, http.StatusBadRequest, "member_id and claim_type are required")
		return
	}

	var claimID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO pending_claims (
			member_id, scheme_id, claim_type, claim_form_no, date_of_claim,
			date_of_leaving, leaving_reason, estimated_amount, supporting_docs,
			status, submitted_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',$10)
		RETURNING id
	`, req.MemberID, schemeID, req.ClaimType, req.ClaimFormNo, req.DateOfClaim,
		req.DateOfLeaving, req.LeavingReason, req.EstimatedAmt, req.SupportingDocs, userID).Scan(&claimID)

	if err != nil {
		slog.Error("submit pending claim failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to submit claim")
		return
	}

	respondCreated(w, map[string]interface{}{"id": claimID, "status": "pending"})
}

func (s *Server) getPendingClaim(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	id := chi.URLParam(r, "id")

	var claim struct {
		ID              string
		MemberID        string
		ClaimType       string
		ClaimFormNo     string
		DateOfClaim     string
		DateOfLeaving   string
		LeavingReason   string
		EstimatedAmt    int64
		SupportingDocs  string
		Status          string
		SubmittedBy     string
		RejectionReason string
		CreatedAt       time.Time
		ReviewedAt      *time.Time
	}

	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, member_id, claim_type, claim_form_no, date_of_claim,
		       date_of_leaving, leaving_reason, COALESCE(estimated_amount, 0),
		       COALESCE(supporting_docs::text, ''), status, submitted_by,
		       COALESCE(rejection_reason, ''), created_at, reviewed_at
		FROM pending_claims
		WHERE id = $1 AND scheme_id = $2
	`, id, schemeID).Scan(&claim.ID, &claim.MemberID, &claim.ClaimType, &claim.ClaimFormNo,
		&claim.DateOfClaim, &claim.DateOfLeaving, &claim.LeavingReason, &claim.EstimatedAmt,
		&claim.SupportingDocs, &claim.Status, &claim.SubmittedBy, &claim.RejectionReason,
		&claim.CreatedAt, &claim.ReviewedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "pending claim not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": claim.ID, "member_id": claim.MemberID, "claim_type": claim.ClaimType,
		"claim_form_no": claim.ClaimFormNo, "date_of_claim": claim.DateOfClaim,
		"date_of_leaving": claim.DateOfLeaving, "leaving_reason": claim.LeavingReason,
		"estimated_amount": claim.EstimatedAmt, "supporting_docs": claim.SupportingDocs,
		"status": claim.Status, "submitted_by": claim.SubmittedBy,
		"rejection_reason": claim.RejectionReason, "created_at": claim.CreatedAt,
		"reviewed_at": claim.ReviewedAt,
	})
}

func (s *Server) approvePendingClaim(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)
	id := chi.URLParam(r, "id")

	var pending struct {
		MemberID      string
		ClaimType     string
		ClaimFormNo   string
		DateOfClaim   string
		DateOfLeaving string
		LeavingReason string
		EstimatedAmt  int64
	}

	err := s.db.QueryRowContext(r.Context(), `
		SELECT member_id, claim_type, claim_form_no, date_of_claim,
		       date_of_leaving, leaving_reason, COALESCE(estimated_amount, 0)
		FROM pending_claims
		WHERE id = $1 AND scheme_id = $2 AND status = 'pending'
	`, id, schemeID).Scan(&pending.MemberID, &pending.ClaimType, &pending.ClaimFormNo,
		&pending.DateOfClaim, &pending.DateOfLeaving, &pending.LeavingReason, &pending.EstimatedAmt)

	if err != nil {
		respondError(w, http.StatusNotFound, "pending claim not found or already processed")
		return
	}

	var claimID string
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO claims (
			member_id, scheme_id, claim_type, claim_form_no, date_of_claim,
			date_of_leaving, leaving_reason, amount, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'submitted')
		RETURNING id
	`, pending.MemberID, schemeID, pending.ClaimType, pending.ClaimFormNo,
		pending.DateOfClaim, pending.DateOfLeaving, pending.LeavingReason,
		pending.EstimatedAmt).Scan(&claimID)

	if err != nil {
		slog.Error("approve pending claim failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create claim")
		return
	}

	_, err = s.db.ExecContext(r.Context(), `
		UPDATE pending_claims SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2
	`, userID, id)

	recordEvent(s.db, r.Context(), schemeID, "claim", claimID, "claim_submitted", map[string]interface{}{
		"claim_type": pending.ClaimType, "member_id": pending.MemberID,
	}, userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{"status": "approved", "claim_id": claimID})
}

func (s *Server) rejectPendingClaim(w http.ResponseWriter, r *http.Request) {
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
		UPDATE pending_claims SET status = 'rejected', rejection_reason = $1,
			reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3 AND scheme_id = $4 AND status = 'pending'
	`, req.Reason, userID, id, schemeID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to reject claim")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "pending claim not found or already processed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
