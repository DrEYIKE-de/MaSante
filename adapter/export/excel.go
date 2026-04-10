// Package export provides driven adapters for generating Excel and PDF reports.
package export

import (
	"fmt"
	"io"

	"github.com/masante/masante/domain"
	"github.com/xuri/excelize/v2"
)

// PatientsToExcel writes a patient list to an Excel file.
func PatientsToExcel(w io.Writer, patients []domain.Patient, title string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Patients"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Code", "Nom", "Prenom", "Sexe", "Telephone", "Quartier", "Statut", "Score Risque", "Derniere inscription"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Bold header style.
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetRowStyle(sheet, 1, 1, style)

	for row, p := range patients {
		r := row + 2
		f.SetCellValue(sheet, cellName(1, r), p.Code)
		f.SetCellValue(sheet, cellName(2, r), p.LastName)
		f.SetCellValue(sheet, cellName(3, r), p.FirstName)
		f.SetCellValue(sheet, cellName(4, r), p.Sex)
		f.SetCellValue(sheet, cellName(5, r), p.Phone)
		f.SetCellValue(sheet, cellName(6, r), p.District)
		f.SetCellValue(sheet, cellName(7, r), string(p.Status))
		f.SetCellValue(sheet, cellName(8, r), p.RiskScore)
		f.SetCellValue(sheet, cellName(9, r), p.EnrollmentDate.Format("02/01/2006"))
	}

	// Auto-width columns.
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 18)
	}

	return f.Write(w)
}

// AppointmentsToExcel writes an appointment list to an Excel file.
func AppointmentsToExcel(w io.Writer, apts []domain.Appointment) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Rendez-vous"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Date", "Heure", "Patient", "Code", "Type", "Statut", "Notes"}
	for i, h := range headers {
		f.SetCellValue(sheet, cellName(i+1, 1), h)
	}

	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	for row, a := range apts {
		r := row + 2
		f.SetCellValue(sheet, cellName(1, r), a.Date.Format("02/01/2006"))
		f.SetCellValue(sheet, cellName(2, r), a.Time)
		f.SetCellValue(sheet, cellName(3, r), a.PatientName)
		f.SetCellValue(sheet, cellName(4, r), a.PatientCode)
		f.SetCellValue(sheet, cellName(5, r), string(a.Type))
		f.SetCellValue(sheet, cellName(6, r), string(a.Status))
		f.SetCellValue(sheet, cellName(7, r), a.Notes)
	}

	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 18)
	}

	return f.Write(w)
}

// MonthlyReportExcel generates a monthly summary Excel.
func MonthlyReportExcel(w io.Writer, month string, patientCounts map[domain.PatientStatus]int, aptCounts map[domain.AppointmentStatus]int) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Rapport " + month
	f.SetSheetName("Sheet1", sheet)

	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	f.SetCellValue(sheet, "A1", "Rapport mensuel — "+month)
	f.SetCellStyle(sheet, "A1", "A1", style)

	f.SetCellValue(sheet, "A3", "Patients")
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A3", "A3", bold)

	row := 4
	for status, count := range patientCounts {
		f.SetCellValue(sheet, cellName(1, row), string(status))
		f.SetCellValue(sheet, cellName(2, row), count)
		row++
	}

	row += 1
	f.SetCellValue(sheet, cellName(1, row), "Rendez-vous")
	f.SetCellStyle(sheet, cellName(1, row), cellName(1, row), bold)
	row++
	for status, count := range aptCounts {
		f.SetCellValue(sheet, cellName(1, row), string(status))
		f.SetCellValue(sheet, cellName(2, row), count)
		row++
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 12)

	return f.Write(w)
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

// FormatFileName generates a safe filename for downloads.
func FormatFileName(prefix, ext string) string {
	return fmt.Sprintf("%s.%s", prefix, ext)
}
