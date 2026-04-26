package member

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"pension-manager/internal/db"
)

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

type PendingMemberInput struct {
	ID                 string `json:"id"`
	MemberNo           string `json:"member_no"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	OtherNames         string `json:"other_names"`
	Gender             string `json:"gender"`
	DateOfBirth        string `json:"date_of_birth"`
	Nationality        string `json:"nationality"`
	IDNumber           string `json:"id_number"`
	KRAPin             string `json:"kra_pin"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	PostalAddress      string `json:"postal_address"`
	PostalCode         string `json:"postal_code"`
	Town               string `json:"town"`
	MaritalStatus      string `json:"marital_status"`
	SpouseName         string `json:"spouse_name"`
	NextOfKin          string `json:"next_of_kin"`
	NextOfKinPhone     string `json:"next_of_kin_phone"`
	BankName           string `json:"bank_name"`
	BankBranch         string `json:"bank_branch"`
	BankAccount        string `json:"bank_account"`
	PayrollNo          string `json:"payroll_no"`
	Designation        string `json:"designation"`
	Department         string `json:"department"`
	SponsorID          string `json:"sponsor_id"`
	DateFirstAppt      string `json:"date_first_appt"`
	DateJoinedScheme   string `json:"date_joined_scheme"`
	ExpectedRetirement string `json:"expected_retirement"`
	BasicSalary        int64  `json:"basic_salary"`
}

type ChangeRequest struct {
	ID              string          `json:"id"`
	SchemeID        string          `json:"scheme_id"`
	MemberID        string          `json:"member_id"`
	RequestType     ChangeType      `json:"request_type"`
	Status          RequestStatus   `json:"status"`
	BeforeData      json.RawMessage `json:"before_data,omitempty"`
	AfterData       json.RawMessage `json:"after_data"`
	RequestedBy     string          `json:"requested_by"`
	ApprovedBy      *string         `json:"approved_by,omitempty"`
	RejectedBy      *string         `json:"rejected_by,omitempty"`
	RejectionReason *string         `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty"`
}

type ChangeType string

const (
	ChangeContactDetails    ChangeType = "contact_details"
	ChangeBeneficiaryAdd    ChangeType = "beneficiary_add"
	ChangeBeneficiaryRemove ChangeType = "beneficiary_remove"
	ChangeBeneficiaryAlloc  ChangeType = "beneficiary_allocation"
	ChangePhoto             ChangeType = "photo"
	ChangeBankDetails       ChangeType = "bank_details"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
)

type MemberCreateRequest struct {
	SchemeID           string    `json:"scheme_id"`
	MemberNo           string    `json:"member_no"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	OtherNames         string    `json:"other_names,omitempty"`
	Gender             string    `json:"gender,omitempty"`
	DateOfBirth        time.Time `json:"date_of_birth"`
	Nationality        string    `json:"nationality,omitempty"`
	IDNumber           string    `json:"id_number,omitempty"`
	KRAPIN             string    `json:"kra_pin,omitempty"`
	Email              string    `json:"email,omitempty"`
	Phone              string    `json:"phone,omitempty"`
	PostalAddress      string    `json:"postal_address,omitempty"`
	PostalCode         string    `json:"postal_code,omitempty"`
	Town               string    `json:"town,omitempty"`
	MaritalStatus      string    `json:"marital_status,omitempty"`
	SpouseName         string    `json:"spouse_name,omitempty"`
	NextOfKin          string    `json:"next_of_kin,omitempty"`
	NextOfKinPhone     string    `json:"next_of_kin_phone,omitempty"`
	BankName           string    `json:"bank_name,omitempty"`
	BankBranch         string    `json:"bank_branch,omitempty"`
	BankAccount        string    `json:"bank_account,omitempty"`
	PayrollNo          string    `json:"payroll_no,omitempty"`
	Designation        string    `json:"designation,omitempty"`
	Department         string    `json:"department,omitempty"`
	SponsorID          string    `json:"sponsor_id,omitempty"`
	DateFirstAppt      time.Time `json:"date_first_appt,omitempty"`
	DateJoinedScheme   time.Time `json:"date_joined_scheme"`
	ExpectedRetirement time.Time `json:"expected_retirement,omitempty"`
	BasicSalary        int64     `json:"basic_salary,omitempty"`
}

type MemberUpdateRequest struct {
	Phone          *string `json:"phone,omitempty"`
	Email          *string `json:"email,omitempty"`
	PostalAddress  *string `json:"postal_address,omitempty"`
	PostalCode     *string `json:"postal_code,omitempty"`
	Town           *string `json:"town,omitempty"`
	BankName       *string `json:"bank_name,omitempty"`
	BankBranch     *string `json:"bank_branch,omitempty"`
	BankAccount    *string `json:"bank_account,omitempty"`
	NextOfKin      *string `json:"next_of_kin,omitempty"`
	NextOfKinPhone *string `json:"next_of_kin_phone,omitempty"`
	Designation    *string `json:"designation,omitempty"`
	Department     *string `json:"department,omitempty"`
	BasicSalary    *int64  `json:"basic_salary,omitempty"`
}

func (s *Service) CreateMember(ctx context.Context, req *MemberCreateRequest, createdBy string) (string, error) {
	var memberID string

	err := s.db.Transactional(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO pending_member_registrations (
				id, scheme_id, member_no, first_name, last_name, other_names, gender,
				date_of_birth, nationality, id_number, kra_pin, email, phone,
				postal_address, postal_code, town, marital_status, spouse_name,
				next_of_kin, next_of_kin_phone, bank_name, bank_branch, bank_account,
				payroll_no, designation, department, sponsor_id, date_first_appt,
				date_joined_scheme, expected_retirement, basic_salary,
				status, requested_by, created_at
			) VALUES (
				uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
				$13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
				$27, $28, $29, 'pending', $30, NOW()
			) RETURNING id
		`

		var dob, dateJoined, dateFirstAppt, expectedRet sql.NullTime
		dob.Time = req.DateOfBirth
		dob.Valid = !req.DateOfBirth.IsZero()
		dateJoined.Time = req.DateJoinedScheme
		dateJoined.Valid = !req.DateJoinedScheme.IsZero()
		if !req.DateFirstAppt.IsZero() {
			dateFirstAppt.Time = req.DateFirstAppt
			dateFirstAppt.Valid = true
		}
		if !req.ExpectedRetirement.IsZero() {
			expectedRet.Time = req.ExpectedRetirement
			expectedRet.Valid = true
		}

		err := tx.QueryRowContext(ctx, query,
			req.SchemeID, req.MemberNo, req.FirstName, req.LastName, req.OtherNames, req.Gender,
			dob, req.Nationality, req.IDNumber, req.KRAPIN, req.Email, req.Phone,
			req.PostalAddress, req.PostalCode, req.Town, req.MaritalStatus, req.SpouseName,
			req.NextOfKin, req.NextOfKinPhone, req.BankName, req.BankBranch, req.BankAccount,
			req.PayrollNo, req.Designation, req.Department, req.SponsorID,
			dateFirstAppt, dateJoined, expectedRet, req.BasicSalary, createdBy,
		).Scan(&memberID)

		if err != nil {
			return fmt.Errorf("create pending registration: %w", err)
		}

		s.logAuditEvent(ctx, tx, req.SchemeID, "member_registration_request", memberID, createdBy,
			map[string]interface{}{
				"action":     "member_registration_requested",
				"member_no":  req.MemberNo,
				"first_name": req.FirstName,
				"last_name":  req.LastName,
			})

		return nil
	})

	if err != nil {
		return "", err
	}
	return memberID, nil
}

func (s *Service) ApproveMemberRegistration(ctx context.Context, pendingID, approvedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var pending map[string]interface{}
		err := tx.QueryRowContext(ctx, `
			SELECT data FROM pending_member_registrations WHERE id = $1 AND status = 'pending'
		`, pendingID).Scan(&pending)
		if err != nil {
			return fmt.Errorf("query pending registration: %w", err)
		}

		memberNo := pending["member_no"].(string)
		schemeID := pending["scheme_id"].(string)

		var exists bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM members WHERE member_no = $1 AND scheme_id = $2)
		`, memberNo, schemeID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("member %s already exists", memberNo)
		}

		memberID, err := s.insertApprovedMember(ctx, tx, pending)
		if err != nil {
			return fmt.Errorf("insert member: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE pending_member_registrations 
			SET status = 'approved', approved_by = $1, reviewed_at = NOW()
			WHERE id = $2
		`, approvedBy, pendingID)
		if err != nil {
			return fmt.Errorf("update pending status: %w", err)
		}

		s.logAuditEvent(ctx, tx, schemeID, "member_registered", memberID, approvedBy,
			map[string]interface{}{
				"action":      "member_registration_approved",
				"member_no":   memberNo,
				"approved_by": approvedBy,
			})

		return nil
	})
}

func (s *Service) insertApprovedMember(ctx context.Context, tx *sql.Tx, pending map[string]interface{}) (string, error) {
	query := `
		INSERT INTO members (
			id, scheme_id, member_no, first_name, last_name, other_names, gender,
			date_of_birth, nationality, id_number, kra_pin, email, phone,
			postal_address, postal_code, town, marital_status, spouse_name,
			next_of_kin, next_of_kin_phone, bank_name, bank_branch, bank_account,
			payroll_no, designation, department, sponsor_id, date_first_appt,
			date_joined_scheme, expected_retirement, basic_salary,
			membership_status, account_balance, created_at, updated_at
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
			$27, $28, $29, 'active', 0, NOW(), NOW()
		) RETURNING id
	`

	var memberID string
	err := tx.QueryRowContext(ctx, query,
		pending["scheme_id"], pending["member_no"], pending["first_name"], pending["last_name"],
		pending["other_names"], pending["gender"], pending["date_of_birth"],
		pending["nationality"], pending["id_number"], pending["kra_pin"],
		pending["email"], pending["phone"], pending["postal_address"],
		pending["postal_code"], pending["town"], pending["marital_status"],
		pending["spouse_name"], pending["next_of_kin"], pending["next_of_kin_phone"],
		pending["bank_name"], pending["bank_branch"], pending["bank_account"],
		pending["payroll_no"], pending["designation"], pending["department"],
		pending["sponsor_id"], pending["date_first_appt"], pending["date_joined_scheme"],
		pending["expected_retirement"], pending["basic_salary"],
	).Scan(&memberID)

	return memberID, err
}

func (s *Service) RejectMemberRegistration(ctx context.Context, pendingID, rejectedBy, reason string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE pending_member_registrations 
		SET status = 'rejected', rejected_by = $1, rejection_reason = $2, reviewed_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`, rejectedBy, reason, pendingID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pending registration not found or already processed")
	}

	var schemeID, memberNo string
	s.db.QueryRowContext(ctx, `
		SELECT scheme_id, member_no FROM pending_member_registrations WHERE id = $1
	`, pendingID).Scan(&schemeID, &memberNo)

	s.logAuditEvent(ctx, nil, schemeID, "member_rejected", pendingID, rejectedBy,
		map[string]interface{}{
			"action":      "member_registration_rejected",
			"member_no":   memberNo,
			"rejected_by": rejectedBy,
			"reason":      reason,
		})

	return nil
}

func (s *Service) CreateUpdateRequest(ctx context.Context, memberID, requestType string, before, after interface{}, requestedBy string) (string, error) {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	var schemeID string
	err := s.db.QueryRowContext(ctx, `SELECT scheme_id FROM members WHERE id = $1`, memberID).Scan(&schemeID)
	if err != nil {
		return "", fmt.Errorf("get member scheme: %w", err)
	}

	var requestID string
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO pending_changes (id, scheme_id, entity_type, entity_id, change_type, requested_by, before_data, after_data, status, created_at)
		VALUES (uuid_generate_v4(), $1, 'member', $2, $3, $4, $5, $6, 'pending', NOW())
		RETURNING id
	`, schemeID, memberID, requestType, requestedBy, beforeJSON, afterJSON).Scan(&requestID)
	if err != nil {
		return "", fmt.Errorf("create change request: %w", err)
	}

	return requestID, nil
}

func (s *Service) ApproveUpdateRequest(ctx context.Context, requestID, approvedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var memberID, changeType string
		var afterData []byte
		var schemeID string

		err := tx.QueryRowContext(ctx, `
			SELECT entity_id, change_type, after_data, scheme_id 
			FROM pending_changes WHERE id = $1 AND status = 'pending'
		`, requestID).Scan(&memberID, &changeType, &afterData, &schemeID)
		if err != nil {
			return fmt.Errorf("query pending change: %w", err)
		}

		var updates []string
		var args []interface{}
		argNum := 1

		var afterMap map[string]interface{}
		json.Unmarshal(afterData, &afterMap)

		for field, value := range afterMap {
			updates = append(updates, fmt.Sprintf("%s = $%d", field, argNum))
			args = append(args, value)
			argNum++
		}
		updates = append(updates, fmt.Sprintf("updated_at = $%d", argNum))
		args = append(args, time.Now())
		argNum++
		args = append(args, memberID)

		query := fmt.Sprintf("UPDATE members SET %s WHERE id = $%d", joinStrings(updates, ", "), argNum)
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("apply update: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE pending_changes SET status = 'approved', approved_by = $1, reviewed_at = NOW()
			WHERE id = $2
		`, approvedBy, requestID)
		if err != nil {
			return fmt.Errorf("update change status: %w", err)
		}

		s.logAuditEvent(ctx, tx, schemeID, "member_update_approved", memberID, approvedBy,
			map[string]interface{}{
				"action":      "member_update_approved",
				"change_type": changeType,
				"approved_by": approvedBy,
			})

		return nil
	})
}

func (s *Service) RejectUpdateRequest(ctx context.Context, requestID, rejectedBy, reason string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE pending_changes SET status = 'rejected', rejected_by = $1, rejection_reason = $2, reviewed_at = NOW()
		WHERE id = $3 AND status = 'pending'
	`, rejectedBy, reason, requestID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("change request not found or already processed")
	}

	return nil
}

func (s *Service) GetPendingRequests(ctx context.Context, schemeID string, limit, offset int) ([]ChangeRequest, int, error) {
	var total int
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pending_changes WHERE scheme_id = $1 AND status = 'pending'
	`, schemeID).Scan(&total)

	query := `
		SELECT id, scheme_id, entity_id, change_type, before_data, after_data, status, 
		       requested_by, approved_by, rejected_by, rejection_reason, created_at, reviewed_at
		FROM pending_changes WHERE scheme_id = $1 AND status = 'pending'
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []ChangeRequest
	for rows.Next() {
		var r ChangeRequest
		err := rows.Scan(&r.ID, &r.SchemeID, &r.MemberID, &r.RequestType, &r.BeforeData, &r.AfterData,
			&r.Status, &r.RequestedBy, &r.ApprovedBy, &r.RejectedBy, &r.RejectionReason,
			&r.CreatedAt, &r.ReviewedAt)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, r)
	}

	return requests, total, nil
}

func (s *Service) logAuditEvent(ctx context.Context, tx *sql.Tx, schemeID, entityType, entityID, actorID string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	if tx != nil {
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	} else {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
