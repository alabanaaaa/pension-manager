package bulk

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"pension-manager/internal/db"
)

// ImportResult holds the result of a bulk import operation
type ImportResult struct {
	TotalRows int             `json:"total_rows"`
	Success   int             `json:"success"`
	Failed    int             `json:"failed"`
	Errors    []ImportError   `json:"errors,omitempty"`
	Warnings  []ImportWarning `json:"warnings,omitempty"`
}

// ImportError holds details about a failed row
type ImportError struct {
	Row    int    `json:"row"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// ImportWarning holds details about a warning
type ImportWarning struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// BulkValidation holds validation results before bulk update
type BulkValidation struct {
	IsValid        bool           `json:"is_valid"`
	NewMembers     []string       `json:"new_members,omitempty"`
	RemovedMembers []string       `json:"removed_members,omitempty"`
	SalaryChanges  []SalaryChange `json:"salary_changes,omitempty"`
	Errors         []string       `json:"errors,omitempty"`
}

// SalaryChange holds salary change details
type SalaryChange struct {
	MemberID  string  `json:"member_id"`
	MemberNo  string  `json:"member_no"`
	OldSalary int64   `json:"old_salary"`
	NewSalary int64   `json:"new_salary"`
	ChangePct float64 `json:"change_pct"`
}

// Service manages bulk processing operations
type Service struct {
	db *db.DB
}

// NewService creates a new bulk processing service
func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// ImportMembersCSV imports members from a CSV file
// Expected columns: member_no,first_name,last_name,other_names,gender,date_of_birth,nationality,id_number,kra_pin,email,phone,department,designation,payroll_no,basic_salary,date_joined_scheme,sponsor_id
func (s *Service) ImportMembersCSV(ctx context.Context, schemeID string, reader io.Reader, createdBy string) (*ImportResult, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV headers: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[strings.TrimSpace(strings.ToLower(h))] = i
	}

	requiredCols := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	result := &ImportResult{}
	var members []memberImport

	rowNum := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Reason: err.Error()})
			result.Failed++
			rowNum++
			continue
		}
		rowNum++
		result.TotalRows++

		m, errs := parseMemberImport(record, colIndex, rowNum, schemeID)
		if len(errs) > 0 {
			result.Errors = append(result.Errors, errs...)
			result.Failed++
			continue
		}

		// Validate against existing data
		warnings, err := s.validateMemberImport(ctx, &m)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Reason: err.Error()})
			result.Failed++
			continue
		}
		result.Warnings = append(result.Warnings, warnings...)

		members = append(members, m)
	}

	// Import in a single transaction
	if len(members) > 0 {
		err = s.db.Transactional(ctx, func(tx *sql.Tx) error {
			for _, m := range members {
				if err := m.insert(ctx, tx, createdBy); err != nil {
					return fmt.Errorf("row %d: %w", m.RowNum, err)
				}
				result.Success++
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("bulk import failed: %w", err)
		}
	}

	result.Failed = result.TotalRows - result.Success
	return result, nil
}

type memberImport struct {
	RowNum           int
	SchemeID         string
	MemberNo         string
	FirstName        string
	LastName         string
	OtherNames       string
	Gender           string
	DateOfBirth      time.Time
	Nationality      string
	IDNumber         string
	KRAPIN           string
	Email            string
	Phone            string
	Department       string
	Designation      string
	PayrollNo        string
	BasicSalary      int64
	DateJoinedScheme time.Time
	SponsorID        string
}

func parseMemberImport(record []string, colIndex map[string]int, rowNum int, schemeID string) (memberImport, []ImportError) {
	var m memberImport
	var errs []ImportError
	m.RowNum = rowNum
	m.SchemeID = schemeID

	get := func(col string) string {
		if idx, ok := colIndex[col]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	m.MemberNo = get("member_no")
	m.FirstName = get("first_name")
	m.LastName = get("last_name")
	m.OtherNames = get("other_names")
	m.Gender = get("gender")
	m.Nationality = get("nationality")
	m.IDNumber = get("id_number")
	m.KRAPIN = get("kra_pin")
	m.Email = get("email")
	m.Phone = get("phone")
	m.Department = get("department")
	m.Designation = get("designation")
	m.PayrollNo = get("payroll_no")
	m.SponsorID = get("sponsor_id")

	if m.MemberNo == "" {
		errs = append(errs, ImportError{Row: rowNum, Field: "member_no", Reason: "required"})
	}
	if m.FirstName == "" {
		errs = append(errs, ImportError{Row: rowNum, Field: "first_name", Reason: "required"})
	}
	if m.LastName == "" {
		errs = append(errs, ImportError{Row: rowNum, Field: "last_name", Reason: "required"})
	}

	dob, err := time.Parse("2006-01-02", get("date_of_birth"))
	if err != nil {
		errs = append(errs, ImportError{Row: rowNum, Field: "date_of_birth", Reason: "invalid format (use YYYY-MM-DD)"})
	} else {
		m.DateOfBirth = dob
	}

	djs, err := time.Parse("2006-01-02", get("date_joined_scheme"))
	if err != nil {
		errs = append(errs, ImportError{Row: rowNum, Field: "date_joined_scheme", Reason: "invalid format (use YYYY-MM-DD)"})
	} else {
		m.DateJoinedScheme = djs
	}

	if salStr := get("basic_salary"); salStr != "" {
		sal, err := strconv.ParseInt(salStr, 10, 64)
		if err != nil {
			errs = append(errs, ImportError{Row: rowNum, Field: "basic_salary", Reason: "invalid number"})
		} else {
			m.BasicSalary = sal
		}
	}

	return m, errs
}

func (s *Service) validateMemberImport(ctx context.Context, m *memberImport) ([]ImportWarning, error) {
	var warnings []ImportWarning

	// Check for duplicate member number
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM members WHERE member_no = $1 AND scheme_id = $2)
	`, m.MemberNo, m.SchemeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		warnings = append(warnings, ImportWarning{
			Row:     m.RowNum,
			Message: fmt.Sprintf("Member %s already exists - will be skipped", m.MemberNo),
		})
	}

	// Validate sponsor exists if provided
	if m.SponsorID != "" {
		var sponsorExists bool
		err := s.db.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM sponsors WHERE id = $1 AND scheme_id = $2)
		`, m.SponsorID, m.SchemeID).Scan(&sponsorExists)
		if err != nil {
			return nil, fmt.Errorf("check sponsor: %w", err)
		}
		if !sponsorExists {
			return nil, fmt.Errorf("sponsor %s not found", m.SponsorID)
		}
	}

	return warnings, nil
}

func (m *memberImport) insert(ctx context.Context, tx *sql.Tx, createdBy string) error {
	// Skip if already exists
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM members WHERE member_no = $1 AND scheme_id = $2)
	`, m.MemberNo, m.SchemeID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // Skip duplicates
	}

	query := `
		INSERT INTO members (id, scheme_id, member_no, first_name, last_name, other_names, gender,
		                     date_of_birth, nationality, id_number, kra_pin, email, phone,
		                     department, designation, payroll_no, basic_salary, date_joined_scheme,
		                     sponsor_id, membership_status, account_balance, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 'active', 0, NOW(), NOW())
		RETURNING id
	`
	_, err = tx.ExecContext(ctx, query,
		m.SchemeID, m.MemberNo, m.FirstName, m.LastName, m.OtherNames, m.Gender,
		m.DateOfBirth, m.Nationality, m.IDNumber, m.KRAPIN, m.Email, m.Phone,
		m.Department, m.Designation, m.PayrollNo, m.BasicSalary, m.DateJoinedScheme,
		m.SponsorID,
	)
	return err
}

// ValidateBulkUpdate compares new data against existing data and reports discrepancies
func (s *Service) ValidateBulkUpdate(ctx context.Context, schemeID string, newMemberNos []string, newSalaries map[string]int64) (*BulkValidation, error) {
	v := &BulkValidation{IsValid: true}

	// Get existing member numbers
	existingRows, err := s.db.QueryContext(ctx, `
		SELECT member_no, basic_salary FROM members WHERE scheme_id = $1
	`, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query existing members: %w", err)
	}
	defer existingRows.Close()

	existingMembers := make(map[string]int64)
	for existingRows.Next() {
		var no string
		var sal int64
		if err := existingRows.Scan(&no, &sal); err != nil {
			continue
		}
		existingMembers[no] = sal
	}

	// Find new members
	newSet := make(map[string]bool)
	for _, no := range newMemberNos {
		newSet[no] = true
		if _, exists := existingMembers[no]; !exists {
			v.NewMembers = append(v.NewMembers, no)
		}
	}

	// Find removed members
	for no := range existingMembers {
		if !newSet[no] {
			v.RemovedMembers = append(v.RemovedMembers, no)
		}
	}

	// Find salary changes
	for memberNo, newSalary := range newSalaries {
		oldSalary, exists := existingMembers[memberNo]
		if exists && oldSalary != newSalary {
			changePct := float64(0)
			if oldSalary > 0 {
				changePct = float64(newSalary-oldSalary) / float64(oldSalary) * 100
			}
			v.SalaryChanges = append(v.SalaryChanges, SalaryChange{
				MemberNo:  memberNo,
				OldSalary: oldSalary,
				NewSalary: newSalary,
				ChangePct: changePct,
			})
		}
	}

	return v, nil
}

// ProcessRetirements processes retirement for eligible members
func (s *Service) ProcessRetirements(ctx context.Context, schemeID string, retirementType string) (*BulkResult, error) {
	return s.processBulkStatusChange(ctx, schemeID, "retired", retirementType)
}

// ProcessEarlyLeavers processes early leavers (active to deferred or refund)
func (s *Service) ProcessEarlyLeavers(ctx context.Context, schemeID string) (*BulkResult, error) {
	return s.processBulkStatusChange(ctx, schemeID, "deferred", "early_leaver")
}

type BulkResult struct {
	Processed int      `json:"processed"`
	Errors    []string `json:"errors,omitempty"`
}

func (s *Service) processBulkStatusChange(ctx context.Context, schemeID, newStatus, reason string) (*BulkResult, error) {
	result := &BulkResult{}

	err := s.db.Transactional(ctx, func(tx *sql.Tx) error {
		query := `
			UPDATE members SET membership_status = $1, updated_at = NOW()
			WHERE scheme_id = $2 AND membership_status = 'active'
			AND expected_retirement <= NOW()
			RETURNING id, member_no
		`
		rows, err := tx.QueryContext(ctx, query, newStatus, schemeID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id, memberNo string
			if err := rows.Scan(&id, &memberNo); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
				continue
			}
			result.Processed++

			// Log audit event
			_, _ = tx.ExecContext(ctx, `
				INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
				VALUES (uuid_generate_v4(), $1, 'member', $2, $3, 'system', $4, NOW())
			`, schemeID, id, "bulk_status_change", fmt.Sprintf("Type: %s, New status: %s", reason, newStatus))
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("bulk status change: %w", err)
	}
	return result, nil
}

// AnnualPosting posts annual contributions for all active members
func (s *Service) AnnualPosting(ctx context.Context, schemeID string, year int) (*BulkResult, error) {
	result := &BulkResult{}

	query := `
		UPDATE members
		SET account_balance = account_balance + (
			SELECT COALESCE(SUM(total_amount), 0)
			FROM contributions
			WHERE member_id = members.id
			AND EXTRACT(YEAR FROM period) = $1
		),
		updated_at = NOW()
		WHERE scheme_id = $2 AND membership_status = 'active'
	`
	res, err := s.db.ExecContext(ctx, query, year, schemeID)
	if err != nil {
		return nil, fmt.Errorf("annual posting: %w", err)
	}

	rows, _ := res.RowsAffected()
	result.Processed = int(rows)
	return result, nil
}

// BatchStatementData holds data for batch statement generation
type BatchStatementData struct {
	MemberID   string `json:"member_id"`
	MemberNo   string `json:"member_no"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	Balance    int64  `json:"balance"`
	Email      string `json:"email"`
	Statement  []byte `json:"-"`
}

// GetBatchStatementData retrieves data for batch statement printing
func (s *Service) GetBatchStatementData(ctx context.Context, schemeID, department, status string, startDate, endDate time.Time) ([]BatchStatementData, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name, m.department,
		       m.account_balance, m.email
		FROM members m WHERE m.scheme_id = $1
	`
	args := []interface{}{schemeID}
	argCount := 1

	if department != "" {
		argCount++
		query += fmt.Sprintf(" AND m.department = $%d", argCount)
		args = append(args, department)
	}
	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND m.membership_status = $%d", argCount)
		args = append(args, status)
	}

	query += " ORDER BY m.department, m.last_name, m.first_name"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query batch data: %w", err)
	}
	defer rows.Close()

	var data []BatchStatementData
	for rows.Next() {
		var d BatchStatementData
		if err := rows.Scan(&d.MemberID, &d.MemberNo, &d.FullName, &d.Department, &d.Balance, &d.Email); err != nil {
			continue
		}
		data = append(data, d)
	}
	return data, rows.Err()
}
