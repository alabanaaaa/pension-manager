package bulk

import (
	"bytes"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func sampleStatementWorkbookBytes(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	sheet := "Sheet1"
	rows := [][]interface{}{
		{"", "PF NUMBER:", "", "0001"},
		{"", "NAME:", "", "HON TEST MEMBER"},
		{"Date", "Details", "DR", "Employee Contribution", "Employer Contribution", "Balance"},
		{"04-30-13", "Pension Contribution", "0.00", "100.00", "200.00", "300.00"},
		{"05-31-13", "Pension Contribution", "0.00", "110.00", "220.00", "630.00"},
	}
	for i, row := range rows {
		for j, value := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}

func TestParseMemberStatementWorkbook(t *testing.T) {
	data := sampleStatementWorkbookBytes(t)
	memberNo, memberName, lines, err := parseMemberStatementWorkbook(data)
	if err != nil {
		t.Fatalf("parse workbook: %v", err)
	}
	if memberNo != "0001" {
		t.Fatalf("member number: got %q", memberNo)
	}
	if memberName != "HON TEST MEMBER" {
		t.Fatalf("member name: got %q", memberName)
	}
	if len(lines) != 2 {
		t.Fatalf("line count: got %d", len(lines))
	}
	if lines[0].Emp != 100 || lines[0].Empr != 200 || lines[0].Balance != 300 {
		t.Fatalf("first line money: %+v", lines[0])
	}
	expected, _ := time.Parse("01-02-06", "04-30-13")
	if !lines[0].Period.Equal(expected) {
		t.Fatalf("period: got %s want %s", lines[0].Period, expected)
	}
}

func TestParseMemberStatementWorkbookMissingMemberNo(t *testing.T) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	row := []interface{}{"Month", "Details", "DR", "Employee Contribution", "Employer Contribution", "Balance"}
	for j, value := range row {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		_ = f.SetCellValue(sheet, cell, value)
	}
	var buf bytes.Buffer
	_ = f.Write(&buf)
	if _, _, _, err := parseMemberStatementWorkbook(buf.Bytes()); err == nil {
		t.Fatal("expected missing member number error")
	}
}

func TestParseMemberStatementWorkbookMissingHeaders(t *testing.T) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	_ = f.SetCellValue(sheet, "B2", "PF NUMBER:")
	_ = f.SetCellValue(sheet, "D2", "0001")
	_ = f.SetCellValue(sheet, "B4", "NAME:")
	_ = f.SetCellValue(sheet, "D4", "HON TEST MEMBER")
	var buf bytes.Buffer
	_ = f.Write(&buf)
	if _, _, _, err := parseMemberStatementWorkbook(buf.Bytes()); err == nil {
		t.Fatal("expected missing header error")
	}
}

func TestParseMemberStatementWorkbookNoRows(t *testing.T) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	headers := []interface{}{"Date", "Details", "DR", "Employee Contribution", "Employer Contribution", "Balance"}
	for j, value := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		_ = f.SetCellValue(sheet, cell, value)
	}
	_ = f.SetCellValue(sheet, "B2", "PF NUMBER:")
	_ = f.SetCellValue(sheet, "D2", "0001")
	var buf bytes.Buffer
	_ = f.Write(&buf)
	if _, _, _, err := parseMemberStatementWorkbook(buf.Bytes()); err == nil {
		t.Fatal("expected no contribution rows error")
	}
}

func TestParseStatementDate(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"04-30-13", true},
		{"2013-04-30", true},
		{"30/04/2013", true},
		{"31/02/2013", false},
	}
	for _, tc := range cases {
		if _, err := parseStatementDate(tc.in); (err == nil) != tc.ok {
			t.Errorf("parse %q: got err=%v, want ok=%v", tc.in, err, tc.ok)
		}
	}
}

func TestParseMoney(t *testing.T) {
	if got := parseMoney("1,234.50"); got != 1235 {
		t.Fatalf("got %d", got)
	}
	if got := parseMoney(""); got != 0 {
		t.Fatalf("got %d", got)
	}
}
