package export

import (
	"fmt"
	"io"

	"github.com/masante/masante/domain"
	"github.com/go-pdf/fpdf"
)

// PatientsToPDF writes a patient list to a PDF file.
func PatientsToPDF(w io.Writer, patients []domain.Patient, title string) error {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	headers := []string{"Code", "Nom", "Prenom", "Sexe", "Tel", "Quartier", "Statut", "Risque"}
	widths := []float64{35, 35, 35, 15, 35, 35, 30, 15}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, p := range patients {
		vals := []string{
			p.Code, p.LastName, p.FirstName, p.Sex,
			p.Phone, p.District, string(p.Status),
			fmt.Sprintf("%d", p.RiskScore),
		}
		for i, v := range vals {
			pdf.CellFormat(widths[i], 6, v, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}

	return pdf.Output(w)
}

// MonthlyReportPDF generates a monthly summary PDF.
func MonthlyReportPDF(w io.Writer, month string, centerName string, patientCounts map[domain.PatientStatus]int, aptCounts map[domain.AppointmentStatus]int) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 12, "Rapport mensuel — "+month, "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 8, centerName, "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, "Patients", "", 1, "", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "", 10)
	total := 0
	for status, count := range patientCounts {
		total += count
		pdf.CellFormat(60, 7, string(status), "", 0, "", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", count), "", 1, "R", false, 0, "")
	}
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(60, 7, "Total", "", 0, "", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", total), "", 1, "R", false, 0, "")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, "Rendez-vous", "", 1, "", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "", 10)
	for status, count := range aptCounts {
		pdf.CellFormat(60, 7, string(status), "", 0, "", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", count), "", 1, "R", false, 0, "")
	}

	return pdf.Output(w)
}
