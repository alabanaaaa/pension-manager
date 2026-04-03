package hospital

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"pension-manager/core/domain"

	"github.com/xuri/excelize/v2"
)

// GenerateMedicalExpenditureExcel generates an Excel report of medical expenditures
func GenerateMedicalExpenditureExcel(expenditures []*domain.MedicalExpenditure, members map[string]*domain.Member, hospitals map[string]*domain.Hospital) (*excelize.File, error) {
	f := excelize.NewFile()

	// Create a sheet for medical expenditures
	sheetName := "Medical Expenditures"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// Set header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
	})
	if err != nil {
		return nil, err
	}

	// Define headers
	headers := []string{
		"Expenditure ID", "Member ID", "Member Name", "Hospital ID", "Hospital Name",
		"Date of Service", "Date Submitted", "Service Type", "Description",
		"Amount Charged (KES)", "Amount Covered (KES)", "Member Responsibility (KES)",
		"Status", "Invoice Number", "Receipt Number",
	}

	// Set header values
	for col, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+col)), 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Set data rows
	for row, exp := range expenditures {
		member := members[exp.MemberID]
		hospital := hospitals[exp.HospitalID]

		memberName := ""
		if member != nil {
			memberName = fmt.Sprintf("%s %s", member.FirstName, member.LastName)
		}

		hospitalName := ""
		if hospital != nil {
			hospitalName = hospital.Name
		}

		data := []interface{}{
			exp.ID,
			exp.MemberID,
			memberName,
			exp.HospitalID,
			hospitalName,
			exp.DateOfService.Format("2006-01-02"),
			exp.DateSubmitted.Format("2006-01-02"),
			exp.ServiceType,
			exp.Description,
			fmt.Sprintf("%.2f", float64(exp.AmountCharged)/100),
			fmt.Sprintf("%.2f", float64(exp.AmountCovered)/100),
			fmt.Sprintf("%.2f", float64(exp.MemberResponsibility)/100),
			exp.Status,
			exp.InvoiceNumber,
			exp.ReceiptNumber,
		}

		for colIdx, value := range data {
			colName := string('A' + uint8(colIdx))
			cell := fmt.Sprintf("%s%d", colName, row+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Set column widths
	for i := 0; i < len(headers); i++ {
		colName := string('A' + uint8(i))
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	return f, nil
}

// GenerateMedicalExpenditureCSV generates a CSV report of medical expenditures
func GenerateMedicalExpenditureCSV(expenditures []*domain.MedicalExpenditure, members map[string]*domain.Member, hospitals map[string]*domain.Hospital) (io.Reader, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Expenditure ID", "Member ID", "Member Name", "Hospital ID", "Hospital Name",
		"Date of Service", "Date Submitted", "Service Type", "Description",
		"Amount Charged (KES)", "Amount Covered (KES)", "Member Responsibility (KES)",
		"Status", "Invoice Number", "Receipt Number",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, exp := range expenditures {
		member := members[exp.MemberID]
		hospital := hospitals[exp.HospitalID]

		memberName := ""
		if member != nil {
			memberName = fmt.Sprintf("%s %s", member.FirstName, member.LastName)
		}

		hospitalName := ""
		if hospital != nil {
			hospitalName = hospital.Name
		}

		record := []string{
			exp.ID,
			exp.MemberID,
			memberName,
			exp.HospitalID,
			hospitalName,
			exp.DateOfService.Format("2006-01-02"),
			exp.DateSubmitted.Format("2006-01-02"),
			exp.ServiceType,
			exp.Description,
			fmt.Sprintf("%.2f", float64(exp.AmountCharged)/100),
			fmt.Sprintf("%.2f", float64(exp.AmountCovered)/100),
			fmt.Sprintf("%.2f", float64(exp.MemberResponsibility)/100),
			exp.Status,
			exp.InvoiceNumber,
			exp.ReceiptNumber,
		}

		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return strings.NewReader(buf.String()), nil
}
