package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerReconciliationRoutes(r chi.Router) {
	r.Get("/api/reconciliation/edi", s.listEDIFiles)
	r.Post("/api/reconciliation/edi", s.uploadEDIFile)
	r.Get("/api/reconciliation/edi/{id}", s.getEDIDetails)
	r.Post("/api/reconciliation/edi/{id}/reconcile", s.reconcileEDI)

	r.Get("/api/reconciliation/unregistered", s.listUnregisteredContributions)
	r.Post("/api/reconciliation/unregistered", s.trackUnregisteredContribution)
	r.Put("/api/reconciliation/unregistered/{id}", s.updateUnregisteredContribution)
}

func (s *Server) listEDIFiles(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	sponsorID := r.URL.Query().Get("sponsor_id")
	status := r.URL.Query().Get("status")

	query := `
		SELECT ef.id, ef.sponsor_id, s.name as sponsor_name, ef.file_name, ef.record_count,
		       ef.total_employee_contributions, ef.total_employer_contributions, ef.status,
		       ef.processed_at, ef.error_message, ef.created_at
		FROM edi_files ef
		JOIN sponsors s ON ef.sponsor_id = s.id
		WHERE s.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if sponsorID != "" {
		argCount++
		query += fmt.Sprintf(" AND ef.sponsor_id = $%d", argCount)
		args = append(args, sponsorID)
	}
	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND ef.status = $%d", argCount)
		args = append(args, status)
	}
	query += " ORDER BY ef.created_at DESC"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var files []map[string]interface{}
	for rows.Next() {
		var id, sponsorID, sponsorName, fileName, status, errorMsg string
		var recordCount int
		var empContribs, erContribs int64
		var processedAt *time.Time
		var createdAt time.Time
		if err := rows.Scan(&id, &sponsorID, &sponsorName, &fileName, &recordCount, &empContribs, &erContribs, &status, &processedAt, &errorMsg, &createdAt); err != nil {
			continue
		}
		f := map[string]interface{}{
			"id": id, "sponsor_id": sponsorID, "sponsor_name": sponsorName, "file_name": fileName,
			"record_count": recordCount, "total_employee_contributions": empContribs,
			"total_employer_contributions": erContribs, "status": status,
			"created_at": createdAt,
		}
		if errorMsg != "" {
			f["error_message"] = errorMsg
		}
		if processedAt != nil {
			f["processed_at"] = *processedAt
		}
		files = append(files, f)
	}
	if files == nil {
		files = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, files)
}

func (s *Server) uploadEDIFile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)

	var req struct {
		SponsorID   string `json:"sponsor_id"`
		FileName    string `json:"file_name"`
		FileHash    string `json:"file_hash"`
		RecordCount int    `json:"record_count"`
		EmpContribs int64  `json:"employee_contributions"`
		ErContribs  int64  `json:"employer_contributions"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var fileID string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO edi_files (sponsor_id, file_name, file_hash, record_count,
			total_employee_contributions, total_employer_contributions, status, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, 'processing', $7)
		RETURNING id
	`, req.SponsorID, req.FileName, req.FileHash, req.RecordCount, req.EmpContribs, req.ErContribs, userID).Scan(&fileID)

	if err != nil {
		slog.Error("upload EDI file failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to upload EDI file")
		return
	}

	respondCreated(w, map[string]interface{}{"id": fileID, "status": "processing"})
}

func (s *Server) getEDIDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var file struct {
		ID                      string
		SponsorID               string
		FileName                string
		FileHash                string
		RecordCount             int
		EmpContribs, ErContribs int64
		Status                  string
		ProcessedAt             *time.Time
		ErrorMessage            string
		CreatedAt               time.Time
	}

	err := s.db.QueryRowContext(r.Context(), `
		SELECT id, sponsor_id, file_name, file_hash, record_count,
		       total_employee_contributions, total_employer_contributions, status,
		       processed_at, COALESCE(error_message, ''), created_at
		FROM edi_files WHERE id = $1
	`, id).Scan(&file.ID, &file.SponsorID, &file.FileName, &file.FileHash, &file.RecordCount,
		&file.EmpContribs, &file.ErContribs, &file.Status, &file.ProcessedAt,
		&file.ErrorMessage, &file.CreatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "EDI file not found")
		return
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, record_type, member_id, matched, match_confidence, discrepancy_notes
		FROM edi_records WHERE edi_file_id = $1
	`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var recID, recordType, memberID, discrepancyNotes string
		var matched bool
		var confidence float64
		if err := rows.Scan(&recID, &recordType, &memberID, &matched, &confidence, &discrepancyNotes); err != nil {
			continue
		}
		records = append(records, map[string]interface{}{
			"id": recID, "record_type": recordType, "member_id": memberID,
			"matched": matched, "match_confidence": confidence, "discrepancy_notes": discrepancyNotes,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"file": map[string]interface{}{
			"id": file.ID, "sponsor_id": file.SponsorID, "file_name": file.FileName,
			"file_hash": file.FileHash, "record_count": file.RecordCount,
			"total_employee_contributions": file.EmpContribs,
			"total_employer_contributions": file.ErContribs, "status": file.Status,
			"processed_at": file.ProcessedAt, "error_message": file.ErrorMessage,
			"created_at": file.CreatedAt,
		},
		"records": records,
	})
}

func (s *Server) reconcileEDI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := s.db.ExecContext(r.Context(), `
		UPDATE edi_files SET status = 'reconciled', processed_at = NOW()
		WHERE id = $1 AND status IN ('processing', 'partial_mismatch', 'mismatch')
	`, id)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to reconcile EDI file")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "EDI file not found or already reconciled")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "reconciled"})
}

func (s *Server) listUnregisteredContributions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	status := r.URL.Query().Get("status")

	query := `
		SELECT id, sponsor_id, sponsor_name, member_name, member_id_no, period,
		       employee_amount, employer_amount, total_amount, payment_ref, payment_date,
		       tracking_status, first_contact_date, last_contact_date, notes, created_at
		FROM unregistered_contributions
		WHERE scheme_id = $1
	`
	args := []interface{}{schemeID}

	if status != "" {
		query += " AND tracking_status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	var contributions []map[string]interface{}
	for rows.Next() {
		var id, sponsorID, sponsorName, memberName, memberIDNo, paymentRef, status, notes string
		var empAmt, erAmt, total int64
		var period, paymentDate, firstContact, lastContact *time.Time
		var createdAt time.Time
		if err := rows.Scan(&id, &sponsorID, &sponsorName, &memberName, &memberIDNo, &period,
			&empAmt, &erAmt, &total, &paymentRef, &paymentDate, &status, &firstContact,
			&lastContact, &notes, &createdAt); err != nil {
			continue
		}
		c := map[string]interface{}{
			"id": id, "sponsor_id": sponsorID, "sponsor_name": sponsorName,
			"member_name": memberName, "member_id_no": memberIDNo, "employee_amount": empAmt,
			"employer_amount": erAmt, "total_amount": total, "payment_ref": paymentRef,
			"tracking_status": status, "notes": notes, "created_at": createdAt,
		}
		if period != nil {
			c["period"] = *period
		}
		if paymentDate != nil {
			c["payment_date"] = *paymentDate
		}
		if firstContact != nil {
			c["first_contact_date"] = *firstContact
		}
		if lastContact != nil {
			c["last_contact_date"] = *lastContact
		}
		contributions = append(contributions, c)
	}
	if contributions == nil {
		contributions = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, contributions)
}

func (s *Server) trackUnregisteredContribution(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var req struct {
		SponsorID   string `json:"sponsor_id"`
		SponsorName string `json:"sponsor_name"`
		MemberName  string `json:"member_name"`
		MemberIDNo  string `json:"member_id_no"`
		Period      string `json:"period"`
		EmpAmount   int64  `json:"employee_amount"`
		ErAmount    int64  `json:"employer_amount"`
		PaymentRef  string `json:"payment_ref"`
		PaymentDate string `json:"payment_date"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	period, _ := time.Parse("2006-01-02", req.Period)
	paymentDate, _ := time.Parse("2006-01-02", req.PaymentDate)
	total := req.EmpAmount + req.ErAmount

	var id string
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO unregistered_contributions (
			scheme_id, sponsor_id, sponsor_name, member_name, member_id_no, period,
			employee_amount, employer_amount, total_amount, payment_ref, payment_date,
			tracking_status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'pending')
		RETURNING id
	`, schemeID, req.SponsorID, req.SponsorName, req.MemberName, req.MemberIDNo,
		period, req.EmpAmount, req.ErAmount, total, req.PaymentRef, paymentDate).Scan(&id)

	if err != nil {
		slog.Error("track unregistered contribution failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to track contribution")
		return
	}

	respondCreated(w, map[string]interface{}{"id": id, "status": "pending"})
}

func (s *Server) updateUnregisteredContribution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		TrackingStatus string `json:"tracking_status"`
		Notes          string `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var updateAt string
	if req.TrackingStatus == "contacted" {
		updateAt = "first_contact_date"
	} else if req.TrackingStatus != "" {
		updateAt = "last_contact_date"
	}

	query := "UPDATE unregistered_contributions SET tracking_status = $1"
	args := []interface{}{req.TrackingStatus}
	if req.Notes != "" {
		query += ", notes = $2"
		args = append(args, req.Notes)
	}
	if updateAt != "" {
		query += fmt.Sprintf(", %s = NOW()", updateAt)
	}
	query += " WHERE id = $" + fmt.Sprint(len(args)+1)
	args = append(args, id)

	_, err := s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
