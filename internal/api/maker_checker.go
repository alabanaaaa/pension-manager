package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// PendingChange represents a change awaiting approval
type PendingChange struct {
	ID              string          `json:"id"`
	EntityType      string          `json:"entity_type"`
	EntityID        string          `json:"entity_id"`
	SchemeID        string          `json:"scheme_id"`
	RequestedBy     string          `json:"requested_by"`
	ChangeType      string          `json:"change_type"`
	BeforeData      json.RawMessage `json:"before_data,omitempty"`
	AfterData       json.RawMessage `json:"after_data"`
	Status          string          `json:"status"`
	ReviewedBy      *string         `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty"`
	RejectionReason *string         `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	// Populated fields
	RequesterName string `json:"requester_name,omitempty"`
	ReviewerName  string `json:"reviewer_name,omitempty"`
}

// registerMakerCheckerRoutes registers maker-checker workflow routes
func (s *Server) registerMakerCheckerRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("admin", "pension_officer", "super_admin"))

		r.Route("/api/pending-changes", func(r chi.Router) {
			r.Get("/", s.handleListPendingChanges)
			r.Get("/{id}", s.handleGetPendingChange)
			r.Post("/{id}/approve", s.handleApproveChange)
			r.Post("/{id}/reject", s.handleRejectChange)
			r.Get("/count", s.handlePendingChangeCount)
		})
	})
}

// handleListPendingChanges handles GET /pending-changes
func (s *Server) handleListPendingChanges(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	entityType := r.URL.Query().Get("entity_type")
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	query := `
		SELECT pc.id, pc.entity_type, pc.entity_id, pc.scheme_id, pc.requested_by,
		       pc.change_type, pc.before_data, pc.after_data, pc.status,
		       pc.reviewed_by, pc.reviewed_at, pc.rejection_reason, pc.created_at,
		       COALESCE(u1.first_name || ' ' || u1.last_name, 'Unknown') as requester_name,
		       COALESCE(u2.first_name || ' ' || u2.last_name, '') as reviewer_name
		FROM pending_changes pc
		LEFT JOIN system_users u1 ON u1.id = pc.requested_by
		LEFT JOIN system_users u2 ON u2.id = pc.reviewed_by
		WHERE pc.scheme_id = $1 AND pc.status = $2
	`
	args := []interface{}{schemeID, status}
	argCount := 2

	if entityType != "" {
		argCount++
		query += fmt.Sprintf(" AND pc.entity_type = $%d", argCount)
		args = append(args, entityType)
	}

	query += " ORDER BY pc.created_at DESC"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list pending changes: %v", err))
		return
	}
	defer rows.Close()

	var changes []PendingChange
	for rows.Next() {
		var c PendingChange
		var reviewedBy, rejectionReason sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.EntityType, &c.EntityID, &c.SchemeID, &c.RequestedBy,
			&c.ChangeType, &c.BeforeData, &c.AfterData, &c.Status,
			&reviewedBy, &reviewedAt, &rejectionReason, &c.CreatedAt,
			&c.RequesterName, &c.ReviewerName,
		); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to scan pending change: %v", err))
			return
		}
		if reviewedBy.Valid {
			c.ReviewedBy = &reviewedBy.String
		}
		if reviewedAt.Valid {
			c.ReviewedAt = &reviewedAt.Time
		}
		if rejectionReason.Valid {
			c.RejectionReason = &rejectionReason.String
		}
		changes = append(changes, c)
	}

	respondJSON(w, http.StatusOK, changes)
}

// handleGetPendingChange handles GET /pending-changes/{id}
func (s *Server) handleGetPendingChange(w http.ResponseWriter, r *http.Request) {
	changeID := chi.URLParam(r, "id")
	if changeID == "" {
		respondError(w, http.StatusBadRequest, "change ID is required")
		return
	}

	query := `
		SELECT pc.id, pc.entity_type, pc.entity_id, pc.scheme_id, pc.requested_by,
		       pc.change_type, pc.before_data, pc.after_data, pc.status,
		       pc.reviewed_by, pc.reviewed_at, pc.rejection_reason, pc.created_at,
		       COALESCE(u1.first_name || ' ' || u1.last_name, 'Unknown') as requester_name,
		       COALESCE(u2.first_name || ' ' || u2.last_name, '') as reviewer_name
		FROM pending_changes pc
		LEFT JOIN system_users u1 ON u1.id = pc.requested_by
		LEFT JOIN system_users u2 ON u2.id = pc.reviewed_by
		WHERE pc.id = $1
	`
	var c PendingChange
	var reviewedBy, rejectionReason sql.NullString
	var reviewedAt sql.NullTime
	err := s.db.QueryRowContext(r.Context(), query, changeID).Scan(
		&c.ID, &c.EntityType, &c.EntityID, &c.SchemeID, &c.RequestedBy,
		&c.ChangeType, &c.BeforeData, &c.AfterData, &c.Status,
		&reviewedBy, &reviewedAt, &rejectionReason, &c.CreatedAt,
		&c.RequesterName, &c.ReviewerName,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "pending change not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get pending change: %v", err))
		return
	}

	if reviewedBy.Valid {
		c.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		c.ReviewedAt = &reviewedAt.Time
	}
	if rejectionReason.Valid {
		c.RejectionReason = &rejectionReason.String
	}

	respondJSON(w, http.StatusOK, c)
}

// handleApproveChange handles POST /pending-changes/{id}/approve
func (s *Server) handleApproveChange(w http.ResponseWriter, r *http.Request) {
	changeID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Notes string `json:"notes,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		var entityType, entityID, changeType string
		var afterData []byte
		var status string
		err := tx.QueryRowContext(r.Context(), `
			SELECT entity_type, entity_id, change_type, after_data, status
			FROM pending_changes WHERE id = $1
		`, changeID).Scan(&entityType, &entityID, &changeType, &afterData, &status)
		if err == sql.ErrNoRows {
			return fmt.Errorf("pending change not found")
		}
		if err != nil {
			return fmt.Errorf("get pending change: %w", err)
		}
		if status != "pending" {
			return fmt.Errorf("change is already %s", status)
		}

		// Apply the change based on entity type
		if err := applyPendingChange(r.Context(), tx, entityType, entityID, changeType, afterData); err != nil {
			return fmt.Errorf("apply change: %w", err)
		}

		// Update pending change status
		_, err = tx.ExecContext(r.Context(), `
			UPDATE pending_changes SET status = 'approved', reviewed_by = $1,
			                           reviewed_at = NOW(), rejection_reason = NULL
			WHERE id = $2
		`, userID, changeID)
		if err != nil {
			return fmt.Errorf("update pending change: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM pending_changes WHERE id = $1), $2, $3, 'maker_checker_approved', $4, $5, NOW())
		`, changeID, entityType, entityID, userID, req.Notes)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved", "change_id": changeID})
}

// handleRejectChange handles POST /pending-changes/{id}/reject
func (s *Server) handleRejectChange(w http.ResponseWriter, r *http.Request) {
	changeID := chi.URLParam(r, "id")
	userID := GetUserID(r)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "rejection reason is required")
		return
	}

	err := s.db.Transactional(r.Context(), func(tx *sql.Tx) error {
		var status string
		err := tx.QueryRowContext(r.Context(), `SELECT status FROM pending_changes WHERE id = $1`, changeID).Scan(&status)
		if err == sql.ErrNoRows {
			return fmt.Errorf("pending change not found")
		}
		if err != nil {
			return fmt.Errorf("get pending change: %w", err)
		}
		if status != "pending" {
			return fmt.Errorf("change is already %s", status)
		}

		_, err = tx.ExecContext(r.Context(), `
			UPDATE pending_changes SET status = 'rejected', reviewed_by = $1,
			                           reviewed_at = NOW(), rejection_reason = $2
			WHERE id = $3
		`, userID, req.Reason, changeID)
		if err != nil {
			return fmt.Errorf("reject change: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM pending_changes WHERE id = $1), entity_type, entity_id, 'maker_checker_rejected', $2, $3, NOW())
		`, changeID, userID, req.Reason)
		return err
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected", "change_id": changeID})
}

// handlePendingChangeCount handles GET /pending-changes/count
func (s *Server) handlePendingChangeCount(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var counts struct {
		Total         int `json:"total"`
		Members       int `json:"members"`
		Beneficiaries int `json:"beneficiaries"`
		Claims        int `json:"claims"`
	}

	err := s.db.QueryRowContext(r.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending') as total,
			COUNT(*) FILTER (WHERE status = 'pending' AND entity_type = 'member') as members,
			COUNT(*) FILTER (WHERE status = 'pending' AND entity_type = 'beneficiary') as beneficiaries,
			COUNT(*) FILTER (WHERE status = 'pending' AND entity_type = 'claim') as claims
		FROM pending_changes WHERE scheme_id = $1
	`, schemeID).Scan(&counts.Total, &counts.Members, &counts.Beneficiaries, &counts.Claims)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get counts: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, counts)
}

// applyPendingChange applies a pending change to the actual entity
func applyPendingChange(ctx context.Context, tx *sql.Tx, entityType, entityID, changeType string, afterData []byte) error {
	switch entityType {
	case "member":
		return applyMemberChange(ctx, tx, entityID, changeType, afterData)
	case "beneficiary":
		return applyBeneficiaryChange(ctx, tx, entityID, changeType, afterData)
	case "claim":
		return applyClaimChange(ctx, tx, entityID, changeType, afterData)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func applyMemberChange(ctx context.Context, tx *sql.Tx, entityID, changeType string, afterData []byte) error {
	var member struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		OtherNames  string `json:"other_names"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		IDNumber    string `json:"id_number"`
		KRAPIN      string `json:"kra_pin"`
		BankName    string `json:"bank_name"`
		BankBranch  string `json:"bank_branch"`
		BankAccount string `json:"bank_account"`
		Department  string `json:"department"`
		Designation string `json:"designation"`
		BasicSalary int64  `json:"basic_salary"`
	}
	if err := json.Unmarshal(afterData, &member); err != nil {
		return fmt.Errorf("parse member data: %w", err)
	}

	query := `
		UPDATE members SET first_name = $1, last_name = $2, other_names = $3,
		                   phone = $4, email = $5, id_number = $6, kra_pin = $7,
		                   bank_name = $8, bank_branch = $9, bank_account = $10,
		                   department = $11, designation = $12, basic_salary = $13,
		                   updated_at = NOW()
		WHERE id = $14
	`
	_, err := tx.ExecContext(ctx, query,
		member.FirstName, member.LastName, member.OtherNames, member.Phone, member.Email,
		member.IDNumber, member.KRAPIN, member.BankName, member.BankBranch, member.BankAccount,
		member.Department, member.Designation, member.BasicSalary, entityID,
	)
	return err
}

func applyBeneficiaryChange(ctx context.Context, tx *sql.Tx, entityID, changeType string, afterData []byte) error {
	var beneficiary struct {
		Name          string  `json:"name"`
		Relationship  string  `json:"relationship"`
		IDNumber      string  `json:"id_number"`
		Phone         string  `json:"phone"`
		AllocationPct float64 `json:"allocation_pct"`
	}
	if err := json.Unmarshal(afterData, &beneficiary); err != nil {
		return fmt.Errorf("parse beneficiary data: %w", err)
	}

	query := `
		UPDATE beneficiaries SET name = $1, relationship = $2, id_number = $3,
		                         phone = $4, allocation_pct = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := tx.ExecContext(ctx, query,
		beneficiary.Name, beneficiary.Relationship, beneficiary.IDNumber,
		beneficiary.Phone, beneficiary.AllocationPct, entityID,
	)
	return err
}

func applyClaimChange(ctx context.Context, tx *sql.Tx, entityID, changeType string, afterData []byte) error {
	var claim struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(afterData, &claim); err != nil {
		return fmt.Errorf("parse claim data: %w", err)
	}

	query := `UPDATE claims SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, claim.Status, entityID)
	return err
}

// CreatePendingChange creates a new pending change entry (helper for other handlers)
func CreatePendingChange(ctx context.Context, tx *sql.Tx, schemeID, entityType, entityID, changeType, requestedBy string, beforeData, afterData interface{}) error {
	var beforeJSON, afterJSON []byte
	var err error

	if beforeData != nil {
		beforeJSON, err = json.Marshal(beforeData)
		if err != nil {
			return fmt.Errorf("marshal before data: %w", err)
		}
	}
	if afterData != nil {
		afterJSON, err = json.Marshal(afterData)
		if err != nil {
			return fmt.Errorf("marshal after data: %w", err)
		}
	}

	query := `
		INSERT INTO pending_changes (id, scheme_id, entity_type, entity_id, change_type, requested_by,
		                             before_data, after_data, status, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, 'pending', NOW())
	`
	_, err = tx.ExecContext(ctx, query, schemeID, entityType, entityID, changeType, requestedBy, beforeJSON, afterJSON)
	return err
}
