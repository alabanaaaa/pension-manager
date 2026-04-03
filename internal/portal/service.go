package portal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"pension-manager/internal/db"
)

// Service manages member portal operations
type Service struct {
	db *db.DB
}

// NewService creates a new member portal service
func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// MemberProfile holds the member's portal profile data
type MemberProfile struct {
	PersonalInfo   PersonalInfo   `json:"personal_info"`
	ContactInfo    ContactInfo    `json:"contact_info"`
	EmploymentInfo EmploymentInfo `json:"employment_info"`
	Beneficiaries  []Beneficiary  `json:"beneficiaries"`
	MedicalLimits  MedicalLimits  `json:"medical_limits"`
	AccountSummary AccountSummary `json:"account_summary"`
}

// PersonalInfo holds member's personal details
type PersonalInfo struct {
	FullName      string    `json:"full_name"`
	NationalID    string    `json:"national_id"`
	Gender        string    `json:"gender"`
	DateOfBirth   time.Time `json:"date_of_birth"`
	PhoneNumber   string    `json:"phone_number"`
	MaritalStatus string    `json:"marital_status"`
	SpouseName    string    `json:"spouse_name,omitempty"`
	KRAPIN        string    `json:"kra_pin,omitempty"`
	Nationality   string    `json:"nationality"`
	Age           int       `json:"age"`
}

// ContactInfo holds member's contact details
type ContactInfo struct {
	Email         string `json:"email"`
	MobileNumber  string `json:"mobile_number"`
	WorkTelephone string `json:"work_telephone,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	PostalAddress string `json:"postal_address,omitempty"`
	Town          string `json:"town,omitempty"`
}

// EmploymentInfo holds member's employment details
type EmploymentInfo struct {
	SchemeName       string    `json:"scheme_name"`
	SponsorName      string    `json:"sponsor_name"`
	SponsorCode      string    `json:"sponsor_code"`
	DateJoinedScheme time.Time `json:"date_joined_scheme"`
	PayrollNo        string    `json:"payroll_no,omitempty"`
	BankName         string    `json:"bank_name,omitempty"`
	BankBranch       string    `json:"bank_branch,omitempty"`
	BankAccount      string    `json:"bank_account,omitempty"`
	Address          string    `json:"address,omitempty"`
	Town             string    `json:"town,omitempty"`
	DateFirstAppt    time.Time `json:"date_first_appt,omitempty"`
	Designation      string    `json:"designation,omitempty"`
	Department       string    `json:"department,omitempty"`
	BasicSalary      int64     `json:"basic_salary"`
	MemberNo         string    `json:"member_no"`
}

// Beneficiary holds beneficiary details for portal display
type Beneficiary struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	NationalID      string    `json:"national_id"`
	MobileTelephone string    `json:"mobile_telephone"`
	Relationship    string    `json:"relationship"`
	PhysicalAddress string    `json:"physical_address"`
	AllocationPct   float64   `json:"allocation_pct"`
	DateOfBirth     time.Time `json:"date_of_birth,omitempty"`
}

// MedicalLimits holds member's medical coverage limits
type MedicalLimits struct {
	InpatientLimit  int64 `json:"inpatient_limit"`
	OutpatientLimit int64 `json:"outpatient_limit"`
}

// AccountSummary holds member's account summary
type AccountSummary struct {
	AccountBalance     int64     `json:"account_balance"`
	TotalWithdrawals   int64     `json:"total_withdrawals"`
	LastContribution   time.Time `json:"last_contribution,omitempty"`
	MembershipStatus   string    `json:"membership_status"`
	ExpectedRetirement time.Time `json:"expected_retirement,omitempty"`
}

// GetMemberProfile retrieves the full member profile for the portal
func (s *Service) GetMemberProfile(ctx context.Context, memberID string) (*MemberProfile, error) {
	query := `
		SELECT m.member_no, m.first_name, m.last_name, m.other_names, m.gender, m.date_of_birth,
		       m.nationality, m.id_number, m.kra_pin, m.email, m.phone, m.postal_address,
		       m.postal_code, m.town, m.marital_status, m.spouse_name, m.payroll_no,
		       m.designation, m.department, m.date_first_appt, m.date_joined_scheme,
		       m.expected_retirement, m.membership_status, m.basic_salary, m.account_balance,
		       m.inpatient_limit, m.outpatient_limit, m.total_withdrawals, m.last_contribution,
		       m.bank_name, m.bank_branch, m.bank_account,
		       COALESCE(s.name, '') as sponsor_name, COALESCE(s.code, '') as sponsor_code
		FROM members m
		LEFT JOIN sponsors s ON s.id = m.sponsor_id
		WHERE m.id = $1
	`
	var profile MemberProfile
	var otherNames, kraPin, postalAddr, postalCode, town, spouseName, payrollNo, designation, department sql.NullString
	var dateFirstAppt, expectedRetirement, lastContribution sql.NullTime
	var sponsorName, sponsorCode, bankName, bankBranch, bankAccount sql.NullString

	err := s.db.QueryRowContext(ctx, query, memberID).Scan(
		&profile.EmploymentInfo.MemberNo,
		&profile.PersonalInfo.FullName, &profile.PersonalInfo.FullName, // first+last combined later
		&otherNames, &profile.PersonalInfo.Gender, &profile.PersonalInfo.DateOfBirth,
		&profile.PersonalInfo.Nationality, &profile.PersonalInfo.NationalID, &kraPin,
		&profile.ContactInfo.Email, &profile.ContactInfo.MobileNumber, &postalAddr,
		&postalCode, &town, &profile.PersonalInfo.MaritalStatus, &spouseName, &payrollNo,
		&designation, &department, &dateFirstAppt, &profile.EmploymentInfo.DateJoinedScheme,
		&expectedRetirement, &profile.AccountSummary.MembershipStatus, &profile.EmploymentInfo.BasicSalary,
		&profile.AccountSummary.AccountBalance, &profile.MedicalLimits.InpatientLimit,
		&profile.MedicalLimits.OutpatientLimit, &profile.AccountSummary.TotalWithdrawals,
		&lastContribution, &bankName, &bankBranch, &bankAccount,
		&sponsorName, &sponsorCode,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get member profile: %w", err)
	}

	// Build full name
	fullName := profile.PersonalInfo.FullName
	if otherNames.Valid {
		fullName += " " + otherNames.String
	}
	profile.PersonalInfo.FullName = fullName

	// Set optional fields
	if kraPin.Valid {
		profile.PersonalInfo.KRAPIN = kraPin.String
	}
	if spouseName.Valid {
		profile.PersonalInfo.SpouseName = spouseName.String
	}
	if postalAddr.Valid {
		profile.ContactInfo.PostalAddress = postalAddr.String
	}
	if postalCode.Valid {
		profile.ContactInfo.PostalCode = postalCode.String
	}
	if town.Valid {
		profile.ContactInfo.Town = town.String
	}
	if payrollNo.Valid {
		profile.EmploymentInfo.PayrollNo = payrollNo.String
	}
	if designation.Valid {
		profile.EmploymentInfo.Designation = designation.String
	}
	if department.Valid {
		profile.EmploymentInfo.Department = department.String
	}
	if dateFirstAppt.Valid {
		profile.EmploymentInfo.DateFirstAppt = dateFirstAppt.Time
	}
	if expectedRetirement.Valid {
		profile.AccountSummary.ExpectedRetirement = expectedRetirement.Time
	}
	if lastContribution.Valid {
		profile.AccountSummary.LastContribution = lastContribution.Time
	}
	if sponsorName.Valid {
		profile.EmploymentInfo.SponsorName = sponsorName.String
	}
	if sponsorCode.Valid {
		profile.EmploymentInfo.SponsorCode = sponsorCode.String
	}
	if bankName.Valid {
		profile.EmploymentInfo.BankName = bankName.String
	}
	if bankBranch.Valid {
		profile.EmploymentInfo.BankBranch = bankBranch.String
	}
	if bankAccount.Valid {
		profile.EmploymentInfo.BankAccount = bankAccount.String
	}

	profile.ContactInfo.WorkTelephone = profile.EmploymentInfo.MemberNo // placeholder
	profile.EmploymentInfo.Address = profile.ContactInfo.PostalAddress
	profile.EmploymentInfo.Town = profile.ContactInfo.Town

	// Calculate age
	now := time.Now()
	profile.PersonalInfo.Age = now.Year() - profile.PersonalInfo.DateOfBirth.Year()
	if now.YearDay() < profile.PersonalInfo.DateOfBirth.YearDay() {
		profile.PersonalInfo.Age--
	}

	// Get beneficiaries
	benefs, err := s.GetMemberBeneficiaries(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get beneficiaries: %w", err)
	}
	profile.Beneficiaries = benefs

	return &profile, nil
}

// GetMemberBeneficiaries retrieves beneficiaries for a member
func (s *Service) GetMemberBeneficiaries(ctx context.Context, memberID string) ([]Beneficiary, error) {
	query := `
		SELECT id, name, relationship, date_of_birth, id_number, phone, physical_address, allocation_pct
		FROM beneficiaries WHERE member_id = $1 ORDER BY allocation_pct DESC
	`
	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("query beneficiaries: %w", err)
	}
	defer rows.Close()

	var benefs []Beneficiary
	for rows.Next() {
		var b Beneficiary
		var idNumber, phone, physicalAddr sql.NullString
		var dob sql.NullTime
		if err := rows.Scan(&b.ID, &b.Name, &b.Relationship, &dob, &idNumber, &phone, &physicalAddr, &b.AllocationPct); err != nil {
			return nil, fmt.Errorf("scan beneficiary: %w", err)
		}
		if idNumber.Valid {
			b.NationalID = idNumber.String
		}
		if phone.Valid {
			b.MobileTelephone = phone.String
		}
		if physicalAddr.Valid {
			b.PhysicalAddress = physicalAddr.String
		}
		if dob.Valid {
			b.DateOfBirth = dob.Time
		}
		benefs = append(benefs, b)
	}
	return benefs, rows.Err()
}

// MemberContribution holds contribution data for portal display
type MemberContribution struct {
	ID             string    `json:"id"`
	Period         time.Time `json:"period"`
	EmployeeAmount int64     `json:"employee_amount"`
	EmployerAmount int64     `json:"employer_amount"`
	AVCAmount      int64     `json:"avc_amount"`
	TotalAmount    int64     `json:"total_amount"`
	PaymentMethod  string    `json:"payment_method"`
	Status         string    `json:"status"`
	ReceiptNo      string    `json:"receipt_no,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetMemberContributions retrieves contributions for a member with date filtering
func (s *Service) GetMemberContributions(ctx context.Context, memberID string, startDate, endDate time.Time) ([]MemberContribution, error) {
	query := `
		SELECT id, period, employee_amount, employer_amount, avc_amount, total_amount,
		       payment_method, status, receipt_no, created_at
		FROM contributions WHERE member_id = $1
	`
	args := []interface{}{memberID}
	argCount := 1

	if !startDate.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND period >= $%d", argCount)
		args = append(args, startDate)
	}
	if !endDate.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND period <= $%d", argCount)
		args = append(args, endDate)
	}

	query += " ORDER BY period DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query contributions: %w", err)
	}
	defer rows.Close()

	var contributions []MemberContribution
	for rows.Next() {
		var c MemberContribution
		var receiptNo sql.NullString
		if err := rows.Scan(&c.ID, &c.Period, &c.EmployeeAmount, &c.EmployerAmount,
			&c.AVCAmount, &c.TotalAmount, &c.PaymentMethod, &c.Status, &receiptNo, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan contribution: %w", err)
		}
		if receiptNo.Valid {
			c.ReceiptNo = receiptNo.String
		}
		contributions = append(contributions, c)
	}
	return contributions, rows.Err()
}

// GetAnnualContributions retrieves contributions grouped by year
func (s *Service) GetAnnualContributions(ctx context.Context, memberID string) ([]AnnualContribution, error) {
	query := `
		SELECT DATE_TRUNC('year', period) as year,
		       SUM(employee_amount) as employee_total,
		       SUM(employer_amount) as employer_total,
		       SUM(avc_amount) as avc_total,
		       SUM(total_amount) as grand_total
		FROM contributions WHERE member_id = $1
		GROUP BY DATE_TRUNC('year', period)
		ORDER BY year DESC
	`
	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("query annual contributions: %w", err)
	}
	defer rows.Close()

	var annual []AnnualContribution
	for rows.Next() {
		var a AnnualContribution
		if err := rows.Scan(&a.Year, &a.EmployeeTotal, &a.EmployerTotal, &a.AVCTotal, &a.GrandTotal); err != nil {
			return nil, fmt.Errorf("scan annual contribution: %w", err)
		}
		annual = append(annual, a)
	}
	return annual, rows.Err()
}

// AnnualContribution holds yearly contribution totals
type AnnualContribution struct {
	Year          time.Time `json:"year"`
	EmployeeTotal int64     `json:"employee_total"`
	EmployerTotal int64     `json:"employer_total"`
	AVCTotal      int64     `json:"avc_total"`
	GrandTotal    int64     `json:"grand_total"`
}

// ChangeRequest represents a member-initiated change request
type ChangeRequest struct {
	ID              string          `json:"id"`
	RequestType     string          `json:"request_type"` // contact_change, add_beneficiary, remove_beneficiary, change_allocation
	Status          string          `json:"status"`       // pending, approved, rejected
	BeforeData      json.RawMessage `json:"before_data,omitempty"`
	AfterData       json.RawMessage `json:"after_data"`
	RejectionReason *string         `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty"`
}

// CreateChangeRequest creates a new change request from a member
func (s *Service) CreateChangeRequest(ctx context.Context, memberID, schemeID, requestType string, beforeData, afterData interface{}) error {
	beforeJSON, _ := json.Marshal(beforeData)
	afterJSON, _ := json.Marshal(afterData)

	query := `
		INSERT INTO pending_changes (id, scheme_id, entity_type, entity_id, change_type, requested_by,
		                             before_data, after_data, status, created_at)
		VALUES (uuid_generate_v4(), $1, 'member', $2, $3, $2, $4, $5, 'pending', NOW())
	`
	_, err := s.db.ExecContext(ctx, query, schemeID, memberID, requestType, beforeJSON, afterJSON)
	if err != nil {
		return fmt.Errorf("create change request: %w", err)
	}
	return nil
}

// GetMemberChangeRequests retrieves change requests for a member
func (s *Service) GetMemberChangeRequests(ctx context.Context, memberID string) ([]ChangeRequest, error) {
	query := `
		SELECT id, change_type, before_data, after_data, status, rejection_reason, created_at, reviewed_at
		FROM pending_changes WHERE entity_type = 'member' AND entity_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("query change requests: %w", err)
	}
	defer rows.Close()

	var requests []ChangeRequest
	for rows.Next() {
		var cr ChangeRequest
		var rejectionReason sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(&cr.ID, &cr.RequestType, &cr.BeforeData, &cr.AfterData,
			&cr.Status, &rejectionReason, &cr.CreatedAt, &reviewedAt); err != nil {
			return nil, fmt.Errorf("scan change request: %w", err)
		}
		if rejectionReason.Valid {
			cr.RejectionReason = &rejectionReason.String
		}
		if reviewedAt.Valid {
			cr.ReviewedAt = &reviewedAt.Time
		}
		requests = append(requests, cr)
	}
	return requests, rows.Err()
}

// Feedback represents member feedback
type Feedback struct {
	ID        string    `json:"id"`
	MemberID  string    `json:"member_id"`
	SchemeID  string    `json:"scheme_id"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"` // open, in_progress, resolved
	CreatedAt time.Time `json:"created_at"`
}

// SubmitFeedback creates a new feedback entry
func (s *Service) SubmitFeedback(ctx context.Context, memberID, schemeID, subject, message string) error {
	query := `
		INSERT INTO feedback (id, member_id, scheme_id, subject, message, status, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, 'open', NOW())
	`
	_, err := s.db.ExecContext(ctx, query, memberID, schemeID, subject, message)
	if err != nil {
		return fmt.Errorf("submit feedback: %w", err)
	}
	return nil
}

// GetMemberFeedback retrieves feedback for a member
func (s *Service) GetMemberFeedback(ctx context.Context, memberID string) ([]Feedback, error) {
	query := `
		SELECT id, member_id, scheme_id, subject, message, status, created_at
		FROM feedback WHERE member_id = $1 ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("query feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []Feedback
	for rows.Next() {
		var f Feedback
		if err := rows.Scan(&f.ID, &f.MemberID, &f.SchemeID, &f.Subject, &f.Message, &f.Status, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		feedbacks = append(feedbacks, f)
	}
	return feedbacks, rows.Err()
}

// TrackLogin records a member login for utilization tracking
func (s *Service) TrackLogin(ctx context.Context, memberID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO member_login_log (id, member_id, login_at)
		VALUES (uuid_generate_v4(), $1, NOW())
	`, memberID)
	return err
}

// GetMemberLoginStats retrieves login statistics for a member
func (s *Service) GetMemberLoginStats(ctx context.Context, memberID string) (*LoginStats, error) {
	stats := &LoginStats{}
	query := `
		SELECT COUNT(*) as total_logins,
		       MAX(login_at) as last_login,
		       COUNT(*) FILTER (WHERE login_at > NOW() - INTERVAL '30 days') as logins_last_30_days
		FROM member_login_log WHERE member_id = $1
	`
	err := s.db.QueryRowContext(ctx, query, memberID).Scan(&stats.TotalLogins, &stats.LastLogin, &stats.LoginsLast30Days)
	if err != nil {
		return nil, fmt.Errorf("get login stats: %w", err)
	}
	return stats, nil
}

// LoginStats holds member login statistics
type LoginStats struct {
	TotalLogins      int        `json:"total_logins"`
	LastLogin        *time.Time `json:"last_login,omitempty"`
	LoginsLast30Days int        `json:"logins_last_30_days"`
}
