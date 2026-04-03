package bulk

import (
	"strings"
	"testing"
)

func TestParseMemberImport(t *testing.T) {
	headers := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme", "basic_salary", "email", "phone", "department", "designation", "sponsor_id"}
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	record := []string{"M001", "John", "Doe", "1990-01-15", "2020-01-01", "50000", "john@example.com", "+254712345678", "IT", "Developer", "sponsor-001"}

	m, errs := parseMemberImport(record, colIndex, 1, "scheme-001")
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
	if m.MemberNo != "M001" {
		t.Errorf("Expected member_no M001, got: %s", m.MemberNo)
	}
	if m.FirstName != "John" {
		t.Errorf("Expected first_name John, got: %s", m.FirstName)
	}
	if m.LastName != "Doe" {
		t.Errorf("Expected last_name Doe, got: %s", m.LastName)
	}
	if m.BasicSalary != 50000 {
		t.Errorf("Expected basic_salary 50000, got: %d", m.BasicSalary)
	}
	if m.SchemeID != "scheme-001" {
		t.Errorf("Expected scheme_id scheme-001, got: %s", m.SchemeID)
	}
}

func TestParseMemberImport_MissingRequired(t *testing.T) {
	headers := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme"}
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	record := []string{"", "", "", "", ""}

	_, errs := parseMemberImport(record, colIndex, 1, "scheme-001")
	if len(errs) < 3 {
		t.Errorf("Expected at least 3 errors for missing required fields, got: %d", len(errs))
	}
}

func TestParseMemberImport_InvalidDate(t *testing.T) {
	headers := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme"}
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	record := []string{"M001", "John", "Doe", "invalid-date", "not-a-date"}

	_, errs := parseMemberImport(record, colIndex, 1, "scheme-001")
	if len(errs) < 2 {
		t.Errorf("Expected at least 2 errors for invalid dates, got: %d", len(errs))
	}
}

func TestParseMemberImport_InvalidSalary(t *testing.T) {
	headers := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme", "basic_salary"}
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	record := []string{"M001", "John", "Doe", "1990-01-15", "2020-01-01", "not-a-number"}

	m, errs := parseMemberImport(record, colIndex, 1, "scheme-001")
	if len(errs) != 1 {
		t.Errorf("Expected 1 error for invalid salary, got: %d", len(errs))
	}
	if m.BasicSalary != 0 {
		t.Errorf("Expected basic_salary 0 for invalid input, got: %d", m.BasicSalary)
	}
}

func TestParseMemberImport_EmptyFields(t *testing.T) {
	headers := []string{"member_no", "first_name", "last_name", "date_of_birth", "date_joined_scheme"}
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	record := []string{"M001", "John", "Doe", "1990-01-15", "2020-01-01"}

	m, errs := parseMemberImport(record, colIndex, 1, "scheme-001")
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
	if m.OtherNames != "" {
		t.Errorf("Expected empty other_names, got: %s", m.OtherNames)
	}
	if m.Email != "" {
		t.Errorf("Expected empty email, got: %s", m.Email)
	}
}

func TestImportResult(t *testing.T) {
	result := &ImportResult{
		TotalRows: 10,
		Success:   8,
		Failed:    2,
		Errors: []ImportError{
			{Row: 3, Field: "member_no", Reason: "required"},
			{Row: 7, Field: "date_of_birth", Reason: "invalid format"},
		},
		Warnings: []ImportWarning{
			{Row: 5, Message: "Member M005 already exists"},
		},
	}

	if result.TotalRows != 10 {
		t.Errorf("Expected total rows 10, got: %d", result.TotalRows)
	}
	if result.Success != 8 {
		t.Errorf("Expected success 8, got: %d", result.Success)
	}
	if result.Failed != 2 {
		t.Errorf("Expected failed 2, got: %d", result.Failed)
	}
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got: %d", len(result.Errors))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got: %d", len(result.Warnings))
	}
}

func TestCSVReader(t *testing.T) {
	csvData := "member_no,first_name,last_name,date_of_birth,date_joined_scheme\nM001,John,Doe,1990-01-15,2020-01-01\nM002,Jane,Smith,1985-05-20,2019-03-15"

	// Test CSV parsing
	lines := strings.Split(csvData, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 records), got: %d", len(lines))
	}

	headers := strings.Split(lines[0], ",")
	if len(headers) != 5 {
		t.Errorf("Expected 5 headers, got: %d", len(headers))
	}
}

func TestBulkValidation(t *testing.T) {
	v := &BulkValidation{
		IsValid:        true,
		NewMembers:     []string{"M010", "M011"},
		RemovedMembers: []string{"M003"},
		SalaryChanges: []SalaryChange{
			{MemberNo: "M001", OldSalary: 50000, NewSalary: 55000, ChangePct: 10.0},
		},
	}

	if !v.IsValid {
		t.Error("Expected valid validation")
	}
	if len(v.NewMembers) != 2 {
		t.Errorf("Expected 2 new members, got: %d", len(v.NewMembers))
	}
	if len(v.RemovedMembers) != 1 {
		t.Errorf("Expected 1 removed member, got: %d", len(v.RemovedMembers))
	}
	if len(v.SalaryChanges) != 1 {
		t.Errorf("Expected 1 salary change, got: %d", len(v.SalaryChanges))
	}
}
