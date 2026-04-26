package reconciliation

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

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

type RemittanceSchedule struct {
	ID            string    `json:"id"`
	SchemeID      string    `json:"scheme_id"`
	SponsorID     string    `json:"sponsor_id"`
	Period        time.Time `json:"period"`
	TotalAmount   int64     `json:"total_amount"`
	EmployeeTotal int64     `json:"employee_total"`
	EmployerTotal int64     `json:"employer_total"`
	Status        string    `json:"status"`
	ReceiptNo     string    `json:"receipt_no,omitempty"`
	ChequeNo      string    `json:"cheque_no,omitempty"`
	BankName      string    `json:"bank_name,omitempty"`
	PaymentMethod string    `json:"payment_method"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	OnHold        bool      `json:"on_hold"`
	HoldReason    string    `json:"hold_reason,omitempty"`
}

type ScheduleLine struct {
	MemberNo             string `json:"member_no"`
	MemberName           string `json:"member_name"`
	Department           string `json:"department"`
	BasicSalary          int64  `json:"basic_salary"`
	EmployeeContribution int64  `json:"employee_contribution"`
	EmployerContribution int64  `json:"employer_contribution"`
	AVC                  int64  `json:"avc"`
	TotalContribution    int64  `json:"total_contribution"`
}

type ReconciliationResult struct {
	ScheduleID        string       `json:"schedule_id"`
	ScheduleAmount    int64        `json:"schedule_amount"`
	ReceivedAmount    int64        `json:"received_amount"`
	Difference        int64        `json:"difference"`
	Status            string       `json:"status"`
	MemberDifferences []MemberDiff `json:"member_differences,omitempty"`
	NewMembers        []string     `json:"new_members,omitempty"`
	RemovedMembers    []string     `json:"removed_members,omitempty"`
	Errors            []string     `json:"errors,omitempty"`
}

type MemberDiff struct {
	MemberNo   string `json:"member_no"`
	MemberName string `json:"member_name"`
	Field      string `json:"field"`
	Expected   int64  `json:"expected"`
	Actual     int64  `json:"actual"`
	Diff       int64  `json:"difference"`
}

type SponsorSubmission struct {
	ScheduleID  string
	Lines       []ScheduleLine
	TotalAmount int64
}

func (s *Service) ImportRemittanceSchedule(ctx context.Context, schemeID, sponsorID string, period time.Time, reader io.Reader, paymentMethod, receiptNo, chequeNo, bankName string, createdBy string) (*RemittanceSchedule, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV headers: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var lines []ScheduleLine
	var totalAmount int64

	rowNum := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rowNum++
			continue
		}
		rowNum++

		line, err := parseScheduleLine(record, colIndex, rowNum)
		if err != nil {
			continue
		}
		lines = append(lines, line)
		totalAmount += line.TotalContribution
	}

	var scheduleID string
	err = s.db.Transactional(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO remittance_schedules (
				id, scheme_id, sponsor_id, period, total_amount, status, receipt_no, 
				cheque_no, bank_name, payment_method, created_by, created_at, on_hold
			) VALUES (
				uuid_generate_v4(), $1, $2, $3, $4, 'pending', $5, $6, $7, $8, $9, NOW(), false
			) RETURNING id
		`
		err := tx.QueryRowContext(ctx, query,
			schemeID, sponsorID, period, totalAmount, receiptNo, chequeNo, bankName, paymentMethod, createdBy,
		).Scan(&scheduleID)
		if err != nil {
			return fmt.Errorf("insert schedule: %w", err)
		}

		for _, line := range lines {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO remittance_schedule_lines (
					id, schedule_id, member_no, member_name, department, basic_salary,
					employee_contribution, employer_contribution, avc, total_contribution
				) VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, scheduleID, line.MemberNo, line.MemberName, line.Department, line.BasicSalary,
				line.EmployeeContribution, line.EmployerContribution, line.AVC, line.TotalContribution)
			if err != nil {
				return fmt.Errorf("insert line: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &RemittanceSchedule{
		ID:            scheduleID,
		SchemeID:      schemeID,
		SponsorID:     sponsorID,
		Period:        period,
		TotalAmount:   totalAmount,
		Status:        "pending",
		ReceiptNo:     receiptNo,
		ChequeNo:      chequeNo,
		BankName:      bankName,
		PaymentMethod: paymentMethod,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
	}, nil
}

func parseScheduleLine(record []string, colIndex map[string]int, rowNum int) (ScheduleLine, error) {
	get := func(col string) string {
		if idx, ok := colIndex[col]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	parseInt := func(s string) int64 {
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseInt(strings.ReplaceAll(s, ",", ""), 10, 64)
		return v
	}

	return ScheduleLine{
		MemberNo:             get("member_no"),
		MemberName:           get("member_name"),
		Department:           get("department"),
		BasicSalary:          parseInt(get("basic_salary")),
		EmployeeContribution: parseInt(get("employee_contribution")),
		EmployerContribution: parseInt(get("employer_contribution")),
		AVC:                  parseInt(get("avc")),
		TotalContribution:    parseInt(get("total_contribution")),
	}, nil
}

func (s *Service) ReconcileSchedule(ctx context.Context, scheduleID string) (*ReconciliationResult, error) {
	result := &ReconciliationResult{ScheduleID: scheduleID}

	var scheduleTotal int64
	var schemeID, sponsorID string
	err := s.db.QueryRowContext(ctx, `
		SELECT total_amount, scheme_id, sponsor_id FROM remittance_schedules WHERE id = $1
	`, scheduleID).Scan(&scheduleTotal, &schemeID, &sponsorID)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	result.ScheduleAmount = scheduleTotal

	rows, err := s.db.QueryContext(ctx, `
		SELECT m.member_no, m.first_name || ' ' || m.last_name,
		       COALESCE(m.basic_salary, 0),
		       COALESCE((
		           SELECT c.employee_amount FROM contributions c
		           WHERE c.member_id = m.id AND c.period = (
		               SELECT period FROM remittance_schedules WHERE id = $1
		           )
		       ), 0) as employee_contrib,
		       COALESCE((
		           SELECT c.employer_amount FROM contributions c
		           WHERE c.member_id = m.id AND c.period = (
		               SELECT period FROM remittance_schedules WHERE id = $1
		           )
		       ), 0) as employer_contrib,
		       COALESCE((
		           SELECT c.total_amount FROM contributions c
		           WHERE c.member_id = m.id AND c.period = (
		               SELECT period FROM remittance_schedules WHERE id = $1
		           )
		       ), 0) as total_contrib
		FROM remittance_schedule_lines rsl
		JOIN members m ON m.member_no = rsl.member_no
		WHERE rsl.schedule_id = $1
	`, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expectedAmount int64
	for rows.Next() {
		var memberNo, memberName string
		var expectedEmp, expectedEr, actualTotal int64
		if err := rows.Scan(&memberNo, &memberName, &expectedEmp, &expectedEr, &actualTotal); err != nil {
			continue
		}
		expectedAmount += expectedEmp + expectedEr

		if actualTotal != (expectedEmp + expectedEr) {
			result.MemberDifferences = append(result.MemberDifferences, MemberDiff{
				MemberNo:   memberNo,
				MemberName: memberName,
				Field:      "contribution",
				Expected:   expectedEmp + expectedEr,
				Actual:     actualTotal,
				Diff:       actualTotal - (expectedEmp + expectedEr),
			})
		}
	}

	_, receivedAmount, _ := s.getReceivedAmount(ctx, schemeID)
	result.ReceivedAmount = receivedAmount
	result.Difference = receivedAmount - scheduleTotal

	if result.Difference == 0 && len(result.MemberDifferences) == 0 {
		result.Status = "matched"
	} else if result.Difference > 0 {
		result.Status = "over_remittance"
	} else {
		result.Status = "under_remittance"
	}

	return result, nil
}

func (s *Service) getReceivedAmount(ctx context.Context, schemeID string) (string, int64, error) {
	var amount int64
	var paymentRef string
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0), MAX(payment_ref)
		FROM contributions WHERE scheme_id = $1 AND status = 'confirmed'
	`, schemeID).Scan(&amount, &paymentRef)
	return paymentRef, amount, err
}

func (s *Service) PutScheduleOnHold(ctx context.Context, scheduleID, reason, userID string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE remittance_schedules SET on_hold = true, hold_reason = $1 WHERE id = $2
	`, reason, scheduleID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}

func (s *Service) ReleaseSchedule(ctx context.Context, scheduleID, userID string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE remittance_schedules SET on_hold = false, hold_reason = NULL WHERE id = $1
	`, scheduleID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}

func (s *Service) AllocateContributions(ctx context.Context, scheduleID string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT member_no, department, employee_contribution, employer_contribution, avc
			FROM remittance_schedule_lines WHERE schedule_id = $1
		`, scheduleID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var memberNo, department string
			var empContrib, erContrib, avc int64
			if err := rows.Scan(&memberNo, &department, &empContrib, &erContrib, &avc); err != nil {
				continue
			}

			var period time.Time
			var schemeID, sponsorID string
			err := tx.QueryRowContext(ctx, `
				SELECT period, scheme_id, sponsor_id FROM remittance_schedules WHERE id = $1
			`, scheduleID).Scan(&period, &schemeID, &sponsorID)
			if err != nil {
				continue
			}

			var memberID string
			err = tx.QueryRowContext(ctx, `
				SELECT id FROM members WHERE member_no = $1 AND scheme_id = $2
			`, memberNo, schemeID).Scan(&memberID)
			if err != nil {
				continue
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO contributions (
					id, member_id, scheme_id, sponsor_id, period,
					employee_amount, employer_amount, avc_amount, total_amount,
					payment_method, status, registered, created_at
				) VALUES (
					uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8,
					'bank_transfer', 'confirmed', true, NOW()
				)
			`, memberID, schemeID, sponsorID, period, empContrib, erContrib, avc, empContrib+erContrib+avc)
			if err != nil {
				return fmt.Errorf("allocate contribution for %s: %w", memberNo, err)
			}

			_, _ = tx.ExecContext(ctx, `
				UPDATE members SET account_balance = account_balance + $1, last_contribution = NOW()
				WHERE id = $2
			`, empContrib+erContrib+avc, memberID)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE remittance_schedules SET status = 'allocated' WHERE id = $1
		`, scheduleID)

		return err
	})
}

func (s *Service) WarnContributionIrregularity(ctx context.Context, schemeID string, period time.Time) ([]string, error) {
	var warnings []string

	rows, err := s.db.QueryContext(ctx, `
		SELECT member_no, first_name || ' ' || last_name, basic_salary,
		       (SELECT total_amount FROM contributions c WHERE c.member_id = m.id
		        AND DATE_TRUNC('month', c.period) = DATE_TRUNC('month', $2)
		        AND c.scheme_id = m.scheme_id) as actual_contrib,
		       ROUND(m.basic_salary * m.member_contribution_rate / 100) as expected_contrib
		FROM members m
		WHERE m.scheme_id = $1 AND m.membership_status = 'active'
	`, schemeID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var memberNo, memberName string
		var basicSalary int64
		var actualContrib, expectedContrib sql.NullFloat64

		if err := rows.Scan(&memberNo, &memberName, &basicSalary, &actualContrib, &expectedContrib); err != nil {
			continue
		}

		if !actualContrib.Valid || actualContrib.Float64 == 0 {
			warnings = append(warnings, fmt.Sprintf("Member %s (%s): No contribution received", memberNo, memberName))
			continue
		}

		if !expectedContrib.Valid || expectedContrib.Float64 == 0 {
			continue
		}

		diff := actualContrib.Float64 - expectedContrib.Float64
		if diff > 1000 || diff < -1000 {
			warnings = append(warnings, fmt.Sprintf("Member %s (%s): Contribution irregularity (expected: %.0f, actual: %.0f)",
				memberNo, memberName, expectedContrib.Float64, actualContrib.Float64))
		}
	}

	return warnings, nil
}

func (s *Service) TrackUnregisteredContributions(ctx context.Context, schemeID string, period time.Time) ([]UnregisteredContribution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, member_id, total_amount, payment_method, created_at
		FROM contributions
		WHERE scheme_id = $1 AND DATE_TRUNC('month', period) = DATE_TRUNC('month', $2)
		AND registered = false
	`, schemeID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributions []UnregisteredContribution
	for rows.Next() {
		var c UnregisteredContribution
		if err := rows.Scan(&c.ID, &c.MemberID, &c.Amount, &c.PaymentMethod, &c.CreatedAt); err != nil {
			continue
		}
		contributions = append(contributions, c)
	}
	return contributions, rows.Err()
}

type UnregisteredContribution struct {
	ID            string    `json:"id"`
	MemberID      string    `json:"member_id"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
}
