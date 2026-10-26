package bulk

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/google/uuid"
)

// StatementImportResult summarises a Labo-style member statement import.
type StatementImportResult struct {
	MemberNo       string `json:"member_no"`
	MemberName     string `json:"member_name"`
	MemberID       string `json:"member_id"`
	Periods        int    `json:"periods"`
	InsertedRows   int    `json:"inserted_rows"`
	UpdatedRows    int    `json:"updated_rows"`
	FirstPeriod    string `json:"first_period,omitempty"`
	LastPeriod     string `json:"last_period,omitempty"`
	AccountBalance int64  `json:"account_balance"`
}

// statementLine is a parsed monthly row from a Labo-style workbook.
type statementLine struct {
	Period  time.Time
	Detail  string
	DR      int64
	Emp     int64
	Empr    int64
	Balance int64
}

// parseMemberStatementWorkbook extracts member details and monthly rows from a
// MANUAL STATEMENT LABO-style Excel upload without requiring a database.
func parseMemberStatementWorkbook(data []byte) (string, string, []statementLine, error) {
	f, err := excelize.OpenReader(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return "", "", nil, fmt.Errorf("open statement workbook: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", "", nil, fmt.Errorf("workbook has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return "", "", nil, fmt.Errorf("read statement rows: %w", err)
	}

	memberNo := ""
	memberName := ""
	headerRow := -1
	col := map[string]int{}

	for i, row := range rows {
		for j, cell := range row {
			normal := strings.ToLower(strings.TrimSpace(cell))
			switch normal {
			case "pf number:", "pf number":
				if memberNo == "" {
					for k := j + 1; k < len(row); k++ {
						if strings.TrimSpace(row[k]) != "" {
							memberNo = strings.TrimSpace(row[k])
							break
						}
					}
				}
			case "name:", "name":
				if memberName == "" {
					for k := j + 1; k < len(row); k++ {
						if strings.TrimSpace(row[k]) != "" {
							memberName = strings.TrimSpace(row[k])
							break
						}
					}
				}
			}
		}

		if headerRow < 0 && foundHeaderRow(row) {
			for j, cell := range row {
				h := strings.ToLower(strings.TrimSpace(cell))
				if h == "date" || h == "details" || h == "dr" || h == "employee contribution" || h == "employer contribution" || h == "balance" {
					col[h] = j
				}
			}
			headerRow = i
		}
	}

	if memberNo == "" {
		return "", "", nil, fmt.Errorf("could not extract member number from statement")
	}
	if memberName == "" {
		memberName = memberNo
	}
	if headerRow < 0 || !hasColumns(col, "date", "details") {
		return "", "", nil, fmt.Errorf("could not find contribution table headers in statement")
	}

	var lines []statementLine
	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		dateCell := cellAt(row, col["date"])
		if strings.TrimSpace(dateCell) == "" {
			break
		}
		period, err := parseStatementDate(dateCell)
		if err != nil {
			continue
		}
		line := statementLine{
			Period:  period,
			Detail:  cellAt(row, col["details"]),
			DR:      parseMoney(cellAt(row, col["dr"])),
			Emp:     parseMoney(cellAt(row, col["employee contribution"])),
			Empr:    parseMoney(cellAt(row, col["employer contribution"])),
			Balance: parseMoney(cellAt(row, col["balance"])),
		}
		if line.Detail == "" {
			line.Detail = "Pension Contribution"
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "", "", nil, fmt.Errorf("no contribution rows found in statement")
	}

	return memberNo, memberName, lines, nil
}

// ImportMemberStatement parses a MANUAL STATEMENT LABO-style workbook and
// upserts the member and monthly contribution records.
func (s *Service) ImportMemberStatement(ctx context.Context, schemeID string, reader io.Reader, createdBy string) (*StatementImportResult, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}

	memberNo, memberName, lines, err := parseMemberStatementWorkbook(data)
	if err != nil {
		return nil, err
	}

	memberID, _, err := findOrCreateMemberStatement(ctx, s.db, schemeID, memberNo, memberName)
	if err != nil {
		return nil, err
	}

	err = s.db.Transactional(ctx, func(tx *sql.Tx) error {
		if err := deleteMemberContributions(ctx, tx, memberID); err != nil {
			return err
		}
		for _, line := range lines {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO contributions (member_id, scheme_id, period, employee_amount,
				                          employer_amount, avc_amount, status, payment_method,
				                          receipt_no)
				VALUES ($1, $2, $3, $4, $5, 0, 'reconciled', 'bank_transfer',
				        $6)
			`, memberID, schemeID, line.Period, line.Emp, line.Empr,
				"STMT-"+memberNo+"-"+line.Period.Format("200601")); err != nil {
				return fmt.Errorf("insert statement contribution: %w", err)
			}
		}
		last := lines[len(lines)-1].Balance
		if _, err := tx.ExecContext(ctx, `UPDATE members SET account_balance = $1, updated_at = NOW() WHERE id = $2`, last, memberID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &StatementImportResult{
		MemberNo:       memberNo,
		MemberName:     memberName,
		MemberID:       memberID,
		Periods:        len(lines),
		InsertedRows:   len(lines),
		FirstPeriod:    lines[0].Period.Format("2006-01-02"),
		LastPeriod:     lines[len(lines)-1].Period.Format("2006-01-02"),
		AccountBalance: lines[len(lines)-1].Balance,
	}, nil
}

func findOrCreateMemberStatement(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, schemeID, memberNo, name string) (string, bool, error) {
	var memberID string
	row := q.QueryRowContext(ctx, `SELECT id FROM members WHERE scheme_id = $1 AND member_no = $2`, schemeID, memberNo)
	switch err := row.Scan(&memberID); err {
	case nil:
		return memberID, false, nil
	case sql.ErrNoRows:
		// continue to insert
	default:
		return "", false, err
	}

	id := uuid.New().String()
	first := strings.TrimSpace(name)
	last := "-"
	fields := strings.Fields(name)
	if len(fields) > 1 {
		first = fields[0]
		last = fields[len(fields)-1]
	}
	if executor, ok := q.(interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}); ok {
		_, err := executor.ExecContext(ctx, `
			INSERT INTO members (id, scheme_id, member_no, first_name, last_name, email, phone, department,
		                     date_of_birth, date_joined_scheme, membership_status, portal_enabled, account_balance)
		VALUES ($1, $2, $3, $4, $5, '', '', '', CURRENT_DATE, CURRENT_DATE, 'active', true, 0)
		`, id, schemeID, memberNo, first, last)
		if err != nil {
			return "", true, err
		}
	} else {
		return "", false, fmt.Errorf("database handle does not support ExecContext")
	}
	return id, true, nil
}

func deleteMemberContributions(ctx context.Context, tx *sql.Tx, memberID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM contributions WHERE member_id = $1`, memberID)
	return err
}

func parseStatementDate(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, fmt.Errorf("blank date")
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"01-02-06",
		"1-2-06",
		"01/02/2006",
		"02/01/2006",
		"01/02/06",
		"02/01/06",
		"01.02.06",
		"January 2, 2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse statement date %q", v)
}

func parseMoney(v string) int64 {
	v = strings.TrimSpace(strings.ReplaceAll(v, ",", ""))
	if v == "" {
		return 0
	}
	g, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return int64(math.Round(g))
}

func cellAt(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func foundHeaderRow(row []string) bool {
	count := 0
	for _, cell := range row {
		h := strings.ToLower(strings.TrimSpace(cell))
		if h == "date" || h == "details" || h == "dr" || h == "employee contribution" || strings.HasPrefix(h, "employee") || h == "employer contribution" || strings.HasPrefix(h, "employer") || h == "balance" {
			count++
		}
	}
	return count >= 4
}

func hasColumns(col map[string]int, keys ...string) bool {
	for _, key := range keys {
		if _, ok := col[key]; !ok {
			return false
		}
	}
	return true
}

func colHasHeader(col map[string]int, key string) bool {
	for k := range col {
		if k == key {
			return true
		}
	}
	return false
}
